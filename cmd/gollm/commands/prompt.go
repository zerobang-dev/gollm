package commands

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"github.com/zerobang-dev/gollm/pkg/config"
	"github.com/zerobang-dev/gollm/pkg/logger"
	"github.com/zerobang-dev/gollm/pkg/llm"
)

// readPromptFromArgs reads the prompt from command arguments or stdin
func readPromptFromArgs(cmd *cobra.Command, args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}

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
		return "", fmt.Errorf("no input provided via argument or stdin")
	}

	return string(buffer), nil
}

// queryLLM sends a prompt to an LLM and returns the response
func queryLLM(ctx context.Context, prompt string, modelFlag string, systemPromptFlag string, temperatureFlag float64, queryAllFlag bool) (interface{}, error) {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}
	
	// Initialize logger
	var queryLogger *logger.Logger
	configDir := config.GetConfigDir()
	queryLogger, err = logger.NewLogger(configDir)
	if err != nil {
		// Just log a warning but continue without logging
		fmt.Printf("Warning: Query logging disabled - %v\n", err)
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 120 * time.Second,
	}

	// Set up options
	options := []llm.Option{
		llm.WithMaxTokens(1000),
		llm.WithTemperature(temperatureFlag),
	}

	// If system prompt is provided, add it as a custom parameter
	if systemPromptFlag != "" {
		options = append(options, llm.WithCustomParam("system", systemPromptFlag))
	}

	// Initialize API keys map - not used directly here

	// Close logger when function returns
	if queryLogger != nil {
		defer queryLogger.Close()
	}

	if queryAllFlag {
		// Query all providers flag is set
		return queryAllProviders(ctx, prompt, cfg, httpClient, options, queryLogger)
	} else {
		// Regular single provider query
		return querySingleProvider(ctx, prompt, modelFlag, cfg, httpClient, options, queryLogger)
	}
}

// queryAllProviders queries all available providers and returns results
func queryAllProviders(ctx context.Context, prompt string, cfg *config.Config, httpClient *http.Client, options []llm.Option, queryLogger *logger.Logger) (map[string]llm.ProviderResponse, error) {
	// Collect API keys for all available providers
	allApiKeys := make(map[string]string)
	for provider := range llm.SupportedProviders {
		apiKey := cfg.GetAPIKey(provider)
		if apiKey != "" {
			allApiKeys[provider] = apiKey
		}
	}

	// Check if we have any API keys
	if len(allApiKeys) == 0 {
		return nil, fmt.Errorf("no API keys found. Set at least one provider API key with: gollm set <provider> --api-key YOUR_API_KEY")
	}

	// Create LLM service with all API keys
	service := llm.NewService(allApiKeys, httpClient)
	
	// Set logger if available
	if queryLogger != nil {
		service.SetLogger(queryLogger)
	}

	// Create and start spinner
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Querying all configured providers..."
	s.Start()

	// Query all providers
	results := service.QueryAll(ctx, prompt, options...)

	// Stop spinner
	s.Stop()

	return results, nil
}

// querySingleProvider queries a single provider and returns the result
func querySingleProvider(ctx context.Context, prompt string, modelFlag string, cfg *config.Config, httpClient *http.Client, options []llm.Option, queryLogger *logger.Logger) (*struct {
	Response    string
	ElapsedTime time.Duration
}, error) {
	// Validate model
	if !llm.IsValidModel(modelFlag) {
		return nil, fmt.Errorf("unknown model: %s", modelFlag)
	}

	// Get provider for model
	providerName, _ := llm.GetProviderForModel(modelFlag)

	// Get API key from config or environment
	apiKey := cfg.GetAPIKey(providerName)
	if apiKey == "" {
		return nil, fmt.Errorf("%s API key not found. Set it with: gollm set %s --api-key YOUR_API_KEY",
			providerName, providerName)
	}

	// Create LLM service with single API key
	apiKeys := make(map[string]string)
	apiKeys[providerName] = apiKey
	service := llm.NewService(apiKeys, httpClient)
	
	// Set logger if available
	if queryLogger != nil {
		service.SetLogger(queryLogger)
	}

	// Create and start spinner
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = fmt.Sprintf(" Querying %s model %s...", providerName, modelFlag)
	s.Start()

	// Query the model with timing
	response, elapsedTime, err := service.QueryWithTiming(ctx, prompt, modelFlag, options...)

	// Stop spinner
	s.Stop()

	if err != nil {
		return nil, fmt.Errorf("error querying model: %w", err)
	}

	return &struct {
		Response    string
		ElapsedTime time.Duration
	}{
		Response:    response,
		ElapsedTime: elapsedTime,
	}, nil
}