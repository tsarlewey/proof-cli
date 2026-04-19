#!/usr/bin/env bash
# smoke-test-write.sh — run opt-in write smoke tests against the live Proof API.
#
# Each test creates a real resource and deletes it in the same run. Requires
# PROOF_SMOKE_WRITE=1 to opt in.
#
# Usage:
#   PROOF_SMOKE_WRITE=1 make smoke-write

set -u
set -o pipefail

PROOF_BIN=${PROOF_BIN:-./proof}
PASS=0
FAIL=0
FAILURES=()

if [ ! -x "$PROOF_BIN" ]; then
    echo "Binary not found or not executable at $PROOF_BIN."
    echo "Run 'make build' first, or set PROOF_BIN to the binary path."
    exit 2
fi

if [ "${PROOF_SMOKE_WRITE:-}" != "1" ]; then
    echo "Refusing to run write smokes: PROOF_SMOKE_WRITE != 1."
    echo "Set PROOF_SMOKE_WRITE=1 explicitly to opt in."
    exit 2
fi

if ! command -v jq >/dev/null 2>&1; then
    echo "This script needs 'jq' to extract resource IDs from responses."
    echo "Install jq (e.g. 'brew install jq') and retry."
    exit 2
fi

record() {
    local name="$1" rc="$2" out="$3"
    if [ "$rc" -eq 0 ]; then
        printf '  PASS  %s\n' "$name"
        PASS=$((PASS + 1))
    else
        printf '  FAIL  %s (exit %d)\n' "$name" "$rc"
        printf '    --- output ---\n'
        printf '%s\n' "$out" | sed 's/^/    /'
        printf '    --------------\n'
        FAIL=$((FAIL + 1))
        FAILURES+=("$name")
    fi
}

# Resolve the PDF fixture. If PROOF_SMOKE_PDF is set, use it; otherwise pick
# the first *.pdf in test/smoke/.
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
if [ -n "${PROOF_SMOKE_PDF:-}" ]; then
    PDF_PATH="$PROOF_SMOKE_PDF"
else
    PDF_PATH=$(find "$SCRIPT_DIR/../test/smoke" -maxdepth 1 -name '*.pdf' -type f 2>/dev/null | head -n 1)
fi
if [ -z "$PDF_PATH" ] || [ ! -r "$PDF_PATH" ]; then
    echo "No PDF fixture found under test/smoke/ and PROOF_SMOKE_PDF is not set."
    echo "Drop a PDF into test/smoke/ or set PROOF_SMOKE_PDF to a valid path."
    exit 2
fi

echo "Running write smoke tests with $PROOF_BIN"

echo
echo "=== Business draft transactions (create + delete) ==="
email="smoke+$(date +%s)@example.com"

set +e
create_out=$("$PROOF_BIN" --pretty=false business transactions create \
    --email "$email" \
    --first-name "Smoke" \
    --last-name "Test" \
    --document "$PDF_PATH" \
    --name "smoke-test draft" \
    --draft 2>&1)
create_rc=$?
set -e 2>/dev/null || true
record "business transactions create (draft)" "$create_rc" "$create_out"

if [ "$create_rc" -eq 0 ]; then
    txn_id=$(printf '%s' "$create_out" | jq -r '.id // empty')
    if [ -z "$txn_id" ]; then
        echo "  FAIL  business transactions create — could not extract id from response"
        printf '%s\n' "$create_out" | sed 's/^/    /'
        FAIL=$((FAIL + 1))
        FAILURES+=("business transactions create (no id)")
    else
        set +e
        delete_out=$("$PROOF_BIN" business transactions delete "$txn_id" 2>&1)
        delete_rc=$?
        set -e 2>/dev/null || true
        record "business transactions delete ($txn_id)" "$delete_rc" "$delete_out"
        if [ "$delete_rc" -ne 0 ]; then
            echo "  !! LEAKED transaction $txn_id — investigate and delete manually"
        fi
    fi
fi

echo
echo "=== Summary ==="
printf 'Passed: %d\n' "$PASS"
printf 'Failed: %d\n' "$FAIL"
if [ "$FAIL" -gt 0 ]; then
    echo "Failures:"
    for f in "${FAILURES[@]}"; do printf '  - %s\n' "$f"; done
    exit 1
fi
