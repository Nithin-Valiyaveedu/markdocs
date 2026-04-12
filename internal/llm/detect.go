package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/nithinworks/markdocs/internal/config"
)

// AutoDetect returns the first available LLMProvider detected from the
// environment. Detection order: Anthropic → OpenAI → Ollama.
// Returns an error with instructions if no provider is found.
func AutoDetect(ctx context.Context) (LLMProvider, error) {
	provider, ok := config.DetectProvider()
	if !ok {
		return nil, fmt.Errorf("no LLM provider detected — run 'markdocs init' to configure one")
	}

	switch provider {
	case config.ProviderAnthropic:
		return NewAnthropicProvider(os.Getenv("ANTHROPIC_API_KEY"), ""), nil
	case config.ProviderOpenAI:
		return NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"), "", ""), nil
	case config.ProviderOllama:
		return NewOllamaProvider("http://localhost:11434", "llama3.2")
	default:
		return nil, fmt.Errorf("unknown provider detected: %s", provider)
	}
}
