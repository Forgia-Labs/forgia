#!/usr/bin/env bash
set -euo pipefail

# Forgia E2E Tests
# Run from repo root: ./tests/e2e.sh

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FORGIA="$ROOT_DIR/bin/forgia"
TEST_DIR=$(mktemp -d)
PASSED=0
FAILED=0
TOTAL=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
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

assert_dir_exists() {
  local desc="$1" dir="$2"
  TOTAL=$((TOTAL + 1))
  if [[ -d "$dir" ]]; then
    printf "  ${GREEN}PASS${NC} %s\n" "$desc"
    PASSED=$((PASSED + 1))
  else
    printf "  ${RED}FAIL${NC} %s\n" "$desc"
    printf "    dir not found: %s\n" "$dir"
    FAILED=$((FAILED + 1))
  fi
}

assert_exit_code() {
  local desc="$1" expected="$2"
  shift 2
  TOTAL=$((TOTAL + 1))
  local actual
  set +e
  "$@" >/dev/null 2>&1
  actual=$?
  set -e
  if [[ "$expected" == "$actual" ]]; then
    printf "  ${GREEN}PASS${NC} %s\n" "$desc"
    PASSED=$((PASSED + 1))
  else
    printf "  ${RED}FAIL${NC} %s\n" "$desc"
    printf "    expected exit code: %s, got: %s\n" "$expected" "$actual"
    FAILED=$((FAILED + 1))
  fi
}

# --- Tests ---

echo "=== Forgia E2E Tests ==="
echo "  Test dir: $TEST_DIR"
echo ""

# =====================================================
echo "--- forgia init ---"
# =====================================================

cd "$TEST_DIR"
git init --initial-branch=main -q

# Run init (no interactive prompt since .forgia doesn't exist)
output=$("$FORGIA" init 2>&1)

assert_contains "init outputs scaffolding message" "Scaffolding" "$output"
assert_dir_exists "creates .forgia/" "$TEST_DIR/.forgia"
assert_file_exists "creates constitution.md" "$TEST_DIR/.forgia/constitution.md"
assert_file_exists "creates config.toml" "$TEST_DIR/.forgia/config.toml"
assert_file_exists "creates _dashboard.md" "$TEST_DIR/.forgia/_dashboard.md"
assert_dir_exists "creates fd/" "$TEST_DIR/.forgia/fd"
assert_dir_exists "creates sdd/" "$TEST_DIR/.forgia/sdd"
assert_dir_exists "creates ops/" "$TEST_DIR/.forgia/ops"
assert_dir_exists "creates dev-guide/" "$TEST_DIR/.forgia/dev-guide"
assert_dir_exists "creates dev-guide/principles/" "$TEST_DIR/.forgia/dev-guide/principles"
assert_dir_exists "creates dev-guide/lang/" "$TEST_DIR/.forgia/dev-guide/lang"
assert_file_exists "creates fd template" "$TEST_DIR/.forgia/fd/_templates/fd-template.md"
assert_file_exists "creates sdd template" "$TEST_DIR/.forgia/sdd/_templates/sdd-template.md"
assert_file_exists "creates ops template" "$TEST_DIR/.forgia/ops/_templates/ops-task.md"
assert_file_exists "creates shell conventions (always)" "$TEST_DIR/.forgia/dev-guide/lang/shell.md"
assert_file_exists "creates clean-code principles" "$TEST_DIR/.forgia/dev-guide/principles/clean-code.md"
assert_file_exists "creates solid principles" "$TEST_DIR/.forgia/dev-guide/principles/solid.md"
assert_file_exists "creates design-patterns principles" "$TEST_DIR/.forgia/dev-guide/principles/design-patterns.md"

echo ""

# =====================================================
echo "--- forgia init (language detection) ---"
# =====================================================

# Test Rust detection
cd "$TEST_DIR"
touch Cargo.toml
rm -rf .forgia
output=$("$FORGIA" init 2>&1)
assert_contains "detects Rust" "rust" "$output"
assert_file_exists "copies rust conventions" "$TEST_DIR/.forgia/dev-guide/lang/rust.md"
rm Cargo.toml

# Test Python detection
rm -rf .forgia
touch pyproject.toml
output=$("$FORGIA" init 2>&1)
assert_contains "detects Python" "python" "$output"
assert_file_exists "copies python conventions" "$TEST_DIR/.forgia/dev-guide/lang/python.md"
rm pyproject.toml

