#!/usr/bin/env bash
set -euo pipefail

# Claude Code Runner
# Executes an SDD using Claude Code agent.
# Uses Max subscription — no API key needed.

VAULT_DIR=".forgia"
SDD_FILE="${1:?Usage: claude.sh <sdd-file>}"
CONFIG_FILE="$VAULT_DIR/config.toml"

if [[ ! -f "$SDD_FILE" ]]; then
  echo "Error: SDD file not found: $SDD_FILE" >&2
  exit 1
fi

# Parse config (simple TOML parsing — skip comments)
auto_approve="true"
max_turns="200"

if [[ -f "$CONFIG_FILE" ]]; then
  auto_approve=$(grep -A5 '\[runner\.claude\]' "$CONFIG_FILE" | grep -v '^#' | grep 'auto_approve' | head -1 | sed 's/.*= *//' | tr -d ' ' || echo "true")
  max_turns=$(grep -A5 '\[runner\.claude\]' "$CONFIG_FILE" | grep -v '^#' | grep 'max_turns' | head -1 | sed 's/.*= *//' | tr -d ' ' || echo "200")
fi

# Extract SDD metadata
sdd_id=$(grep '^id:' "$SDD_FILE" | head -1 | sed 's/id: *//;s/"//g')
sdd_title=$(grep '^title:' "$SDD_FILE" | head -1 | sed 's/title: *"*//;s/"*$//')

# Setup log directory
mkdir -p "$VAULT_DIR/logs"
log_file="$VAULT_DIR/logs/exec-${sdd_id}-$(date +%Y-%m-%dT%H:%M:%S).log"
start_time=$SECONDS

echo "=== Forgia Claude Runner ==="
echo "  SDD:          $sdd_id — $sdd_title"
echo "  Auto-approve: $auto_approve"
echo "  Max turns:    $max_turns"
echo "  Log:          $log_file"
echo ""

# Log header
{
  echo "=== Forgia Exec Log ==="
  echo "SDD: $sdd_id"
  echo "File: $SDD_FILE"
  echo "Started: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "Runner: claude-code"
  echo "---"
} > "$log_file"

# Execute via claude CLI
if ! command -v claude >/dev/null 2>&1; then
  echo "Error: 'claude' CLI not found. Install Claude Code first." >&2
  exit 1
fi

# Build system context (constitution + guardrails + principles + lang conventions)
system_context=""

if [[ -f "$VAULT_DIR/constitution.md" ]]; then
  system_context+="## Constitution"$'\n'
  system_context+="$(cat "$VAULT_DIR/constitution.md")"$'\n\n'
fi

if [[ -f "$VAULT_DIR/guardrails/deny.toml" ]]; then
  system_context+="## Security Guardrails — MANDATORY"$'\n'
  system_context+="$(cat "$VAULT_DIR/guardrails/deny.toml")"$'\n'
  system_context+="NEVER read secrets, NEVER execute key export commands, NEVER write to .env or forgia meta files."$'\n\n'
fi

if [[ -d "$VAULT_DIR/dev-guide/principles" ]]; then
  for f in "$VAULT_DIR/dev-guide/principles/"*.md; do
    [[ -f "$f" ]] && system_context+="$(cat "$f")"$'\n\n'
  done
fi
if [[ -d "$VAULT_DIR/dev-guide/lang" ]]; then
  for f in "$VAULT_DIR/dev-guide/lang/"*.md; do
    [[ -f "$f" ]] && system_context+="$(cat "$f")"$'\n\n'
  done
fi

# Build the task prompt
task_prompt="Execute this SDD. Read the file ${SDD_FILE} for the full spec. Implement everything in the Scope section, follow Constraints and Best Practices, write tests as specified, verify all Acceptance Criteria. When done, update the Work Log section in ${SDD_FILE} with agent=claude-code, date=$(date +%Y-%m-%d), decisions, output files, and retrospective. Commit with a message referencing ${sdd_id}."

