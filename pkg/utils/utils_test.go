package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
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
// helpers.go tests
// ============================================================================

func TestBuildQueryParams_NilInput(t *testing.T) {
	result := BuildQueryParams(nil)
	assert.Empty(t, result)
}

func TestBuildQueryParams_NonStruct(t *testing.T) {
	result := BuildQueryParams("not a struct")
	assert.Empty(t, result)
}

func TestBuildQueryParams_NilPointer(t *testing.T) {
	var nilPtr *struct{ Name string }
	result := BuildQueryParams(nilPtr)
	assert.Empty(t, result)
}

func TestBuildQueryParams_StringFields(t *testing.T) {
	type Params struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Empty string `json:"empty"`
	}
	params := Params{Name: "john", Email: "john@example.com", Empty: ""}

	result := BuildQueryParams(params)

	assert.Equal(t, "john", result.Get("name"))
	assert.Equal(t, "john@example.com", result.Get("email"))
	assert.Empty(t, result.Get("empty")) // empty strings are skipped
}

func TestBuildQueryParams_IntFields(t *testing.T) {
	type Params struct {
		Page  int   `json:"page"`
		Limit int   `json:"limit"`
		Zero  int   `json:"zero"`
		Big   int64 `json:"big"`
	}
	params := Params{Page: 1, Limit: 10, Zero: 0, Big: 999999}

	result := BuildQueryParams(params)

	assert.Equal(t, "1", result.Get("page"))
	assert.Equal(t, "10", result.Get("limit"))
	assert.Empty(t, result.Get("zero")) // zero ints are skipped
	assert.Equal(t, "999999", result.Get("big"))
}

func TestBuildQueryParams_PointerToStruct(t *testing.T) {
	type Params struct {
		Name string `json:"name"`
	}
	params := &Params{Name: "john"}

	result := BuildQueryParams(params)

	assert.Equal(t, "john", result.Get("name"))
}

func TestBuildQueryParams_TimePointer(t *testing.T) {
	type Params struct {
		CreatedAt *time.Time `json:"created_at"`
	}
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	params := Params{CreatedAt: &now}

	result := BuildQueryParams(params)

	assert.Equal(t, "2025-01-15T10:30:00Z", result.Get("created_at"))
}

func TestBuildQueryParams_SkipsUntaggedFields(t *testing.T) {
	type Params struct {
		Tagged   string `json:"tagged"`
		Untagged string
	}
	params := Params{Tagged: "value", Untagged: "ignored"}

	result := BuildQueryParams(params)

	assert.Equal(t, "value", result.Get("tagged"))
	assert.Empty(t, result.Get("Untagged"))
}

func TestBuildQueryParams_SkipsIgnoredFields(t *testing.T) {
	type Params struct {
		Included string `json:"included"`
		Ignored  string `json:"-"`
	}
	params := Params{Included: "value", Ignored: "ignored"}

	result := BuildQueryParams(params)

	assert.Equal(t, "value", result.Get("included"))
	assert.Empty(t, result.Get("Ignored"))
}

func TestBuildQueryParams_HandlesOmitempty(t *testing.T) {
	type Params struct {
		Name string `json:"name,omitempty"`
	}
	params := Params{Name: "john"}

	result := BuildQueryParams(params)

	assert.Equal(t, "john", result.Get("name"))
}

