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

# Patch a user (partial update)
proof scim users patch <organization-id> <user-id> \
  --operations '[{"op":"replace","path":"active","value":false}]'

# Delete a user
proof scim users delete <organization-id> <user-id>
```

#### Schemas

```bash
# Get user schema
proof scim schemas user-schema <organization-id>

# Get service provider configuration
proof scim schemas service-provider-config <organization-id>

# Get resource types
proof scim schemas resource-types <organization-id>
```

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

### Code Formatting

```bash
go fmt ./...
go vet ./...
```

## Support

For issues and feature requests, please visit the [GitHub repository](https://github.com/tsarlewey/proof-cli).

## License

