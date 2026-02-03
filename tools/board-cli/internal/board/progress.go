package board

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// ProgressData holds parsed progress.md content.
type ProgressData struct {
	Status     Status
	AssignedTo string
	CreatedAt  time.Time
	LastUpdate time.Time
	BlockedBy  []string
	Blocks     []string
	Checklist  []ChecklistItem
	Notes      string
}

// ParseProgressFile reads and parses a progress.md file.
func ParseProgressFile(path string) (*ProgressData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading progress file: %w", err)
	}
	return ParseProgress(string(data))
}

// ParseProgress parses progress.md content.
func ParseProgress(content string) (*ProgressData, error) {
	pd := &ProgressData{
		Status: StatusBacklog,
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	currentSection := ""

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Detect sections
		if strings.HasPrefix(trimmed, "## ") {
			currentSection = strings.ToLower(strings.TrimPrefix(trimmed, "## "))
			continue
		}

		switch currentSection {
		case "status":
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				s, err := ParseStatus(trimmed)
				if err == nil {
					pd.Status = s
				}
			}
		case "assigned to":
			if trimmed != "" && trimmed != "(none)" {
				pd.AssignedTo = trimmed
			}
		case "created":
			if trimmed != "" {
				t, err := time.Parse(time.RFC3339, trimmed)
				if err == nil {
					pd.CreatedAt = t
				}
			}
		case "last update":
			if trimmed != "" {
				t, err := time.Parse(time.RFC3339, trimmed)
				if err == nil {
					pd.LastUpdate = t
				}
			}
		case "blocked by", "blockedby":
			if strings.HasPrefix(trimmed, "- ") {
				id := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
				if id != "" && id != "(none)" {
					pd.BlockedBy = append(pd.BlockedBy, id)
				}
			}
		case "blocks":
			if strings.HasPrefix(trimmed, "- ") {
				id := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
				if id != "" && id != "(none)" {
					pd.Blocks = append(pd.Blocks, id)
				}
			}
		case "checklist":
			if strings.HasPrefix(trimmed, "- [x] ") {
				text := strings.TrimPrefix(trimmed, "- [x] ")
				pd.Checklist = append(pd.Checklist, ChecklistItem{Text: text, Checked: true})
			} else if strings.HasPrefix(trimmed, "- [ ] ") {
				text := strings.TrimPrefix(trimmed, "- [ ] ")
				pd.Checklist = append(pd.Checklist, ChecklistItem{Text: text, Checked: false})
			}
		case "notes":
			if trimmed != "" {
				if pd.Notes != "" {
					pd.Notes += "\n"
				}
				pd.Notes += trimmed
			}
		}
	}

	return pd, scanner.Err()
}

// WriteProgress writes progress.md content.
func WriteProgress(pd *ProgressData) string {
	var b strings.Builder

	b.WriteString("## Status\n")
	b.WriteString(string(pd.Status))
	b.WriteString("\n\n")

	b.WriteString("## Assigned To\n")
	if pd.AssignedTo == "" {
		b.WriteString("(none)\n")
	} else {
		b.WriteString(pd.AssignedTo)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	b.WriteString("## Created\n")
	if !pd.CreatedAt.IsZero() {
		b.WriteString(pd.CreatedAt.UTC().Format(time.RFC3339))
	}
	b.WriteString("\n\n")

	b.WriteString("## Last Update\n")
	b.WriteString(pd.LastUpdate.UTC().Format(time.RFC3339))
	b.WriteString("\n\n")

	b.WriteString("## Blocked By\n")
	if len(pd.BlockedBy) == 0 {
		b.WriteString("- (none)\n")
	} else {
		for _, id := range pd.BlockedBy {
			fmt.Fprintf(&b, "- %s\n", id)
		}
	}
	b.WriteString("\n")

	b.WriteString("## Blocks\n")
	if len(pd.Blocks) == 0 {
		b.WriteString("- (none)\n")
	} else {
		for _, id := range pd.Blocks {
			fmt.Fprintf(&b, "- %s\n", id)
		}
	}
	b.WriteString("\n")

	b.WriteString("## Checklist\n")
	if len(pd.Checklist) == 0 {
		b.WriteString("(empty)\n")
	} else {
		for _, item := range pd.Checklist {
			if item.Checked {
				fmt.Fprintf(&b, "- [x] %s\n", item.Text)
			} else {
				fmt.Fprintf(&b, "- [ ] %s\n", item.Text)
			}
		}
	}
	b.WriteString("\n")

	b.WriteString("## Notes\n")
	if pd.Notes != "" {
		b.WriteString(pd.Notes)
		b.WriteString("\n")
	}

	return b.String()
}

// WriteProgressFile writes progress data to a file.
// Automatically updates LastUpdate to current time.
func WriteProgressFile(path string, pd *ProgressData) error {
	pd.LastUpdate = time.Now().UTC()
	content := WriteProgress(pd)
	return os.WriteFile(path, []byte(content), 0644)
}
