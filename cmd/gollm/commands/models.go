package commands

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/zerobang-dev/gollm/pkg/llm"
)

// listModelsCmd represents the models command
var listModelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List all supported models",
	Long:  `Display a list of all supported LLM models organized by provider.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a tabwriter for clean columnar output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		// Add header
		fmt.Fprintln(w, "PROVIDER\tMODEL")
		fmt.Fprintln(w, "--------\t-----")

		// Get providers and sort them for consistent output
		providers := make([]string, 0, len(llm.SupportedProviders))
		for provider := range llm.SupportedProviders {
			providers = append(providers, provider)
		}
		sort.Strings(providers)

		// Display models for each provider
		for _, provider := range providers {
			providerInfo := llm.SupportedProviders[provider]

			// Sort models within each provider for consistent output
			models := make([]string, len(providerInfo.Models))
			copy(models, providerInfo.Models)
			sort.Strings(models)

			// Print each model
			for _, model := range models {
				fmt.Fprintf(w, "%s\t%s\n", provider, model)
			}
		}

		// Flush the tabwriter
		w.Flush()

		return nil
	},
}

func init() {
	// No flags needed for this command
}
