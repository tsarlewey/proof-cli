# Proof CLI

An unofficial command-line interface for interacting with the Proof API. The Proof CLI provides access to Business, Real Estate, and SCIM APIs through a unified interface.

# This is an alpha product and should be treated as such. It is not an official product from Proof.

## Installation

### From Source

```bash
git clone https://github.com/tsarlewey/proof-cli.git
cd proof-cli
go build -o proof
./proof --help
```

### Install via Go

```bash
go install github.com/tsarlewey/proof-cli@latest
```

## Getting Started

### 1. Configuration

Before using the CLI, you need to configure your API key:

```bash
# Set your API key
proof config set-api-key YOUR_API_KEY

# Or set via environment variable
export PROOF_API_KEY=YOUR_API_KEY

# Verify your configuration
proof config get
```
In order to get an API Key, you'll need an account with API Usage. You can get that here - https://www.proof.com/pricing#business

After your account is created and upgrade use https://dev.proof.com/docs/api-keys to setup your key.

### OAuth 2.0 Authentication

The CLI supports OAuth 2.0 client credentials flow for secure API authentication:

```bash
# Configure OAuth credentials
proof config set-oauth <client-id> <client-secret> [--scope="read write"]

# Test OAuth authentication
proof config test-oauth

# Disable OAuth (fallback to API key)
proof config disable-oauth
```

OAuth tokens are automatically refreshed before expiration (5-minute buffer).

### 2. Basic Usage

The CLI is organized into three main API groups:

- `business` - Business API operations (transactions, documents, webhooks, notaries, etc.)
- `real-estate` - Real Estate/Mortgage API operations
- `scim` - SCIM API operations for user management

```bash
# Get help for any command
proof --help
proof business --help
proof real-estate --help
proof scim --help
```

## API Reference

### Business API

The Business API provides comprehensive transaction and document management capabilities.

#### Transactions

```bash
# List all transactions
proof business transactions list

# List with filtering
proof business transactions list --limit 10 --status completed

# Get a specific transaction
proof business transactions get <transaction-id>

# Create a new transaction
proof business transactions create \
  --email "signer@example.com" \
  --first-name "John" \
  --last-name "Doe" \
  --document "/path/to/document.pdf" \
  --name "Contract Signing" \
  --draft

# Multiple documents, scheduled notary meeting, and a redirect
proof business transactions create \
  --email "signer@example.com" \
  --document "/path/to/deed.pdf" \
  --document "/path/to/rider.pdf" \
  --notary-meeting-time "2026-09-02T15:30:00Z" \
  --allowed-notary-states "CA,NY" \
  --notary-note "Verify the property address" \
  --payer sender \
  --idv-use-case ACCOUNT_RECOVERY \
  --recipient-details-config "name=locked" \
  --redirect-url "https://example.com/done"

# See the full parameter set
proof business transactions create --help

# Activate a draft transaction
proof business transactions activate <transaction-id>

# Delete a transaction
proof business transactions delete <transaction-id>
```

#### Documents

```bash
# Add a document to a transaction
proof business documents add <transaction-id> /path/to/document.pdf \
  --filename "Contract.pdf" \
  --requirement "esign" \
  --esign-required

# Get a document
proof business documents get <transaction-id> <document-id>

# Get document as hosted URL (after completion)
proof business documents get <transaction-id> <document-id> --encoding uri

# Delete a document
proof business documents delete <document-id>
```

#### Webhooks

```bash
# List webhooks (v2)
proof business webhooks list

# Get webhook details
proof business webhooks get-v2 <webhook-id>

# Create a webhook
proof business webhooks create \
  --url "https://example.com/webhook" \
  --name "Transaction Updates" \
  --events "transaction.created,transaction.completed"

# Update a webhook
proof business webhooks update <webhook-id> \
  --url "https://new-url.com/webhook"

# Delete a webhook
proof business webhooks delete <webhook-id>

# Get webhook events
proof business webhooks events <webhook-id>

# List available event subscriptions
proof business webhooks subscriptions
```

#### Notaries

```bash
# List notaries
proof business notaries list

# List notaries by state
proof business notaries list --state CA

# Get notary details
proof business notaries get <notary-id>

# Create a notary
proof business notaries create \
  --email "notary@example.com" \
  --first-name "Jane" \
  --last-name "Smith" \
  --state "CA"

# Delete a notary
proof business notaries delete <notary-id>
```

#### Templates

```bash
# List document templates
proof business templates list

# List with pagination
proof business templates list --limit 50 --offset 100
```

#### Referrals

