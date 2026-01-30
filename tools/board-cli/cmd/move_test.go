package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMoveTask(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	moveToFlag = "STORY-02"

	// Move TASK-03 from STORY-01 to STORY-02
	err := runMove(moveCmd, []string{"TASK-03"})
	if err != nil {
		t.Fatalf("runMove: %v", err)
	}

	// Old path should not exist
	oldPath := filepath.Join(bd, "EPIC-01_recording", "STORY-01_audio-capture", "TASK-03_tests")
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("old path still exists")
	}

	// New path should exist
	newPath := filepath.Join(bd, "EPIC-01_recording", "STORY-02_amplitude", "TASK-03_tests")
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("new path does not exist")
	}
}

func TestMoveStoryToEpic(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	moveToFlag = "EPIC-02"

	err := runMove(moveCmd, []string{"STORY-02"})
	if err != nil {
		t.Fatalf("runMove: %v", err)
	}

	newPath := filepath.Join(bd, "EPIC-02_storage", "STORY-02_amplitude")
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("story not found in new epic")
	}
}

func TestMoveInvalidType(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	moveToFlag = "TASK-01"

	// Try to move task to task — should fail
	err := runMove(moveCmd, []string{"TASK-03"})
	if err == nil {
		t.Fatal("expected error for invalid target type")
	}
	if !strings.Contains(err.Error(), "stories") {
		t.Errorf("error = %q, want mention of stories", err.Error())
	}
}

func TestMoveEpicFails(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	moveToFlag = "EPIC-02"

	err := runMove(moveCmd, []string{"EPIC-01"})
	if err == nil {
		t.Fatal("expected error for moving epic")
	}
	if !strings.Contains(err.Error(), "cannot be moved") {
		t.Errorf("error = %q, want 'cannot be moved'", err.Error())
	}
}

func TestMoveNotFound(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	moveToFlag = "STORY-01"

	err := runMove(moveCmd, []string{"TASK-999"})
	if err == nil {
		t.Fatal("expected error")
	}
}
