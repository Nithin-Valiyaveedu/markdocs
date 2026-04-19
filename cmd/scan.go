package cmd

import (
	"fmt"
	"os"

	"github.com/Nithin-Valiyaveedu/markdocs/internal/skill"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/ui"
	"github.com/spf13/cobra"
)

var scanAddAll bool

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Detect libraries in the project that lack skill files.",
	Long: `Reads package.json, go.mod, requirements.txt and other dependency files
to find libraries that don't have compiled skill files yet.`,
	RunE: runScan,
}

func init() {
	scanCmd.Flags().BoolVar(&scanAddAll, "add-all", false, "Automatically add skills for all missing libraries")
}

func runScan(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	// Scan project dependencies
	depSets, err := skill.ScanProject(cwd)
	if err != nil {
		return fmt.Errorf("scanning project: %w", err)
	}

	if len(depSets) == 0 {
		ui.Info("No dependency files found (package.json, go.mod, requirements.txt).")
		return nil
	}

	// Build a flat list of all libraries across all dep files
	type libEntry struct {
		name   string
		source string
	}
	var allLibs []libEntry
	seen := make(map[string]bool)
	for _, deps := range depSets {
		for _, lib := range deps.Libraries {
			if !seen[lib] {
				seen[lib] = true
				allLibs = append(allLibs, libEntry{name: lib, source: deps.Source})
			}
		}
	}

	// Print table
	tbl := ui.NewTable("LIBRARY", "SOURCE", "STATUS")
	var missing []libEntry
	for _, lib := range allLibs {
		if skill.SkillExists(cwd, lib.name) {
			tbl.Row(lib.name, lib.source, ui.StyleSuccess.Render("✓ compiled"))
		} else {
			tbl.Row(lib.name, lib.source, ui.StyleWarning.Render("○ missing"))
			missing = append(missing, lib)
		}
	}
	tbl.Print()

	ui.Blank()
	ui.Info(fmt.Sprintf("%d libraries found, %d compiled, %d missing",
		len(allLibs), len(allLibs)-len(missing), len(missing)))

	if len(missing) == 0 || !scanAddAll {
		return nil
	}

	// --add-all: run pipeline for each missing library
	compiler, providerName, modelName := newCompiler()

	ui.Blank()
	ui.Section(fmt.Sprintf("Adding %d missing skills", len(missing)))

	added := 0
	for _, lib := range missing {
		ui.Blank()
		ui.Section(lib.name)

		urls, err := compiler.SuggestURLs(ctx, lib.name)
		if err != nil {
			ui.Error(fmt.Sprintf("skipping %s: %s", lib.name, err))
			continue
		}
		if len(urls) == 0 {
			ui.Error(fmt.Sprintf("no URLs found for %s", lib.name))
			continue
		}

		_, err = runAddPipeline(ctx, lib.name, urls[0], compiler, providerName, modelName, cwd)
		if err != nil {
			ui.Error(fmt.Sprintf("failed: %s", err))
			continue
		}
		added++
	}

	ui.Blank()
	ui.Success(fmt.Sprintf("Added %d skill(s).", added))
	return nil
}
