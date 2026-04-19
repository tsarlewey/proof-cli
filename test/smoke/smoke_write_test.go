//go:build smoke_write

// Write smokes mutate real resources in the live Proof account. Each test
// pairs a create with a t.Cleanup-registered delete so resources do not leak,
// even if later assertions fail.
//
// Opt-in only. Run with:
//
//	make smoke-write-go
//	# or
//	PROOF_SMOKE_WRITE=1 PROOF_BIN=$PWD/proof go test -tags=smoke_write -v ./test/smoke/...

package smoke

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// requireWriteOptIn skips the test unless PROOF_SMOKE_WRITE=1. This is a
// deliberate belt-and-suspenders guard on top of the smoke_write build tag.
func requireWriteOptIn(t *testing.T) {
	t.Helper()
	if os.Getenv("PROOF_SMOKE_WRITE") != "1" {
		t.Skip("PROOF_SMOKE_WRITE != 1; write smokes are opt-in to avoid accidental mutation")
	}
}

// smokePDFPath returns the path to a PDF fixture used by transaction-create
// smokes. If PROOF_SMOKE_PDF is set, that path wins. Otherwise the first *.pdf
// file next to this test file is used. Skips the test if no fixture is found.
func smokePDFPath(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("PROOF_SMOKE_PDF"); p != "" {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("PROOF_SMOKE_PDF=%s not found: %v", p, err)
		}
		return p
	}
	_, thisFile, _, _ := runtime.Caller(0)
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(thisFile), "*.pdf"))
	if err != nil || len(matches) == 0 {
		t.Skipf("no PDF fixture in %s; set PROOF_SMOKE_PDF to override", filepath.Dir(thisFile))
	}
	return matches[0]
}

// TestCreateDeleteBusinessDraftTransaction creates a draft business
// transaction (which does not trigger signer notifications) and tears it down
// in the same test. Requires PROOF_SMOKE_WRITE=1.
func TestCreateDeleteBusinessDraftTransaction(t *testing.T) {
	requireWriteOptIn(t)

	docPath := smokePDFPath(t)
	email := fmt.Sprintf("smoke+%d@example.com", time.Now().UnixNano())

	stdout, stderr, rc := runProof(t,
		"--pretty=false", "business", "transactions", "create",
		"--email", email,
		"--first-name", "Smoke",
		"--last-name", "Test",
		"--document", docPath,
		"--name", "smoke-test draft",
		"--draft",
	)
	assertSuccess(t, "business transactions create", stdout, stderr, rc)
	assertJSON(t, "business transactions create", stdout)

	var resp struct {
		Id string `json:"id"`
	}
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		t.Fatalf("could not parse transaction create response: %v\n%s", err, stdout)
	}
	if resp.Id == "" {
		t.Fatalf("transaction create response missing id field: %s", stdout)
	}

	txnID := resp.Id
	t.Cleanup(func() {
		out, errOut, code := runProof(t, "business", "transactions", "delete", txnID)
		if code != 0 {
			t.Errorf("LEAKED draft transaction %s — delete failed: exit=%d stdout=%s stderr=%s",
				txnID, code, out, errOut)
		}
	})
}
