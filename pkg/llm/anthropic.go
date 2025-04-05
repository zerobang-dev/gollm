package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// AnthropicProvider implements the Provider interface for Anthropic API
type AnthropicProvider struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// anthropicMessage represents a message in the Anthropic API
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicRequest represents a request to the Anthropic API
type anthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	Temperature float64            `json:"temperature"`
	TopP        float64            `json:"top_p,omitempty"`
}

// anthropicContentBlock represents a content block in the Anthropic API response
type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// anthropicResponse represents a response from the Anthropic API
type anthropicResponse struct {
	Content []anthropicContentBlock `json:"content"`
	Error   *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey string, httpClient *http.Client) *AnthropicProvider {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &AnthropicProvider{
		apiKey:     apiKey,
		httpClient: httpClient,
		baseURL:    "https://api.anthropic.com/v1/messages",
	}
}

// Query implements the Provider interface
func (p *AnthropicProvider) Query(ctx context.Context, prompt string, options ...Option) (string, error) {
	// Apply options
	opts := &RequestOptions{
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	for _, option := range options {
		option(opts)
	}

	// Model is required
	if opts.Model == "" {
		return "", errors.New("model is required for Anthropic provider")
	}

	// Create request payload
	req := anthropicRequest{
		Model:       opts.Model,
		MaxTokens:   opts.MaxTokens,
		Messages:    []anthropicMessage{{Role: "user", Content: prompt}},
		Temperature: opts.Temperature,
	}

	// Add system prompt if specified
	if system, ok := opts.CustomParams["system"].(string); ok && system != "" {
		req.System = system
	}

	// Add top_p if specified
	if topP, ok := opts.CustomParams["top_p"].(float64); ok {
		req.TopP = topP
	}

	// Convert to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.baseURL,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer func() {
			if err := resp.Body.Close(); err != nil {
				// Just log the error, can't return it here
				fmt.Printf("Error closing response body: %v\n", err)
			}
		}()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		var errResp anthropicResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
			return "", fmt.Errorf("API error (%s): %s", errResp.Error.Type, errResp.Error.Message)
		}
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result anthropicResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing response: %w", err)
	}

	// Check for empty response
	if len(result.Content) == 0 {
		return "", errors.New("empty response from Anthropic API")
	}

	// Find text content
	for _, block := range result.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}

	return "", errors.New("no text content in response")
}
