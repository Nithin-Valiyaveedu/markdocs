package llm

import (
	"context"
	"fmt"

	"github.com/nithinworks/markdocs/internal/config"
)

// LLMProvider is the interface all LLM backends must implement.
type LLMProvider interface {
	Complete(ctx context.Context, prompt string) (string, error)
	Model() string
}

// NewProvider constructs the appropriate LLM provider from the given config.
func NewProvider(cfg *config.Config) (LLMProvider, error) {
	switch cfg.Provider {
	case config.ProviderAnthropic:
		return NewAnthropicProvider(cfg.APIKey, cfg.Model), nil
	case config.ProviderOpenAI:
		return NewOpenAIProvider(cfg.APIKey, cfg.Model, cfg.BaseURL), nil
	case config.ProviderOllama:
		host := cfg.OllamaHost
		if host == "" {
			host = "http://localhost:11434"
		}
		return NewOllamaProvider(host, cfg.Model)
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
	}
}
