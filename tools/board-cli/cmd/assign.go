package cmd

import (
	"fmt"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/spf13/cobra"
)

var assignCmd = &cobra.Command{
	Use:   "assign <ID>",
	Short: "Assign element to an agent",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssign,
}

var assignAgent string

func init() {
	rootCmd.AddCommand(assignCmd)
	assignCmd.Flags().StringVar(&assignAgent, "agent", "", "Agent name (required)")
	assignCmd.MarkFlagRequired("agent")
}

func runAssign(cmd *cobra.Command, args []string) error {
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

	pd.AssignedTo = assignAgent

	if err := board.WriteProgressFile(elem.ProgressPath(), pd); err != nil {
		return fmt.Errorf("writing progress for %s: %w", id, err)
	}

	fmt.Printf("%s: assigned to %s\n", id, assignAgent)
	return nil
}
