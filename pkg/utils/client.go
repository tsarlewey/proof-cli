// Not yet used to be used
package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type ProofClient struct {
	config     *Config
	httpClient *http.Client
	apiKey     string
	oauthToken *OAuthToken
}

func NewProofClient() (*ProofClient, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	client := &ProofClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}

	// Check if OAuth is enabled
	if config.OAuth != nil && config.OAuth.Enabled {
		// Use OAuth authentication - get token on first use to avoid unnecessary calls
		// Token will be retrieved when needed in getValidOAuthToken()
	} else {
		// Use API key authentication
		apiKey, err := GetAPIKey()
		if err != nil {
			return nil, fmt.Errorf("failed to get API key: %w", err)
		}
		client.apiKey = apiKey
	}

	return client, nil
}

// AuthenticateOAuth performs OAuth authentication and returns a token
func (c *ProofClient) AuthenticateOAuth() (*OAuthToken, error) {
	req, err := PrepareOAuthTokenRequest(c.config)
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", req.URL, strings.NewReader(req.FormData))
	if err != nil {
		return nil, fmt.Errorf("error creating OAuth request: %w", err)
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Make request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making OAuth request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading OAuth response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OAuth authentication failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse token response using helper from oauth.go
	token, err := ParseOAuthTokenResponse(respBody)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// getValidOAuthToken gets a valid OAuth token, refreshing if necessary
func (c *ProofClient) getValidOAuthToken() (*OAuthToken, error) {
	needsRefresh, err := ShouldRefreshToken(c.config)
	if err != nil {
		return nil, err
	}

	if needsRefresh {
		// Get new token
		token, err := c.AuthenticateOAuth()
		if err != nil {
			return nil, fmt.Errorf("error getting OAuth token: %w", err)
		}

		// Save the new token
		if err := SaveOAuthToken(token); err != nil {
			return nil, fmt.Errorf("error saving OAuth token: %w", err)
		}

		return token, nil
	}

	// Load existing valid token
	return LoadOAuthToken()
}

// TestOAuthAuthentication tests OAuth authentication by getting a fresh token
func (c *ProofClient) TestOAuthAuthentication() (*OAuthToken, error) {
	// Force a fresh OAuth token request
	token, err := c.AuthenticateOAuth()
	if err != nil {
		return nil, fmt.Errorf("error testing OAuth authentication: %w", err)
	}

	// Save the new token
	if err := SaveOAuthToken(token); err != nil {
		return nil, fmt.Errorf("error saving OAuth token: %w", err)
	}

	return token, nil
}

// RequestOptions allows customizing requests with additional headers
type RequestOptions struct {
	ContentType string
	Accept      string
}

func (c *ProofClient) Request(method, path string, body any, opts ...*RequestOptions) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.config.APIEndpoint, path)

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication header
	if c.config.OAuth != nil && c.config.OAuth.Enabled {
		// Use OAuth authentication
		token, err := c.getValidOAuthToken()
		if err != nil {
			return nil, fmt.Errorf("failed to get OAuth token: %w", err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	} else {
		req.Header.Set("ApiKey", c.apiKey)
	}

	// Set content type from options or use default
	contentType := "application/json"
	if len(opts) > 0 && opts[0] != nil && opts[0].ContentType != "" {
		contentType = opts[0].ContentType
	}
	req.Header.Set("Content-Type", contentType)

	// Set accept header from options or use default
	accept := "application/json"
	if len(opts) > 0 && opts[0] != nil && opts[0].Accept != "" {
		accept = opts[0].Accept
	}
	req.Header.Set("Accept", accept)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *ProofClient) Get(path string, opts ...*RequestOptions) ([]byte, error) {
	return c.Request("GET", path, nil, opts...)
}

func (c *ProofClient) Post(path string, body any, opts ...*RequestOptions) ([]byte, error) {
	return c.Request("POST", path, body, opts...)
}

func (c *ProofClient) Put(path string, body any, opts ...*RequestOptions) ([]byte, error) {
	return c.Request("PUT", path, body, opts...)
}

func (c *ProofClient) Patch(path string, body any, opts ...*RequestOptions) ([]byte, error) {
	return c.Request("PATCH", path, body, opts...)
}

func (c *ProofClient) Delete(path string, opts ...*RequestOptions) ([]byte, error) {
	return c.Request("DELETE", path, nil, opts...)
}

func HandleError(err error, message string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", message, err)
		os.Exit(1)
	}
}

func Must[T any](val T, err error) T {
	if err != nil {
		HandleError(err, "Operation failed")
	}
	return val
}
