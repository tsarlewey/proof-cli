//go:build smoke

// Package smoke runs read-only end-to-end tests against the live Proof API.
//
// Excluded from the default test suite by the `smoke` build tag. Run with:
//
//	make smoke-go
//	# or
//	go build -o proof && PROOF_BIN=$PWD/proof go test -tags=smoke -v ./test/smoke/...
//
// Requires a working config at ~/.proof-cli/config.json. SCIM tests are
// skipped unless PROOF_SMOKE_ORG_ID is set.
package smoke

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var proofBin string

func TestMain(m *testing.M) {
	bin := os.Getenv("PROOF_BIN")
	if bin == "" {
		bin = "../../proof"
	}
	abs, err := filepath.Abs(bin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolving PROOF_BIN: %v\n", err)
		os.Exit(2)
	}
	if _, err := os.Stat(abs); err != nil {
		fmt.Fprintf(os.Stderr, "proof binary not found at %s; run 'make build' first or set PROOF_BIN\n", abs)
		os.Exit(2)
	}
	proofBin = abs
	os.Exit(m.Run())
}

// runProof runs the CLI with the given args and returns stdout, stderr, exit code.
func runProof(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(proofBin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run proof: %v", err)
		}
	}
	return stdout.String(), stderr.String(), exitCode
}

// assertSuccess fails the test if exit code is non-zero, logging output for debugging.
func assertSuccess(t *testing.T, name string, stdout, stderr string, exitCode int) {
	t.Helper()
	if exitCode != 0 {
		t.Fatalf("%s failed with exit code %d\nstdout: %s\nstderr: %s", name, exitCode, stdout, stderr)
	}
}

// assertJSON fails the test if stdout isn't valid JSON.
func assertJSON(t *testing.T, name, stdout string) {
	t.Helper()
	trimmed := strings.TrimSpace(stdout)
	if trimmed == "" {
		t.Fatalf("%s: expected JSON output, got empty stdout", name)
	}
	var v any
	if err := json.Unmarshal([]byte(trimmed), &v); err != nil {
		t.Fatalf("%s: expected valid JSON, parse error: %v", name, err)
	}
}

// Basic commands (no API hit) — confirms the binary runs and config loads.

func TestHelp(t *testing.T) {
	stdout, stderr, rc := runProof(t, "--help")
	assertSuccess(t, "--help", stdout, stderr, rc)
}

func TestConfigGet(t *testing.T) {
	stdout, stderr, rc := runProof(t, "config", "get")
	assertSuccess(t, "config get", stdout, stderr, rc)
}

// Business API — read-only endpoints.

func TestBusinessTransactionsList(t *testing.T) {
	stdout, stderr, rc := runProof(t, "--pretty=false", "business", "transactions", "list", "--limit", "1")
	assertSuccess(t, "business transactions list", stdout, stderr, rc)
	assertJSON(t, "business transactions list", stdout)
}

func TestBusinessWebhookSubscriptions(t *testing.T) {
	stdout, stderr, rc := runProof(t, "--pretty=false", "business", "webhooks", "subscriptions")
	assertSuccess(t, "business webhooks subscriptions", stdout, stderr, rc)
	assertJSON(t, "business webhooks subscriptions", stdout)
}

func TestBusinessNotariesList(t *testing.T) {
	stdout, stderr, rc := runProof(t, "--pretty=false", "business", "notaries", "list")
	assertSuccess(t, "business notaries list", stdout, stderr, rc)
	assertJSON(t, "business notaries list", stdout)
}

func TestBusinessTemplatesList(t *testing.T) {
	stdout, stderr, rc := runProof(t, "--pretty=false", "business", "templates", "list", "--limit", "1")
	assertSuccess(t, "business templates list", stdout, stderr, rc)
	assertJSON(t, "business templates list", stdout)
}

// Real Estate API — read-only endpoints.

func TestRealEstateTransactionsList(t *testing.T) {
	stdout, stderr, rc := runProof(t, "--pretty=false", "real-estate", "transactions", "list", "--limit", "1")
	assertSuccess(t, "real-estate transactions list", stdout, stderr, rc)
	assertJSON(t, "real-estate transactions list", stdout)
}

func TestRealEstateWebhooksList(t *testing.T) {
	stdout, stderr, rc := runProof(t, "--pretty=false", "real-estate", "webhooks", "list")
	assertSuccess(t, "real-estate webhooks list", stdout, stderr, rc)
	assertJSON(t, "real-estate webhooks list", stdout)
}

func TestRealEstateVerifyAddress(t *testing.T) {
	stdout, stderr, rc := runProof(t,
		"--pretty=false", "real-estate", "verify-address",
		"--line1", "1600 Pennsylvania Ave",
		"--city", "Washington",
		"--state", "DC",
		"--postal-code", "20500",
	)
	assertSuccess(t, "real-estate verify-address", stdout, stderr, rc)
	assertJSON(t, "real-estate verify-address", stdout)
}

// SCIM API — read-only, requires PROOF_SMOKE_ORG_ID.

func TestSCIMUsersList(t *testing.T) {
	orgID := os.Getenv("PROOF_SMOKE_ORG_ID")
	if orgID == "" {
		t.Skip("PROOF_SMOKE_ORG_ID not set; skipping SCIM test")
	}
	stdout, stderr, rc := runProof(t, "--pretty=false", "scim", "users", "list", orgID, "--count", "1")
	assertSuccess(t, "scim users list", stdout, stderr, rc)
	assertJSON(t, "scim users list", stdout)
}

func TestSCIMUserSchema(t *testing.T) {
	orgID := os.Getenv("PROOF_SMOKE_ORG_ID")
	if orgID == "" {
		t.Skip("PROOF_SMOKE_ORG_ID not set; skipping SCIM test")
	}
	stdout, stderr, rc := runProof(t, "--pretty=false", "scim", "schemas", "user", orgID)
	assertSuccess(t, "scim schemas user", stdout, stderr, rc)
	assertJSON(t, "scim schemas user", stdout)
}
