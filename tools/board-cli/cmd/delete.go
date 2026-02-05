package cmd

import (
	"fmt"
	"os"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

// DeleteResponse represents the JSON output for delete command
type DeleteResponse struct {
	Deleted DeletedElement `json:"deleted"`
	Message string         `json:"message"`
}

// DeletedElement represents the deleted element info
type DeletedElement struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

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
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("loading board: %v", err), nil)
			return nil
		}
		return fmt.Errorf("loading board: %w", err)
	}

	elem := b.FindByID(id)
	if elem == nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.NotFound, fmt.Sprintf("Element %s not found", id), map[string]interface{}{
				"id": id,
			})
			return nil
		}
		return fmt.Errorf("element %s not found", id)
	}

	// Check for children
	children := b.Children(elem)
	if len(children) > 0 && !deleteForce {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.ValidationError, fmt.Sprintf("%s has %d children — use --force to delete recursively", id, len(children)), map[string]interface{}{
				"id":            id,
				"childrenCount": len(children),
			})
			return nil
		}
		return fmt.Errorf("%s has %d children — use --force to delete recursively", id, len(children))
	}

	// Get element name from README before deletion (for JSON output)
	var elemName string
	if JSONEnabled() {
		rd, err := board.ParseReadmeFile(elem.ReadmePath())
		if err == nil {
			elemName = rd.Title
		} else {
			elemName = elem.Name
		}
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
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("deleting %s: %v", id, err), nil)
			return nil
		}
		return fmt.Errorf("deleting %s: %w", id, err)
	}

	// JSON output
	if JSONEnabled() {
		response := DeleteResponse{
			Deleted: DeletedElement{
				ID:   elem.ID(),
				Type: string(elem.Type),
				Name: elemName,
			},
			Message: "Element deleted successfully",
		}
		return output.PrintJSON(os.Stdout, response)
	}

	fmt.Printf("Deleted %s\n", id)
	return nil
}
