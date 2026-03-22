#!/usr/bin/env bash
set -euo pipefail

# Forgia Tests — [knowledge] config section
# Run from repo root: ./tests/e2e-knowledge-config.sh
#
# Covers:
#   Unit 1: Config with [knowledge] section → correct variables
#   Unit 2: Config without [knowledge] section → defaults
#   Unit 3: auto_index = false → KNOWLEDGE_AUTO_INDEX is "false"
#   E2E 1:  forgia init respects auto_index = false (skips indexing)
#   E2E 2:  New project from template has [knowledge] section

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FORGIA="$ROOT_DIR/bin/forgia"
TEST_DIR=$(mktemp -d)
MOCK_DIR=$(mktemp -d)
PASSED=0
FAILED=0
TOTAL=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

cleanup() {
  rm -rf "$TEST_DIR" "$MOCK_DIR"
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
  if echo "$haystack" | grep -qF "$needle"; then
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
  if ! echo "$haystack" | grep -qF "$needle"; then
    printf "  ${GREEN}PASS${NC} %s\n" "$desc"
    PASSED=$((PASSED + 1))
  else
    printf "  ${RED}FAIL${NC} %s\n" "$desc"
    printf "    expected NOT to contain: %s\n" "$needle"
    FAILED=$((FAILED + 1))
  fi
}

# --- Create mock codebase-memory-mcp ---

cat > "$MOCK_DIR/codebase-memory-mcp" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
case "${1:-}" in
  cli)
    case "${2:-}" in
      list_projects)
        echo '[{"path":"./","node_count":42,"edge_count":99}]'
        ;;
      index_repository)
        # Record that indexing was called
        echo "INDEXING" >> "${MOCK_CBM_LOG:-/dev/null}"
        echo '{"symbols_indexed":42}'
        ;;
      *) echo "Unknown: ${2:-}" >&2; exit 1 ;;
    esac
    ;;
  *) echo "mock codebase-memory-mcp" ;;
esac
MOCK
chmod +x "$MOCK_DIR/codebase-memory-mcp"

CLEAN_PATH="/usr/bin:/bin:/usr/sbin:/sbin"

# --- Tests ---

echo "=== Forgia Tests — [knowledge] config ==="
echo "  Test dir: $TEST_DIR"
echo ""

# =====================================================
echo "--- Unit 1: Config with [knowledge] → correct variables ---"
# =====================================================

# Source just the config function from bin/forgia
# We test _cfg_load_knowledge in isolation by sourcing the script
# in a subshell with VAULT_DIR pointing to our test config.

mkdir -p "$TEST_DIR/unit1/.forgia"
cat > "$TEST_DIR/unit1/.forgia/config.toml" <<'TOML'
[runner]
default = "claude"

[knowledge]
provider = "my-custom-provider"
auto_index = false
auto_sync = false

[watcher]
enabled = false
TOML

result=$(
  VAULT_DIR="$TEST_DIR/unit1/.forgia" \
  bash -c '
    source "'"$FORGIA"'"_cfg_load_knowledge_test 2>/dev/null || true
    # Source the file to get the function, but prevent main() from running
    eval "$(sed -n "/^_cfg_load_knowledge/,/^}/p" "'"$FORGIA"'")"
    eval "$(sed -n "/^export KNOWLEDGE_/p" "'"$FORGIA"'")"
    VAULT_DIR="'"$TEST_DIR/unit1/.forgia"'"
    _cfg_load_knowledge
    echo "PROVIDER=$KNOWLEDGE_PROVIDER"
    echo "AUTO_INDEX=$KNOWLEDGE_AUTO_INDEX"
    echo "AUTO_SYNC=$KNOWLEDGE_AUTO_SYNC"
  ' 2>/dev/null
)

assert_contains "provider parsed correctly" \
  "PROVIDER=my-custom-provider" "$result"
assert_contains "auto_index parsed correctly" \
  "AUTO_INDEX=false" "$result"
assert_contains "auto_sync parsed correctly" \
  "AUTO_SYNC=false" "$result"

echo ""

# =====================================================
echo "--- Unit 2: Config without [knowledge] → defaults ---"
# =====================================================

mkdir -p "$TEST_DIR/unit2/.forgia"
cat > "$TEST_DIR/unit2/.forgia/config.toml" <<'TOML'
[runner]
default = "claude"

