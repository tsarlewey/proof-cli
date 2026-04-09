package scim

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ptr returns a pointer to the given value
func ptr[T any](v T) *T {
	return &v
}

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

// mockJSONResponse creates a mock HTTP response with the given status code and body
func mockJSONResponse(statusCode int, body any) *http.Response {
	jsonBytes, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

// TestNewClient verifies client creation
func TestNewClient(t *testing.T) {
	t.Run("creates client with default http client", func(t *testing.T) {
		client, err := NewClient("https://api.example.com")
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "https://api.example.com/", client.Server)
		assert.NotNil(t, client.Client)
	})

	t.Run("adds trailing slash to server URL", func(t *testing.T) {
		client, err := NewClient("https://api.example.com")
		require.NoError(t, err)
		assert.Equal(t, "https://api.example.com/", client.Server)
	})

	t.Run("preserves existing trailing slash", func(t *testing.T) {
		client, err := NewClient("https://api.example.com/")
		require.NoError(t, err)
		assert.Equal(t, "https://api.example.com/", client.Server)
	})

	t.Run("accepts custom HTTP client", func(t *testing.T) {
		customClient := &http.Client{}
		client, err := NewClient("https://api.example.com", WithHTTPClient(customClient))
		require.NoError(t, err)
		assert.Equal(t, customClient, client.Client)
	})
}

