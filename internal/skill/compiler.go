package skill

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/Nithin-Valiyaveedu/markdocs/internal/llm"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/search"
)

const urlDiscoveryPrompt = `What are the official documentation URLs for "{{.Library}}"?
Return ONLY a JSON array of 3-5 URLs, most important first. No explanation, just the JSON array.
Example: ["https://docs.example.com", "https://example.com/api"]`

const skillCompilePrompt = `You are compiling documentation into a Claude Code skill file.

Library: {{.Library}}
URL: {{.URL}}
Project Framework: {{.ProjectFramework}}
Model: {{.Model}}

Scraped Documentation:
---
{{.ScrapedMarkdown}}
---

Generate a skill file with EXACTLY these sections, in this order:
# {{.Library}} — markdocs skill
## What This Is
## Installation (project-specific)
## Key Concepts
## API / Usage Patterns
## Your Project Config (detected)
## Hidden Gotchas
## Common Errors
## Version Notes

After the content, output these metadata lines (do not include them in the markdown body):
CATEGORY: <one of: frontend, backend, testing, infra, database, payments, auth, devtools>
DESCRIPTION: <one sentence describing what this library does and when Claude should load this skill. Max 200 characters. Front-load the library name and key use case.>
WHEN_TO_USE: <comma-separated trigger phrases or situations, e.g. "user mentions reactflow, building a node graph, flow diagram". Max 200 characters.>

Rules:
- Extract facts, not prose
- Flag anything undocumented but real (hidden gotchas)
- Note version-specific behaviour
- Use the project's detected patterns and config
- Keep it concise — this goes into an LLM context window
- Keep total content under 400 lines`

// CompileInput holds the inputs needed to compile a skill file.
type CompileInput struct {
	Library          string
	URL              string
	ScrapedMarkdown  string
	ProjectFramework string
	Model            string
}

// CompileOutput holds the result of a skill compilation.
type CompileOutput struct {
	Markdown    string
	Category    string
	Description string
	WhenToUse   string
}

// Compiler discovers doc URLs and compiles documentation into skill files.
type Compiler interface {
	SuggestURLs(ctx context.Context, library string) ([]string, error)
	Compile(ctx context.Context, input CompileInput) (*CompileOutput, error)
}

// LLMCompiler implements Compiler using an LLM provider.
type LLMCompiler struct {
	provider llm.LLMProvider
	// searchFn is the web search backend. Defaults to search.DocURLs.
	// Override in tests to avoid real network calls.
	searchFn func(library string, max int) ([]string, error)
}

var _ Compiler = (*LLMCompiler)(nil)

// NewLLMCompiler creates a new LLMCompiler using the given provider.
func NewLLMCompiler(provider llm.LLMProvider) *LLMCompiler {
	return &LLMCompiler{provider: provider, searchFn: search.DocURLs}
}

// SuggestURLs discovers documentation URLs for the given library.
// It searches the web first (DuckDuckGo) and validates reachability, falling
// back to asking the LLM if web search yields nothing usable.
func (c *LLMCompiler) SuggestURLs(ctx context.Context, library string) ([]string, error) {
	// Step 1: web search
	webURLs, err := c.searchFn(library, 8)
	if err == nil && len(webURLs) > 0 {
		validated := search.ValidateURLs(webURLs)
		if len(validated) > 0 {
			// Return top 5 validated results
			if len(validated) > 5 {
				validated = validated[:5]
			}
			return validated, nil
		}
	}

	// Step 2: fall back to LLM
	prompt, err := renderTemplate(urlDiscoveryPrompt, map[string]string{"Library": library})
	if err != nil {
		return nil, fmt.Errorf("rendering url discovery prompt: %w", err)
	}

	response, err := c.provider.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("llm url discovery: %w", err)
	}

	urls, err := parseJSONStringArray(response)
	if err != nil {
		return nil, fmt.Errorf("parsing url response: %w", err)
	}
	if len(urls) == 0 {
		return nil, fmt.Errorf("no URLs returned for library %q", library)
	}

	// Validate LLM-suggested URLs too
	validated := search.ValidateURLs(urls)
	if len(validated) > 0 {
		return validated, nil
	}
	return urls, nil
}

// Compile uses the LLM to compile scraped documentation into a structured skill file.
func (c *LLMCompiler) Compile(ctx context.Context, input CompileInput) (*CompileOutput, error) {
	prompt, err := renderTemplate(skillCompilePrompt, input)
	if err != nil {
		return nil, fmt.Errorf("rendering compile prompt: %w", err)
	}

	response, err := c.provider.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("llm compile: %w", err)
	}

	return parseCompileResponse(response), nil
}

// parseCompileResponse extracts markdown content and metadata lines from the LLM response.
func parseCompileResponse(response string) *CompileOutput {
	lines := strings.Split(response, "\n")

	var contentLines []string
	out := &CompileOutput{Category: "general"}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "CATEGORY:"):
			if v := strings.TrimSpace(strings.TrimPrefix(trimmed, "CATEGORY:")); v != "" {
				out.Category = strings.ToLower(v)
			}
		case strings.HasPrefix(trimmed, "DESCRIPTION:"):
			out.Description = strings.TrimSpace(strings.TrimPrefix(trimmed, "DESCRIPTION:"))
		case strings.HasPrefix(trimmed, "WHEN_TO_USE:"):
			out.WhenToUse = strings.TrimSpace(strings.TrimPrefix(trimmed, "WHEN_TO_USE:"))
		default:
			contentLines = append(contentLines, line)
		}
	}

	out.Markdown = strings.TrimSpace(strings.Join(contentLines, "\n"))
	return out
}

// parseJSONStringArray extracts a JSON string array from an LLM response,
// stripping any markdown code fences first.
func parseJSONStringArray(response string) ([]string, error) {
	// Strip markdown code fences
	cleaned := response
	if idx := strings.Index(cleaned, "```"); idx != -1 {
		cleaned = cleaned[idx:]
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		if end := strings.Index(cleaned, "```"); end != -1 {
			cleaned = cleaned[:end]
		}
	}

	// Find the JSON array bounds
	start := strings.Index(cleaned, "[")
	end := strings.LastIndex(cleaned, "]")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON array found in response: %q", response)
	}
	jsonStr := cleaned[start : end+1]

	var urls []string
	if err := json.Unmarshal([]byte(jsonStr), &urls); err != nil {
		return nil, fmt.Errorf("parsing JSON array %q: %w", jsonStr, err)
	}
	return urls, nil
}

// renderTemplate renders a text template with the given data.
func renderTemplate(tmplStr string, data interface{}) (string, error) {
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
