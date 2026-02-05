package cmd

import (
	"fmt"
	"os"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

// UnassignResponse represents the JSON response for unassign command
type UnassignResponse struct {
	Updated UpdatedElement `json:"updated"`
	Message string         `json:"message"`
}

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

	pd, err := board.ParseProgressFile(elem.ProgressPath())
	if err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("reading progress for %s: %v", id, err), nil)
			return nil
		}
		return fmt.Errorf("reading progress for %s: %w", id, err)
	}

	if pd.AssignedTo == "" {
		if JSONEnabled() {
			// Get name from README
			rd, _ := board.ParseReadmeFile(elem.ReadmePath())
			name := ""
			if rd != nil {
				name = rd.Title
			}

			response := UnassignResponse{
				Updated: UpdatedElement{
					ID:       elem.ID(),
					Type:     string(elem.Type),
					Name:     name,
					Status:   string(pd.Status),
					Assignee: "",
				},
				Message: "Not assigned to anyone",
			}
			return output.PrintJSON(os.Stdout, response)
		}
		fmt.Printf("%s: not assigned to anyone\n", id)
		return nil
	}

	pd.AssignedTo = ""

	if err := board.WriteProgressFile(elem.ProgressPath(), pd); err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("writing progress for %s: %v", id, err), nil)
			return nil
		}
		return fmt.Errorf("writing progress for %s: %w", id, err)
	}

	if JSONEnabled() {
		// Get name from README
		rd, _ := board.ParseReadmeFile(elem.ReadmePath())
		name := ""
		if rd != nil {
			name = rd.Title
		}

		response := UnassignResponse{
			Updated: UpdatedElement{
				ID:       elem.ID(),
				Type:     string(elem.Type),
				Name:     name,
				Status:   string(pd.Status),
				Assignee: "",
			},
			Message: "Unassigned",
		}
		return output.PrintJSON(os.Stdout, response)
	}

	fmt.Printf("%s: unassigned\n", id)
	return nil
}
