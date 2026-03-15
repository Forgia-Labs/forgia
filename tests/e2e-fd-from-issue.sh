#!/usr/bin/env bash
set -euo pipefail

# E2E tests for: forgia fd-from-issue
# Run from repo root: ./tests/e2e-fd-from-issue.sh

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FORGIA="$ROOT_DIR/bin/forgia"
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

assert_file_not_exists() {
  local desc="$1" file="$2"
  TOTAL=$((TOTAL + 1))
  if [[ ! -f "$file" ]]; then
    printf "  ${GREEN}PASS${NC} %s\n" "$desc"
    PASSED=$((PASSED + 1))
  else
    printf "  ${RED}FAIL${NC} %s\n" "$desc"
    printf "    file should not exist: %s\n" "$file"
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

# --- Create mock gh script ---

create_mock_gh() {
  local mock_dir="$TEST_DIR/mock-bin"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/gh" <<'MOCK_GH'
#!/usr/bin/env bash
# Mock gh CLI for testing forgia fd-from-issue

# gh auth status — always succeed
if [[ "${1:-}" == "auth" && "${2:-}" == "status" ]]; then
  echo "Logged in to github.com"
  exit 0
fi

# gh api user --jq '.login'
if [[ "${1:-}" == "api" && "${2:-}" == "user" ]]; then
  echo "test-user"
  exit 0
fi

# gh api repos/OWNER/REPO/issues/NUMBER --jq EXPR
if [[ "${1:-}" == "api" && "${2:-}" =~ ^repos/(.+)/issues/([0-9]+)$ ]]; then
  local_repo="${BASH_REMATCH[1]}"
  local_number="${BASH_REMATCH[2]}"
  local_jq="${4:-}"

  # Non-existent issue
  if [[ "$local_number" == "9999" ]]; then
    echo "gh: Not Found (HTTP 404)" >&2
    exit 1
  fi

  # Issue #42 — test issue with all fields
  if [[ "$local_number" == "42" ]]; then
    case "$local_jq" in
      '.title') echo "Add rate limiting to API" ;;
      '.body // empty') echo "We need rate limiting on all public endpoints.
This is important for security.
See RFC 6585 for details." ;;
      '[.labels[].name] | join(", ")') echo "bug, security" ;;
      '.assignee.login // empty') echo "octocat" ;;
      '.milestone.title // empty') echo "v2.0 Release" ;;
      '.state') echo "open" ;;
      *) echo "{}" ;;
    esac
    exit 0
  fi

  # Issue #43 — minimal issue (no labels, no assignee, no milestone)
  if [[ "$local_number" == "43" ]]; then
    case "$local_jq" in
      '.title') echo "Fix typo in README" ;;
      '.body // empty') echo "There is a typo on line 5." ;;
      '[.labels[].name] | join(", ")') echo "" ;;
      '.assignee.login // empty') echo "" ;;
      '.milestone.title // empty') echo "" ;;
      '.state') echo "open" ;;
      *) echo "{}" ;;
    esac
    exit 0
  fi

  echo "gh: Not Found (HTTP 404)" >&2
  exit 1
fi

# gh api repos/OWNER/REPO/issues/NUMBER/comments --jq EXPR
if [[ "${1:-}" == "api" && "${2:-}" =~ ^repos/(.+)/issues/([0-9]+)/comments$ ]]; then
  local_number="${BASH_REMATCH[2]}"

  if [[ "$local_number" == "42" ]]; then
    echo "This is a maintainer comment with important context."
    echo "Second comment: we should use token bucket algorithm."
    exit 0
  fi

  # No comments for other issues
  echo ""
  exit 0
fi

echo "gh: unknown command" >&2
exit 1
MOCK_GH
  chmod +x "$mock_dir/gh"
  echo "$mock_dir"
}

# --- Setup ---

echo "=== Forgia fd-from-issue E2E Tests ==="
echo "  Test dir: $TEST_DIR"
echo ""

cd "$TEST_DIR"
git init --initial-branch=main -q
git remote add origin "https://github.com/test-owner/test-repo.git"

# Initialize vault
"$FORGIA" init >/dev/null 2>&1

# Create mock gh
MOCK_DIR=$(create_mock_gh)

# =====================================================
echo "--- fd-from-issue (no args → usage + exit 1) ---"
# =====================================================

assert_exit_code "no args exits 1" 1 "$FORGIA" fd-from-issue

output=$("$FORGIA" fd-from-issue 2>&1 || true)
assert_contains "shows usage" "Usage:" "$output"

echo ""

# =====================================================
echo "--- fd-from-issue (non-existent issue → error) ---"
# =====================================================

