#!/usr/bin/env bash
set -euo pipefail

# Forgia Installer
# Usage: git clone ... && cd forgia && ./install.sh [--no-interactive]

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
INTERACTIVE=true

for arg in "$@"; do
  case "$arg" in
    --no-interactive) INTERACTIVE=false ;;
  esac
done

echo "=== Forgia Installer ==="
echo ""

# Sanity check — Go must be installed
if ! command -v go >/dev/null 2>&1; then
  echo "Error: Go is required but not installed." >&2
  echo "  Install Go: https://go.dev/dl/" >&2
  exit 1
fi

# --- 1. Build the Go binary ---

echo "→ Building forgia..."
(cd "$REPO_DIR" && go build -o build/forgia ./cmd/forgia)
echo "✓ Built: $REPO_DIR/build/forgia"

# --- 2. Symlink forgia binary ---

mkdir -p "$INSTALL_DIR"

if [[ -e "$INSTALL_DIR/forgia" ]] && [[ ! -L "$INSTALL_DIR/forgia" ]]; then
  echo "⚠ $INSTALL_DIR/forgia exists and is not a symlink."
  if [[ "$INTERACTIVE" == "true" ]]; then
    read -rp "  Overwrite? (y/N) " confirm
    [[ "$confirm" =~ ^[Yy]$ ]] || { echo "Aborted."; exit 0; }
  else
    echo "  Overwriting (--no-interactive)"
  fi
fi

ln -sf "$REPO_DIR/build/forgia" "$INSTALL_DIR/forgia"
echo "✓ forgia linked: $INSTALL_DIR/forgia → $REPO_DIR/build/forgia"

# --- 3. Add to PATH automatically ---

# Detect the right shell RC file
detect_shell_rc() {
  case "${SHELL:-}" in
    */zsh)
      [[ -f "$HOME/.zshrc" ]] && echo "$HOME/.zshrc" && return
      echo "$HOME/.zshrc" && return
      ;;
    */bash)
      [[ -f "$HOME/.bashrc" ]] && echo "$HOME/.bashrc" && return
      [[ -f "$HOME/.bash_profile" ]] && echo "$HOME/.bash_profile" && return
      echo "$HOME/.bashrc" && return
      ;;
  esac

  [[ -f "$HOME/.zshrc" ]] && echo "$HOME/.zshrc" && return
  [[ -f "$HOME/.bashrc" ]] && echo "$HOME/.bashrc" && return
  [[ -f "$HOME/.profile" ]] && echo "$HOME/.profile" && return

  echo ""
}

PATH_LINE='export PATH="$HOME/.local/bin:$PATH"'
shell_rc=$(detect_shell_rc)

if echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
  echo "✓ $INSTALL_DIR already in PATH"
elif [[ -n "$shell_rc" ]]; then
  if grep -qF '.local/bin' "$shell_rc" 2>/dev/null; then
    echo "✓ PATH entry already in $shell_rc (restart shell to activate)"
  else
    if [[ "$INTERACTIVE" == "true" ]]; then
      echo ""
      echo "  $INSTALL_DIR is not in your PATH."
      read -rp "  Add to $shell_rc? (Y/n) " confirm
      if [[ ! "$confirm" =~ ^[Nn]$ ]]; then
        echo "" >> "$shell_rc"
        echo "# Forgia CLI" >> "$shell_rc"
        echo "$PATH_LINE" >> "$shell_rc"
        echo "✓ PATH added to $shell_rc"
        export PATH="$INSTALL_DIR:$PATH"
      else
        echo "  Skipped. Add manually:"
        echo "    echo '$PATH_LINE' >> $shell_rc"
      fi
    else
      echo "" >> "$shell_rc"
      echo "# Forgia CLI" >> "$shell_rc"
      echo "$PATH_LINE" >> "$shell_rc"
      echo "✓ PATH added to $shell_rc (--no-interactive)"
      export PATH="$INSTALL_DIR:$PATH"
    fi
  fi
else
  echo "⚠ Could not detect shell RC file. Add manually:"
  echo "    $PATH_LINE"
fi

# --- 4. Install Claude Code commands ---

if command -v mise >/dev/null 2>&1; then
  echo ""
  echo "→ Installing Claude Code commands..."
  (cd "$REPO_DIR" && mise trust 2>/dev/null || true)
  (cd "$REPO_DIR" && mise run claude:install) || echo "  ⚠ Failed (non-blocking)"

  echo ""
  echo "→ Installing optional tools (fswatch, yq)..."
  (cd "$REPO_DIR" && mise run tools:install) || echo "  ⚠ Some tools failed (non-blocking)"
else
  echo ""
  echo "⚠ mise not found — skipping Claude Code commands and tools"
  echo "  Install mise: https://mise.jdx.dev"
  echo "  Then run: cd $REPO_DIR && mise run claude:install"
fi

# --- 5. Verify ---

echo ""
echo "=== Verifying ==="

if command -v forgia >/dev/null 2>&1; then
  echo "✓ forgia is in PATH"
else
  echo "✗ forgia not found in PATH"
  echo "  Restart your terminal, then try: forgia doctor"
fi

# --- 6. Summary ---

echo ""
echo "=== Installation Complete ==="
echo ""
echo "Next steps:"
echo "  1. Open a new terminal (or: source $shell_rc)"
echo "  2. cd /path/to/your/project"
echo "  3. forgia init"
echo "  4. forgia doctor"
echo ""
echo "First time? https://github.com/Deepzima/forgia#quick-start"
