#!/usr/bin/env bash
set -euo pipefail

# Forgia E2E Tests — Beads + AutoSpec integration
# Run from repo root: ./tests/e2e-beads-autospec.sh

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

echo "=== Forgia E2E Tests — Beads + AutoSpec ==="
echo "  Test dir: $TEST_DIR"
echo ""

cd "$TEST_DIR"
git init --initial-branch=main -q

# =====================================================
echo "--- forgia init: config.toml scaffolding ---"
# =====================================================

"$FORGIA" init >/dev/null 2>&1

assert_file_exists "creates config.toml" "$TEST_DIR/.forgia/config.toml"

echo ""

# =====================================================
echo "--- config.toml: beads section ---"
# =====================================================

config="$TEST_DIR/.forgia/config.toml"
config_content=$(cat "$config")

assert_contains "config has [beads]" '\[beads\]' "$config_content"
assert_contains "config has auto_create_tasks" 'auto_create_tasks' "$config_content"
assert_contains "config has auto_close_tasks" 'auto_close_tasks' "$config_content"
assert_contains "config has dependency_aware_batch" 'dependency_aware_batch' "$config_content"

echo ""

# =====================================================
echo "--- config.toml: runner sections ---"
# =====================================================

assert_contains "config has [runner]" '\[runner\]' "$config_content"
assert_contains "config has [runner.claude]" '\[runner.claude\]' "$config_content"
assert_contains "config has [runner.openhands]" '\[runner.openhands\]' "$config_content"
assert_contains "config has [watcher]" '\[watcher\]' "$config_content"
assert_contains "config has worktree setting" 'worktree' "$config_content"
assert_contains "config has auto_approve setting" 'auto_approve' "$config_content"
assert_contains "config has max_turns setting" 'max_turns' "$config_content"
assert_contains "config has openhands image" 'openhands' "$config_content"
assert_contains "config has openhands model" 'model' "$config_content"
assert_contains "config has watcher debounce" 'debounce' "$config_content"

echo ""

# =====================================================
echo "--- YAML SDD template ---"
# =====================================================

assert_file_exists "yaml sdd template exists" "$ROOT_DIR/modules/vault-template/sdd/_templates/sdd-template.yaml"

yaml_content=$(cat "$ROOT_DIR/modules/vault-template/sdd/_templates/sdd-template.yaml")

assert_contains "yaml has meta section" 'meta:' "$yaml_content"
assert_contains "yaml has meta.id" 'id:' "$yaml_content"
assert_contains "yaml has meta.fd" 'fd:' "$yaml_content"
assert_contains "yaml has meta.status" 'status:' "$yaml_content"
assert_contains "yaml has scope" 'scope:' "$yaml_content"
assert_contains "yaml has interfaces" 'interfaces:' "$yaml_content"
assert_contains "yaml has constraints" 'constraints:' "$yaml_content"
assert_contains "yaml has best_practices" 'best_practices:' "$yaml_content"
assert_contains "yaml has test_requirements" 'test_requirements:' "$yaml_content"
assert_contains "yaml has acceptance_criteria" 'acceptance_criteria:' "$yaml_content"
assert_contains "yaml has context" 'context:' "$yaml_content"
assert_contains "yaml has constitution_check" 'constitution_check:' "$yaml_content"
assert_contains "yaml has work_log" 'work_log:' "$yaml_content"
assert_contains "yaml has bd_task_id" 'bd_task_id:' "$yaml_content"
assert_contains "yaml has principles_loaded" 'principles_loaded:' "$yaml_content"
assert_contains "yaml has lang_loaded" 'lang_loaded:' "$yaml_content"
assert_contains "yaml has retrospective" 'retrospective:' "$yaml_content"

echo ""

# =====================================================
echo "--- validate-sdd.sh: exists and executable ---"
# =====================================================

assert_file_exists "validate-sdd.sh exists" "$ROOT_DIR/modules/runners/validate-sdd.sh"

TOTAL=$((TOTAL + 1))
if [[ -x "$ROOT_DIR/modules/runners/validate-sdd.sh" ]]; then
  printf "  ${GREEN}PASS${NC} validate-sdd.sh is executable\n"
  PASSED=$((PASSED + 1))
