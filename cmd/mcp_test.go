package cmd

import (
	"regexp"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func findLeaf(t *testing.T, tool string) *cobra.Command {
	t.Helper()
	for _, c := range mcpLeaves(rootCmd) {
		if mcpToolName(c) == tool {
			return c
		}
	}
	t.Fatalf("no tool %q", tool)
	return nil
}

func TestMCPTools_NamesAndSchemas(t *testing.T) {
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)
	seen := map[string]bool{}
	for _, c := range mcpLeaves(rootCmd) {
		tool := mcpTool(c)
		assert.Regexp(t, validName, tool.Name)
		assert.False(t, seen[tool.Name], "duplicate tool %s", tool.Name)
		seen[tool.Name] = true

		schema := tool.InputSchema.(map[string]any)
		props := schema["properties"].(map[string]any)
		for _, name := range mcpPositionals(c) {
			assert.Contains(t, props, name, tool.Name)
			assert.Contains(t, schema["required"], name, tool.Name)
		}
		assert.NotContains(t, props, "help")
		assert.NotContains(t, props, "pretty", "persistent flags are the server's job")
	}
	assert.NotContains(t, seen, "config_set_api_key")
	assert.Contains(t, seen, "business_transactions_create")
}

func TestMCPTools_Annotations(t *testing.T) {
	get := mcpTool(findLeaf(t, "business_transactions_get")).Annotations
	assert.True(t, get.ReadOnlyHint)
	assert.False(t, *get.DestructiveHint)

	cancel := mcpTool(findLeaf(t, "business_transactions_cancel")).Annotations
	assert.False(t, cancel.ReadOnlyHint)
	assert.True(t, *cancel.DestructiveHint)

	assert.True(t, mcpTool(findLeaf(t, "scim_schemas_user")).Annotations.ReadOnlyHint)
	assert.False(t, mcpTool(findLeaf(t, "business_transactions_create")).Annotations.ReadOnlyHint)
}

func TestMCPArgv(t *testing.T) {
	create := findLeaf(t, "business_transactions_create")
	argv, err := mcpArgv(create, []byte(`{"email":"a@b.co","draft":false,"document":["one.pdf","two.pdf"]}`))
	require.NoError(t, err)
	assert.Equal(t, []string{"business", "transactions", "create", "--pretty=false"}, argv[:4])
	assert.ElementsMatch(t, []string{"--email=a@b.co", "--draft=false", "--document=one.pdf", "--document=two.pdf"}, argv[4:])

	list := findLeaf(t, "business_transactions_list")
	argv, err = mcpArgv(list, []byte(`{"limit":25}`))
	require.NoError(t, err)
	assert.Contains(t, argv, "--limit=25", "integers must not render as floats")

	get := findLeaf(t, "business_documents_get")
	argv, err = mcpArgv(get, []byte(`{"transaction-id":"ot_1","document-id":"do_2"}`))
	require.NoError(t, err)
	assert.Equal(t, []string{"business", "documents", "get", "--pretty=false", "ot_1", "do_2"}, argv)

	_, err = mcpArgv(get, []byte(`{"transaction-id":"ot_1"}`))
	assert.ErrorContains(t, err, `"document-id"`)

	_, err = mcpArgv(list, []byte(`{"bogus":1}`))
	assert.ErrorContains(t, err, `unknown argument "bogus"`)
}
