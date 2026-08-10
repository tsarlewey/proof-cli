package cmd

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The generated structs carry the upstream spec's literal "common_name *required"
// and "csr *required" keys, so this builder exists to emit the real wire names.
// If it ever regresses to the generated struct, these assertions fail.
func TestCertCreateBody_UsesRealWireKeys(t *testing.T) {
	t.Run("common_name with profile", func(t *testing.T) {
		raw, err := certCreateBody("organization_authenticity_al2", "common_name", "Acme Signing Authority")
		require.NoError(t, err)

		var got map[string]string
		require.NoError(t, json.Unmarshal(raw, &got))
		assert.Equal(t, map[string]string{
			"common_name":         "Acme Signing Authority",
			"certificate_profile": "organization_authenticity_al2",
		}, got)
	})

	t.Run("csr without profile omits the key", func(t *testing.T) {
		raw, err := certCreateBody("", "csr", "-----BEGIN CERTIFICATE REQUEST-----")
		require.NoError(t, err)

		var got map[string]string
		require.NoError(t, json.Unmarshal(raw, &got))
		assert.Equal(t, map[string]string{"csr": "-----BEGIN CERTIFICATE REQUEST-----"}, got)
		assert.NotContains(t, got, "certificate_profile")
	})

	t.Run("never emits the spec's asterisk-suffixed keys", func(t *testing.T) {
		raw, err := certCreateBody("organization_authenticity_al1", "common_name", "Acme")
		require.NoError(t, err)
		assert.NotContains(t, string(raw), "*required")
	})
}
