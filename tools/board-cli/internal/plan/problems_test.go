package plan

import (
	"strings"
	"testing"

	"github.com/aagrigore/task-board/internal/board"
)

func TestDetectProblems_NoCycle(t *testing.T) {
	a := makeElement(board.TaskType, 1, "a")
	b := makeElement(board.TaskType, 2, "b", "TASK-01")
	c := makeElement(board.TaskType, 3, "c", "TASK-02")
	elements := []*board.Element{a, b, c}

	plan := BuildPlan(elements)
	problems := DetectProblems(plan, elements)

	for _, p := range problems {
		if p.Severity == "error" {
			t.Errorf("unexpected error problem: %s", p.Message)
		}
	}
}

func TestDetectProblems_SimpleCycle(t *testing.T) {
	// A blocks B, B blocks A
	a := makeElement(board.TaskType, 1, "a", "TASK-02")
	b := makeElement(board.TaskType, 2, "b", "TASK-01")
	elements := []*board.Element{a, b}

	plan := BuildPlan(elements)
	problems := DetectProblems(plan, elements)

	found := false
	for _, p := range problems {
		if p.Severity == "error" && p.Type == "cycle" {
			found = true
			if !strings.Contains(p.Message, "cycle") {
				t.Errorf("cycle message should mention 'cycle': %s", p.Message)
			}
			if len(p.Elements) != 2 {
				t.Errorf("cycle elements = %d, want 2", len(p.Elements))
			}
		}
	}
	if !found {
		t.Fatal("expected cycle error problem, found none")
	}
}

func TestDetectProblems_ComplexCycle(t *testing.T) {
	// A -> B -> C -> A
	a := makeElement(board.TaskType, 1, "a", "TASK-03")
	b := makeElement(board.TaskType, 2, "b", "TASK-01")
	c := makeElement(board.TaskType, 3, "c", "TASK-02")
	elements := []*board.Element{a, b, c}

	plan := BuildPlan(elements)
	problems := DetectProblems(plan, elements)

	found := false
	for _, p := range problems {
		if p.Severity == "error" && p.Type == "cycle" {
			found = true
			if !strings.Contains(p.Message, "cycle") {
				t.Errorf("cycle message should mention 'cycle': %s", p.Message)
			}
			if len(p.Elements) != 3 {
				t.Errorf("cycle elements = %d, want 3", len(p.Elements))
			}
		}
	}
	if !found {
		t.Fatal("expected cycle error problem, found none")
	}
}

func TestDetectProblems_CriticalPathLength5(t *testing.T) {
	// Linear chain of 5 elements
	a := makeElement(board.TaskType, 1, "a")
	b := makeElement(board.TaskType, 2, "b", "TASK-01")
	c := makeElement(board.TaskType, 3, "c", "TASK-02")
	d := makeElement(board.TaskType, 4, "d", "TASK-03")
	e := makeElement(board.TaskType, 5, "e", "TASK-04")
	elements := []*board.Element{a, b, c, d, e}

	plan := BuildPlan(elements)

	if len(plan.CriticalPath) != 5 {
		t.Fatalf("critical path len = %d, want 5", len(plan.CriticalPath))
	}

	problems := DetectProblems(plan, elements)

	found := false
	for _, p := range problems {
		if p.Type == "critical-path" {
			found = true
			if p.Severity != "info" {
				t.Errorf("critical-path severity = %s, want info", p.Severity)
			}
			if len(p.Elements) != 5 {
				t.Errorf("critical-path elements = %d, want 5", len(p.Elements))
			}
		}
	}
	if !found {
		t.Fatal("expected critical-path info problem, found none")
	}
}

func TestDetectProblems_CriticalPathDiamond(t *testing.T) {
	// Diamond: A->B, A->C, B->D, C->D
	// Critical path = 3 (A -> B -> D or A -> C -> D)
	a := makeElement(board.TaskType, 1, "a")
	b := makeElement(board.TaskType, 2, "b", "TASK-01")
	c := makeElement(board.TaskType, 3, "c", "TASK-01")
	d := makeElement(board.TaskType, 4, "d", "TASK-02", "TASK-03")
	elements := []*board.Element{a, b, c, d}

	plan := BuildPlan(elements)

	if len(plan.CriticalPath) != 3 {
		t.Fatalf("critical path len = %d, want 3", len(plan.CriticalPath))
	}

	problems := DetectProblems(plan, elements)

	found := false
	for _, p := range problems {
		if p.Type == "critical-path" {
			found = true
			if len(p.Elements) != 3 {
				t.Errorf("critical-path elements = %d, want 3", len(p.Elements))
			}
			// Path starts with A and ends with D
			if p.Elements[0] != "TASK-01" {
				t.Errorf("critical-path start = %s, want TASK-01", p.Elements[0])
			}
			if p.Elements[2] != "TASK-04" {
				t.Errorf("critical-path end = %s, want TASK-04", p.Elements[2])
			}
		}
	}
	if !found {
		t.Fatal("expected critical-path info problem, found none")
	}
}

func TestDetectProblems_SingleElement(t *testing.T) {
	a := makeElement(board.TaskType, 1, "a")
	elements := []*board.Element{a}

	plan := BuildPlan(elements)
	problems := DetectProblems(plan, elements)

	// No error problems
	for _, p := range problems {
		if p.Severity == "error" {
			t.Errorf("unexpected error problem: %s", p.Message)
		}
	}

	// Critical path should be nil for single element with dist=0
	// (computeCriticalPath returns nil when maxDist==0 since maxID stays "")
	// Actually, let's check what BuildPlan gives us
	if plan.CriticalPath != nil && len(plan.CriticalPath) > 0 {
		// If there is a critical path entry, it's fine as info
		found := false
		for _, p := range problems {
			if p.Type == "critical-path" {
				found = true
			}
		}
		if !found {
			t.Error("expected critical-path problem when CriticalPath is non-empty")
		}
	}
}

func TestFormatCriticalPath(t *testing.T) {
	a := makeElement(board.TaskType, 1, "a")
	b := makeElement(board.TaskType, 2, "b")
	c := makeElement(board.TaskType, 6, "c")

	result := FormatCriticalPath([]*board.Element{a, b, c})
	expected := "TASK-01 -> TASK-02 -> TASK-06 (3 elements)"

	if result != expected {
		t.Errorf("FormatCriticalPath = %q, want %q", result, expected)
	}
}

func TestFormatCriticalPath_Empty(t *testing.T) {
	result := FormatCriticalPath([]*board.Element{})
	if result != "" {
		t.Errorf("FormatCriticalPath(empty) = %q, want empty", result)
	}
}

func TestFormatCriticalPath_Single(t *testing.T) {
	a := makeElement(board.TaskType, 1, "a")
	result := FormatCriticalPath([]*board.Element{a})
	expected := "TASK-01 (1 elements)"
	if result != expected {
		t.Errorf("FormatCriticalPath(single) = %q, want %q", result, expected)
	}
}
