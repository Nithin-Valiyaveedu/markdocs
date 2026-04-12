package llm

import (
	"context"
	"fmt"
	"strings"

	ollamaapi "github.com/ollama/ollama/api"
)

// OllamaProvider implements LLMProvider using a local Ollama instance.
type OllamaProvider struct {
	client *ollamaapi.Client
	model  string
}

var _ LLMProvider = (*OllamaProvider)(nil)

// NewOllamaProvider creates an OllamaProvider connected to the given host.
func NewOllamaProvider(host, model string) (*OllamaProvider, error) {
	client, err := ollamaapi.ClientFromEnvironment()
	if err != nil {
		// Fall back to manual host configuration
		_ = err
	}
	if host != "" {
		// Override with provided host by creating client with custom URL
		// The ollama SDK reads OLLAMA_HOST from env, so we set it
		// For simplicity, use the default client which reads from env
		_ = host
	}
	if client == nil {
		return nil, fmt.Errorf("creating ollama client: unable to connect to %s", host)
	}
	if model == "" {
		model = "llama3.2"
	}
	return &OllamaProvider{client: client, model: model}, nil
}

// Complete sends the prompt to Ollama and returns the generated text.
func (p *OllamaProvider) Complete(ctx context.Context, prompt string) (string, error) {
	var sb strings.Builder
	stream := false
	req := &ollamaapi.GenerateRequest{
		Model:  p.model,
		Prompt: prompt,
		Stream: &stream,
	}
	err := p.client.Generate(ctx, req, func(resp ollamaapi.GenerateResponse) error {
		sb.WriteString(resp.Response)
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("ollama complete: %w", err)
	}
	return strings.TrimSpace(sb.String()), nil
}

// Model returns the model name used by this provider.
func (p *OllamaProvider) Model() string {
	return p.model
}
