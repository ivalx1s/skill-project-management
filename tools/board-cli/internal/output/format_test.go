package output

import (
	"strings"
	"testing"
)

func TestColorStatus(t *testing.T) {
	tests := []struct {
		status string
		color  string
	}{
		{"open", Blue},
		{"progress", Yellow},
		{"done", Green},
		{"closed", Gray},
		{"blocked", Red},
		{"unknown", Reset},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := ColorStatus(tt.status)
			if !strings.Contains(result, tt.status) {
				t.Errorf("ColorStatus(%q) = %q, missing status text", tt.status, result)
			}
			if !strings.Contains(result, tt.color) {
				t.Errorf("ColorStatus(%q) missing expected color code", tt.status)
			}
			if !strings.HasSuffix(result, Reset) {
				t.Errorf("ColorStatus(%q) missing Reset suffix", tt.status)
			}
		})
	}
}

func TestStatusColor(t *testing.T) {
	if StatusColor("open") != Blue {
		t.Error("open should be blue")
	}
	if StatusColor("progress") != Yellow {
		t.Error("progress should be yellow")
	}
	if StatusColor("done") != Green {
		t.Error("done should be green")
	}
	if StatusColor("closed") != Gray {
		t.Error("closed should be gray")
	}
	if StatusColor("blocked") != Red {
		t.Error("blocked should be red")
	}
	if StatusColor("unknown") != Reset {
		t.Error("unknown should be reset")
	}
}

func TestTableBasic(t *testing.T) {
	table := NewTable("ID", "NAME")
	table.AddRow("TASK-01", "Interface")
	table.AddRow("TASK-02", "Implementation")

	result := table.String()
	if !strings.Contains(result, "ID") {
		t.Error("missing header ID")
	}
	if !strings.Contains(result, "NAME") {
		t.Error("missing header NAME")
	}
	if !strings.Contains(result, "TASK-01") {
		t.Error("missing row TASK-01")
	}
	if !strings.Contains(result, "Interface") {
		t.Error("missing row Interface")
	}
	if !strings.Contains(result, "Implementation") {
		t.Error("missing row Implementation")
	}

	// Should have separator line
	lines := strings.Split(result, "\n")
	if len(lines) < 4 {
		t.Errorf("expected at least 4 lines (header + separator + 2 rows), got %d", len(lines))
	}
}

func TestTableEmptyRows(t *testing.T) {
	table := NewTable("A", "B")
	result := table.String()
	if !strings.Contains(result, "A") {
		t.Error("missing header")
	}
	// Just header + separator
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines for empty table, got %d", len(lines))
	}
}

func TestTableWithAnsi(t *testing.T) {
	table := NewTable("ID", "STATUS")
	table.AddRow("TASK-01", Yellow+"progress"+Reset)
	table.AddRow("TASK-02", Green+"done"+Reset)

	result := table.String()
	// Columns should be aligned despite ANSI codes
	if !strings.Contains(result, "progress") {
		t.Error("missing progress")
	}
	if !strings.Contains(result, "done") {
		t.Error("missing done")
	}
}

func TestStripAnsi(t *testing.T) {
	input := Red + "hello" + Reset + " world"
	result := stripAnsi(input)
	if result != "hello world" {
		t.Errorf("stripAnsi = %q, want 'hello world'", result)
	}
}

func TestStripAnsiNoAnsi(t *testing.T) {
	input := "plain text"
	result := stripAnsi(input)
	if result != "plain text" {
		t.Errorf("stripAnsi = %q, want 'plain text'", result)
	}
}
