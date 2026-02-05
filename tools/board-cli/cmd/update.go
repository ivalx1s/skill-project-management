package cmd

import (
	"fmt"
	"os"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

// UpdateResponse represents the JSON response for update command
type UpdateResponse struct {
	Updated UpdatedElement `json:"updated"`
	Message string         `json:"message"`
}

// UpdatedElement represents an updated element in JSON output
type UpdatedElement struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Assignee string `json:"assignee"`
}

var updateCmd = &cobra.Command{
	Use:   "update <ID>",
	Short: "Update element README.md fields",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

var (
	updateTitle       string
	updateDescription string
	updateScope       string
	updateAC          string
)

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New title")
	updateCmd.Flags().StringVar(&updateDescription, "description", "", "New description")
	updateCmd.Flags().StringVar(&updateScope, "scope", "", "New scope")
	updateCmd.Flags().StringVar(&updateAC, "ac", "", "New acceptance criteria")
}

func runUpdate(cmd *cobra.Command, args []string) error {
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

	rd, err := board.ParseReadmeFile(elem.ReadmePath())
	if err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("reading README.md: %v", err), nil)
			return nil
		}
		return fmt.Errorf("reading README.md: %w", err)
	}

	changed := false
	var changedFields []string
	if cmd.Flags().Changed("title") {
		rd.Title = updateTitle
		changed = true
		changedFields = append(changedFields, "title")
	}
	if cmd.Flags().Changed("description") {
		rd.Description = updateDescription
		changed = true
		changedFields = append(changedFields, "description")
	}
	if cmd.Flags().Changed("scope") {
		rd.Scope = updateScope
		changed = true
		changedFields = append(changedFields, "scope")
	}
	if cmd.Flags().Changed("ac") {
		rd.AC = updateAC
		changed = true
		changedFields = append(changedFields, "ac")
	}

	if !changed {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.ValidationError, "No changes specified. Use --title, --description, --scope, or --ac flags.", nil)
			return nil
		}
		fmt.Println("No changes specified. Use --title, --description, --scope, or --ac flags.")
		return nil
	}

	if err := board.WriteReadmeFile(elem.ReadmePath(), rd); err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("writing README.md: %v", err), nil)
			return nil
		}
		return fmt.Errorf("writing README.md: %w", err)
	}

	// Get progress data for status and assignee
	pd, _ := board.ParseProgressFile(elem.ProgressPath())
	status := ""
	assignee := ""
	if pd != nil {
		status = string(pd.Status)
		assignee = pd.AssignedTo
	}

	if JSONEnabled() {
		response := UpdateResponse{
			Updated: UpdatedElement{
				ID:       elem.ID(),
				Type:     string(elem.Type),
				Name:     rd.Title,
				Status:   status,
				Assignee: assignee,
			},
			Message: fmt.Sprintf("Updated fields: %v", changedFields),
		}
		return output.PrintJSON(os.Stdout, response)
	}

	fmt.Printf("Updated %s\n", id)
	return nil
}
