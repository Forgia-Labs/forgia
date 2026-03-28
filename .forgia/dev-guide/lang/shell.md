# Shell Conventions (Bash/Zsh)

> Auto-loaded when `.sh` or `.zsh` files are detected, or when the project uses shell scripts.

## Script Header

Every script must start with:

```bash
#!/usr/bin/env bash
set -euo pipefail
```

- `set -e`: exit on error
- `set -u`: error on undefined variables
- `set -o pipefail`: catch errors in pipes

## Variables

- Use `local` for function-scoped variables
- Quote all variable expansions: `"$var"` not `$var`
- Use `${var:-default}` for optional variables with defaults
- Use `${var:?error message}` for required variables
- Constants in `UPPER_CASE`, locals in `lower_case`

```bash
readonly CONFIG_DIR="/etc/myapp"
local output_file="${1:?Usage: process <file>}"
```

## Functions

- Naming: `prefix_action()` (e.g., `fd_new`, `ops_status`, `oh_start`)
- Always use `local` for variables inside functions
- Usage strings on stderr: `echo "Usage: ..." >&2`
- Return codes: 0=success, 1=logic error, 64=usage error, 127=tool missing

```bash
cmd_install() {
  local module="${1:?Module name required}"
  local module_path="$MODULE_DIR/$module"

  if [[ ! -d "$module_path" ]]; then
    echo "Unknown module: $module" >&2
    return 1
  fi

  echo "Installing $module"
  "$module_path/install.sh"
}
```

## Conditionals

- Use `[[ ]]` not `[ ]` (bash/zsh extended test)
- Use `(( ))` for arithmetic
- String comparison: `==` inside `[[ ]]`
- File tests: `-f` (file), `-d` (dir), `-x` (executable), `-e` (exists)
- Command check: `command -v tool >/dev/null 2>&1`

## Loops & Iteration

- Use `for f in dir/*.md` over `find` when possible
- Guard globs: `[[ -f "$f" ]] || continue`
- Use `while IFS= read -r line` for reading files
- Avoid subshell loops when modifying variables

## Platform Detection

```bash
case "$(uname -s)" in
  Darwin) platform="macos" ;;
  Linux)  platform="linux" ;;
  *)      echo "Unsupported platform" >&2; exit 1 ;;
esac
```

## Dependencies

- Prefer coreutils commands over external tools
- Use `command -v` to check tool availability before using
- Install via package manager when needed (brew, apt)

## Style

- Linter: `shellcheck` — mandatory
- Indent: 2 spaces
- Max line length: 100 characters (break with `\`)
- Heredocs: use `<<'EOF'` (single-quoted) to prevent expansion when not needed
- No trailing whitespace

## Testing

- Syntax check: `bash -n script.sh`
- Lint: `shellcheck script.sh`
- Test manually in a clean environment (container or fresh shell)
- For complex scripts: use `bats` testing framework

## What NOT to Do

- Don't use `eval` — it's almost always a security risk
- Don't parse `ls` output — use globs
- Don't use backticks `` `cmd` `` — use `$(cmd)`
- Don't use `echo` for error messages — use `echo ... >&2`
- Don't rely on `$?` across multiple commands — check immediately
- Don't use `cd` without `|| exit` or in a subshell `(cd dir && ...)`
