package cmd

import (
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Configure LLM provider and credentials.",
	Long:  "Interactively configure the LLM provider (Anthropic, OpenAI, or Ollama) used to find and compile documentation.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
