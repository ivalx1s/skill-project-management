package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/plan"
)

// captureOutput captures stdout during function execution.
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func resetPlanFlags() {
	planSave = false
	planCriticalPath = false
	planPhase = 0
	planRender = false
	planActive = false
	planLayout = "hierarchy"
	planFormat = "svg"
}

func TestPlanNoArgs(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	resetPlanFlags()

	out := captureOutput(t, func() {
		err := runPlan(planCmd, []string{})
		if err != nil {
			t.Fatalf("runPlan no args: %v", err)
		}
	})

	// Should show epics in plan
	if !strings.Contains(out, "Plan: Project") {
		t.Errorf("expected 'Plan: Project' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "EPIC-01") {
		t.Errorf("expected EPIC-01 in output, got:\n%s", out)
	}
	if !strings.Contains(out, "EPIC-02") {
		t.Errorf("expected EPIC-02 in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Phase 1") {
		t.Errorf("expected Phase 1 in output, got:\n%s", out)
	}
}

func TestPlanEpic(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	resetPlanFlags()

	out := captureOutput(t, func() {
		err := runPlan(planCmd, []string{"EPIC-01"})
		if err != nil {
			t.Fatalf("runPlan EPIC-01: %v", err)
		}
	})

	if !strings.Contains(out, "STORY-01") {
		t.Errorf("expected STORY-01 in output, got:\n%s", out)
	}
	if !strings.Contains(out, "STORY-02") {
		t.Errorf("expected STORY-02 in output, got:\n%s", out)
	}
}

func TestPlanStory(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	resetPlanFlags()

	out := captureOutput(t, func() {
		err := runPlan(planCmd, []string{"STORY-01"})
		if err != nil {
			t.Fatalf("runPlan STORY-01: %v", err)
		}
	})

	// STORY-01 has tasks: TASK-01, TASK-02 (blocked by TASK-01), TASK-03, BUG-01
	if !strings.Contains(out, "TASK-01") {
		t.Errorf("expected TASK-01 in output, got:\n%s", out)
	}
	if !strings.Contains(out, "TASK-02") {
		t.Errorf("expected TASK-02 in output, got:\n%s", out)
	}
	// TASK-02 is blocked by TASK-01, so there should be Phase 2
	if !strings.Contains(out, "Phase 2") {
		t.Errorf("expected Phase 2 in output (TASK-02 blocked by TASK-01), got:\n%s", out)
	}
	// Critical path should be shown
	if !strings.Contains(out, "Critical Path") {
		t.Errorf("expected 'Critical Path' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "TASK-01") && !strings.Contains(out, "TASK-02") {
		t.Errorf("expected critical path with TASK-01 and TASK-02, got:\n%s", out)
	}
}

func TestPlanSave(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	resetPlanFlags()
	planSave = true

	out := captureOutput(t, func() {
		err := runPlan(planCmd, []string{})
		if err != nil {
			t.Fatalf("runPlan --save: %v", err)
		}
	})

	planPath := filepath.Join(bd, "plan.md")
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		t.Fatalf("plan.md not created at %s", planPath)
	}

	content, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("reading plan.md: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "# Plan: Project") {
		t.Errorf("plan.md missing header, got:\n%s", s)
	}
	if !strings.Contains(s, "## Phase 1") {
		t.Errorf("plan.md missing Phase 1, got:\n%s", s)
	}
	if !strings.Contains(s, "## Critical Path") {
		t.Errorf("plan.md missing Critical Path, got:\n%s", s)
	}
	if !strings.Contains(s, "## Warnings") {
		t.Errorf("plan.md missing Warnings, got:\n%s", s)
	}

	if !strings.Contains(out, "Plan saved to") {
		t.Errorf("expected save confirmation in output, got:\n%s", out)
	}
}

func TestPlanSaveEpic(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	resetPlanFlags()
	planSave = true

	out := captureOutput(t, func() {
		err := runPlan(planCmd, []string{"EPIC-01"})
		if err != nil {
			t.Fatalf("runPlan --save EPIC-01: %v", err)
		}
	})

	// plan.md should be in the epic's directory
	epicDir := filepath.Join(bd, "EPIC-01_recording")
	planPath := filepath.Join(epicDir, "plan.md")
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		t.Fatalf("plan.md not created at %s (output: %s)", planPath, out)
	}

	content, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("reading plan.md: %v", err)
	}

	if !strings.Contains(string(content), "EPIC-01") {
		t.Errorf("plan.md should reference EPIC-01, got:\n%s", string(content))
	}
}

