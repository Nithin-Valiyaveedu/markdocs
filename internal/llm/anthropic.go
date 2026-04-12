package llm

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicProvider implements LLMProvider using the Anthropic API.
type AnthropicProvider struct {
	client *anthropic.Client
	model  string
}

var _ LLMProvider = (*AnthropicProvider)(nil)

// NewAnthropicProvider creates an AnthropicProvider with the given API key and model.
func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
	if model == "" {
		model = "claude-haiku-4-5-20251001"
	}
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &AnthropicProvider{client: &client, model: model}
}

// Complete sends the prompt to Anthropic and returns the generated text.
func (p *AnthropicProvider) Complete(ctx context.Context, prompt string) (string, error) {
	msg, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: 4096,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("anthropic complete: %w", err)
	}
	if len(msg.Content) == 0 {
		return "", fmt.Errorf("anthropic complete: empty response")
	}
	return msg.Content[0].Text, nil
}

// Model returns the model name used by this provider.
func (p *AnthropicProvider) Model() string {
	return p.model
}
