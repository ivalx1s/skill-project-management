package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aagrigore/task-board/internal/board"
)

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

	// Verify directory exists
	epicDir := filepath.Join(bd, "EPIC-01_test-epic")
	if _, err := os.Stat(epicDir); os.IsNotExist(err) {
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

	// Verify counters
	c, err := board.ReadCounters(bd)
	if err != nil {
		t.Fatalf("ReadCounters: %v", err)
	}
	if c.Epic != 1 {
		t.Errorf("epic counter = %d, want 1", c.Epic)
	}
}

func TestCreateStory(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	createName = "new-story"
	createDescription = "A new story"
	createEpicFlag = "EPIC-01"

	err := runCreateStory(createStoryCmd, nil)
	if err != nil {
		t.Fatalf("runCreateStory: %v", err)
	}

	// Counter should have incremented
	c, err := board.ReadCounters(bd)
	if err != nil {
		t.Fatalf("ReadCounters: %v", err)
	}
	if c.Story != 4 {
		t.Errorf("story counter = %d, want 4", c.Story)
	}

	// Verify directory
	storyDir := filepath.Join(bd, "EPIC-01_recording", "STORY-04_new-story")
	if _, err := os.Stat(storyDir); os.IsNotExist(err) {
		t.Fatal("story directory not created")
	}
}

func TestCreateTask(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	createName = "new-task"
	createDescription = "A new task"
	createStoryFlag = "STORY-01"

	err := runCreateTask(createTaskCmd, nil)
	if err != nil {
		t.Fatalf("runCreateTask: %v", err)
	}

	c, err := board.ReadCounters(bd)
	if err != nil {
		t.Fatalf("ReadCounters: %v", err)
	}
	if c.Task != 5 {
		t.Errorf("task counter = %d, want 5", c.Task)
	}
}

func TestCreateBug(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	createName = "new-bug"
	createDescription = "A new bug"
	createStoryFlag = "STORY-01"

	err := runCreateBug(createBugCmd, nil)
	if err != nil {
		t.Fatalf("runCreateBug: %v", err)
	}

	c, err := board.ReadCounters(bd)
	if err != nil {
		t.Fatalf("ReadCounters: %v", err)
	}
	if c.Bug != 2 {
		t.Errorf("bug counter = %d, want 2", c.Bug)
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
