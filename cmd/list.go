package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Nithin-Valiyaveedu/markdocs/internal/skill"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/ui"
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
		ui.Info("No skills compiled yet. Run 'markdocs add <library>' to get started.")
		return nil
	}

	tbl := ui.NewTable("NAME", "CATEGORY", "MODEL", "COMPILED", "SOURCES")

	shown := 0
	for _, f := range files {
		stale := skillIsStale(f.Meta.Compiled)
		if listStale && !stale {
			continue
		}

		age := skill.CompiledAge(f.Meta.Compiled)
		var ageStr string
		if stale {
			ageStr = ui.StyleWarning.Render(age)
		} else {
			ageStr = ui.StyleSuccess.Render(age)
		}

		name := f.Meta.Name
		if name == "" {
			name = skillNameFromPath(f.Path)
		}

		tbl.Row(
			name,
			f.Meta.Category,
			f.Meta.Model,
			ageStr,
			fmt.Sprintf("%d URL(s)", len(f.Meta.Sources)),
		)
		shown++
	}

	tbl.Print()

	if shown == 0 && listStale {
		ui.Info("All skills are up to date.")
	} else {
		ui.Blank()
		ui.Info(fmt.Sprintf("%d skill(s) found.", shown))
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
