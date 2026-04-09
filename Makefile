.PHONY: build generate download-specs clean install fmt vet tools

# Build the CLI binary
build:
	go build -o proof

# Install the CLI
install:
	go install

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Install tool dependencies
tools:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Download OpenAPI specs from dev.proof.com
download-specs:
	@mkdir -p openapi
	@echo "Downloading Business API spec..."
	@curl -sL "https://dev.proof.com/openapi/proof-business-api-specification.json" -o openapi/business.json
	@echo "Downloading Real Estate API spec..."
	@curl -sL "https://dev.proof.com/openapi/proof-real-estate-api-specification.json" -o openapi/realestate.json
	@echo "Downloading SCIM API spec..."
	@curl -sL "https://dev.proof.com/openapi/proof-scim-api-specification.json" -o openapi/scim.json
	@echo "Downloading Logs API spec..."
	@curl -sL "https://dev.proof.com/openapi/proof-logs-api-specification.json" -o openapi/logs.json
	@echo "Downloading Certificates API spec..."
	@curl -sL "https://dev.proof.com/openapi/organization-certificates-openapi-specification.json" -o openapi/certificates.json
	@echo "Fixing deep $ref references in specs..."
	@python3 scripts/fix-openapi-refs.py openapi/business.json
	@python3 scripts/fix-openapi-refs.py openapi/realestate.json
	@python3 scripts/fix-openapi-refs.py openapi/scim.json
	@python3 scripts/fix-openapi-refs.py openapi/logs.json
	@python3 scripts/fix-openapi-refs.py openapi/certificates.json
	@echo "All specs downloaded and fixed!"

# oapi-codegen command - use go run to avoid PATH issues
OAPI_CODEGEN = go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Generate SDK clients from OpenAPI specs
generate:
	@echo "Generating Business SDK..."
	@mkdir -p pkg/sdk/business
	@$(OAPI_CODEGEN) --config pkg/sdk/business/oapi-codegen.yaml openapi/business.json
	@echo "Generating Real Estate SDK..."
	@mkdir -p pkg/sdk/realestate
	@$(OAPI_CODEGEN) --config pkg/sdk/realestate/oapi-codegen.yaml openapi/realestate.json
	@echo "Generating SCIM SDK..."
	@mkdir -p pkg/sdk/scim
	@$(OAPI_CODEGEN) --config pkg/sdk/scim/oapi-codegen.yaml openapi/scim.json
	@echo "Generating Logs SDK..."
	@mkdir -p pkg/sdk/logs
	@$(OAPI_CODEGEN) --config pkg/sdk/logs/oapi-codegen.yaml openapi/logs.json
	@echo "Generating Certificates SDK..."
	@mkdir -p pkg/sdk/certificates
	@$(OAPI_CODEGEN) --config pkg/sdk/certificates/oapi-codegen.yaml openapi/certificates.json
	@echo "SDK generation complete!"

# Clean generated files and binaries
clean:
	rm -f proof proof-cli
	rm -rf openapi/
	rm -f pkg/sdk/*/client.gen.go

# Run all checks
check: fmt vet build
