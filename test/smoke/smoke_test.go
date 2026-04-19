//go:build smoke

// Read-only smoke tests against the live Proof API. Shared helpers live in
// helpers_test.go. Run with:
//
//	make smoke-go
//	# or
//	go build -o proof && PROOF_BIN=$PWD/proof go test -tags=smoke -v ./test/smoke/...
//
// Requires a working config at ~/.proof-cli/config.json. SCIM tests are
// skipped unless PROOF_SMOKE_ORG_ID is set.
package smoke

import (
	"os"
	"testing"
)

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
