package plan

import (
	"testing"

	"github.com/aagrigore/task-board/internal/board"
)

func makeElement(typ board.ElementType, num int, name string, blockedBy ...string) *board.Element {
	return &board.Element{
		Type:      typ,
		Number:    num,
		Name:      name,
		BlockedBy: blockedBy,
	}
}

func TestBuildGraphLinear(t *testing.T) {
	// A -> B -> C
	a := makeElement(board.TaskType, 1, "a")
	b := makeElement(board.TaskType, 2, "b", "TASK-01")
	c := makeElement(board.TaskType, 3, "c", "TASK-02")

	g := BuildGraph([]*board.Element{a, b, c})

	if g.InDegree["TASK-01"] != 0 {
		t.Errorf("TASK-01 indegree = %d, want 0", g.InDegree["TASK-01"])
	}
	if g.InDegree["TASK-02"] != 1 {
		t.Errorf("TASK-02 indegree = %d, want 1", g.InDegree["TASK-02"])
	}
	if g.InDegree["TASK-03"] != 1 {
		t.Errorf("TASK-03 indegree = %d, want 1", g.InDegree["TASK-03"])
	}
}

func TestBuildPlanLinear(t *testing.T) {
	a := makeElement(board.TaskType, 1, "a")
	b := makeElement(board.TaskType, 2, "b", "TASK-01")
	c := makeElement(board.TaskType, 3, "c", "TASK-02")

	plan := BuildPlan([]*board.Element{a, b, c})

	if plan.HasCycle {
		t.Fatal("unexpected cycle")
	}
	if len(plan.Phases) != 3 {
		t.Fatalf("phases = %d, want 3", len(plan.Phases))
	}
	if plan.Phases[0].Elements[0].ID() != "TASK-01" {
		t.Errorf("phase 1 = %v, want TASK-01", plan.Phases[0].Elements[0].ID())
	}
	if plan.Phases[1].Elements[0].ID() != "TASK-02" {
		t.Errorf("phase 2 = %v, want TASK-02", plan.Phases[1].Elements[0].ID())
	}
	if plan.Phases[2].Elements[0].ID() != "TASK-03" {
		t.Errorf("phase 3 = %v, want TASK-03", plan.Phases[2].Elements[0].ID())
	}
}

func TestBuildPlanParallel(t *testing.T) {
	// A and B are independent, C depends on both
	a := makeElement(board.TaskType, 1, "a")
	b := makeElement(board.TaskType, 2, "b")
	c := makeElement(board.TaskType, 3, "c", "TASK-01", "TASK-02")

	plan := BuildPlan([]*board.Element{a, b, c})

	if plan.HasCycle {
		t.Fatal("unexpected cycle")
	}
	if len(plan.Phases) != 2 {
		t.Fatalf("phases = %d, want 2", len(plan.Phases))
	}
	if len(plan.Phases[0].Elements) != 2 {
		t.Errorf("phase 1 has %d elements, want 2", len(plan.Phases[0].Elements))
	}
	if len(plan.Phases[1].Elements) != 1 {
		t.Errorf("phase 2 has %d elements, want 1", len(plan.Phases[1].Elements))
	}
}

func TestBuildPlanDiamond(t *testing.T) {
	// Diamond: A -> B, A -> C, B -> D, C -> D
	a := makeElement(board.TaskType, 1, "a")
	b := makeElement(board.TaskType, 2, "b", "TASK-01")
	c := makeElement(board.TaskType, 3, "c", "TASK-01")
	d := makeElement(board.TaskType, 4, "d", "TASK-02", "TASK-03")

	plan := BuildPlan([]*board.Element{a, b, c, d})

	if plan.HasCycle {
		t.Fatal("unexpected cycle")
	}
	if len(plan.Phases) != 3 {
		t.Fatalf("phases = %d, want 3", len(plan.Phases))
	}
	// Phase 1: A, Phase 2: B+C, Phase 3: D
	if len(plan.Phases[0].Elements) != 1 {
		t.Errorf("phase 1 has %d elements, want 1", len(plan.Phases[0].Elements))
	}
	if len(plan.Phases[1].Elements) != 2 {
		t.Errorf("phase 2 has %d elements, want 2", len(plan.Phases[1].Elements))
	}
	if len(plan.Phases[2].Elements) != 1 {
		t.Errorf("phase 3 has %d elements, want 1", len(plan.Phases[2].Elements))
	}
}

func TestBuildPlanCycle(t *testing.T) {
	// A -> B -> A (cycle)
	a := makeElement(board.TaskType, 1, "a", "TASK-02")
	b := makeElement(board.TaskType, 2, "b", "TASK-01")

	plan := BuildPlan([]*board.Element{a, b})

	if !plan.HasCycle {
		t.Fatal("expected cycle")
	}
	if len(plan.CycleNodes) != 2 {
		t.Errorf("cycle nodes = %d, want 2", len(plan.CycleNodes))
	}
}

func TestBuildPlanSingleElement(t *testing.T) {
	a := makeElement(board.TaskType, 1, "a")

	plan := BuildPlan([]*board.Element{a})

	if plan.HasCycle {
		t.Fatal("unexpected cycle")
	}
	if len(plan.Phases) != 1 {
		t.Fatalf("phases = %d, want 1", len(plan.Phases))
	}
}

func TestBuildPlanEmpty(t *testing.T) {
	plan := BuildPlan([]*board.Element{})

	if plan.HasCycle {
		t.Fatal("unexpected cycle")
	}
	if len(plan.Phases) != 0 {
		t.Errorf("phases = %d, want 0", len(plan.Phases))
	}
}

func TestCriticalPathLinear(t *testing.T) {
	a := makeElement(board.TaskType, 1, "a")
	b := makeElement(board.TaskType, 2, "b", "TASK-01")
	c := makeElement(board.TaskType, 3, "c", "TASK-02")

	plan := BuildPlan([]*board.Element{a, b, c})

	if len(plan.CriticalPath) != 3 {
		t.Fatalf("critical path len = %d, want 3", len(plan.CriticalPath))
	}
	if plan.CriticalPath[0].ID() != "TASK-01" || plan.CriticalPath[2].ID() != "TASK-03" {
		t.Errorf("critical path = %v %v %v", plan.CriticalPath[0].ID(), plan.CriticalPath[1].ID(), plan.CriticalPath[2].ID())
	}
}

func TestCriticalPathDiamond(t *testing.T) {
	// Diamond: critical path is length 3 (A -> B -> D or A -> C -> D)
	a := makeElement(board.TaskType, 1, "a")
	b := makeElement(board.TaskType, 2, "b", "TASK-01")
	c := makeElement(board.TaskType, 3, "c", "TASK-01")
	d := makeElement(board.TaskType, 4, "d", "TASK-02", "TASK-03")

	plan := BuildPlan([]*board.Element{a, b, c, d})

	if len(plan.CriticalPath) != 3 {
		t.Fatalf("critical path len = %d, want 3", len(plan.CriticalPath))
	}
	// Should start with A and end with D
	if plan.CriticalPath[0].ID() != "TASK-01" {
		t.Errorf("critical path start = %s, want TASK-01", plan.CriticalPath[0].ID())
	}
	if plan.CriticalPath[2].ID() != "TASK-04" {
		t.Errorf("critical path end = %s, want TASK-04", plan.CriticalPath[2].ID())
	}
}
