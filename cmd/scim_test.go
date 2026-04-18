package cmd

import (
	"encoding/json"
	"testing"

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
