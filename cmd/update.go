package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nithin-Valiyaveedu/markdocs/internal/scraper"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/skill"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/ui"
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

	compiler, providerName, modelName := newCompiler()

	sc := scraper.NewWaterfall()
	updated := 0
	skipped := 0

	for _, f := range targets {
		name := f.Meta.Name
		if name == "" {
			name = strings.TrimSuffix(filepath.Base(f.Path), ".md")
		}

		ui.Section(name)

		if len(f.Meta.Sources) == 0 {
			ui.Warning("no sources recorded — skipping")
			skipped++
			continue
		}

		// Re-scrape the primary source URL
		primaryURL := f.Meta.Sources[0]
		var newContent string
		err := ui.Spin(fmt.Sprintf("Fetching %s...", primaryURL), func() error {
			var err error
			newContent, err = sc.Scrape(primaryURL)
			return err
		})
		if err != nil {
			ui.Error(fmt.Sprintf("fetch failed: %s — skipping", err))
			skipped++
			continue
		}

		newChecksum := skill.ContentChecksum(newContent)
		if newChecksum == f.Meta.Checksum {
			ui.Success("up to date")
			skipped++
			continue
		}

		ui.Recompiling("content changed — recompiling...")
		_, err = runAddPipeline(ctx, name, primaryURL, compiler, providerName, modelName, cwd, true)
		if err != nil {
			ui.Error(fmt.Sprintf("recompile failed: %s", err))
			skipped++
			continue
		}
		updated++
	}

	ui.Blank()
	ui.Success(fmt.Sprintf("Updated %d, skipped %d skill(s).", updated, skipped))
	return nil
}