func TestFormatOutput(t *testing.T) {
	// FormatOutput is a stub that returns empty string
	result, err := FormatOutput("data", "json")
	assert.NoError(t, err)
	assert.Empty(t, result)
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

func TestProofClient_Request_WithAPIKey(t *testing.T) {
	// Create a mock transport
	mockTransport := &MockRoundTripper{
		Response: mockJSONResponse(200, map[string]string{"status": "ok"}),
	}

	// Create client with API key auth
	client := &ProofClient{
		config: &Config{
			APIEndpoint: "https://api.proof.com",
			OAuth:       nil, // No OAuth, use API key
		},
		httpClient: &http.Client{Transport: mockTransport},
		apiKey:     "test-api-key",
	}

	resp, err := client.Get("/test")

	require.NoError(t, err)
	assert.Contains(t, string(resp), "ok")
	assert.Equal(t, "test-api-key", mockTransport.LastReq.Header.Get("ApiKey"))
	assert.Equal(t, "application/json", mockTransport.LastReq.Header.Get("Content-Type"))
}

func TestProofClient_Request_WithBody(t *testing.T) {
	mockTransport := &MockRoundTripper{
		Response: mockJSONResponse(201, map[string]string{"id": "123"}),
	}

	client := &ProofClient{
		config: &Config{
			APIEndpoint: "https://api.proof.com",
		},
		httpClient: &http.Client{Transport: mockTransport},
		apiKey:     "test-api-key",
	}

	body := map[string]string{"name": "test"}
	resp, err := client.Post("/items", body)

	require.NoError(t, err)
	assert.Contains(t, string(resp), "123")
	assert.Equal(t, "POST", mockTransport.LastReq.Method)
}

func TestProofClient_Request_CustomContentType(t *testing.T) {
	mockTransport := &MockRoundTripper{
		Response: mockJSONResponse(200, map[string]string{}),
	}

	client := &ProofClient{
		config: &Config{
			APIEndpoint: "https://api.proof.com",
		},
		httpClient: &http.Client{Transport: mockTransport},
		apiKey:     "test-api-key",
	}

	opts := &RequestOptions{
		ContentType: "application/scim+json",
		Accept:      "application/scim+json",
	}
	_, err := client.Get("/scim/users", opts)

	require.NoError(t, err)
	assert.Equal(t, "application/scim+json", mockTransport.LastReq.Header.Get("Content-Type"))
	assert.Equal(t, "application/scim+json", mockTransport.LastReq.Header.Get("Accept"))
}

func TestProofClient_Request_APIError(t *testing.T) {
	mockTransport := &MockRoundTripper{
		Response: mockJSONResponse(400, map[string]string{"error": "bad request"}),
	}

	client := &ProofClient{
		config: &Config{
			APIEndpoint: "https://api.proof.com",
		},
		httpClient: &http.Client{Transport: mockTransport},
		apiKey:     "test-api-key",
	}

	_, err := client.Get("/invalid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error (status 400)")
}

func TestProofClient_HTTPMethods(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		call           func(c *ProofClient) ([]byte, error)
		expectedMethod string
	}{
		{
			name:           "GET",
			call:           func(c *ProofClient) ([]byte, error) { return c.Get("/test") },
			expectedMethod: "GET",
		},
		{
			name:           "POST",
			call:           func(c *ProofClient) ([]byte, error) { return c.Post("/test", nil) },
			expectedMethod: "POST",
		},
		{
			name:           "PUT",
			call:           func(c *ProofClient) ([]byte, error) { return c.Put("/test", nil) },
			expectedMethod: "PUT",
		},
		{
			name:           "PATCH",
			call:           func(c *ProofClient) ([]byte, error) { return c.Patch("/test", nil) },
			expectedMethod: "PATCH",
		},
		{
			name:           "DELETE",
			call:           func(c *ProofClient) ([]byte, error) { return c.Delete("/test") },
			expectedMethod: "DELETE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport := &MockRoundTripper{
				Response: mockJSONResponse(200, map[string]string{}),
			}

			client := &ProofClient{
				config: &Config{
					APIEndpoint: "https://api.proof.com",
				},
				httpClient: &http.Client{Transport: mockTransport},
				apiKey:     "test-api-key",
			}

			_, err := tt.call(client)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedMethod, mockTransport.LastReq.Method)
		})
	}
}

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
// URL building tests
// ============================================================================

func TestBuildQueryParams_URLEncoding(t *testing.T) {
	type Params struct {
		Query string `json:"q"`
	}
	params := Params{Query: "hello world"}

	result := BuildQueryParams(params)

	// url.Values.Encode() should encode spaces
	encoded := result.Encode()
	assert.Contains(t, encoded, "q=hello+world")
}

func TestBuildQueryParams_SpecialCharacters(t *testing.T) {
	type Params struct {
		Email string `json:"email"`
	}
	params := Params{Email: "user@example.com"}

	result := BuildQueryParams(params)

	encoded := result.Encode()
	// @ should be URL encoded
	decodedEmail, err := url.QueryUnescape(result.Get("email"))
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", decodedEmail)
	_ = encoded
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

func TestProofClient_Request_WithOAuth(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// First, set up a config with a valid OAuth token already saved
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
		AccessToken: "saved-access-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	err = SaveOAuthToken(token)
	require.NoError(t, err)

	// Now create a client that will use this token
	mockTransport := &MockRoundTripper{
		Response: mockJSONResponse(200, map[string]string{"status": "ok"}),
	}

	client := &ProofClient{
		config:     config,
		httpClient: &http.Client{Transport: mockTransport},
	}

	resp, err := client.Get("/test")

	require.NoError(t, err)
	assert.Contains(t, string(resp), "ok")
	// Verify OAuth Bearer token was used
	assert.Equal(t, "Bearer saved-access-token", mockTransport.LastReq.Header.Get("Authorization"))
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

// ============================================================================
// Error handling tests
// ============================================================================

func TestProofClient_Request_NetworkError(t *testing.T) {
	mockTransport := &MockRoundTripper{
		Response: nil,
		Err:      assert.AnError,
	}

	client := &ProofClient{
		config: &Config{
			APIEndpoint: "https://api.proof.com",
		},
		httpClient: &http.Client{Transport: mockTransport},
		apiKey:     "test-api-key",
	}

	_, err := client.Get("/test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute request")
}

func TestProofClient_Request_ServerErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"400 Bad Request", 400},
		{"401 Unauthorized", 401},
		{"403 Forbidden", 403},
		{"404 Not Found", 404},
		{"500 Internal Server Error", 500},
		{"502 Bad Gateway", 502},
		{"503 Service Unavailable", 503},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport := &MockRoundTripper{
				Response: mockJSONResponse(tt.statusCode, map[string]string{"error": "error"}),
			}

			client := &ProofClient{
				config: &Config{
					APIEndpoint: "https://api.proof.com",
				},
				httpClient: &http.Client{Transport: mockTransport},
				apiKey:     "test-api-key",
			}

			_, err := client.Get("/test")

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "API error")
		})
	}
}

