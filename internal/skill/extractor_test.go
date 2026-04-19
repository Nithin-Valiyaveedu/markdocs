package skill

import (
	"context"
	"strings"
	"testing"
)

const sampleDocs = `# Zustand

Zustand is a small, fast state management library for React.

## Installation

Install via npm:

` + "```bash" + `
npm install zustand
` + "```" + `

Or with yarn:

` + "```bash" + `
yarn add zustand
` + "```" + `

## Core Concepts

Zustand uses a store-based approach where state is managed outside React components.
Stores are created with the create function and accessed via hooks.

## API Reference

### create

` + "```ts" + `
const useStore = create((set) => ({
  count: 0,
  increment: () => set((state) => ({ count: state.count + 1 })),
}))
` + "```" + `

## Configuration Options

You can configure middleware like devtools and immer:

` + "```ts" + `
const useStore = create(devtools(immer((set) => ({ ... }))))
` + "```" + `

## Common Errors

### Stale closures

If you read state inside an event handler without using getState(), you may get stale values.

## Changelog

### v5.0

Breaking: removed deprecated create() overload.
`

func TestExtractSections(t *testing.T) {
	sections := extractSections(sampleDocs)

	if sections["installation"] == "" {
		t.Error("expected installation section to be populated")
	}
	if !strings.Contains(sections["installation"], "npm install zustand") {
		t.Errorf("installation section missing npm command, got: %s", sections["installation"])
	}

	if sections["concepts"] == "" {
		t.Error("expected concepts section to be populated")
	}
	if !strings.Contains(sections["concepts"], "store-based") {
		t.Errorf("concepts section missing expected content, got: %s", sections["concepts"])
	}

	if sections["api"] == "" {
		t.Error("expected api section to be populated")
	}
	if !strings.Contains(sections["api"], "create") {
		t.Errorf("api section missing expected content, got: %s", sections["api"])
	}

	if sections["config"] == "" {
		t.Error("expected config section to be populated")
	}

	if sections["errors"] == "" {
		t.Error("expected errors section to be populated")
	}
	if !strings.Contains(sections["errors"], "Stale closures") {
		t.Errorf("errors section missing expected content, got: %s", sections["errors"])
	}

	if sections["versions"] == "" {
		t.Error("expected versions section to be populated")
	}
	if !strings.Contains(sections["versions"], "v5.0") {
		t.Errorf("versions section missing expected content, got: %s", sections["versions"])
	}
}

func TestFirstParagraph(t *testing.T) {
	result := firstParagraph(sampleDocs)
	if result == "" {
		t.Fatal("expected non-empty first paragraph")
	}
	if !strings.Contains(result, "Zustand") {
		t.Errorf("expected first paragraph to mention Zustand, got: %s", result)
	}
	// Should not include heading text
	if strings.HasPrefix(result, "#") {
		t.Errorf("first paragraph should not start with #, got: %s", result)
	}
}

func TestMatchSection(t *testing.T) {
	cases := []struct {
		heading string
		want    string
	}{
		{"## Installation", "installation"},
		{"## Getting Started", "installation"},
		{"## API Reference", "api"},
		{"## Core Concepts", "concepts"},
		{"## Configuration", "config"},
		{"## Common Errors", "errors"},
		{"## Changelog", "versions"},
		{"## Migration Guide", "versions"},
		{"## About This Library", ""},
	}

	for _, tc := range cases {
		got := matchSection(tc.heading)
		if got != tc.want {
			t.Errorf("matchSection(%q) = %q, want %q", tc.heading, got, tc.want)
		}
	}
}

func TestStructuredExtractorCompile(t *testing.T) {
	e := &StructuredExtractor{
		searchFn:  func(library string, max int) ([]string, error) { return nil, nil },
		resolveFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
	}

	out, err := e.Compile(context.Background(), CompileInput{
		Library:         "zustand",
		URL:             "https://zustand.docs.pmnd.rs/",
		ScrapedMarkdown: sampleDocs,
	})
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	if out.Category != "frontend" {
		t.Errorf("expected category 'frontend' (zustand is in registry), got %q", out.Category)
	}
	if out.Description == "" {
		t.Error("expected non-empty description")
	}
	if out.WhenToUse == "" {
		t.Error("expected non-empty when_to_use")
	}
	if !strings.Contains(out.Markdown, "## What This Is") {
		t.Error("expected markdown to contain '## What This Is'")
	}
	if !strings.Contains(out.Markdown, "## Installation") {
		t.Error("expected markdown to contain '## Installation'")
	}
	if !strings.Contains(out.Markdown, "## API / Usage Patterns") {
		t.Error("expected markdown to contain '## API / Usage Patterns'")
	}
}

func TestStructuredExtractorCompileUnknownLibrary(t *testing.T) {
	e := &StructuredExtractor{
		searchFn:  func(library string, max int) ([]string, error) { return nil, nil },
		resolveFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
	}

	out, err := e.Compile(context.Background(), CompileInput{
		Library:         "some-unknown-lib",
		URL:             "https://example.com/docs",
		ScrapedMarkdown: "# some-unknown-lib\n\nA library for doing things.\n\n## Installation\n\nnpm install some-unknown-lib\n",
	})
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}
	// Should default to "general" when not in registry
	if out.Category != "general" {
		t.Errorf("expected category 'general' for unknown library, got %q", out.Category)
	}
}

func TestCategoryForLibrary(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"react", "frontend"},
		{"react-query", "frontend"},  // prefix match: "react" prefix
		{"zustand", "frontend"},
		{"prisma", "database"},
		{"jest", "testing"},
		{"stripe", "payments"},
		{"completely-unknown-lib", ""},
	}

	for _, tc := range cases {
		got := CategoryForLibrary(tc.name)
		if got != tc.want {
			t.Errorf("CategoryForLibrary(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}
