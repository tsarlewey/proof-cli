---
name: proof
description: Use the Proof API (proof.com) through the `proof` CLI to send documents for remote online notarization or e-signature, verify signer identity, track transaction status, manage webhooks, mortgage/real-estate closings, SCIM users, certificates, and verifiable credentials. Use when the user wants to notarize or e-sign a document, create/check/cancel a Proof transaction, set up Proof webhooks, or integrate an app with Proof.
---

# Proof CLI

`proof` is an unofficial CLI over the Proof API. Every command prints the API's JSON response; non-2xx responses exit 1 with the error body on stderr.

## Setup (check before first use)

```bash
command -v proof || go install github.com/tsarlewey/proof-cli@latest   # or a release binary from github.com/tsarlewey/proof-cli/releases
proof config get                                                       # shows endpoint and whether a key is set
```

- Auth: `PROOF_API_KEY` env var, or `proof config set-api-key <key>`. Never print, log, or commit the key. If none is set, ask the user; keys come from https://dev.proof.com/docs/api-keys.
- Environment: production is `https://api.proof.com`; the sandbox is `https://api.fairfax.proof.com` (`proof config set-endpoint <url>`, or `PROOF_API_ENDPOINT=<url>` for one command or session). **Use the sandbox for testing and anything exploratory.** Check `proof config get` before any write, and tell the user which environment you're hitting.
- Add `--pretty=false` when you'll parse output (e.g. pipe to `jq`).

## Safety rules

Proof transactions reach real people: activating one emails the signer, and notarizations are legal acts.

1. **Always create transactions with `--draft`.** Without it, `create` sends immediately.
2. **Ask the user before** `activate`, `cancel`, `recall`, `delete`, `resend-email`, `resend-sms`, `real-estate transactions place-order`, `certificates revoke`, `scim users delete`. State the transaction ID, signer, and environment when asking.
3. Only use signer emails the user gave you. Never guess or invent one.
4. Read before writing: `get` the transaction and check its status first.

## Core workflow: send a document for notarization or e-signature

```bash
# 1. Draft it (use a file path; repeat --document for several)
proof business transactions create --draft --pretty=false \
  --email signer@example.com --first-name Ada --last-name Lovelace \
  --document ./contract.pdf --name "Contract signing" | jq -r .id     # -> ot_...

# 2. Review it with the user
proof business transactions get <transaction-id>

# 3. After the user confirms, send it
proof business transactions activate <transaction-id>

# 4. Track it
proof business transactions get <transaction-id> | jq .status
```

- E-sign only instead of notarization: add a document with `proof business documents add <transaction-id> <file> --requirement esign` or set `--type`; check `proof business transactions create --help`.
- Multiple signers: `--signers-file signers.json`, a JSON array of signer objects (max 10, each needs `email`); `--email` stays the primary signer.
- Edit a draft: `patch` changes only the fields you pass; `update` replaces the whole draft (anything omitted reverts to default). Prefer `patch`.
- Stop it: `cancel`, `recall` (optional `--reason`), and `delete` behave differently and the API spec doesn't define them precisely. Describe the options to the user, link https://dev.proof.com/reference, and let them choose.
- Downloads: `proof business documents get <transaction-id> <document-id> --encoding uri` for a hosted URL of the completed document.

## Command map

| Goal | Command |
|---|---|
| List / find transactions | `proof business transactions list --limit 20 --status completed` |
| Transaction details and status | `proof business transactions get <id>` |
| Add, get, delete documents | `proof business documents add\|get\|delete` (`--template-id` applies a specific template) |
| Templates | `proof business templates list`, `proof real-estate templates list` |
| Notaries for a transaction | `proof business transactions eligible-notaries <id>` |
| Webhooks (status callbacks) | `proof business webhooks list\|create\|subscriptions\|events <id>` |
| Mortgage / real-estate closings | `proof real-estate transactions list\|get\|create\|place-order\|cancel` |
| User provisioning (SCIM) | `proof scim users list\|get\|create\|patch\|delete <org-id> ...` |
| Security event logs | `proof logs list` |
| Organization certificates | `proof certificates list\|create\|sign\|revoke` |
| Verifiable credential presentation URL | `proof credentials authorize-url` |

Every command has `--help` listing every flag with a description; read it rather than guessing flag names. Flags you don't pass are omitted from the request, so Proof applies the organization's defaults.

## Building an integration in code

For Go, use the SDK the CLI is built on: `go get github.com/tsarlewey/proof-sdk-go` (see its AGENTS.md and examples/). For other languages, the OpenAPI specs are at https://dev.proof.com and in `proof-sdk-go/openapi/`. Use the CLI to try calls against the sandbox first, then port the working request to code.

## MCP

`proof mcp` serves every command above as an MCP tool over stdio (`proof mcp --read-only` exposes only reads). Destructive tools carry `destructiveHint`. The safety rules still apply.
