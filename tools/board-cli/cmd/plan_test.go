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
	if !strings.Contains(out, testEpic1ID) {
		t.Errorf("expected EPIC-01 in output, got:\n%s", out)
	}
	if !strings.Contains(out, testEpic2ID) {
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
		err := runPlan(planCmd, []string{testEpic1ID})
		if err != nil {
			t.Fatalf("runPlan EPIC-01: %v", err)
		}
	})

	if !strings.Contains(out, testStory1ID) {
		t.Errorf("expected STORY-01 in output, got:\n%s", out)
	}
	if !strings.Contains(out, testStory2ID) {
		t.Errorf("expected STORY-02 in output, got:\n%s", out)
	}
}

func TestPlanStory(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	resetPlanFlags()

	out := captureOutput(t, func() {
		err := runPlan(planCmd, []string{testStory1ID})
		if err != nil {
			t.Fatalf("runPlan STORY-01: %v", err)
		}
	})

	// STORY-01 has tasks: TASK-01, TASK-02 (blocked by TASK-01), TASK-03, BUG-01
	if !strings.Contains(out, testTask1ID) {
		t.Errorf("expected TASK-01 in output, got:\n%s", out)
	}
	if !strings.Contains(out, testTask2ID) {
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
	if !strings.Contains(out, testTask1ID) && !strings.Contains(out, testTask2ID) {
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
		err := runPlan(planCmd, []string{testEpic1ID})
		if err != nil {
			t.Fatalf("runPlan --save EPIC-01: %v", err)
		}
	})

	// plan.md should be in the epic's directory
	epicDir := filepath.Join(bd, testEpic1ID+"_recording")
	planPath := filepath.Join(epicDir, "plan.md")
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		t.Fatalf("plan.md not created at %s (output: %s)", planPath, out)
	}

	content, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("reading plan.md: %v", err)
	}

	if !strings.Contains(string(content), testEpic1ID) {
		t.Errorf("plan.md should reference EPIC-01, got:\n%s", string(content))
	}
}

