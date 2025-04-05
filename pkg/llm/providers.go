package llm

// ProviderModel defines a structure for provider information
type ProviderModel struct {
	// Models supported by this provider
	Models []string
}

// SupportedProviders maps provider names to their information
var SupportedProviders = map[string]ProviderModel{
	"anthropic": {
		Models: []string{
			"claude-3-7-sonnet-latest",
		},
	},
	"deepseek": {
		Models: []string{
			"deepseek-coder",
			"deepseek-chat",
		},
	},
	"google": {
		Models: []string{
			"gemini-2.5-pro-exp-03-25",
			"gemini-2.0-flash",
			"gemini-2.0-flash-lite",
			"gemini-1.5-flash",
			"gemini-1.5-flash-8b",
		},
	},
	// Add more providers here
}

// ModelToProvider maps from model name to provider name (generated at init)
var ModelToProvider map[string]string

func init() {
	// Build model to provider map
	ModelToProvider = make(map[string]string)
	for provider, info := range SupportedProviders {
		for _, model := range info.Models {
			ModelToProvider[model] = provider
		}
	}
}

// IsValidProvider checks if a provider is supported
func IsValidProvider(provider string) bool {
	_, ok := SupportedProviders[provider]
	return ok
}

// GetModelsForProvider returns the models for a provider
func GetModelsForProvider(provider string) []string {
	if info, ok := SupportedProviders[provider]; ok {
		return info.Models
	}
	return nil
}

// GetProviderForModel returns the provider for a model
func GetProviderForModel(model string) (string, bool) {
	provider, ok := ModelToProvider[model]
	return provider, ok
}

// IsValidModel checks if a model is supported
func IsValidModel(model string) bool {
	_, ok := ModelToProvider[model]
	return ok
}
