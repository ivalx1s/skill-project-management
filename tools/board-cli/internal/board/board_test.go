package board

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestBoard(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	boardDir := filepath.Join(dir, ".task-board")
	os.MkdirAll(boardDir, 0755)

	// Write system.md
	WriteCounters(boardDir, &Counters{Epic: 1, Story: 1, Task: 2, Bug: 0})

	// Create EPIC-01_recording
	epicDir := filepath.Join(boardDir, "EPIC-01_recording")
	os.MkdirAll(epicDir, 0755)
	os.WriteFile(filepath.Join(epicDir, "README.md"), []byte("# EPIC-01: recording\n\n## Description\nRecording epic\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(epicDir, "progress.md"), []byte("## Status\nprogress\n\n## Blocked By\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// Create STORY-01_audio inside epic
	storyDir := filepath.Join(epicDir, "STORY-01_audio")
	os.MkdirAll(storyDir, 0755)
	os.WriteFile(filepath.Join(storyDir, "README.md"), []byte("# STORY-01: audio\n\n## Description\nAudio story\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(storyDir, "progress.md"), []byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// Create TASK-01_interface inside story
	taskDir := filepath.Join(storyDir, "TASK-01_interface")
	os.MkdirAll(taskDir, 0755)
	os.WriteFile(filepath.Join(taskDir, "README.md"), []byte("# TASK-01: interface\n\n## Description\nInterface task\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(taskDir, "progress.md"), []byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Checklist\n- [ ] Step 1\n- [x] Step 2\n\n## Notes\n"), 0644)

	// Create TASK-02_impl inside story, blocked by TASK-01
	task2Dir := filepath.Join(storyDir, "TASK-02_impl")
	os.MkdirAll(task2Dir, 0755)
	os.WriteFile(filepath.Join(task2Dir, "README.md"), []byte("# TASK-02: impl\n\n## Description\nImpl task\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(task2Dir, "progress.md"), []byte("## Status\nopen\n\n## Blocked By\n- TASK-01\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	return boardDir
}

func TestLoadBoard(t *testing.T) {
	boardDir := setupTestBoard(t)

	b, err := Load(boardDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(b.Elements) != 4 {
		t.Fatalf("Elements count = %d, want 4", len(b.Elements))
	}

	// Check epic
	epic := b.FindByID("EPIC-01")
	if epic == nil {
		t.Fatal("EPIC-01 not found")
	}
	if epic.Status != StatusProgress {
		t.Errorf("EPIC-01 status = %v, want progress", epic.Status)
	}

	// Check story
	story := b.FindByID("STORY-01")
	if story == nil {
		t.Fatal("STORY-01 not found")
	}
	if story.ParentID != "EPIC-01" {
		t.Errorf("STORY-01 parent = %v, want EPIC-01", story.ParentID)
	}

	// Check task
	task := b.FindByID("TASK-01")
	if task == nil {
		t.Fatal("TASK-01 not found")
	}
	if len(task.Checklist) != 2 {
		t.Errorf("TASK-01 checklist len = %d, want 2", len(task.Checklist))
	}

	// Check blocked task
	task2 := b.FindByID("TASK-02")
	if task2 == nil {
		t.Fatal("TASK-02 not found")
	}
	if len(task2.BlockedBy) != 1 || task2.BlockedBy[0] != "TASK-01" {
		t.Errorf("TASK-02 blockedBy = %v, want [TASK-01]", task2.BlockedBy)
	}
}

func TestAncestry(t *testing.T) {
	boardDir := setupTestBoard(t)

	b, err := Load(boardDir)
	if err != nil {
		t.Fatal(err)
	}

	task := b.FindByID("TASK-01")
	ancestry := b.Ancestry(task)
	if ancestry != "EPIC-01 > STORY-01 > TASK-01" {
		t.Errorf("Ancestry = %q, want 'EPIC-01 > STORY-01 > TASK-01'", ancestry)
	}
}

func TestParentOf(t *testing.T) {
	boardDir := setupTestBoard(t)
	b, err := Load(boardDir)
	if err != nil {
		t.Fatal(err)
	}

	// Task's parent is story
	task := b.FindByID("TASK-01")
	parent := b.ParentOf(task)
	if parent == nil || parent.ID() != "STORY-01" {
		t.Errorf("ParentOf(TASK-01) = %v, want STORY-01", parent)
	}

	// Story's parent is epic
	story := b.FindByID("STORY-01")
	parent = b.ParentOf(story)
	if parent == nil || parent.ID() != "EPIC-01" {
		t.Errorf("ParentOf(STORY-01) = %v, want EPIC-01", parent)
	}

	// Epic has no parent
	epic := b.FindByID("EPIC-01")
	parent = b.ParentOf(epic)
	if parent != nil {
		t.Errorf("ParentOf(EPIC-01) = %v, want nil", parent)
	}
}

func TestHasCrossChildDependency(t *testing.T) {
	dir := t.TempDir()
	boardDir := filepath.Join(dir, ".task-board")
	os.MkdirAll(boardDir, 0755)
	WriteCounters(boardDir, &Counters{Epic: 1, Story: 2, Task: 2, Bug: 0})

	// EPIC-01 with two stories
	epicDir := filepath.Join(boardDir, "EPIC-01_test")
	os.MkdirAll(epicDir, 0755)
	os.WriteFile(filepath.Join(epicDir, "README.md"), []byte("# EPIC-01: test\n\n## Description\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(epicDir, "progress.md"), []byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	story1Dir := filepath.Join(epicDir, "STORY-01_a")
	os.MkdirAll(story1Dir, 0755)
	os.WriteFile(filepath.Join(story1Dir, "README.md"), []byte("# STORY-01: a\n\n## Description\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(story1Dir, "progress.md"), []byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	story2Dir := filepath.Join(epicDir, "STORY-02_b")
	os.MkdirAll(story2Dir, 0755)
	os.WriteFile(filepath.Join(story2Dir, "README.md"), []byte("# STORY-02: b\n\n## Description\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(story2Dir, "progress.md"), []byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-01 in STORY-01, TASK-02 in STORY-02, TASK-02 blocked by TASK-01
	task1Dir := filepath.Join(story1Dir, "TASK-01_x")
	os.MkdirAll(task1Dir, 0755)
	os.WriteFile(filepath.Join(task1Dir, "README.md"), []byte("# TASK-01: x\n\n## Description\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(task1Dir, "progress.md"), []byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Blocks\n- TASK-02\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	task2Dir := filepath.Join(story2Dir, "TASK-02_y")
	os.MkdirAll(task2Dir, 0755)
	os.WriteFile(filepath.Join(task2Dir, "README.md"), []byte("# TASK-02: y\n\n## Description\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(task2Dir, "progress.md"), []byte("## Status\nopen\n\n## Blocked By\n- TASK-01\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	b, err := Load(boardDir)
	if err != nil {
		t.Fatal(err)
	}

	story1 := b.FindByID("STORY-01")
	story2 := b.FindByID("STORY-02")

	// STORY-02 has child TASK-02 blocked by TASK-01 (child of STORY-01)
	if !b.HasCrossChildDependency(story2, story1) {
		t.Error("expected cross-child dependency from STORY-02 to STORY-01")
	}

	// Reverse should be false
	if b.HasCrossChildDependency(story1, story2) {
		t.Error("unexpected cross-child dependency from STORY-01 to STORY-02")
	}
}

func TestFilterByStatus(t *testing.T) {
	boardDir := setupTestBoard(t)

	b, err := Load(boardDir)
	if err != nil {
		t.Fatal(err)
	}

	open := FilterByStatus(b.Elements, StatusOpen)
	if len(open) != 3 { // story + 2 tasks
		t.Errorf("Open count = %d, want 3", len(open))
	}

	progress := FilterByStatus(b.Elements, StatusProgress)
	if len(progress) != 1 { // epic
		t.Errorf("Progress count = %d, want 1", len(progress))
	}
}