func TestProofClient_Request_SuccessStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"200 OK", 200},
		{"201 Created", 201},
		{"204 No Content", 204},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport := &MockRoundTripper{
				Response: mockJSONResponse(tt.statusCode, map[string]string{}),
			}

			client := &ProofClient{
				config: &Config{
					APIEndpoint: "https://api.proof.com",
				},
				httpClient: &http.Client{Transport: mockTransport},
				apiKey:     "test-api-key",
			}

			_, err := client.Get("/test")

			assert.NoError(t, err)
		})
	}
}

// ============================================================================
// Edge cases and additional coverage
// ============================================================================

func TestBuildQueryParams_NilTimePointer(t *testing.T) {
	type Params struct {
		CreatedAt *time.Time `json:"created_at"`
	}
	params := Params{CreatedAt: nil}

	result := BuildQueryParams(params)

	assert.Empty(t, result.Get("created_at"))
}

func TestBuildQueryParams_AllIntTypes(t *testing.T) {
	type Params struct {
		Int8Val  int8  `json:"int8"`
		Int16Val int16 `json:"int16"`
		Int32Val int32 `json:"int32"`
		Int64Val int64 `json:"int64"`
	}
	params := Params{
		Int8Val:  8,
		Int16Val: 16,
		Int32Val: 32,
		Int64Val: 64,
	}

	result := BuildQueryParams(params)

	assert.Equal(t, "8", result.Get("int8"))
	assert.Equal(t, "16", result.Get("int16"))
	assert.Equal(t, "32", result.Get("int32"))
	assert.Equal(t, "64", result.Get("int64"))
}

func TestProofClient_Request_MarshalBody(t *testing.T) {
	mockTransport := &MockRoundTripper{
		Response: mockJSONResponse(200, map[string]string{}),
	}

	client := &ProofClient{
		config: &Config{
			APIEndpoint: "https://api.proof.com",
		},
		httpClient: &http.Client{Transport: mockTransport},
		apiKey:     "test-api-key",
	}

	body := struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}{
		Name:  "test",
		Count: 42,
	}

	_, err := client.Post("/items", body)

	require.NoError(t, err)

	// Verify body was marshaled correctly
	bodyBytes, _ := io.ReadAll(mockTransport.LastReq.Body)
	assert.Contains(t, string(bodyBytes), "test")
	assert.Contains(t, string(bodyBytes), "42")
}

func TestProofClient_Request_NilOptions(t *testing.T) {
	mockTransport := &MockRoundTripper{
		Response: mockJSONResponse(200, map[string]string{}),
	}

	client := &ProofClient{
		config: &Config{
			APIEndpoint: "https://api.proof.com",
		},
		httpClient: &http.Client{Transport: mockTransport},
		apiKey:     "test-api-key",
	}

	// Pass nil options explicitly
	_, err := client.Get("/test", nil)

	require.NoError(t, err)
	assert.Equal(t, "application/json", mockTransport.LastReq.Header.Get("Content-Type"))
	assert.Equal(t, "application/json", mockTransport.LastReq.Header.Get("Accept"))
}

func TestProofClient_Request_EmptyOptions(t *testing.T) {
	mockTransport := &MockRoundTripper{
		Response: mockJSONResponse(200, map[string]string{}),
	}

	client := &ProofClient{
		config: &Config{
			APIEndpoint: "https://api.proof.com",
		},
		httpClient: &http.Client{Transport: mockTransport},
		apiKey:     "test-api-key",
	}

	// Pass empty options
	opts := &RequestOptions{}
	_, err := client.Get("/test", opts)

	require.NoError(t, err)
	// Should use defaults
	assert.Equal(t, "application/json", mockTransport.LastReq.Header.Get("Content-Type"))
	assert.Equal(t, "application/json", mockTransport.LastReq.Header.Get("Accept"))
}
