package llm

import (
	"context"
)

// Provider defines the interface for LLM providers
type Provider interface {
	// Query sends a prompt to the LLM and returns the response
	Query(ctx context.Context, prompt string, options ...Option) (string, error)
}

// Option is a functional option for configuring LLM requests
type Option func(*RequestOptions)

// RequestOptions contains configuration for an LLM request
type RequestOptions struct {
	Model       string
	MaxTokens   int
	Temperature float64

	// Provider-specific parameters stored as key-value pairs
	CustomParams map[string]interface{}
}

// WithModel sets the model for the request
func WithModel(model string) Option {
	return func(o *RequestOptions) {
		o.Model = model
	}
}

// WithMaxTokens sets the maximum number of tokens to generate
func WithMaxTokens(maxTokens int) Option {
	return func(o *RequestOptions) {
		o.MaxTokens = maxTokens
	}
}

// WithTemperature sets the sampling temperature
func WithTemperature(temperature float64) Option {
	return func(o *RequestOptions) {
		o.Temperature = temperature
	}
}

// WithCustomParam sets a provider-specific parameter
func WithCustomParam(key string, value interface{}) Option {
	return func(o *RequestOptions) {
		if o.CustomParams == nil {
			o.CustomParams = make(map[string]interface{})
		}
		o.CustomParams[key] = value
	}
}
