package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Nithin-Valiyaveedu/markdocs/internal/config"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/llm"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/ui"
	"github.com/spf13/cobra"
)

var initSkipDetect bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Configure LLM provider and credentials.",
	Long:  "Interactively configure the LLM provider (Anthropic, OpenAI, or Ollama) used to find and compile documentation.",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initSkipDetect, "skip-detect", false, "Skip auto-detection of available providers")
}

func runInit(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Try auto-detection first unless skipped
	if !initSkipDetect {
		if detected, ok := config.DetectProvider(); ok {
			use, err := ui.UseDetected(string(detected))
			if err != nil {
				return fmt.Errorf("prompt: %w", err)
			}
			if use {
				cfg, err := buildDetectedConfig(detected)
				if err != nil {
					return err
				}
				if err := validateAndSave(ctx, cfg); err != nil {
					return err
				}
				ui.Success("Configuration saved.")
				return nil
			}
			// User declined — fall through to manual setup
		}
	}

	// Manual provider selection
	idx, err := ui.SelectProvider("")
	if err != nil {
		return fmt.Errorf("provider selection: %w", err)
	}

	var cfg *config.Config
	switch idx {
	case 0:
		cfg, err = promptAnthropicConfig()
	case 1:
		cfg, err = promptOpenAIConfig("", "")
	case 2:
		cfg, err = promptOpenAICompatibleConfig()
	case 3:
		cfg, err = promptOllamaConfig()
	}
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}

	if err := validateAndSave(ctx, cfg); err != nil {
		return err
	}
	ui.Success("Configuration saved to ~/.markdocs/config.json")
	return nil
}

func validateAndSave(ctx context.Context, cfg *config.Config) error {
	var validationErr error
	err := ui.Spin("Validating credentials...", func() error {
		provider, err := llm.NewProvider(cfg)
		if err != nil {
			return err
		}
		_, err = provider.Complete(ctx, "Reply with the single word: ok")
		return err
	})
	if err != nil {
		validationErr = err
	}

	if validationErr != nil {
		fmt.Fprintf(os.Stderr, "✗ Validation failed: %s\n", validationErr)
		os.Exit(2)
	}

	return config.Save(cfg)
}

func buildDetectedConfig(provider config.ProviderName) (*config.Config, error) {
	switch provider {
	case config.ProviderAnthropic:
		return promptAnthropicConfig()
	case config.ProviderOpenAI:
		return promptOpenAIConfig("", "")
	case config.ProviderOllama:
		return promptOllamaConfig()
	}
	return &config.Config{Provider: provider}, nil
}

func promptAnthropicConfig() (*config.Config, error) {
	apiKey, err := ui.Secret("Anthropic API key")
	if err != nil {
		return nil, err
	}
	model, err := ui.InputWithDefault("Model", "claude-haiku-4-5-20251001")
	if err != nil {
		return nil, err
	}
	return &config.Config{Provider: config.ProviderAnthropic, APIKey: apiKey, Model: model}, nil
}

func promptOpenAIConfig(defaultModel, baseURL string) (*config.Config, error) {
	if defaultModel == "" {
		defaultModel = "gpt-4o-mini"
	}
	apiKey, err := ui.Secret("OpenAI API key")
	if err != nil {
		return nil, err
	}
	model, err := ui.InputWithDefault("Model", defaultModel)
	if err != nil {
		return nil, err
	}
	return &config.Config{Provider: config.ProviderOpenAI, APIKey: apiKey, Model: model, BaseURL: baseURL}, nil
}

func promptOpenAICompatibleConfig() (*config.Config, error) {
	baseURL, err := ui.InputWithDefault("Base URL", "https://api.groq.com/openai/v1")
	if err != nil {
		return nil, err
	}
	return promptOpenAIConfig("llama-3.1-8b-instant", baseURL)
}

func promptOllamaConfig() (*config.Config, error) {
	host, err := ui.InputWithDefault("Ollama host", "http://localhost:11434")
	if err != nil {
		return nil, err
	}
	model, err := ui.InputWithDefault("Model", "llama3.2")
	if err != nil {
		return nil, err
	}
	return &config.Config{Provider: config.ProviderOllama, OllamaHost: host, Model: model}, nil
}
