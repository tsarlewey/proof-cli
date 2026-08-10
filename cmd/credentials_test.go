package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsarlewey/proof-sdk-go/credentials"
)

// authorizeFlags builds a command carrying the authorize-url flag set, with
// the given overrides applied.
func authorizeFlags(t *testing.T, overrides map[string]string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	registerAuthorizeFlags(cmd)
	for name, value := range overrides {
		require.NoError(t, cmd.Flags().Set(name, value))
	}
	return cmd
}

func TestBuildAuthorizeParams_FragmentMode(t *testing.T) {
	cmd := authorizeFlags(t, map[string]string{
		"client-id":     "client-123",
		"response-mode": "fragment",
		"redirect-uri":  "https://app.example.com/cb",
		"scope":         "openid",
		"login-hint":    "user@example.com",
		"nonce":         "nonce-abc",
		"state":         "state-xyz",
	})

	params, err := buildAuthorizeParams(cmd)
	require.NoError(t, err)

	assert.Equal(t, "client-123", params.ClientId)
	assert.Equal(t, credentials.AuthorizeVerifiableCredentialPresentationParamsResponseType("vp_token"), params.ResponseType)
	require.NotNil(t, params.RedirectUri)
	assert.Equal(t, "https://app.example.com/cb", *params.RedirectUri)
	assert.Nil(t, params.ResponseUri)
	require.NotNil(t, params.State)
	assert.Equal(t, "state-xyz", *params.State)
}

func TestBuildAuthorizeParams_DirectPostMode(t *testing.T) {
	cmd := authorizeFlags(t, map[string]string{
		"client-id":     "client-123",
		"response-mode": "direct_post",
		"response-uri":  "https://app.example.com/post",
		"scope":         "openid",
		"login-hint":    "user@example.com",
		"nonce":         "nonce-abc",
	})

	params, err := buildAuthorizeParams(cmd)
	require.NoError(t, err)

	require.NotNil(t, params.ResponseUri)
	assert.Equal(t, "https://app.example.com/post", *params.ResponseUri)
	assert.Nil(t, params.RedirectUri)
	assert.Nil(t, params.State, "state should stay unset when the flag is empty")
}

// The API rejects the wrong URI for a given response mode, so the CLI catches
// each mismatch before building a request.
func TestBuildAuthorizeParams_RejectsMismatchedURIs(t *testing.T) {
	cases := []struct {
		name      string
		overrides map[string]string
		wantErr   string
	}{
		{
			name:      "fragment without redirect-uri",
			overrides: map[string]string{"response-mode": "fragment"},
			wantErr:   "--redirect-uri is required",
		},
		{
			name:      "fragment with response-uri",
			overrides: map[string]string{"response-mode": "fragment", "redirect-uri": "https://a", "response-uri": "https://b"},
			wantErr:   "--response-uri is not permitted",
		},
		{
			name:      "direct_post without response-uri",
			overrides: map[string]string{"response-mode": "direct_post"},
			wantErr:   "--response-uri is required",
		},
		{
			name:      "direct_post with redirect-uri",
			overrides: map[string]string{"response-mode": "direct_post", "response-uri": "https://a", "redirect-uri": "https://b"},
			wantErr:   "--redirect-uri is not permitted",
		},
		{
			name:      "unknown response mode",
			overrides: map[string]string{"response-mode": "query"},
			wantErr:   "invalid --response-mode",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := authorizeFlags(t, tc.overrides)
			params, err := buildAuthorizeParams(cmd)
			assert.Nil(t, params)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// The authorize endpoint is a browser redirect, so the useful output is a URL
// carrying every parameter as a query string.
func TestAuthorizeURLContainsAllParams(t *testing.T) {
	cmd := authorizeFlags(t, map[string]string{
		"client-id":     "client-123",
		"response-mode": "fragment",
		"redirect-uri":  "https://app.example.com/cb",
		"scope":         "openid vc",
		"login-hint":    "user@example.com",
		"nonce":         "nonce-abc",
		"state":         "state-xyz",
	})
	params, err := buildAuthorizeParams(cmd)
	require.NoError(t, err)

	req, err := credentials.NewAuthorizeVerifiableCredentialPresentationRequest("https://api.proof.com", params)
	require.NoError(t, err)

	assert.Equal(t, "/verifiable-credentials/v1/presentation/authorize", req.URL.Path)
	q := req.URL.Query()
	assert.Equal(t, "client-123", q.Get("client_id"))
	assert.Equal(t, "vp_token", q.Get("response_type"))
	assert.Equal(t, "fragment", q.Get("response_mode"))
	assert.Equal(t, "https://app.example.com/cb", q.Get("redirect_uri"))
	assert.Equal(t, "openid vc", q.Get("scope"))
	assert.Equal(t, "user@example.com", q.Get("login_hint"))
	assert.Equal(t, "nonce-abc", q.Get("nonce"))
	assert.Equal(t, "state-xyz", q.Get("state"))
}
