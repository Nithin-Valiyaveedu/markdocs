package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/nithinworks/markdocs/internal/llm"
	"github.com/nithinworks/markdocs/internal/scraper"
	"github.com/nithinworks/markdocs/internal/skill"
)

// providerFromConfig constructs the LLM provider from the loaded appConfig.
func providerFromConfig() (llm.LLMProvider, error) {
	if appConfig == nil {
		return nil, fmt.Errorf("no provider configured — run 'markdocs init' first")
	}
	provider, err := llm.NewProvider(appConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ Failed to create provider: %s\n", err)
		os.Exit(2)
	}
	return provider, nil
}

// PipelineResult holds the output of a successful runAddPipeline call.
type PipelineResult struct {
	SkillPath string
	Category  string
}

// runAddPipeline executes the full add pipeline: scrape → compile → write.
// It is called by both `add` and `scan --add-all`.
func runAddPipeline(ctx context.Context, library, url string, provider llm.LLMProvider, cwd string) (*PipelineResult, error) {
	green := color.New(color.FgGreen).SprintFunc()

	// Step 1: Scrape
	s := spinner.New(spinner.CharSets[14], 80*time.Millisecond)
	s.Suffix = fmt.Sprintf(" Scraping %s...", url)
	s.Start()

	sc := scraper.NewWaterfall()
	scraped, err := sc.Scrape(url)
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("scraping %s: %w", url, err)
	}
	fmt.Printf("  %s Scraped %d chars from %s\n", green("✓"), len(scraped), url)

	// Step 2: Compile
	s = spinner.New(spinner.CharSets[14], 80*time.Millisecond)
	s.Suffix = " Compiling skill..."
	s.Start()

	framework := skill.DetectFramework(cwd)
	compiler := skill.NewLLMCompiler(provider)
	compiled, err := compiler.Compile(ctx, skill.CompileInput{
		Library:          library,
		URL:              url,
		ScrapedMarkdown:  scraped,
		ProjectFramework: framework,
		Model:            provider.Model(),
	})
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("compiling skill for %s: %w", library, err)
	}
	fmt.Printf("  %s Compiled skill (category: %s)\n", green("✓"), compiled.Category)

	// Step 3: Write — metadata lives entirely in the skill file's frontmatter
	meta := skill.NewSkillMeta(
		library,
		string(appConfig.Provider),
		provider.Model(),
		compiled.Category,
		framework,
		[]string{url},
		scraped,
	)
	writer := skill.NewFSWriter(cwd)
	path, err := writer.Write(library, compiled.Category, compiled.Markdown, meta)
	if err != nil {
		return nil, fmt.Errorf("writing skill for %s: %w", library, err)
	}
	fmt.Printf("  %s Written to %s\n", green("✓"), path)

	return &PipelineResult{SkillPath: path, Category: compiled.Category}, nil
}
