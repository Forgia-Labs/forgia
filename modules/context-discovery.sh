#!/usr/bin/env bash
# modules/context-discovery.sh — Discover project context files for FD inclusion
# Can be sourced (provides discover_context, discover_context_formatted)
# or called directly (./modules/context-discovery.sh [--format=...] [--max=N] [--include=PATH])

# Category priorities (lower = higher priority)
_CD_CAT_ADR=1
_CD_CAT_TESTING=2
_CD_CAT_OPS=3
_CD_CAT_CONTRIBUTING=4
_CD_CAT_ARCHITECTURE=5
_CD_CAT_COMMUNITY=6
_CD_CAT_AGENT=7
_CD_CAT_FORGIA=8
_CD_CAT_DESIGN=9
_CD_CAT_DOCS=10
_CD_CAT_USER=11

_CD_MAX_SIZE=102400  # 100KB
_CD_MAX_FILES=30
_CD_SKIP_DIRS="node_modules vendor .git target dist build _templates"

# Collect discovered files into _cd_results array
# Each entry: "priority|category|path"
_cd_results=()

_cd_should_skip_path() {
  local path="$1"
  # Skip hidden dirs except .forgia/ and .claude/
  case "$path" in
    .forgia/*|.claude/*) return 1 ;;
    .*) return 0 ;;
  esac
  # Skip known vendor/build dirs and template dirs
  for skip in $_CD_SKIP_DIRS; do
    case "$path" in
      "$skip"/*|*/"$skip"/*) return 0 ;;
    esac
  done
  return 1
}

_cd_check_size() {
  local path="$1"
  if [[ ! -f "$path" ]]; then
    return 1
  fi
  local size
  size=$(wc -c < "$path" 2>/dev/null) || return 1
  # Trim whitespace (macOS wc pads)
  size="${size##* }"
  size="${size// /}"
  [[ "$size" -le "$_CD_MAX_SIZE" ]]
}

_cd_add() {
  local priority="$1" category="$2" path="$3"
  if _cd_should_skip_path "$path"; then
    return
  fi
  if ! _cd_check_size "$path"; then
    return
  fi
  _cd_results+=("${priority}|${category}|${path}")
}

_cd_scan_glob() {
  local priority="$1" category="$2" pattern="$3"
  # Use find for controlled traversal instead of bash glob
  # This handles the case where globs don't match (nullglob not set)
  local IFS=$'\n'
  local matches
  matches=$(eval "for f in $pattern; do [[ -f \"\$f\" ]] && echo \"\$f\"; done" 2>/dev/null) || true
  for match in $matches; do
    _cd_add "$priority" "$category" "$match"
  done
}

_cd_scan_find() {
  local priority="$1" category="$2" dir="$3" maxdepth="${4:-3}"
  [[ -d "$dir" ]] || return 0
  local IFS=$'\n'
  local matches
  matches=$(find "$dir" -maxdepth "$maxdepth" -name '*.md' -type f 2>/dev/null) || true
  for match in $matches; do
    _cd_add "$priority" "$category" "$match"
  done
}

_cd_scan_file() {
  local priority="$1" category="$2" file="$3"
  [[ -f "$file" ]] && _cd_add "$priority" "$category" "$file"
}

_cd_scan_top_level_docs() {
  [[ -d "docs" ]] || return 0
  local IFS=$'\n'
  local matches
  matches=$(find "docs" -maxdepth 1 -name '*.md' -type f 2>/dev/null) || true
  for match in $matches; do
    # Skip if already found by a more specific pattern
    local dominated=false
    for existing in "${_cd_results[@]}"; do
      local existing_path="${existing#*|}"
      existing_path="${existing_path#*|}"
      if [[ "$existing_path" == "$match" ]]; then
        dominated=true
        break
      fi
    done
    "$dominated" || _cd_add "$_CD_CAT_DOCS" "Docs" "$match"
  done
}

_cd_discover() {
  _cd_results=()

  # ADR
  _cd_scan_find "$_CD_CAT_ADR" "ADR" "docs/decisions" 1
  _cd_scan_find "$_CD_CAT_ADR" "ADR" "docs/adr" 1

  # Testing
  _cd_scan_find "$_CD_CAT_TESTING" "Testing" "docs/testing" 1

  # Operations
  _cd_scan_find "$_CD_CAT_OPS" "Operations" "docs/operations" 1
  _cd_scan_find "$_CD_CAT_OPS" "Operations" "docs/ops" 1

  # Contributing
  _cd_scan_file "$_CD_CAT_CONTRIBUTING" "Contributing" "CONTRIBUTING.md"

  # Architecture
  _cd_scan_file "$_CD_CAT_ARCHITECTURE" "Architecture" "ARCHITECTURE.md"
  _cd_scan_file "$_CD_CAT_ARCHITECTURE" "Architecture" "docs/ARCHITECTURE.md"

  # Community
  _cd_scan_file "$_CD_CAT_COMMUNITY" "Community" "CODE_OF_CONDUCT.md"

  # Agent
  _cd_scan_file "$_CD_CAT_AGENT" "Agent" "CLAUDE.md"
  _cd_scan_file "$_CD_CAT_AGENT" "Agent" ".claude/CLAUDE.md"

  # Forgia
  _cd_scan_file "$_CD_CAT_FORGIA" "Forgia" ".forgia/constitution.md"
  _cd_scan_find "$_CD_CAT_FORGIA" "Forgia" ".forgia/dev-guide" 3

  # Design
  _cd_scan_find "$_CD_CAT_DESIGN" "Design" "docs/design" 1

  # Top-level docs (deduplicated)
  _cd_scan_top_level_docs
}

_cd_sorted_results() {
  local max="${1:-$_CD_MAX_FILES}"
  # Sort by priority (first field), then deduplicate by path, limit to max
  printf '%s\n' "${_cd_results[@]}" | sort -t'|' -k1,1n -k3,3 | \
    awk -F'|' '!seen[$3]++' | head -n "$max"
}

# --- Public API ---

discover_context() {
  discover_context_formatted "markdown"
}

discover_context_formatted() {
  local format="${1:-markdown}"
  _cd_discover

  if [[ ${#_cd_results[@]} -eq 0 ]]; then
    return 1
  fi

  local sorted
  sorted=$(_cd_sorted_results "$_CD_MAX_FILES")

  if [[ -z "$sorted" ]]; then
    return 1
  fi

  case "$format" in
    markdown)
      echo "### Auto-discovered Context"
      echo ""
      while IFS='|' read -r _priority category path; do
        echo "- \`$path\` — $category"
      done <<< "$sorted"
      ;;
    plain)
      while IFS='|' read -r _priority category path; do
        local upper_cat
        upper_cat=$(echo "$category" | tr '[:lower:]' '[:upper:]')
        echo "$upper_cat $path"
      done <<< "$sorted"
      ;;
    json)
      echo "["
      local first=true
      while IFS='|' read -r _priority category path; do
        if "$first"; then
          first=false
        else
          echo ","
        fi
        # Escape for JSON
        local escaped_path="${path//\\/\\\\}"
        escaped_path="${escaped_path//\"/\\\"}"
        local escaped_cat="${category//\\/\\\\}"
        escaped_cat="${escaped_cat//\"/\\\"}"
        printf '  {"path": "%s", "category": "%s"}' "$escaped_path" "$escaped_cat"
      done <<< "$sorted"
      echo ""
      echo "]"
      ;;
    *)
      echo "Error: unknown format '$format'" >&2
      return 1
      ;;
  esac
}

# --- CLI entry point ---

_cd_main() {
  local format="markdown"
  local max="$_CD_MAX_FILES"
  local includes=()

  for arg in "$@"; do
    case "$arg" in
      --format=*) format="${arg#--format=}" ;;
      --max=*) max="${arg#--max=}" ;;
      --include=*) includes+=("${arg#--include=}") ;;
      --help|-h)
        cat <<USAGE
Usage: context-discovery.sh [options]

Options:
  --format=FORMAT    Output format: markdown (default), plain, json
  --max=N            Maximum files to list (default: 30)
  --include=PATH     Additional path to include (repeatable)

Exit codes:
  0   Context found
  1   No context found
USAGE
        return 0
        ;;
      *)
        echo "Error: unknown option '$arg'" >&2
        return 1
        ;;
    esac
  done

  _CD_MAX_FILES="$max"
  _cd_discover

  # Add --include paths
  for inc in "${includes[@]}"; do
    if [[ -f "$inc" ]]; then
      _cd_add "$_CD_CAT_USER" "User-specified" "$inc"
    elif [[ -d "$inc" ]]; then
      local IFS=$'\n'
      local matches
      matches=$(find "$inc" -maxdepth 1 -name '*.md' -type f 2>/dev/null) || true
      for match in $matches; do
        _cd_add "$_CD_CAT_USER" "User-specified" "$match"
      done
    fi
  done

  if [[ ${#_cd_results[@]} -eq 0 ]]; then
    return 1
  fi

  local sorted
  sorted=$(_cd_sorted_results "$_CD_MAX_FILES")

  if [[ -z "$sorted" ]]; then
    return 1
  fi

  case "$format" in
    markdown)
      echo "### Auto-discovered Context"
      echo ""
      while IFS='|' read -r _priority category path; do
        echo "- \`$path\` — $category"
      done <<< "$sorted"
      ;;
    plain)
      while IFS='|' read -r _priority category path; do
        local upper_cat
        upper_cat=$(echo "$category" | tr '[:lower:]' '[:upper:]')
        echo "$upper_cat $path"
      done <<< "$sorted"
      ;;
    json)
      echo "["
      local first=true
      while IFS='|' read -r _priority category path; do
        if "$first"; then
          first=false
        else
          echo ","
        fi
        local escaped_path="${path//\\/\\\\}"
        escaped_path="${escaped_path//\"/\\\"}"
        local escaped_cat="${category//\\/\\\\}"
        escaped_cat="${escaped_cat//\"/\\\"}"
        printf '  {"path": "%s", "category": "%s"}' "$escaped_path" "$escaped_cat"
      done <<< "$sorted"
      echo ""
      echo "]"
      ;;
    *)
      echo "Error: unknown format '$format'" >&2
      return 1
      ;;
  esac
}

# If executed directly (not sourced), run main
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  _cd_main "$@"
fi
