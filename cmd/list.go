package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/nithinworks/markdocs/internal/skill"
	"github.com/spf13/cobra"
)

var listStale bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all compiled skills and their status.",
	Long:  "Lists all skill files compiled for the current project by scanning .claude/skills/**/*.md.",
	RunE:  runList,
}

func init() {
	listCmd.Flags().BoolVar(&listStale, "stale", false, "Show only skills compiled more than 7 days ago")
}

func runList(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	files, err := skill.GlobSkills(cwd)
	if err != nil {
		return fmt.Errorf("scanning skills: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No skills compiled yet. Run 'markdocs add <library>' to get started.")
		return nil
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tCATEGORY\tMODEL\tCOMPILED\tSOURCES")
	fmt.Fprintln(w, "----\t--------\t-----\t--------\t-------")

	shown := 0
	for _, f := range files {
		stale := skillIsStale(f.Meta.Compiled)
		if listStale && !stale {
			continue
		}

		age := skill.CompiledAge(f.Meta.Compiled)
		ageStr := age
		if stale {
			ageStr = yellow(age)
		} else {
			ageStr = green(age)
		}

		name := f.Meta.Name
		if name == "" {
			name = skillNameFromPath(f.Path)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d URL(s)\n",
			name,
			f.Meta.Category,
			f.Meta.Model,
			ageStr,
			len(f.Meta.Sources),
		)
		shown++
	}
	w.Flush()

	if shown == 0 && listStale {
		fmt.Println("All skills are up to date.")
	} else {
		fmt.Printf("\n%d skill(s) found.\n", shown)
	}
	return nil
}

// skillIsStale returns true if the skill was compiled more than 7 days ago.
func skillIsStale(compiled string) bool {
	t, err := time.Parse(time.RFC3339, compiled)
	if err != nil {
		return true
	}
	return time.Since(t) > 7*24*time.Hour
}

// skillNameFromPath extracts the filename without extension as a fallback name.
func skillNameFromPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, ".md")
}
