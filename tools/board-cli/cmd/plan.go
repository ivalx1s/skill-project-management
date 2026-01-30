package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/aagrigore/task-board/internal/plan"
	"github.com/spf13/cobra"
)

var (
	planSave         bool
	planCriticalPath bool
	planPhase        int
	planRender       bool
	planFormat       string
	planLayout       string
	planActive       bool
)

var planCmd = &cobra.Command{
	Use:   "plan [ID]",
	Short: "Show execution plan with phases and critical path",
	Long: `Generate an execution plan showing phases (parallel groups) and critical path.

Without arguments, plans at the project level (epics).
With an EPIC ID, plans its stories.
With a STORY ID, plans its tasks/bugs.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPlan,
}

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.Flags().BoolVar(&planSave, "save", false, "Save plan as plan.md")
	planCmd.Flags().BoolVar(&planCriticalPath, "critical-path", false, "Show only critical path")
	planCmd.Flags().IntVar(&planPhase, "phase", 0, "Show only specific phase number")
	planCmd.Flags().BoolVar(&planRender, "render", false, "Render dependency graph via Graphviz")
	planCmd.Flags().StringVar(&planFormat, "format", "svg", "Render output format: svg, png, pdf")
	planCmd.Flags().StringVar(&planLayout, "layout", "hierarchy", "Graph layout: hierarchy (epic/story clusters) or phases (phase clusters)")
	planCmd.Flags().BoolVar(&planActive, "active", false, "Show only active elements (exclude done/closed)")
}

func runPlan(cmd *cobra.Command, args []string) error {
	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	scopeID := ""
	if len(args) > 0 {
		scopeID = args[0]
	}

	elements, err := plan.ScopeElements(b, scopeID)
	if err != nil {
		return err
	}

	p := plan.BuildPlan(elements)

	if p.HasCycle {
		fmt.Fprintf(os.Stderr, "Error: dependency cycle detected involving: %s\n", strings.Join(p.CycleNodes, ", "))
		os.Exit(1)
	}

	scopeName := scopeLabel(b, scopeID)

	if planRender {
		return renderGraph(b, scopeID, elements, p)
	}

	if planSave {
		return savePlanMD(b, scopeID, scopeName, p)
	}

	printPlan(scopeName, p)
	return nil
}

func scopeLabel(b *board.Board, scopeID string) string {
	if scopeID == "" {
		return "Project"
	}
	elem := b.FindByID(scopeID)
	if elem == nil {
		return scopeID
	}
	return fmt.Sprintf("%s: %s", elem.ID(), elem.Title)
}

func printPlan(scopeName string, p *plan.Plan) {
	// Critical path only mode
	if planCriticalPath {
		printCriticalPath(p)
		return
	}

	fmt.Printf("%sPlan: %s%s\n\n", output.Bold, scopeName, output.Reset)

	for _, phase := range p.Phases {
		if planPhase > 0 && phase.Number != planPhase {
			continue
		}

		label := fmt.Sprintf("Phase %d", phase.Number)
		if phase.Number == 1 {
			label += " (no dependencies)"
		}
		fmt.Printf("%s%s%s\n", output.Bold, label, output.Reset)

		table := output.NewTable("ID", "STATUS", "NAME", "BLOCKED BY")
		for _, e := range phase.Elements {
			blockedBy := ""
			if len(e.BlockedBy) > 0 {
				blockedBy = strings.Join(e.BlockedBy, ", ")
			}
			table.AddRow(
				e.ID(),
				output.ColorStatus(string(e.Status)),
				e.Name,
				blockedBy,
			)
		}
		fmt.Print(table.String())
		fmt.Println()
	}

	if planPhase == 0 {
		printCriticalPath(p)
	}
}

func printCriticalPath(p *plan.Plan) {
	if len(p.CriticalPath) == 0 {
		fmt.Println("No critical path (no dependencies).")
		return
	}

	fmt.Printf("%sCritical Path%s\n", output.Bold, output.Reset)
	var ids []string
	for _, e := range p.CriticalPath {
		ids = append(ids, e.ID())
	}
	fmt.Printf("  %s (%d phases)\n", strings.Join(ids, " -> "), len(p.Phases))
}

func savePlanMD(b *board.Board, scopeID, scopeName string, p *plan.Plan) error {
	mdPath, err := planMDPath(b, scopeID)
	if err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Plan: %s\n\n", scopeName))
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02")))

	for _, phase := range p.Phases {
		label := fmt.Sprintf("## Phase %d", phase.Number)
		if phase.Number == 1 {
			label += " (no dependencies)"
		}
		sb.WriteString(label + "\n")

		for _, e := range phase.Elements {
			line := fmt.Sprintf("- %s: %s", e.ID(), e.Title)
			if len(e.BlockedBy) > 0 {
				line += fmt.Sprintf(" (blocked by: %s)", strings.Join(e.BlockedBy, ", "))
			}
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Critical Path\n")
	if len(p.CriticalPath) == 0 {
		sb.WriteString("No critical path (no dependencies).\n\n")
	} else {
		var ids []string
		for _, e := range p.CriticalPath {
			ids = append(ids, e.ID())
		}
		sb.WriteString(fmt.Sprintf("%s (%d phases)\n\n", strings.Join(ids, " -> "), len(p.Phases)))
	}

	sb.WriteString("## Warnings\n")
	sb.WriteString("- No issues found\n")

	if err := os.MkdirAll(filepath.Dir(mdPath), 0755); err != nil {
		return fmt.Errorf("creating directory for plan.md: %w", err)
	}

	if err := os.WriteFile(mdPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("writing plan.md: %w", err)
	}

	fmt.Printf("Plan saved to %s\n", mdPath)
	return nil
}

func filterActiveElements(elements []*board.Element) []*board.Element {
	var result []*board.Element
	for _, e := range elements {
		if e.Status != board.StatusDone && e.Status != board.StatusClosed {
			result = append(result, e)
		}
	}
	return result
}

func renderGraph(b *board.Board, scopeID string, elements []*board.Element, p *plan.Plan) error {
	var dot string

	switch planLayout {
	case "phases":
		// Collect all descendants, build plan from them, render with phase clusters.
		allElements, err := plan.AllDescendants(b, scopeID)
		if err != nil {
			return err
		}
		// Exclude the root element itself (we want its children in the graph).
		if scopeID != "" {
			var filtered []*board.Element
			for _, e := range allElements {
				if e.ID() != strings.ToUpper(scopeID) {
					filtered = append(filtered, e)
				}
			}
			allElements = filtered
		}
		if planActive {
			allElements = filterActiveElements(allElements)
		}
		fullPlan := plan.BuildPlan(allElements)
		if fullPlan.HasCycle {
			return fmt.Errorf("dependency cycle detected involving: %s", strings.Join(fullPlan.CycleNodes, ", "))
		}
		dot = plan.GenerateDOT(fullPlan, allElements)
	default:
		// Hierarchy layout: full tree with epic/story clusters.
		allElements, err := plan.AllDescendants(b, scopeID)
		if err != nil {
			return err
		}
		if planActive {
			allElements = filterActiveElements(allElements)
		}
		dot = plan.GenerateFullDOT(b, allElements)
	}

	suffix := "plan"
	if planLayout == "phases" {
		suffix = "plan-phases"
	}
	if planActive {
		suffix += "-active"
	}
	outputPath, err := plan.RenderOutputPath(b, scopeID, suffix, planFormat)
	if err != nil {
		return err
	}

	if err := plan.RenderDOT(dot, outputPath, planFormat); err != nil {
		return err
	}

	fmt.Printf("Graph rendered to %s\n", outputPath)
	return nil
}

func planMDPath(b *board.Board, scopeID string) (string, error) {
	if scopeID == "" {
		// Project level: .task-board/plan.md
		return filepath.Join(boardDir, "plan.md"), nil
	}

	elem := b.FindByID(scopeID)
	if elem == nil {
		return "", fmt.Errorf("element %s not found", scopeID)
	}

	// Element's own directory contains the plan.md
	return filepath.Join(elem.Path, "plan.md"), nil
}