// TestNewClientWithResponses verifies response client creation
func TestNewClientWithResponses(t *testing.T) {
	client, err := NewClientWithResponses("https://api.example.com")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

// Note: SCIM SDK has confusingly named methods due to poor operationIds in the OpenAPI spec.
// The mapping is:
// - RetrieveResourceTypesCopy: List Users (GET)
// - CreateUserCopy: Get User (GET)
// - CreateUserCopy1: Update User (PUT)
// - ReplaceUserCopy: Patch User (PATCH)
// - ReplaceUserCopy1: Delete User (DELETE)
// - RetrieveUsersSchemaCopy: Get Resource Types (GET)

// TestNewRetrieveResourceTypesCopyRequest verifies list users request generation
func TestNewRetrieveResourceTypesCopyRequest(t *testing.T) {
	t.Run("generates request without params", func(t *testing.T) {
		req, err := NewRetrieveResourceTypesCopyRequest("https://api.example.com/", "org-123", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/Users")
	})

	t.Run("includes startIndex parameter", func(t *testing.T) {
		params := &RetrieveResourceTypesCopyParams{
			StartIndex: ptr(int32(1)),
		}
		req, err := NewRetrieveResourceTypesCopyRequest("https://api.example.com/", "org-123", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "startIndex=1")
	})

	t.Run("includes count parameter", func(t *testing.T) {
		params := &RetrieveResourceTypesCopyParams{
			Count: ptr(int32(50)),
		}
		req, err := NewRetrieveResourceTypesCopyRequest("https://api.example.com/", "org-123", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "count=50")
	})

	t.Run("includes filter parameter", func(t *testing.T) {
		params := &RetrieveResourceTypesCopyParams{
			Filter: ptr(`userName eq "test@example.com"`),
		}
		req, err := NewRetrieveResourceTypesCopyRequest("https://api.example.com/", "org-123", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "filter=")
	})

	t.Run("includes multiple parameters", func(t *testing.T) {
		params := &RetrieveResourceTypesCopyParams{
			StartIndex: ptr(int32(1)),
			Count:      ptr(int32(100)),
		}
		req, err := NewRetrieveResourceTypesCopyRequest("https://api.example.com/", "org-123", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "startIndex=1")
		assert.Contains(t, req.URL.RawQuery, "count=100")
	})
}

// TestNewCreateUserCopyRequest verifies get user request generation (confusing name due to bad operationId)
func TestNewCreateUserCopyRequest(t *testing.T) {
	t.Run("generates GET request for specific user", func(t *testing.T) {
		req, err := NewCreateUserCopyRequest("https://api.example.com/", "org-123", "user-456", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/Users/user-456")
	})

	t.Run("includes accept header parameter", func(t *testing.T) {
		params := &CreateUserCopyParams{
			Accept: ptr("application/json"),
		}
		req, err := NewCreateUserCopyRequest("https://api.example.com/", "org-123", "user-456", params)
		require.NoError(t, err)
		assert.Equal(t, "application/json", req.Header.Get("accept"))
	})
}

// TestNewReplaceUserCopy1Request verifies delete user request generation (confusing name due to bad operationId)
func TestNewReplaceUserCopy1Request(t *testing.T) {
	t.Run("generates DELETE request for user", func(t *testing.T) {
		req, err := NewReplaceUserCopy1Request("https://api.example.com/", "org-123", "user-456", nil)
		require.NoError(t, err)
		assert.Equal(t, "DELETE", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/Users/user-456")
	})
}

// TestNewReplaceUserCopyRequest verifies patch user request generation (confusing name due to bad operationId)
func TestNewReplaceUserCopyRequest(t *testing.T) {
	t.Run("generates PATCH request for user", func(t *testing.T) {
		body := ReplaceUserCopyJSONRequestBody{
			Operations: struct {
				Op    string  `json:"op"`
				Path  *string `json:"path,omitempty"`
				Value *string `json:"value,omitempty"`
			}{
				Op:    "replace",
				Path:  ptr("active"),
				Value: ptr("false"),
			},
		}
		req, err := NewReplaceUserCopyRequest("https://api.example.com/", "org-123", "user-456", nil, body)
		require.NoError(t, err)
		assert.Equal(t, "PATCH", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/Users/user-456")
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	})
}

// TestNewCreateUserRequest verifies create user request generation
func TestNewCreateUserRequest(t *testing.T) {
	t.Run("generates POST request to create user", func(t *testing.T) {
		body := CreateUserJSONRequestBody{
			UserName: "newuser@example.com",
			Name: &struct {
				FamilyName *string `json:"familyName,omitempty"`
				GivenName  *string `json:"givenName,omitempty"`
			}{
				GivenName:  ptr("John"),
				FamilyName: ptr("Doe"),
			},
			Emails: &[]string{"newuser@example.com"},
		}
		req, err := NewCreateUserRequest("https://api.example.com/", "org-123", nil, body)
		require.NoError(t, err)
		assert.Equal(t, "POST", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/Users")
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	})
}

// TestNewRetrieveUsersSchemaRequest verifies schema retrieval request generation
func TestNewRetrieveUsersSchemaRequest(t *testing.T) {
	t.Run("generates request for user schema", func(t *testing.T) {
		req, err := NewRetrieveUsersSchemaRequest("https://api.example.com/", "org-123")
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/Schemas/Users")
	})
}

// TestNewRetrieveServiceProviderConfigCopyRequest verifies resource types retrieval
func TestNewRetrieveServiceProviderConfigCopyRequest(t *testing.T) {
	t.Run("generates request for resource types", func(t *testing.T) {
		req, err := NewRetrieveServiceProviderConfigCopyRequest("https://api.example.com/", "org-123")
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/ResourceTypes")
	})
}

// TestNewRetrieveUsersSchemaCopyRequest verifies service provider config retrieval
func TestNewRetrieveUsersSchemaCopyRequest(t *testing.T) {
	t.Run("generates request for service provider config", func(t *testing.T) {
		req, err := NewRetrieveUsersSchemaCopyRequest("https://api.example.com/", "org-123")
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/ServiceProviderConfig")
	})
}

// TestParseRetrieveResourceTypesCopyResponse verifies list users response parsing
func TestParseRetrieveResourceTypesCopyResponse(t *testing.T) {
	t.Run("parses 200 response", func(t *testing.T) {
		body := struct {
			Resources    []interface{} `json:"Resources"`
			TotalResults int           `json:"totalResults"`
		}{
			Resources:    []interface{}{},
			TotalResults: 0,
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParseRetrieveResourceTypesCopyResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 200, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON200)
	})

	t.Run("parses 403 error response", func(t *testing.T) {
		body := struct {
			Errors []string `json:"errors"`
		}{
			Errors: []string{"Access denied"},
		}
		resp := mockJSONResponse(403, body)

		parsed, err := ParseRetrieveResourceTypesCopyResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 403, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON403)
	})

	t.Run("parses 404 error response", func(t *testing.T) {
		body := struct {
			Errors []string `json:"errors"`
		}{
			Errors: []string{"Not found"},
		}
		resp := mockJSONResponse(404, body)

		parsed, err := ParseRetrieveResourceTypesCopyResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 404, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON404)
	})

	t.Run("handles empty response body", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 204,
			Status:     http.StatusText(204),
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
			Header:     http.Header{},
		}

		parsed, err := ParseRetrieveResourceTypesCopyResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 204, parsed.StatusCode())
	})
}

// TestResponseStatusMethods verifies status methods on response types
func TestResponseStatusMethods(t *testing.T) {
	testCases := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "RetrieveResourceTypesCopyResponse",
			testFunc: func(t *testing.T) {
				resp := &RetrieveResourceTypesCopyResponse{HTTPResponse: &http.Response{StatusCode: 200, Status: "200 OK"}}
				assert.Equal(t, 200, resp.StatusCode())
				assert.Equal(t, "200 OK", resp.Status())

				nilResp := &RetrieveResourceTypesCopyResponse{}
				assert.Equal(t, 0, nilResp.StatusCode())
				assert.Equal(t, http.StatusText(0), nilResp.Status())
			},
		},
		{
			name: "CreateUserCopyResponse",
			testFunc: func(t *testing.T) {
				resp := &CreateUserCopyResponse{HTTPResponse: &http.Response{StatusCode: 200, Status: "200 OK"}}
				assert.Equal(t, 200, resp.StatusCode())
				assert.Equal(t, "200 OK", resp.Status())

				nilResp := &CreateUserCopyResponse{}
				assert.Equal(t, 0, nilResp.StatusCode())
			},
		},
		{
			name: "ReplaceUserCopy1Response (Delete User)",
			testFunc: func(t *testing.T) {
				resp := &ReplaceUserCopy1Response{HTTPResponse: &http.Response{StatusCode: 204, Status: "204 No Content"}}
				assert.Equal(t, 204, resp.StatusCode())
				assert.Equal(t, "204 No Content", resp.Status())

				nilResp := &ReplaceUserCopy1Response{}
				assert.Equal(t, 0, nilResp.StatusCode())
			},
		},
		{
			name: "ReplaceUserCopyResponse (Patch User)",
			testFunc: func(t *testing.T) {
				resp := &ReplaceUserCopyResponse{HTTPResponse: &http.Response{StatusCode: 200, Status: "200 OK"}}
				assert.Equal(t, 200, resp.StatusCode())
				assert.Equal(t, "200 OK", resp.Status())

				nilResp := &ReplaceUserCopyResponse{}
				assert.Equal(t, 0, nilResp.StatusCode())
			},
		},
		{
			name: "CreateUserResponse",
			testFunc: func(t *testing.T) {
				resp := &CreateUserResponse{HTTPResponse: &http.Response{StatusCode: 201, Status: "201 Created"}}
				assert.Equal(t, 201, resp.StatusCode())
				assert.Equal(t, "201 Created", resp.Status())

				nilResp := &CreateUserResponse{}
				assert.Equal(t, 0, nilResp.StatusCode())
			},
		},
		{
			name: "RetrieveUsersSchemaResponse",
			testFunc: func(t *testing.T) {
				resp := &RetrieveUsersSchemaResponse{HTTPResponse: &http.Response{StatusCode: 200, Status: "200 OK"}}
				assert.Equal(t, 200, resp.StatusCode())
				assert.Equal(t, "200 OK", resp.Status())

				nilResp := &RetrieveUsersSchemaResponse{}
				assert.Equal(t, 0, nilResp.StatusCode())
			},
		},
		{
			name: "RetrieveServiceProviderConfigCopyResponse",
			testFunc: func(t *testing.T) {
				resp := &RetrieveServiceProviderConfigCopyResponse{HTTPResponse: &http.Response{StatusCode: 200, Status: "200 OK"}}
				assert.Equal(t, 200, resp.StatusCode())
				assert.Equal(t, "200 OK", resp.Status())

				nilResp := &RetrieveServiceProviderConfigCopyResponse{}
				assert.Equal(t, 0, nilResp.StatusCode())
			},
		},
		{
			name: "RetrieveUsersSchemaCopyResponse",
			testFunc: func(t *testing.T) {
				resp := &RetrieveUsersSchemaCopyResponse{HTTPResponse: &http.Response{StatusCode: 200, Status: "200 OK"}}
				assert.Equal(t, 200, resp.StatusCode())
				assert.Equal(t, "200 OK", resp.Status())

				nilResp := &RetrieveUsersSchemaCopyResponse{}
				assert.Equal(t, 0, nilResp.StatusCode())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.testFunc)
	}
}

