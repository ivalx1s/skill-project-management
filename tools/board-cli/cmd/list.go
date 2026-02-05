package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <epics|stories|tasks|bugs>",
	Short: "List board elements",
	Args:  cobra.ExactArgs(1),
	RunE:  runList,
}

var (
	listStatus string
	listEpic   string
	listStory  string
)

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status")
	listCmd.Flags().StringVar(&listEpic, "epic", "", "Filter by epic ID")
	listCmd.Flags().StringVar(&listStory, "story", "", "Filter by story ID")
}

// ListResponse is the JSON response structure for list command
type ListResponse struct {
	Elements []ListElement `json:"elements"`
	Count    int           `json:"count"`
	Filters  ListFilters   `json:"filters"`
}

// ListElement represents an element in the list JSON output
type ListElement struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"`
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	Assignee  string   `json:"assignee"`
	Parent    string   `json:"parent"`
	Path      string   `json:"path"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
	BlockedBy []string `json:"blockedBy"`
	Blocks    []string `json:"blocks"`
}

// ListFilters shows which filters were applied
type ListFilters struct {
	Type   string `json:"type"`
	Story  string `json:"story,omitempty"`
	Epic   string `json:"epic,omitempty"`
	Status string `json:"status,omitempty"`
}

func runList(cmd *cobra.Command, args []string) error {
	elemType, err := board.ParseElementType(args[0])
	if err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.ValidationError, err.Error(), nil)
		}
		return err
	}

	b, err := board.Load(boardDir)
	if err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("loading board: %v", err), nil)
		}
		return fmt.Errorf("loading board: %w", err)
	}

	elements := b.FindByType(elemType)

	// Apply filters
	if listStatus != "" {
		status, err := board.ParseStatus(listStatus)
		if err != nil {
			if JSONEnabled() {
				output.PrintError(os.Stderr, output.InvalidStatus, err.Error(), nil)
			}
			return err
		}
		elements = board.FilterByStatus(elements, status)
	}

	if listEpic != "" {
		elements = board.FilterByParent(elements, listEpic)
	}

	if listStory != "" {
		elements = board.FilterByParent(elements, listStory)
	}

	// JSON output
	if JSONEnabled() {
		return printListJSON(b, elements, string(elemType))
	}

	// Table output
	if len(elements) == 0 {
		fmt.Println("No elements found.")
		return nil
	}

	table := output.NewTable("ID", "STATUS", "NAME", "PARENT")
	for _, e := range elements {
		statusStr := output.ColorStatus(string(e.Status))
		// Show [BLOCKED by X] if element has active blockers
		if activeBlockers := b.ActiveBlockers(e); len(activeBlockers) > 0 {
			var blockerIDs []string
			for _, blocker := range activeBlockers {
				blockerIDs = append(blockerIDs, blocker.ID())
			}
			statusStr += fmt.Sprintf(" %s[BLOCKED by %s]%s", output.Red, blockerIDs[0], output.Reset)
			if len(blockerIDs) > 1 {
				statusStr = statusStr[:len(statusStr)-len(output.Reset)] + fmt.Sprintf("+%d%s", len(blockerIDs)-1, output.Reset)
			}
		}
		table.AddRow(
			e.ID(),
			statusStr,
			e.Name,
			e.ParentID,
		)
	}

	fmt.Print(table.String())
	return nil
}

func printListJSON(b *board.Board, elements []*board.Element, elemType string) error {
	listElements := make([]ListElement, 0, len(elements))

	for _, e := range elements {
		// Build relative path from element's absolute path
		relPath := ""
		if e.Path != "" {
			// Extract just the path within the board (after .task-board/)
			if idx := filepath.Base(filepath.Dir(e.Path)); idx != "." {
				relPath = buildElementPath(e, b)
			}
		}

		// Format timestamps
		createdAt := ""
		if !e.CreatedAt.IsZero() {
			createdAt = e.CreatedAt.Format("2006-01-02T15:04:05Z")
		}
		updatedAt := ""
		if !e.LastUpdate.IsZero() {
			updatedAt = e.LastUpdate.Format("2006-01-02T15:04:05Z")
		}

		// Ensure slices are not nil for JSON
		blockedBy := e.BlockedBy
		if blockedBy == nil {
			blockedBy = []string{}
		}
		blocks := e.Blocks
		if blocks == nil {
			blocks = []string{}
		}

		listElements = append(listElements, ListElement{
			ID:        e.ID(),
			Type:      string(e.Type),
			Name:      e.Name,
			Status:    string(e.Status),
			Assignee:  e.AssignedTo,
			Parent:    e.ParentID,
			Path:      relPath,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			BlockedBy: blockedBy,
			Blocks:    blocks,
		})
	}

	response := ListResponse{
		Elements: listElements,
		Count:    len(listElements),
		Filters: ListFilters{
			Type:   elemType,
			Story:  listStory,
			Epic:   listEpic,
			Status: listStatus,
		},
	}

	return output.PrintJSON(os.Stdout, response)
}

// buildElementPath constructs the path like "EPIC-260205-foo/STORY-260205-xyz/TASK-260205-abc"
func buildElementPath(e *board.Element, b *board.Board) string {
	parts := []string{e.DirName()}

	// Walk up the parent chain
	currentParentID := e.ParentID
	for currentParentID != "" {
		parent := b.FindByID(currentParentID)
		if parent == nil {
			break
		}
		parts = append([]string{parent.DirName()}, parts...)
		currentParentID = parent.ParentID
	}

	return filepath.Join(parts...)
}
