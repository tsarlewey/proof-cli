package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PatchOperation represents a SCIM PATCH operation for testing
type PatchOperation struct {
	Op    string `json:"op"`
	Path  string `json:"path,omitempty"`
	Value any    `json:"value,omitempty"`
}

// parsePatchOperation parses a string in "op:path:value" format
// This is the same logic used in scimPatchUserCmd
func parsePatchOperation(op string) (*PatchOperation, error) {
	parts := strings.SplitN(op, ":", 3)
	if len(parts) < 2 {
		return nil, nil // Invalid format
	}

	patchOp := &PatchOperation{
		Op:   parts[0],
		Path: parts[1],
	}

	// Add value if provided
	if len(parts) == 3 {
		value := parts[2]
		// Try to parse as JSON, fallback to string
		var jsonValue interface{}
		if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
			patchOp.Value = jsonValue
		} else {
			patchOp.Value = value
		}
	}

	return patchOp, nil
}

func TestScimPatchOperation_Parse_Simple(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		expectedOp   string
		expectedPath string
		expectedVal  interface{}
	}{
		{
			name:         "replace active with string",
			input:        "replace:active:false",
			expectedOp:   "replace",
			expectedPath: "active",
			expectedVal:  false, // JSON parsed to bool
		},
		{
			name:         "replace active with true",
			input:        "replace:active:true",
			expectedOp:   "replace",
			expectedPath: "active",
			expectedVal:  true,
		},
		{
			name:         "add role",
			input:        "add:roles:admin",
			expectedOp:   "add",
			expectedPath: "roles",
			expectedVal:  "admin", // plain string
		},
		{
			name:         "remove path",
			input:        "remove:emails",
			expectedOp:   "remove",
			expectedPath: "emails",
			expectedVal:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parsePatchOperation(tc.input)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.expectedOp, result.Op)
			assert.Equal(t, tc.expectedPath, result.Path)
			assert.Equal(t, tc.expectedVal, result.Value)
		})
	}
}

func TestScimPatchOperation_ParseJSON(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedVal interface{}
	}{
		{
			name:        "JSON number",
			input:       "replace:count:42",
			expectedVal: float64(42), // JSON numbers are float64
		},
		{
			name:        "JSON string",
			input:       `replace:name:"John"`,
			expectedVal: "John",
		},
		{
			name:        "JSON array",
			input:       `replace:roles:["admin","user"]`,
			expectedVal: []interface{}{"admin", "user"},
		},
		{
			name:        "JSON object",
			input:       `replace:config:{"key":"value"}`,
			expectedVal: map[string]interface{}{"key": "value"},
		},
		{
			name:        "JSON null",
			input:       "replace:data:null",
			expectedVal: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parsePatchOperation(tc.input)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.expectedVal, result.Value)
		})
	}
}

func TestScimPatchOperation_ParseInvalid(t *testing.T) {
	// Single part (no colon) should return nil
	result, err := parsePatchOperation("invalid")

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestScimPatchOperation_ParseEmptyValue(t *testing.T) {
	// Operation with empty value after colon
	result, err := parsePatchOperation("replace:active:")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "replace", result.Op)
	assert.Equal(t, "active", result.Path)
	assert.Equal(t, "", result.Value) // Empty string value
}

func TestScimPatchOperation_ParseColonInValue(t *testing.T) {
	// Value containing colons should be preserved
	result, err := parsePatchOperation("replace:url:https://example.com:8080/path")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "replace", result.Op)
	assert.Equal(t, "url", result.Path)
	assert.Equal(t, "https://example.com:8080/path", result.Value)
}

func TestScimPatchOperation_ParseComplexPath(t *testing.T) {
	// SCIM paths can be complex like "name.givenName"
	result, err := parsePatchOperation("replace:name.givenName:John")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "replace", result.Op)
	assert.Equal(t, "name.givenName", result.Path)
	assert.Equal(t, "John", result.Value)
}

func TestScimPatchOperation_Operations(t *testing.T) {
	// Test all standard SCIM operations
	operations := []string{"add", "replace", "remove"}

	for _, op := range operations {
		t.Run(op, func(t *testing.T) {
			input := op + ":path:value"
			result, err := parsePatchOperation(input)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, op, result.Op)
		})
	}
}

func TestScimPatchOperation_MultipleParsing(t *testing.T) {
	// Simulate parsing multiple operations as in the command
	operations := []string{
		"replace:active:false",
		"add:roles:[\"admin\"]",
		"replace:name.givenName:John",
	}

	var patchOps []PatchOperation
	for _, op := range operations {
		parsed, err := parsePatchOperation(op)
		require.NoError(t, err)
		require.NotNil(t, parsed)
		patchOps = append(patchOps, *parsed)
	}

	assert.Len(t, patchOps, 3)
	assert.Equal(t, "replace", patchOps[0].Op)
	assert.Equal(t, "add", patchOps[1].Op)
	assert.Equal(t, "replace", patchOps[2].Op)
}
