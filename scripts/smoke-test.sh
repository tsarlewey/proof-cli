#!/usr/bin/env bash
# smoke-test.sh — run read-only smoke tests against the live Proof API.
#
# Usage:
#   make smoke                   # build + run all non-SCIM smoke tests
#   PROOF_SMOKE_ORG_ID=abc make smoke   # also run SCIM tests
#
# The script hits only read-only endpoints (list/get/verify-address) and does
# not mutate any resources.

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

run_test() {
    local name="$1"
    shift
    local out rc
    set +e
    out=$("$@" 2>&1)
    rc=$?
    set -e 2>/dev/null || true
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

echo "Running smoke tests with $PROOF_BIN"

echo
echo "=== Basic commands (no API) ==="
run_test "help"           "$PROOF_BIN" --help
run_test "config get"     "$PROOF_BIN" config get

echo
echo "=== Business API (read-only) ==="
run_test "business transactions list"      "$PROOF_BIN" --pretty=false business transactions list --limit 1
run_test "business webhooks subscriptions" "$PROOF_BIN" --pretty=false business webhooks subscriptions
run_test "business notaries list"          "$PROOF_BIN" --pretty=false business notaries list
run_test "business templates list"         "$PROOF_BIN" --pretty=false business templates list --limit 1

echo
echo "=== Real Estate API (read-only) ==="
run_test "real-estate transactions list" "$PROOF_BIN" --pretty=false real-estate transactions list --limit 1
run_test "real-estate webhooks list"     "$PROOF_BIN" --pretty=false real-estate webhooks list
run_test "real-estate verify-address"    "$PROOF_BIN" --pretty=false real-estate verify-address \
    --line1 "1600 Pennsylvania Ave" --city "Washington" --state "DC" --postal-code "20500"

if [ -n "${PROOF_SMOKE_ORG_ID:-}" ]; then
    echo
    echo "=== SCIM API (read-only) ==="
    run_test "scim users list"   "$PROOF_BIN" --pretty=false scim users list "$PROOF_SMOKE_ORG_ID" --count 1
    run_test "scim schemas user" "$PROOF_BIN" --pretty=false scim schemas user "$PROOF_SMOKE_ORG_ID"
else
    echo
    echo "Skipping SCIM tests (set PROOF_SMOKE_ORG_ID to enable)"
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
