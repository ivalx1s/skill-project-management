package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/aagrigore/task-board/internal/board"
)

// newIDPattern matches the new distributed ID format: TYPE-YYMMDD-xxxxxx
var newIDDirPattern = regexp.MustCompile(`^(EPIC|STORY|TASK|BUG)-\d{6}-[0-9a-z]{6}_`)

func findCreatedDir(parent string, prefix string) string {
	entries, err := os.ReadDir(parent)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() && newIDDirPattern.MatchString(e.Name()) && e.Name()[:len(prefix)] == prefix {
			return filepath.Join(parent, e.Name())
		}
	}
	return ""
}

func TestCreateEpic(t *testing.T) {
	dir := t.TempDir()
	bd := filepath.Join(dir, ".task-board")
	boardDir = bd

	createName = "test-epic"
	createDescription = "A test epic"

	err := runCreateEpic(createEpicCmd, nil)
	if err != nil {
		t.Fatalf("runCreateEpic: %v", err)
	}

	// Find created epic directory (new format)
	epicDir := findCreatedDir(bd, "EPIC")
	if epicDir == "" {
		t.Fatal("epic directory not created")
	}

	// Verify README.md
	if _, err := os.Stat(filepath.Join(epicDir, "README.md")); os.IsNotExist(err) {
		t.Fatal("README.md not created")
	}

	// Verify progress.md
	if _, err := os.Stat(filepath.Join(epicDir, "progress.md")); os.IsNotExist(err) {
		t.Fatal("progress.md not created")
	}
}

func TestCreateStory(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	createName = "new-story"
	createDescription = "A new story"
	createEpicFlag = testEpic1ID

	err := runCreateStory(createStoryCmd, nil)
	if err != nil {
		t.Fatalf("runCreateStory: %v", err)
	}

	// Find created story directory inside test epic
	epicDir := filepath.Join(bd, testEpic1ID+"_recording")
	storyDir := findCreatedDir(epicDir, "STORY")
	if storyDir == "" {
		t.Fatal("story directory not created")
	}

	// Verify README.md exists
	if _, err := os.Stat(filepath.Join(storyDir, "README.md")); os.IsNotExist(err) {
		t.Fatal("README.md not created")
	}
}

func TestCreateTask(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	createName = "new-task"
	createDescription = "A new task"
	createStoryFlag = testStory1ID

	err := runCreateTask(createTaskCmd, nil)
	if err != nil {
		t.Fatalf("runCreateTask: %v", err)
	}

	// Find created task directory inside test story
	storyDir := filepath.Join(bd, testEpic1ID+"_recording", testStory1ID+"_audio-capture")
	taskDir := findCreatedDir(storyDir, "TASK")
	if taskDir == "" {
		t.Fatal("task directory not created")
	}

	// Verify board can load the new element
	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("board.Load: %v", err)
	}

	// Count tasks
	taskCount := 0
	for _, e := range b.Elements {
		if e.Type == board.TaskType {
			taskCount++
		}
	}
	// Should have one more task than before (4 original + 1 new = 5)
	if taskCount != 5 {
		t.Errorf("task count = %d, want 5", taskCount)
	}
}

func TestCreateBug(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	createName = "new-bug"
	createDescription = "A new bug"
	createStoryFlag = testStory1ID

	err := runCreateBug(createBugCmd, nil)
	if err != nil {
		t.Fatalf("runCreateBug: %v", err)
	}

	// Verify board can load the new bug
	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("board.Load: %v", err)
	}

	// Count bugs
	bugCount := 0
	for _, e := range b.Elements {
		if e.Type == board.BugType {
			bugCount++
		}
	}
	// Should have 2 bugs now
	if bugCount != 2 {
		t.Errorf("bug count = %d, want 2", bugCount)
	}
}

func TestCreateWithInvalidParent(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	createName = "orphan"
	createDescription = "orphan"
	createStoryFlag = "STORY-999"

	err := runCreateTask(createTaskCmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid parent")
	}
}