func TestPlanSaveStory(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	resetPlanFlags()
	planSave = true

	out := captureOutput(t, func() {
		err := runPlan(planCmd, []string{testStory1ID})
		if err != nil {
			t.Fatalf("runPlan --save STORY-01: %v", err)
		}
	})

	storyDir := filepath.Join(bd, testEpic1ID+"_recording", testStory1ID+"_audio-capture")
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
		err := runPlan(planCmd, []string{testStory1ID})
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
		err := runPlan(planCmd, []string{testStory1ID})
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

	// Local IDs for this test
	cycleEpicID := "EPIC-260101-cycle1"
	cycleStoryID := "STORY-260101-cycle2"
	cycleTask1ID := "TASK-260101-cycle3"
	cycleTask2ID := "TASK-260101-cycle4"

	epicDir := filepath.Join(bd, cycleEpicID+"_cycle-test")
	os.MkdirAll(epicDir, 0755)
	os.WriteFile(filepath.Join(epicDir, "README.md"),
		[]byte("# "+cycleEpicID+": Cycle Test\n\n## Description\nTest\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Test\n"), 0644)
	os.WriteFile(filepath.Join(epicDir, "progress.md"),
		[]byte("## Status\nbacklog\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	storyDir := filepath.Join(epicDir, cycleStoryID+"_cycle-story")
	os.MkdirAll(storyDir, 0755)
	os.WriteFile(filepath.Join(storyDir, "README.md"),
		[]byte("# "+cycleStoryID+": Cycle Story\n\n## Description\nTest\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Test\n"), 0644)
	os.WriteFile(filepath.Join(storyDir, "progress.md"),
		[]byte("## Status\nbacklog\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-A blocked by TASK-B
	task1Dir := filepath.Join(storyDir, cycleTask1ID+"_alpha")
	os.MkdirAll(task1Dir, 0755)
	os.WriteFile(filepath.Join(task1Dir, "README.md"),
		[]byte("# "+cycleTask1ID+": Alpha\n\n## Description\nTest\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Test\n"), 0644)
	os.WriteFile(filepath.Join(task1Dir, "progress.md"),
		[]byte("## Status\nbacklog\n\n## Blocked By\n- "+cycleTask2ID+"\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-B blocked by TASK-A
	task2Dir := filepath.Join(storyDir, cycleTask2ID+"_beta")
	os.MkdirAll(task2Dir, 0755)
	os.WriteFile(filepath.Join(task2Dir, "README.md"),
		[]byte("# "+cycleTask2ID+": Beta\n\n## Description\nTest\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Test\n"), 0644)
	os.WriteFile(filepath.Join(task2Dir, "progress.md"),
		[]byte("## Status\nbacklog\n\n## Blocked By\n- "+cycleTask1ID+"\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	boardDir = bd
	resetPlanFlags()

	// The cycle detection calls os.Exit(1), so we test the plan directly
	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("loading board: %v", err)
	}

	elements, err := plan.ScopeElements(b, cycleStoryID)
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
		cycleSet[strings.ToUpper(id)] = true
	}
	if !cycleSet[strings.ToUpper(cycleTask1ID)] || !cycleSet[strings.ToUpper(cycleTask2ID)] {
		t.Errorf("expected cycle tasks in cycle nodes, got: %v", p.CycleNodes)
	}
}

// Mixed status board IDs (local to these tests)
const (
	mixedEpic1ID  = "EPIC-260101-mix001"
	mixedEpic2ID  = "EPIC-260101-mix002"
	mixedStory1ID = "STORY-260101-mix003"
	mixedTask1ID  = "TASK-260101-mix004"
	mixedTask2ID  = "TASK-260101-mix005"
	mixedTask3ID  = "TASK-260101-mix006"
	mixedTask4ID  = "TASK-260101-mix007"
)

// setupMixedStatusBoard creates a board with elements in various statuses:
//   - EPIC-01 (development) -> STORY-01 (backlog) -> TASK-01 (done), TASK-02 (backlog), TASK-03 (closed), TASK-04 (blocked)
//   - EPIC-02 (done)
//
// Used for testing the --active filter.
func setupMixedStatusBoard(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bd := filepath.Join(dir, ".task-board")
	os.MkdirAll(bd, 0755)

	// EPIC-01 (development)
	epicDir := filepath.Join(bd, mixedEpic1ID+"_active-epic")
	os.MkdirAll(epicDir, 0755)
	os.WriteFile(filepath.Join(epicDir, "README.md"),
		[]byte("# "+mixedEpic1ID+": Active Epic\n\n## Description\nActive\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Works\n"), 0644)
	os.WriteFile(filepath.Join(epicDir, "progress.md"),
		[]byte("## Status\ndevelopment\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// EPIC-02 (done)
	epic2Dir := filepath.Join(bd, mixedEpic2ID+"_done-epic")
	os.MkdirAll(epic2Dir, 0755)
	os.WriteFile(filepath.Join(epic2Dir, "README.md"),
		[]byte("# "+mixedEpic2ID+": Done Epic\n\n## Description\nFinished\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Done\n"), 0644)
	os.WriteFile(filepath.Join(epic2Dir, "progress.md"),
		[]byte("## Status\ndone\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// STORY-01 (backlog) inside EPIC-01
	storyDir := filepath.Join(epicDir, mixedStory1ID+"_mixed-story")
	os.MkdirAll(storyDir, 0755)
	os.WriteFile(filepath.Join(storyDir, "README.md"),
		[]byte("# "+mixedStory1ID+": Mixed Story\n\n## Description\nMixed statuses\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Test\n"), 0644)
	os.WriteFile(filepath.Join(storyDir, "progress.md"),
		[]byte("## Status\nbacklog\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-01 (done) inside STORY-01
	task1Dir := filepath.Join(storyDir, mixedTask1ID+"_done-task")
	os.MkdirAll(task1Dir, 0755)
	os.WriteFile(filepath.Join(task1Dir, "README.md"),
		[]byte("# "+mixedTask1ID+": Done Task\n\n## Description\nCompleted\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Done\n"), 0644)
	os.WriteFile(filepath.Join(task1Dir, "progress.md"),
		[]byte("## Status\ndone\n\n## Blocked By\n- (none)\n\n## Blocks\n- "+mixedTask2ID+"\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-02 (backlog) inside STORY-01, blocked by TASK-01
	task2Dir := filepath.Join(storyDir, mixedTask2ID+"_open-task")
	os.MkdirAll(task2Dir, 0755)
	os.WriteFile(filepath.Join(task2Dir, "README.md"),
		[]byte("# "+mixedTask2ID+": Open Task\n\n## Description\nStill open\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Open\n"), 0644)
	os.WriteFile(filepath.Join(task2Dir, "progress.md"),
		[]byte("## Status\nbacklog\n\n## Blocked By\n- "+mixedTask1ID+"\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-03 (closed) inside STORY-01
	task3Dir := filepath.Join(storyDir, mixedTask3ID+"_closed-task")
	os.MkdirAll(task3Dir, 0755)
	os.WriteFile(filepath.Join(task3Dir, "README.md"),
		[]byte("# "+mixedTask3ID+": Closed Task\n\n## Description\nClosed\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Closed\n"), 0644)
	os.WriteFile(filepath.Join(task3Dir, "progress.md"),
		[]byte("## Status\nclosed\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-04 (blocked) inside STORY-01
	task4Dir := filepath.Join(storyDir, mixedTask4ID+"_blocked-task")
	os.MkdirAll(task4Dir, 0755)
	os.WriteFile(filepath.Join(task4Dir, "README.md"),
		[]byte("# "+mixedTask4ID+": Blocked Task\n\n## Description\nBlocked\n\n## Scope\ntest\n\n## Acceptance Criteria\n- Blocked\n"), 0644)
	os.WriteFile(filepath.Join(task4Dir, "progress.md"),
		[]byte("## Status\nblocked\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	return bd
}

func TestActiveFilterRemovesDoneAndClosed(t *testing.T) {
	elements := []*board.Element{
		{Title: "Open Task", Status: board.StatusToDev},
		{Title: "Progress Task", Status: board.StatusDevelopment},
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
		{Title: "Open 1", Status: board.StatusToDev},
		{Title: "Progress 1", Status: board.StatusDevelopment},
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
	allElements, err := plan.AllDescendants(b, mixedEpic1ID)
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
		ids[strings.ToUpper(e.ID())] = true
	}
	if ids[strings.ToUpper(mixedTask1ID)] {
		t.Error("TASK-01 (done) should be excluded by --active filter")
	}
	if ids[strings.ToUpper(mixedTask3ID)] {
		t.Error("TASK-03 (closed) should be excluded by --active filter")
	}

	// Verify active elements are still present
	if !ids[strings.ToUpper(mixedTask2ID)] {
		t.Error("TASK-02 (open) should be included by --active filter")
	}
	if !ids[strings.ToUpper(mixedTask4ID)] {
		t.Error("TASK-04 (blocked) should be included by --active filter")
	}

	// Generate DOT and verify content
	dot := plan.GenerateFullDOT(b, filtered)
	if strings.Contains(strings.ToLower(dot), strings.ToLower(mixedTask1ID)) {
		t.Errorf("DOT output should not contain TASK-01 (done), got:\n%s", dot)
	}
	if strings.Contains(strings.ToLower(dot), strings.ToLower(mixedTask3ID)) {
		t.Errorf("DOT output should not contain TASK-03 (closed), got:\n%s", dot)
	}
	if !strings.Contains(strings.ToLower(dot), strings.ToLower(mixedTask2ID)) {
		t.Errorf("DOT output should contain TASK-02 (open), got:\n%s", dot)
	}
	if !strings.Contains(strings.ToLower(dot), strings.ToLower(mixedTask4ID)) {
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
	allElements, err := plan.AllDescendants(b, mixedEpic1ID)
	if err != nil {
		t.Fatalf("AllDescendants: %v", err)
	}

	// Exclude root element (like renderGraph does for phases)
	var withoutRoot []*board.Element
	for _, e := range allElements {
		if strings.ToUpper(e.ID()) != strings.ToUpper(mixedEpic1ID) {
			withoutRoot = append(withoutRoot, e)
		}
	}

	filtered := filterActiveElements(withoutRoot)

	fullPlan := plan.BuildPlan(filtered)
	if fullPlan.HasCycle {
		t.Fatalf("unexpected cycle: %v", fullPlan.CycleNodes)
	}

	dot := plan.GenerateDOT(fullPlan, filtered)
	if strings.Contains(strings.ToLower(dot), strings.ToLower(mixedTask1ID)) {
		t.Errorf("DOT output should not contain TASK-01 (done), got:\n%s", dot)
	}
	if strings.Contains(strings.ToLower(dot), strings.ToLower(mixedTask3ID)) {
		t.Errorf("DOT output should not contain TASK-03 (closed), got:\n%s", dot)
	}
	if !strings.Contains(strings.ToLower(dot), strings.ToLower(mixedTask2ID)) {
		t.Errorf("DOT output should contain TASK-02 (open), got:\n%s", dot)
	}
	if !strings.Contains(strings.ToLower(dot), strings.ToLower(mixedTask4ID)) {
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
		if strings.ToUpper(e.ID()) == strings.ToUpper(mixedEpic2ID) {
			t.Error("EPIC-02 (done) should be excluded by --active filter at project level")
		}
	}

	// EPIC-01 is development, should be included
	found := false
	for _, e := range filtered {
		if strings.ToUpper(e.ID()) == strings.ToUpper(mixedEpic1ID) {
			found = true
			break
		}
	}
	if !found {
		t.Error("EPIC-01 (development) should be included by --active filter at project level")
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
