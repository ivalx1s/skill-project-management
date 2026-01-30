package plan

import (
	"fmt"
	"strings"

	"github.com/aagrigore/task-board/internal/board"
)

// statusColor returns the fillcolor for a given element status.
func statusColor(s board.Status) string {
	switch s {
	case board.StatusOpen:
		return "#ffffff"
	case board.StatusProgress:
		return "#fff3cd"
	case board.StatusDone:
		return "#d4edda"
	case board.StatusClosed:
		return "#e2e3e5"
	case board.StatusBlocked:
		return "#f8d7da"
	default:
		return "#ffffff"
	}
}

// safeDOTID converts an element ID like "TASK-01" to a DOT-safe identifier "TASK_01".
func safeDOTID(id string) string {
	return strings.ReplaceAll(id, "-", "_")
}

// GenerateDOT produces a Graphviz DOT representation of the plan.
// The plan provides phase grouping; elements provides the full element data
// (for status colors, names, and blocked-by edges).
func GenerateDOT(p *Plan, elements []*board.Element) string {
	var b strings.Builder

	b.WriteString("digraph plan {\n")
	b.WriteString("  rankdir=LR;\n")
	b.WriteString("  node [shape=box, style=filled, fontname=\"Helvetica\"];\n")

	// Index elements by ID for quick lookup.
	inScope := make(map[string]bool, len(elements))
	for _, e := range elements {
		inScope[e.ID()] = true
	}

	// Subgraph per phase.
	for _, phase := range p.Phases {
		b.WriteString(fmt.Sprintf("\n  subgraph cluster_phase_%d {\n", phase.Number))
		b.WriteString(fmt.Sprintf("    label=\"Phase %d\";\n", phase.Number))
		b.WriteString("    style=dashed;\n")

		for _, e := range phase.Elements {
			id := safeDOTID(e.ID())
			color := statusColor(e.Status)
			label := fmt.Sprintf("%s\\n%s", e.ID(), e.Name)
			b.WriteString(fmt.Sprintf("    %s [label=\"%s\", fillcolor=\"%s\"];\n", id, label, color))
		}

		b.WriteString("  }\n")
	}

	// Edges: for each element, draw from each blocker to this element.
	b.WriteString("\n")
	for _, e := range elements {
		for _, blockerID := range e.BlockedBy {
			if !inScope[blockerID] {
				continue
			}
			b.WriteString(fmt.Sprintf("  %s -> %s;\n", safeDOTID(blockerID), safeDOTID(e.ID())))
		}
	}

	writeLegend(&b)

	b.WriteString("}\n")
	return b.String()
}

