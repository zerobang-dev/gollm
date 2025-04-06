package llm

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/zerobang-dev/gollm/pkg/logger"
)

// ProviderResponse represents a response from a provider along with metadata
type ProviderResponse struct {
	Response    string        // The text response from the provider
	Model       string        // The model used for the response
	Provider    string        // The provider name
	Error       error         // Error, if any occurred during the query
	ElapsedTime time.Duration // Time taken to get the response
}

// Service manages LLM providers
type Service struct {
	providers  map[string]Provider
	httpClient *http.Client
	logger     *logger.Logger // Optional query logger
}

// NewService creates a new LLM service with API keys and an optional HTTP client
func NewService(apiKeys map[string]string, httpClient *http.Client) *Service {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	service := &Service{
		providers:  make(map[string]Provider),
		httpClient: httpClient,
	}

	// Initialize providers with their respective API keys
	if anthropicKey, ok := apiKeys["anthropic"]; ok && anthropicKey != "" {
		service.providers["anthropic"] = NewAnthropicProvider(anthropicKey, httpClient)
	}

	if deepseekKey, ok := apiKeys["deepseek"]; ok && deepseekKey != "" {
		service.providers["deepseek"] = NewDeepseekProvider(deepseekKey, httpClient)
	}

	if googleKey, ok := apiKeys["google"]; ok && googleKey != "" {
		service.providers["google"] = NewGoogleProvider(googleKey, httpClient)
	}

	return service
}

// SetLogger sets the query logger for the service
func (s *Service) SetLogger(l *logger.Logger) {
	s.logger = l
}

// QueryWithTiming sends a prompt to the model and returns the response with timing information
func (s *Service) QueryWithTiming(ctx context.Context, prompt, modelName string, options ...Option) (string, time.Duration, error) {
	// Validate model
	if !IsValidModel(modelName) {
		return "", 0, fmt.Errorf("unknown model: %s", modelName)
	}

	// Determine provider based on model name
	providerName, _ := GetProviderForModel(modelName)

	// Get provider
	provider, ok := s.providers[providerName]
	if !ok {
		return "", 0, fmt.Errorf("provider %s not configured", providerName)
	}

	// Add model to options
	options = append([]Option{WithModel(modelName)}, options...)

	// Extract temperature for logging
	temperature := 0.7 // default
	for _, opt := range options {
		if opt == nil {
			continue
		}
		reqOpts := &RequestOptions{}
		opt(reqOpts)
		if reqOpts.Temperature != 0 {
			temperature = reqOpts.Temperature
		}
	}

	// Start the timer
	startTime := time.Now()

	// Query the provider
	response, err := provider.Query(ctx, prompt, options...)

	// Calculate elapsed time
	elapsedTime := time.Since(startTime)

	// Log query if logger is configured
	if err == nil && s.logger != nil {
		// Only log successful queries
		// Use a goroutine to avoid blocking the response
		go func() {
			if logErr := s.logger.LogQuery(prompt, modelName, response, elapsedTime, temperature); logErr != nil {
				// Just print the error but don't fail the request
				fmt.Fprintf(os.Stderr, "Failed to log query: %v\n", logErr)
			}
		}()
	}

	return response, elapsedTime, err
}

// Query sends a prompt to the model using the appropriate provider
func (s *Service) Query(ctx context.Context, prompt, modelName string, options ...Option) (string, error) {
	response, _, err := s.QueryWithTiming(ctx, prompt, modelName, options...)
	return response, err
}

// GetDefaultModelForProvider returns the first (default) model for a provider
func GetDefaultModelForProvider(providerName string) (string, error) {
	models := GetModelsForProvider(providerName)
	if len(models) == 0 {
		return "", fmt.Errorf("no models available for provider %s", providerName)
	}

	// Return the first model as the default
	return models[0], nil
}

// QueryAll sends a prompt to all configured providers and returns their responses
func (s *Service) QueryAll(ctx context.Context, prompt string, options ...Option) map[string]ProviderResponse {
	results := make(map[string]ProviderResponse)
	resultsMutex := sync.Mutex{}

	// Create a wait group to wait for all queries to complete
	var wg sync.WaitGroup

	// Query each provider concurrently
	for providerName, provider := range s.providers {
		wg.Add(1)

		go func(providerName string, provider Provider) {
			defer wg.Done()

			// Find a valid model for this provider
			defaultModel, err := GetDefaultModelForProvider(providerName)
			if err != nil {
				resultsMutex.Lock()
				results[providerName] = ProviderResponse{
					Provider: providerName,
					Error:    err,
				}
				resultsMutex.Unlock()
				return
			}

			// Create a copy of options and add the model
			providerOptions := make([]Option, len(options)+1)
			providerOptions[0] = WithModel(defaultModel)
			copy(providerOptions[1:], options)

			// Start the timer
			startTime := time.Now()

			// Query the provider
			response, err := provider.Query(ctx, prompt, providerOptions...)

			// Calculate elapsed time
			elapsedTime := time.Since(startTime)

			// Store the result
			resultsMutex.Lock()
			results[providerName] = ProviderResponse{
				Response:    response,
				Model:       defaultModel,
				Provider:    providerName,
				Error:       err,
				ElapsedTime: elapsedTime,
			}
			resultsMutex.Unlock()
		}(providerName, provider)
	}

	// Wait for all queries to complete
	wg.Wait()

	return results
}
