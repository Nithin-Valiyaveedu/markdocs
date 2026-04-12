package cmd

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all compiled skills and their status.",
	Long:  "Lists all skill files compiled for the current project, with their compilation date, model used, and freshness status.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	listCmd.Flags().Bool("stale", false, "Show only skills whose source has changed")
}
