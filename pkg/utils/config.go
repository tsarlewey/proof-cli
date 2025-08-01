package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the configuration for the CLI
type Config struct {
	APIEndpoint string        `json:"api_endpoint"`
	Timeout     time.Duration `json:"timeout"`
	OAuth       *OAuthConfig  `json:"oauth,omitempty"`
	APIKey      string        `json:"api_key,omitempty"`
	OAuthToken  *OAuthToken   `json:"oauth_token,omitempty"`
}

// OAuthConfig represents OAuth configuration
type OAuthConfig struct {
	Enabled      bool   `json:"enabled"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope,omitempty"`
}

// OAuthToken represents an OAuth access token
type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
}

// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
	// Default config
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		Timeout:     30 * time.Second,
	}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config, fmt.Errorf("error getting home directory: %w", err)
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Join(homeDir, ".proof-cli")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return config, fmt.Errorf("error creating config directory: %w", err)
		}
	}

	// Check if config file exists
	configFile := filepath.Join(configDir, "config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create default config file
		configJSON, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return config, fmt.Errorf("error marshaling config: %w", err)
		}

		if err := os.WriteFile(configFile, configJSON, 0600); err != nil {
			return config, fmt.Errorf("error writing config file: %w", err)
		}

		return config, nil
	}

	// Read config file
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return config, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse config file
	if err := json.Unmarshal(configData, config); err != nil {
		return config, fmt.Errorf("error parsing config file: %w", err)
	}

	return config, nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(config *Config) error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Join(homeDir, ".proof-cli")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("error creating config directory: %w", err)
		}
	}

	// Marshal config
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	// Write config file with restricted permissions since it now contains sensitive data
	configFile := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configFile, configJSON, 0600); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// GetAPIKey gets the API key from the environment or config
func GetAPIKey() (string, error) {
	// Check environment variable first
	apiKey := os.Getenv("PROOF_API_KEY")
	if apiKey != "" {
		return apiKey, nil
	}

	// Load config and get API key
	config, err := LoadConfig()
	if err != nil {
		return "", fmt.Errorf("error loading config: %w", err)
	}

	if config.APIKey == "" {
		return "", fmt.Errorf("API key not found. Set PROOF_API_KEY environment variable or run 'proof config set-api-key'")
	}

	return config.APIKey, nil
}

// SaveAPIKey saves the API key to the config
func SaveAPIKey(apiKey string) error {
	// Load current config
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	// Update API key
	config.APIKey = apiKey

	// Save config
	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	return nil
}
