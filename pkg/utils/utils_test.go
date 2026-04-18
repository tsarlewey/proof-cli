package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRoundTripper is a mock implementation of http.RoundTripper for testing
type MockRoundTripper struct {
	Response *http.Response
	Err      error
	LastReq  *http.Request
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.LastReq = req
	return m.Response, m.Err
}

// mockJSONResponse creates a mock HTTP response with JSON body
func mockJSONResponse(statusCode int, body any) *http.Response {
	jsonBytes, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

// ============================================================================
// oauth.go tests
// ============================================================================

func TestPrepareOAuthTokenRequest_Success(t *testing.T) {
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		OAuth: &OAuthConfig{
			Enabled:      true,
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Scope:        "read write",
		},
	}

	req, err := PrepareOAuthTokenRequest(config)

	require.NoError(t, err)
	assert.Equal(t, "https://api.proof.com/oauth/v2/token", req.URL)
	assert.Contains(t, req.FormData, "grant_type=client_credentials")
	assert.Contains(t, req.FormData, "scope=read+write")
	assert.Equal(t, "application/x-www-form-urlencoded", req.Headers["Content-Type"])
	assert.Equal(t, "application/json", req.Headers["Accept"])
	assert.Contains(t, req.Headers["Authorization"], "Basic ")
}

func TestPrepareOAuthTokenRequest_OAuthDisabled(t *testing.T) {
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		OAuth: &OAuthConfig{
			Enabled: false,
		},
	}

	_, err := PrepareOAuthTokenRequest(config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OAuth not enabled")
}

func TestPrepareOAuthTokenRequest_NilOAuth(t *testing.T) {
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		OAuth:       nil,
	}

	_, err := PrepareOAuthTokenRequest(config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OAuth not enabled")
}

func TestPrepareOAuthTokenRequest_MissingCredentials(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "missing client ID",
			config: &Config{
				APIEndpoint: "https://api.proof.com",
				OAuth: &OAuthConfig{
					Enabled:      true,
					ClientID:     "",
					ClientSecret: "secret",
				},
			},
		},
		{
			name: "missing client secret",
			config: &Config{
				APIEndpoint: "https://api.proof.com",
				OAuth: &OAuthConfig{
					Enabled:      true,
					ClientID:     "client-id",
					ClientSecret: "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PrepareOAuthTokenRequest(tt.config)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "client ID and secret are required")
		})
	}
}

func TestPrepareOAuthTokenRequest_NoScope(t *testing.T) {
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		OAuth: &OAuthConfig{
			Enabled:      true,
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Scope:        "",
		},
	}

	req, err := PrepareOAuthTokenRequest(config)

	require.NoError(t, err)
	assert.NotContains(t, req.FormData, "scope=")
}

func TestParseOAuthTokenResponse_Success(t *testing.T) {
	response := map[string]any{
		"access_token": "test-access-token",
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        "read write",
	}
	responseBody, _ := json.Marshal(response)

	token, err := ParseOAuthTokenResponse(responseBody)

	require.NoError(t, err)
	assert.Equal(t, "test-access-token", token.AccessToken)
	assert.Equal(t, "Bearer", token.TokenType)
	assert.Equal(t, 3600, token.ExpiresIn)
	assert.Equal(t, "read write", token.Scope)
	// ExpiresAt should be set to approximately now + 3600 seconds
	assert.WithinDuration(t, time.Now().Add(3600*time.Second), token.ExpiresAt, 5*time.Second)
}

