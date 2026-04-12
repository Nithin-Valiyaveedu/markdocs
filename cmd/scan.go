package cmd

import (
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Detect libraries in the project that lack skill files.",
	Long: `Reads package.json, go.mod, requirements.txt and other dependency files
to find libraries that don't have compiled skill files yet.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	scanCmd.Flags().Bool("add-all", false, "Automatically add skills for all missing libraries")
}
