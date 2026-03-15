#!/usr/bin/env bash
set -euo pipefail

# Validate an SDD file (YAML or Markdown) for completeness before execution.
# Returns 0 if valid, 1 if issues found.

SDD_FILE="${1:?Usage: validate-sdd.sh <sdd-file>}"

if [[ ! -f "$SDD_FILE" ]]; then
  echo "Error: File not found: $SDD_FILE" >&2
  exit 1
fi

errors=()
warnings=()

# Detect format
if [[ "$SDD_FILE" == *.yaml ]] || [[ "$SDD_FILE" == *.yml ]]; then
  format="yaml"
else
  format="markdown"
fi

echo "=== SDD Validation: $(basename "$SDD_FILE") ==="
echo "  Format: $format"
echo ""

if [[ "$format" == "yaml" ]]; then
  # YAML validation
  if command -v yq >/dev/null 2>&1; then
    # Check required fields
    for field in meta.id meta.fd meta.title meta.status; do
      val=$(yq -r ".${field}" "$SDD_FILE" 2>/dev/null || echo "")
      if [[ -z "$val" ]] || [[ "$val" == "null" ]] || [[ "$val" == '{{'"'*'"'}}' ]]; then
        errors+=("Missing required field: $field")
      fi
    done

    # Check scope is filled
    scope=$(yq -r '.scope' "$SDD_FILE" 2>/dev/null || echo "")
    if [[ -z "$scope" ]] || [[ "$scope" == *"What to build"* ]]; then
      errors+=("Scope is empty or still contains template placeholder")
    fi

    # Check acceptance criteria exist
    ac_count=$(yq -r '.acceptance_criteria | length' "$SDD_FILE" 2>/dev/null || echo "0")
    if (( ac_count == 0 )); then
      errors+=("No acceptance criteria defined")
    else
      # Check for empty criteria
      empty_ac=$(yq -r '.acceptance_criteria[] | select(.criterion == "") | .criterion' "$SDD_FILE" 2>/dev/null | wc -l)
      if (( empty_ac > 0 )); then
        warnings+=("$empty_ac acceptance criteria have empty descriptions")
      fi
    fi

    # Check test requirements
    test_count=$(yq -r '.test_requirements | length' "$SDD_FILE" 2>/dev/null || echo "0")
    if (( test_count == 0 )); then
      warnings+=("No test requirements defined")
    fi

    # Check constraints
    lang=$(yq -r '.constraints.language' "$SDD_FILE" 2>/dev/null || echo "")
    if [[ -z "$lang" ]] || [[ "$lang" == "null" ]]; then
      warnings+=("No language specified in constraints")
    fi

    # Check constitution
    for check in code_standards commit_conventions no_hardcoded_secrets tests_sufficient; do
      val=$(yq -r ".constitution_check.${check}" "$SDD_FILE" 2>/dev/null || echo "false")
      if [[ "$val" != "true" ]]; then
        warnings+=("Constitution check not confirmed: $check")
      fi
    done

  else
    warnings+=("yq not installed — YAML deep validation skipped (install: brew install yq)")
    # Basic check: file is valid YAML
    if command -v python3 >/dev/null 2>&1; then
      if ! python3 -c "import yaml; yaml.safe_load(open('$SDD_FILE'))" 2>/dev/null; then
        errors+=("File is not valid YAML")
      fi
    fi
  fi

else
  # Markdown validation
  # Check frontmatter fields
  for field in id fd title status; do
    if ! grep -q "^${field}:" "$SDD_FILE"; then
      errors+=("Missing frontmatter field: $field")
    fi
  done

  # Check sections exist
  for section in "## Scope" "## Interfaces" "## Constraints" "## Test Requirements" "## Acceptance Criteria" "## Work Log"; do
    if ! grep -q "$section" "$SDD_FILE"; then
      errors+=("Missing section: $section")
    fi
  done

  # Check acceptance criteria have content
  ac_count=$(grep -c '^\- \[ \]' "$SDD_FILE" 2>/dev/null || echo "0")
  if (( ac_count == 0 )); then
    errors+=("No acceptance criteria defined (no checkboxes found)")
  fi

  # Check for template placeholders still present
  if grep -q '{{.*}}' "$SDD_FILE"; then
    placeholders=$(grep -c '{{.*}}' "$SDD_FILE")
    warnings+=("$placeholders template placeholders still present")
  fi

  # Check scope has content (not just comments)
  scope_content=$(sed -n '/^## Scope/,/^## /{/^## /d;/^<!--/d;/^$/d;p;}' "$SDD_FILE" 2>/dev/null | wc -l)
  if (( scope_content == 0 )); then
    errors+=("Scope section is empty")
  fi
fi

# Report
if (( ${#errors[@]} > 0 )); then
  echo "ERRORS (must fix before execution):"
  for e in "${errors[@]}"; do
    echo "  ✗ $e"
  done
  echo ""
fi

if (( ${#warnings[@]} > 0 )); then
  echo "WARNINGS (review recommended):"
  for w in "${warnings[@]}"; do
    echo "  ⚠ $w"
  done
  echo ""
fi

if (( ${#errors[@]} == 0 )) && (( ${#warnings[@]} == 0 )); then
  echo "✓ SDD is valid and ready for execution"
  exit 0
elif (( ${#errors[@]} == 0 )); then
  echo "✓ SDD is valid with ${#warnings[@]} warning(s)"
  exit 0
else
  echo "✗ SDD has ${#errors[@]} error(s) — fix before execution"
  exit 1
fi