func TestPlanSaveStory(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	resetPlanFlags()
	planSave = true

	out := captureOutput(t, func() {
		err := runPlan(planCmd, []string{"STORY-01"})
		if err != nil {
			t.Fatalf("runPlan --save STORY-01: %v", err)
		}
	})

	storyDir := filepath.Join(bd, "EPIC-01_recording", "STORY-01_audio-capture")
	planPath := filepath.Join(storyDir, "plan.md")
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		t.Fatalf("plan.md not created at %s (output: %s)", planPath, out)
	}
}

func TestPlanCriticalPathFlag(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	resetPlanFlags()
	planCriticalPath = true

	out := captureOutput(t, func() {
		err := runPlan(planCmd, []string{"STORY-01"})
		if err != nil {
			t.Fatalf("runPlan --critical-path: %v", err)
		}
	})

	if !strings.Contains(out, "Critical Path") {
		t.Errorf("expected 'Critical Path' in output, got:\n%s", out)
	}
	// Should NOT show phase tables
	if strings.Contains(out, "Phase 1") {
		t.Errorf("--critical-path should not show phases, got:\n%s", out)
	}
}

func TestPlanPhaseFlag(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	resetPlanFlags()
	planPhase = 1

	out := captureOutput(t, func() {
		err := runPlan(planCmd, []string{"STORY-01"})
		if err != nil {
			t.Fatalf("runPlan --phase 1: %v", err)
		}
	})

	if !strings.Contains(out, "Phase 1") {
		t.Errorf("expected Phase 1 in output, got:\n%s", out)
	}
	// Should NOT show Phase 2
	if strings.Contains(out, "Phase 2") {
		t.Errorf("--phase 1 should not show Phase 2, got:\n%s", out)
	}
	// Should NOT show critical path when filtering by phase
	if strings.Contains(out, "Critical Path") {
		t.Errorf("--phase should not show critical path, got:\n%s", out)
	}
}

