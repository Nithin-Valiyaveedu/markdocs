package cmd

import (
	"fmt"
	"os"

	"github.com/Nithin-Valiyaveedu/markdocs/internal/llm"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/skill"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/ui"
	"github.com/spf13/cobra"
)

var addNoInteractive bool

var addCmd = &cobra.Command{
	Use:   "add <library>",
	Short: "Find, scrape, compile, and write a skill for a library.",
	Long: `Discovers documentation URLs for the given library using the configured LLM,
presents them for interactive selection, scrapes the pages, compiles the content
into a structured skill file, and writes it to .claude/skills/<category>/<library>.md.`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().BoolVar(&addNoInteractive, "no-interactive", false, "Skip URL selection prompt, use first suggested URL")
}

func runAdd(cmd *cobra.Command, args []string) error {
	library := args[0]
	ctx := cmd.Context()

	provider, err := llm.NewProvider(appConfig)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to create provider: %s", err))
		os.Exit(2)
	}

	// Step 1: Discover URLs
	var urls []string
	ui.Step(1, fmt.Sprintf("Asking LLM for %s documentation URLs...", library))
	err = ui.Spin("", func() error {
		compiler := skill.NewLLMCompiler(provider)
		var err error
		urls, err = compiler.SuggestURLs(ctx, library)
		return err
	})
	if err != nil {
		ui.Error(fmt.Sprintf("URL discovery failed: %s", err))
		os.Exit(2)
	}
	ui.StepDone(1, fmt.Sprintf("Found %d documentation URL(s)", len(urls)))

	// Step 2: Select URL
	var selectedURL string
	if addNoInteractive {
		selectedURL = urls[0]
		ui.Info(fmt.Sprintf("Using: %s", selectedURL))
	} else {
		selectedURL, err = ui.SelectWithManual("Select documentation URL to scrape", urls)
		if err != nil {
			return fmt.Errorf("url selection: %w", err)
		}
	}

	// Step 3–5: Scrape → Compile → Write (shared pipeline)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	result, err := runAddPipeline(ctx, library, selectedURL, provider, cwd)
	if err != nil {
		ui.Error(fmt.Sprintf("Pipeline failed: %s", err))
		os.Exit(2)
	}

	ui.Blank()
	ui.Success(fmt.Sprintf("Skill ready: %s", result.SkillPath))
	ui.Info("Claude Code will pick it up automatically next session.")
	return nil
}
