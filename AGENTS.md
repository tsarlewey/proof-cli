# AGENTS.md

## Using the `proof` CLI

Read [skills/proof/SKILL.md](skills/proof/SKILL.md): setup, auth, sandbox vs production, the safety rules (always `--draft`, confirm before `activate`/`cancel`/`delete`), the core notarization workflow, and a command map. It is the canonical usage guide for agents; this file doesn't repeat it.

- Install: `go install github.com/tsarlewey/proof-cli@latest`, or a binary from [Releases](https://github.com/tsarlewey/proof-cli/releases).
- MCP: `proof mcp` serves every API command as a tool over stdio (`--read-only` for reads only).
- Claude Code plugin (skill + MCP server): `/plugin marketplace add tsarlewey/proof-cli`, then `/plugin install proof@proof`.
- Writing code instead: use [proof-sdk-go](https://github.com/tsarlewey/proof-sdk-go) (Go) or the OpenAPI specs at https://dev.proof.com.

## Working on this repo

See [CLAUDE.md](CLAUDE.md) for architecture and conventions. `make check` (fmt, vet, build, test-race) must pass. SDK code is generated in proof-sdk-go; never hand-edit API clients here. New commands get MCP tools automatically (`cmd/mcp.go`); if a new leaf command reads only, or can't be undone, add its name to `mcpReadOnly` or `mcpDestructive`.
