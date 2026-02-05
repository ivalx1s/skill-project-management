package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI",
	Long:  "Launch the interactive terminal UI for browsing the task board.",
	RunE:  runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(cmd *cobra.Command, args []string) error {
	// Find the TUI binary
	tuiBinary, err := findTUIBinary()
	if err != nil {
		return err
	}

	// Get absolute path to board directory
	absBoardDir, err := filepath.Abs(boardDir)
	if err != nil {
		return fmt.Errorf("resolving board path: %w", err)
	}

	// Launch TUI with board directory
	tuiCmd := exec.Command(tuiBinary, "--board-dir", absBoardDir)
	tuiCmd.Stdin = os.Stdin
	tuiCmd.Stdout = os.Stdout
	tuiCmd.Stderr = os.Stderr

	return tuiCmd.Run()
}

// findTUIBinary locates the board-tui binary
func findTUIBinary() (string, error) {
	// 1. Check next to this binary
	if execPath, err := os.Executable(); err == nil {
		dir := filepath.Dir(execPath)
		candidate := filepath.Join(dir, "board-tui")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	// 2. Check in PATH
	if path, err := exec.LookPath("board-tui"); err == nil {
		return path, nil
	}

	// 3. Check ~/.local/bin
	if home, err := os.UserHomeDir(); err == nil {
		candidate := filepath.Join(home, ".local", "bin", "board-tui")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("board-tui binary not found\n\nInstall it with:\n  cd tools/board-tui && swift build -c release\n  cp .build/release/board-tui ~/.local/bin/")
}
