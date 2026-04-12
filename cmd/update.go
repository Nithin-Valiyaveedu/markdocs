package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/scraper"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/skill"
	"github.com/spf13/cobra"
)

var updateAll bool

var updateCmd = &cobra.Command{
	Use:   "update [skill]",
	Short: "Recompile skills whose source documentation has changed.",
	Long: `Checks the source URLs stored in each skill file's frontmatter and recompiles
any skill whose content checksum has changed since the last compile.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().BoolVar(&updateAll, "all", false, "Update all skills")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	if len(args) == 0 && !updateAll {
		return fmt.Errorf("specify a skill name or use --all")
	}

	// Collect target skill files
	allFiles, err := skill.GlobSkills(cwd)
	if err != nil {
		return fmt.Errorf("scanning skills: %w", err)
	}

	var targets []skill.SkillFile
	if updateAll {
		targets = allFiles
	} else {
		target := args[0]
		for _, f := range allFiles {
			name := f.Meta.Name
			if name == "" {
				name = strings.TrimSuffix(filepath.Base(f.Path), ".md")
			}
			if name == target {
				targets = append(targets, f)
				break
			}
		}
		if len(targets) == 0 {
			return fmt.Errorf("skill %q not found — run 'markdocs list' to see available skills", target)
		}
	}

	provider, err := providerFromConfig()
	if err != nil {
		return err
	}

	sc := scraper.NewWaterfall()
	updated := 0
	skipped := 0

	for _, f := range targets {
		name := f.Meta.Name
		if name == "" {
			name = strings.TrimSuffix(filepath.Base(f.Path), ".md")
		}

		fmt.Printf("Checking %s...\n", name)

		if len(f.Meta.Sources) == 0 {
			fmt.Printf("  %s no sources recorded — skipping\n", yellow("○"))
			skipped++
			continue
		}

		// Re-scrape the primary source URL
		primaryURL := f.Meta.Sources[0]
		s := spinner.New(spinner.CharSets[14], 80*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Fetching %s...", primaryURL)
		s.Start()

		newContent, err := sc.Scrape(primaryURL)
		s.Stop()
		if err != nil {
			fmt.Printf("  %s fetch failed: %s — skipping\n", red("✗"), err)
			skipped++
			continue
		}

		newChecksum := skill.ContentChecksum(newContent)
		if newChecksum == f.Meta.Checksum {
			fmt.Printf("  %s up to date\n", green("✓"))
			skipped++
			continue
		}

		fmt.Printf("  %s content changed — recompiling...\n", yellow("↻"))
		_, err = runAddPipeline(ctx, name, primaryURL, provider, cwd)
		if err != nil {
			fmt.Printf("  %s recompile failed: %s\n", red("✗"), err)
			skipped++
			continue
		}
		updated++
	}

	fmt.Printf("\n%s Updated %d, skipped %d skill(s).\n", green("✓"), updated, skipped)
	return nil
}
