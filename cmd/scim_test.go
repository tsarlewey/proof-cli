package cmd

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSCIMPatchOp_Simple(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		expectedOp   string
		expectedPath string
		expectedVal  any
	}{
		{
			name:         "replace active with false",
			input:        "replace:active:false",
			expectedOp:   "replace",
			expectedPath: "active",
			expectedVal:  false,
		},
		{
			name:         "replace active with true",
			input:        "replace:active:true",
			expectedOp:   "replace",
			expectedPath: "active",
			expectedVal:  true,
		},
		{
			name:         "add role plain string",
			input:        "add:roles:admin",
			expectedOp:   "add",
			expectedPath: "roles",
			expectedVal:  "admin",
		},
		{
			name:         "remove path no value",
			input:        "remove:emails",
			expectedOp:   "remove",
			expectedPath: "emails",
			expectedVal:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			op, err := parseSCIMPatchOp(tc.input)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedOp, op.Op)
			assert.Equal(t, tc.expectedPath, op.Path)
			assert.Equal(t, tc.expectedVal, op.Value)
		})
	}
}

func TestParseSCIMPatchOp_JSONValues(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedVal any
	}{
		{"JSON number", "replace:count:42", float64(42)},
		{"JSON quoted string", `replace:name:"John"`, "John"},
		{"JSON array", `replace:roles:["admin","user"]`, []any{"admin", "user"}},
		{"JSON object", `replace:config:{"key":"value"}`, map[string]any{"key": "value"}},
		{"JSON null", "replace:data:null", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			op, err := parseSCIMPatchOp(tc.input)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedVal, op.Value)
		})
	}
}

func TestParseSCIMPatchOp_Invalid(t *testing.T) {
	_, err := parseSCIMPatchOp("invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid operation format")
}

func TestParseSCIMPatchOp_EmptyValueAfterColon(t *testing.T) {
	op, err := parseSCIMPatchOp("replace:active:")
	require.NoError(t, err)
	assert.Equal(t, "replace", op.Op)
	assert.Equal(t, "active", op.Path)
	assert.Equal(t, "", op.Value)
}

func TestParseSCIMPatchOp_ColonInValue(t *testing.T) {
	op, err := parseSCIMPatchOp("replace:url:https://example.com:8080/path")
	require.NoError(t, err)
	assert.Equal(t, "replace", op.Op)
	assert.Equal(t, "url", op.Path)
	assert.Equal(t, "https://example.com:8080/path", op.Value)
}

func TestParseSCIMPatchOp_ComplexPath(t *testing.T) {
	op, err := parseSCIMPatchOp("replace:name.givenName:John")
	require.NoError(t, err)
	assert.Equal(t, "name.givenName", op.Path)
	assert.Equal(t, "John", op.Value)
}

func TestBuildSCIMPatchBody_Empty(t *testing.T) {
	_, err := buildSCIMPatchBody(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one operation")

	_, err = buildSCIMPatchBody([]string{})
	require.Error(t, err)
}

func TestBuildSCIMPatchBody_InvalidOpFails(t *testing.T) {
	_, err := buildSCIMPatchBody([]string{"replace:active:false", "garbage"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid operation format")
}

func TestBuildSCIMPatchBody_ShapeAndRoundTrip(t *testing.T) {
	// The upstream spec models Operations as a single object, but SCIM PATCH
	// requires an array. This test confirms the body actually has an array.
	raw := []string{
		"replace:active:false",
		`add:roles:["admin","user"]`,
		"replace:name.givenName:John",
	}

	body, err := buildSCIMPatchBody(raw)
	require.NoError(t, err)

	var decoded struct {
		Operations []scimPatchOperation `json:"Operations"`
	}
	require.NoError(t, json.Unmarshal(body, &decoded))

	require.Len(t, decoded.Operations, 3)

	assert.Equal(t, "replace", decoded.Operations[0].Op)
	assert.Equal(t, "active", decoded.Operations[0].Path)
	assert.Equal(t, false, decoded.Operations[0].Value)

	assert.Equal(t, "add", decoded.Operations[1].Op)
	assert.Equal(t, "roles", decoded.Operations[1].Path)
	assert.Equal(t, []any{"admin", "user"}, decoded.Operations[1].Value)

	assert.Equal(t, "replace", decoded.Operations[2].Op)
	assert.Equal(t, "name.givenName", decoded.Operations[2].Path)
	assert.Equal(t, "John", decoded.Operations[2].Value)
}

func TestBuildSCIMPatchBody_OmitsNilValueField(t *testing.T) {
	// A "remove" op has no value — the JSON should omit the value field
	// rather than serialize it as null.
	body, err := buildSCIMPatchBody([]string{"remove:emails"})
	require.NoError(t, err)

	assert.NotContains(t, string(body), `"value"`)
	assert.Contains(t, string(body), `"op":"remove"`)
	assert.Contains(t, string(body), `"path":"emails"`)
}

// SCIM models emails and roles as arrays of {value: ...} objects, and active
// as a real bool. A regression to plain string slices would silently send a
// body the API rejects.
func TestBuildSCIMUserBody_Shape(t *testing.T) {
	cmd := &cobra.Command{}
	registerSCIMUserFlags(cmd)
	for name, value := range map[string]string{
		"username":    "user@example.com",
		"given-name":  "Ada",
		"family-name": "Lovelace",
		"email":       "user@example.com",
		"roles":       "admin,employee",
		"external-id": "ext-42",
		"active":      "false",
	} {
		require.NoError(t, cmd.Flags().Set(name, value))
	}

	body := buildSCIMUserBody(cmd)

	assert.Equal(t, "user@example.com", body.UserName)
	require.NotNil(t, body.Active)
	assert.False(t, *body.Active)

	require.NotNil(t, body.Name)
	assert.Equal(t, "Ada", *body.Name.GivenName)
	assert.Equal(t, "Lovelace", *body.Name.FamilyName)

	require.NotNil(t, body.Emails)
	require.Len(t, *body.Emails, 1)
	assert.Equal(t, "user@example.com", *(*body.Emails)[0].Value)

	require.NotNil(t, body.Roles)
	require.Len(t, *body.Roles, 2)
	assert.Equal(t, "admin", *(*body.Roles)[0].Value)
	assert.Equal(t, "employee", *(*body.Roles)[1].Value)

	require.NotNil(t, body.ExternalId)
	assert.Equal(t, "ext-42", *body.ExternalId)
}

func TestBuildSCIMUserBody_OmitsUnsetOptionalFields(t *testing.T) {
	cmd := &cobra.Command{}
	registerSCIMUserFlags(cmd)
	require.NoError(t, cmd.Flags().Set("username", "user@example.com"))

	body := buildSCIMUserBody(cmd)

	assert.Nil(t, body.Name)
	assert.Nil(t, body.Emails)
	assert.Nil(t, body.Roles)
	assert.Nil(t, body.ExternalId)
}
