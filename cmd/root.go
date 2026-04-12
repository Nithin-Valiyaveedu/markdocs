package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "markdocs",
	Short: "Compile library documentation into Claude Code skill files.",
	Long: `markdocs scrapes library documentation and compiles it into structured
Claude Code skill files (.claude/skills/<category>/<library>.md).

Configure a provider with 'markdocs init', then add skills with 'markdocs add <library>'.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(listCmd)
}
