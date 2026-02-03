package board

import (
	"os"
	"path/filepath"
	"testing"
)

// Test IDs in new distributed format
const (
	tEpic1   = "EPIC-260101-aaaaaa"
	tStory1  = "STORY-260101-bbbbbb"
	tTask1   = "TASK-260101-cccccc"
	tTask2   = "TASK-260101-dddddd"
	tStory2  = "STORY-260101-eeeeee"
)

func setupTestBoard(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	boardDir := filepath.Join(dir, ".task-board")
	os.MkdirAll(boardDir, 0755)

	// Create EPIC
	epicDir := filepath.Join(boardDir, tEpic1+"_recording")
	os.MkdirAll(epicDir, 0755)
	os.WriteFile(filepath.Join(epicDir, "README.md"), []byte("# "+tEpic1+": recording\n\n## Description\nRecording epic\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(epicDir, "progress.md"), []byte("## Status\ndevelopment\n\n## Blocked By\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// Create STORY inside epic
	storyDir := filepath.Join(epicDir, tStory1+"_audio")
	os.MkdirAll(storyDir, 0755)
	os.WriteFile(filepath.Join(storyDir, "README.md"), []byte("# "+tStory1+": audio\n\n## Description\nAudio story\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(storyDir, "progress.md"), []byte("## Status\nto-dev\n\n## Blocked By\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// Create TASK-01 inside story
	taskDir := filepath.Join(storyDir, tTask1+"_interface")
	os.MkdirAll(taskDir, 0755)
	os.WriteFile(filepath.Join(taskDir, "README.md"), []byte("# "+tTask1+": interface\n\n## Description\nInterface task\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(taskDir, "progress.md"), []byte("## Status\nto-dev\n\n## Blocked By\n- (none)\n\n## Checklist\n- [ ] Step 1\n- [x] Step 2\n\n## Notes\n"), 0644)

	// Create TASK-02 inside story, blocked by TASK-01
	task2Dir := filepath.Join(storyDir, tTask2+"_impl")
	os.MkdirAll(task2Dir, 0755)
	os.WriteFile(filepath.Join(task2Dir, "README.md"), []byte("# "+tTask2+": impl\n\n## Description\nImpl task\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(task2Dir, "progress.md"), []byte("## Status\nto-dev\n\n## Blocked By\n- "+tTask1+"\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

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
	epic := b.FindByID(tEpic1)
	if epic == nil {
		t.Fatalf("%s not found", tEpic1)
	}
	if epic.Status != StatusDevelopment {
		t.Errorf("%s status = %v, want development", tEpic1, epic.Status)
	}

	// Check story
	story := b.FindByID(tStory1)
	if story == nil {
		t.Fatalf("%s not found", tStory1)
	}
	if story.ParentID != tEpic1 {
		t.Errorf("%s parent = %v, want %s", tStory1, story.ParentID, tEpic1)
	}

	// Check task
	task := b.FindByID(tTask1)
	if task == nil {
		t.Fatalf("%s not found", tTask1)
	}
	if len(task.Checklist) != 2 {
		t.Errorf("%s checklist len = %d, want 2", tTask1, len(task.Checklist))
	}

	// Check blocked task
	task2 := b.FindByID(tTask2)
	if task2 == nil {
		t.Fatalf("%s not found", tTask2)
	}
	if len(task2.BlockedBy) != 1 || task2.BlockedBy[0] != tTask1 {
		t.Errorf("%s blockedBy = %v, want [%s]", tTask2, task2.BlockedBy, tTask1)
	}
}

func TestAncestry(t *testing.T) {
	boardDir := setupTestBoard(t)

	b, err := Load(boardDir)
	if err != nil {
		t.Fatal(err)
	}

	task := b.FindByID(tTask1)
	if task == nil {
		t.Fatalf("%s not found", tTask1)
	}
	ancestry := b.Ancestry(task)
	expected := tEpic1 + " > " + tStory1 + " > " + tTask1
	if ancestry != expected {
		t.Errorf("Ancestry = %q, want %q", ancestry, expected)
	}
}

func TestParentOf(t *testing.T) {
	boardDir := setupTestBoard(t)
	b, err := Load(boardDir)
	if err != nil {
		t.Fatal(err)
	}

	// Task's parent is story
	task := b.FindByID(tTask1)
	if task == nil {
		t.Fatalf("%s not found", tTask1)
	}
	parent := b.ParentOf(task)
	if parent == nil || parent.ID() != tStory1 {
		t.Errorf("ParentOf(%s) = %v, want %s", tTask1, parent, tStory1)
	}

	// Story's parent is epic
	story := b.FindByID(tStory1)
	parent = b.ParentOf(story)
	if parent == nil || parent.ID() != tEpic1 {
		t.Errorf("ParentOf(%s) = %v, want %s", tStory1, parent, tEpic1)
	}

	// Epic has no parent
	epic := b.FindByID(tEpic1)
	parent = b.ParentOf(epic)
	if parent != nil {
		t.Errorf("ParentOf(%s) = %v, want nil", tEpic1, parent)
	}
}

func TestHasCrossChildDependency(t *testing.T) {
	dir := t.TempDir()
	boardDir := filepath.Join(dir, ".task-board")
	os.MkdirAll(boardDir, 0755)

	epicID := "EPIC-260101-ffffff"
	story1ID := "STORY-260101-gggggg"
	story2ID := "STORY-260101-hhhhhh"
	task1ID := "TASK-260101-iiiiii"
	task2ID := "TASK-260101-jjjjjj"

	// EPIC with two stories
	epicDir := filepath.Join(boardDir, epicID+"_test")
	os.MkdirAll(epicDir, 0755)
	os.WriteFile(filepath.Join(epicDir, "README.md"), []byte("# "+epicID+": test\n\n## Description\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(epicDir, "progress.md"), []byte("## Status\nto-dev\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	story1Dir := filepath.Join(epicDir, story1ID+"_a")
	os.MkdirAll(story1Dir, 0755)
	os.WriteFile(filepath.Join(story1Dir, "README.md"), []byte("# "+story1ID+": a\n\n## Description\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(story1Dir, "progress.md"), []byte("## Status\nto-dev\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	story2Dir := filepath.Join(epicDir, story2ID+"_b")
	os.MkdirAll(story2Dir, 0755)
	os.WriteFile(filepath.Join(story2Dir, "README.md"), []byte("# "+story2ID+": b\n\n## Description\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(story2Dir, "progress.md"), []byte("## Status\nto-dev\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-01 in STORY-01, TASK-02 in STORY-02, TASK-02 blocked by TASK-01
	task1Dir := filepath.Join(story1Dir, task1ID+"_x")
	os.MkdirAll(task1Dir, 0755)
	os.WriteFile(filepath.Join(task1Dir, "README.md"), []byte("# "+task1ID+": x\n\n## Description\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(task1Dir, "progress.md"), []byte("## Status\nto-dev\n\n## Blocked By\n- (none)\n\n## Blocks\n- "+task2ID+"\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	task2Dir := filepath.Join(story2Dir, task2ID+"_y")
	os.MkdirAll(task2Dir, 0755)
	os.WriteFile(filepath.Join(task2Dir, "README.md"), []byte("# "+task2ID+": y\n\n## Description\n\n## Scope\n\n## Acceptance Criteria\n"), 0644)
	os.WriteFile(filepath.Join(task2Dir, "progress.md"), []byte("## Status\nto-dev\n\n## Blocked By\n- "+task1ID+"\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	b, err := Load(boardDir)
	if err != nil {
		t.Fatal(err)
	}

	story1 := b.FindByID(story1ID)
	story2 := b.FindByID(story2ID)

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

	toDev := FilterByStatus(b.Elements, StatusToDev)
	if len(toDev) != 3 { // story + 2 tasks
		t.Errorf("to-dev count = %d, want 3", len(toDev))
	}

	development := FilterByStatus(b.Elements, StatusDevelopment)
	if len(development) != 1 { // epic
		t.Errorf("development count = %d, want 1", len(development))
	}
}
