package cmd

import (
	"fmt"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/spf13/cobra"
)

var unassignCmd = &cobra.Command{
	Use:   "unassign <ID>",
	Short: "Unassign element from its agent",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnassign,
}

func init() {
	rootCmd.AddCommand(unassignCmd)
}

func runUnassign(cmd *cobra.Command, args []string) error {
	id := args[0]

	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	elem := b.FindByID(id)
	if elem == nil {
		return fmt.Errorf("element %s not found", id)
	}

	pd, err := board.ParseProgressFile(elem.ProgressPath())
	if err != nil {
		return fmt.Errorf("reading progress for %s: %w", id, err)
	}

	if pd.AssignedTo == "" {
		fmt.Printf("%s: not assigned to anyone\n", id)
		return nil
	}

	pd.AssignedTo = ""

	if err := board.WriteProgressFile(elem.ProgressPath(), pd); err != nil {
		return fmt.Errorf("writing progress for %s: %w", id, err)
	}

	fmt.Printf("%s: unassigned\n", id)
	return nil
}