```bash
# Create a referral campaign
proof business referrals create \
  --name "Partner Referrals" \
  --cover-payment \
  --redirect-url "https://example.com/signup"

# Generate a referral code
proof business referrals generate-code <campaign-id> \
  --expires-at "2024-12-31T23:59:59Z"
```

#### Integrations

```bash
# Create an Adobe integration
proof business integrations create \
  --name "ADOBE" \
  --org-id "your-org-id" \
  --account-id "adobe-account-id" \
  --environment "production"

# Create a DocuTech integration
proof business integrations create \
  --name "DOCUTECH" \
  --org-id "your-org-id"
```

### Real Estate API

The Real Estate API specializes in mortgage and real estate transaction management.

#### Transactions

```bash
# List real estate transactions
proof real-estate transactions list

# List with filtering
proof real-estate transactions list --status "in_progress" --limit 20

# Get transaction details
proof real-estate transactions get <transaction-id>

# Create a transaction
proof real-estate transactions create \
  --type "purchase" \
  --file-number "RE-2024-001" \
  --loan-number "LN-2024-001"

# Place an order for a transaction
proof real-estate transactions place-order <transaction-id>
```

#### Documents

```bash
# List documents
proof real-estate documents list --transaction-id <transaction-id>

# Get document details
proof real-estate documents get <document-id>

# Upload a document
proof real-estate documents upload <transaction-id> /path/to/document.pdf \
  --type "purchase_agreement" \
  --external-id "PA-001"
```

#### Webhooks

```bash
# List real estate webhooks
proof real-estate webhooks list

# Create a webhook
proof real-estate webhooks create "https://example.com/re-webhook" \
  --subscriptions "transaction.created,document.uploaded"
```

#### Address Verification

```bash
# Verify an address
proof real-estate verify-address \
  --line1 "123 Main St" \
  --city "San Francisco" \
  --state "CA" \
  --postal-code "94102"
```

### SCIM API

The SCIM API provides standardized user management capabilities.

#### Users

```bash
# List users in an organization
proof scim users list <organization-id>

# List with pagination
proof scim users list <organization-id> --start-index 1 --count 25

# Get user details
proof scim users get <organization-id> <user-id>

# Create a user
proof scim users create <organization-id> \
  --username "user@example.com" \
  --given-name "John" \
  --family-name "Doe" \
  --email "user@example.com" \
  --active

# Update a user (full replacement)
proof scim users update <organization-id> <user-id> \
  --username "updated@example.com" \
  --given-name "Jane"

# Patch a user (partial update); --operation is repeatable, format op:path[:value]
proof scim users patch <organization-id> <user-id> \
  --operation 'replace:active:false'

# Delete a user
proof scim users delete <organization-id> <user-id>
```

#### Schemas

```bash
# Get user schema
proof scim schemas user <organization-id>

# Get service provider configuration
proof scim schemas service-provider-config <organization-id>

# Get resource types
proof scim schemas resource-types <organization-id>
```

### Security Events API

Read the organization's OCSF-formatted security event log.

```bash
# List recent security events
proof logs list

# Page through results
proof logs list --limit 100
proof logs list --cursor <next_cursor-from-previous-response>

# Filter
proof logs list --since 2026-01-01T00:00:00Z
proof logs list --class-uid 1001 --severity-id 3
```

### Certificates API

Issue, use, and revoke organization certificates.

```bash
# List certificates
proof certificates list --limit 25 --offset 0

# Get one certificate
proof certificates get <certificate-id>

# Issue a certificate with a Proof-generated key
proof certificates create \
  --common-name "Acme Signing Authority" \
  --profile organization_authenticity_al2

# Issue a certificate from your own CSR
proof certificates create-from-csr --csr "$(cat request.pem)"

# Sign base64-encoded SHA256 digests (--digest is repeatable, max 25)
proof certificates sign <certificate-id> --digest "BOGQnhPlcpXqM7fAH6tvFPI4QOXsIyXMiBKtFpblmjU="

# Revoke
proof certificates revoke <certificate-id> --reason "key compromise"
```

### Verifiable Credentials API

Requests a Verifiable Credential presentation from an End-User. This endpoint is
a **browser redirect**, not a server-to-server call — the CLI prints the URL for
you to send the End-User to, and does not follow it.

```bash
# Fragment mode: the vp_token comes back on your redirect URI
proof credentials authorize-url \
  --client-id <oauth-client-id> \
  --response-mode fragment \
  --redirect-uri https://app.example.com/callback \
  --scope "openid" \
  --login-hint user@example.com \
  --nonce "$(openssl rand -hex 16)" \
  --state "$(openssl rand -hex 8)"

# direct_post mode: Proof POSTs the vp_token to your response URI
proof credentials authorize-url \
  --client-id <oauth-client-id> \
  --response-mode direct_post \
  --response-uri https://app.example.com/vp \
  --scope "openid" \
  --login-hint user@example.com \
  --nonce "$(openssl rand -hex 16)"
```

