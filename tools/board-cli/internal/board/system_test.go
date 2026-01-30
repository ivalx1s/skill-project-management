package board

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCounters(t *testing.T) {
	content := `## Counters
- epic: 9
- story: 43
- task: 156
- bug: 11
`
	c, err := parseCounters(content)
	if err != nil {
		t.Fatal(err)
	}
	if c.Epic != 9 {
		t.Errorf("Epic = %d, want 9", c.Epic)
	}
	if c.Story != 43 {
		t.Errorf("Story = %d, want 43", c.Story)
	}
	if c.Task != 156 {
		t.Errorf("Task = %d, want 156", c.Task)
	}
	if c.Bug != 11 {
		t.Errorf("Bug = %d, want 11", c.Bug)
	}
}

func TestCountersIncrement(t *testing.T) {
	c := &Counters{Epic: 5, Story: 10, Task: 20, Bug: 3}
	got := c.Increment(TaskType)
	if got != 21 {
		t.Errorf("Increment(Task) = %d, want 21", got)
	}
	if c.Task != 21 {
		t.Errorf("Task = %d, want 21", c.Task)
	}
}

func TestReadWriteCounters(t *testing.T) {
	dir := t.TempDir()
	boardDir := filepath.Join(dir, ".task-board")
	os.MkdirAll(boardDir, 0755)

	original := &Counters{Epic: 3, Story: 7, Task: 42, Bug: 5}
	if err := WriteCounters(boardDir, original); err != nil {
		t.Fatal(err)
	}

	loaded, err := ReadCounters(boardDir)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.Epic != 3 || loaded.Story != 7 || loaded.Task != 42 || loaded.Bug != 5 {
		t.Errorf("got %+v, want %+v", loaded, original)
	}
}
