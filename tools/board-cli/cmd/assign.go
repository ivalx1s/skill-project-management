package cmd

import (
	"fmt"
	"os"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

// AssignResponse represents the JSON response for assign command
type AssignResponse struct {
	Updated UpdatedElement `json:"updated"`
	Message string         `json:"message"`
}

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

	pd.AssignedTo = assignAgent

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

		response := AssignResponse{
			Updated: UpdatedElement{
				ID:       elem.ID(),
				Type:     string(elem.Type),
				Name:     name,
				Status:   string(pd.Status),
				Assignee: assignAgent,
			},
			Message: fmt.Sprintf("Assigned to %s", assignAgent),
		}
		return output.PrintJSON(os.Stdout, response)
	}

	fmt.Printf("%s: assigned to %s\n", id, assignAgent)
	return nil
}
