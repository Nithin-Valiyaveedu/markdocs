package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/nithinworks/markdocs/internal/llm"
	"github.com/nithinworks/markdocs/internal/skill"
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
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	provider, err := llm.NewProvider(appConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to create provider: %s\n", red("✗"), err)
		os.Exit(2)
	}

	// Step 1: Discover URLs
	s := spinner.New(spinner.CharSets[14], 80*time.Millisecond)
	s.Suffix = fmt.Sprintf(" Asking LLM for %s documentation URLs...", library)
	s.Start()

	compiler := skill.NewLLMCompiler(provider)
	urls, err := compiler.SuggestURLs(ctx, library)
	s.Stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s URL discovery failed: %s\n", red("✗"), err)
		os.Exit(2)
	}
	fmt.Printf("%s Found %d documentation URL(s)\n", green("✓"), len(urls))

	// Step 2: Select URL
	var selectedURL string
	if addNoInteractive {
		selectedURL = urls[0]
		fmt.Printf("  Using: %s\n", selectedURL)
	} else {
		selectedURL, err = selectURL(urls)
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
		fmt.Fprintf(os.Stderr, "%s Pipeline failed: %s\n", red("✗"), err)
		os.Exit(2)
	}

	fmt.Printf("\n%s Skill ready: %s\n", green("✓"), result.SkillPath)
	fmt.Println("  Claude Code will pick it up automatically next session.")
	return nil
}

// selectURL presents a URL selection prompt with an option to enter manually.
func selectURL(urls []string) (string, error) {
	items := append(urls, "↵ Enter URL manually")

	sel := promptui.Select{
		Label: "Select documentation URL to scrape",
		Items: items,
		Size:  8,
	}
	idx, _, err := sel.Run()
	if err != nil {
		return "", err
	}

	if idx == len(urls) {
		// Manual entry
		p := promptui.Prompt{Label: "Documentation URL"}
		return p.Run()
	}
	return urls[idx], nil
}
