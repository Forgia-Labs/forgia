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
  auto_approve=$(grep -A5 '\[runner\.claude\]' "$CONFIG_FILE" | grep 'auto_approve' | sed 's/.*= *//' | tr -d ' ' || echo "true")
  max_turns=$(grep -A5 '\[runner\.claude\]' "$CONFIG_FILE" | grep 'max_turns' | sed 's/.*= *//' | tr -d ' ' || echo "200")
  use_worktree=$(grep -A5 '\[runner\.claude\]' "$CONFIG_FILE" | grep 'worktree' | sed 's/.*= *//' | tr -d ' ' || echo "true")
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

prompt=$(cat <<PROMPT
You are executing an SDD (Spec-Driven Development) task. Follow the spec exactly.

## Constitution
$constitution

## Coding Principles & Conventions
$context

## SDD to Execute
$sdd_content

## Instructions
1. Read and understand the full SDD above
2. Implement everything in the Scope section
3. Follow all Constraints and Best Practices
4. Write tests as specified in Test Requirements
5. Verify all Acceptance Criteria are met
6. When done, update the Work Log section of the SDD file ($SDD_FILE) with:
   - Agent: claude-code
   - Date: $(date +%Y-%m-%d)
   - Decisions: key implementation decisions you made
   - Output: files created/modified
   - Retrospective: what went well, what was tricky
7. Commit your changes with a descriptive message referencing $sdd_id
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

echo "→ Launching Claude Code agent..."
echo ""

# Execute via claude CLI
if command -v claude >/dev/null 2>&1; then
  echo "$prompt" | claude "${claude_args[@]}" --print
else
  echo "Error: 'claude' CLI not found. Install Claude Code first." >&2
  exit 1
fi

echo ""
echo "=== Claude Runner Complete ==="
echo "  SDD: $sdd_id"
echo "  Check $SDD_FILE for the Work Log"