func TestParseOAuthTokenResponse_InvalidJSON(t *testing.T) {
	_, err := ParseOAuthTokenResponse([]byte("invalid json"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing OAuth token")
}

func TestOAuthToken_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "expired token",
			expiresAt: time.Now().Add(-1 * time.Hour),
			expected:  true,
		},
		{
			name:      "token expiring soon (within 5 min buffer)",
			expiresAt: time.Now().Add(3 * time.Minute),
			expected:  true,
		},
		{
			name:      "valid token",
			expiresAt: time.Now().Add(1 * time.Hour),
			expected:  false,
		},
		{
			name:      "token at exactly 5 min boundary",
			expiresAt: time.Now().Add(5 * time.Minute),
			expected:  true, // buffer is >= 5 minutes
		},
		{
			name:      "token just past 5 min boundary",
			expiresAt: time.Now().Add(6 * time.Minute),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &OAuthToken{
				AccessToken: "test-token",
				ExpiresAt:   tt.expiresAt,
			}
			assert.Equal(t, tt.expected, token.IsExpired())
		})
	}
}

// ============================================================================
// config.go tests - using temp directory
// ============================================================================

// setupTestConfigDir creates a temp directory and sets HOME to use it
func setupTestConfigDir(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "proof-cli-test")
	require.NoError(t, err)

	// Save original HOME
	originalHome := os.Getenv("HOME")

	// Set HOME to temp directory
	os.Setenv("HOME", tempDir)

	// Return cleanup function
	cleanup := func() {
		os.Setenv("HOME", originalHome)
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestLoadConfig_CreatesDefaultConfig(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	config, err := LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, "https://api.proof.com", config.APIEndpoint)
	assert.Equal(t, 30*time.Second, config.Timeout)
}

func TestSaveConfig_AndLoadConfig(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Save config
	config := &Config{
		APIEndpoint: "https://custom.api.com",
		Timeout:     60 * time.Second,
		OAuth: &OAuthConfig{
			Enabled:      true,
			ClientID:     "my-client-id",
			ClientSecret: "my-secret",
			Scope:        "read",
		},
		APIKey: "test-api-key",
	}
	err := SaveConfig(config)
	require.NoError(t, err)

	// Load config
	loaded, err := LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, "https://custom.api.com", loaded.APIEndpoint)
	assert.Equal(t, 60*time.Second, loaded.Timeout)
	assert.True(t, loaded.OAuth.Enabled)
	assert.Equal(t, "my-client-id", loaded.OAuth.ClientID)
	assert.Equal(t, "my-secret", loaded.OAuth.ClientSecret)
	assert.Equal(t, "read", loaded.OAuth.Scope)
	assert.Equal(t, "test-api-key", loaded.APIKey)
}

func TestLoadConfig_FilePermissions(t *testing.T) {
	tempDir, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Load config (creates default)
	_, err := LoadConfig()
	require.NoError(t, err)

	// Check file permissions
	configFile := filepath.Join(tempDir, ".proof-cli", "config.json")
	info, err := os.Stat(configFile)
	require.NoError(t, err)

	// File should have restricted permissions (0600)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestGetAPIKey_FromEnvironment(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Set environment variable
	os.Setenv("PROOF_API_KEY", "env-api-key")
	defer os.Unsetenv("PROOF_API_KEY")

	apiKey, err := GetAPIKey()

	require.NoError(t, err)
	assert.Equal(t, "env-api-key", apiKey)
}

func TestGetAPIKey_FromConfig(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Ensure env var is not set
	os.Unsetenv("PROOF_API_KEY")

	// Save config with API key
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		Timeout:     30 * time.Second,
		APIKey:      "config-api-key",
	}
	err := SaveConfig(config)
	require.NoError(t, err)

	apiKey, err := GetAPIKey()

	require.NoError(t, err)
	assert.Equal(t, "config-api-key", apiKey)
}

func TestGetAPIKey_NotFound(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Ensure env var is not set
	os.Unsetenv("PROOF_API_KEY")

	_, err := GetAPIKey()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key not found")
}

func TestSaveAPIKey(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	err := SaveAPIKey("new-api-key")
	require.NoError(t, err)

	// Verify by loading config
	os.Unsetenv("PROOF_API_KEY")
	apiKey, err := GetAPIKey()
	require.NoError(t, err)
	assert.Equal(t, "new-api-key", apiKey)
}

