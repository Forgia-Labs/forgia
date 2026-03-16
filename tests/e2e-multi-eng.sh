#!/usr/bin/env bash
set -euo pipefail

# Forgia E2E Tests — Multi-Engineer Workflow
# Run from repo root: ./tests/e2e-multi-eng.sh

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

assert_file_not_contains() {
  local desc="$1" needle="$2" file="$3"
  TOTAL=$((TOTAL + 1))
  if [[ -f "$file" ]] && ! grep -q "$needle" "$file"; then
    printf "  ${GREEN}PASS${NC} %s\n" "$desc"
    PASSED=$((PASSED + 1))
  else
    printf "  ${RED}FAIL${NC} %s\n" "$desc"
    printf "    file: %s, expected NOT to contain: %s\n" "$file" "$needle"
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

echo "=== Forgia E2E Tests — Multi-Engineer ==="
echo "  Test dir: $TEST_DIR"
echo ""

cd "$TEST_DIR"
git init --initial-branch=main -q
git config user.name "TestUser"
git config user.email "test@example.com"

# =====================================================
echo "--- forgia init: .gitignore ---"
# =====================================================

"$FORGIA" init >/dev/null 2>&1

assert_file_exists "creates .forgia/.gitignore" "$TEST_DIR/.forgia/.gitignore"

gitignore_content=$(cat "$TEST_DIR/.forgia/.gitignore")
assert_contains "gitignore excludes logs/" "logs/" "$gitignore_content"
assert_contains "gitignore excludes run/" "run/" "$gitignore_content"
assert_contains "gitignore excludes *.pid" '*.pid' "$gitignore_content"
assert_contains "gitignore excludes .beads/" ".beads/" "$gitignore_content"

echo ""

# =====================================================
echo "--- forgia init: .gitignore not overwritten ---"
# =====================================================

echo "# custom" > "$TEST_DIR/.forgia/.gitignore"
"$FORGIA" init <<< "y" >/dev/null 2>&1

custom_content=$(cat "$TEST_DIR/.forgia/.gitignore")
assert_eq "gitignore not overwritten" "# custom" "$custom_content"

echo ""

# =====================================================
echo "--- forgia init: CODEOWNERS (no .github/) ---"
# =====================================================

rm -rf .forgia
"$FORGIA" init >/dev/null 2>&1

TOTAL=$((TOTAL + 1))
if [[ ! -f ".github/CODEOWNERS" ]]; then
  printf "  ${GREEN}PASS${NC} %s\n" "no CODEOWNERS when .github/ missing"
  PASSED=$((PASSED + 1))
else
  printf "  ${RED}FAIL${NC} %s\n" "no CODEOWNERS when .github/ missing"
  FAILED=$((FAILED + 1))
fi

echo ""

# =====================================================
echo "--- forgia init: CODEOWNERS (with .github/) ---"
# =====================================================

rm -rf .forgia
mkdir -p .github
"$FORGIA" init >/dev/null 2>&1

assert_file_exists "creates CODEOWNERS" ".github/CODEOWNERS"

co_content=$(cat ".github/CODEOWNERS")
assert_contains "CODEOWNERS has constitution" "constitution.md" "$co_content"
assert_contains "CODEOWNERS has guardrails" "guardrails/" "$co_content"
assert_contains "CODEOWNERS has architecture" "architecture/" "$co_content"
assert_contains "CODEOWNERS has fd/ open to all" '.forgia/fd/' "$co_content"
assert_contains "CODEOWNERS uses git user" "TestUser" "$co_content"

echo ""

# =====================================================
echo "--- forgia init: CODEOWNERS not overwritten ---"
# =====================================================

echo "# custom owners" > ".github/CODEOWNERS"
rm -rf .forgia
"$FORGIA" init >/dev/null 2>&1

custom_co=$(cat ".github/CODEOWNERS")
assert_eq "CODEOWNERS not overwritten" "# custom owners" "$custom_co"

echo ""

# =====================================================
echo "--- FD template: author field ---"
# =====================================================

fd_template=$(cat "$TEST_DIR/.forgia/fd/_templates/fd-template.md")
assert_contains "FD template has author field" "author:" "$fd_template"
assert_contains "FD template has author placeholder" '{{AUTHOR}}' "$fd_template"

echo ""

# =====================================================
echo "--- forgia status: shows author ---"
# =====================================================

cat > "$TEST_DIR/.forgia/fd/FD-001-test.md" <<'FD'
---
id: "FD-001"
title: "Test Feature"
status: planned
author: "ferruvich"
---
FD

output=$("$FORGIA" status 2>&1)
assert_contains "status shows author" "ferruvich" "$output"

echo ""

# =====================================================
echo "--- forgia status: no author graceful ---"
# =====================================================

cat > "$TEST_DIR/.forgia/fd/FD-002-no-author.md" <<'FD'
---
id: "FD-002"
title: "No Author Feature"
status: planned
---
FD

output=$("$FORGIA" status 2>&1)
assert_contains "status shows FD without author" "FD-002" "$output"
assert_contains "status shows title without author" "No Author Feature" "$output"

echo ""

# =====================================================
echo "--- .gitignore: logs excluded from git ---"
# =====================================================

mkdir -p "$TEST_DIR/.forgia/logs"
echo "test log" > "$TEST_DIR/.forgia/logs/test.log"
git add .forgia/ 2>/dev/null || true
git_status=$(git status --short .forgia/logs/ 2>/dev/null)
TOTAL=$((TOTAL + 1))
# logs/ should not appear in git status (ignored)
if [[ -z "$git_status" ]] || echo "$git_status" | grep -q '!!'; then
  printf "  ${GREEN}PASS${NC} %s\n" "logs/ excluded from git tracking"
  PASSED=$((PASSED + 1))
else
  printf "  ${RED}FAIL${NC} %s\n" "logs/ excluded from git tracking"
  printf "    git status: %s\n" "$git_status"
  FAILED=$((FAILED + 1))
fi

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
