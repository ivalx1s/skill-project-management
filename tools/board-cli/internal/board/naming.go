package board

import (
	"fmt"
	"regexp"
	"strings"
)

// Directory format: EPIC-260203-abc123_recording (distributed ID)
var dirPattern = regexp.MustCompile(`^(EPIC|STORY|TASK|BUG)-(\d{6})-([0-9a-z]{6})_(.+)$`)

// ParseDirName parses a directory name into type, rawID, and name.
// Format: TYPE-YYMMDD-xxxxxx_name
// Returns: type, 0 (legacy number), name, rawID, error
func ParseDirName(dirName string) (ElementType, int, string, error) {
	matches := dirPattern.FindStringSubmatch(dirName)
	if matches == nil {
		return "", 0, "", fmt.Errorf("invalid element directory name: %s (expected TYPE-YYMMDD-xxxxxx_name)", dirName)
	}

	prefix := matches[1]
	elemType := prefixToType(prefix)
	if elemType == "" {
		return "", 0, "", fmt.Errorf("unknown prefix: %s", prefix)
	}

	// Return 0 for number (legacy field, not used with distributed IDs)
	return elemType, 0, matches[4], nil
}

// ExtractRawID extracts the full ID from a directory name.
// E.g., "TASK-260101-aaaaaa_interface" -> "TASK-260101-aaaaaa"
func ExtractRawID(dirName string) string {
	matches := dirPattern.FindStringSubmatch(dirName)
	if matches == nil {
		return ""
	}
	return fmt.Sprintf("%s-%s-%s", matches[1], matches[2], matches[3])
}

// ID format: TASK-260203-abc123 (case-insensitive matching, pattern uses lowercase)
var idPattern = regexp.MustCompile(`^(epic|story|task|bug)-(\d{6})-([0-9a-z]{6})$`)

// ParseID parses an element ID into type.
// Format: TYPE-YYMMDD-xxxxxx
func ParseID(id string) (ElementType, int, error) {
	matches := idPattern.FindStringSubmatch(strings.ToLower(id))
	if matches == nil {
		return "", 0, fmt.Errorf("invalid element ID: %s (expected TYPE-YYMMDD-xxxxxx)", id)
	}

	elemType := prefixToType(matches[1])
	if elemType == "" {
		return "", 0, fmt.Errorf("unknown prefix in ID: %s", id)
	}

	// Return 0 for number (legacy field, not used with distributed IDs)
	return elemType, 0, nil
}

// prefixToType converts a string prefix to ElementType.
func prefixToType(prefix string) ElementType {
	switch strings.ToUpper(prefix) {
	case "EPIC":
		return EpicType
	case "STORY":
		return StoryType
	case "TASK":
		return TaskType
	case "BUG":
		return BugType
	default:
		return ""
	}
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