# Build claude args
claude_args=()
if [[ "$auto_approve" == "true" ]]; then
  claude_args+=(--dangerously-skip-permissions)
fi
if [[ "$max_turns" != "0" ]]; then
  claude_args+=(--max-turns "$max_turns")
fi

echo "→ Agent working..."

# Snapshot current git state for file tracking
git_snapshot_before=$(git status --short 2>/dev/null || true)

# Background file monitor — shows changes every 5 seconds
(
  prev_snapshot="$git_snapshot_before"
  while true; do
    sleep 5
    # Check if parent process is still alive
    kill -0 $$ 2>/dev/null || break

    current=$(git status --short 2>/dev/null || true)
    if [[ "$current" != "$prev_snapshot" ]]; then
      elapsed=$(( SECONDS - start_time ))
      mins=$(( elapsed / 60 ))
      secs=$(( elapsed % 60 ))

      # Find new/changed files
      new_files=$(echo "$current" | grep '^?' | sed 's/^?? //' || true)
      mod_files=$(echo "$current" | grep '^ M\|^M' | sed 's/^ *M //' || true)

      for f in $new_files; do
        if ! echo "$prev_snapshot" | grep -qF "$f" 2>/dev/null; then
          printf "  ├── [%02d:%02d] Created %s\n" "$mins" "$secs" "$f"
          echo "[$(date +%H:%M:%S)] Created $f" >> "$log_file"
        fi
      done

      for f in $mod_files; do
        if ! echo "$prev_snapshot" | grep -qF "$f" 2>/dev/null; then
          printf "  ├── [%02d:%02d] Modified %s\n" "$mins" "$secs" "$f"
          echo "[$(date +%H:%M:%S)] Modified $f" >> "$log_file"
        fi
      done

      prev_snapshot="$current"
    fi
  done
) &
monitor_pid=$!

# Run claude
claude "${claude_args[@]}" \
  --append-system-prompt "$system_context" \
  -p "$task_prompt" 2>&1 | tee -a "$log_file"
exit_code=${PIPESTATUS[0]}

# Stop the file monitor
kill "$monitor_pid" 2>/dev/null || true
wait "$monitor_pid" 2>/dev/null || true

# Calculate duration
elapsed=$(( SECONDS - start_time ))
mins=$(( elapsed / 60 ))
secs=$(( elapsed % 60 ))

# Collect results
git_snapshot_after=$(git status --short 2>/dev/null || true)
new_commits=$(git log --oneline -5 2>/dev/null | head -5 || true)

# Write JSON execution report
report_file="$VAULT_DIR/logs/exec-${sdd_id}-$(date +%Y-%m-%dT%H:%M:%S).json"
{
  cat <<REPORT
{
  "sdd": "$sdd_id",
  "file": "$SDD_FILE",
  "runner": "claude-code",
  "started": "$(date -u -v-${elapsed}S +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u +%Y-%m-%dT%H:%M:%SZ)",
  "completed": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "duration_seconds": $elapsed,
  "exit_code": $exit_code,
  "status": "$(if [[ $exit_code -eq 0 ]]; then echo "success"; else echo "failed"; fi)"
}
REPORT
} > "$report_file"

# Log footer
{
  echo "---"
  echo "Completed: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "Duration: ${mins}m ${secs}s"
  echo "Exit code: $exit_code"
} >> "$log_file"

# Summary
echo ""
printf "  └── [%02d:%02d] Done (exit code: %d)\n" "$mins" "$secs" "$exit_code"
echo ""
echo "=== Claude Runner Complete ==="
echo "  SDD:      $sdd_id"
echo "  Duration: ${mins}m ${secs}s"
echo "  Log:      $log_file"
echo "  Report:   $report_file"
echo "  Work Log: $SDD_FILE"

if [[ $exit_code -ne 0 ]]; then
  echo ""
  echo "Error: Claude exited with code $exit_code" >&2
  exit $exit_code
fi
