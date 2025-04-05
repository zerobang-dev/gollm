package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewDeepseekProvider ensures the provider is initialized correctly
func TestNewDeepseekProvider(t *testing.T) {
	apiKey := "test-api-key"
	client := &http.Client{}

	provider := NewDeepseekProvider(apiKey, client)

	if provider.apiKey != apiKey {
		t.Errorf("Expected API key to be %q, got %q", apiKey, provider.apiKey)
	}

	if provider.httpClient != client {
		t.Errorf("Expected HTTP client to be %v, got %v", client, provider.httpClient)
	}

	if provider.baseURL != "https://api.deepseek.com/v1/chat/completions" {
		t.Errorf("Expected baseURL to be %q, got %q", "https://api.deepseek.com/v1/chat/completions", provider.baseURL)
	}
}

// TestNewDeepseekProviderWithNilClient ensures the provider handles nil HTTP client
func TestNewDeepseekProviderWithNilClient(t *testing.T) {
	apiKey := "test-api-key"
	provider := NewDeepseekProvider(apiKey, nil)

	if provider.httpClient == nil {
		t.Error("Expected default HTTP client, got nil")
	}

	if provider.httpClient != http.DefaultClient {
		t.Errorf("Expected HTTP client to be default client, got different client")
	}
}

// TestDeepseekProviderMissingModel tests error handling for missing model
func TestDeepseekProviderMissingModel(t *testing.T) {
	provider := NewDeepseekProvider("test-key", nil)
	_, err := provider.Query(context.Background(), "Test prompt")

	if err == nil {
		t.Error("Expected error for missing model, got nil")
	}

	expectedErr := "model is required for Deepseek provider"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
	}
}

// TestDeepseekProviderQuery tests successful queries
func TestDeepseekProviderQuery(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header to be application/json, got %s", r.Header.Get("Content-Type"))
		}

		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header to be Bearer test-key, got %s", r.Header.Get("Authorization"))
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		// Parse request
		var req deepseekRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("Failed to parse request body: %v", err)
		}

		// Check request fields
		if req.Model != "deepseek-chat" {
			t.Errorf("Expected model to be deepseek-chat, got %s", req.Model)
		}

		if len(req.Messages) == 0 || req.Messages[0].Role != "user" || req.Messages[0].Content != "Test prompt" {
			t.Errorf("Expected user message with content 'Test prompt', got %v", req.Messages)
		}

		// Return mock response
		mockResponse := deepseekResponse{
			ID:      "resp-123",
			Object:  "chat.completion",
			Created: 1712227200,
			Model:   "deepseek-chat",
			Choices: []deepseekChoice{
				{
					Index: 0,
					Message: deepseekMessage{
						Role:    "assistant",
						Content: "This is a mock response from Deepseek",
					},
					FinishReason: "stop",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create provider using test server URL
	provider := &DeepseekProvider{
		apiKey:     "test-key",
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	// Test query
	response, err := provider.Query(
		context.Background(),
		"Test prompt",
		WithModel("deepseek-chat"),
		WithMaxTokens(1000),
		WithTemperature(0.7),
	)

	// Check for errors
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}

	// Check response
	expected := "This is a mock response from Deepseek"
	if response != expected {
		t.Errorf("Expected response %q, got %q", expected, response)
	}
}

// TestDeepseekProviderQueryWithSystemPrompt tests query with system prompt
func TestDeepseekProviderQueryWithSystemPrompt(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		// Parse request
		var req deepseekRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("Failed to parse request body: %v", err)
		}

		// Check system message
		if len(req.Messages) < 2 || req.Messages[0].Role != "system" || req.Messages[0].Content != "You are a helpful assistant" {
			t.Errorf("Expected system message, got %v", req.Messages)
		}

		// Return mock response
		mockResponse := deepseekResponse{
			Choices: []deepseekChoice{
				{
					Message: deepseekMessage{
						Content: "Response with system prompt",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create provider using test server URL
	provider := &DeepseekProvider{
		apiKey:     "test-key",
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	// Test query with system prompt
	_, err := provider.Query(
		context.Background(),
		"Test prompt",
		WithModel("deepseek-chat"),
		WithCustomParam("system", "You are a helpful assistant"),
	)

	// Check for errors
	if err != nil {
		t.Fatalf("Query with system prompt returned error: %v", err)
	}
}

// TestDeepseekProviderAPIError tests handling of API errors
func TestDeepseekProviderAPIError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errResponse := deepseekResponse{
			Error: &struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			}{
				Message: "Invalid API key",
				Type:    "authentication_error",
				Code:    "invalid_api_key",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errResponse)
	}))
	defer server.Close()

	// Create provider using test server URL
	provider := &DeepseekProvider{
		apiKey:     "invalid-key",
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	// Test query with invalid API key
	_, err := provider.Query(
		context.Background(),
		"Test prompt",
		WithModel("deepseek-chat"),
	)

	// Check for errors
	if err == nil {
		t.Fatal("Expected error for invalid API key, got nil")
	}

	// Check error message
	if err.Error() != "API error (authentication_error): Invalid API key" {
		t.Errorf("Expected error message about API key, got: %v", err)
	}
}

// TestDeepseekProviderEmptyResponse tests handling of empty responses
func TestDeepseekProviderEmptyResponse(t *testing.T) {
	// Create a test server that returns an empty response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		emptyResponse := deepseekResponse{
			Choices: []deepseekChoice{}, // Empty choices
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(emptyResponse)
	}))
	defer server.Close()

	// Create provider using test server URL
	provider := &DeepseekProvider{
		apiKey:     "test-key",
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	// Test query
	_, err := provider.Query(
		context.Background(),
		"Test prompt",
		WithModel("deepseek-chat"),
	)

	// Check for errors
	if err == nil {
		t.Fatal("Expected error for empty response, got nil")
	}

	// Check error message
	if err.Error() != "empty response from Deepseek API" {
		t.Errorf("Expected error about empty response, got: %v", err)
	}
}

// TestDeepseekProviderNetworkError tests handling of network errors
func TestDeepseekProviderNetworkError(t *testing.T) {
	// Create provider with invalid URL to simulate network error
	provider := &DeepseekProvider{
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		baseURL:    "http://invalid-url-that-does-not-exist.example",
	}

	// Test query with invalid URL
	_, err := provider.Query(
		context.Background(),
		"Test prompt",
		WithModel("deepseek-chat"),
	)

	// Check for errors
	if err == nil {
		t.Fatal("Expected network error, got nil")
	}
}

// TestDeepseekProviderContext tests context handling
func TestDeepseekProviderContext(t *testing.T) {
	// Create a test server with a delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request has a context
		if r.Context() == nil {
			t.Error("Request context is nil")
		}

		// Return a valid response
		mockResponse := deepseekResponse{
			Choices: []deepseekChoice{
				{
					Message: deepseekMessage{
						Content: "Response",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create provider using test server URL
	provider := &DeepseekProvider{
		apiKey:     "test-key",
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	// Create a context
	ctx := context.Background()

	// Test query with context
	_, err := provider.Query(
		ctx,
		"Test prompt",
		WithModel("deepseek-chat"),
	)

	// Check for errors
	if err != nil {
		t.Fatalf("Query with context returned error: %v", err)
	}
}
