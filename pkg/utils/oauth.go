package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// PrepareOAuthRequest prepares OAuth request parameters and headers
type OAuthRequest struct {
	URL      string
	FormData string
	Headers  map[string]string
}

// PrepareOAuthTokenRequest creates the OAuth request parameters without making HTTP calls
func PrepareOAuthTokenRequest(config *Config) (*OAuthRequest, error) {
	if config.OAuth == nil || !config.OAuth.Enabled {
		return nil, fmt.Errorf("OAuth not enabled in configuration")
	}

	if config.OAuth.ClientID == "" || config.OAuth.ClientSecret == "" {
		return nil, fmt.Errorf("OAuth client ID and secret are required")
	}

	// Prepare OAuth token request URL
	tokenURL := fmt.Sprintf("%s/oauth/v2/token", config.APIEndpoint)

	// Prepare form data
	formData := url.Values{}
	formData.Set("grant_type", "client_credentials")
	if config.OAuth.Scope != "" {
		formData.Set("scope", config.OAuth.Scope)
	}

	// Prepare headers
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"Accept":       "application/json",
	}

	// Set Basic Auth header with client credentials
	auth := base64.StdEncoding.EncodeToString([]byte(config.OAuth.ClientID + ":" + config.OAuth.ClientSecret))
	headers["Authorization"] = "Basic " + auth

	return &OAuthRequest{
		URL:      tokenURL,
		FormData: formData.Encode(),
		Headers:  headers,
	}, nil
}

// ParseOAuthTokenResponse parses the OAuth token response
func ParseOAuthTokenResponse(responseBody []byte) (*OAuthToken, error) {
	var token OAuthToken
	if err := json.Unmarshal(responseBody, &token); err != nil {
		return nil, fmt.Errorf("error parsing OAuth token: %w", err)
	}

	// Set expiration time
	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	return &token, nil
}

// SaveOAuthToken saves the OAuth token to config
func SaveOAuthToken(token *OAuthToken) error {
	// Load current config
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	// Update OAuth token
	config.OAuthToken = token

	// Save config
	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	return nil
}

// LoadOAuthToken loads the OAuth token from config
func LoadOAuthToken() (*OAuthToken, error) {
	// Load config
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	if config.OAuthToken == nil {
		return nil, fmt.Errorf("OAuth token not found")
	}

	return config.OAuthToken, nil
}

// IsTokenExpired checks if the OAuth token is expired (with 5 minute buffer)
func (t *OAuthToken) IsExpired() bool {
	return time.Now().Add(5 * time.Minute).After(t.ExpiresAt)
}

// ShouldRefreshToken checks if we need to refresh the OAuth token
func ShouldRefreshToken(config *Config) (bool, error) {
	if config.OAuth == nil || !config.OAuth.Enabled {
		return false, fmt.Errorf("OAuth not enabled")
	}

	// Try to load existing token
	token, err := LoadOAuthToken()
	if err != nil || token.IsExpired() {
		return true, nil // Need to refresh
	}

	return false, nil // Token is still valid
}
