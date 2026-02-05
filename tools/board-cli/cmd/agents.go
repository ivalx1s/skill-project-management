package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

// AgentsResponse is the JSON response structure for agents command
type AgentsResponse struct {
	Agents      []AgentInfo  `json:"agents"`
	TotalAgents int          `json:"totalAgents"`
	Filters     AgentFilters `json:"filters"`
}

// AgentInfo represents an agent with their assigned elements
type AgentInfo struct {
	Name             string                 `json:"name"`
	AssignedElements []AgentAssignedElement `json:"assignedElements"`
	TotalAssigned    int                    `json:"totalAssigned"`
	StaleCount       int                    `json:"staleCount"`
}

// AgentAssignedElement represents an element assigned to an agent
type AgentAssignedElement struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	UpdatedAt  string  `json:"updatedAt"`
	StaleSince *string `json:"staleSince"`
}

// AgentFilters shows which filters were applied
type AgentFilters struct {
	StaleMinutes int `json:"staleMinutes"`
}

var agentsAll bool
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

// childProgress returns progress for an element's children, or the element's own status if no children.
// Shows both to-review and done counts for visibility into agent work.
func childProgress(b *board.Board, e *board.Element) string {
	children := b.Children(e)
	if len(children) == 0 {
		return string(e.Status)
	}
	review := 0
	done := 0
	for _, c := range children {
		if c.Status == board.StatusToReview {
			review++
		}
		if c.Status == board.StatusDone || c.Status == board.StatusClosed {
			done++
		}
	}
	total := len(children)
	if review > 0 && done > 0 {
		return fmt.Sprintf("%d/%d to-review, %d/%d done", review, total, done, total)
	}
	if review > 0 {
		return fmt.Sprintf("%d/%d to-review", review, total)
	}
	if done > 0 {
		return fmt.Sprintf("%d/%d done", done, total)
	}
	return fmt.Sprintf("0/%d", total)
}

func runAgents(cmd *cobra.Command, args []string) error {
	b, err := board.Load(boardDir)
	if err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("loading board: %v", err), nil)
		}
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

	// JSON output
	if JSONEnabled() {
		return printAgentsJSON(assigned, now, freshness)
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

func printAgentsJSON(assigned []*board.Element, now time.Time, freshness time.Duration) error {
	// Group elements by agent name
	agentMap := make(map[string][]*board.Element)
	for _, e := range assigned {
		agentMap[e.AssignedTo] = append(agentMap[e.AssignedTo], e)
	}

	// Build response
	agents := make([]AgentInfo, 0, len(agentMap))
	for agentName, elements := range agentMap {
		assignedElements := make([]AgentAssignedElement, 0, len(elements))
		staleCount := 0

		for _, e := range elements {
			// Format updatedAt
			updatedAt := ""
			if !e.LastUpdate.IsZero() {
				updatedAt = e.LastUpdate.Format("2006-01-02T15:04:05Z")
			}

			// Determine staleSince - element is stale if not done/closed and updated longer than freshness ago
			var staleSince *string
			isDone := e.Status == board.StatusDone || e.Status == board.StatusClosed
			if !isDone && !e.LastUpdate.IsZero() && now.Sub(e.LastUpdate) > freshness {
				staleTime := e.LastUpdate.Add(freshness).Format("2006-01-02T15:04:05Z")
				staleSince = &staleTime
				staleCount++
			}

			assignedElements = append(assignedElements, AgentAssignedElement{
				ID:         e.ID(),
				Type:       string(e.Type),
				Name:       e.Name,
				Status:     string(e.Status),
				UpdatedAt:  updatedAt,
				StaleSince: staleSince,
			})
		}

		agents = append(agents, AgentInfo{
			Name:             agentName,
			AssignedElements: assignedElements,
			TotalAssigned:    len(elements),
			StaleCount:       staleCount,
		})
	}

	response := AgentsResponse{
		Agents:      agents,
		TotalAgents: len(agents),
		Filters: AgentFilters{
			StaleMinutes: int(freshness.Minutes()),
		},
	}

	return output.PrintJSON(os.Stdout, response)
}