else
  printf "  ${RED}FAIL${NC} validate-sdd.sh is executable\n"
  FAILED=$((FAILED + 1))
fi

echo ""

# =====================================================
echo "--- validate-sdd.sh: markdown validation ---"
# =====================================================

# Empty SDD should fail
cat > "$TEST_DIR/empty-sdd.md" <<'SDD'
---
---
# Empty
SDD

assert_exit_code "empty SDD fails validation" 1 "$ROOT_DIR/modules/runners/validate-sdd.sh" "$TEST_DIR/empty-sdd.md"

# Valid SDD should pass
cat > "$TEST_DIR/valid-sdd.md" <<'SDD'
---
id: SDD-001
fd: FD-001
title: "Test SDD"
status: planned
---

## Scope

Build the API endpoint for user auth.

## Interfaces / Interfacce

| Interface | Type | Description |
|-----------|------|-------------|
| POST /auth | API | Login endpoint |

## Constraints / Vincoli

- Language: Rust
- Framework: axum

## Test Requirements

| Type | What | Coverage |
|------|------|----------|
| Unit | Auth logic | 80% |

## Acceptance Criteria / Criteri di Accettazione

- [ ] POST /auth returns JWT
- [ ] Invalid credentials return 401
- [ ] Tests pass

## Context / Contesto

- [ ] `src/auth.rs`

## Constitution Check

- [x] Respects code standards
- [x] No hardcoded secrets

## Work Log / Diario di Lavoro

### Agent
- **Executor**: pending
SDD

assert_exit_code "valid SDD passes validation" 0 "$ROOT_DIR/modules/runners/validate-sdd.sh" "$TEST_DIR/valid-sdd.md"

# SDD missing frontmatter fields
cat > "$TEST_DIR/missing-fields.md" <<'SDD'
---
title: "No ID"
---

## Scope

Something

## Acceptance Criteria

- [ ] criterion
SDD

output=$("$ROOT_DIR/modules/runners/validate-sdd.sh" "$TEST_DIR/missing-fields.md" 2>&1 || true)
assert_contains "reports missing id" "id" "$output"

echo ""

# =====================================================
echo "--- validate-sdd.sh: nonexistent file ---"
# =====================================================

assert_exit_code "nonexistent file exits 1" 1 "$ROOT_DIR/modules/runners/validate-sdd.sh" "$TEST_DIR/nope.md"

echo ""

# =====================================================
echo "--- forgia validate command ---"
# =====================================================

output=$("$FORGIA" 2>&1 || true)
assert_contains "usage shows validate command" "validate" "$output"

echo ""

# =====================================================
echo "--- forgia validate: single file ---"
# =====================================================

assert_exit_code "validate valid SDD passes" 0 "$FORGIA" validate "$TEST_DIR/valid-sdd.md"
assert_exit_code "validate empty SDD fails" 1 "$FORGIA" validate "$TEST_DIR/empty-sdd.md"

echo ""

# =====================================================
echo "--- forgia validate: FD directory ---"
# =====================================================

mkdir -p "$TEST_DIR/.forgia/sdd/FD-TEST"
cp "$TEST_DIR/valid-sdd.md" "$TEST_DIR/.forgia/sdd/FD-TEST/SDD-001-auth.md"
assert_exit_code "validate FD directory passes" 0 "$FORGIA" validate "FD-TEST"

cp "$TEST_DIR/empty-sdd.md" "$TEST_DIR/.forgia/sdd/FD-TEST/SDD-002-broken.md"
assert_exit_code "validate FD directory with bad SDD fails" 1 "$FORGIA" validate "FD-TEST"

echo ""

# =====================================================
echo "--- forgia exec: pre-validation ---"
# =====================================================

# exec should fail on invalid SDD (validation gate)
output=$("$FORGIA" exec "$TEST_DIR/empty-sdd.md" 2>&1 || true)
assert_contains "exec rejects invalid SDD" "validation failed" "$output"

echo ""

# =====================================================
echo "--- forgia batch: error on missing FD dir ---"
# =====================================================

assert_exit_code "batch on nonexistent FD exits 1" 1 "$FORGIA" batch "FD-NOPE"

echo ""

