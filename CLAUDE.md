# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Current Repository State

The Proof CLI is a Cobra-based front-end over the generated SDK clients now hosted in [`github.com/tsarlewey/proof-sdk-go`](https://github.com/tsarlewey/proof-sdk-go). All SDK generation, OpenAPI spec management, and per-API client code lives in that repo; this repo only consumes it. Use `make check` (fmt + vet + build + test-race) as the canonical verification command.

### Recent Changes
- Extracted `pkg/sdk/` into its own module, `github.com/tsarlewey/proof-sdk-go`. Imports in `cmd/*.go` now use `github.com/tsarlewey/proof-sdk-go/<pkg>`.
- Moved SDK generation tooling (Makefile targets, `openapi/` specs, `scripts/fix-openapi-refs.py`, `scripts/fix-scim-operation-ids.py`) to the SDK repo. The CLI Makefile no longer has `generate` / `download-specs` / `regenerate` targets.
- Added `checkAPIStatus` helper in `cmd/root.go` so SDK calls that return non-2xx responses now exit 1 with the API's error body printed to stderr (previously they exited 0).
- Earlier cleanup: removed dead `ProofClient` CRUD helpers, `BuildQueryParams`, `Must[T]`, logrus. Consolidated `if err != nil { fmt.Println; os.Exit(1) }` onto `utils.HandleError`. Added `isSuccess`, `parseDateFlag` in `cmd/root.go` and `scimBoolString` in `cmd/scim.go`. Collapsed `NewAuthenticatedDoer` / `NewAuthenticatedDoerWithProvider` into a single `NewAuthenticatedDoer(AuthProvider)` constructor.

---

## Common Development Commands

### Build
```bash
go build -o proof
```

### Run
```bash
go run main.go
# or after building:
./proof
```

### Test
```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run with race detector
go test ./... -race

# Run specific package tests
go test ./pkg/utils/... -v
go test ./cmd/... -v
```

### Install
```bash
go install
# or
go install github.com/tsarlewey/proof-cli@latest
```

### Dependencies
```bash
go mod download
go mod tidy
```

### Formatting & Linting
```bash
go fmt ./...
go vet ./...
```

### SDK Regeneration

SDK generation lives in `github.com/tsarlewey/proof-sdk-go`. From *that* repo:

```bash
make regenerate   # download-specs + generate + build + test
```

After a regeneration, bump the SDK version there, tag it, and in this repo run `go get github.com/tsarlewey/proof-sdk-go@<new-version>` + `go mod tidy`.

---

## High-Level Architecture

The Proof CLI is a Cobra-based command-line application that interacts with the Proof API.

### Directory Structure
```
proof-cli/
├── cmd/                    # Cobra commands
│   ├── root.go            # Base command, global flags, SDK client factories, checkAPIStatus
│   ├── config.go          # Configuration management (OAuth, API key)
│   ├── business.go        # Business API commands
│   ├── real_estate.go     # Real Estate/Mortgage API commands
│   ├── scim.go            # SCIM identity management commands
│   └── example.go         # Example demonstration commands
├── pkg/
│   └── utils/             # Configuration, OAuth helpers, AuthProvider-implementing ProofClient
├── scripts/               # Smoke test shell scripts
└── test/smoke/            # End-to-end smoke tests (build-tagged smoke / smoke_write)
```

SDK generation, OpenAPI specs, and the generated clients live in `github.com/tsarlewey/proof-sdk-go`.

### SDK Architecture

SDK clients are imported from `github.com/tsarlewey/proof-sdk-go` (subpackages `business`, `realestate`, `scim`, `logs`, `certificates`, `common`). The SDK repo owns generation via `oapi-codegen`.

#### Authentication Adapter
`common.AuthenticatedDoer` (in the SDK repo) wraps any `AuthProvider` (two methods: `AddAuthHeaders(*http.Request) error`, `HTTPClient() *http.Client`) to inject auth headers on every request. `*utils.ProofClient` in this repo implements that interface (`pkg/utils/client.go:138-155`). Usage:

```go
authDoer := common.NewAuthenticatedDoer(proofClient)
client, err := business.NewClientWithResponses(
    proofClient.GetConfig().APIEndpoint,
    business.WithHTTPClient(authDoer),
)
```

`NewAuthenticatedDoer` takes the `AuthProvider` interface (not the concrete `*ProofClient`), which lets tests pass in mocks.

#### SDK Client Factories
Clients are lazily initialized in `cmd/root.go`:
- `getBusinessClient()` - Business API
- `getRealEstateClient()` - Real Estate/Mortgage API
- `getSCIMClient()` - SCIM API

Each factory wires `common.NewAuthenticatedDoer(proofClient)` into the generated SDK's `WithHTTPClient` option. There is no shared factory — the three look similar but each returns a different concrete SDK type, and the triplication is deliberate (see P1/P2 cleanup guardrails).

#### Shared Command Helpers (`cmd/root.go`)
- `initializeForAPICall` — lazy `ProofClient` setup; used as `PreRun` on every API-calling command.
- `PrintResponse` / `PrintVerbose` — response output with optional pretty-printing + colorization.
- `parseDateFlag(name, value, layout)` — parses an optional date flag or exits with a clear message.
- `isSuccess(code)` — 2xx status-code check, used after delete operations.
- `colorizeJSON` — regex-based JSON colorizer; do not rewrite as a tokenizer.

SDK call errors flow through `utils.HandleError(err, "action phrase")` for transport errors, then `checkAPIStatus(resp.StatusCode(), resp.Body, "action phrase")` for non-2xx responses. Both exit 1 on failure and print the context to stderr.

### Configuration Layer (`pkg/utils/`)
- `config.go`: Configuration management and storage
  - Stores config in `~/.proof-cli/config.json`
  - API key stored separately in `~/.proof-cli/api_key` (permissions 0600)
  - OAuth configuration with client credentials
  - Supports environment variable `PROOF_API_KEY`
- `oauth.go`: OAuth 2.0 helper functions (no HTTP calls)
  - Token management (save/load/expiration checking)
  - Request preparation and response parsing
  - OAuth configuration validation
- `client.go`: HTTP client with authentication
  - Centralized HTTP communication
  - Automatic OAuth vs API key authentication
  - Token refresh management
  - Request/response handling

---

## Key Architectural Decisions

- **Authentication**: Dual authentication support
  - OAuth 2.0 client credentials flow (preferred)
  - API key authentication (fallback)
  - Automatic token refresh with 5-minute expiration buffer
- **SDK Source**: Generated SDK clients live in `github.com/tsarlewey/proof-sdk-go` and are consumed by this repo as a regular Go module dependency
- **Separation of Concerns**: Clean architecture pattern
  - Config commands only handle configuration management
  - OAuth helpers only prepare requests and parse responses
  - Client handles all HTTP communication
- **Configuration Storage**:
  - Main config: `~/.proof-cli/config.json`
  - API key: `~/.proof-cli/api_key` (permissions 0600)
  - OAuth tokens: `~/.proof-cli/oauth_token` (permissions 0600)
- **Security**:
  - OAuth client credentials use HTTP Basic Authentication
  - Secure file permissions for sensitive data
  - Bearer token authentication for API requests
- Uses Cobra for command structure and argument parsing
- All API interactions go through generated SDK clients wrapped with `common.AuthenticatedDoer`; `ProofClient` holds config + auth state only (no generic HTTP CRUD helpers)
- No logging framework — `PrintVerbose` + `fmt` is sufficient for a CLI
- Uses Nix flake for reproducible builds (flake.nix)

---

## Known Issues

### SCIM Patch Operation Structure
The upstream SCIM OpenAPI spec defines the PATCH `Operations` field as a single object rather than an array. SCIM requires an array. The workaround in `cmd/scim.go` uses `PatchUserWithBodyWithResponse` with a manually marshaled JSON body (`scimPatchOperation` / `scimPatchRequest` structs) to send a proper array.

### SCIM Active-Field Type
The upstream SCIM spec types the user `active` field as `*string` ("true"/"false") instead of `*bool`. The `scimBoolString(b bool) *string` helper in `cmd/scim.go` bridges the CLI's bool flag to the SDK's string pointer. Keep the helper package-private to `cmd` — this is a SCIM-specific wart and doesn't belong in `pkg/utils`.

---

## OAuth 2.0 Authentication

The CLI supports OAuth 2.0 client credentials flow for secure API authentication:

### Configuration Commands
```bash
# Set OAuth credentials
./proof config set-oauth <client-id> <client-secret> [--scope="read write"]

# Test OAuth authentication
./proof config test-oauth

# Disable OAuth (fallback to API key)
./proof config disable-oauth

# View current configuration
./proof config get
```

### OAuth Flow Implementation
1. **Client Registration**: Use `set-oauth` to configure client credentials
2. **Token Request**: Client automatically requests tokens using HTTP Basic Auth
3. **Token Storage**: Tokens saved to `~/.proof-cli/oauth_token` with 0600 permissions
4. **Automatic Refresh**: Tokens refreshed automatically with 5-minute buffer before expiration
5. **API Requests**: Bearer tokens automatically added to `Authorization` header

### OAuth Configuration Structure
```json
{
  "api_endpoint": "https://api.proof.com",
  "timeout": "30s",
  "oauth": {
    "enabled": true,
    "client_id": "your-client-id",
    "client_secret": "your-client-secret",
    "scope": "read write"
  }
}
```

### Token Structure
```json
{
  "access_token": "eyJ...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "expires_at": "2025-07-31T12:00:00Z",
  "scope": "read write"
}
```

---

## Testing Patterns

All SDK tests follow consistent patterns:

```go
// Mock HTTP transport
type MockRoundTripper struct {
    Response *http.Response
    Err      error
    LastReq  *http.Request
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
    m.LastReq = req
    return m.Response, m.Err
}

// Create mock JSON response
func mockJSONResponse(statusCode int, body any) *http.Response {
    jsonBytes, _ := json.Marshal(body)
    return &http.Response{
        StatusCode: statusCode,
        Status:     http.StatusText(statusCode),
        Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
        Header:     http.Header{"Content-Type": []string{"application/json"}},
    }
}

// Generic pointer helper
func ptr[T any](v T) *T { return &v }
```

Use `require.NoError()` for fatal assertions, `assert.*` for non-fatal assertions.
