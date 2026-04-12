package cmd

import (
	"fmt"
	"os"

	"github.com/nithinworks/markdocs/internal/config"
	"github.com/spf13/cobra"
)

// appConfig holds the loaded configuration, available to all subcommands.
var appConfig *config.Config

var rootCmd = &cobra.Command{
	Use:   "markdocs",
	Short: "Compile library documentation into Claude Code skill files.",
	Long: `markdocs scrapes library documentation and compiles it into structured
Claude Code skill files (.claude/skills/<category>/<library>.md).

Configure a provider with 'markdocs init', then add skills with 'markdocs add <library>'.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// init command creates the config, so skip loading it
		if cmd.Name() == "init" {
			return nil
		}
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("run 'markdocs init' first to configure a provider: %w", err)
		}
		appConfig = cfg
		return nil
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
