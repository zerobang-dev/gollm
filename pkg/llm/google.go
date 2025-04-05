package llm

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GoogleProvider implements the Provider interface for Google Vertex AI API
type GoogleProvider struct {
	apiKey string
	client *genai.Client
}

// NewGoogleProvider creates a new Google Vertex AI provider
func NewGoogleProvider(apiKey string, _ *http.Client) *GoogleProvider {
	// Create a new genai client
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		// Return empty provider that will fail on first use
		return &GoogleProvider{
			apiKey: apiKey,
		}
	}

	return &GoogleProvider{
		apiKey: apiKey,
		client: client,
	}
}

// Query implements the Provider interface
func (p *GoogleProvider) Query(ctx context.Context, prompt string, options ...Option) (string, error) {
	// Check if client is initialized
	if p.client == nil {
		return "", errors.New("google client not initialized")
	}

	// Apply options
	opts := &RequestOptions{
		MaxTokens:   1024,
		Temperature: 0.7,
	}

	for _, option := range options {
		option(opts)
	}

	// Model is required
	if opts.Model == "" {
		return "", errors.New("model is required for Google provider")
	}

	// Create a new model
	model := p.client.GenerativeModel(opts.Model)

	// Configure model
	model.SetTemperature(float32(opts.Temperature))
	model.SetMaxOutputTokens(int32(opts.MaxTokens))

	// Add top_p if specified
	if topP, ok := opts.CustomParams["top_p"].(float64); ok {
		model.SetTopP(float32(topP))
	}

	// Add top_k if specified
	if topK, ok := opts.CustomParams["top_k"].(int); ok {
		model.SetTopK(int32(topK))
	}

	// Set system instructions if specified
	if system, ok := opts.CustomParams["system"].(string); ok && system != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(system)},
			Role:  "system",
		}
	}

	// Create prompt part (single part)
	promptPart := genai.Text(prompt)

	// Generate content
	resp, err := model.GenerateContent(ctx, promptPart)
	if err != nil {
		return "", fmt.Errorf("error generating content: %w", err)
	}

	// Process response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty response from Google API")
	}

	// Return the text from the first candidate's first part
	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", errors.New("unexpected response type from Google API")
	}

	return string(text), nil
}

// Close closes the provider's resources
func (p *GoogleProvider) Close() error {
	if p.client != nil {
		if err := p.client.Close(); err != nil {
				return fmt.Errorf("error closing Google client: %w", err)
			}
	}
	return nil
}