func TestSaveOAuthToken_AndLoadOAuthToken(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// First load config to create the file
	_, err := LoadConfig()
	require.NoError(t, err)

	token := &OAuthToken{
		AccessToken: "test-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		Scope:       "read write",
	}

	err = SaveOAuthToken(token)
	require.NoError(t, err)

	loaded, err := LoadOAuthToken()
	require.NoError(t, err)
	assert.Equal(t, "test-token", loaded.AccessToken)
	assert.Equal(t, "Bearer", loaded.TokenType)
	assert.Equal(t, 3600, loaded.ExpiresIn)
}

func TestLoadOAuthToken_NotFound(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	_, err := LoadOAuthToken()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OAuth token not found")
}

func TestShouldRefreshToken_OAuthDisabled(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	config := &Config{
		APIEndpoint: "https://api.proof.com",
		OAuth:       nil,
	}

	_, err := ShouldRefreshToken(config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OAuth not enabled")
}

func TestShouldRefreshToken_NoExistingToken(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	config := &Config{
		APIEndpoint: "https://api.proof.com",
		OAuth: &OAuthConfig{
			Enabled: true,
		},
	}

	needsRefresh, err := ShouldRefreshToken(config)

	require.NoError(t, err)
	assert.True(t, needsRefresh) // No token means we need to refresh
}

func TestShouldRefreshToken_ValidToken(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Create config with valid token
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		OAuth: &OAuthConfig{
			Enabled: true,
		},
	}
	err := SaveConfig(config)
	require.NoError(t, err)

	// Save a valid token
	token := &OAuthToken{
		AccessToken: "valid-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour), // Valid for 1 hour
	}
	err = SaveOAuthToken(token)
	require.NoError(t, err)

	needsRefresh, err := ShouldRefreshToken(config)

	require.NoError(t, err)
	assert.False(t, needsRefresh) // Token is valid, no refresh needed
}

func TestShouldRefreshToken_ExpiredToken(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Create config with expired token
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		OAuth: &OAuthConfig{
			Enabled: true,
		},
	}
	err := SaveConfig(config)
	require.NoError(t, err)

	// Save an expired token
	token := &OAuthToken{
		AccessToken: "expired-token",
		ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}
	err = SaveOAuthToken(token)
	require.NoError(t, err)

	needsRefresh, err := ShouldRefreshToken(config)

	require.NoError(t, err)
	assert.True(t, needsRefresh) // Token is expired, needs refresh
}

// ============================================================================
// client.go tests
// ============================================================================

func TestProofClient_AddAuthHeaders_APIKey(t *testing.T) {
	client := &ProofClient{
		config: &Config{
			APIEndpoint: "https://api.proof.com",
			OAuth:       nil,
		},
		apiKey: "test-api-key",
	}

	req, _ := http.NewRequest("GET", "https://api.proof.com/test", nil)
	err := client.AddAuthHeaders(req)

	require.NoError(t, err)
	assert.Equal(t, "test-api-key", req.Header.Get("ApiKey"))
}

func TestProofClient_HTTPClient(t *testing.T) {
	httpClient := &http.Client{Timeout: 60 * time.Second}
	client := &ProofClient{
		httpClient: httpClient,
	}

	assert.Equal(t, httpClient, client.HTTPClient())
}

func TestProofClient_GetConfig(t *testing.T) {
	config := &Config{
		APIEndpoint: "https://custom.api.com",
	}
	client := &ProofClient{
		config: config,
	}

	assert.Equal(t, config, client.GetConfig())
}

func TestNewProofClient_WithAPIKey(t *testing.T) {
	tempDir, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Set up config without OAuth
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		Timeout:     30 * time.Second,
		APIKey:      "stored-api-key",
	}
	err := SaveConfig(config)
	require.NoError(t, err)

	client, err := NewProofClient()

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://api.proof.com", client.config.APIEndpoint)
	assert.Equal(t, "stored-api-key", client.apiKey)
	_ = tempDir // silence unused warning
}

