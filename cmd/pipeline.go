package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Nithin-Valiyaveedu/markdocs/internal/llm"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/scraper"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/skill"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/ui"
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

// newCompiler returns the appropriate Compiler based on the --llm flag.
// In LLM mode, it requires appConfig to be loaded.
// providerName and modelName are set to identify the backend in skill frontmatter.
func newCompiler() (compiler skill.Compiler, providerName, modelName string) {
	if useLLM {
		provider, err := providerFromConfig()
		if err != nil {
			ui.Error(err.Error())
			os.Exit(2)
		}
		return skill.NewLLMCompiler(provider), string(appConfig.Provider), provider.Model()
	}
	return skill.NewStructuredExtractor(), "structured", ""
}

// PipelineResult holds the output of a successful runAddPipeline call.
type PipelineResult struct {
	SkillPath string
	Category  string
}

// runAddPipeline executes the full add pipeline: scrape → compile → [review] → write.
// It is called by `add`, `scan --add-all`, and `update`.
// providerName and modelName are stored in the skill file frontmatter for traceability.
// When noInteractive is false, a draft review prompt is shown before writing.
// Returns nil, nil if the user discards the draft.
func runAddPipeline(ctx context.Context, library, url string, compiler skill.Compiler, providerName, modelName string, cwd string, noInteractive bool) (*PipelineResult, error) {
	// Step 1: Scrape
	var scraped string
	err := ui.Spin(fmt.Sprintf("Scraping %s...", url), func() error {
		sc := scraper.NewWaterfall()
		var err error
		scraped, err = sc.Scrape(url)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("scraping %s: %w", url, err)
	}
	ui.StepDone(1, fmt.Sprintf("Scraped %d chars from %s", len(scraped), url))

	// Step 2: Compile
	var compiled *skill.CompileOutput
	err = ui.Spin("Compiling skill...", func() error {
		framework := skill.DetectFramework(cwd)
		var err error
		compiled, err = compiler.Compile(ctx, skill.CompileInput{
			Library:          library,
			URL:              url,
			ScrapedMarkdown:  scraped,
			ProjectFramework: framework,
			Model:            modelName,
		})
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("compiling skill for %s: %w", library, err)
	}
	ui.StepDone(2, fmt.Sprintf("Compiled skill (category: %s)", compiled.Category))

	// Step 3: Review draft (skipped in non-interactive mode)
	if !noInteractive {
		action, editedMarkdown, err := ui.ReviewDraft(compiled.Markdown)
		if err != nil {
			return nil, fmt.Errorf("draft review: %w", err)
		}
		if action == ui.ReviewDiscard {
			ui.Warning("Discarded — no skill written.")
			return nil, nil
		}
		if action == ui.ReviewEdit {
			compiled.Markdown = editedMarkdown
		}
	}

	// Step 4: Write — metadata lives entirely in the skill file's frontmatter
	framework := skill.DetectFramework(cwd)
	meta := skill.NewSkillMeta(
		library,
		providerName,
		modelName,
		compiled.Category,
		framework,
		compiled.Description,
		compiled.WhenToUse,
		[]string{url},
		scraped,
	)
	writer := skill.NewFSWriter(cwd)
	path, err := writer.Write(library, compiled.Category, compiled.Markdown, meta)
	if err != nil {
		return nil, fmt.Errorf("writing skill for %s: %w", library, err)
	}
	ui.StepDone(3, fmt.Sprintf("Written to %s", path))

	return &PipelineResult{SkillPath: path, Category: compiled.Category}, nil
}
