package commands

import (
	"github.com/spf13/cobra"
)

var (
	modelFlag string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "gollm",
	Short: "A CLI for interacting with LLMs",
	Long: `A command-line interface for interacting with Large Language Models (LLMs).
Use it to chat, get completions, or stream responses from different LLM providers.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&modelFlag, "model", "m", "claude-3-sonnet-20240229", "LLM model to use")

	// Add commands
	rootCmd.AddCommand(completeCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(listModelsCmd)
}
