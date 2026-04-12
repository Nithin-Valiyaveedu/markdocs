package skill

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider is a test double for llm.LLMProvider.
type mockProvider struct {
	response string
	err      error
	Prompts  []string
	model    string
}

func (m *mockProvider) Complete(ctx context.Context, prompt string) (string, error) {
	m.Prompts = append(m.Prompts, prompt)
	return m.response, m.err
}

func (m *mockProvider) Model() string { return m.model }

func TestSuggestURLsParsing(t *testing.T) {
	provider := &mockProvider{
		response: `["https://react.dev/learn", "https://react.dev/reference"]`,
		model:    "test",
	}
	compiler := NewLLMCompiler(provider)
	urls, err := compiler.SuggestURLs(context.Background(), "react")
	require.NoError(t, err)
	assert.Len(t, urls, 2)
	assert.Contains(t, urls, "https://react.dev/learn")
}

func TestSuggestURLsWithCodeFence(t *testing.T) {
	provider := &mockProvider{
		response: "```json\n[\"https://react.dev\"]\n```",
		model:    "test",
	}
	compiler := NewLLMCompiler(provider)
	urls, err := compiler.SuggestURLs(context.Background(), "react")
	require.NoError(t, err)
	assert.Len(t, urls, 1)
}

func TestSuggestURLsMalformed(t *testing.T) {
	provider := &mockProvider{response: "Here are some links: react.dev", model: "test"}
	compiler := NewLLMCompiler(provider)
	_, err := compiler.SuggestURLs(context.Background(), "react")
	assert.Error(t, err)
}

func TestSuggestURLsLLMError(t *testing.T) {
	provider := &mockProvider{err: errors.New("timeout"), model: "test"}
	compiler := NewLLMCompiler(provider)
	_, err := compiler.SuggestURLs(context.Background(), "react")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "llm url discovery")
}

func TestCompileOutput(t *testing.T) {
	mockResponse := `# React — markdocs skill
## What This Is
React is a JavaScript library.
## Installation (project-specific)
npm install react
## Key Concepts
Components, JSX, Hooks.
## API / Usage Patterns
useState, useEffect.
## Your Project Config (detected)
Using next.js.
## Hidden Gotchas
None known.
## Common Errors
Missing key prop.
## Version Notes
React 18 adds concurrent features.
CATEGORY: frontend`

	provider := &mockProvider{response: mockResponse, model: "test-model"}
	compiler := NewLLMCompiler(provider)
	out, err := compiler.Compile(context.Background(), CompileInput{
		Library:          "React",
		URL:              "https://react.dev",
		ScrapedMarkdown:  "React docs content...",
		ProjectFramework: "next.js",
		Model:            "test-model",
	})
	require.NoError(t, err)
	assert.Equal(t, "frontend", out.Category)
	assert.Contains(t, out.Markdown, "# React — markdocs skill")
	assert.NotContains(t, out.Markdown, "CATEGORY:")
}

func TestChecksumConsistency(t *testing.T) {
	content := "some scraped documentation content"
	sum1 := ContentChecksum(content)
	sum2 := ContentChecksum(content)
	assert.Equal(t, sum1, sum2)
	assert.Len(t, sum1, 71) // "sha256:" prefix (7) + 64 hex chars
}
