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

// DeepseekProvider implements the Provider interface for Deepseek API
type DeepseekProvider struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// deepseekMessage represents a message in the Deepseek API
type deepseekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// deepseekRequest represents a request to the Deepseek API
type deepseekRequest struct {
	Model       string            `json:"model"`
	Messages    []deepseekMessage `json:"messages"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float64           `json:"temperature,omitempty"`
	TopP        float64           `json:"top_p,omitempty"`
	Stream      bool              `json:"stream,omitempty"`
}

// deepseekChoice represents a choice in the Deepseek API response
type deepseekChoice struct {
	Index        int             `json:"index"`
	Message      deepseekMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

// deepseekResponse represents a response from the Deepseek API
type deepseekResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []deepseekChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// NewDeepseekProvider creates a new Deepseek provider
func NewDeepseekProvider(apiKey string, httpClient *http.Client) *DeepseekProvider {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &DeepseekProvider{
		apiKey:     apiKey,
		httpClient: httpClient,
		baseURL:    "https://api.deepseek.com/v1/chat/completions",
	}
}

// Query implements the Provider interface
func (p *DeepseekProvider) Query(ctx context.Context, prompt string, options ...Option) (string, error) {
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
		return "", errors.New("model is required for Deepseek provider")
	}

	// Create request payload
	req := deepseekRequest{
		Model:       opts.Model,
		MaxTokens:   opts.MaxTokens,
		Messages:    []deepseekMessage{{Role: "user", Content: prompt}},
		Temperature: opts.Temperature,
		Stream:      false,
	}

	// Add system prompt if specified
	if system, ok := opts.CustomParams["system"].(string); ok && system != "" {
		req.Messages = append([]deepseekMessage{{Role: "system", Content: system}}, req.Messages...)
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
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

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
		var errResp deepseekResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
			return "", fmt.Errorf("API error (%s): %s", errResp.Error.Type, errResp.Error.Message)
		}
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result deepseekResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing response: %w", err)
	}

	// Check for empty choices
	if len(result.Choices) == 0 {
		return "", errors.New("empty response from Deepseek API")
	}

	// Return the content from the first choice
	return result.Choices[0].Message.Content, nil
}