# Test TypeScript detection
rm -rf .forgia
touch tsconfig.json
output=$("$FORGIA" init 2>&1)
assert_contains "detects TypeScript" "typescript" "$output"
assert_file_exists "copies typescript conventions" "$TEST_DIR/.forgia/dev-guide/lang/typescript.md"
rm tsconfig.json

# Test Go detection
rm -rf .forgia
touch go.mod
output=$("$FORGIA" init 2>&1)
assert_contains "detects Go" "go" "$output"
assert_file_exists "copies go conventions" "$TEST_DIR/.forgia/dev-guide/lang/go.md"
rm go.mod

echo ""

# =====================================================
echo "--- forgia init (idempotency) ---"
# =====================================================

rm -rf .forgia
"$FORGIA" init >/dev/null 2>&1

# Modify constitution (should NOT be overwritten)
echo "# Custom rules" > "$TEST_DIR/.forgia/constitution.md"
"$FORGIA" init <<< "y" >/dev/null 2>&1
content=$(cat "$TEST_DIR/.forgia/constitution.md")
assert_eq "constitution not overwritten" "# Custom rules" "$content"

echo ""

# =====================================================
echo "--- forgia status ---"
# =====================================================

rm -rf .forgia
"$FORGIA" init >/dev/null 2>&1

output=$("$FORGIA" status 2>&1)
assert_contains "shows dashboard header" "Forgia Dashboard" "$output"
assert_contains "shows FD section" "Feature Designs" "$output"
assert_contains "shows SDD section" "Execution Specs" "$output"
assert_contains "shows no FDs message" "no FDs yet" "$output"
assert_contains "shows summary" "Summary:" "$output"

echo ""

# =====================================================
echo "--- forgia status (with FDs) ---"
# =====================================================

cat > "$TEST_DIR/.forgia/fd/FD-001-test.md" <<'FD'
---
id: FD-001
title: "Test Feature"
status: planned
---
FD

output=$("$FORGIA" status 2>&1)
assert_contains "shows FD-001" "FD-001" "$output"
assert_contains "shows FD title" "Test Feature" "$output"
assert_contains "shows planned status" "planned" "$output"
assert_contains "summary shows 1 FD" "1 FDs" "$output"

echo ""

# =====================================================
echo "--- forgia status (with SDDs) ---"
# =====================================================

mkdir -p "$TEST_DIR/.forgia/sdd/FD-001"
cat > "$TEST_DIR/.forgia/sdd/FD-001/SDD-001-api.md" <<'SDD'
---
id: SDD-001
title: "API Module"
status: planned
agent: ""
---
SDD

output=$("$FORGIA" status 2>&1)
assert_contains "shows SDD-001" "SDD-001" "$output"
assert_contains "shows SDD title" "API Module" "$output"
assert_contains "summary shows 1 SDD" "1 SDDs" "$output"

echo ""

# =====================================================
echo "--- forgia doctor ---"
# =====================================================

output=$("$FORGIA" doctor 2>&1)
assert_contains "shows doctor header" "Forgia Doctor" "$output"
assert_contains "checks vault" "vault" "$output"
assert_contains "checks constitution" "constitution" "$output"
assert_contains "shows summary" "Summary:" "$output"

echo ""

# =====================================================
echo "--- forgia (no args) ---"
# =====================================================

assert_exit_code "no args exits 1" 1 "$FORGIA"

output=$("$FORGIA" 2>&1 || true)
assert_contains "shows usage" "Usage:" "$output"

echo ""

# =====================================================
echo "--- forgia (unknown command) ---"
# =====================================================

assert_exit_code "unknown command exits 1" 1 "$FORGIA" foobar

echo ""

# =====================================================
echo "--- forgia status (no vault) ---"
# =====================================================

cd "$TEST_DIR"
rm -rf .forgia
assert_exit_code "status without vault exits 1" 1 "$FORGIA" status

echo ""

# =====================================================
echo "--- forgia exec (missing file) ---"
# =====================================================

assert_exit_code "exec with missing file exits 1" 1 "$FORGIA" exec nonexistent.md

echo ""

# =====================================================
echo "--- runner resolution ---"
# =====================================================

rm -rf .forgia
"$FORGIA" init >/dev/null 2>&1

