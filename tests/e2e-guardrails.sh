#!/usr/bin/env bash
set -euo pipefail

# Forgia E2E Tests — Guardrails
# Run from repo root: ./tests/e2e-guardrails.sh

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FORGIA="$ROOT_DIR/bin/forgia"
TEST_DIR=$(mktemp -d)
PASSED=0
FAILED=0
TOTAL=0

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

cleanup() {
  rm -rf "$TEST_DIR"
}
trap cleanup EXIT

# --- Helpers ---

assert_contains() {
  local desc="$1" needle="$2" haystack="$3"
  TOTAL=$((TOTAL + 1))
  if echo "$haystack" | grep -q "$needle"; then
    printf "  ${GREEN}PASS${NC} %s\n" "$desc"
    PASSED=$((PASSED + 1))
  else
    printf "  ${RED}FAIL${NC} %s\n" "$desc"
    printf "    expected to contain: %s\n" "$needle"
    FAILED=$((FAILED + 1))
  fi
}

assert_not_contains() {
  local desc="$1" needle="$2" haystack="$3"
  TOTAL=$((TOTAL + 1))
  if ! echo "$haystack" | grep -q "$needle"; then
    printf "  ${GREEN}PASS${NC} %s\n" "$desc"
    PASSED=$((PASSED + 1))
  else
    printf "  ${RED}FAIL${NC} %s\n" "$desc"
    printf "    expected NOT to contain: %s\n" "$needle"
    FAILED=$((FAILED + 1))
  fi
}

assert_file_exists() {
  local desc="$1" file="$2"
  TOTAL=$((TOTAL + 1))
  if [[ -f "$file" ]]; then
    printf "  ${GREEN}PASS${NC} %s\n" "$desc"
    PASSED=$((PASSED + 1))
  else
    printf "  ${RED}FAIL${NC} %s\n" "$desc"
    printf "    file not found: %s\n" "$file"
    FAILED=$((FAILED + 1))
  fi
}

assert_file_contains() {
  local desc="$1" needle="$2" file="$3"
  TOTAL=$((TOTAL + 1))
  if [[ -f "$file" ]] && grep -q "$needle" "$file"; then
    printf "  ${GREEN}PASS${NC} %s\n" "$desc"
    PASSED=$((PASSED + 1))
  else
    printf "  ${RED}FAIL${NC} %s\n" "$desc"
    printf "    file: %s, expected to contain: %s\n" "$file" "$needle"
    FAILED=$((FAILED + 1))
  fi
}

assert_eq() {
  local desc="$1" expected="$2" actual="$3"
  TOTAL=$((TOTAL + 1))
  if [[ "$expected" == "$actual" ]]; then
    printf "  ${GREEN}PASS${NC} %s\n" "$desc"
    PASSED=$((PASSED + 1))
  else
    printf "  ${RED}FAIL${NC} %s\n" "$desc"
    printf "    expected: %s\n" "$expected"
    printf "    actual:   %s\n" "$actual"
    FAILED=$((FAILED + 1))
  fi
}

# --- Tests ---

echo "=== Forgia E2E Tests — Guardrails ==="
echo "  Test dir: $TEST_DIR"
echo ""

cd "$TEST_DIR"
git init --initial-branch=main -q

# =====================================================
echo "--- forgia init: guardrails scaffolding ---"
# =====================================================

"$FORGIA" init >/dev/null 2>&1

assert_file_exists "creates guardrails/" "$TEST_DIR/.forgia/guardrails/deny.toml"
assert_file_exists "creates guardrails/ignore" "$TEST_DIR/.forgia/guardrails/ignore"

echo ""

# =====================================================
echo "--- deny.toml: structure ---"
# =====================================================

deny="$TEST_DIR/.forgia/guardrails/deny.toml"

assert_file_contains "has [read] section" '^\[read\]' "$deny"
assert_file_contains "has [execute] section" '^\[execute\]' "$deny"
assert_file_contains "has [write] section" '^\[write\]' "$deny"

