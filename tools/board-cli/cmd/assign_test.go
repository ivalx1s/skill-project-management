package cmd

import (
	"testing"
	"time"

	"github.com/aagrigore/task-board/internal/board"
)

func TestAssignSetsAssignedTo(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	assignAgent = "agent-1"

	err := runAssign(assignCmd, []string{"STORY-03"})
	if err != nil {
		t.Fatalf("runAssign: %v", err)
	}

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	elem := b.FindByID("STORY-03")
	if elem.AssignedTo != "agent-1" {
		t.Errorf("AssignedTo = %q, want %q", elem.AssignedTo, "agent-1")
	}
}

func TestAssignUpdatesLastUpdate(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	assignAgent = "agent-2"

	before := time.Now().UTC().Add(-time.Second)

	err := runAssign(assignCmd, []string{"TASK-01"})
	if err != nil {
		t.Fatalf("runAssign: %v", err)
	}

	after := time.Now().UTC().Add(time.Second)

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	elem := b.FindByID("TASK-01")
	if elem.LastUpdate.Before(before) || elem.LastUpdate.After(after) {
		t.Errorf("LastUpdate = %v, want between %v and %v", elem.LastUpdate, before, after)
	}
}

func TestUnassignClearsAssignedTo(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	// First assign
	assignAgent = "agent-1"
	err := runAssign(assignCmd, []string{"STORY-03"})
	if err != nil {
		t.Fatalf("runAssign: %v", err)
	}

	// Then unassign
	err = runUnassign(unassignCmd, []string{"STORY-03"})
	if err != nil {
		t.Fatalf("runUnassign: %v", err)
	}

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	elem := b.FindByID("STORY-03")
	if elem.AssignedTo != "" {
		t.Errorf("AssignedTo = %q, want empty", elem.AssignedTo)
	}
}

func TestUnassignOnUnassignedElement(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	// STORY-03 is not assigned to anyone by default
	err := runUnassign(unassignCmd, []string{"STORY-03"})
	if err != nil {
		t.Fatalf("runUnassign: %v", err)
	}
	// Should print "not assigned to anyone" but not error
}

func TestAssignNotFoundElement(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	assignAgent = "agent-1"

	err := runAssign(assignCmd, []string{"STORY-999"})
	if err == nil {
		t.Fatal("expected error for missing element")
	}
}
