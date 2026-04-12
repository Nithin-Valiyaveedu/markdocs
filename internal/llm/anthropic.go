package llm

import (
	"context"
	"errors"
)

// AnthropicProvider implements LLMProvider using the Anthropic API.
// Full implementation added in Phase 12.
type AnthropicProvider struct {
	apiKey string
	model  string
}

var _ LLMProvider = (*AnthropicProvider)(nil)

// NewAnthropicProvider creates an AnthropicProvider with the given credentials.
func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
	if model == "" {
		model = "claude-haiku-4-5-20251001"
	}
	return &AnthropicProvider{apiKey: apiKey, model: model}
}

// Complete sends the prompt to Anthropic and returns the generated text.
func (p *AnthropicProvider) Complete(ctx context.Context, prompt string) (string, error) {
	return "", errors.New("anthropic provider: not yet implemented — coming in Phase 12")
}

// Model returns the model name used by this provider.
func (p *AnthropicProvider) Model() string {
	return p.model
}
