package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zerobang-dev/go-llm/pkg/config"
	"github.com/zerobang-dev/go-llm/pkg/llm"
)

var apiKeyFlag string

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set [provider]",
	Short: "Configure provider settings",
	Long:  `Set configuration values for LLM providers, such as API keys.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerName := args[0]

		// Validate provider
		if !llm.IsValidProvider(providerName) {
			return fmt.Errorf("unsupported provider: %s", providerName)
		}

		// API key is required
		if apiKeyFlag == "" {
			return fmt.Errorf("API key is required (use --api-key flag)")
		}

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		// Set API key
		if err := cfg.SetAPIKey(providerName, apiKeyFlag); err != nil {
			return fmt.Errorf("error setting API key: %w", err)
		}

		// Save config
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("error saving config: %w", err)
		}

		fmt.Printf("API key for %s has been set.\n", providerName)
		fmt.Printf("Configuration saved to ~/.config/go-llm/config.yml\n")

		return nil
	},
}

func init() {
	setCmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "API key for the provider")
	setCmd.MarkFlagRequired("api-key")
}