func TestPlanCycleDetection(t *testing.T) {
	// Create a board with a cycle: TASK-A blocked by TASK-B, TASK-B blocked by TASK-A
	dir := t.TempDir()
	bd := filepath.Join(dir, ".task-board")
	os.MkdirAll(bd, 0755)

	board.WriteCounters(bd, &board.Counters{Epic: 1, Story: 1, Task: 2, Bug: 0})

	epicDir := filepath.Join(bd, "EPIC-01_cycle-test")
	os.MkdirAll(epicDir, 0755)
	os.WriteFile(filepath.Join(epicDir, "README.md"),
		[]byte("# EPIC-01: Cycle Test\n\n## Description\nTest\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Test\n"), 0644)
	os.WriteFile(filepath.Join(epicDir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	storyDir := filepath.Join(epicDir, "STORY-01_cycle-story")
	os.MkdirAll(storyDir, 0755)
	os.WriteFile(filepath.Join(storyDir, "README.md"),
		[]byte("# STORY-01: Cycle Story\n\n## Description\nTest\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Test\n"), 0644)
	os.WriteFile(filepath.Join(storyDir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-01 blocked by TASK-02
	task1Dir := filepath.Join(storyDir, "TASK-01_alpha")
	os.MkdirAll(task1Dir, 0755)
	os.WriteFile(filepath.Join(task1Dir, "README.md"),
		[]byte("# TASK-01: Alpha\n\n## Description\nTest\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Test\n"), 0644)
	os.WriteFile(filepath.Join(task1Dir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- TASK-02\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-02 blocked by TASK-01
	task2Dir := filepath.Join(storyDir, "TASK-02_beta")
	os.MkdirAll(task2Dir, 0755)
	os.WriteFile(filepath.Join(task2Dir, "README.md"),
		[]byte("# TASK-02: Beta\n\n## Description\nTest\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Test\n"), 0644)
	os.WriteFile(filepath.Join(task2Dir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- TASK-01\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	boardDir = bd
	resetPlanFlags()

	// The cycle detection calls os.Exit(1), so we test the plan directly
	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("loading board: %v", err)
	}

	elements, err := plan.ScopeElements(b, "STORY-01")
	if err != nil {
		t.Fatalf("scope elements: %v", err)
	}

	p := plan.BuildPlan(elements)
	if !p.HasCycle {
		t.Fatal("expected cycle to be detected")
	}
	if len(p.CycleNodes) == 0 {
		t.Fatal("expected cycle nodes to be listed")
	}

	// Verify cycle nodes contain both tasks
	cycleSet := make(map[string]bool)
	for _, id := range p.CycleNodes {
		cycleSet[id] = true
	}
	if !cycleSet["TASK-01"] || !cycleSet["TASK-02"] {
		t.Errorf("expected TASK-01 and TASK-02 in cycle nodes, got: %v", p.CycleNodes)
	}
}

// setupMixedStatusBoard creates a board with elements in various statuses:
//   - EPIC-01 (progress) -> STORY-01 (open) -> TASK-01 (done), TASK-02 (open), TASK-03 (closed), TASK-04 (blocked)
//   - EPIC-02 (done)
//
// Used for testing the --active filter.
func setupMixedStatusBoard(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bd := filepath.Join(dir, ".task-board")
	os.MkdirAll(bd, 0755)

	board.WriteCounters(bd, &board.Counters{Epic: 2, Story: 1, Task: 4, Bug: 0})

	// EPIC-01 (progress)
	epicDir := filepath.Join(bd, "EPIC-01_active-epic")
	os.MkdirAll(epicDir, 0755)
	os.WriteFile(filepath.Join(epicDir, "README.md"),
		[]byte("# EPIC-01: Active Epic\n\n## Description\nActive\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Works\n"), 0644)
	os.WriteFile(filepath.Join(epicDir, "progress.md"),
		[]byte("## Status\nprogress\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// EPIC-02 (done)
	epic2Dir := filepath.Join(bd, "EPIC-02_done-epic")
	os.MkdirAll(epic2Dir, 0755)
	os.WriteFile(filepath.Join(epic2Dir, "README.md"),
		[]byte("# EPIC-02: Done Epic\n\n## Description\nFinished\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Done\n"), 0644)
	os.WriteFile(filepath.Join(epic2Dir, "progress.md"),
		[]byte("## Status\ndone\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// STORY-01 (open) inside EPIC-01
	storyDir := filepath.Join(epicDir, "STORY-01_mixed-story")
	os.MkdirAll(storyDir, 0755)
	os.WriteFile(filepath.Join(storyDir, "README.md"),
		[]byte("# STORY-01: Mixed Story\n\n## Description\nMixed statuses\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Test\n"), 0644)
	os.WriteFile(filepath.Join(storyDir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-01 (done) inside STORY-01
	task1Dir := filepath.Join(storyDir, "TASK-01_done-task")
	os.MkdirAll(task1Dir, 0755)
	os.WriteFile(filepath.Join(task1Dir, "README.md"),
		[]byte("# TASK-01: Done Task\n\n## Description\nCompleted\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Done\n"), 0644)
	os.WriteFile(filepath.Join(task1Dir, "progress.md"),
		[]byte("## Status\ndone\n\n## Blocked By\n- (none)\n\n## Blocks\n- TASK-02\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-02 (open) inside STORY-01, blocked by TASK-01
	task2Dir := filepath.Join(storyDir, "TASK-02_open-task")
	os.MkdirAll(task2Dir, 0755)
	os.WriteFile(filepath.Join(task2Dir, "README.md"),
		[]byte("# TASK-02: Open Task\n\n## Description\nStill open\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Open\n"), 0644)
	os.WriteFile(filepath.Join(task2Dir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- TASK-01\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-03 (closed) inside STORY-01
	task3Dir := filepath.Join(storyDir, "TASK-03_closed-task")
	os.MkdirAll(task3Dir, 0755)
	os.WriteFile(filepath.Join(task3Dir, "README.md"),
		[]byte("# TASK-03: Closed Task\n\n## Description\nClosed\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Closed\n"), 0644)
	os.WriteFile(filepath.Join(task3Dir, "progress.md"),
		[]byte("## Status\nclosed\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-04 (blocked) inside STORY-01
	task4Dir := filepath.Join(storyDir, "TASK-04_blocked-task")
	os.MkdirAll(task4Dir, 0755)
	os.WriteFile(filepath.Join(task4Dir, "README.md"),
		[]byte("# TASK-04: Blocked Task\n\n## Description\nBlocked\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Blocked\n"), 0644)
	os.WriteFile(filepath.Join(task4Dir, "progress.md"),
		[]byte("## Status\nblocked\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	return bd
}

func TestActiveFilterRemovesDoneAndClosed(t *testing.T) {
	elements := []*board.Element{
		{Title: "Open Task", Status: board.StatusOpen},
		{Title: "Progress Task", Status: board.StatusProgress},
		{Title: "Done Task", Status: board.StatusDone},
		{Title: "Closed Task", Status: board.StatusClosed},
		{Title: "Blocked Task", Status: board.StatusBlocked},
	}

	result := filterActiveElements(elements)

	if len(result) != 3 {
		t.Fatalf("expected 3 active elements, got %d", len(result))
	}

	for _, e := range result {
		if e.Status == board.StatusDone || e.Status == board.StatusClosed {
			t.Errorf("filterActiveElements should not include %s element (%s)", e.Status, e.Title)
		}
	}

	// Verify the kept elements
	titles := make(map[string]bool)
	for _, e := range result {
		titles[e.Title] = true
	}
	for _, expected := range []string{"Open Task", "Progress Task", "Blocked Task"} {
		if !titles[expected] {
			t.Errorf("expected %q in result, not found", expected)
		}
	}
}

func TestActiveFilterEmptyInput(t *testing.T) {
	result := filterActiveElements(nil)
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}

	result = filterActiveElements([]*board.Element{})
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestActiveFilterAllDone(t *testing.T) {
	elements := []*board.Element{
		{Title: "Done 1", Status: board.StatusDone},
		{Title: "Done 2", Status: board.StatusDone},
		{Title: "Closed 1", Status: board.StatusClosed},
	}

	result := filterActiveElements(elements)
	if result != nil {
		t.Errorf("expected nil when all elements are done/closed, got %d elements", len(result))
	}
}

func TestActiveFilterNoneDone(t *testing.T) {
	elements := []*board.Element{
		{Title: "Open 1", Status: board.StatusOpen},
		{Title: "Progress 1", Status: board.StatusProgress},
	}

	result := filterActiveElements(elements)
	if len(result) != 2 {
		t.Fatalf("expected 2 elements when none are done/closed, got %d", len(result))
	}
}

func TestActiveRenderHierarchyExcludesDone(t *testing.T) {
	bd := setupMixedStatusBoard(t)
	boardDir = bd
	resetPlanFlags()
	planActive = true

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("loading board: %v", err)
	}

	// Simulate what renderGraph does for hierarchy layout with --active
	allElements, err := plan.AllDescendants(b, "EPIC-01")
	if err != nil {
		t.Fatalf("AllDescendants: %v", err)
	}

	filtered := filterActiveElements(allElements)

	// Verify done/closed elements are excluded
	for _, e := range filtered {
		if e.Status == board.StatusDone || e.Status == board.StatusClosed {
			t.Errorf("active filter should exclude %s (%s), but it was included", e.ID(), e.Status)
		}
	}

	// Verify TASK-01 (done) and TASK-03 (closed) are not present
	ids := make(map[string]bool)
	for _, e := range filtered {
		ids[e.ID()] = true
	}
	if ids["TASK-01"] {
		t.Error("TASK-01 (done) should be excluded by --active filter")
	}
	if ids["TASK-03"] {
		t.Error("TASK-03 (closed) should be excluded by --active filter")
	}

	// Verify active elements are still present
	if !ids["TASK-02"] {
		t.Error("TASK-02 (open) should be included by --active filter")
	}
	if !ids["TASK-04"] {
		t.Error("TASK-04 (blocked) should be included by --active filter")
	}

	// Generate DOT and verify content
	dot := plan.GenerateFullDOT(b, filtered)
	if strings.Contains(dot, "TASK-01") {
		t.Errorf("DOT output should not contain TASK-01 (done), got:\n%s", dot)
	}
	if strings.Contains(dot, "TASK-03") {
		t.Errorf("DOT output should not contain TASK-03 (closed), got:\n%s", dot)
	}
	if !strings.Contains(dot, "TASK-02") {
		t.Errorf("DOT output should contain TASK-02 (open), got:\n%s", dot)
	}
	if !strings.Contains(dot, "TASK-04") {
		t.Errorf("DOT output should contain TASK-04 (blocked), got:\n%s", dot)
	}
}

func TestActiveRenderPhasesExcludesDone(t *testing.T) {
	bd := setupMixedStatusBoard(t)
	boardDir = bd
	resetPlanFlags()
	planActive = true

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("loading board: %v", err)
	}

	// Simulate what renderGraph does for phases layout with --active
	allElements, err := plan.AllDescendants(b, "EPIC-01")
	if err != nil {
		t.Fatalf("AllDescendants: %v", err)
	}

	// Exclude root element (like renderGraph does for phases)
	var withoutRoot []*board.Element
	for _, e := range allElements {
		if e.ID() != "EPIC-01" {
			withoutRoot = append(withoutRoot, e)
		}
	}

	filtered := filterActiveElements(withoutRoot)

	fullPlan := plan.BuildPlan(filtered)
	if fullPlan.HasCycle {
		t.Fatalf("unexpected cycle: %v", fullPlan.CycleNodes)
	}

	dot := plan.GenerateDOT(fullPlan, filtered)
	if strings.Contains(dot, "TASK-01") {
		t.Errorf("DOT output should not contain TASK-01 (done), got:\n%s", dot)
	}
	if strings.Contains(dot, "TASK-03") {
		t.Errorf("DOT output should not contain TASK-03 (closed), got:\n%s", dot)
	}
	if !strings.Contains(dot, "TASK-02") {
		t.Errorf("DOT output should contain TASK-02 (open), got:\n%s", dot)
	}
	if !strings.Contains(dot, "TASK-04") {
		t.Errorf("DOT output should contain TASK-04 (blocked), got:\n%s", dot)
	}
}

func TestActiveProjectLevelExcludesDoneEpic(t *testing.T) {
	bd := setupMixedStatusBoard(t)
	boardDir = bd
	resetPlanFlags()
	planActive = true

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("loading board: %v", err)
	}

	// Project-level: all descendants from root
	allElements, err := plan.AllDescendants(b, "")
	if err != nil {
		t.Fatalf("AllDescendants: %v", err)
	}

	filtered := filterActiveElements(allElements)

	// EPIC-02 is done, should be excluded
	for _, e := range filtered {
		if e.ID() == "EPIC-02" {
			t.Error("EPIC-02 (done) should be excluded by --active filter at project level")
		}
	}

	// EPIC-01 is progress, should be included
	found := false
	for _, e := range filtered {
		if e.ID() == "EPIC-01" {
			found = true
			break
		}
	}
	if !found {
		t.Error("EPIC-01 (progress) should be included by --active filter at project level")
	}
}

func TestPlanInvalidScope(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	resetPlanFlags()

	err := runPlan(planCmd, []string{"EPIC-99"})
	if err == nil {
		t.Fatal("expected error for invalid scope ID")
	}
}
