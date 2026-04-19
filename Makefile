.PHONY: build install fmt vet test test-race coverage check clean smoke smoke-go smoke-write smoke-write-go help

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

# Run tests
test:
	go test ./...

# Run tests with the race detector
test-race:
	go test ./... -race

# Run tests and emit a coverage profile
coverage:
	go test ./... -coverprofile=coverage.out
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | tail -n 1

# Run all checks (format, vet, build, test with race detector)
check: fmt vet build test-race

# Remove the built binary
clean:
	rm -f proof proof-cli

# Run read-only smoke tests (shell) against the live Proof API.
# Set PROOF_SMOKE_ORG_ID to also exercise SCIM endpoints.
smoke: build
	./scripts/smoke-test.sh

# Run read-only smoke tests (Go, build-tag smoke) against the live Proof API.
smoke-go: build
	PROOF_BIN=$(PWD)/proof go test -tags=smoke -v ./test/smoke/...

# Run write smoke tests (shell, create+delete). Requires PROOF_SMOKE_WRITE=1.
smoke-write: build
	./scripts/smoke-test-write.sh

# Run write smoke tests (Go, build-tag smoke_write). Requires PROOF_SMOKE_WRITE=1.
smoke-write-go: build
	PROOF_BIN=$(PWD)/proof go test -tags=smoke_write -v ./test/smoke/...

# Show available targets
help:
	@echo "Available targets:"
	@echo "  build          Build the ./proof binary"
	@echo "  install        Install the CLI via go install"
	@echo "  fmt            Run go fmt"
	@echo "  vet            Run go vet"
	@echo "  test           Run all tests"
	@echo "  test-race      Run all tests with -race"
	@echo "  coverage       Run tests with coverage profile"
	@echo "  clean          Remove the built binary"
	@echo "  check          fmt + vet + build + test-race"
	@echo "  smoke          Run shell smoke tests against live API (set PROOF_SMOKE_ORG_ID for SCIM)"
	@echo "  smoke-go       Run Go smoke tests (build-tag smoke) against live API"
	@echo "  smoke-write    Run shell write smoke tests (PROOF_SMOKE_WRITE=1)"
	@echo "  smoke-write-go Run Go write smoke tests (build-tag smoke_write)"
	@echo ""
	@echo "SDK regeneration moved to github.com/tsarlewey/proof-sdk-go."
