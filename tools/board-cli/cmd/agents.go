package cmd

import (
	"fmt"
	"time"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

var agentsAll   bool
var agentsStale int

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Show sub-agent dashboard",
	Args:  cobra.NoArgs,
	RunE:  runAgents,
}

func init() {
	rootCmd.AddCommand(agentsCmd)
	agentsCmd.Flags().BoolVar(&agentsAll, "all", false, "Show all assigned elements (including done/closed)")
	agentsCmd.Flags().IntVar(&agentsStale, "stale", 30, "Freshness window in minutes (done entries older than this are hidden)")
}

// humanTime formats a time.Time as a human-readable relative string.
func humanTime(t time.Time, now time.Time) string {
	if t.IsZero() {
		return "-"
	}
	d := now.Sub(t)
	if d < 0 {
		d = 0
	}
	switch {
	case d < 60*time.Second:
		return fmt.Sprintf("%d sec ago", int(d.Seconds()))
	case d < 60*time.Minute:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	}
}

// childProgress returns "N/M done" for an element's children, or the element's own status if no children.
func childProgress(b *board.Board, e *board.Element) string {
	children := b.Children(e)
	if len(children) == 0 {
		return string(e.Status)
	}
	done := 0
	for _, c := range children {
		if c.Status == board.StatusDone || c.Status == board.StatusClosed {
			done++
		}
	}
	return fmt.Sprintf("%d/%d done", done, len(children))
}

func runAgents(cmd *cobra.Command, args []string) error {
	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	now := time.Now().UTC()
	freshness := time.Duration(agentsStale) * time.Minute

	// Collect assigned elements
	var assigned []*board.Element
	for _, e := range b.Elements {
		if e.AssignedTo == "" {
			continue
		}
		if !agentsAll {
			// Filter: show if not done/closed, OR if updated within freshness window
			isDone := e.Status == board.StatusDone || e.Status == board.StatusClosed
			isFresh := !e.LastUpdate.IsZero() && now.Sub(e.LastUpdate) <= freshness
			if isDone && !isFresh {
				continue
			}
		}
		assigned = append(assigned, e)
	}

	if len(assigned) == 0 {
		fmt.Println("No active agents.")
		return nil
	}

	fmt.Println(output.Bold + "Sub-Agent Dashboard" + output.Reset)
	fmt.Println()

	table := output.NewTable("AGENT", "SCOPE", "STATUS", "PROGRESS", "LAST UPDATE")
	for _, e := range assigned {
		scope := fmt.Sprintf("%s: %s", e.ID(), e.Name)
		table.AddRow(
			e.AssignedTo,
			scope,
			output.ColorStatus(string(e.Status)),
			childProgress(b, e),
			humanTime(e.LastUpdate, now),
		)
	}
	fmt.Print(table.String())

	// Footer
	active := 0
	done := 0
	for _, e := range assigned {
		if e.Status == board.StatusDone || e.Status == board.StatusClosed {
			done++
		} else {
			active++
		}
	}
	fmt.Printf("\nTotal: %d agents, %d active, %d done\n", len(assigned), active, done)

	return nil
}
