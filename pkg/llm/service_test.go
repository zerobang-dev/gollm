package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// setupAnthropicMockServer creates a mock server for Anthropic API
func setupAnthropicMockServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return mock response
		mockResponse := anthropicResponse{
			Content: []anthropicContentBlock{
				{
					Type: "text",
					Text: "Mock Anthropic response",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(mockResponse); err != nil {
			t.Fatalf("Failed to encode Anthropic mock response: %v", err)
		}
	}))
}

// setupDeepseekMockServer creates a mock server for Deepseek API
func setupDeepseekMockServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return mock response
		mockResponse := deepseekResponse{
			Choices: []deepseekChoice{
				{
					Message: deepseekMessage{
						Content: "Mock Deepseek response",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(mockResponse); err != nil {
			t.Fatalf("Failed to encode Deepseek mock response: %v", err)
		}
	}))
}

// TestService tests the Service with multiple providers
func TestService(t *testing.T) {
	// Setup mock servers
	anthropicServer := setupAnthropicMockServer(t)
	defer anthropicServer.Close()

	deepseekServer := setupDeepseekMockServer(t)
	defer deepseekServer.Close()

	// Override provider implementations for testing
	origModelToProvider := ModelToProvider
	defer func() { ModelToProvider = origModelToProvider }()

	// Test model mapping
	ModelToProvider = map[string]string{
		"claude-3-7-sonnet-latest": "anthropic",
		"deepseek-chat":            "deepseek",
	}

	// Create HTTP client
	client := &http.Client{}

	// Create custom Anthropic provider that uses the mock server
	anthropicProvider := &AnthropicProvider{
		apiKey:     "test-anthropic-key",
		httpClient: client,
		baseURL:    anthropicServer.URL,
	}

	// Create custom Deepseek provider that uses the mock server
	deepseekProvider := &DeepseekProvider{
		apiKey:     "test-deepseek-key",
		httpClient: client,
		baseURL:    deepseekServer.URL,
	}

	// Create service with both providers
	service := &Service{
		providers: map[string]Provider{
			"anthropic": anthropicProvider,
			"deepseek":  deepseekProvider,
		},
		httpClient: client,
	}

	// Test querying Anthropic provider
	t.Run("Query Anthropic provider", func(t *testing.T) {
		response, err := service.Query(
			context.Background(),
			"Test prompt",
			"claude-3-7-sonnet-latest",
		)

		if err != nil {
			t.Fatalf("Query Anthropic provider returned error: %v", err)
		}

		expected := "Mock Anthropic response"
		if response != expected {
			t.Errorf("Expected response %q, got %q", expected, response)
		}
	})

	// Test querying Deepseek provider
	t.Run("Query Deepseek provider", func(t *testing.T) {
		response, err := service.Query(
			context.Background(),
			"Test prompt",
			"deepseek-chat",
		)

		if err != nil {
			t.Fatalf("Query Deepseek provider returned error: %v", err)
		}

		expected := "Mock Deepseek response"
		if response != expected {
			t.Errorf("Expected response %q, got %q", expected, response)
		}
	})

	// Test querying unknown model
	t.Run("Query unknown model", func(t *testing.T) {
		_, err := service.Query(
			context.Background(),
			"Test prompt",
			"unknown-model",
		)

		if err == nil {
			t.Fatal("Expected error for unknown model, got nil")
		}

		expected := "unknown model: unknown-model"
		if err.Error() != expected {
			t.Errorf("Expected error %q, got %q", expected, err.Error())
		}
	})

	// Test querying unconfigured provider
	t.Run("Query unconfigured provider", func(t *testing.T) {
		// Temporarily add a model mapping to a provider that isn't configured
		ModelToProvider["test-model"] = "unconfigured"
		defer delete(ModelToProvider, "test-model")

		_, err := service.Query(
			context.Background(),
			"Test prompt",
			"test-model",
		)

		if err == nil {
			t.Fatal("Expected error for unconfigured provider, got nil")
		}

		expected := "provider unconfigured not configured"
		if err.Error() != expected {
			t.Errorf("Expected error %q, got %q", expected, err.Error())
		}
	})

	// Test QueryAll method
	t.Run("QueryAll", func(t *testing.T) {
		// Setup supported providers and models
		origSupportedProviders := SupportedProviders
		defer func() { SupportedProviders = origSupportedProviders }()

		SupportedProviders = map[string]ProviderModel{
			"anthropic": {
				Models: []string{"claude-3-7-sonnet-latest"},
			},
			"deepseek": {
				Models: []string{"deepseek-chat", "deepseek-coder"},
			},
		}

		// Call QueryAll
		results := service.QueryAll(
			context.Background(),
			"Test prompt all providers",
		)

		// Check that we got responses from both providers
		if len(results) != 2 {
			t.Errorf("Expected responses from 2 providers, got %d", len(results))
		}

		// Check Anthropic response
		if anthropicResult, ok := results["anthropic"]; !ok {
			t.Error("Expected response from Anthropic provider")
		} else {
			if anthropicResult.Error != nil {
				t.Errorf("Anthropic provider returned error: %v", anthropicResult.Error)
			}

			if anthropicResult.Model != "claude-3-7-sonnet-latest" {
				t.Errorf("Expected model 'claude-3-7-sonnet-latest', got %q", anthropicResult.Model)
			}

			if anthropicResult.Response != "Mock Anthropic response" {
				t.Errorf("Expected response 'Mock Anthropic response', got %q", anthropicResult.Response)
			}

			if anthropicResult.Provider != "anthropic" {
				t.Errorf("Expected provider 'anthropic', got %q", anthropicResult.Provider)
			}
		}

		// Check Deepseek response
		if deepseekResult, ok := results["deepseek"]; !ok {
			t.Error("Expected response from Deepseek provider")
		} else {
			if deepseekResult.Error != nil {
				t.Errorf("Deepseek provider returned error: %v", deepseekResult.Error)
			}

			if deepseekResult.Model != "deepseek-chat" {
				t.Errorf("Expected model 'deepseek-chat', got %q", deepseekResult.Model)
			}

			if deepseekResult.Response != "Mock Deepseek response" {
				t.Errorf("Expected response 'Mock Deepseek response', got %q", deepseekResult.Response)
			}

			if deepseekResult.Provider != "deepseek" {
				t.Errorf("Expected provider 'deepseek', got %q", deepseekResult.Provider)
			}
		}
	})
}

