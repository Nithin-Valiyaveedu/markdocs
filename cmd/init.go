package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/config"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/llm"
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
	green := color.New(color.FgGreen).SprintFunc()

	// Try auto-detection first unless skipped
	if !initSkipDetect {
		if detected, ok := config.DetectProvider(); ok {
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Detected %s from environment. Use it?", string(detected)),
				IsConfirm: true,
			}
			_, err := prompt.Run()
			if err == nil {
				// User confirmed — build config from detected provider
				cfg, err := buildDetectedConfig(detected)
				if err != nil {
					return err
				}
				if err := validateAndSave(ctx, cfg); err != nil {
					return err
				}
				fmt.Println(green("✓ Configuration saved."))
				return nil
			}
			// User declined — fall through to manual setup
		}
	}

	// Manual provider selection
	providerSelect := promptui.Select{
		Label: "Select LLM provider",
		Items: []string{"Anthropic", "OpenAI", "OpenAI-compatible (Groq, Together, etc.)", "Ollama (local)"},
	}
	idx, _, err := providerSelect.Run()
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
	fmt.Println(green("✓ Configuration saved to ~/.markdocs/config.json"))
	return nil
}

func validateAndSave(ctx context.Context, cfg *config.Config) error {
	s := spinner.New(spinner.CharSets[14], 80*time.Millisecond)
	s.Suffix = " Validating credentials..."
	s.Start()

	provider, err := llm.NewProvider(cfg)
	var validationErr error
	if err == nil {
		_, validationErr = provider.Complete(ctx, "Reply with the single word: ok")
	} else {
		validationErr = err
	}
	s.Stop()

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
	apiKey, err := promptSecret("Anthropic API key")
	if err != nil {
		return nil, err
	}
	model, err := promptWithDefault("Model", "claude-haiku-4-5-20251001")
	if err != nil {
		return nil, err
	}
	return &config.Config{Provider: config.ProviderAnthropic, APIKey: apiKey, Model: model}, nil
}

func promptOpenAIConfig(defaultModel, baseURL string) (*config.Config, error) {
	if defaultModel == "" {
		defaultModel = "gpt-4o-mini"
	}
	apiKey, err := promptSecret("OpenAI API key")
	if err != nil {
		return nil, err
	}
	model, err := promptWithDefault("Model", defaultModel)
	if err != nil {
		return nil, err
	}
	return &config.Config{Provider: config.ProviderOpenAI, APIKey: apiKey, Model: model, BaseURL: baseURL}, nil
}

func promptOpenAICompatibleConfig() (*config.Config, error) {
	baseURL, err := promptWithDefault("Base URL", "https://api.groq.com/openai/v1")
	if err != nil {
		return nil, err
	}
	return promptOpenAIConfig("llama-3.1-8b-instant", baseURL)
}

func promptOllamaConfig() (*config.Config, error) {
	host, err := promptWithDefault("Ollama host", "http://localhost:11434")
	if err != nil {
		return nil, err
	}
	model, err := promptWithDefault("Model", "llama3.2")
	if err != nil {
		return nil, err
	}
	return &config.Config{Provider: config.ProviderOllama, OllamaHost: host, Model: model}, nil
}

func promptSecret(label string) (string, error) {
	p := promptui.Prompt{Label: label, Mask: '*'}
	return p.Run()
}

func promptWithDefault(label, defaultVal string) (string, error) {
	p := promptui.Prompt{Label: label, Default: defaultVal}
	return p.Run()
}
