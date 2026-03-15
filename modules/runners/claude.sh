#!/usr/bin/env bash
set -euo pipefail

# Claude Code Runner
# Executes an SDD using Claude Code sub-agent in an isolated git worktree.
# Uses Max subscription — no API key needed.

VAULT_DIR=".forgia"
SDD_FILE="${1:?Usage: claude.sh <sdd-file>}"
CONFIG_FILE="$VAULT_DIR/config.toml"

if [[ ! -f "$SDD_FILE" ]]; then
  echo "Error: SDD file not found: $SDD_FILE" >&2
  exit 1
fi

# Parse config (simple TOML parsing)
auto_approve="true"
max_turns="200"
use_worktree="true"

if [[ -f "$CONFIG_FILE" ]]; then
  auto_approve=$(grep -A5 '\[runner\.claude\]' "$CONFIG_FILE" | grep -v '^#' | grep 'auto_approve' | head -1 | sed 's/.*= *//' | tr -d ' ' || echo "true")
  max_turns=$(grep -A5 '\[runner\.claude\]' "$CONFIG_FILE" | grep -v '^#' | grep 'max_turns' | head -1 | sed 's/.*= *//' | tr -d ' ' || echo "200")
  use_worktree=$(grep -A5 '\[runner\.claude\]' "$CONFIG_FILE" | grep -v '^#' | grep 'worktree' | head -1 | sed 's/.*= *//' | tr -d ' ' || echo "true")
fi

# Extract SDD metadata
sdd_id=$(grep '^id:' "$SDD_FILE" | head -1 | sed 's/id: *//')
sdd_title=$(grep '^title:' "$SDD_FILE" | head -1 | sed 's/title: *"*//;s/"*$//')

echo "=== Forgia Claude Runner ==="
echo "  SDD:        $sdd_id — $sdd_title"
echo "  Worktree:   $use_worktree"
echo "  Auto-approve: $auto_approve"
echo "  Max turns:  $max_turns"
echo ""

# Build the prompt from the SDD
sdd_content=$(cat "$SDD_FILE")

# Load principles and lang conventions as context
context=""
if [[ -d "$VAULT_DIR/dev-guide/principles" ]]; then
  for f in "$VAULT_DIR/dev-guide/principles/"*.md; do
    [[ -f "$f" ]] && context+="$(cat "$f")"$'\n\n'
  done
fi
if [[ -d "$VAULT_DIR/dev-guide/lang" ]]; then
  for f in "$VAULT_DIR/dev-guide/lang/"*.md; do
    [[ -f "$f" ]] && context+="$(cat "$f")"$'\n\n'
  done
fi

# Load constitution
constitution=""
if [[ -f "$VAULT_DIR/constitution.md" ]]; then
  constitution=$(cat "$VAULT_DIR/constitution.md")
fi

# Load guardrails
guardrails=""
if [[ -f "$VAULT_DIR/guardrails/deny.toml" ]]; then
  guardrails=$(cat "$VAULT_DIR/guardrails/deny.toml")
fi

prompt=$(cat <<PROMPT
You are executing an SDD (Spec-Driven Development) task. Follow the spec exactly.

## Constitution
$constitution

## Security Guardrails — MANDATORY
The following deny list is ABSOLUTE. Violating any rule is a critical failure.

$guardrails

Summary of guardrails:
- NEVER read files matching [read] patterns (secrets, keys, credentials)
- NEVER execute commands matching [execute] patterns (key export, env enumeration)
- NEVER write to files matching [write] patterns (secrets, forgia meta)
- If you need a secret value, use a placeholder and instruct the user to set it
- If you encounter a file that might contain secrets, SKIP it and warn the user

## Coding Principles & Conventions
$context

## SDD to Execute
$sdd_content

## Instructions
1. Read and understand the full SDD above
2. Review the Security Guardrails — never violate them
3. Implement everything in the Scope section
4. Follow all Constraints and Best Practices
5. Write tests as specified in Test Requirements
6. Verify all Acceptance Criteria are met
7. When done, update the Work Log section of the SDD file ($SDD_FILE) with:
   - Agent: claude-code
   - Date: $(date +%Y-%m-%d)
   - Decisions: key implementation decisions you made
   - Output: files created/modified
   - Retrospective: what went well, what was tricky
8. Commit your changes with a descriptive message referencing $sdd_id
PROMPT
)

# Build claude command
claude_args=()

if [[ "$auto_approve" == "true" ]]; then
  claude_args+=(--dangerously-skip-permissions)
fi

if [[ "$max_turns" != "0" ]]; then
  claude_args+=(--max-turns "$max_turns")
fi

echo "→ Launching Claude Code agent (session-isolated)..."
echo ""

# Execute via claude CLI — each SDD gets its own isolated session
# This prevents context pollution between SDDs and saves tokens
if ! command -v claude >/dev/null 2>&1; then
  echo "Error: 'claude' CLI not found. Install Claude Code first." >&2
  exit 1
fi

# Session isolation: --no-conversation-history ensures a fresh context
# Each SDD runs in its own session with only its spec + conventions loaded
# Check if worktree is possible (needs at least one commit)
can_worktree="false"
if [[ "$use_worktree" == "true" ]] && git rev-parse HEAD >/dev/null 2>&1; then
  can_worktree="true"
fi

if [[ "$can_worktree" == "true" ]]; then
  # Create temporary worktree for full git isolation
  worktree_dir=$(mktemp -d)
  branch_name="forgia/${sdd_id:-sdd}-$(date +%s)"

  echo "  Worktree: $worktree_dir"
  echo "  Branch:   $branch_name"

  git worktree add -b "$branch_name" "$worktree_dir" HEAD 2>/dev/null

  # Run in worktree
  (cd "$worktree_dir" && echo "$prompt" | claude "${claude_args[@]}" --print)

  # Merge back if there are changes
  if (cd "$worktree_dir" && git diff --quiet HEAD 2>/dev/null); then
    echo "  No changes — cleaning up worktree"
    git worktree remove "$worktree_dir" 2>/dev/null || rm -rf "$worktree_dir"
    git branch -D "$branch_name" 2>/dev/null || true
  else
    echo "  Changes detected on branch: $branch_name"
    echo "  Worktree: $worktree_dir"
    echo "  Merge with: git merge $branch_name"
  fi
else
  if [[ "$use_worktree" == "true" ]]; then
    echo "  Warning: worktree requested but no commits exist — running directly"
  fi
  # Run directly (no worktree isolation)
  echo "$prompt" | claude "${claude_args[@]}" --print
fi

echo ""
echo "=== Claude Runner Complete ==="
echo "  SDD: $sdd_id"
echo "  Check $SDD_FILE for the Work Log"
