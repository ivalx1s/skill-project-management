#!/usr/bin/env zsh
#
# Setup script for project-management skill.
# Installs Go (if missing), builds the CLI, symlinks to PATH.
#

set -euo pipefail

SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CLI_DIR="$SKILL_DIR/tools/board-cli"
BINARY="$CLI_DIR/task-board"
BIN_DIR="$HOME/.local/bin"

# --- Colors ---
red()   { print -P "%F{red}$1%f" }
green() { print -P "%F{green}$1%f" }
yellow(){ print -P "%F{yellow}$1%f" }

# --- 1. Check / install Go ---
install_go() {
  if command -v go &>/dev/null; then
    green "Go already installed: $(go version)"
    return
  fi

  yellow "Go not found. Installing via Homebrew..."

  if ! command -v brew &>/dev/null; then
    red "Homebrew not found. Install it first: https://brew.sh"
    exit 1
  fi

  brew install go
  green "Go installed: $(go version)"
}

# --- 2. Build CLI ---
build_cli() {
  green "Building task-board..."
  cd "$CLI_DIR"
  go build -o task-board .
  green "Built: $BINARY"
}

# --- 3. Symlink to PATH ---
symlink_cli() {
  mkdir -p "$BIN_DIR"

  if [[ -L "$BIN_DIR/task-board" ]]; then
    local existing
    existing="$(readlink "$BIN_DIR/task-board")"
    if [[ "$existing" == "$BINARY" ]]; then
      green "Symlink already correct: $BIN_DIR/task-board -> $BINARY"
      return
    fi
    yellow "Updating symlink (was: $existing)"
    rm "$BIN_DIR/task-board"
  elif [[ -e "$BIN_DIR/task-board" ]]; then
    red "$BIN_DIR/task-board exists and is not a symlink. Skipping."
    return
  fi

  ln -s "$BINARY" "$BIN_DIR/task-board"
  green "Symlinked: $BIN_DIR/task-board -> $BINARY"
}

# --- 4. Verify PATH ---
check_path() {
  if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    yellow "WARNING: $BIN_DIR is not in PATH."
    yellow "Add to ~/.zshrc:  export PATH=\"\$HOME/.local/bin:\$PATH\""
  fi
}

# --- 5. Verify ---
verify() {
  if command -v task-board &>/dev/null; then
    green "Verified: $(task-board --version 2>&1 || echo 'task-board available')"
  else
    yellow "task-board not found in PATH. Check PATH settings."
  fi
}

# --- Run ---
print ""
green "=== project-management skill setup ==="
print ""
install_go
build_cli
symlink_cli
check_path
verify
print ""
green "=== Done ==="