// TestNewService tests creating a new service
func TestNewService(t *testing.T) {
	// Create API keys
	apiKeys := map[string]string{
		"anthropic": "test-anthropic-key",
		"deepseek":  "test-deepseek-key",
	}

	// Create service
	service := NewService(apiKeys, nil)

	// Check that providers were initialized
	if len(service.providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(service.providers))
	}

	// Check for Anthropic provider
	if _, ok := service.providers["anthropic"]; !ok {
		t.Error("Expected Anthropic provider to be initialized")
	}

	// Check for Deepseek provider
	if _, ok := service.providers["deepseek"]; !ok {
		t.Error("Expected Deepseek provider to be initialized")
	}

	// Check that HTTP client was initialized
	if service.httpClient == nil {
		t.Error("Expected HTTP client to be initialized")
	}
}

// TestNewServiceWithMissingKeys tests service creation with missing API keys
func TestNewServiceWithMissingKeys(t *testing.T) {
	// Create service with only one API key
	apiKeys := map[string]string{
		"anthropic": "test-anthropic-key",
		// Deepseek key intentionally omitted
	}

	service := NewService(apiKeys, nil)

	// Check that only one provider was initialized
	if len(service.providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(service.providers))
	}

	// Check for Anthropic provider
	if _, ok := service.providers["anthropic"]; !ok {
		t.Error("Expected Anthropic provider to be initialized")
	}

	// Check that Deepseek provider wasn't initialized
	if _, ok := service.providers["deepseek"]; ok {
		t.Error("Expected Deepseek provider not to be initialized")
	}
}

// TestGetDefaultModelForProvider tests the GetDefaultModelForProvider function
func TestGetDefaultModelForProvider(t *testing.T) {
	// Setup
	origSupportedProviders := SupportedProviders
	defer func() { SupportedProviders = origSupportedProviders }()

	SupportedProviders = map[string]ProviderModel{
		"provider1": {
			Models: []string{"model1", "model2"},
		},
		"provider2": {
			Models: []string{"model3"},
		},
		"empty-provider": {
			Models: []string{},
		},
	}

	// Test valid provider with multiple models
	t.Run("Valid provider with multiple models", func(t *testing.T) {
		model, err := GetDefaultModelForProvider("provider1")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if model != "model1" {
			t.Errorf("Expected first model 'model1', got %q", model)
		}
	})

	// Test valid provider with single model
	t.Run("Valid provider with single model", func(t *testing.T) {
		model, err := GetDefaultModelForProvider("provider2")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if model != "model3" {
			t.Errorf("Expected model 'model3', got %q", model)
		}
	})

	// Test provider with no models
	t.Run("Provider with no models", func(t *testing.T) {
		_, err := GetDefaultModelForProvider("empty-provider")

		if err == nil {
			t.Error("Expected error for provider with no models, got nil")
		}
	})

	// Test non-existent provider
	t.Run("Non-existent provider", func(t *testing.T) {
		_, err := GetDefaultModelForProvider("non-existent")

		if err == nil {
			t.Error("Expected error for non-existent provider, got nil")
		}
	})
}
