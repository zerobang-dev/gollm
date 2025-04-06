package logger

import (
	"os"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "gollm-logger-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

	// Create a new logger
	logger, err := NewLogger(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			t.Errorf("Failed to close logger: %v", err)
		}
	}()

	// Log a test query
	err = logger.LogQuery(
		"test prompt",
		"test-model",
		"test response",
		100*time.Millisecond,
		0.7,
	)
	if err != nil {
		t.Fatalf("Failed to log query: %v", err)
	}

	// Retrieve recent queries
	queries, err := logger.GetRecentQueries(10)
	if err != nil {
		t.Fatalf("Failed to retrieve queries: %v", err)
	}

	// Verify query was logged properly
	if len(queries) != 1 {
		t.Fatalf("Expected 1 query, got %d", len(queries))
	}

	q := queries[0]
	if q.Prompt != "test prompt" {
		t.Errorf("Expected prompt 'test prompt', got %q", q.Prompt)
	}
	if q.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got %q", q.Model)
	}
	if q.Response != "test response" {
		t.Errorf("Expected response 'test response', got %q", q.Response)
	}
	if q.Duration != 100 {
		t.Errorf("Expected duration 100ms, got %d", q.Duration)
	}
	if q.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", q.Temperature)
	}

	// Test search functionality
	err = logger.LogQuery(
		"another test with golang",
		"test-model",
		"response about golang",
		150*time.Millisecond,
		0.8,
	)
	if err != nil {
		t.Fatalf("Failed to log second query: %v", err)
	}

	// Search for "golang"
	searchResults, err := logger.SearchQueries("golang", 10)
	if err != nil {
		t.Fatalf("Failed to search queries: %v", err)
	}

	if len(searchResults) != 1 {
		t.Fatalf("Expected 1 search result, got %d", len(searchResults))
	}

	if searchResults[0].Prompt != "another test with golang" {
		t.Errorf("Expected to find query with 'golang', got: %q", searchResults[0].Prompt)
	}
}