package llm

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/zerobang-dev/gollm/pkg/logger"
)

// MockProvider implements the Provider interface for testing
type MockProvider struct {
	Response string
}

// Query implements the Provider interface
func (p *MockProvider) Query(ctx context.Context, prompt string, options ...Option) (string, error) {
	return p.Response, nil
}

// Close implements the Provider interface
func (p *MockProvider) Close() error {
	return nil
}

func TestServiceWithLogger(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "gollm-service-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a logger
	testLogger, err := logger.NewLogger(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer testLogger.Close()

	// Create a mock provider
	mockProvider := &MockProvider{Response: "This is a test response"}

	// Create a service
	service := &Service{
		providers: map[string]Provider{
			"test": mockProvider,
		},
		httpClient: nil,
		logger:     testLogger,
	}

	// Set up a model mapping for testing
	originalModelToProvider := ModelToProvider
	ModelToProvider = map[string]string{"test-model": "test"}
	defer func() { ModelToProvider = originalModelToProvider }()

	originalSupportedProviders := SupportedProviders
	SupportedProviders = map[string]ProviderModel{
		"test": {Models: []string{"test-model"}},
	}
	defer func() { SupportedProviders = originalSupportedProviders }()

	// Test querying with the logger
	response, _, err := service.QueryWithTiming(
		context.Background(),
		"Test prompt",
		"test-model",
		WithTemperature(0.8),
	)

	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if response != "This is a test response" {
		t.Errorf("Expected response 'This is a test response', got '%s'", response)
	}

	// Give the goroutine time to log the query
	time.Sleep(100 * time.Millisecond)

	// Verify the query was logged
	queries, err := testLogger.GetRecentQueries(10)
	if err != nil {
		t.Fatalf("Failed to get recent queries: %v", err)
	}

	if len(queries) != 1 {
		t.Fatalf("Expected 1 query logged, got %d", len(queries))
	}

	q := queries[0]
	if q.Prompt != "Test prompt" {
		t.Errorf("Expected prompt 'Test prompt', got '%s'", q.Prompt)
	}
	if q.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", q.Model)
	}
	if q.Temperature != 0.8 {
		t.Errorf("Expected temperature 0.8, got %f", q.Temperature)
	}
}