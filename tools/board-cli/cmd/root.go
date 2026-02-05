package cmd

import (
	"fmt"
	"os"

	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

var boardDir string
var jsonOutput bool

var rootCmd = &cobra.Command{
	Use:   "task-board",
	Short: "File-based task board CLI",
	Long:  "Manage a file-based task board with epics, stories, tasks, and bugs.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if jsonOutput {
			output.PrintError(os.Stderr, output.InternalError, err.Error(), nil)
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&boardDir, "board-dir", ".task-board", "Path to the board directory")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
}

// JSONEnabled returns true if JSON output is enabled
func JSONEnabled() bool {
	return jsonOutput
}