// TestClientWithResponsesMethods verifies client with responses wrapper methods
func TestClientWithResponsesMethods(t *testing.T) {
	mockRT := &MockRoundTripper{
		Response: mockJSONResponse(200, struct {
			Resources []interface{} `json:"Resources"`
		}{
			Resources: []interface{}{},
		}),
	}

	client, err := NewClientWithResponses("https://api.example.com",
		WithHTTPClient(&http.Client{Transport: mockRT}))
	require.NoError(t, err)

	t.Run("RetrieveResourceTypesCopyWithResponse", func(t *testing.T) {
		resp, err := client.RetrieveResourceTypesCopyWithResponse(context.Background(), "org-123", nil)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode())
	})
}

// TestClientMakesRequest verifies client makes correct requests
func TestClientMakesRequest(t *testing.T) {
	t.Run("RetrieveResourceTypesCopy (List Users)", func(t *testing.T) {
		mockRT := &MockRoundTripper{
			Response: mockJSONResponse(200, struct {
				Resources []interface{} `json:"Resources"`
			}{
				Resources: []interface{}{},
			}),
		}

		client, err := NewClient("https://api.example.com",
			WithHTTPClient(&http.Client{Transport: mockRT}))
		require.NoError(t, err)

		_, err = client.RetrieveResourceTypesCopy(context.Background(), "org-123", nil)
		require.NoError(t, err)

		assert.NotNil(t, mockRT.LastReq)
		assert.Equal(t, "GET", mockRT.LastReq.Method)
		assert.Contains(t, mockRT.LastReq.URL.Path, "/org-123/Users")
	})

	t.Run("CreateUserCopy (Get User)", func(t *testing.T) {
		mockRT := &MockRoundTripper{
			Response: mockJSONResponse(200, struct{}{}),
		}

		client, err := NewClient("https://api.example.com",
			WithHTTPClient(&http.Client{Transport: mockRT}))
		require.NoError(t, err)

		_, err = client.CreateUserCopy(context.Background(), "org-123", "user-456", nil)
		require.NoError(t, err)

		assert.NotNil(t, mockRT.LastReq)
		assert.Equal(t, "GET", mockRT.LastReq.Method)
		assert.Contains(t, mockRT.LastReq.URL.Path, "user-456")
	})
}

// TestWithBaseURL verifies base URL override
func TestWithBaseURL(t *testing.T) {
	client, err := NewClient("https://api.example.com",
		WithBaseURL("https://custom.api.example.com/"))
	require.NoError(t, err)
	assert.Equal(t, "https://custom.api.example.com/", client.Server)
}

// TestWithRequestEditorFn verifies request editor functionality
func TestWithRequestEditorFn(t *testing.T) {
	editorCalled := false
	editor := func(ctx context.Context, req *http.Request) error {
		editorCalled = true
		req.Header.Set("X-Custom-Header", "test-value")
		return nil
	}

	mockRT := &MockRoundTripper{
		Response: mockJSONResponse(200, struct {
			Resources []interface{} `json:"Resources"`
		}{
			Resources: []interface{}{},
		}),
	}

	client, err := NewClient("https://api.example.com",
		WithHTTPClient(&http.Client{Transport: mockRT}),
		WithRequestEditorFn(editor))
	require.NoError(t, err)

	_, err = client.RetrieveResourceTypesCopy(context.Background(), "org-123", nil)
	require.NoError(t, err)

	assert.True(t, editorCalled)
	assert.Equal(t, "test-value", mockRT.LastReq.Header.Get("X-Custom-Header"))
}