// GenerateFullDOT produces a Graphviz DOT with the full hierarchy rendered
// on a single graph. Elements are clustered by their parent (epics contain
// stories, stories contain tasks/bugs).
func GenerateFullDOT(brd *board.Board, elements []*board.Element) string {
	var b strings.Builder

	b.WriteString("digraph plan {\n")
	b.WriteString("  rankdir=TB;\n")
	b.WriteString("  compound=true;\n")
	b.WriteString("  node [shape=box, style=filled, fontname=\"Helvetica\", fontsize=11];\n")
	b.WriteString("  edge [color=\"#666666\"];\n")

	// Index all elements in scope.
	inScope := make(map[string]bool, len(elements))
	for _, e := range elements {
		inScope[e.ID()] = true
	}

	// Group elements by type and parent.
	epics := filterType(elements, board.EpicType)
	storiesByEpic := make(map[string][]*board.Element)
	tasksByStory := make(map[string][]*board.Element)
	// Loose stories/tasks (parent not in scope).
	var looseStories []*board.Element
	var looseTasks []*board.Element

	for _, e := range elements {
		switch e.Type {
		case board.StoryType:
			if inScope[e.ParentID] {
				storiesByEpic[e.ParentID] = append(storiesByEpic[e.ParentID], e)
			} else {
				looseStories = append(looseStories, e)
			}
		case board.TaskType, board.BugType:
			if inScope[e.ParentID] {
				tasksByStory[e.ParentID] = append(tasksByStory[e.ParentID], e)
			} else {
				looseTasks = append(looseTasks, e)
			}
		}
	}

	clusterIdx := 0

	// Render epics as top-level clusters.
	for _, epic := range epics {
		clusterIdx++
		b.WriteString(fmt.Sprintf("\n  subgraph cluster_%d {\n", clusterIdx))
		b.WriteString(fmt.Sprintf("    label=\"%s: %s\";\n", epic.ID(), epic.Name))
		b.WriteString("    style=rounded;\n")
		b.WriteString(fmt.Sprintf("    color=\"%s\";\n", clusterBorderColor(epic.Status)))
		b.WriteString("    fontname=\"Helvetica Bold\";\n")
		b.WriteString("    fontsize=13;\n")

		// Epic node itself.
		writeNode(&b, epic, "    ")

		// Stories inside this epic.
		stories := storiesByEpic[epic.ID()]
		for _, story := range stories {
			tasks := tasksByStory[story.ID()]
			if len(tasks) > 0 {
				clusterIdx++
				b.WriteString(fmt.Sprintf("\n    subgraph cluster_%d {\n", clusterIdx))
				b.WriteString(fmt.Sprintf("      label=\"%s: %s\";\n", story.ID(), story.Name))
				b.WriteString("      style=dashed;\n")
				b.WriteString(fmt.Sprintf("      color=\"%s\";\n", clusterBorderColor(story.Status)))
				b.WriteString("      fontname=\"Helvetica\";\n")
				b.WriteString("      fontsize=11;\n")

				writeNode(&b, story, "      ")
				for _, t := range tasks {
					writeNode(&b, t, "      ")
				}
				b.WriteString("    }\n")
			} else {
				writeNode(&b, story, "    ")
			}
		}

		b.WriteString("  }\n")
	}

	// Loose stories (when rendering from a story scope, these are direct children).
	for _, story := range looseStories {
		tasks := tasksByStory[story.ID()]
		if len(tasks) > 0 {
			clusterIdx++
			b.WriteString(fmt.Sprintf("\n  subgraph cluster_%d {\n", clusterIdx))
			b.WriteString(fmt.Sprintf("    label=\"%s: %s\";\n", story.ID(), story.Name))
			b.WriteString("    style=dashed;\n")
			b.WriteString(fmt.Sprintf("    color=\"%s\";\n", clusterBorderColor(story.Status)))
			b.WriteString("    fontname=\"Helvetica\";\n")

			writeNode(&b, story, "    ")
			for _, t := range tasks {
				writeNode(&b, t, "    ")
			}
			b.WriteString("  }\n")
		} else {
			writeNode(&b, story, "  ")
		}
	}

	// Loose tasks.
	for _, t := range looseTasks {
		writeNode(&b, t, "  ")
	}

	// Edges: all blocked-by within scope.
	b.WriteString("\n")
	for _, e := range elements {
		for _, blockerID := range e.BlockedBy {
			if !inScope[blockerID] {
				continue
			}
			b.WriteString(fmt.Sprintf("  %s -> %s;\n", safeDOTID(blockerID), safeDOTID(e.ID())))
		}
	}

	writeLegend(&b)

	b.WriteString("}\n")
	return b.String()
}

func writeNode(b *strings.Builder, e *board.Element, indent string) {
	id := safeDOTID(e.ID())
	color := statusColor(e.Status)
	shape := "box"
	if e.Type == board.EpicType {
		shape = "box3d"
	} else if e.Type == board.BugType {
		shape = "octagon"
	}
	label := fmt.Sprintf("%s\\n%s", e.ID(), e.Name)
	b.WriteString(fmt.Sprintf("%s%s [label=\"%s\", fillcolor=\"%s\", shape=%s];\n", indent, id, label, color, shape))
}

func writeLegend(b *strings.Builder) {
	b.WriteString("\n  subgraph cluster_legend {\n")
	b.WriteString("    label=\"Legend\";\n")
	b.WriteString("    style=rounded;\n")
	b.WriteString("    color=\"#999999\";\n")
	b.WriteString("    fontname=\"Helvetica Bold\";\n")
	b.WriteString("    fontsize=11;\n")
	b.WriteString("    node [shape=box, style=filled, fontname=\"Helvetica\", fontsize=9, width=1.2];\n")
	b.WriteString("    leg_open [label=\"open\", fillcolor=\"#ffffff\"];\n")
	b.WriteString("    leg_progress [label=\"progress\", fillcolor=\"#fff3cd\"];\n")
	b.WriteString("    leg_done [label=\"done\", fillcolor=\"#d4edda\"];\n")
	b.WriteString("    leg_blocked [label=\"blocked\", fillcolor=\"#f8d7da\"];\n")
	b.WriteString("    leg_closed [label=\"closed\", fillcolor=\"#e2e3e5\"];\n")
	b.WriteString("    leg_open -> leg_progress -> leg_done [style=invis];\n")
	b.WriteString("    leg_blocked -> leg_closed [style=invis];\n")
	b.WriteString("  }\n")
}

func clusterBorderColor(s board.Status) string {
	switch s {
	case board.StatusDone:
		return "#28a745"
	case board.StatusProgress:
		return "#ffc107"
	case board.StatusBlocked:
		return "#dc3545"
	default:
		return "#6c757d"
	}
}

func filterType(elements []*board.Element, t board.ElementType) []*board.Element {
	var result []*board.Element
	for _, e := range elements {
		if e.Type == t {
			result = append(result, e)
		}
	}
	return result
}
