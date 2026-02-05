package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

// SummaryResponse is the JSON response structure for summary command
type SummaryResponse struct {
	Summary SummaryData `json:"summary"`
}

// SummaryData contains the summary information
type SummaryData struct {
	ByType  map[string]TypeStats    `json:"byType"`
	Active  []SummaryActiveElement  `json:"active"`
	Blocked []SummaryBlockedElement `json:"blocked"`
}

// TypeStats contains counts by status group for a type
type TypeStats struct {
	Total   int `json:"total"`
	Todo    int `json:"todo"`
	Active  int `json:"active"`
	Done    int `json:"done"`
	Closed  int `json:"closed"`
	Blocked int `json:"blocked"`
}

// SummaryActiveElement represents an active element in JSON output
type SummaryActiveElement struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Ancestry string `json:"ancestry"`
}

// SummaryBlockedElement represents a blocked element in JSON output
type SummaryBlockedElement struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	BlockedBy []string `json:"blockedBy"`
	Ancestry  string   `json:"ancestry"`
}

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show board summary",
	RunE:  runSummary,
}

func init() {
	rootCmd.AddCommand(summaryCmd)
}

func runSummary(cmd *cobra.Command, args []string) error {
	b, err := board.Load(boardDir)
	if err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("loading board: %v", err), nil)
		}
		return fmt.Errorf("loading board: %w", err)
	}

	// Count by type and status
	// Group: TODO (backlog, to-dev), ACTIVE (analysis, development, to-review, reviewing), DONE, CLOSED, BLOCKED
	counts := map[board.ElementType]*TypeStats{
		board.EpicType:  {},
		board.StoryType: {},
		board.TaskType:  {},
		board.BugType:   {},
	}

	for _, e := range b.Elements {
		s := counts[e.Type]
		s.Total++
		switch e.Status {
		case board.StatusBacklog, board.StatusToDev:
			s.Todo++
		case board.StatusAnalysis, board.StatusDevelopment, board.StatusToReview, board.StatusReviewing:
			s.Active++
		case board.StatusDone:
			s.Done++
		case board.StatusClosed:
			s.Closed++
		case board.StatusBlocked:
			s.Blocked++
		}
	}

	// Active (analysis, development, to-review, reviewing)
	var active []*board.Element
	for _, e := range b.Elements {
		switch e.Status {
		case board.StatusAnalysis, board.StatusDevelopment, board.StatusToReview, board.StatusReviewing:
			active = append(active, e)
		}
	}

	// Blocked (explicit status or has active blockers)
	var blocked []*board.Element
	for _, e := range b.Elements {
		if e.Status == board.StatusBlocked || b.IsBlocked(e) {
			blocked = append(blocked, e)
		}
	}

	// JSON output
	if JSONEnabled() {
		return printSummaryJSON(b, counts, active, blocked)
	}

	// Text output
	if len(b.Elements) == 0 {
		fmt.Println("Board is empty.")
		return nil
	}

	fmt.Println(output.Bold + "Board Summary" + output.Reset)
	fmt.Println()

	table := output.NewTable("TYPE", "TOTAL", "TODO", "ACTIVE", "DONE", "CLOSED", "BLOCKED")
	for _, t := range []board.ElementType{board.EpicType, board.StoryType, board.TaskType, board.BugType} {
		s := counts[t]
		if s.Total == 0 {
			continue
		}
		table.AddRow(
			string(t),
			fmt.Sprintf("%d", s.Total),
			fmt.Sprintf("%d", s.Todo),
			output.Yellow+fmt.Sprintf("%d", s.Active)+output.Reset,
			output.Green+fmt.Sprintf("%d", s.Done)+output.Reset,
			fmt.Sprintf("%d", s.Closed),
			output.Red+fmt.Sprintf("%d", s.Blocked)+output.Reset,
		)
	}
	fmt.Print(table.String())

	if len(active) > 0 {
		fmt.Println()
		fmt.Println(output.Bold + "Active" + output.Reset)
		for _, e := range active {
			fmt.Printf("  %s %s (%s) [%s]\n", b.Ancestry(e), output.Gray+"—"+output.Reset, e.Name, e.Status)
		}
	}

	if len(blocked) > 0 {
		fmt.Println()
		fmt.Println(output.Bold + "Blocked" + output.Reset)
		for _, e := range blocked {
			if e.Status == board.StatusBlocked {
				fmt.Printf("  %s %s blocked (external)\n", b.Ancestry(e), output.Gray+"—"+output.Reset)
			} else {
				activeBlockers := b.ActiveBlockers(e)
				var blockerIDs []string
				for _, blocker := range activeBlockers {
					blockerIDs = append(blockerIDs, blocker.ID())
				}
				fmt.Printf("  %s %s blocked by: %s\n", b.Ancestry(e), output.Gray+"—"+output.Reset, strings.Join(blockerIDs, ", "))
			}
		}
	}

	return nil
}

// printSummaryJSON outputs summary data as JSON
func printSummaryJSON(b *board.Board, counts map[board.ElementType]*TypeStats, active []*board.Element, blocked []*board.Element) error {
	// Build byType map
	byType := map[string]TypeStats{
		"epic":  {Total: 0, Todo: 0, Active: 0, Done: 0, Closed: 0, Blocked: 0},
		"story": {Total: 0, Todo: 0, Active: 0, Done: 0, Closed: 0, Blocked: 0},
		"task":  {Total: 0, Todo: 0, Active: 0, Done: 0, Closed: 0, Blocked: 0},
		"bug":   {Total: 0, Todo: 0, Active: 0, Done: 0, Closed: 0, Blocked: 0},
	}

	for elemType, s := range counts {
		byType[string(elemType)] = TypeStats{
			Total:   s.Total,
			Todo:    s.Todo,
			Active:  s.Active,
			Done:    s.Done,
			Closed:  s.Closed,
			Blocked: s.Blocked,
		}
	}

	// Build active list
	activeList := make([]SummaryActiveElement, 0, len(active))
	for _, e := range active {
		activeList = append(activeList, SummaryActiveElement{
			ID:       e.ID(),
			Name:     e.Name,
			Status:   string(e.Status),
			Ancestry: b.Ancestry(e),
		})
	}

	// Build blocked list
	blockedList := make([]SummaryBlockedElement, 0, len(blocked))
	for _, e := range blocked {
		blockerIDs := []string{}
		if e.Status == board.StatusBlocked {
			blockerIDs = []string{"external"}
		} else {
			activeBlockers := b.ActiveBlockers(e)
			for _, blocker := range activeBlockers {
				blockerIDs = append(blockerIDs, blocker.ID())
			}
		}
		blockedList = append(blockedList, SummaryBlockedElement{
			ID:        e.ID(),
			Name:      e.Name,
			BlockedBy: blockerIDs,
			Ancestry:  b.Ancestry(e),
		})
	}

	response := SummaryResponse{
		Summary: SummaryData{
			ByType:  byType,
			Active:  activeList,
			Blocked: blockedList,
		},
	}

	return output.PrintJSON(os.Stdout, response)
}
