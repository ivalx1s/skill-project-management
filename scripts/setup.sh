#!/usr/bin/env zsh
#
# Setup script for project-management skill.
# Installs Go (if missing), builds CLI + TUI, symlinks to PATH.
#

set -euo pipefail

SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CLI_DIR="$SKILL_DIR/tools/board-cli"
TUI_DIR="$SKILL_DIR/tools/board-tui"
CLI_BINARY="$CLI_DIR/task-board"
TUI_BINARY="$TUI_DIR/task-board-tui"
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
  green "Building task-board (CLI)..."
  cd "$CLI_DIR"
  go build -o task-board .
  green "Built: $CLI_BINARY"
}

# --- 2b. Build TUI ---
build_tui() {
  green "Building task-board-tui (TUI)..."
  cd "$TUI_DIR"
  go build -o task-board-tui .
  green "Built: $TUI_BINARY"
}

# --- 3. Symlink to PATH ---
symlink_binary() {
  local name="$1"
  local target="$2"
  local link="$BIN_DIR/$name"

  mkdir -p "$BIN_DIR"

  if [[ -L "$link" ]]; then
    local existing
    existing="$(readlink "$link")"
    if [[ "$existing" == "$target" ]]; then
      green "Symlink already correct: $link -> $target"
      return
    fi
    yellow "Updating symlink (was: $existing)"
    rm "$link"
  elif [[ -e "$link" ]]; then
    red "$link exists and is not a symlink. Skipping."
    return
  fi

  ln -s "$target" "$link"
  green "Symlinked: $link -> $target"
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
  if command -v task-board-tui &>/dev/null; then
    green "Verified: task-board-tui available"
  else
    yellow "task-board-tui not found in PATH. Check PATH settings."
  fi
}

# --- Run ---
print ""
green "=== project-management skill setup ==="
print ""
install_go
build_cli
build_tui
symlink_binary "task-board" "$CLI_BINARY"
symlink_binary "task-board-tui" "$TUI_BINARY"
check_path
verify
print ""
green "=== Done ==="
