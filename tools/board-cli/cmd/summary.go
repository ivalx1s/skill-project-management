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
	type stats struct {
		total    int
		open     int
		progress int
		done     int
		closed   int
		blocked  int
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
		case board.StatusOpen:
			s.open++
		case board.StatusProgress:
			s.progress++
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

	table := output.NewTable("TYPE", "TOTAL", "OPEN", "PROGRESS", "DONE", "CLOSED", "BLOCKED")
	for _, t := range []board.ElementType{board.EpicType, board.StoryType, board.TaskType, board.BugType} {
		s := counts[t]
		if s.total == 0 {
			continue
		}
		table.AddRow(
			string(t),
			fmt.Sprintf("%d", s.total),
			fmt.Sprintf("%d", s.open),
			output.Yellow+fmt.Sprintf("%d", s.progress)+output.Reset,
			output.Green+fmt.Sprintf("%d", s.done)+output.Reset,
			fmt.Sprintf("%d", s.closed),
			output.Red+fmt.Sprintf("%d", s.blocked)+output.Reset,
		)
	}
	fmt.Print(table.String())

	// In progress
	var inProgress []*board.Element
	for _, e := range b.Elements {
		if e.Status == board.StatusProgress {
			inProgress = append(inProgress, e)
		}
	}
	if len(inProgress) > 0 {
		fmt.Println()
		fmt.Println(output.Bold + "In Progress" + output.Reset)
		for _, e := range inProgress {
			fmt.Printf("  %s %s (%s)\n", b.Ancestry(e), output.Gray+"—"+output.Reset, e.Name)
		}
	}

	// Blocked
	var blocked []*board.Element
	for _, e := range b.Elements {
		if e.Status == board.StatusBlocked || len(e.BlockedBy) > 0 {
			blocked = append(blocked, e)
		}
	}
	if len(blocked) > 0 {
		fmt.Println()
		fmt.Println(output.Bold + "Blocked" + output.Reset)
		for _, e := range blocked {
			blockers := strings.Join(e.BlockedBy, ", ")
			if blockers == "" {
				blockers = "(no blockers listed)"
			}
			fmt.Printf("  %s %s blocked by: %s\n", b.Ancestry(e), output.Gray+"—"+output.Reset, blockers)
		}
	}

	return nil
}