# Default runner from config.toml should be "claude"
config_runner=$(grep 'default' "$TEST_DIR/.forgia/config.toml" | head -1 | sed 's/.*= *"*//;s/"*$//')
assert_eq "default runner is claude" "claude" "$config_runner"

echo ""

# =====================================================
echo "--- template content validation ---"
# =====================================================

# Constitution has required sections
const_content=$(cat "$TEST_DIR/.forgia/constitution.md")
assert_contains "constitution has Principles" "Principles" "$const_content"
assert_contains "constitution has Code Standards" "Code Standards" "$const_content"
assert_contains "constitution has Commit Conventions" "Commit Conventions" "$const_content"

# FD template has required sections
fd_content=$(cat "$TEST_DIR/.forgia/fd/_templates/fd-template.md")
assert_contains "fd template has Problem" "Problem" "$fd_content"
assert_contains "fd template has Architecture" "Architecture" "$fd_content"
assert_contains "fd template has mermaid" "mermaid" "$fd_content"
assert_contains "fd template has Interfaces" "Interfaces" "$fd_content"
assert_contains "fd template has SDD Previsti" "SDD Previsti" "$fd_content"
assert_contains "fd template has Verification" "Verification" "$fd_content"

# SDD template has required sections
sdd_content=$(cat "$TEST_DIR/.forgia/sdd/_templates/sdd-template.md")
assert_contains "sdd template has Scope" "Scope" "$sdd_content"
assert_contains "sdd template has Interfaces" "Interfaces" "$sdd_content"
assert_contains "sdd template has Constraints" "Constraints" "$sdd_content"
assert_contains "sdd template has Test Requirements" "Test Requirements" "$sdd_content"
assert_contains "sdd template has Acceptance Criteria" "Acceptance Criteria" "$sdd_content"
assert_contains "sdd template has Work Log" "Work Log" "$sdd_content"

echo ""

# =====================================================
echo "--- config.toml validation ---"
# =====================================================

config_content=$(cat "$TEST_DIR/.forgia/config.toml")
assert_contains "config has [runner]" "[runner]" "$config_content"
assert_contains "config has [runner.claude]" "[runner.claude]" "$config_content"
assert_contains "config has [runner.openhands]" "[runner.openhands]" "$config_content"
assert_contains "config has [watcher]" "[watcher]" "$config_content"

echo ""

# =====================================================
echo "--- context-discovery ---"
# =====================================================

# Helper to create a mock workspace with known context files
setup_context_workspace() {
  local dir="$1"
  mkdir -p "$dir/docs/decisions"
  echo "# ADR 001" > "$dir/docs/decisions/001-kustomize.md"
  echo "# Contributing" > "$dir/CONTRIBUTING.md"
  echo "# Architecture" > "$dir/ARCHITECTURE.md"
  mkdir -p "$dir/.forgia/dev-guide/lang"
  echo "# Go" > "$dir/.forgia/dev-guide/lang/go.md"
  echo "# Constitution" > "$dir/.forgia/constitution.md"
}

CD_SCRIPT="$ROOT_DIR/modules/context-discovery.sh"

# Test: discovers ADR files
cd_ws=$(mktemp -d)
setup_context_workspace "$cd_ws"
cd "$cd_ws"
cd_output=$("$CD_SCRIPT" --format=markdown) && cd_ec=0 || cd_ec=$?
assert_contains "discovers ADR files in docs/decisions/" "docs/decisions/001-kustomize.md" "$cd_output"

# Test: discovers CONTRIBUTING.md
assert_contains "discovers CONTRIBUTING.md" "CONTRIBUTING.md" "$cd_output"

# Test: discovers ARCHITECTURE.md
assert_contains "discovers ARCHITECTURE.md" "ARCHITECTURE.md" "$cd_output"

# Test: discovers .forgia/constitution.md
assert_contains "discovers .forgia/constitution.md" ".forgia/constitution.md" "$cd_output"

# Test: discovers .forgia/dev-guide/ files
assert_contains "discovers .forgia/dev-guide/ files" ".forgia/dev-guide/lang/go.md" "$cd_output"
rm -rf "$cd_ws"

# Test: skips files > 100KB
cd_ws=$(mktemp -d)
cd "$cd_ws"
mkdir -p docs/decisions
echo "# Small" > docs/decisions/small.md
dd if=/dev/zero bs=1024 count=110 2>/dev/null | tr '\0' 'x' > docs/decisions/huge.md
cd_output=$("$CD_SCRIPT" --format=plain) && cd_ec=0 || cd_ec=$?
cd_has_huge=false
echo "$cd_output" | grep -q "huge.md" && cd_has_huge=true
TOTAL=$((TOTAL + 1))
if [[ "$cd_has_huge" == "false" ]]; then
  printf "  ${GREEN}PASS${NC} %s\n" "skips files > 100KB"
  PASSED=$((PASSED + 1))
