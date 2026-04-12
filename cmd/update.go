package cmd

import (
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [skill]",
	Short: "Recompile skills whose source documentation has changed.",
	Long: `Checks the source URLs for compiled skills and recompiles any whose content
has changed since the last compile (detected via checksum comparison).`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	updateCmd.Flags().Bool("all", false, "Update all skills")
}
