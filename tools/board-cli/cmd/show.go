package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

// ShowElementJSON represents the JSON output for show command
type ShowElementJSON struct {
	ID                 string              `json:"id"`
	Type               string              `json:"type"`
	Name               string              `json:"name"`
	Status             string              `json:"status"`
	Assignee           string              `json:"assignee"`
	Parent             string              `json:"parent"`
	Path               string              `json:"path"`
	CreatedAt          string              `json:"createdAt"`
	UpdatedAt          string              `json:"updatedAt"`
	BlockedBy          []string            `json:"blockedBy"`
	Blocks             []string            `json:"blocks"`
	Description        string              `json:"description"`
	AcceptanceCriteria string              `json:"acceptanceCriteria"`
	Checklist          []ChecklistItemJSON `json:"checklist"`
	Notes              []NoteJSON          `json:"notes"`
}

// ChecklistItemJSON represents a checklist item in JSON output
type ChecklistItemJSON struct {
	Text string `json:"text"`
	Done bool   `json:"done"`
}

// NoteJSON represents a note entry in JSON output
type NoteJSON struct {
	Timestamp string `json:"timestamp"`
	Text      string `json:"text"`
}

// ShowResponse is the top-level JSON response for show command
type ShowResponse struct {
	Element ShowElementJSON `json:"element"`
}

var showCmd = &cobra.Command{
	Use:   "show <ID>",
	Short: "Show full element details",
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) error {
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
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("reading progress: %v", err), nil)
			return nil
		}
		return fmt.Errorf("reading progress: %w", err)
	}

	rd, err := board.ParseReadmeFile(elem.ReadmePath())
	if err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("reading README: %v", err), nil)
			return nil
		}
		return fmt.Errorf("reading README: %w", err)
	}

	// JSON output
	if JSONEnabled() {
		return outputShowJSON(b, elem, pd, rd)
	}

	// Text output (original)
	// Header
	fmt.Printf("%s%s: %s%s\n", output.Bold, elem.ID(), rd.Title, output.Reset)
	fmt.Printf("Path: %s\n", b.Ancestry(elem))
	fmt.Printf("Status: %s\n", output.ColorStatus(string(pd.Status)))
	fmt.Println()

	// Description
	if rd.Description != "" {
		fmt.Println("Description:")
		for _, line := range strings.Split(rd.Description, "\n") {
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}

	// Scope
	if rd.Scope != "" {
		fmt.Println("Scope:")
		for _, line := range strings.Split(rd.Scope, "\n") {
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}

	// Acceptance Criteria
	if rd.AC != "" {
		fmt.Println("Acceptance Criteria:")
		for _, line := range strings.Split(rd.AC, "\n") {
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}

	// Checklist
	if len(pd.Checklist) > 0 {
		fmt.Println("Checklist:")
		for i, item := range pd.Checklist {
			mark := "[ ]"
			if item.Checked {
				mark = "[x]"
			}
			fmt.Printf("  %d. %s %s\n", i+1, mark, item.Text)
		}
		fmt.Println()
	}

	// Blocked By
	if len(pd.BlockedBy) > 0 {
		fmt.Printf("Blocked By: %s\n", strings.Join(pd.BlockedBy, ", "))
	} else {
		fmt.Println("Blocked By: (none)")
	}

	// Blocks
	if len(pd.Blocks) > 0 {
		fmt.Printf("Blocks: %s\n", strings.Join(pd.Blocks, ", "))
	} else {
		fmt.Println("Blocks: (none)")
	}

	// Notes
	if pd.Notes != "" {
		fmt.Println()
		fmt.Println("Notes:")
		for _, line := range strings.Split(pd.Notes, "\n") {
			fmt.Printf("  %s\n", line)
		}
	}

	return nil
}

// outputShowJSON outputs the show command result as JSON
func outputShowJSON(b *board.Board, elem *board.Element, pd *board.ProgressData, rd *board.ReadmeData) error {
	// Build checklist
	checklist := make([]ChecklistItemJSON, len(pd.Checklist))
	for i, item := range pd.Checklist {
		checklist[i] = ChecklistItemJSON{
			Text: item.Text,
			Done: item.Checked,
		}
	}

	// Build notes - parse notes string into structured format
	// Notes in progress.md are stored as plain text, we'll treat each line as a note
	var notes []NoteJSON
	if pd.Notes != "" {
		for _, line := range strings.Split(pd.Notes, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				notes = append(notes, NoteJSON{
					Timestamp: "", // Notes don't have timestamps in current format
					Text:      line,
				})
			}
		}
	}

	// Ensure empty arrays instead of null
	blockedBy := pd.BlockedBy
	if blockedBy == nil {
		blockedBy = []string{}
	}
	blocks := pd.Blocks
	if blocks == nil {
		blocks = []string{}
	}
	if notes == nil {
		notes = []NoteJSON{}
	}

	// Format timestamps
	createdAt := ""
	if !pd.CreatedAt.IsZero() {
		createdAt = pd.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	updatedAt := ""
	if !pd.LastUpdate.IsZero() {
		updatedAt = pd.LastUpdate.UTC().Format("2006-01-02T15:04:05Z")
	}

	response := ShowResponse{
		Element: ShowElementJSON{
			ID:                 elem.ID(),
			Type:               string(elem.Type),
			Name:               rd.Title,
			Status:             string(pd.Status),
			Assignee:           pd.AssignedTo,
			Parent:             elem.ParentID,
			Path:               b.Ancestry(elem),
			CreatedAt:          createdAt,
			UpdatedAt:          updatedAt,
			BlockedBy:          blockedBy,
			Blocks:             blocks,
			Description:        rd.Description,
			AcceptanceCriteria: rd.AC,
			Checklist:          checklist,
			Notes:              notes,
		},
	}

	return output.PrintJSON(os.Stdout, response)
}
