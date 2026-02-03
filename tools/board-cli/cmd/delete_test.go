package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aagrigore/task-board/internal/board"
)

func TestDeleteLeaf(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	// Delete TASK-03 (leaf, no children, no deps)
	err := runDelete(deleteCmd, []string{testTask3ID})
	if err != nil {
		t.Fatalf("runDelete: %v", err)
	}

	// Verify directory removed
	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.FindByID(testTask3ID) != nil {
		t.Error("TASK-03 still exists after delete")
	}
}

func TestDeleteWithChildrenRefused(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	deleteForce = false

	err := runDelete(deleteCmd, []string{testEpic1ID})
	if err == nil {
		t.Fatal("expected error when deleting parent without --force")
	}
	if !strings.Contains(err.Error(), "children") {
		t.Errorf("error = %q, want mention of children", err.Error())
	}
}

func TestDeleteWithForce(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	deleteForce = true
	defer func() { deleteForce = false }()

	err := runDelete(deleteCmd, []string{testEpic1ID})
	if err != nil {
		t.Fatalf("runDelete --force: %v", err)
	}

	// Verify directory is gone
	epicDir := filepath.Join(bd, "EPIC-01_recording")
	if _, err := os.Stat(epicDir); !os.IsNotExist(err) {
		t.Error("EPIC-01 directory still exists after force delete")
	}
}

func TestDeleteCleansDeps(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	// TASK-01 blocks TASK-02. Delete TASK-01, verify TASK-02's blockedBy is cleaned.
	deleteForce = false
	err := runDelete(deleteCmd, []string{testTask1ID})
	if err != nil {
		t.Fatalf("runDelete: %v", err)
	}

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	task2 := b.FindByID(testTask2ID)
	if task2 == nil {
		t.Fatal("TASK-02 not found")
	}
	if len(task2.BlockedBy) > 0 {
		t.Errorf("TASK-02 still has blockedBy: %v", task2.BlockedBy)
	}
}

func TestDeleteNotFound(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	err := runDelete(deleteCmd, []string{"TASK-999"})
	if err == nil {
		t.Fatal("expected error for missing element")
	}
}