func TestNewProofClient_WithOAuthEnabled(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Set up config with OAuth enabled
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		Timeout:     30 * time.Second,
		OAuth: &OAuthConfig{
			Enabled:      true,
			ClientID:     "test-client",
			ClientSecret: "test-secret",
		},
	}
	err := SaveConfig(config)
	require.NoError(t, err)

	client, err := NewProofClient()

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.True(t, client.config.OAuth.Enabled)
	assert.Empty(t, client.apiKey) // API key not set when OAuth is enabled
}

func TestNewProofClient_NoAPIKey(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Ensure env var is not set
	os.Unsetenv("PROOF_API_KEY")

	// Set up config without OAuth and without API key
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		Timeout:     30 * time.Second,
		OAuth:       nil,
		APIKey:      "",
	}
	err := SaveConfig(config)
	require.NoError(t, err)

	_, err = NewProofClient()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get API key")
}

// ============================================================================
// Config types tests
// ============================================================================

func TestConfig_JSONSerialization(t *testing.T) {
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		Timeout:     30 * time.Second,
		OAuth: &OAuthConfig{
			Enabled:      true,
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			Scope:        "read write",
		},
		APIKey: "api-key",
		OAuthToken: &OAuthToken{
			AccessToken: "token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		},
	}

	jsonBytes, err := json.Marshal(config)
	require.NoError(t, err)

	var decoded Config
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, config.APIEndpoint, decoded.APIEndpoint)
	assert.Equal(t, config.OAuth.ClientID, decoded.OAuth.ClientID)
	assert.Equal(t, config.APIKey, decoded.APIKey)
}

// ============================================================================
// OAuth authentication tests (with mocked HTTP)
// ============================================================================

func TestProofClient_AuthenticateOAuth_Success(t *testing.T) {
	// Create mock transport that returns a valid token response
	tokenResponse := map[string]any{
		"access_token": "mock-access-token",
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        "read write",
	}
	mockTransport := &MockRoundTripper{
		Response: mockJSONResponse(200, tokenResponse),
	}

	client := &ProofClient{
		config: &Config{
			APIEndpoint: "https://api.proof.com",
			OAuth: &OAuthConfig{
				Enabled:      true,
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				Scope:        "read write",
			},
		},
		httpClient: &http.Client{Transport: mockTransport},
	}

	token, err := client.AuthenticateOAuth()

	require.NoError(t, err)
	assert.Equal(t, "mock-access-token", token.AccessToken)
	assert.Equal(t, "Bearer", token.TokenType)
	assert.Equal(t, 3600, token.ExpiresIn)

	// Verify request was correct
	assert.Equal(t, "POST", mockTransport.LastReq.Method)
	assert.Contains(t, mockTransport.LastReq.URL.String(), "/oauth/v2/token")
	assert.Contains(t, mockTransport.LastReq.Header.Get("Authorization"), "Basic ")
}

func TestProofClient_AuthenticateOAuth_Failure(t *testing.T) {
	// Create mock transport that returns an error response
	mockTransport := &MockRoundTripper{
		Response: mockJSONResponse(401, map[string]string{"error": "invalid_client"}),
	}

	client := &ProofClient{
		config: &Config{
			APIEndpoint: "https://api.proof.com",
			OAuth: &OAuthConfig{
				Enabled:      true,
				ClientID:     "test-client-id",
				ClientSecret: "wrong-secret",
				Scope:        "read",
			},
		},
		httpClient: &http.Client{Transport: mockTransport},
	}

	_, err := client.AuthenticateOAuth()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OAuth authentication failed")
	assert.Contains(t, err.Error(), "401")
}

func TestProofClient_AuthenticateOAuth_OAuthDisabled(t *testing.T) {
	client := &ProofClient{
		config: &Config{
			APIEndpoint: "https://api.proof.com",
			OAuth:       nil,
		},
		httpClient: &http.Client{},
	}

	_, err := client.AuthenticateOAuth()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OAuth not enabled")
}

