package board

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ReadmeData holds parsed README.md content.
type ReadmeData struct {
	Title       string
	Description string
	Scope       string
	AC          string // acceptance criteria
}

// ParseReadmeFile reads and parses a README.md file.
func ParseReadmeFile(path string) (*ReadmeData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading readme file: %w", err)
	}
	return ParseReadme(string(data))
}

// ParseReadme parses README.md content.
func ParseReadme(content string) (*ReadmeData, error) {
	rd := &ReadmeData{}

	scanner := bufio.NewScanner(strings.NewReader(content))
	currentSection := ""
	var sectionContent strings.Builder

	flushSection := func() {
		text := strings.TrimSpace(sectionContent.String())
		switch currentSection {
		case "title":
			// Title is set from ## line, not content
		case "description":
			rd.Description = text
		case "scope":
			rd.Scope = text
		case "acceptance criteria":
			rd.AC = text
		}
		sectionContent.Reset()
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// H1 = title
		if strings.HasPrefix(trimmed, "# ") && !strings.HasPrefix(trimmed, "## ") {
			rd.Title = strings.TrimPrefix(trimmed, "# ")
			currentSection = "title"
			continue
		}

		// H2 = section
		if strings.HasPrefix(trimmed, "## ") {
			flushSection()
			currentSection = strings.ToLower(strings.TrimPrefix(trimmed, "## "))
			continue
		}

		if currentSection != "" && currentSection != "title" {
			sectionContent.WriteString(line)
			sectionContent.WriteString("\n")
		}
	}
	flushSection()

	return rd, scanner.Err()
}

// WriteReadme generates README.md content.
func WriteReadme(rd *ReadmeData) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", rd.Title)
	fmt.Fprintf(&b, "## Description\n%s\n\n", rd.Description)
	fmt.Fprintf(&b, "## Scope\n%s\n\n", rd.Scope)
	fmt.Fprintf(&b, "## Acceptance Criteria\n%s\n", rd.AC)
	return b.String()
}

// WriteReadmeFile writes readme data to a file.
func WriteReadmeFile(path string, rd *ReadmeData) error {
	content := WriteReadme(rd)
	return os.WriteFile(path, []byte(content), 0644)
}
