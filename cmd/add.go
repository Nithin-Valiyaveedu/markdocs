package cmd

import (
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <library>",
	Short: "Find, scrape, compile, and write a skill for a library.",
	Long: `Discovers documentation URLs for the given library using the configured LLM,
presents them for interactive selection, scrapes the pages, compiles the content
into a structured skill file, and writes it to .claude/skills/<category>/<library>.md.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	addCmd.Flags().Bool("no-interactive", false, "Skip URL selection prompt, use first suggested URL")
}