echo ""

# =====================================================
echo "--- deny.toml: read patterns ---"
# =====================================================

deny_content=$(cat "$deny")

assert_contains "denies .env" '.env' "$deny_content"
assert_contains "denies *.pem" '*.pem' "$deny_content"
assert_contains "denies *.key" '*.key' "$deny_content"
assert_contains "denies aws credentials" '.aws/credentials' "$deny_content"
assert_contains "denies ssh keys" '.ssh/id_' "$deny_content"
assert_contains "denies gnupg" '.gnupg/' "$deny_content"
assert_contains "denies password-store" '.password-store/' "$deny_content"
assert_contains "denies kubeconfig" 'kubeconfig' "$deny_content"
assert_contains "denies tfstate" '.tfstate' "$deny_content"
assert_contains "allows .env.example (negation)" '.env.example' "$deny_content"

echo ""

# =====================================================
echo "--- deny.toml: execute patterns ---"
# =====================================================

assert_contains "denies cat ssh keys" 'cat ~/.ssh/id_' "$deny_content"
assert_contains "denies gpg export" 'gpg --export-secret' "$deny_content"
assert_contains "denies pass show" 'pass show' "$deny_content"
assert_contains "denies bw get" 'bw get' "$deny_content"
assert_contains "denies op item get" 'op item get' "$deny_content"
assert_contains "denies env grep key" 'env | grep -i key' "$deny_content"
assert_contains "denies env grep token" 'env | grep -i token' "$deny_content"
assert_contains "denies env grep secret" 'env | grep -i secret' "$deny_content"
assert_contains "denies echo API keys" 'echo \$ANTHROPIC_API_KEY' "$deny_content"
assert_contains "denies security find-password" 'security find-generic-password' "$deny_content"

echo ""

# =====================================================
echo "--- deny.toml: write patterns ---"
# =====================================================

assert_contains "denies writing .env" '.env' "$deny_content"
assert_contains "denies writing constitution" 'constitution.md' "$deny_content"
assert_contains "denies writing config.toml" 'config.toml' "$deny_content"
assert_contains "denies writing deny.toml itself" 'deny.toml' "$deny_content"
assert_contains "denies writing *.pem" '*.pem' "$deny_content"

echo ""

# =====================================================
echo "--- ignore: patterns ---"
# =====================================================

ignore="$TEST_DIR/.forgia/guardrails/ignore"
ignore_content=$(cat "$ignore")

assert_contains "ignores .env" '.env' "$ignore_content"
assert_contains "ignores *.pem" '*.pem' "$ignore_content"
assert_contains "ignores .gnupg/" '.gnupg/' "$ignore_content"
assert_contains "ignores node_modules/" 'node_modules/' "$ignore_content"
assert_contains "ignores target/" 'target/' "$ignore_content"
assert_contains "ignores __pycache__/" '__pycache__/' "$ignore_content"
assert_contains "ignores .DS_Store" '.DS_Store' "$ignore_content"
assert_contains "allows .env.example (negation)" '!.env.example' "$ignore_content"

echo ""

# =====================================================
echo "--- forgia init: guardrails not overwritten ---"
# =====================================================

echo "# Custom deny rules" > "$TEST_DIR/.forgia/guardrails/deny.toml"
echo "# Custom ignore" > "$TEST_DIR/.forgia/guardrails/ignore"
"$FORGIA" init <<< "y" >/dev/null 2>&1

deny_after=$(cat "$TEST_DIR/.forgia/guardrails/deny.toml")
ignore_after=$(cat "$TEST_DIR/.forgia/guardrails/ignore")
assert_eq "deny.toml not overwritten" "# Custom deny rules" "$deny_after"
assert_eq "ignore not overwritten" "# Custom ignore" "$ignore_after"

echo ""

# =====================================================
echo "--- forgia doctor: guardrails check ---"
# =====================================================

