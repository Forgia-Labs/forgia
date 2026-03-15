#!/usr/bin/env bash
# tests/mock-gh.sh — mock gh CLI for testing
# Intercepts: gh api repos/*/issues/*, gh auth status
# Returns canned JSON responses from tests/fixtures/

set -euo pipefail

# Resolve real path (handles symlinks)
_MOCK_GH_REAL="${BASH_SOURCE[0]}"
if command -v readlink >/dev/null 2>&1; then
  _MOCK_GH_REAL="$(readlink -f "${BASH_SOURCE[0]}" 2>/dev/null || readlink "${BASH_SOURCE[0]}" 2>/dev/null || echo "${BASH_SOURCE[0]}")"
fi
FIXTURES_DIR="$(cd "$(dirname "$_MOCK_GH_REAL")/fixtures" && pwd)"

if [[ "$1" == "api" ]]; then
  endpoint="$2"

  # Match: repos/{owner}/{repo}/issues/{number}
  if [[ "$endpoint" =~ repos/([^/]+)/([^/]+)/issues/([0-9]+)$ ]]; then
    number="${BASH_REMATCH[3]}"
    fixture="$FIXTURES_DIR/issue-${number}.json"
    if [[ -f "$fixture" ]]; then
      # Handle --jq flag
      if [[ "${3:-}" == "--jq" ]]; then
        # Use python for jq simulation if available, otherwise cat
        if command -v python3 >/dev/null 2>&1; then
          python3 -c "
import json, sys
data = json.load(open('$fixture'))
expr = '${4}'
# Simple jq expressions
if expr == '.title': print(data['title'])
elif expr == '.body': print(data.get('body', ''))
elif expr == '.body // empty': print(data.get('body', ''))
elif expr == '.state': print(data.get('state', 'open'))
elif expr == '.assignee.login // empty': print(data.get('assignee', {}).get('login', '') if data.get('assignee') else '')
elif expr == '.milestone.title // empty': print(data.get('milestone', {}).get('title', '') if data.get('milestone') else '')
elif expr.startswith('[.labels'): print(', '.join(l['name'] for l in data.get('labels', [])))
else: json.dump(data, sys.stdout)
"
        else
          cat "$fixture"
        fi
      else
        cat "$fixture"
      fi
    else
      echo "Error: issue not found" >&2
      exit 1
    fi

  # Match: repos/{owner}/{repo}/issues/{number}/comments
  elif [[ "$endpoint" =~ repos/([^/]+)/([^/]+)/issues/([0-9]+)/comments$ ]]; then
    number="${BASH_REMATCH[3]}"
    fixture="$FIXTURES_DIR/issue-${number}-comments.json"
    if [[ -f "$fixture" ]]; then
      # Handle --jq flag for comments
      if [[ "${3:-}" == "--jq" ]]; then
        if command -v python3 >/dev/null 2>&1; then
          python3 -c "
import json, sys
data = json.load(open('$fixture'))
expr = '${4}'
if expr == '.[].body':
    for c in data:
        print(c['body'])
else:
    json.dump(data, sys.stdout)
"
        else
          cat "$fixture"
        fi
      else
        cat "$fixture"
      fi
    else
      echo "[]"
    fi
  else
    echo "Error: unknown endpoint: $endpoint" >&2
    exit 1
  fi

elif [[ "$1" == "auth" && "$2" == "status" ]]; then
  echo "Logged in to github.com as testuser"
  exit 0

else
  echo "Error: mock-gh does not support: $*" >&2
  exit 1
fi
