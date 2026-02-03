package cmd

import (
	"fmt"
	"strings"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

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
		return fmt.Errorf("loading board: %w", err)
	}

	if len(b.Elements) == 0 {
		fmt.Println("Board is empty.")
		return nil
	}

	// Count by type and status
	// Group: TODO (backlog, to-dev), ACTIVE (analysis, development, to-review, reviewing), DONE, CLOSED, BLOCKED
	type stats struct {
		total   int
		todo    int // backlog + to-dev
		active  int // analysis + development + to-review + reviewing
		done    int
		closed  int
		blocked int
	}

	counts := map[board.ElementType]*stats{
		board.EpicType:  {},
		board.StoryType: {},
		board.TaskType:  {},
		board.BugType:   {},
	}

	for _, e := range b.Elements {
		s := counts[e.Type]
		s.total++
		switch e.Status {
		case board.StatusBacklog, board.StatusToDev:
			s.todo++
		case board.StatusAnalysis, board.StatusDevelopment, board.StatusToReview, board.StatusReviewing:
			s.active++
		case board.StatusDone:
			s.done++
		case board.StatusClosed:
			s.closed++
		case board.StatusBlocked:
			s.blocked++
		}
	}

	// Summary table
	fmt.Println(output.Bold + "Board Summary" + output.Reset)
	fmt.Println()

	table := output.NewTable("TYPE", "TOTAL", "TODO", "ACTIVE", "DONE", "CLOSED", "BLOCKED")
	for _, t := range []board.ElementType{board.EpicType, board.StoryType, board.TaskType, board.BugType} {
		s := counts[t]
		if s.total == 0 {
			continue
		}
		table.AddRow(
			string(t),
			fmt.Sprintf("%d", s.total),
			fmt.Sprintf("%d", s.todo),
			output.Yellow+fmt.Sprintf("%d", s.active)+output.Reset,
			output.Green+fmt.Sprintf("%d", s.done)+output.Reset,
			fmt.Sprintf("%d", s.closed),
			output.Red+fmt.Sprintf("%d", s.blocked)+output.Reset,
		)
	}
	fmt.Print(table.String())

	// Active (analysis, development, to-review, reviewing)
	var active []*board.Element
	for _, e := range b.Elements {
		switch e.Status {
		case board.StatusAnalysis, board.StatusDevelopment, board.StatusToReview, board.StatusReviewing:
			active = append(active, e)
		}
	}
	if len(active) > 0 {
		fmt.Println()
		fmt.Println(output.Bold + "Active" + output.Reset)
		for _, e := range active {
			fmt.Printf("  %s %s (%s) [%s]\n", b.Ancestry(e), output.Gray+"—"+output.Reset, e.Name, e.Status)
		}
	}

	// Blocked (explicit status or has active blockers)
	var blocked []*board.Element
	for _, e := range b.Elements {
		if e.Status == board.StatusBlocked || b.IsBlocked(e) {
			blocked = append(blocked, e)
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
