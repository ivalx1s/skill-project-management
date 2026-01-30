package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var boardDir string

var rootCmd = &cobra.Command{
	Use:   "task-board",
	Short: "File-based task board CLI",
	Long:  "Manage a file-based task board with epics, stories, tasks, and bugs.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&boardDir, "board-dir", ".task-board", "Path to the board directory")
}