func TestProofClient_AuthenticateOAuth_NetworkError(t *testing.T) {
	mockTransport := &MockRoundTripper{
		Response: nil,
		Err:      assert.AnError,
	}

	client := &ProofClient{
		config: &Config{
			APIEndpoint: "https://api.proof.com",
			OAuth: &OAuthConfig{
				Enabled:      true,
				ClientID:     "test-client-id",
				ClientSecret: "test-secret",
			},
		},
		httpClient: &http.Client{Transport: mockTransport},
	}

	_, err := client.AuthenticateOAuth()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error making OAuth request")
}

func TestProofClient_AddAuthHeaders_OAuth(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Set up config with valid OAuth token
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		OAuth: &OAuthConfig{
			Enabled:      true,
			ClientID:     "test-client",
			ClientSecret: "test-secret",
		},
	}
	err := SaveConfig(config)
	require.NoError(t, err)

	// Save a valid token
	token := &OAuthToken{
		AccessToken: "oauth-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	err = SaveOAuthToken(token)
	require.NoError(t, err)

	client := &ProofClient{
		config:     config,
		httpClient: &http.Client{},
	}

	req, _ := http.NewRequest("GET", "https://api.proof.com/test", nil)
	err = client.AddAuthHeaders(req)

	require.NoError(t, err)
	assert.Equal(t, "Bearer oauth-token", req.Header.Get("Authorization"))
}

func TestProofClient_getValidOAuthToken_RefreshNeeded(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Set up config with expired token
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		OAuth: &OAuthConfig{
			Enabled:      true,
			ClientID:     "test-client",
			ClientSecret: "test-secret",
		},
	}
	err := SaveConfig(config)
	require.NoError(t, err)

	// Save an expired token
	expiredToken := &OAuthToken{
		AccessToken: "expired-token",
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}
	err = SaveOAuthToken(expiredToken)
	require.NoError(t, err)

	// Create mock transport for token refresh
	newTokenResponse := map[string]any{
		"access_token": "new-access-token",
		"token_type":   "Bearer",
		"expires_in":   3600,
	}
	mockTransport := &MockRoundTripper{
		Response: mockJSONResponse(200, newTokenResponse),
	}

	client := &ProofClient{
		config:     config,
		httpClient: &http.Client{Transport: mockTransport},
	}

	token, err := client.getValidOAuthToken()

	require.NoError(t, err)
	assert.Equal(t, "new-access-token", token.AccessToken)
}

func TestProofClient_getValidOAuthToken_UseExisting(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Set up config with valid token
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		OAuth: &OAuthConfig{
			Enabled:      true,
			ClientID:     "test-client",
			ClientSecret: "test-secret",
		},
	}
	err := SaveConfig(config)
	require.NoError(t, err)

	// Save a valid token
	validToken := &OAuthToken{
		AccessToken: "valid-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	err = SaveOAuthToken(validToken)
	require.NoError(t, err)

	client := &ProofClient{
		config:     config,
		httpClient: &http.Client{},
	}

	token, err := client.getValidOAuthToken()

	require.NoError(t, err)
	assert.Equal(t, "valid-token", token.AccessToken)
}

func TestProofClient_TestOAuthAuthentication(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Set up config
	config := &Config{
		APIEndpoint: "https://api.proof.com",
		OAuth: &OAuthConfig{
			Enabled:      true,
			ClientID:     "test-client",
			ClientSecret: "test-secret",
		},
	}
	err := SaveConfig(config)
	require.NoError(t, err)

	// Create mock transport
	tokenResponse := map[string]any{
		"access_token": "test-token",
		"token_type":   "Bearer",
		"expires_in":   3600,
	}
	mockTransport := &MockRoundTripper{
		Response: mockJSONResponse(200, tokenResponse),
	}

	client := &ProofClient{
		config:     config,
		httpClient: &http.Client{Transport: mockTransport},
	}

	token, err := client.TestOAuthAuthentication()

	require.NoError(t, err)
	assert.Equal(t, "test-token", token.AccessToken)

	// Verify token was saved
	loaded, err := LoadOAuthToken()
	require.NoError(t, err)
	assert.Equal(t, "test-token", loaded.AccessToken)
}
