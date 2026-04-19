//go:build smoke || smoke_write

// Shared helpers for the read-only smoke tests (tag: smoke) and the opt-in
// write smoke tests (tag: smoke_write). Build-tagged with OR so either tag
// pulls this file in, but a plain `go test ./...` still sees "no test files".
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