else
  printf "  ${RED}FAIL${NC} %s\n" "skips files > 100KB"
  FAILED=$((FAILED + 1))
fi
rm -rf "$cd_ws"

# Test: skips _templates/ directory
cd_ws=$(mktemp -d)
cd "$cd_ws"
mkdir -p _templates docs/decisions
echo "# Template" > _templates/template.md
echo "# Real" > docs/decisions/real.md
cd_output=$("$CD_SCRIPT" --format=plain) && cd_ec=0 || cd_ec=$?
cd_has_template=false
echo "$cd_output" | grep -q "_templates" && cd_has_template=true
TOTAL=$((TOTAL + 1))
if [[ "$cd_has_template" == "false" ]]; then
  printf "  ${GREEN}PASS${NC} %s\n" "skips _templates/ directory"
  PASSED=$((PASSED + 1))
else
  printf "  ${RED}FAIL${NC} %s\n" "skips _templates/ directory"
  FAILED=$((FAILED + 1))
fi
rm -rf "$cd_ws"

# Test: limits to 30 files max
cd_ws=$(mktemp -d)
cd "$cd_ws"
mkdir -p docs/decisions
for i in $(seq 1 40); do
  printf -v padded "%03d" "$i"
  echo "# ADR $padded" > "docs/decisions/${padded}-adr.md"
done
cd_output=$("$CD_SCRIPT" --format=plain) && cd_ec=0 || cd_ec=$?
cd_line_count=$(echo "$cd_output" | wc -l | tr -d ' ')
assert_eq "limits to 30 files max" "30" "$cd_line_count"
rm -rf "$cd_ws"

# Test: returns exit 1 on empty directory
cd_ws=$(mktemp -d)
cd "$cd_ws"
set +e
"$CD_SCRIPT" >/dev/null 2>&1
cd_ec=$?
set -e
assert_eq "returns exit 1 on empty directory" "1" "$cd_ec"
rm -rf "$cd_ws"

# Test: --format=markdown produces markdown list
cd_ws=$(mktemp -d)
cd "$cd_ws"
echo "# Contributing" > CONTRIBUTING.md
cd_output=$("$CD_SCRIPT" --format=markdown) && cd_ec=0 || cd_ec=$?
assert_contains "--format=markdown produces markdown list" "### Auto-discovered Context" "$cd_output"
rm -rf "$cd_ws"

# Test: --format=plain produces CATEGORY path format
cd_ws=$(mktemp -d)
cd "$cd_ws"
echo "# Contributing" > CONTRIBUTING.md
cd_output=$("$CD_SCRIPT" --format=plain) && cd_ec=0 || cd_ec=$?
assert_contains "--format=plain produces CATEGORY path format" "CONTRIBUTING CONTRIBUTING.md" "$cd_output"
rm -rf "$cd_ws"

# Test: --include adds extra paths
cd_ws=$(mktemp -d)
cd "$cd_ws"
mkdir -p extra
echo "# Extra" > extra/notes.md
cd_output=$("$CD_SCRIPT" --format=plain --include=extra) && cd_ec=0 || cd_ec=$?
assert_contains "--include adds extra paths" "extra/notes.md" "$cd_output"
rm -rf "$cd_ws"

# Return to TEST_DIR for remaining tests
cd "$TEST_DIR"

echo ""

# =====================================================
echo "--- forgia fd-from-issue ---"
# =====================================================

# Setup: create mock gh symlink
FDI_DIR=$(mktemp -d)
FDI_BIN="$FDI_DIR/bin"
mkdir -p "$FDI_BIN"
ln -sf "$ROOT_DIR/tests/mock-gh.sh" "$FDI_BIN/gh"
FDI_ORIG_PATH="$PATH"

fdi_cleanup() {
  rm -rf "$FDI_DIR"
  export PATH="$FDI_ORIG_PATH"
  cd "$TEST_DIR"
}

