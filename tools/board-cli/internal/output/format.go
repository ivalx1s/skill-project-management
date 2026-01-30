package output

import (
	"fmt"
	"strings"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Gray   = "\033[90m"
)

func StatusColor(status string) string {
	switch status {
	case "open":
		return Blue
	case "progress":
		return Yellow
	case "done":
		return Green
	case "closed":
		return Gray
	case "blocked":
		return Red
	default:
		return Reset
	}
}

func ColorStatus(status string) string {
	return StatusColor(status) + status + Reset
}

// Table prints a simple aligned table.
type Table struct {
	Headers []string
	Rows    [][]string
	widths  []int
}

func NewTable(headers ...string) *Table {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	return &Table{Headers: headers, widths: widths}
}

func (t *Table) AddRow(cols ...string) {
	row := make([]string, len(t.Headers))
	for i := 0; i < len(t.Headers) && i < len(cols); i++ {
		row[i] = cols[i]
		// Strip ANSI for width calculation
		stripped := stripAnsi(cols[i])
		if len(stripped) > t.widths[i] {
			t.widths[i] = len(stripped)
		}
	}
	t.Rows = append(t.Rows, row)
}

func (t *Table) String() string {
	var b strings.Builder

	// Header
	for i, h := range t.Headers {
		if i > 0 {
			b.WriteString("  ")
		}
		fmt.Fprintf(&b, "%-*s", t.widths[i], h)
	}
	b.WriteString("\n")

	// Separator
	for i, w := range t.widths {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(strings.Repeat("-", w))
	}
	b.WriteString("\n")

	// Rows
	for _, row := range t.Rows {
		for i, col := range row {
			if i > 0 {
				b.WriteString("  ")
			}
			stripped := stripAnsi(col)
			padding := t.widths[i] - len(stripped)
			b.WriteString(col)
			if padding > 0 {
				b.WriteString(strings.Repeat(" ", padding))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func stripAnsi(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}
