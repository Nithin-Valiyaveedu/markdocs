package llm

import (
	"context"
	"errors"
)

// OpenAIProvider implements LLMProvider using the OpenAI API (or any
// OpenAI-compatible endpoint via BaseURL).
// Full implementation added in Phase 12.
type OpenAIProvider struct {
	apiKey  string
	model   string
	baseURL string // empty = use default OpenAI endpoint
}

var _ LLMProvider = (*OpenAIProvider)(nil)

// NewOpenAIProvider creates an OpenAIProvider with the given credentials.
// Set baseURL to a non-empty string to use an OpenAI-compatible endpoint
// (e.g. Groq, Together AI).
func NewOpenAIProvider(apiKey, model, baseURL string) *OpenAIProvider {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIProvider{apiKey: apiKey, model: model, baseURL: baseURL}
}

// Complete sends the prompt to the OpenAI API and returns the generated text.
func (p *OpenAIProvider) Complete(ctx context.Context, prompt string) (string, error) {
	return "", errors.New("openai provider: not yet implemented — coming in Phase 12")
}

// Model returns the model name used by this provider.
func (p *OpenAIProvider) Model() string {
	return p.model
}
