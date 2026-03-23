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

assert_dir_exists "creates architecture/" "$TEST_DIR/.forgia/architecture"
assert_file_exists "creates system-context.yaml" "$TEST_DIR/.forgia/architecture/system-context.yaml"
assert_file_exists "creates containers.yaml" "$TEST_DIR/.forgia/architecture/containers.yaml"
assert_file_exists "creates technology-decisions.yaml" "$TEST_DIR/.forgia/architecture/technology-decisions.yaml"
assert_file_exists "creates quality-attributes.yaml" "$TEST_DIR/.forgia/architecture/quality-attributes.yaml"
assert_file_exists "creates constraints.yaml" "$TEST_DIR/.forgia/architecture/constraints.yaml"
assert_file_exists "creates glossary.yaml" "$TEST_DIR/.forgia/architecture/glossary.yaml"
assert_dir_exists "creates contexts/" "$TEST_DIR/.forgia/contexts"
assert_file_exists "creates contexts/_template.yaml" "$TEST_DIR/.forgia/contexts/_template.yaml"
assert_dir_exists "creates learnings/" "$TEST_DIR/.forgia/learnings"
assert_file_exists "creates learnings/_template.yaml" "$TEST_DIR/.forgia/learnings/_template.yaml"

echo ""

# =====================================================
echo "--- forgia init (architecture idempotency) ---"
# =====================================================

# Modify a file inside architecture/ — re-init should NOT overwrite
echo "# Custom content" > "$TEST_DIR/.forgia/architecture/system-context.yaml"
"$FORGIA" init <<< "y" >/dev/null 2>&1
custom_content=$(cat "$TEST_DIR/.forgia/architecture/system-context.yaml")
assert_eq "architecture/ not overwritten on re-init" "# Custom content" "$custom_content"

echo ""

# =====================================================
echo "--- YAML template validation ---"
# =====================================================

# Detect a working YAML parser
yaml_parser=""
if command -v python3 >/dev/null 2>&1 && python3 -c "import yaml" 2>/dev/null; then
  yaml_parser="python3"
elif command -v ruby >/dev/null 2>&1; then
  yaml_parser="ruby"
fi

yaml_ok=true
if [[ -n "$yaml_parser" ]]; then
  for f in "$ROOT_DIR"/modules/vault-template/architecture/*.yaml \
           "$ROOT_DIR"/modules/vault-template/contexts/_template.yaml \
           "$ROOT_DIR"/modules/vault-template/learnings/_template.yaml; do
    local_ok=true
    if [[ "$yaml_parser" == "python3" ]]; then
      python3 -c "import yaml; yaml.safe_load(open('$f'))" 2>/dev/null || local_ok=false
    else
      ruby -ryaml -e "YAML.safe_load(File.read('$f'))" 2>/dev/null || local_ok=false
    fi
    if [[ "$local_ok" == "false" ]]; then
      yaml_ok=false
      printf "  %sFAIL%s YAML parse: %s\n" "$RED" "$NC" "$(basename "$f")"
      TOTAL=$((TOTAL + 1))
      FAILED=$((FAILED + 1))
    fi
  done
fi
if [[ "$yaml_ok" == "true" ]]; then
  TOTAL=$((TOTAL + 1))
  PASSED=$((PASSED + 1))
  printf "  %sPASS%s all YAML templates are parseable\n" "$GREEN" "$NC"
fi

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