# Setup FDI workspace
FDI_WS="$FDI_DIR/ws"
mkdir -p "$FDI_WS"
cd "$FDI_WS"
git init --initial-branch=main -q
git remote add origin "https://github.com/test-org/test-repo.git"
"$FORGIA" init >/dev/null 2>&1

# Test: no args shows usage and exits 1
assert_exit_code "no args shows usage and exits 1" 1 "$FORGIA" fd-from-issue
fdi_output=$("$FORGIA" fd-from-issue 2>&1 || true)
assert_contains "no args shows usage text" "Usage:" "$fdi_output"

# Test: missing gh CLI shows error
fdi_output=$(PATH="/usr/bin:/bin" "$FORGIA" fd-from-issue 42 2>&1 || true)
assert_contains "missing gh CLI shows error" "gh CLI not found" "$fdi_output"

# Test: non-existent issue shows error
export PATH="$FDI_BIN:$FDI_ORIG_PATH"
fdi_output=$("$FORGIA" fd-from-issue 9999 --repo=test-org/test-repo 2>&1 || true)
assert_contains "non-existent issue shows error" "Could not fetch issue" "$fdi_output"

# Test: vault not initialized shows error
fdi_novault=$(mktemp -d)
cd "$fdi_novault"
git init --initial-branch=main -q
git remote add origin "https://github.com/test-org/test-repo.git"
fdi_output=$("$FORGIA" fd-from-issue 42 2>&1 || true)
assert_contains "vault not initialized shows error" "Vault not initialized" "$fdi_output"
rm -rf "$fdi_novault"

# Return to FDI workspace
cd "$FDI_WS"

# Test: creates FD file at correct path
fdi_output=$("$FORGIA" fd-from-issue 42 --repo=test-org/test-repo 2>&1)
fdi_file=$(find "$FDI_WS/.forgia/fd" -maxdepth 1 -name 'FD-*-add-ingress-rate-limiting-to-nginx-module.md' 2>/dev/null | head -1)
assert_file_exists "creates FD file at correct path" "$fdi_file"

# Test: FD filename is kebab-case of title
TOTAL=$((TOTAL + 1))
if [[ -n "$fdi_file" ]] && echo "$fdi_file" | grep -q "add-ingress-rate-limiting-to-nginx-module"; then
  printf "  ${GREEN}PASS${NC} %s\n" "FD filename is kebab-case of title"
  PASSED=$((PASSED + 1))
else
  printf "  ${RED}FAIL${NC} %s\n" "FD filename is kebab-case of title"
  FAILED=$((FAILED + 1))
fi

if [[ -f "$fdi_file" ]]; then
  fdi_content=$(cat "$fdi_file")

  # Test: FD has all template sections
  assert_contains "FD has all template sections (Problem)" "## Problem" "$fdi_content"
  assert_contains "FD has all template sections (Architecture)" "## Architecture" "$fdi_content"
  assert_contains "FD has all template sections (Interfaces)" "## Interfaces" "$fdi_content"
  assert_contains "FD has all template sections (Notes)" "## Notes" "$fdi_content"

  # Test: frontmatter has upstream_issue field
  assert_contains "frontmatter has upstream_issue field" "upstream_issue:" "$fdi_content"
  assert_contains "upstream_issue has repo#number" "test-org/test-repo#42" "$fdi_content"

  # Test: frontmatter tags from labels
  assert_contains "frontmatter tags from labels" "enhancement" "$fdi_content"
  assert_contains "frontmatter tags include nginx" "nginx" "$fdi_content"

  # Test: frontmatter priority derived from labels (enhancement → medium)
  assert_contains "frontmatter priority derived from labels (enhancement → medium)" "priority: medium" "$fdi_content"

  # Test: frontmatter assignee populated
  assert_contains "frontmatter assignee populated" "testuser" "$fdi_content"

  # Test: Problem section contains issue body
  assert_contains "Problem section contains issue body" "nginx ingress module" "$fdi_content"

  # Test: Notes section contains comments
  assert_contains "Notes section contains comments" "limit_req module" "$fdi_content"

  # Test: Notes section contains discovered context
  assert_contains "Notes section contains discovered context" "Auto-discovered Context" "$fdi_content"

  # Test: Constraints mentions milestone deadline
  assert_contains "Constraints mentions milestone deadline" "v1.30" "$fdi_content"
fi

