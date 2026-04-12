package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/nithinworks/markdocs/internal/skill"
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

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Scan project dependencies
	depSets, err := skill.ScanProject(cwd)
	if err != nil {
		return fmt.Errorf("scanning project: %w", err)
	}

	if len(depSets) == 0 {
		fmt.Println("No dependency files found (package.json, go.mod, requirements.txt).")
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
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "LIBRARY\tSOURCE\tSTATUS")
	fmt.Fprintln(w, "-------\t------\t------")

	var missing []libEntry
	for _, lib := range allLibs {
		if skill.SkillExists(cwd, lib.name) {
			fmt.Fprintf(w, "%s\t%s\t%s\n", lib.name, lib.source, green("✓ compiled"))
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\n", lib.name, lib.source, yellow("○ missing"))
			missing = append(missing, lib)
		}
	}
	w.Flush()

	fmt.Printf("\n%d libraries found, %d compiled, %d missing\n",
		len(allLibs), len(allLibs)-len(missing), len(missing))

	if len(missing) == 0 || !scanAddAll {
		return nil
	}

	// --add-all: run pipeline for each missing library
	if appConfig == nil {
		return fmt.Errorf("no provider configured — run 'markdocs init' first")
	}
	provider, err := providerFromConfig()
	if err != nil {
		return err
	}

	fmt.Printf("\nAdding %d missing skills...\n", len(missing))
	added := 0
	for _, lib := range missing {
		fmt.Printf("\n[%s]\n", lib.name)

		// Use LLM to discover URL, pick first one automatically
		compiler := skill.NewLLMCompiler(provider)
		urls, err := compiler.SuggestURLs(ctx, lib.name)
		if err != nil {
			fmt.Printf("  %s skipping %s: %s\n", red("✗"), lib.name, err)
			continue
		}
		if len(urls) == 0 {
			fmt.Printf("  %s no URLs found for %s\n", red("✗"), lib.name)
			continue
		}

		_, err = runAddPipeline(ctx, lib.name, urls[0], provider, cwd)
		if err != nil {
			fmt.Printf("  %s failed: %s\n", red("✗"), err)
			continue
		}
		added++
	}

	fmt.Printf("\n%s Added %d skill(s).\n", green("✓"), added)
	return nil
}
