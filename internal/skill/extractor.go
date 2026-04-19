package skill

import (
	"context"
	"fmt"
	"strings"

	"github.com/Nithin-Valiyaveedu/markdocs/internal/resolver"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/search"
)

// sectionEntry pairs a skill section key with its heading keyword patterns.
type sectionEntry struct {
	key      string
	keywords []string
}

// sectionMappings is an ordered list of section → keyword mappings.
// Order matters: more specific sections are checked first so that broad keywords
// like "guide" or "overview" don't shadow more precise matches (e.g. "migration guide").
var sectionMappings = []sectionEntry{
	{"versions", []string{"changelog", "migration", "upgrade", "breaking", "release note", "what's new", "history"}},
	{"installation", []string{"install", "getting started", "quick start", "quickstart", "setup", "prerequisites", "requirement"}},
	{"errors", []string{"error", "troubleshoot", "common issue", "faq", "problem", "debug", "fix"}},
	{"gotchas", []string{"gotcha", "pitfall", "caveat", "warning", "limitation", "caveats", "known issue"}},
	{"config", []string{"config", "configuration", "options", "settings", "environment", "customize", "parameter"}},
	{"api", []string{"api", "reference", "methods", "functions", "interface", "endpoints", "usage", "example", "hook"}},
	// concepts last — "guide", "overview", "introduction" are broad and should only match
	// when no more specific section has matched.
	{"concepts", []string{"concept", "overview", "architecture", "how it works", "introduction", "guide", "basics", "core", "fundamental"}},
}

// StructuredExtractor implements Compiler using deterministic heading-based
// extraction from scraped documentation — no LLM required.
type StructuredExtractor struct {
	searchFn  func(library string, max int) ([]string, error)
	resolveFn func(ctx context.Context, library string) ([]string, error)
}

var _ Compiler = (*StructuredExtractor)(nil)

// NewStructuredExtractor creates a StructuredExtractor using DuckDuckGo web search
// and package registry resolution.
func NewStructuredExtractor() *StructuredExtractor {
	r := resolver.New()
	return &StructuredExtractor{
		searchFn:  search.DocURLs,
		resolveFn: r.Resolve,
	}
}

// SuggestURLs discovers documentation URLs using a three-layer approach:
//  1. Package registry APIs (npm, PyPI, crates.io, pkg.go.dev) — authoritative
//  2. markdocs curated registry DocHints — verified overrides
//  3. DuckDuckGo web search — last resort
func (e *StructuredExtractor) SuggestURLs(ctx context.Context, library string) ([]string, error) {
	// Layer 1: Package registry — maintainer-declared docs URL
	if regURLs, err := e.resolveFn(ctx, library); err == nil && len(regURLs) > 0 {
		validated := search.ValidateURLs(regURLs)
		if len(validated) > 0 {
			if len(validated) > 5 {
				validated = validated[:5]
			}
			return validated, nil
		}
	}

	// Layer 2: Curated DocHints registry
	if lib := RegistryByName(sanitizeName(library)); lib != nil && len(lib.DocHints) > 0 {
		return lib.DocHints, nil
	}

	// Layer 3: DuckDuckGo web search
	urls, searchErr := e.searchFn(library, 8)
	if searchErr == nil && len(urls) > 0 {
		validated := search.ValidateURLs(urls)
		if len(validated) > 0 {
			if len(validated) > 5 {
				validated = validated[:5]
			}
			return validated, nil
		}
	}

	if searchErr != nil {
		return nil, fmt.Errorf("web search for %q: %w", library, searchErr)
	}
	return nil, fmt.Errorf("no documentation URLs found for %q", library)
}

