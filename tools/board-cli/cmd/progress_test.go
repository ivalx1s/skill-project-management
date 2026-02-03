package cmd

import (
	"strings"
	"testing"

	"github.com/aagrigore/task-board/internal/board"
)

func TestProgressStatusSet(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	err := runProgressStatus(progressStatusCmd, []string{testTask3ID, "development"})
	if err != nil {
		t.Fatalf("runProgressStatus: %v", err)
	}

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	task3 := b.FindByID(testTask3ID)
	if task3.Status != board.StatusDevelopment {
		t.Errorf("TASK-03 status = %v, want development", task3.Status)
	}
}

func TestProgressStatusBlockedPrevented(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	// TASK-02 is blocked by TASK-01 (which is to-dev)
	err := runProgressStatus(progressStatusCmd, []string{testTask2ID, "development"})
	if err == nil {
		t.Fatal("expected error for blocked element")
	}
	if !strings.Contains(err.Error(), "blocked by") {
		t.Errorf("error = %q, want 'blocked by'", err.Error())
	}
}

func TestProgressCheckAndUncheck(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	// Check item 1 of TASK-01
	err := toggleChecklistItem(testTask1ID, "1", true)
	if err != nil {
		t.Fatalf("check: %v", err)
	}

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	task1 := b.FindByID(testTask1ID)
	if !task1.Checklist[0].Checked {
		t.Error("item 1 should be checked")
	}

	// Uncheck item 2
	boardDir = bd
	err = toggleChecklistItem(testTask1ID, "2", false)
	if err != nil {
		t.Fatalf("uncheck: %v", err)
	}

	b, err = board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	task1 = b.FindByID(testTask1ID)
	if task1.Checklist[1].Checked {
		t.Error("item 2 should be unchecked")
	}
}

func TestProgressCheckOutOfRange(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	err := toggleChecklistItem(testTask1ID, "99", true)
	if err == nil {
		t.Fatal("expected error for out of range")
	}
}

func TestProgressAddItem(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	err := runProgressAddItem(progressAddItemCmd, []string{testTask1ID, "New step"})
	if err != nil {
		t.Fatalf("runProgressAddItem: %v", err)
	}

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	task1 := b.FindByID(testTask1ID)
	if len(task1.Checklist) != 3 {
		t.Errorf("checklist len = %d, want 3", len(task1.Checklist))
	}
	if task1.Checklist[2].Text != "New step" {
		t.Errorf("item text = %q, want 'New step'", task1.Checklist[2].Text)
	}
}