# Reset with real guardrails
rm -rf .forgia
"$FORGIA" init >/dev/null 2>&1

output=$("$FORGIA" doctor 2>&1)
assert_contains "doctor checks deny.toml" "guardrails" "$output"
assert_contains "doctor shows OK for deny.toml" "OK" "$output"

echo ""

# =====================================================
echo "--- forgia doctor: missing guardrails ---"
# =====================================================

rm -f "$TEST_DIR/.forgia/guardrails/deny.toml"
output=$("$FORGIA" doctor 2>&1)
assert_contains "doctor warns missing deny.toml" "MISSING" "$output"

echo ""

# =====================================================
echo "--- constitution: security section ---"
# =====================================================

rm -rf .forgia
"$FORGIA" init >/dev/null 2>&1

const_content=$(cat "$TEST_DIR/.forgia/constitution.md")
assert_contains "constitution has Security section" "Security" "$const_content"
assert_contains "constitution mentions guardrails" "guardrails" "$const_content"
assert_contains "constitution mentions deny.toml" "deny.toml" "$const_content"
assert_contains "constitution mentions fail-closed" "fail-closed" "$const_content"

echo ""

# =====================================================
echo "--- runner: claude prompt includes guardrails ---"
# =====================================================

# We can't run the full claude runner (needs claude CLI),
# but we can verify the script loads deny.toml

runner_content=$(cat "$ROOT_DIR/modules/runners/claude.sh")
assert_contains "claude runner loads deny.toml" 'guardrails/deny.toml' "$runner_content"
assert_contains "claude runner has Security Guardrails section" 'Security Guardrails' "$runner_content"
assert_contains "claude runner says MANDATORY" 'MANDATORY' "$runner_content"
assert_contains "claude runner says NEVER read" 'NEVER read' "$runner_content"
assert_contains "claude runner says NEVER execute" 'NEVER execute' "$runner_content"
assert_contains "claude runner says NEVER write" 'NEVER write' "$runner_content"

echo ""

# =====================================================
echo "--- runner: openhands prompt includes guardrails ---"
# =====================================================

runner_oh=$(cat "$ROOT_DIR/modules/runners/openhands.sh")
assert_contains "openhands runner loads deny.toml" 'guardrails/deny.toml' "$runner_oh"
assert_contains "openhands runner has Security Guardrails" 'Security Guardrails' "$runner_oh"
assert_contains "openhands runner says MANDATORY" 'MANDATORY' "$runner_oh"

echo ""

# =====================================================
echo "--- claude-commands: fd-review checks security ---"
# =====================================================

review_content=$(cat "$ROOT_DIR/modules/claude-commands/fd-review.md")
assert_contains "fd-review has Security Compliance" 'Security' "$review_content"
assert_contains "fd-review checks guardrails" 'guardrails' "$review_content"
assert_contains "fd-review checks plaintext secrets" 'plaintext' "$review_content"

echo ""

# =====================================================
echo "--- claude-commands: fd-sdd loads guardrails ---"
# =====================================================

sdd_content=$(cat "$ROOT_DIR/modules/claude-commands/fd-sdd.md")
assert_contains "fd-sdd loads guardrails" 'guardrails' "$sdd_content"
assert_contains "fd-sdd loads deny.toml" 'deny.toml' "$sdd_content"

echo ""

# =====================================================
# Summary
# =====================================================

echo "=== Results ==="
echo ""
printf "  Total:  %d\n" "$TOTAL"
printf "  ${GREEN}Passed: %d${NC}\n" "$PASSED"
if (( FAILED > 0 )); then
  printf "  ${RED}Failed: %d${NC}\n" "$FAILED"
else
  printf "  Failed: %d\n" "$FAILED"
fi
echo ""

if (( FAILED > 0 )); then
  echo "FAILED"
  exit 1
else
  echo "ALL TESTS PASSED"
  exit 0
fi
