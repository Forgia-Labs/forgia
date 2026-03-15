#!/usr/bin/env bash
set -euo pipefail

# E2E tests for: modules/context-discovery.sh
# Run from repo root: ./tests/e2e-context-discovery.sh

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT="$ROOT_DIR/modules/context-discovery.sh"
TEST_DIR=$(mktemp -d)
PASSED=0
FAILED=0
TOTAL=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

cleanup() {
  rm -rf "$TEST_DIR"
}
trap cleanup EXIT

# --- Test helpers ---

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

assert_contains() {
  local desc="$1" needle="$2" haystack="$3"
  TOTAL=$((TOTAL + 1))
  if echo "$haystack" | grep -qF -- "$needle"; then
    printf "  ${GREEN}PASS${NC} %s\n" "$desc"
    PASSED=$((PASSED + 1))
  else
    printf "  ${RED}FAIL${NC} %s\n" "$desc"
    printf "    expected to contain: %s\n" "$needle"
    printf "    actual: %s\n" "$haystack"
    FAILED=$((FAILED + 1))
  fi
}

assert_not_contains() {
  local desc="$1" needle="$2" haystack="$3"
  TOTAL=$((TOTAL + 1))
  if ! echo "$haystack" | grep -qF -- "$needle"; then
    printf "  ${GREEN}PASS${NC} %s\n" "$desc"
    PASSED=$((PASSED + 1))
  else
    printf "  ${RED}FAIL${NC} %s\n" "$desc"
    printf "    expected NOT to contain: %s\n" "$needle"
    FAILED=$((FAILED + 1))
  fi
}

assert_exit_code() {
  local desc="$1" expected="$2" actual="$3"
  TOTAL=$((TOTAL + 1))
  if [[ "$expected" == "$actual" ]]; then
    printf "  ${GREEN}PASS${NC} %s\n" "$desc"
    PASSED=$((PASSED + 1))
  else
    printf "  ${RED}FAIL${NC} %s\n" "$desc"
    printf "    expected exit code: %s, got: %s\n" "$expected" "$actual"
    FAILED=$((FAILED + 1))
  fi
}

# Create a fresh workspace in TEST_DIR
setup_workspace() {
  local ws="$TEST_DIR/ws_$$_$RANDOM"
  mkdir -p "$ws"
  echo "$ws"
}

# --- Tests ---

echo "=== Context Discovery Tests ==="

# Test 1: Empty directory returns exit code 1
echo ""
echo "--- Test: empty directory returns 1 ---"
ws=$(setup_workspace)
cd "$ws"
output=$("$SCRIPT" 2>&1) && ec=0 || ec=$?
assert_exit_code "empty dir returns exit code 1" "1" "$ec"
assert_eq "empty dir produces no output" "" "$output"

# Test 2: Discover ADR files
echo ""
echo "--- Test: discover ADR files ---"
ws=$(setup_workspace)
cd "$ws"
mkdir -p docs/decisions docs/adr
echo "# ADR 001" > docs/decisions/001-kustomize.md
echo "# ADR 002" > docs/decisions/002-helm.md
echo "# ADR 003" > docs/adr/003-docker.md
output=$("$SCRIPT" --format=markdown) && ec=0 || ec=$?
assert_exit_code "ADR discovery exits 0" "0" "$ec"
assert_contains "finds decisions dir" "docs/decisions/001-kustomize.md" "$output"
assert_contains "finds decisions dir 2" "docs/decisions/002-helm.md" "$output"
assert_contains "finds adr dir" "docs/adr/003-docker.md" "$output"
assert_contains "ADR category" "ADR" "$output"

# Test 3: Discover CONTRIBUTING.md and ARCHITECTURE.md
echo ""
echo "--- Test: discover CONTRIBUTING.md and ARCHITECTURE.md ---"
ws=$(setup_workspace)
cd "$ws"
echo "# Contributing" > CONTRIBUTING.md
echo "# Arch" > ARCHITECTURE.md
output=$("$SCRIPT" --format=markdown) && ec=0 || ec=$?
assert_exit_code "exits 0" "0" "$ec"
assert_contains "finds CONTRIBUTING.md" "CONTRIBUTING.md" "$output"
assert_contains "Contributing category" "Contributing" "$output"
assert_contains "finds ARCHITECTURE.md" "ARCHITECTURE.md" "$output"
assert_contains "Architecture category" "Architecture" "$output"

# Test 4: Discover .forgia/ context files
echo ""
echo "--- Test: discover .forgia/ context files ---"
ws=$(setup_workspace)
cd "$ws"
mkdir -p .forgia/dev-guide/lang
echo "# Constitution" > .forgia/constitution.md
echo "# Go guide" > .forgia/dev-guide/lang/go.md
output=$("$SCRIPT" --format=markdown) && ec=0 || ec=$?
assert_exit_code "exits 0" "0" "$ec"
assert_contains "finds constitution" ".forgia/constitution.md" "$output"
assert_contains "finds dev-guide" ".forgia/dev-guide/lang/go.md" "$output"
assert_contains "Forgia category" "Forgia" "$output"

