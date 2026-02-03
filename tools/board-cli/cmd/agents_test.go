package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aagrigore/task-board/internal/board"
)

// writeAssignee writes a progress.md with assignee, status, and last update.
func writeAssignee(t *testing.T, dir, agent string, status board.Status, lastUpdate time.Time) {
	t.Helper()
	content := fmt.Sprintf("## Status\n%s\n\n## Assigned To\n%s\n\n## Created\n%s\n\n## Last Update\n%s\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n",
		string(status),
		agent,
		lastUpdate.UTC().Format(time.RFC3339),
		lastUpdate.UTC().Format(time.RFC3339),
	)
	err := os.WriteFile(filepath.Join(dir, "progress.md"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("writeAssignee: %v", err)
	}
}

func TestAgentsEmptyDashboard(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	agentsAll = false

	out := captureOutput(t, func() {
		err := runAgents(agentsCmd, nil)
		if err != nil {
			t.Fatalf("runAgents: %v", err)
		}
	})

	if !strings.Contains(out, "No active agents.") {
		t.Fatalf("expected 'No active agents.', got: %s", out)
	}
}

func TestAgentsOneAgent(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	agentsAll = false

	// Assign agent-1 to STORY-01 with recent update
	storyDir := filepath.Join(bd, testEpic1ID+"_recording", testStory1ID+"_audio-capture")
	writeAssignee(t, storyDir, "agent-1", board.StatusDevelopment, time.Now().UTC())

	out := captureOutput(t, func() {
		err := runAgents(agentsCmd, nil)
		if err != nil {
			t.Fatalf("runAgents: %v", err)
		}
	})

	if !strings.Contains(out, "Sub-Agent Dashboard") {
		t.Fatalf("expected dashboard header, got: %s", out)
	}
	if !strings.Contains(out, "agent-1") {
		t.Fatalf("expected agent-1 in output, got: %s", out)
	}
	if !strings.Contains(out, testStory1ID) {
		t.Fatalf("expected STORY-01 in output, got: %s", out)
	}
	if !strings.Contains(out, "Total: 1 agents, 1 active, 0 done") {
		t.Fatalf("expected footer with 1 agent, got: %s", out)
	}
}

func TestAgentsAllShowsDone(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	// Assign agent-1 to STORY-01 with done status and old update (should be filtered without --all)
	storyDir := filepath.Join(bd, testEpic1ID+"_recording", testStory1ID+"_audio-capture")
	writeAssignee(t, storyDir, "agent-1", board.StatusDone, time.Now().UTC().Add(-2*time.Hour))

	// Without --all: should be filtered (done + old update)
	agentsAll = false
	out := captureOutput(t, func() {
		err := runAgents(agentsCmd, nil)
		if err != nil {
			t.Fatalf("runAgents: %v", err)
		}
	})
	if !strings.Contains(out, "No active agents.") {
		t.Fatalf("expected 'No active agents.' without --all, got: %s", out)
	}

	// With --all: should show
	agentsAll = true
	out = captureOutput(t, func() {
		err := runAgents(agentsCmd, nil)
		if err != nil {
			t.Fatalf("runAgents: %v", err)
		}
	})
	if !strings.Contains(out, "agent-1") {
		t.Fatalf("expected agent-1 with --all, got: %s", out)
	}
	if !strings.Contains(out, "1 done") {
		t.Fatalf("expected '1 done' in footer, got: %s", out)
	}
}

func TestAgentsFreshnessFilter(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	agentsAll = false

	// Done + fresh (within 30 min) — should show
	storyDir := filepath.Join(bd, testEpic1ID+"_recording", testStory1ID+"_audio-capture")
	writeAssignee(t, storyDir, "agent-fresh", board.StatusDone, time.Now().UTC().Add(-10*time.Minute))

	out := captureOutput(t, func() {
		err := runAgents(agentsCmd, nil)
		if err != nil {
			t.Fatalf("runAgents: %v", err)
		}
	})

	if !strings.Contains(out, "agent-fresh") {
		t.Fatalf("expected fresh done agent to show, got: %s", out)
	}
}

func TestHumanTime(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		input    time.Time
		expected string
	}{
		{time.Time{}, "-"},
		{now.Add(-30 * time.Second), "30 sec ago"},
		{now.Add(-5 * time.Minute), "5 min ago"},
		{now.Add(-3 * time.Hour), "3 hours ago"},
		{now.Add(-48 * time.Hour), "2 days ago"},
	}

	for _, tt := range tests {
		got := humanTime(tt.input, now)
		if got != tt.expected {
			t.Errorf("humanTime(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestChildProgress(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	// Set TASK-01 and TASK-02 to done, TASK-03 stays open
	task1Dir := filepath.Join(bd, testEpic1ID+"_recording", testStory1ID+"_audio-capture", testTask1ID+"_interface")
	task2Dir := filepath.Join(bd, testEpic1ID+"_recording", testStory1ID+"_audio-capture", testTask2ID+"_impl")

	writeAssignee(t, task1Dir, "", board.StatusDone, time.Now().UTC())
	writeAssignee(t, task2Dir, "", board.StatusDone, time.Now().UTC())

	// Assign agent to STORY-01
	storyDir := filepath.Join(bd, testEpic1ID+"_recording", testStory1ID+"_audio-capture")
	writeAssignee(t, storyDir, "agent-1", board.StatusDevelopment, time.Now().UTC())

	out := captureOutput(t, func() {
		agentsAll = false
		err := runAgents(agentsCmd, nil)
		if err != nil {
			t.Fatalf("runAgents: %v", err)
		}
	})

	// STORY-01 has children: TASK-01(done), TASK-02(done), TASK-03(open), BUG-01(open) = 2/4 done
	if !strings.Contains(out, "2/4 done") {
		t.Fatalf("expected '2/4 done' progress, got: %s", out)
	}
}

func TestChildProgressNoChildren(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	// Assign agent to TASK-01 (a leaf task, no children)
	taskDir := filepath.Join(bd, testEpic1ID+"_recording", testStory1ID+"_audio-capture", testTask1ID+"_interface")
	writeAssignee(t, taskDir, "agent-leaf", board.StatusDevelopment, time.Now().UTC())

	agentsAll = false
	out := captureOutput(t, func() {
		err := runAgents(agentsCmd, nil)
		if err != nil {
			t.Fatalf("runAgents: %v", err)
		}
	})

	// Task has no children, should show its own status
	if !strings.Contains(out, "development") {
		t.Fatalf("expected 'development' for leaf task, got: %s", out)
	}
}