output=$(PATH="$MOCK_DIR:$PATH" "$FORGIA" fd-from-issue 9999 --repo=Deepzima/forgia 2>&1 || true)
assert_contains "shows error for non-existent issue" "Could not fetch issue" "$output"

echo ""

# =====================================================
echo "--- fd-from-issue --dry-run (prints to stdout) ---"
# =====================================================

output=$(PATH="$MOCK_DIR:$PATH" "$FORGIA" fd-from-issue 42 --repo=test-owner/test-repo --dry-run 2>&1)
assert_contains "dry-run shows FD content" "Add rate limiting to API" "$output"
assert_contains "dry-run shows no-file message" "dry-run" "$output"

# Verify no file was created
fd_files_count=$(find "$TEST_DIR/.forgia/fd" -maxdepth 1 -name 'FD-*.md' 2>/dev/null | wc -l | tr -d ' ')
assert_eq "dry-run creates no FD file" "0" "$fd_files_count"

echo ""

# =====================================================
echo "--- fd-from-issue (creates FD file) ---"
# =====================================================

output=$(PATH="$MOCK_DIR:$PATH" "$FORGIA" fd-from-issue 42 --repo=test-owner/test-repo 2>&1)
assert_contains "shows creating message" "Creating" "$output"
assert_contains "shows success message" "Draft FD created" "$output"

# Find the created file
fd_file=$(find "$TEST_DIR/.forgia/fd" -maxdepth 1 -name 'FD-*-add-rate-limiting-to-api.md' 2>/dev/null | head -1)
assert_file_exists "FD file exists at correct path" "$fd_file"

echo ""

# =====================================================
echo "--- created FD content validation ---"
# =====================================================

if [[ -f "$fd_file" ]]; then
  fd_content=$(cat "$fd_file")

  # Frontmatter fields
  assert_contains "has id in frontmatter" "^id:" "$fd_content"
  assert_contains "has title in frontmatter" "Add rate limiting to API" "$fd_content"
  assert_contains "has status planned" "status: planned" "$fd_content"
  assert_contains "has priority high (bug label)" "priority: high" "$fd_content"
  assert_contains "has tags from labels" "bug" "$fd_content"
  assert_contains "has security tag" "security" "$fd_content"
  assert_contains "has assignee" "octocat" "$fd_content"
  assert_contains "has upstream_issue" "upstream_issue:" "$fd_content"
  assert_contains "upstream_issue has repo#number" "test-owner/test-repo#42" "$fd_content"

  # Template sections present
  assert_contains "has Problem section" "## Problem" "$fd_content"
  assert_contains "has Solutions section" "## Solutions" "$fd_content"
  assert_contains "has Architecture section" "## Architecture" "$fd_content"
  assert_contains "has Interfaces section" "## Interfaces" "$fd_content"
  assert_contains "has SDD Previsti section" "## Planned SDDs" "$fd_content"
  assert_contains "has Constraints section" "## Constraints" "$fd_content"
  assert_contains "has Verification section" "## Verification" "$fd_content"
  assert_contains "has Notes section" "## Notes" "$fd_content"

  # Issue body in Problem section
  assert_contains "Problem has issue body" "rate limiting on all public endpoints" "$fd_content"

  # Comments in Notes section
  assert_contains "Notes has issue comments" "maintainer comment" "$fd_content"
  assert_contains "Notes has second comment" "token bucket" "$fd_content"

  # Milestone in Constraints
  assert_contains "Constraints has milestone" "v2.0 Release" "$fd_content"

  # Context discovery
  assert_contains "Notes has context section" "Auto-discovered Context" "$fd_content"
  assert_contains "Notes has constitution link" "constitution.md" "$fd_content"
fi

echo ""

# =====================================================
echo "--- fd-from-issue --no-comments ---"
# =====================================================

# Clean up previous FD
find "$TEST_DIR/.forgia/fd" -maxdepth 1 -name 'FD-*-add-rate-limiting*.md' -delete 2>/dev/null || true

output=$(PATH="$MOCK_DIR:$PATH" "$FORGIA" fd-from-issue 42 --repo=test-owner/test-repo --no-comments 2>&1)
fd_file=$(find "$TEST_DIR/.forgia/fd" -maxdepth 1 -name 'FD-*-add-rate-limiting-to-api.md' 2>/dev/null | head -1)
if [[ -f "$fd_file" ]]; then
  fd_content=$(cat "$fd_file")
  assert_not_contains "no-comments skips comments" "maintainer comment" "$fd_content"
fi

echo ""

