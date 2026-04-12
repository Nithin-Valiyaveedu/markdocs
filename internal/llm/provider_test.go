package llm

import (
	"context"
	"testing"

	"github.com/nithinworks/markdocs/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLMProvider is a test double for LLMProvider, reusable across all packages.
type MockLLMProvider struct {
	response string
	err      error
	model    string
	Prompts  []string
}

var _ LLMProvider = (*MockLLMProvider)(nil)

func (m *MockLLMProvider) Complete(ctx context.Context, prompt string) (string, error) {
	m.Prompts = append(m.Prompts, prompt)
	return m.response, m.err
}

func (m *MockLLMProvider) Model() string {
	return m.model
}

func TestNewProviderAnthropic(t *testing.T) {
	cfg := &config.Config{Provider: config.ProviderAnthropic, APIKey: "key", Model: "model"}
	p, err := NewProvider(cfg)
	require.NoError(t, err)
	_, ok := p.(*AnthropicProvider)
	assert.True(t, ok, "expected *AnthropicProvider")
}

func TestNewProviderOpenAI(t *testing.T) {
	cfg := &config.Config{Provider: config.ProviderOpenAI, APIKey: "key", Model: "model"}
	p, err := NewProvider(cfg)
	require.NoError(t, err)
	_, ok := p.(*OpenAIProvider)
	assert.True(t, ok, "expected *OpenAIProvider")
}

func TestNewProviderUnknown(t *testing.T) {
	cfg := &config.Config{Provider: "unknown"}
	_, err := NewProvider(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown provider")
}

func TestAnthropicModel(t *testing.T) {
	p := NewAnthropicProvider("key", "my-model")
	assert.Equal(t, "my-model", p.Model())
}

func TestAnthropicModelDefault(t *testing.T) {
	p := NewAnthropicProvider("key", "")
	assert.Equal(t, "claude-haiku-4-5-20251001", p.Model())
}

func TestOpenAIModel(t *testing.T) {
	p := NewOpenAIProvider("key", "my-model", "")
	assert.Equal(t, "my-model", p.Model())
}

func TestOpenAIModelDefault(t *testing.T) {
	p := NewOpenAIProvider("key", "", "")
	assert.Equal(t, "gpt-4o-mini", p.Model())
}
