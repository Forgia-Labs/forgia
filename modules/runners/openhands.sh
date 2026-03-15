#!/usr/bin/env bash
set -euo pipefail

# OpenHands Runner
# Executes an SDD using OpenHands in a Docker container.
# Uses API keys — pay-per-token.

VAULT_DIR=".forgia"
SDD_FILE="${1:?Usage: openhands.sh <sdd-file>}"
CONFIG_FILE="$VAULT_DIR/config.toml"

if [[ ! -f "$SDD_FILE" ]]; then
  echo "Error: SDD file not found: $SDD_FILE" >&2
  exit 1
fi

# Parse config
oh_image="ghcr.io/all-hands-ai/openhands:latest"
oh_model="anthropic/claude-sonnet-4-20250514"
oh_workspace="/opt/workspace"
oh_max_iter="100"
oh_ui_port="3000"

if [[ -f "$CONFIG_FILE" ]]; then
  oh_image=$(grep -A10 '\[runner\.openhands\]' "$CONFIG_FILE" | grep 'image' | sed 's/.*= *"*//;s/"*$//' || echo "$oh_image")
  oh_model=$(grep -A10 '\[runner\.openhands\]' "$CONFIG_FILE" | grep 'model' | head -1 | sed 's/.*= *"*//;s/"*$//' || echo "$oh_model")
  oh_workspace=$(grep -A10 '\[runner\.openhands\]' "$CONFIG_FILE" | grep 'workspace_mount' | sed 's/.*= *"*//;s/"*$//' || echo "$oh_workspace")
  oh_max_iter=$(grep -A10 '\[runner\.openhands\]' "$CONFIG_FILE" | grep 'max_iterations' | sed 's/.*= *//' | tr -d ' ' || echo "$oh_max_iter")
  oh_ui_port=$(grep -A10 '\[runner\.openhands\]' "$CONFIG_FILE" | grep 'ui_port' | sed 's/.*= *//' | tr -d ' ' || echo "$oh_ui_port")
fi

# Extract SDD metadata
sdd_id=$(grep '^id:' "$SDD_FILE" | head -1 | sed 's/id: *//')
sdd_title=$(grep '^title:' "$SDD_FILE" | head -1 | sed 's/title: *"*//;s/"*$//')

echo "=== Forgia OpenHands Runner ==="
echo "  SDD:        $sdd_id — $sdd_title"
echo "  Image:      $oh_image"
echo "  Model:      $oh_model"
echo "  Max iter:   $oh_max_iter"
echo "  UI port:    $oh_ui_port"
echo ""

# Check Docker
if ! command -v docker >/dev/null 2>&1; then
  echo "Error: Docker not found. Install Docker first." >&2
  exit 1
fi

if ! docker info >/dev/null 2>&1; then
  echo "Error: Docker daemon not running." >&2
  exit 1
fi

# Check OpenHands image
if ! docker image inspect "$oh_image" >/dev/null 2>&1; then
  echo "→ Pulling OpenHands image..."
  docker pull "$oh_image"
fi

# Build the prompt from SDD + context
sdd_content=$(cat "$SDD_FILE")

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

constitution=""
if [[ -f "$VAULT_DIR/constitution.md" ]]; then
  constitution=$(cat "$VAULT_DIR/constitution.md")
fi

guardrails=""
if [[ -f "$VAULT_DIR/guardrails/deny.toml" ]]; then
  guardrails=$(cat "$VAULT_DIR/guardrails/deny.toml")
fi

# Write the full prompt to a temp file (mounted into container)
prompt_file=$(mktemp)
cat > "$prompt_file" <<PROMPT
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
7. When done, create a file WORK_LOG.md with:
   - Agent: openhands
   - Date: $(date +%Y-%m-%d)
   - Decisions: key implementation decisions you made
   - Output: files created/modified
   - Retrospective: what went well, what was tricky
7. Commit your changes with a descriptive message referencing $sdd_id
PROMPT

# Determine API key to pass
api_env=""
if [[ -n "${ANTHROPIC_API_KEY:-}" ]]; then
  api_env="-e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY"
elif [[ -n "${OPENAI_API_KEY:-}" ]]; then
  api_env="-e OPENAI_API_KEY=$OPENAI_API_KEY"
else
  echo "Warning: No API key found (ANTHROPIC_API_KEY or OPENAI_API_KEY)" >&2
  echo "  If using a local model (ollama), this is fine." >&2
fi

# Launch container
project_dir="$(pwd)"
container_name="forgia-${sdd_id:-sdd}"

echo "→ Launching OpenHands container: $container_name"
echo ""

port_args=""
if [[ "$oh_ui_port" != "0" ]]; then
  port_args="-p ${oh_ui_port}:3000"
  echo "  UI available at: http://localhost:${oh_ui_port}"
fi

# shellcheck disable=SC2086
docker run \
  --name "$container_name" \
  --rm \
  -v "${project_dir}:${oh_workspace}" \
  -v "${prompt_file}:/tmp/sdd-prompt.md:ro" \
  $port_args \
  $api_env \
  -e LLM_MODEL="$oh_model" \
  -e MAX_ITERATIONS="$oh_max_iter" \
  -e WORKSPACE_BASE="$oh_workspace" \
  "$oh_image" \
  --task-file /tmp/sdd-prompt.md

# Cleanup
rm -f "$prompt_file"

# Copy work log back to SDD if it exists
if [[ -f "WORK_LOG.md" ]]; then
  echo ""
  echo "→ Merging Work Log into SDD..."
  work_log_content=$(cat WORK_LOG.md)

  # Append to SDD's Work Log section
  if grep -q '## Work Log' "$SDD_FILE"; then
    # Replace empty Work Log with actual content
    sed -i.bak "/## Work Log/,/## /{ /## Work Log/a\\
$work_log_content
}" "$SDD_FILE"
    rm -f "${SDD_FILE}.bak"
  fi

  rm -f WORK_LOG.md
  echo "  Work Log merged into $SDD_FILE"
fi

echo ""
echo "=== OpenHands Runner Complete ==="
echo "  SDD: $sdd_id"
echo "  Check $SDD_FILE for the Work Log"
