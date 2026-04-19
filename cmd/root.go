package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/Nithin-Valiyaveedu/markdocs/internal/config"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/ui"
	"github.com/spf13/cobra"
)

var Version = resolveVersion()

func resolveVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return "dev"
}

// appConfig holds the loaded configuration, available to all subcommands.
// It is nil when running without --llm (structured extraction mode).
var appConfig *config.Config

// useLLM controls whether commands use LLM compilation (requires a configured provider).
// When false (default), structured heading-based extraction is used and no API key is needed.
var useLLM bool

var rootCmd = &cobra.Command{
	Use:   "markdocs",
	Short: "Compile library documentation into Claude Code skill files.",
	Long: `markdocs scrapes library documentation and compiles it into structured
Claude Code skill files (.claude/skills/<category>/<library>.md).

By default, markdocs uses structured extraction — no API key required.
Pass --llm to use an LLM provider instead (configure with 'markdocs init').`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ui.PrintBanner(Version)
		ui.Blank()
		// init creates the config, so skip loading it
		if cmd.Name() == "init" {
			return nil
		}
		// Config is only required when --llm is explicitly set
		if !useLLM {
			return nil
		}
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("--llm requires a configured provider — run 'markdocs init' first: %w", err)
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
	rootCmd.PersistentFlags().BoolVar(&useLLM, "llm", false, "Use LLM compilation instead of structured extraction (requires 'markdocs init')")
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(listCmd)
}
