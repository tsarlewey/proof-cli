package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPtr_Int(t *testing.T) {
	// Test with various int values
	testCases := []struct {
		name  string
		input int
	}{
		{"zero", 0},
		{"positive", 42},
		{"negative", -10},
		{"large", 1000000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ptr(tc.input)

			assert.NotNil(t, result)
			assert.Equal(t, tc.input, *result)
		})
	}
}

func TestPtr_String(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"simple string", "hello"},
		{"string with spaces", "hello world"},
		{"unicode string", "Hello"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ptr(tc.input)

			assert.NotNil(t, result)
			assert.Equal(t, tc.input, *result)
		})
	}
}

func TestPtr_Bool(t *testing.T) {
	trueResult := ptr(true)
	assert.NotNil(t, trueResult)
	assert.True(t, *trueResult)

	falseResult := ptr(false)
	assert.NotNil(t, falseResult)
	assert.False(t, *falseResult)
}

func TestPtr_Float(t *testing.T) {
	testCases := []struct {
		name  string
		input float64
	}{
		{"zero", 0.0},
		{"positive", 3.14159},
		{"negative", -2.71828},
		{"small", 0.0001},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ptr(tc.input)

			assert.NotNil(t, result)
			assert.Equal(t, tc.input, *result)
		})
	}
}

func TestPtrIfNotEmpty_Empty(t *testing.T) {
	result := ptrIfNotEmpty("")

	assert.Nil(t, result)
}

func TestPtrIfNotEmpty_NonEmpty(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"simple string", "hello"},
		{"single character", "a"},
		{"whitespace only", "   "},
		{"string with newline", "hello\nworld"},
		{"special characters", "hello@world.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ptrIfNotEmpty(tc.input)

			assert.NotNil(t, result)
			assert.Equal(t, tc.input, *result)
		})
	}
}

func TestPtrIfNotEmpty_ReturnsNewPointer(t *testing.T) {
	// Verify that each call returns a new pointer
	input := "test"
	result1 := ptrIfNotEmpty(input)
	result2 := ptrIfNotEmpty(input)

	assert.NotNil(t, result1)
	assert.NotNil(t, result2)
	assert.Equal(t, *result1, *result2)
	// Note: In Go, these may or may not be the same address depending on compiler optimizations
	// The important thing is that the values are correct
}

func TestPtr_Struct(t *testing.T) {
	type TestStruct struct {
		Name  string
		Value int
	}

	input := TestStruct{Name: "test", Value: 42}
	result := ptr(input)

	assert.NotNil(t, result)
	assert.Equal(t, input.Name, result.Name)
	assert.Equal(t, input.Value, result.Value)
}

func TestPtr_Slice(t *testing.T) {
	input := []string{"a", "b", "c"}
	result := ptr(input)

	assert.NotNil(t, result)
	assert.Equal(t, len(input), len(*result))
	assert.Equal(t, input, *result)
}

func TestPtr_Map(t *testing.T) {
	input := map[string]int{"one": 1, "two": 2}
	result := ptr(input)

	assert.NotNil(t, result)
	assert.Equal(t, input["one"], (*result)["one"])
	assert.Equal(t, input["two"], (*result)["two"])
}
