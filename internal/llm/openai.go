package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// OpenAIProvider implements LLMProvider using the OpenAI API (or any
// OpenAI-compatible endpoint via BaseURL — e.g. Groq, Together AI).
type OpenAIProvider struct {
	client  *openai.Client
	model   string
	baseURL string
}

var _ LLMProvider = (*OpenAIProvider)(nil)

// NewOpenAIProvider creates an OpenAIProvider with the given credentials.
// Set baseURL to a non-empty string to use an OpenAI-compatible endpoint.
func NewOpenAIProvider(apiKey, model, baseURL string) *OpenAIProvider {
	if model == "" {
		model = "gpt-4o-mini"
	}
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	client := openai.NewClient(opts...)
	return &OpenAIProvider{client: &client, model: model, baseURL: baseURL}
}

// Complete sends the prompt to the OpenAI API and returns the generated text.
func (p *OpenAIProvider) Complete(ctx context.Context, prompt string) (string, error) {
	completion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: p.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	})
	if err != nil {
		return "", fmt.Errorf("openai complete: %w", err)
	}
	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("openai complete: empty response")
	}
	return completion.Choices[0].Message.Content, nil
}

// Model returns the model name used by this provider.
func (p *OpenAIProvider) Model() string {
	return p.model
}
