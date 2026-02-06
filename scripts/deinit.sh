#!/usr/bin/env zsh
#
# Deinit script for project-management skill.
# Removes built binaries, symlinks, and config. Does NOT uninstall Go.
#

set -euo pipefail

SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CLI_DIR="$SKILL_DIR/tools/board-cli"
TUI_DIR="$SKILL_DIR/tools/board-tui"
BIN_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.config/board-tui"

# --- Colors ---
red()   { print -P "%F{red}$1%f" }
green() { print -P "%F{green}$1%f" }
yellow(){ print -P "%F{yellow}$1%f" }

# --- 1. Remove symlinks from PATH ---
remove_symlink() {
  local link="$1"
  if [[ -L "$link" ]]; then
    rm "$link"
    green "Removed symlink: $link"
  elif [[ -e "$link" ]]; then
    yellow "Skipping $link (not a symlink)"
  fi
}

remove_symlinks() {
  green "Removing symlinks..."
  remove_symlink "$BIN_DIR/task-board"
  remove_symlink "$BIN_DIR/task-board-tui"
  # Legacy binary name
  remove_symlink "$BIN_DIR/board-tui"
}

# --- 2. Remove built binaries ---
remove_binaries() {
  green "Removing built binaries..."
  local removed=0

  for bin in "$CLI_DIR/task-board" "$TUI_DIR/task-board-tui" "$TUI_DIR/board-tui"; do
    if [[ -f "$bin" ]]; then
      rm "$bin"
      green "Removed: $bin"
      removed=$(( removed + 1 ))
    fi
  done

  if (( removed == 0 )); then
    yellow "No binaries found to remove"
  fi
}

# --- 3. Remove config ---
remove_config() {
  if [[ -d "$CONFIG_DIR" ]]; then
    rm -rf "$CONFIG_DIR"
    green "Removed config: $CONFIG_DIR"
  else
    yellow "No config directory found"
  fi
}

# --- 4. Verify ---
verify_clean() {
  green "Verifying cleanup..."
  local clean=true

  if command -v task-board &>/dev/null; then
    red "WARN: task-board still found in PATH: $(which task-board)"
    clean=false
  else
    green "task-board: not in PATH"
  fi

  if command -v task-board-tui &>/dev/null; then
    red "WARN: task-board-tui still found in PATH: $(which task-board-tui)"
    clean=false
  else
    green "task-board-tui: not in PATH"
  fi

  if [[ -d "$CONFIG_DIR" ]]; then
    red "WARN: config still exists: $CONFIG_DIR"
    clean=false
  else
    green "Config: clean"
  fi

  if $clean; then
    green "All clean!"
  fi
}

# --- Run ---
print ""
green "=== project-management skill deinit ==="
print ""
remove_symlinks
remove_binaries
remove_config
verify_clean
print ""
green "=== Done ==="
