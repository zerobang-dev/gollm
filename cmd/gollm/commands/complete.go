package commands

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/zerobang-dev/gollm/pkg/config"
	"github.com/zerobang-dev/gollm/pkg/llm"
)

var (
	systemPromptFlag string
	temperatureFlag  float64
	queryAllFlag     bool
	verboseFlag      bool
)

// completeCmd represents the complete command
var completeCmd = &cobra.Command{
	Use:   "complete [prompt]",
	Short: "Get a completion from an LLM",
	Long:  `Send a prompt to an LLM and get a completion response. If no prompt is provided, reads from stdin.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var prompt string
		if len(args) == 1 {
			prompt = args[0]
		} else {
			// Read from stdin
			buffer := make([]byte, 0, 4096)
			temp := make([]byte, 1024)
			stdin := cmd.InOrStdin()

			for {
				n, err := stdin.Read(temp)
				if n > 0 {
					buffer = append(buffer, temp[:n]...)
				}
				if err != nil || n < len(temp) {
					break
				}
			}

			if len(buffer) == 0 {
				return fmt.Errorf("no input provided via argument or stdin")
			}

			prompt = string(buffer)
		}

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		// Create HTTP client with timeout
		httpClient := &http.Client{
			Timeout: 120 * time.Second,
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		// Set up options
		options := []llm.Option{
			llm.WithMaxTokens(1000),
			llm.WithTemperature(temperatureFlag),
		}

		// If system prompt is provided, add it as a custom parameter
		if systemPromptFlag != "" {
			options = append(options, llm.WithCustomParam("system", systemPromptFlag))
		}

		// Initialize API keys map
		apiKeys := make(map[string]string)

		if queryAllFlag {
			// Query all providers flag is set

			// Collect API keys for all available providers
			for provider := range llm.SupportedProviders {
				apiKey := cfg.GetAPIKey(provider)
				if apiKey != "" {
					apiKeys[provider] = apiKey
				}
			}

			// Check if we have any API keys
			if len(apiKeys) == 0 {
				return fmt.Errorf("no API keys found. Set at least one provider API key with: gollm set <provider> --api-key YOUR_API_KEY")
			}

			// Create LLM service with all API keys
			service := llm.NewService(apiKeys, httpClient)

			// Create and start spinner
			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Suffix = " Querying all configured providers..."
			s.Start()

			// Query all providers
			results := service.QueryAll(ctx, prompt, options...)

			// Stop spinner
			s.Stop()

			// Create a new tabwriter for formatted output
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			// Add header
			fmt.Fprintln(w, "PROVIDER\tMODEL\tTIME\tRESPONSE")
			fmt.Fprintln(w, "--------\t-----\t----\t--------")

			// Sort providers for consistent output
			sortedProviders := make([]string, 0, len(results))
			for provider := range results {
				sortedProviders = append(sortedProviders, provider)
			}
			sort.Strings(sortedProviders)

			// Print each result
			for _, provider := range sortedProviders {
				result := results[provider]
				var responseText string

				if result.Error != nil {
					// Format error
					responseText = fmt.Sprintf("ERROR: %v", result.Error)
				} else {
					// Get response and handle multi-line responses
					responseLines := strings.Split(result.Response, "\n")
					if len(responseLines) > 1 {
						// For multi-line responses, just show first line with indication
						responseText = responseLines[0] + " [...]"
					} else {
						responseText = result.Response
					}

					// Truncate long responses
					if len(responseText) > 60 {
						responseText = responseText[:57] + "..."
					}
				}

				// Format elapsed time
				timeStr := fmt.Sprintf("%dms", result.ElapsedTime.Milliseconds())

				// Print provider, model, time, and response
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", result.Provider, result.Model, timeStr, responseText)
			}

			// Flush the tabwriter
			w.Flush()

			// For each provider, print the full response
			fmt.Println("\nDetailed responses:")
			fmt.Println("------------------")
			for _, provider := range sortedProviders {
				result := results[provider]

				fmt.Printf("\n## %s (%s)\n\n", result.Provider, result.Model)
				if result.Error != nil {
					fmt.Printf("ERROR: %v\n", result.Error)
				} else {
					fmt.Println(result.Response)
				}
			}
		} else {
			// Regular single provider query

			// Validate model
			if !llm.IsValidModel(modelFlag) {
				return fmt.Errorf("unknown model: %s", modelFlag)
			}

			// Get provider for model
			providerName, _ := llm.GetProviderForModel(modelFlag)

			// Get API key from config or environment
			apiKey := cfg.GetAPIKey(providerName)
			if apiKey == "" {
				return fmt.Errorf("%s API key not found. Set it with: gollm set %s --api-key YOUR_API_KEY",
					providerName, providerName)
			}

			// Create LLM service with single API key
			apiKeys[providerName] = apiKey
			service := llm.NewService(apiKeys, httpClient)

			// Create and start spinner
			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Suffix = fmt.Sprintf(" Querying %s model %s...", providerName, modelFlag)
			s.Start()

			// Query the model with timing
			response, elapsedTime, err := service.QueryWithTiming(ctx, prompt, modelFlag, options...)

			// Stop spinner
			s.Stop()

			if err != nil {
				return fmt.Errorf("error querying model: %w", err)
			}

			if verboseFlag {
				// Create a new tabwriter for formatted output with colors
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

				// Define colors
				headerColor := color.New(color.FgCyan, color.Bold)
				modelColor := color.New(color.FgGreen)
				timeColor := color.New(color.FgYellow)

				// Add header with colors
				headerColor.Fprintln(w, "PROMPT\tMODEL\tTIME\tRESPONSE")
				headerColor.Fprintln(w, "------\t-----\t----\t--------")

				// Format prompt (truncate if too long)
				promptText := prompt
				if len(promptText) > 40 {
					promptText = promptText[:37] + "..."
				}

				// Format response for display (handle multiline)
				responseText := response
				responseLines := strings.Split(responseText, "\n")
				if len(responseLines) > 1 {
					responseText = responseLines[0] + " [...]"
				}

				// Truncate long responses for table
				if len(responseText) > 60 {
					responseText = responseText[:57] + "..."
				}

				// Print row with colored model and time
				fmt.Fprintf(w, "%s\t", promptText)
				modelColor.Fprintf(w, "%s\t", modelFlag)
				timeColor.Fprintf(w, "%dms\t", elapsedTime.Milliseconds())
				fmt.Fprintf(w, "%s\n", responseText)

				// Flush the tabwriter
				w.Flush()

				// Print full response after the table
				fmt.Println("\nFull response:")
				fmt.Println("-------------")
				fmt.Println(response)
			} else {
				// Print timing information
				fmt.Printf("Time: %dms\n\n", elapsedTime.Milliseconds())

				// Print response
				fmt.Println(response)
			}
		}

		return nil
	},
}

func init() {
	// Add flags specific to the complete command
	completeCmd.Flags().StringVarP(&systemPromptFlag, "system", "s", "", "System prompt to provide context")
	completeCmd.Flags().Float64VarP(&temperatureFlag, "temperature", "t", 0.7, "Temperature for response generation (0.0-1.0)")
	completeCmd.Flags().BoolVarP(&queryAllFlag, "all", "a", false, "Query all configured providers and compare responses")
	completeCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Display detailed response information in a colorful table")
}
