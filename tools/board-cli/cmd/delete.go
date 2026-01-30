package cmd

import (
	"fmt"
	"os"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <ID>",
	Short: "Delete a board element",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

var deleteForce bool

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Force delete even if element has children")
}

func runDelete(cmd *cobra.Command, args []string) error {
	id := args[0]

	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	elem := b.FindByID(id)
	if elem == nil {
		return fmt.Errorf("element %s not found", id)
	}

	// Check for children
	children := b.Children(elem)
	if len(children) > 0 && !deleteForce {
		return fmt.Errorf("%s has %d children — use --force to delete recursively", id, len(children))
	}

	// Clean up dependency references in other elements
	for _, other := range b.Elements {
		if other.ID() == elem.ID() {
			continue
		}

		pd, err := board.ParseProgressFile(other.ProgressPath())
		if err != nil {
			continue
		}

		changed := false

		// Remove from BlockedBy
		var newBlockedBy []string
		for _, bid := range pd.BlockedBy {
			if bid == elem.ID() {
				changed = true
				continue
			}
			newBlockedBy = append(newBlockedBy, bid)
		}
		pd.BlockedBy = newBlockedBy

		// Remove from Blocks
		var newBlocks []string
		for _, bid := range pd.Blocks {
			if bid == elem.ID() {
				changed = true
				continue
			}
			newBlocks = append(newBlocks, bid)
		}
		pd.Blocks = newBlocks

		if changed {
			board.WriteProgressFile(other.ProgressPath(), pd)
		}
	}

	// Delete the directory
	if err := os.RemoveAll(elem.Path); err != nil {
		return fmt.Errorf("deleting %s: %w", id, err)
	}

	fmt.Printf("Deleted %s\n", id)
	return nil
}
