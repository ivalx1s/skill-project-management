package plan

import (
	"strings"
	"testing"

	"github.com/aagrigore/task-board/internal/board"
)

func makeElementWithStatus(typ board.ElementType, num int, name string, status board.Status, blockedBy ...string) *board.Element {
	return &board.Element{
		Type:      typ,
		Number:    num,
		Name:      name,
		Status:    status,
		BlockedBy: blockedBy,
	}
}

func TestGenerateDOTBasicStructure(t *testing.T) {
	a := makeElementWithStatus(board.TaskType, 1, "interface", board.StatusDone)
	b := makeElementWithStatus(board.TaskType, 2, "implementation", board.StatusProgress, "TASK-01")

	elements := []*board.Element{a, b}
	plan := BuildPlan(elements)
	dot := GenerateDOT(plan, elements)

	// Must contain digraph header.
	if !strings.Contains(dot, "digraph plan {") {
		t.Error("missing digraph header")
	}
	// Must contain rankdir.
	if !strings.Contains(dot, "rankdir=LR") {
		t.Error("missing rankdir=LR")
	}
	// Must contain node defaults.
	if !strings.Contains(dot, "node [shape=box, style=filled, fontname=\"Helvetica\"]") {
		t.Error("missing node defaults")
	}
	// Must contain closing brace.
	if !strings.HasSuffix(strings.TrimSpace(dot), "}") {
		t.Error("missing closing brace")
	}
}

func TestGenerateDOTPhaseCluster(t *testing.T) {
	a := makeElementWithStatus(board.TaskType, 1, "first", board.StatusOpen)
	b := makeElementWithStatus(board.TaskType, 2, "second", board.StatusOpen, "TASK-01")

	elements := []*board.Element{a, b}
	plan := BuildPlan(elements)
	dot := GenerateDOT(plan, elements)

	if !strings.Contains(dot, "subgraph cluster_phase_1") {
		t.Error("missing phase 1 cluster")
	}
	if !strings.Contains(dot, "subgraph cluster_phase_2") {
		t.Error("missing phase 2 cluster")
	}
	if !strings.Contains(dot, `label="Phase 1"`) {
		t.Error("missing Phase 1 label")
	}
	if !strings.Contains(dot, `label="Phase 2"`) {
		t.Error("missing Phase 2 label")
	}
}

func TestGenerateDOTNodeLabels(t *testing.T) {
	a := makeElementWithStatus(board.TaskType, 1, "interface", board.StatusDone)
	b := makeElementWithStatus(board.StoryType, 5, "auth", board.StatusProgress)

	elements := []*board.Element{a, b}
	plan := BuildPlan(elements)
	dot := GenerateDOT(plan, elements)

	// Node IDs must be safe (underscore, not dash).
	if !strings.Contains(dot, "TASK_01") {
		t.Error("missing safe node ID TASK_01")
	}
	if !strings.Contains(dot, "STORY_05") {
		t.Error("missing safe node ID STORY_05")
	}
	// Labels must contain original ID with dash.
	if !strings.Contains(dot, `TASK-01\ninterface`) {
		t.Error("missing label for TASK-01")
	}
	if !strings.Contains(dot, `STORY-05\nauth`) {
		t.Error("missing label for STORY-05")
	}
}

func TestGenerateDOTNodeColors(t *testing.T) {
	tests := []struct {
		status board.Status
		color  string
	}{
		{board.StatusOpen, "#ffffff"},
		{board.StatusProgress, "#fff3cd"},
		{board.StatusDone, "#d4edda"},
		{board.StatusClosed, "#e2e3e5"},
		{board.StatusBlocked, "#f8d7da"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			e := makeElementWithStatus(board.TaskType, 1, "test", tt.status)
			elements := []*board.Element{e}
			plan := BuildPlan(elements)
			dot := GenerateDOT(plan, elements)

			expected := `fillcolor="` + tt.color + `"`
			if !strings.Contains(dot, expected) {
				t.Errorf("status %s: expected %s in DOT output", tt.status, expected)
			}
		})
	}
}

func TestGenerateDOTEdges(t *testing.T) {
	a := makeElementWithStatus(board.TaskType, 1, "first", board.StatusDone)
	b := makeElementWithStatus(board.TaskType, 2, "second", board.StatusOpen, "TASK-01")
	c := makeElementWithStatus(board.TaskType, 3, "third", board.StatusOpen, "TASK-01", "TASK-02")

	elements := []*board.Element{a, b, c}
	plan := BuildPlan(elements)
	dot := GenerateDOT(plan, elements)

	if !strings.Contains(dot, "TASK_01 -> TASK_02") {
		t.Error("missing edge TASK_01 -> TASK_02")
	}
	if !strings.Contains(dot, "TASK_01 -> TASK_03") {
		t.Error("missing edge TASK_01 -> TASK_03")
	}
	if !strings.Contains(dot, "TASK_02 -> TASK_03") {
		t.Error("missing edge TASK_02 -> TASK_03")
	}
}

func TestGenerateDOTEdgesOutOfScopeIgnored(t *testing.T) {
	// b depends on TASK-99 which is not in scope — should not appear as edge.
	a := makeElementWithStatus(board.TaskType, 1, "first", board.StatusDone)
	b := makeElementWithStatus(board.TaskType, 2, "second", board.StatusOpen, "TASK-99", "TASK-01")

	elements := []*board.Element{a, b}
	plan := BuildPlan(elements)
	dot := GenerateDOT(plan, elements)

	if strings.Contains(dot, "TASK_99") {
		t.Error("out-of-scope TASK_99 should not appear in DOT output")
	}
	if !strings.Contains(dot, "TASK_01 -> TASK_02") {
		t.Error("missing in-scope edge TASK_01 -> TASK_02")
	}
}

func TestGenerateDOTEmpty(t *testing.T) {
	elements := []*board.Element{}
	plan := BuildPlan(elements)
	dot := GenerateDOT(plan, elements)

	if !strings.Contains(dot, "digraph plan {") {
		t.Error("empty plan should still produce valid digraph header")
	}
	if !strings.HasSuffix(strings.TrimSpace(dot), "}") {
		t.Error("empty plan should still have closing brace")
	}
	// Should not contain phase subgraphs (legend is ok).
	if strings.Contains(dot, "cluster_phase") {
		t.Error("empty plan should not have phase subgraphs")
	}
}

func TestSafeDOTID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"TASK-01", "TASK_01"},
		{"EPIC-12", "EPIC_12"},
		{"BUG-03", "BUG_03"},
		{"STORY-99", "STORY_99"},
		{"NO-DASH-HERE", "NO_DASH_HERE"},
	}

	for _, tt := range tests {
		got := safeDOTID(tt.input)
		if got != tt.want {
			t.Errorf("safeDOTID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestStatusColor(t *testing.T) {
	tests := []struct {
		status board.Status
		want   string
	}{
		{board.StatusOpen, "#ffffff"},
		{board.StatusProgress, "#fff3cd"},
		{board.StatusDone, "#d4edda"},
		{board.StatusClosed, "#e2e3e5"},
		{board.StatusBlocked, "#f8d7da"},
		{board.Status("unknown"), "#ffffff"},
	}

	for _, tt := range tests {
		got := statusColor(tt.status)
		if got != tt.want {
			t.Errorf("statusColor(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}
