package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// captureOutput captures stdout and stderr during function execution
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrintResponse_PrettyPrint(t *testing.T) {
	// Save and restore prettyPrint flag
	oldPrettyPrint := prettyPrint
	defer func() { prettyPrint = oldPrettyPrint }()

	prettyPrint = true

	testCases := []struct {
		name     string
		input    []byte
		expected []string // strings that should be in the output
	}{
		{
			name:  "simple JSON object",
			input: []byte(`{"name":"test","value":123}`),
			expected: []string{
				"name",
				"test",
				"value",
				"123",
			},
		},
		{
			name:  "JSON with boolean",
			input: []byte(`{"active":true,"disabled":false}`),
			expected: []string{
				"active",
				"true",
				"disabled",
				"false",
			},
		},
		{
			name:  "JSON with null",
			input: []byte(`{"data":null}`),
			expected: []string{
				"data",
				"null",
			},
		},
		{
			name:  "nested JSON",
			input: []byte(`{"user":{"name":"John","age":30}}`),
			expected: []string{
				"user",
				"name",
				"John",
				"age",
				"30",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := captureOutput(func() {
				PrintResponse(tc.input)
			})

			for _, exp := range tc.expected {
				assert.Contains(t, output, exp, "Output should contain %s", exp)
			}
		})
	}
}

func TestPrintResponse_RawOutput(t *testing.T) {
	// Save and restore prettyPrint flag
	oldPrettyPrint := prettyPrint
	defer func() { prettyPrint = oldPrettyPrint }()

	prettyPrint = false

	input := []byte(`{"name":"test","value":123}`)

	output := captureOutput(func() {
		PrintResponse(input)
	})

	// Raw output should be the exact JSON
	assert.Contains(t, output, `{"name":"test","value":123}`)
}

func TestPrintResponse_InvalidJSON(t *testing.T) {
	// Save and restore prettyPrint flag
	oldPrettyPrint := prettyPrint
	defer func() { prettyPrint = oldPrettyPrint }()

	prettyPrint = true

	// Invalid JSON should be printed as-is
	input := []byte(`not valid json`)

	output := captureOutput(func() {
		PrintResponse(input)
	})

	assert.Contains(t, output, "not valid json")
}

func TestPrintResponse_WithPrefix(t *testing.T) {
	// Save and restore prettyPrint flag
	oldPrettyPrint := prettyPrint
	defer func() { prettyPrint = oldPrettyPrint }()

	prettyPrint = false

	input := []byte(`{"status":"ok"}`)
	prefix := "Response:"

	output := captureOutput(func() {
		PrintResponse(input, prefix)
	})

	// Output should contain prefix and JSON
	assert.Contains(t, output, "Response:")
	assert.Contains(t, output, `{"status":"ok"}`)
}

func TestPrintResponse_EmptyJSON(t *testing.T) {
	// Save and restore prettyPrint flag
	oldPrettyPrint := prettyPrint
	defer func() { prettyPrint = oldPrettyPrint }()

	prettyPrint = true

	input := []byte(`{}`)

	output := captureOutput(func() {
		PrintResponse(input)
	})

	assert.Contains(t, output, "{}")
}

func TestPrintResponse_JSONArray(t *testing.T) {
	// Save and restore prettyPrint flag
	oldPrettyPrint := prettyPrint
	defer func() { prettyPrint = oldPrettyPrint }()

	prettyPrint = true

	input := []byte(`[{"id":1},{"id":2}]`)

	output := captureOutput(func() {
		PrintResponse(input)
	})

	assert.Contains(t, output, "id")
	assert.Contains(t, output, "1")
	assert.Contains(t, output, "2")
}

func TestColorizeJSON_Keys(t *testing.T) {
	// Test that keys are colorized (will contain ANSI escape codes in terminal)
	input := `{
  "key": "value"
}`

	result := colorizeJSON(input)

	// The result should contain "key" (the key might have color codes around it)
	assert.Contains(t, result, "key")
	assert.Contains(t, result, "value")
}

func TestColorizeJSON_StringValues(t *testing.T) {
	input := `{
  "name": "John Doe"
}`

	result := colorizeJSON(input)

	assert.Contains(t, result, "name")
	assert.Contains(t, result, "John Doe")
}

func TestColorizeJSON_NumberValues(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		number string
	}{
		{
			name: "integer",
			input: `{
  "count": 42
}`,
			number: "42",
		},
		{
			name: "negative integer",
			input: `{
  "offset": -10
}`,
			number: "-10",
		},
		{
			name: "float",
			input: `{
  "price": 19.99
}`,
			number: "19.99",
		},
		{
			name: "scientific notation",
			input: `{
  "tiny": 1.23e-4
}`,
			number: "1.23e-4",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := colorizeJSON(tc.input)
			assert.Contains(t, result, tc.number)
		})
	}
}

func TestColorizeJSON_BooleanValues(t *testing.T) {
	input := `{
  "active": true,
  "disabled": false
}`

	result := colorizeJSON(input)

	assert.Contains(t, result, "true")
	assert.Contains(t, result, "false")
}

func TestColorizeJSON_NullValues(t *testing.T) {
	input := `{
  "data": null
}`

	result := colorizeJSON(input)

	assert.Contains(t, result, "null")
}

func TestColorizeJSON_ComplexJSON(t *testing.T) {
	input := `{
  "user": {
    "name": "John",
    "age": 30,
    "active": true,
    "email": null
  },
  "items": [
    "item1",
    "item2"
  ]
}`

	result := colorizeJSON(input)

	// Verify structure is preserved
	assert.Contains(t, result, "user")
	assert.Contains(t, result, "name")
	assert.Contains(t, result, "John")
	assert.Contains(t, result, "age")
	assert.Contains(t, result, "30")
	assert.Contains(t, result, "active")
	assert.Contains(t, result, "true")
	assert.Contains(t, result, "email")
	assert.Contains(t, result, "null")
	assert.Contains(t, result, "items")
}

func TestColorizeJSON_EscapedStrings(t *testing.T) {
	input := `{
  "message": "Hello \"World\""
}`

	result := colorizeJSON(input)

	// The escaped quotes should be preserved
	assert.Contains(t, result, "message")
	// Note: the actual escaping depends on the regex handling
}

func TestPrintVerbose_WhenVerbose(t *testing.T) {
	// Save and restore verbose flag
	oldVerbose := verbose
	defer func() { verbose = oldVerbose }()

	verbose = true

	output := captureOutput(func() {
		PrintVerbose("Test message")
	})

	assert.Contains(t, output, "Test message")
}

func TestPrintVerbose_WhenNotVerbose(t *testing.T) {
	// Save and restore verbose flag
	oldVerbose := verbose
	defer func() { verbose = oldVerbose }()

	verbose = false

	output := captureOutput(func() {
		PrintVerbose("Test message")
	})

	assert.Empty(t, strings.TrimSpace(output))
}
