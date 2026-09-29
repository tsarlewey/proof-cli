package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// mcpSkip are top-level commands that make no sense as agent tools: local
// config, shell setup, and demos. Credentials come from PROOF_API_KEY or the
// config file the user already set up.
var mcpSkip = map[string]bool{"config": true, "completion": true, "example": true, "help": true, "version": true, "mcp": true}

// Leaf command names that only read. Everything else writes to Proof.
var mcpReadOnly = map[string]bool{
	"list": true, "get": true, "get-v2": true, "events": true, "subscriptions": true,
	"eligible-notaries": true, "resource-types": true, "service-provider-config": true,
	"verify-address": true, "authorize-url": true,
}

// mcpIsReadOnly also covers `scim schemas *`, whose leaf names (e.g. "user")
// read like nouns rather than verbs.
func mcpIsReadOnly(c *cobra.Command) bool {
	return mcpReadOnly[c.Name()] || c.Parent().Name() == "schemas"
}

// Writes that can't be undone or that notify real people.
var mcpDestructive = map[string]bool{
	"delete": true, "cancel": true, "recall": true, "revoke": true, "activate": true,
	"place-order": true, "resend-email": true, "resend-sms": true,
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Serve the CLI as an MCP server over stdio",
	Long: `Run a Model Context Protocol server on stdin/stdout so AI agents can call Proof.

Every API command becomes one tool (e.g. business_transactions_create), with an
input schema built from its arguments and flags. Tools are annotated read-only
or destructive so clients can ask before canceling, deleting, or activating.

Register with Claude Code:
  claude mcp add proof -e PROOF_API_KEY=... -- proof mcp`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		self, err := os.Executable()
		if err != nil {
			return err
		}
		readOnly, _ := cmd.Flags().GetBool("read-only")

		server := mcp.NewServer(&mcp.Implementation{Name: "proof", Version: Version}, nil)
		for _, c := range mcpLeaves(rootCmd) {
			if readOnly && !mcpIsReadOnly(c) {
				continue
			}
			server.AddTool(mcpTool(c), mcpHandler(self, c))
		}
		err = server.Run(cmd.Context(), &mcp.StdioTransport{})
		// The SDK reports a client hanging up as "server is closing" (-32004),
		// with the underlying io.EOF flattened to text. That's a normal exit.
		var rpcErr *jsonrpc.Error
		if errors.As(err, &rpcErr) && rpcErr.Code == -32004 {
			return nil
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.Flags().Bool("read-only", false, "Only expose tools that don't change anything in Proof")
}

// mcpLeaves returns every runnable API command under root.
func mcpLeaves(root *cobra.Command) []*cobra.Command {
	var out []*cobra.Command
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		for _, sub := range c.Commands() {
			if sub.Hidden || (c == root && mcpSkip[sub.Name()]) {
				continue
			}
			if sub.HasSubCommands() {
				walk(sub)
			} else if sub.Runnable() {
				out = append(out, sub)
			}
		}
	}
	walk(root)
	return out
}

// mcpPositionals reads argument names from Use, e.g. "get <transaction-id>".
func mcpPositionals(c *cobra.Command) []string {
	var names []string
	for _, f := range strings.Fields(c.Use)[1:] {
		names = append(names, strings.Trim(f, "<>[]"))
	}
	return names
}

func mcpToolName(c *cobra.Command) string {
	path := strings.Fields(c.CommandPath())[1:]
	return strings.ReplaceAll(strings.Join(path, "_"), "-", "_")
}

func mcpTool(c *cobra.Command) *mcp.Tool {
	props := map[string]any{}
	required := []string{}
	for _, name := range mcpPositionals(c) {
		props[name] = map[string]any{"type": "string", "description": "Positional argument " + name}
		required = append(required, name)
	}
	c.NonInheritedFlags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}
		props[f.Name] = mcpFlagSchema(f)
		if _, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok {
			required = append(required, f.Name)
		}
	})

	desc := c.Short
	if c.Long != "" && c.Long != c.Short {
		desc += "\n\n" + c.Long
	}
	desc += "\n\nCLI equivalent: " + c.CommandPath()

	name := c.Name()
	destructive := mcpDestructive[name]
	return &mcp.Tool{
		Name:        mcpToolName(c),
		Description: desc,
		InputSchema: map[string]any{"type": "object", "properties": props, "required": required},
		Annotations: &mcp.ToolAnnotations{
			Title:           c.Short,
			ReadOnlyHint:    mcpIsReadOnly(c),
			DestructiveHint: &destructive,
		},
	}
}

func mcpFlagSchema(f *pflag.Flag) map[string]any {
	s := map[string]any{"description": f.Usage}
	switch f.Value.Type() {
	case "bool":
		s["type"] = "boolean"
	case "int":
		s["type"] = "integer"
	case "stringSlice":
		s["type"] = "array"
		s["items"] = map[string]any{"type": "string"}
	default:
		s["type"] = "string"
	}
	return s
}

// mcpArgv turns tool arguments back into the command line the CLI parses.
func mcpArgv(c *cobra.Command, raw json.RawMessage) ([]string, error) {
	args := map[string]any{}
	if len(raw) > 0 {
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.UseNumber()
		if err := dec.Decode(&args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
	}

	argv := append(strings.Fields(c.CommandPath())[1:], "--pretty=false")
	for _, name := range mcpPositionals(c) {
		v, ok := args[name].(string)
		if !ok || v == "" {
			return nil, fmt.Errorf("missing required argument %q", name)
		}
		argv = append(argv, v)
		delete(args, name)
	}
	for name, v := range args {
		if c.NonInheritedFlags().Lookup(name) == nil {
			return nil, fmt.Errorf("unknown argument %q", name)
		}
		if list, ok := v.([]any); ok {
			for _, item := range list {
				argv = append(argv, fmt.Sprintf("--%s=%v", name, item))
			}
			continue
		}
		argv = append(argv, fmt.Sprintf("--%s=%v", name, v))
	}
	return argv, nil
}

// mcpHandler runs the command as a subprocess: the command handlers exit the
// process on error and share global client state, neither of which a
// long-running server can survive in-process.
func mcpHandler(self string, c *cobra.Command) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		argv, err := mcpArgv(c, req.Params.Arguments)
		if err != nil {
			return mcpText(err.Error(), true), nil
		}
		var stdout, stderr bytes.Buffer
		run := exec.CommandContext(ctx, self, argv...)
		run.Stdout, run.Stderr = &stdout, &stderr
		if err := run.Run(); err != nil {
			return mcpText(strings.TrimSpace(stderr.String()+"\n"+stdout.String()), true), nil
		}
		return mcpText(stdout.String(), false), nil
	}
}

func mcpText(s string, isError bool) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: s}}, IsError: isError}
}
