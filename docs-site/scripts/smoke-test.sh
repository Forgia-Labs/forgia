#!/usr/bin/env bash
set -euo pipefail

OUTPUT_DIR=".output/public"

# Critical pages that must be present after nuxt generate
PAGES=(
  "index.html"
  "docs/index.html"
  "docs/getting-started/index.html"
  "docs/getting-started/installation/index.html"
  "docs/getting-started/first-feature/index.html"
  "docs/fd/index.html"
  "docs/sdd/index.html"
  "docs/cli/index.html"
  "docs/constitution/index.html"
)

# Ensure generate has been run
if [[ ! -d "$OUTPUT_DIR" ]]; then
  echo "ERROR: $OUTPUT_DIR not found — run 'pnpm run generate' first" >&2
  exit 1
fi

errors=0

# Check each required page exists
for page in "${PAGES[@]}"; do
  local_path="$OUTPUT_DIR/$page"
  if [[ ! -f "$local_path" ]]; then
    echo "MISSING: $local_path" >&2
    errors=$(( errors + 1 ))
  fi
done

# Check content source files don't contain hardcoded /forgia/ prefix in links
# (internal links must use /docs/... and let Nuxt handle baseURL)
while IFS= read -r -d '' file; do
  if grep -q '/forgia/docs/' "$file"; then
    echo "HARDCODED BASEURL: $file contains /forgia/docs/ — use /docs/ instead" >&2
    errors=$(( errors + 1 ))
  fi
done < <(find content -name "*.md" -print0 2>/dev/null)

if [[ $errors -gt 0 ]]; then
  echo "Smoke test FAILED: $errors error(s)" >&2
  exit 1
fi

echo "Smoke test passed — $((${#PAGES[@]})) pages verified"
