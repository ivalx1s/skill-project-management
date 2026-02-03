package cmd

import (
	"testing"

	"github.com/aagrigore/task-board/internal/board"
)

func TestProgressNotesAppend(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	progressNotesSet = false

	// TASK-01 already has notes "Started work"
	err := runProgressNotes(progressNotesCmd, []string{testTask1ID, "More work done"})
	if err != nil {
		t.Fatalf("runProgressNotes: %v", err)
	}

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	elem := b.FindByID(testTask1ID)
	pd, err := board.ParseProgressFile(elem.ProgressPath())
	if err != nil {
		t.Fatalf("ParseProgressFile: %v", err)
	}

	if pd.Notes != "Started work\nMore work done" {
		t.Errorf("notes = %q, want 'Started work\\nMore work done'", pd.Notes)
	}
}

func TestProgressNotesSet(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	progressNotesSet = true
	defer func() { progressNotesSet = false }()

	err := runProgressNotes(progressNotesCmd, []string{testTask1ID, "Replaced notes"})
	if err != nil {
		t.Fatalf("runProgressNotes: %v", err)
	}

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	elem := b.FindByID(testTask1ID)
	pd, err := board.ParseProgressFile(elem.ProgressPath())
	if err != nil {
		t.Fatalf("ParseProgressFile: %v", err)
	}

	if pd.Notes != "Replaced notes" {
		t.Errorf("notes = %q, want 'Replaced notes'", pd.Notes)
	}
}

func TestProgressNotesEmpty(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	progressNotesSet = false

	// TASK-03 has no notes
	err := runProgressNotes(progressNotesCmd, []string{testTask3ID, "First note"})
	if err != nil {
		t.Fatalf("runProgressNotes: %v", err)
	}

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	elem := b.FindByID(testTask3ID)
	pd, err := board.ParseProgressFile(elem.ProgressPath())
	if err != nil {
		t.Fatalf("ParseProgressFile: %v", err)
	}

	if pd.Notes != "First note" {
		t.Errorf("notes = %q, want 'First note'", pd.Notes)
	}
}