// Compile extracts skill sections from scraped markdown using heading pattern matching.
// It never calls an LLM. If the library's category cannot be detected from the registry,
// CompileOutput.Category will be empty — callers should prompt the user for it.
func (e *StructuredExtractor) Compile(_ context.Context, input CompileInput) (*CompileOutput, error) {
	sections := extractSections(input.ScrapedMarkdown)

	whatThisIs := firstParagraph(input.ScrapedMarkdown)

	// Category from registry (exact then prefix match); fall back to "general"
	category := CategoryForLibrary(input.Library)
	if category == "" {
		category = "general"
	}

	description := whatThisIs
	if len(description) > 200 {
		description = description[:197] + "..."
	}

	whenToUse := fmt.Sprintf("user mentions %s, working with %s libraries", input.Library, input.Library)
	if category != "" {
		whenToUse = fmt.Sprintf("user mentions %s, working with %s libraries", input.Library, category)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s — markdocs skill\n\n", input.Library))

	sb.WriteString("## What This Is\n")
	if whatThisIs != "" {
		sb.WriteString(whatThisIs + "\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("%s library documentation.\n\n", input.Library))
	}

	writeSkillSection(&sb, "## Installation (project-specific)", sections["installation"])
	writeSkillSection(&sb, "## Key Concepts", sections["concepts"])
	writeSkillSection(&sb, "## API / Usage Patterns", sections["api"])
	writeSkillSection(&sb, "## Your Project Config (detected)", sections["config"])
	writeSkillSection(&sb, "## Hidden Gotchas", sections["gotchas"])
	writeSkillSection(&sb, "## Common Errors", sections["errors"])
	writeSkillSection(&sb, "## Version Notes", sections["versions"])

	return &CompileOutput{
		Markdown:    strings.TrimSpace(sb.String()),
		Category:    category,
		Description: description,
		WhenToUse:   whenToUse,
	}, nil
}

// extractSections parses scraped markdown into skill section buckets by heading matching.
// Content under unrecognised headings is discarded.
func extractSections(markdown string) map[string]string {
	lines := strings.Split(markdown, "\n")
	sections := make(map[string]string)
	currentKey := ""
	var currentLines []string

	flush := func() {
		if currentKey == "" || len(currentLines) == 0 {
			return
		}
		content := strings.TrimSpace(strings.Join(currentLines, "\n"))
		if content == "" {
			return
		}
		if existing := sections[currentKey]; existing != "" {
			sections[currentKey] = existing + "\n\n" + content
		} else {
			sections[currentKey] = content
		}
	}

	for _, line := range lines {
		if isHeading(line) {
			newKey := matchSection(line)
			if newKey != "" {
				// Matched a skill section: flush previous and start new
				flush()
				currentLines = nil
				currentKey = newKey
			} else if currentKey != "" {
				// Unmatched sub-heading: include as content in the current section
				currentLines = append(currentLines, line)
			}
		} else if currentKey != "" {
			currentLines = append(currentLines, line)
		}
	}
	flush()

	return sections
}

// matchSection returns the skill section key for the given heading line, or "" if none match.
// Sections are checked in priority order (most specific first).
func matchSection(heading string) string {
	lower := strings.ToLower(strings.TrimLeft(heading, "# "))
	for _, entry := range sectionMappings {
		for _, kw := range entry.keywords {
			if strings.Contains(lower, kw) {
				return entry.key
			}
		}
	}
	return ""
}

func isHeading(line string) bool {
	return strings.HasPrefix(line, "#")
}

// firstParagraph returns the first non-heading, non-empty paragraph of the markdown.
func firstParagraph(markdown string) string {
	lines := strings.Split(markdown, "\n")
	var para []string
	inPara := false

	for _, line := range lines {
		if isHeading(line) {
			if inPara {
				break
			}
			continue
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if inPara {
				break
			}
			continue
		}
		para = append(para, trimmed)
		inPara = true
	}

	result := strings.Join(para, " ")
	// Strip common markdown inline formatting that would look odd as a description
	result = strings.ReplaceAll(result, "**", "")
	result = strings.ReplaceAll(result, "__", "")
	return strings.TrimSpace(result)
}

// writeSkillSection writes a section heading and its content (or a placeholder).
func writeSkillSection(sb *strings.Builder, header, content string) {
	sb.WriteString(header + "\n")
	if content != "" {
		sb.WriteString(content + "\n\n")
	} else {
		sb.WriteString("_No information found in documentation._\n\n")
	}
}
