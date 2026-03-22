#!/usr/bin/env bash
set -euo pipefail

# Forgia E2E Tests — codebase-memory-mcp integration
# Run from repo root: ./tests/e2e-codebase-memory.sh

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

# --- Create mock codebase-memory-mcp ---

cat > "$MOCK_DIR/codebase-memory-mcp" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail

case "${1:-}" in
  cli)
    case "${2:-}" in
      list_projects)
        if [[ "${MOCK_CBM_NO_INDEX:-}" == "true" ]]; then
          echo '[]'
        else
          cat <<'JSON'
[{"path":"./","node_count":1234,"edge_count":5678}]
JSON
        fi
        ;;
      index_repository)
        echo '{"symbols_indexed":1234}'
        ;;
      *)
        echo "Unknown CLI command: ${2:-}" >&2
        exit 1
        ;;
    esac
    ;;
  *)
    echo "codebase-memory-mcp mock v0.1.0"
    ;;
esac
MOCK
chmod +x "$MOCK_DIR/codebase-memory-mcp"

# Minimal PATH that excludes codebase-memory-mcp
# but includes all standard system tools
CLEAN_PATH="/usr/bin:/bin:/usr/sbin:/sbin"

# --- Tests ---

echo "=== Forgia E2E Tests — codebase-memory-mcp ==="
echo "  Test dir: $TEST_DIR"
echo "  Mock dir: $MOCK_DIR"
echo ""

# =====================================================
echo "--- 1. forgia init: cbm not installed → skip ---"
# =====================================================

cd "$TEST_DIR"
git init --initial-branch=main -q

output=$(PATH="$CLEAN_PATH" "$FORGIA" init 2>&1)
assert_contains "prints not found message" \
  "codebase-memory-mcp not found" "$output"
assert_contains "mentions optional" \
  "optional" "$output"

echo ""

# =====================================================
echo "--- 2. forgia init: mock cbm → index + stats ---"
# =====================================================

rm -rf "$TEST_DIR/.forgia" "$TEST_DIR/.mcp.json"
output=$(PATH="$MOCK_DIR:$CLEAN_PATH" "$FORGIA" init 2>&1)
assert_contains "prints indexed message" \
  "codebase-memory-mcp indexed" "$output"
assert_contains "prints symbol count" \
  "1234 symbols" "$output"

echo ""

# =====================================================
echo "--- 3. forgia init: generates .mcp.json ---"
# =====================================================

assert_file_exists ".mcp.json created" \
  "$TEST_DIR/.mcp.json"
mcp_content=$(cat "$TEST_DIR/.mcp.json")
assert_contains ".mcp.json has codebase-memory-mcp" \
  "codebase-memory-mcp" "$mcp_content"
assert_contains ".mcp.json has stdio type" \
  "stdio" "$mcp_content"

echo ""

# =====================================================
echo "--- 4. forgia init: merge existing .mcp.json ---"
# =====================================================

rm -rf "$TEST_DIR/.forgia"

# Create an existing .mcp.json with another server
cat > "$TEST_DIR/.mcp.json" <<'EXISTING'
{
  "mcpServers": {
    "other-server": {
      "type": "stdio",
      "command": "other-tool"
    }
  }
}
EXISTING

output=$(PATH="$MOCK_DIR:$CLEAN_PATH" "$FORGIA" init 2>&1)
mcp_content=$(cat "$TEST_DIR/.mcp.json")
assert_contains "preserves other-server" \
  "other-server" "$mcp_content"
assert_contains "adds codebase-memory-mcp" \
  "codebase-memory-mcp" "$mcp_content"

echo ""

# =====================================================
echo "--- 5. forgia doctor: cbm not installed → WARN ---"
# =====================================================

cd "$TEST_DIR"
output=$(PATH="$CLEAN_PATH" "$FORGIA" doctor 2>&1)
assert_contains "doctor shows WARN for cbm" \
  "WARN" "$output"
assert_contains "doctor mentions not installed" \
  "not installed (optional)" "$output"

echo ""

# =====================================================
echo "--- 6. forgia doctor: mock cbm → OK + stats ---"
# =====================================================

output=$(PATH="$MOCK_DIR:$CLEAN_PATH" "$FORGIA" doctor 2>&1)
assert_contains "doctor shows OK for cbm" \
  "OK" "$output"
assert_contains "doctor shows symbol count" \
  "1234 symbols" "$output"
assert_contains "doctor shows edge count" \
  "5678 edges" "$output"

echo ""

# =====================================================
echo "--- 7. forgia status: cbm installed → Knowledge Layer ---"
# =====================================================

output=$(PATH="$MOCK_DIR:$CLEAN_PATH" "$FORGIA" status 2>&1)
assert_contains "status shows Knowledge Layer" \
  "Knowledge Layer" "$output"
assert_contains "status shows Tier 3" \
  "Tier 3" "$output"
assert_contains "status shows symbol count" \
  "1234 symbols" "$output"

echo ""

# =====================================================
echo "--- 8. forgia status: cbm not installed → omit ---"
# =====================================================

output=$(PATH="$CLEAN_PATH" "$FORGIA" status 2>&1)
assert_not_contains "status omits Knowledge Layer" \
  "Knowledge Layer" "$output"

echo ""

# =====================================================
echo "--- 9. regression: existing e2e.sh tests ---"
# =====================================================

TOTAL=$((TOTAL + 1))
if "$ROOT_DIR/tests/e2e.sh" >/dev/null 2>&1; then
  printf "  ${GREEN}PASS${NC} %s\n" \
    "existing e2e.sh tests pass"
  PASSED=$((PASSED + 1))
else
  printf "  ${RED}FAIL${NC} %s\n" \
    "existing e2e.sh tests regression"
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
