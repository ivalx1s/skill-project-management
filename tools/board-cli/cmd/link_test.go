package cmd

import (
	"testing"

	"github.com/aagrigore/task-board/internal/board"
)

func TestLinkCreatesDepBidirectional(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	linkBlockedBy = testTask1ID

	// TASK-03 is not blocked by anything. Link it.
	err := runLink(linkCmd, []string{testTask3ID})
	if err != nil {
		t.Fatalf("runLink: %v", err)
	}

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// TASK-03 should be blocked by TASK-01
	task3 := b.FindByID(testTask3ID)
	found := false
	for _, bid := range task3.BlockedBy {
		if bid == testTask1ID {
			found = true
		}
	}
	if !found {
		t.Errorf("TASK-03 blockedBy = %v, want TASK-01", task3.BlockedBy)
	}

	// TASK-01 should block TASK-03
	task1 := b.FindByID(testTask1ID)
	found = false
	for _, bid := range task1.Blocks {
		if bid == testTask3ID {
			found = true
		}
	}
	if !found {
		t.Errorf("TASK-01 blocks = %v, want TASK-03", task1.Blocks)
	}
}

func TestLinkDuplicate(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	linkBlockedBy = testTask1ID

	// TASK-02 is already blocked by TASK-01
	err := runLink(linkCmd, []string{testTask2ID})
	if err != nil {
		t.Fatalf("runLink duplicate: %v", err)
	}

	// Should not duplicate
	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	task2 := b.FindByID(testTask2ID)
	count := 0
	for _, bid := range task2.BlockedBy {
		if bid == testTask1ID {
			count++
		}
	}
	if count != 1 {
		t.Errorf("TASK-01 appears %d times in TASK-02 blockedBy, want 1", count)
	}
}

func TestLinkNotFound(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	linkBlockedBy = "TASK-999"

	err := runLink(linkCmd, []string{testTask1ID})
	if err == nil {
		t.Fatal("expected error for missing blocker")
	}
}

func TestLinkEscalatesCrossStory(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	linkBlockedBy = testTask1ID // TASK-01 is in STORY-01 (EPIC-01)

	// TASK-04 is in STORY-03 (EPIC-02) — different story AND different epic
	err := runLink(linkCmd, []string{testTask4ID})
	if err != nil {
		t.Fatalf("runLink: %v", err)
	}

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// STORY-03 should be blocked by STORY-01
	story3 := b.FindByID(testStory3ID)
	if story3 == nil {
		t.Fatal("STORY-03 not found")
	}
	foundStory := false
	for _, bid := range story3.BlockedBy {
		if bid == testStory1ID {
			foundStory = true
		}
	}
	if !foundStory {
		t.Errorf("STORY-03 blockedBy = %v, want STORY-01", story3.BlockedBy)
	}

	// EPIC-02 should be blocked by EPIC-01
	epic2 := b.FindByID(testEpic2ID)
	if epic2 == nil {
		t.Fatal("EPIC-02 not found")
	}
	foundEpic := false
	for _, bid := range epic2.BlockedBy {
		if bid == testEpic1ID {
			foundEpic = true
		}
	}
	if !foundEpic {
		t.Errorf("EPIC-02 blockedBy = %v, want EPIC-01", epic2.BlockedBy)
	}
}

func TestLinkSameParentNoEscalation(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	linkBlockedBy = testTask1ID // TASK-01 in STORY-01

	// TASK-03 is also in STORY-01 — same parent, no escalation expected
	err := runLink(linkCmd, []string{testTask3ID})
	if err != nil {
		t.Fatalf("runLink: %v", err)
	}

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// STORY-01 should NOT have any new blocked-by (was none before)
	story1 := b.FindByID(testStory1ID)
	for _, bid := range story1.BlockedBy {
		if bid != "(none)" {
			t.Errorf("STORY-01 should not be blocked by anything, got: %v", story1.BlockedBy)
			break
		}
	}
}