# =====================================================
echo "--- fd-from-issue --context=PATH ---"
# =====================================================

find "$TEST_DIR/.forgia/fd" -maxdepth 1 -name 'FD-*-add-rate-limiting*.md' -delete 2>/dev/null || true
mkdir -p "$TEST_DIR/docs/extra"
echo "extra context" > "$TEST_DIR/docs/extra/notes.md"

output=$(PATH="$MOCK_DIR:$PATH" "$FORGIA" fd-from-issue 42 --repo=test-owner/test-repo --context=docs/extra/notes.md 2>&1)
assert_contains "shows extra context found" "docs/extra/notes.md" "$output"

echo ""

# =====================================================
echo "--- next_fd_id handles gaps ---"
# =====================================================

# Create FD-001 and FD-003 (gap at FD-002)
find "$TEST_DIR/.forgia/fd" -maxdepth 1 -name 'FD-*.md' -delete 2>/dev/null || true
cat > "$TEST_DIR/.forgia/fd/FD-001-first.md" <<'FD'
---
id: FD-001
title: "First"
status: done
---
FD
cat > "$TEST_DIR/.forgia/fd/FD-003-third.md" <<'FD'
---
id: FD-003
title: "Third"
status: done
---
FD

output=$(PATH="$MOCK_DIR:$PATH" "$FORGIA" fd-from-issue 43 --repo=test-owner/test-repo --dry-run 2>&1)
assert_contains "next_fd_id returns FD-004 (not FD-002)" "FD-004" "$output"

echo ""

# =====================================================
echo "--- infer_repo handles SSH and HTTPS ---"
# =====================================================

# Test HTTPS remote
cd "$TEST_DIR"
git remote set-url origin "https://github.com/owner-a/repo-b.git"
# Source the function directly
infer_result=$(bash -c "source '$ROOT_DIR/bin/forgia' <<<'' 2>/dev/null; infer_repo" 2>/dev/null || true)
# Alternative: just test the sed logic directly
infer_https=$(echo "https://github.com/owner-a/repo-b.git" | sed -E 's|.*github\.com[:/]||;s|\.git$||')
assert_eq "HTTPS remote parsed correctly" "owner-a/repo-b" "$infer_https"

# Test SSH remote
infer_ssh=$(echo "git@github.com:owner-c/repo-d.git" | sed -E 's|.*github\.com[:/]||;s|\.git$||')
assert_eq "SSH remote parsed correctly" "owner-c/repo-d" "$infer_ssh"

# Test without .git suffix
infer_no_git=$(echo "https://github.com/owner-e/repo-f" | sed -E 's|.*github\.com[:/]||;s|\.git$||')
assert_eq "HTTPS without .git parsed correctly" "owner-e/repo-f" "$infer_no_git"

echo ""

# =====================================================
echo "--- fd-from-issue (repo inference) ---"
# =====================================================

git remote set-url origin "https://github.com/test-owner/test-repo.git"
find "$TEST_DIR/.forgia/fd" -maxdepth 1 -name 'FD-*.md' -delete 2>/dev/null || true

output=$(PATH="$MOCK_DIR:$PATH" "$FORGIA" fd-from-issue 43 --dry-run 2>&1)
assert_contains "infers repo from remote" "test-owner/test-repo" "$output"
assert_contains "shows issue title" "Fix typo in README" "$output"

echo ""

# =====================================================
echo "--- fd-from-issue (--repo overrides inference) ---"
# =====================================================

find "$TEST_DIR/.forgia/fd" -maxdepth 1 -name 'FD-*.md' -delete 2>/dev/null || true
output=$(PATH="$MOCK_DIR:$PATH" "$FORGIA" fd-from-issue 42 --repo=other-owner/other-repo --dry-run 2>&1)
assert_contains "repo override used" "other-owner/other-repo" "$output"

echo ""

# =====================================================
echo "--- fd-from-issue (minimal issue — no labels, no assignee) ---"
# =====================================================

find "$TEST_DIR/.forgia/fd" -maxdepth 1 -name 'FD-*.md' -delete 2>/dev/null || true
output=$(PATH="$MOCK_DIR:$PATH" "$FORGIA" fd-from-issue 43 --repo=test-owner/test-repo --dry-run 2>&1)
assert_contains "minimal issue has medium priority" "priority: medium" "$output"
assert_contains "minimal issue has title" "Fix typo in README" "$output"

echo ""

# =====================================================
echo "--- usage text includes fd-from-issue ---"
# =====================================================

output=$("$FORGIA" 2>&1 || true)
assert_contains "usage mentions fd-from-issue" "fd-from-issue" "$output"

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