# Test 5: Skip files > 100KB
echo ""
echo "--- Test: skip files > 100KB ---"
ws=$(setup_workspace)
cd "$ws"
mkdir -p docs/decisions
echo "# Small" > docs/decisions/small.md
# Create a >100KB file
dd if=/dev/zero bs=1024 count=110 2>/dev/null | tr '\0' 'x' > docs/decisions/huge.md
output=$("$SCRIPT" --format=plain) && ec=0 || ec=$?
assert_exit_code "exits 0" "0" "$ec"
assert_contains "includes small file" "small.md" "$output"
assert_not_contains "skips huge file" "huge.md" "$output"

# Test 6: Skip _templates/ directory
echo ""
echo "--- Test: skip _templates/ directory ---"
ws=$(setup_workspace)
cd "$ws"
mkdir -p _templates docs/decisions
echo "# Template" > _templates/template.md
echo "# Real" > docs/decisions/real.md
output=$("$SCRIPT" --format=plain) && ec=0 || ec=$?
assert_contains "includes real file" "real.md" "$output"
assert_not_contains "skips template" "_templates" "$output"

# Test 7: Max limit (30 files)
echo ""
echo "--- Test: max file limit ---"
ws=$(setup_workspace)
cd "$ws"
mkdir -p docs/decisions
for i in $(seq 1 40); do
  printf -v padded "%03d" "$i"
  echo "# ADR $padded" > "docs/decisions/${padded}-adr.md"
done
output=$("$SCRIPT" --format=plain) && ec=0 || ec=$?
line_count=$(echo "$output" | wc -l | tr -d ' ')
assert_eq "max 30 files returned" "30" "$line_count"

# Test 8: Custom --max
echo ""
echo "--- Test: custom --max ---"
cd "$ws"
output=$("$SCRIPT" --format=plain --max=5) && ec=0 || ec=$?
line_count=$(echo "$output" | wc -l | tr -d ' ')
assert_eq "--max=5 returns 5 files" "5" "$line_count"

# Test 9: Markdown format
echo ""
echo "--- Test: markdown format ---"
ws=$(setup_workspace)
cd "$ws"
echo "# Contributing" > CONTRIBUTING.md
output=$("$SCRIPT" --format=markdown) && ec=0 || ec=$?
assert_contains "markdown header" "### Auto-discovered Context" "$output"
assert_contains "markdown list item" "- \`CONTRIBUTING.md\` — Contributing" "$output"

# Test 10: Plain format
echo ""
echo "--- Test: plain format ---"
ws=$(setup_workspace)
cd "$ws"
echo "# Contributing" > CONTRIBUTING.md
output=$("$SCRIPT" --format=plain) && ec=0 || ec=$?
assert_contains "plain format" "CONTRIBUTING CONTRIBUTING.md" "$output"

# Test 11: JSON format
echo ""
echo "--- Test: JSON format ---"
ws=$(setup_workspace)
cd "$ws"
echo "# Contributing" > CONTRIBUTING.md
echo "# Arch" > ARCHITECTURE.md
output=$("$SCRIPT" --format=json) && ec=0 || ec=$?
assert_contains "json starts with [" "[" "$output"
assert_contains "json has path key" '"path"' "$output"
assert_contains "json has category key" '"category"' "$output"
assert_contains "json contains CONTRIBUTING" "CONTRIBUTING.md" "$output"

# Test 12: --include adds extra paths
echo ""
echo "--- Test: --include adds extra paths ---"
ws=$(setup_workspace)
cd "$ws"
mkdir -p extra
echo "# Extra" > extra/notes.md
echo "# Single" > standalone.md
output=$("$SCRIPT" --format=plain --include=extra --include=standalone.md) && ec=0 || ec=$?
assert_exit_code "--include exits 0" "0" "$ec"
assert_contains "includes dir files" "extra/notes.md" "$output"
assert_contains "includes single file" "standalone.md" "$output"
assert_contains "user-specified category" "USER-SPECIFIED" "$output"

# Test 13: Category priority sorting (ADR before Contributing before Docs)
echo ""
echo "--- Test: category priority sorting ---"
ws=$(setup_workspace)
cd "$ws"
mkdir -p docs/decisions docs
echo "# Doc" > docs/readme.md
echo "# ADR" > docs/decisions/001.md
echo "# Contributing" > CONTRIBUTING.md
output=$("$SCRIPT" --format=plain) && ec=0 || ec=$?
# ADR should come before CONTRIBUTING which should come before DOCS
first_line=$(echo "$output" | head -1)
assert_contains "ADR first" "ADR" "$first_line"

