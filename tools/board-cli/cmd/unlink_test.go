package cmd

import (
	"strings"
	"testing"

	"github.com/aagrigore/task-board/internal/board"
)

func TestUnlink(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	unlinkBlockedBy = testTask1ID

	// TASK-02 is blocked by TASK-01
	err := runUnlink(unlinkCmd, []string{testTask2ID})
	if err != nil {
		t.Fatalf("runUnlink: %v", err)
	}

	// Verify TASK-02 no longer blocked
	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	task2 := b.FindByID(testTask2ID)
	if len(task2.BlockedBy) > 0 {
		t.Errorf("TASK-02 still blocked by: %v", task2.BlockedBy)
	}

	// Verify TASK-01 no longer blocks TASK-02
	task1 := b.FindByID(testTask1ID)
	for _, bid := range task1.Blocks {
		if bid == testTask2ID {
			t.Error("TASK-01 still blocks TASK-02")
		}
	}
}

func TestUnlinkNotBlocked(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	unlinkBlockedBy = testTask3ID

	// TASK-01 is not blocked by TASK-03
	err := runUnlink(unlinkCmd, []string{testTask1ID})
	if err == nil {
		t.Fatal("expected error for non-existent link")
	}
	if !strings.Contains(err.Error(), "not blocked by") {
		t.Errorf("error = %q, want 'not blocked by'", err.Error())
	}
}

func TestUnlinkNotFound(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	unlinkBlockedBy = testTask1ID

	err := runUnlink(unlinkCmd, []string{"TASK-999"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnlinkDeescalatesCrossStory(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	// First, create cross-epic link: TASK-04 blocked by TASK-01
	linkBlockedBy = testTask1ID
	err := runLink(linkCmd, []string{testTask4ID})
	if err != nil {
		t.Fatalf("runLink: %v", err)
	}

	// Verify escalation happened
	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	story3 := b.FindByID(testStory3ID)
	if len(story3.BlockedBy) == 0 {
		t.Fatal("STORY-03 should be blocked after link")
	}

	// Now unlink
	unlinkBlockedBy = testTask1ID
	err = runUnlink(unlinkCmd, []string{testTask4ID})
	if err != nil {
		t.Fatalf("runUnlink: %v", err)
	}

	// Verify de-escalation: STORY-03 should no longer be blocked
	b, err = board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	story3 = b.FindByID(testStory3ID)
	for _, bid := range story3.BlockedBy {
		if bid == testStory1ID {
			t.Error("STORY-03 still blocked by STORY-01 after de-escalation")
		}
	}

	// EPIC-02 should no longer be blocked
	epic2 := b.FindByID(testEpic2ID)
	for _, bid := range epic2.BlockedBy {
		if bid == testEpic1ID {
			t.Error("EPIC-02 still blocked by EPIC-01 after de-escalation")
		}
	}
}

func TestUnlinkKeepsEscalationWhenOtherCrossLinksExist(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	// Create two cross-story links within same epic:
	// TASK-03 (STORY-01) blocked by TASK-01 (STORY-01) — same parent, no escalation
	// So we need another task in STORY-02 for this test.
	// Actually, let's use cross-epic: TASK-04 (STORY-03/EPIC-02) blocked by TASK-01 AND TASK-02

	linkBlockedBy = testTask1ID
	if err := runLink(linkCmd, []string{testTask4ID}); err != nil {
		t.Fatalf("link1: %v", err)
	}
	linkBlockedBy = testTask2ID
	if err := runLink(linkCmd, []string{testTask4ID}); err != nil {
		t.Fatalf("link2: %v", err)
	}

	// Remove one link — escalation should stay (TASK-04 still blocked by TASK-02)
	unlinkBlockedBy = testTask1ID
	if err := runUnlink(unlinkCmd, []string{testTask4ID}); err != nil {
		t.Fatalf("unlink: %v", err)
	}

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// STORY-03 should STILL be blocked by STORY-01 (TASK-04 still blocked by TASK-02 which is in STORY-01)
	story3 := b.FindByID(testStory3ID)
	foundStory := false
	for _, bid := range story3.BlockedBy {
		if bid == testStory1ID {
			foundStory = true
		}
	}
	if !foundStory {
		t.Errorf("STORY-03 should still be blocked by STORY-01, got: %v", story3.BlockedBy)
	}
}