`--redirect-uri` is required in `fragment` mode and rejected in `direct_post`
mode; `--response-uri` is the reverse. Both URIs must be registered on your
OAuth Application.

## Examples

The CLI includes example commands that demonstrate common workflows:

```bash
# List all available examples
proof example --help

# Business API examples
proof example business-transactions     # List business transactions
proof example business-notary          # Create a notary

# Real Estate API examples
proof example real-estate-transactions  # List real estate transactions
proof example verify-address           # Verify an address

# SCIM API examples
proof example scim-users <org-id>       # List SCIM users
proof example scim-schema <org-id>      # Get SCIM user schema
```

## Configuration

The CLI stores configuration in `~/.proof-cli/`:

- `config.json` - Main configuration file
- `api_key` - API key (permissions 0600)

Configuration options:
- `endpoint` - API endpoint URL
- `timeout` - Request timeout in seconds

```bash
# View current configuration
proof config get

# Set API endpoint
proof config set-endpoint "https://api.proof.com"

# Set request timeout
proof config set-timeout 30

# Set API key
proof config set-api-key "your-api-key"
```

## Global Flags

All commands support these global flags:

- `--pretty` - Pretty print JSON output (default: true)
- `--verbose` - Show additional debug output
- `--help` - Show help information

## Environment Variables

- `PROOF_API_KEY` - API key for authentication
- `PROOF_ENDPOINT` - Override default API endpoint
- `PROOF_TIMEOUT` - Request timeout in seconds

## Error Handling

The CLI provides detailed error messages and uses standard exit codes:

- `0` - Success
- `1` - General error (API error, invalid arguments, etc.)

## Development

### Building from Source

```bash
git clone https://github.com/tsarlewey/proof-cli.git
cd proof-cli
go mod download
go build -o proof
```

### Makefile Targets

The project provides a `Makefile` for common development tasks:

```bash
make build          # Build the ./proof binary
make install        # Install the CLI via go install
make fmt            # Run go fmt
make vet            # Run go vet
make test           # Run all tests
make test-race      # Run all tests with -race
make coverage       # Run tests with coverage profile
make check          # fmt + vet + build + test-race
make clean          # Remove the built binary
make smoke          # Run shell smoke tests against live API
make smoke-go       # Run Go smoke tests (build-tag smoke)
make smoke-write    # Run shell write smoke tests (PROOF_SMOKE_WRITE=1)
make smoke-write-go # Run Go write smoke tests (build-tag smoke_write)
```

### SDK Source

The generated Go SDK clients now live in their own repo: [`github.com/tsarlewey/proof-sdk-go`](https://github.com/tsarlewey/proof-sdk-go). This CLI imports them as a regular Go module dependency. If you need to regenerate the clients after an upstream OpenAPI spec change, work inside the SDK repo — the generation toolchain (OpenAPI specs, `oapi-codegen` configs, preprocessing scripts, `make regenerate`) moved there.

To pick up a new SDK release in this CLI:

```bash
go get github.com/tsarlewey/proof-sdk-go@<new-version>
go mod tidy
make check
```

### Adding or Updating Commands

CLI commands live in `cmd/` (one file per API surface) and are built with [Cobra](https://github.com/spf13/cobra). Each command's `Run` closure calls a method on the generated SDK client via the factory helpers in `cmd/root.go` (`getBusinessClient`, `getRealEstateClient`, `getSCIMClient`). The factories wrap the SDK client with `common.AuthenticatedDoer` (from the SDK repo), which injects OAuth bearer tokens or the API key on every request.

Shared helpers in `cmd/root.go`:

- `initializeForAPICall` — lazy client setup, used as `PreRun` on every API-calling command.
- `PrintResponse` / `PrintVerbose` — response output with optional pretty-printing and colorization.
- `parseDateFlag` — parses optional date-flag values, exits with a clear message on parse failure.
- `isSuccess` — 2xx status-code check.
- `checkAPIStatus` — after every SDK call, exits 1 with the body on stderr if the API returned a non-2xx response.

Errors from SDK calls are funneled through `utils.HandleError(err, "action phrase")` for transport errors and `checkAPIStatus(resp.StatusCode(), resp.Body, "action phrase")` for application-level HTTP errors. Both exit 1.

## Support

For issues and feature requests, please visit the [GitHub repository](https://github.com/tsarlewey/proof-cli).

## License