# Test 14: Sourced usage
echo ""
echo "--- Test: sourced usage ---"
ws=$(setup_workspace)
cd "$ws"
echo "# Contributing" > CONTRIBUTING.md
mkdir -p docs/decisions
echo "# ADR" > docs/decisions/001.md
# Source the module and call discover_context
output=$(bash -c "source '$SCRIPT'; discover_context") && ec=0 || ec=$?
assert_exit_code "sourced exits 0" "0" "$ec"
assert_contains "sourced finds contributing" "CONTRIBUTING.md" "$output"
assert_contains "sourced default markdown" "### Auto-discovered Context" "$output"

# Test 15: discover_context_formatted with plain
echo ""
echo "--- Test: sourced discover_context_formatted ---"
ws=$(setup_workspace)
cd "$ws"
echo "# Contributing" > CONTRIBUTING.md
output=$(bash -c "source '$SCRIPT'; discover_context_formatted plain") && ec=0 || ec=$?
assert_contains "formatted plain" "CONTRIBUTING CONTRIBUTING.md" "$output"

# Test 16: CLAUDE.md and .claude/CLAUDE.md discovery
echo ""
echo "--- Test: Agent files discovery ---"
ws=$(setup_workspace)
cd "$ws"
echo "# Claude" > CLAUDE.md
mkdir -p .claude
echo "# Claude inner" > .claude/CLAUDE.md
output=$("$SCRIPT" --format=plain) && ec=0 || ec=$?
assert_contains "finds CLAUDE.md" "CLAUDE.md" "$output"
assert_contains "finds .claude/CLAUDE.md" ".claude/CLAUDE.md" "$output"
assert_contains "Agent category" "AGENT" "$output"

# Test 17: CODE_OF_CONDUCT.md
echo ""
echo "--- Test: CODE_OF_CONDUCT discovery ---"
ws=$(setup_workspace)
cd "$ws"
echo "# CoC" > CODE_OF_CONDUCT.md
output=$("$SCRIPT" --format=plain) && ec=0 || ec=$?
assert_contains "finds CODE_OF_CONDUCT.md" "CODE_OF_CONDUCT.md" "$output"
assert_contains "Community category" "COMMUNITY" "$output"

# Test 18: Performance - should complete in < 1 second
echo ""
echo "--- Test: performance ---"
ws=$(setup_workspace)
cd "$ws"
# Create a bunch of dirs and files to simulate a real repo
for d in $(seq 1 20); do
  mkdir -p "src/pkg$d"
  for f in $(seq 1 50); do
    echo "code" > "src/pkg$d/file$f.go"
  done
done
mkdir -p docs/decisions
echo "# ADR" > docs/decisions/001.md
start_time=$(date +%s)
output=$("$SCRIPT" --format=plain) && ec=0 || ec=$?
end_time=$(date +%s)
elapsed=$((end_time - start_time))
TOTAL=$((TOTAL + 1))
if [[ "$elapsed" -le 1 ]]; then
  printf "  ${GREEN}PASS${NC} completes in <= 1 second (took ${elapsed}s)\n"
  PASSED=$((PASSED + 1))
else
  printf "  ${RED}FAIL${NC} took ${elapsed}s, expected <= 1s\n"
  FAILED=$((FAILED + 1))
fi

# Test 19: Testing docs discovery
echo ""
echo "--- Test: testing docs discovery ---"
ws=$(setup_workspace)
cd "$ws"
mkdir -p docs/testing
echo "# Test matrix" > docs/testing/matrix.md
output=$("$SCRIPT" --format=plain) && ec=0 || ec=$?
assert_contains "finds testing docs" "docs/testing/matrix.md" "$output"
assert_contains "Testing category" "TESTING" "$output"

# Test 20: Operations docs discovery
echo ""
echo "--- Test: operations docs discovery ---"
ws=$(setup_workspace)
cd "$ws"
mkdir -p docs/operations docs/ops
echo "# Runbook" > docs/operations/runbook.md
echo "# Deploy" > docs/ops/deploy.md
output=$("$SCRIPT" --format=plain) && ec=0 || ec=$?
assert_contains "finds operations" "docs/operations/runbook.md" "$output"
assert_contains "finds ops" "docs/ops/deploy.md" "$output"
assert_contains "Operations category" "OPERATIONS" "$output"

# Test 21: Design docs discovery
echo ""
echo "--- Test: design docs discovery ---"
ws=$(setup_workspace)
cd "$ws"
mkdir -p docs/design
echo "# Design" > docs/design/api.md
output=$("$SCRIPT" --format=plain) && ec=0 || ec=$?
assert_contains "finds design docs" "docs/design/api.md" "$output"
assert_contains "Design category" "DESIGN" "$output"

# --- Summary ---

echo ""
echo "=============================="
printf "Results: ${GREEN}$PASSED passed${NC}, "
if [[ "$FAILED" -gt 0 ]]; then
  printf "${RED}$FAILED failed${NC}, "
else
  printf "0 failed, "
fi
echo "$TOTAL total"

if [[ "$FAILED" -gt 0 ]]; then
  exit 1
fi