[watcher]
enabled = false
TOML

result=$(
  bash -c '
    eval "$(sed -n "/^_cfg_load_knowledge/,/^}/p" "'"$FORGIA"'")"
    eval "$(sed -n "/^export KNOWLEDGE_/p" "'"$FORGIA"'")"
    VAULT_DIR="'"$TEST_DIR/unit2/.forgia"'"
    _cfg_load_knowledge
    echo "PROVIDER=$KNOWLEDGE_PROVIDER"
    echo "AUTO_INDEX=$KNOWLEDGE_AUTO_INDEX"
    echo "AUTO_SYNC=$KNOWLEDGE_AUTO_SYNC"
  ' 2>/dev/null
)

assert_contains "default provider" \
  "PROVIDER=codebase-memory-mcp" "$result"
assert_contains "default auto_index" \
  "AUTO_INDEX=true" "$result"
assert_contains "default auto_sync" \
  "AUTO_SYNC=true" "$result"

echo ""

# =====================================================
echo "--- Unit 3: auto_index = false → KNOWLEDGE_AUTO_INDEX is false ---"
# =====================================================

mkdir -p "$TEST_DIR/unit3/.forgia"
cat > "$TEST_DIR/unit3/.forgia/config.toml" <<'TOML'
[runner]
default = "claude"

[knowledge]
auto_index = false
TOML

result=$(
  bash -c '
    eval "$(sed -n "/^_cfg_load_knowledge/,/^}/p" "'"$FORGIA"'")"
    eval "$(sed -n "/^export KNOWLEDGE_/p" "'"$FORGIA"'")"
    VAULT_DIR="'"$TEST_DIR/unit3/.forgia"'"
    _cfg_load_knowledge
    echo "AUTO_INDEX=$KNOWLEDGE_AUTO_INDEX"
  ' 2>/dev/null
)

assert_eq "auto_index is false" \
  "AUTO_INDEX=false" "$result"

echo ""

# =====================================================
echo "--- E2E 1: forgia init respects auto_index = false ---"
# =====================================================

cd "$TEST_DIR"
rm -rf e2e1
mkdir -p e2e1
cd e2e1
git init --initial-branch=main -q

# First init — creates config from template
PATH="$MOCK_DIR:$CLEAN_PATH" "$FORGIA" init >/dev/null 2>&1

# Now set auto_index = false in the config
# (the config was created from template which has auto_index = true)
sed -i.bak 's/auto_index = true/auto_index = false/' .forgia/config.toml
rm -f .forgia/config.toml.bak

# Create a log file to track if indexing is called
MOCK_LOG="$TEST_DIR/e2e1-index.log"
export MOCK_CBM_LOG="$MOCK_LOG"
rm -f "$MOCK_LOG"

# Re-init (overwrite templates)
output=$(PATH="$MOCK_DIR:$CLEAN_PATH" "$FORGIA" init <<< "y" 2>&1)

# Check that indexing was NOT called
if [[ -f "$MOCK_LOG" ]]; then
  assert_not_contains "indexing skipped when auto_index=false" \
    "INDEXING" "$(cat "$MOCK_LOG" 2>/dev/null || echo "")"
else
  # Log file was never created — indexing never called. PASS.
  TOTAL=$((TOTAL + 1))
  printf "  ${GREEN}PASS${NC} %s\n" "indexing skipped when auto_index=false (no index call)"
  PASSED=$((PASSED + 1))
fi

# Also verify the output doesn't mention indexing
assert_not_contains "output does not mention indexing" \
  "indexing with codebase-memory-mcp" "$output"

unset MOCK_CBM_LOG

echo ""

# =====================================================
echo "--- E2E 2: Template has [knowledge] section ---"
# =====================================================

cd "$TEST_DIR"
rm -rf e2e2
mkdir -p e2e2
cd e2e2
git init --initial-branch=main -q

output=$(PATH="$CLEAN_PATH" "$FORGIA" init 2>&1)

config_content=$(cat .forgia/config.toml)
assert_contains "config.toml has [knowledge] section" \
  "[knowledge]" "$config_content"
assert_contains "config.toml has provider field" \
  'provider = "codebase-memory-mcp"' "$config_content"
assert_contains "config.toml has auto_index field" \
  "auto_index = true" "$config_content"
assert_contains "config.toml has auto_sync field" \
  "auto_sync = true" "$config_content"

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
