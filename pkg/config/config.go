package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zerobang-dev/gollm/pkg/llm"
	"gopkg.in/yaml.v3"
)

// ProviderConfig holds configuration for a specific provider
type ProviderConfig struct {
	APIKey string `yaml:"api_key"`
}

// Config represents the application configuration
type Config struct {
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// Load loads the configuration from the file system
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return empty config
		return &Config{
			Providers: make(map[string]ProviderConfig),
		}, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse config
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Initialize providers map if nil
	if config.Providers == nil {
		config.Providers = make(map[string]ProviderConfig)
	}

	return &config, nil
}

// Save saves the configuration to the file system
func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Convert to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// GetAPIKey returns the API key for the specified provider
func (c *Config) GetAPIKey(provider string) string {
	// Check config file first
	if providerConfig, ok := c.Providers[provider]; ok && providerConfig.APIKey != "" {
		return providerConfig.APIKey
	}

	// Fall back to environment variable (convert to uppercase for env var)
	envKey := fmt.Sprintf("%s_API_KEY", strings.ToUpper(provider))
	return os.Getenv(envKey)
}

// SetAPIKey sets the API key for the specified provider
func (c *Config) SetAPIKey(provider, apiKey string) error {
	// Validate provider
	if !llm.IsValidProvider(provider) {
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	if c.Providers == nil {
		c.Providers = make(map[string]ProviderConfig)
	}

	config := c.Providers[provider]
	config.APIKey = apiKey
	c.Providers[provider] = config

	return nil
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "gollm")
	configPath := filepath.Join(configDir, "config.yml")

	return configPath, nil
}
