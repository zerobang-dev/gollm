package commands

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/zerobang-dev/gollm/pkg/llm"
)

var (
	modelFlag        string
	systemPromptFlag string
	temperatureFlag  float64
	queryAllFlag     bool
	verboseFlag      bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "gollm [prompt]",
	Short: "A CLI for interacting with LLMs",
	Long: `A command-line interface for interacting with Large Language Models (LLMs).
	Use it to chat, get completions, or stream responses from different LLM providers.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read prompt from args or stdin
		prompt, err := readPromptFromArgs(cmd, args)
		if err != nil {
			return err
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		// Query the LLM
		result, err := queryLLM(ctx, prompt, modelFlag, systemPromptFlag, temperatureFlag, queryAllFlag)
		if err != nil {
			return err
		}

		// Display results based on the query type
		if queryAllFlag {
			results, ok := result.(map[string]llm.ProviderResponse)
			if !ok {
				return nil
			}
			return displayProviderResults(results)
		} else {
			response, ok := result.(*struct {
				Response    string
				ElapsedTime time.Duration
			})
			if !ok {
				return nil
			}

			if verboseFlag {
				return displayVerboseResult(prompt, modelFlag, response.Response, response.ElapsedTime)
			} else {
				displaySimpleResult(response.Response, response.ElapsedTime)
				return nil
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&modelFlag, "model", "m", "claude-3-7-sonnet-latest", "LLM model to use")

	// Add flags to the root command
	rootCmd.Flags().StringVarP(&systemPromptFlag, "system", "s", "", "System prompt to provide context")
	rootCmd.Flags().Float64VarP(&temperatureFlag, "temperature", "t", 0.7, "Temperature for response generation (0.0-1.0)")
	rootCmd.Flags().BoolVarP(&queryAllFlag, "all", "a", false, "Query all configured providers and compare responses")
	rootCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Display detailed response information in a colorful table")

	// Add commands
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(listModelsCmd)
}
