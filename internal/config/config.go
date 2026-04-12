package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ProviderName identifies an LLM backend.
type ProviderName string

const (
	ProviderAnthropic ProviderName = "anthropic"
	ProviderOpenAI    ProviderName = "openai"
	ProviderOllama    ProviderName = "ollama"
)

// Config holds markdocs configuration stored at ~/.markdocs/config.json.
type Config struct {
	Provider   ProviderName `json:"provider"`
	Model      string       `json:"model"`
	APIKey     string       `json:"api_key,omitempty"`
	BaseURL    string       `json:"base_url,omitempty"` // for OpenAI-compatible endpoints
	OllamaHost string       `json:"ollama_host,omitempty"`
}

// DefaultConfigPath returns the path to the config file (~/.markdocs/config.json).
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home dir: %w", err)
	}
	return filepath.Join(home, ".markdocs", "config.json"), nil
}

// Exists reports whether the config file exists.
func Exists() bool {
	path, err := DefaultConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// Load reads and returns the config from the default path.
func Load() (*Config, error) {
	path, err := DefaultConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// Save writes cfg to the default config path, creating directories as needed.
func Save(cfg *Config) error {
	path, err := DefaultConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
