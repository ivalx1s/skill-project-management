package board

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var elementDirPattern = regexp.MustCompile(`^(EPIC|STORY|TASK|BUG)-(\d+)_(.+)$`)

// ParseDirName parses a directory name like "EPIC-01_recording" into type, number, name.
func ParseDirName(dirName string) (ElementType, int, string, error) {
	matches := elementDirPattern.FindStringSubmatch(dirName)
	if matches == nil {
		return "", 0, "", fmt.Errorf("invalid element directory name: %s", dirName)
	}

	prefix := matches[1]
	numStr := matches[2]
	name := matches[3]

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid number in directory name %s: %w", dirName, err)
	}

	var elemType ElementType
	switch prefix {
	case "EPIC":
		elemType = EpicType
	case "STORY":
		elemType = StoryType
	case "TASK":
		elemType = TaskType
	case "BUG":
		elemType = BugType
	default:
		return "", 0, "", fmt.Errorf("unknown prefix: %s", prefix)
	}

	return elemType, num, name, nil
}

var idPattern = regexp.MustCompile(`^(EPIC|STORY|TASK|BUG)-(\d+)$`)

// ParseID parses an element ID like "TASK-12" into type and number.
func ParseID(id string) (ElementType, int, error) {
	matches := idPattern.FindStringSubmatch(strings.ToUpper(id))
	if matches == nil {
		return "", 0, fmt.Errorf("invalid element ID: %s (expected format: TYPE-NN)", id)
	}

	prefix := matches[1]
	num, _ := strconv.Atoi(matches[2])

	var elemType ElementType
	switch prefix {
	case "EPIC":
		elemType = EpicType
	case "STORY":
		elemType = StoryType
	case "TASK":
		elemType = TaskType
	case "BUG":
		elemType = BugType
	}

	return elemType, num, nil
}

// SanitizeName converts a human name to a valid directory component.
func SanitizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.TrimSpace(name)
	// Replace spaces and underscores with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	// Remove anything that's not alphanumeric or hyphen
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	// Collapse multiple hyphens
	s := result.String()
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	return s
}