# =====================================================
echo "--- forgia watch: requires fswatch ---"
# =====================================================

if ! command -v fswatch >/dev/null 2>&1; then
  output=$("$FORGIA" watch "FD-001" 2>&1 || true)
  assert_contains "watch warns about fswatch" "fswatch" "$output"
fi

echo ""

# =====================================================
echo "--- mise tasks: bd tasks exist ---"
# =====================================================

mise_content=$(cat "$ROOT_DIR/mise.toml")

assert_contains "mise has bd:install" 'bd:install' "$mise_content"
assert_contains "mise has bd:init" 'bd:init' "$mise_content"
assert_contains "mise has bd:ready" 'bd:ready' "$mise_content"
assert_contains "mise has bd:status" 'bd:status' "$mise_content"
assert_contains "mise has sdd:exec" 'sdd:exec' "$mise_content"
assert_contains "mise has sdd:batch" 'sdd:batch' "$mise_content"
assert_contains "mise has sdd:watch" 'sdd:watch' "$mise_content"
assert_contains "mise has sdd:validate" 'sdd:validate' "$mise_content"
assert_contains "mise has tools:install" 'tools:install' "$mise_content"

echo ""

# =====================================================
echo "--- forgia doctor: checks fswatch and yq ---"
# =====================================================

output=$("$FORGIA" doctor 2>&1)
assert_contains "doctor checks fswatch" "fswatch" "$output"
assert_contains "doctor checks yq" "yq" "$output"
assert_contains "doctor checks beads" "beads" "$output"

echo ""

# =====================================================
echo "--- fd-sdd: beads integration in command ---"
# =====================================================

sdd_cmd=$(cat "$ROOT_DIR/modules/claude-commands/fd-sdd.md")
assert_contains "fd-sdd mentions Beads" "Beads" "$sdd_cmd"
assert_contains "fd-sdd mentions bd" "bd" "$sdd_cmd"
assert_contains "fd-sdd mentions epic" "epic" "$sdd_cmd"

echo ""

# =====================================================
echo "--- fd template: mandatory Mermaid ---"
# =====================================================

fd_tmpl=$(cat "$ROOT_DIR/modules/vault-template/fd/_templates/fd-template.md")
assert_contains "fd template has Integration Context" "Integration Context" "$fd_tmpl"
assert_contains "fd template has Data Flow" "Data Flow" "$fd_tmpl"
assert_contains "fd template has sequenceDiagram" "sequenceDiagram" "$fd_tmpl"
assert_contains "fd template has flowchart" "flowchart" "$fd_tmpl"
assert_contains "fd template says MANDATORY" "OBBLIGATORIO" "$fd_tmpl"

echo ""

# =====================================================
echo "--- fd-review: enforces Mermaid ---"
# =====================================================

review_cmd=$(cat "$ROOT_DIR/modules/claude-commands/fd-review.md")
assert_contains "fd-review checks Integration Context" "Integration Context" "$review_cmd"
assert_contains "fd-review checks Data Flow" "Data Flow" "$review_cmd"
assert_contains "fd-review rejects placeholders" "NOT the default template" "$review_cmd"
assert_contains "fd-review checks Mermaid syntax" "Mermaid syntax" "$review_cmd"

echo ""

# =====================================================
echo "--- runners: claude runner structure ---"
# =====================================================

claude_runner=$(cat "$ROOT_DIR/modules/runners/claude.sh")
assert_contains "claude runner has set -euo pipefail" "set -euo pipefail" "$claude_runner"
assert_contains "claude runner loads constitution" "constitution" "$claude_runner"
assert_contains "claude runner loads guardrails" "guardrails" "$claude_runner"

echo ""

# =====================================================
echo "--- README: prerequisites table ---"
# =====================================================

readme=$(cat "$ROOT_DIR/README.md")
assert_contains "readme has Prerequisites section" "Prerequisites" "$readme"
assert_contains "readme lists mise" "mise" "$readme"
assert_contains "readme lists docker" "docker" "$readme"
assert_contains "readme lists bd" "bd" "$readme"
assert_contains "readme lists fswatch" "fswatch" "$readme"
assert_contains "readme lists yq" "yq" "$readme"

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