# Test: --dry-run prints to stdout, no file created
find "$FDI_WS/.forgia/fd" -maxdepth 1 -name 'FD-*.md' -delete 2>/dev/null || true
fdi_output=$("$FORGIA" fd-from-issue 42 --repo=test-org/test-repo --dry-run 2>&1)
assert_contains "--dry-run prints to stdout" "Add ingress rate limiting to nginx module" "$fdi_output"
assert_contains "--dry-run shows dry-run message" "dry-run" "$fdi_output"
fdi_count=$(find "$FDI_WS/.forgia/fd" -maxdepth 1 -name 'FD-*.md' 2>/dev/null | wc -l | tr -d ' ')
assert_eq "--dry-run creates no file" "0" "$fdi_count"

# Test: --no-comments skips comments section
find "$FDI_WS/.forgia/fd" -maxdepth 1 -name 'FD-*.md' -delete 2>/dev/null || true
fdi_output=$("$FORGIA" fd-from-issue 42 --repo=test-org/test-repo --no-comments --dry-run 2>&1)
TOTAL=$((TOTAL + 1))
if ! echo "$fdi_output" | grep -q "limit_req module"; then
  printf "  ${GREEN}PASS${NC} %s\n" "--no-comments skips comments section"
  PASSED=$((PASSED + 1))
else
  printf "  ${RED}FAIL${NC} %s\n" "--no-comments skips comments section"
  FAILED=$((FAILED + 1))
fi

# Test: --repo overrides remote inference
fdi_output=$("$FORGIA" fd-from-issue 42 --repo=other-org/other-repo --dry-run 2>&1)
assert_contains "--repo overrides remote inference" "other-org/other-repo" "$fdi_output"

# Test: --repo with space separator works
fdi_output=$("$FORGIA" fd-from-issue 42 --repo test-org/test-repo --dry-run 2>&1)
assert_contains "--repo space form works" "Add ingress rate limiting" "$fdi_output"

# Test: --context adds extra paths to notes
mkdir -p "$FDI_WS/docs/custom"
echo "# Custom notes" > "$FDI_WS/docs/custom/notes.md"
fdi_output=$("$FORGIA" fd-from-issue 42 --repo=test-org/test-repo --context=docs/custom/notes.md --dry-run 2>&1)
assert_contains "--context adds extra paths to notes" "docs/custom/notes.md" "$fdi_output"

# Test: --context with space separator works
fdi_output=$("$FORGIA" fd-from-issue 42 --repo=test-org/test-repo --context docs/custom/notes.md --dry-run 2>&1)
assert_contains "--context space form works" "docs/custom/notes.md" "$fdi_output"

# Test: next FD ID increments correctly
find "$FDI_WS/.forgia/fd" -maxdepth 1 -name 'FD-*.md' -delete 2>/dev/null || true
cat > "$FDI_WS/.forgia/fd/FD-001-first.md" <<'FD'
---
id: FD-001
title: "First"
status: done
---
FD
fdi_output=$("$FORGIA" fd-from-issue 42 --repo=test-org/test-repo --dry-run 2>&1)
assert_contains "next FD ID increments correctly" "FD-002" "$fdi_output"

# Test: next FD ID handles gaps (FD-001, FD-003 → FD-004)
find "$FDI_WS/.forgia/fd" -maxdepth 1 -name 'FD-*.md' -delete 2>/dev/null || true
cat > "$FDI_WS/.forgia/fd/FD-001-first.md" <<'FD'
---
id: FD-001
title: "First"
status: done
---
FD
cat > "$FDI_WS/.forgia/fd/FD-003-third.md" <<'FD'
---
id: FD-003
title: "Third"
status: done
---
FD
fdi_output=$("$FORGIA" fd-from-issue 42 --repo=test-org/test-repo --dry-run 2>&1)
assert_contains "next FD ID handles gaps (FD-001, FD-003 → FD-004)" "FD-004" "$fdi_output"

# Test: infer_repo handles SSH URL
infer_ssh=$(echo "git@github.com:owner-c/repo-d.git" | sed -E 's|.*github\.com[:/]||;s|\.git$||')
assert_eq "infer_repo handles SSH URL" "owner-c/repo-d" "$infer_ssh"

# Test: infer_repo handles HTTPS URL
infer_https=$(echo "https://github.com/owner-a/repo-b.git" | sed -E 's|.*github\.com[:/]||;s|\.git$||')
assert_eq "infer_repo handles HTTPS URL" "owner-a/repo-b" "$infer_https"

# Cleanup FDI resources
fdi_cleanup

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
