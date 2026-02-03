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
	moveToFlag = testStory2ID

	// Move TASK-03 from STORY-01 to STORY-02
	err := runMove(moveCmd, []string{testTask3ID})
	if err != nil {
		t.Fatalf("runMove: %v", err)
	}

	// Old path should not exist
	oldPath := filepath.Join(bd, testEpic1ID+"_recording", testStory1ID+"_audio-capture", testTask3ID+"_tests")
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("old path still exists")
	}

	// New path should exist
	newPath := filepath.Join(bd, testEpic1ID+"_recording", testStory2ID+"_amplitude", testTask3ID+"_tests")
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("new path does not exist")
	}
}

func TestMoveStoryToEpic(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	moveToFlag = testEpic2ID

	err := runMove(moveCmd, []string{testStory2ID})
	if err != nil {
		t.Fatalf("runMove: %v", err)
	}

	newPath := filepath.Join(bd, testEpic2ID+"_storage", testStory2ID+"_amplitude")
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("story not found in new epic")
	}
}

func TestMoveInvalidType(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	moveToFlag = testTask1ID

	// Try to move task to task — should fail
	err := runMove(moveCmd, []string{testTask3ID})
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
	moveToFlag = testEpic2ID

	err := runMove(moveCmd, []string{testEpic1ID})
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
	moveToFlag = testStory1ID

	err := runMove(moveCmd, []string{"TASK-999"})
	if err == nil {
		t.Fatal("expected error")
	}
}
