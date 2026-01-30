package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move <ID>",
	Short: "Move element to a different parent",
	Args:  cobra.ExactArgs(1),
	RunE:  runMove,
}

var moveToFlag string

func init() {
	rootCmd.AddCommand(moveCmd)
	moveCmd.Flags().StringVar(&moveToFlag, "to", "", "Target parent ID (required)")
	moveCmd.MarkFlagRequired("to")
}

func runMove(cmd *cobra.Command, args []string) error {
	id := args[0]

	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	elem := b.FindByID(id)
	if elem == nil {
		return fmt.Errorf("element %s not found", id)
	}

	target := b.FindByID(moveToFlag)
	if target == nil {
		return fmt.Errorf("target %s not found", moveToFlag)
	}

	// Validate type compatibility
	switch elem.Type {
	case board.TaskType, board.BugType:
		if target.Type != board.StoryType {
			return fmt.Errorf("tasks and bugs can only be moved to stories (target is %s)", target.Type)
		}
	case board.StoryType:
		if target.Type != board.EpicType {
			return fmt.Errorf("stories can only be moved to epics (target is %s)", target.Type)
		}
	case board.EpicType:
		return fmt.Errorf("epics cannot be moved")
	}

	// Move directory
	dirName := filepath.Base(elem.Path)
	newPath := filepath.Join(target.Path, dirName)

	if err := os.Rename(elem.Path, newPath); err != nil {
		return fmt.Errorf("moving %s: %w", id, err)
	}

	fmt.Printf("Moved %s → %s\n", id, target.ID())
	return nil
}
