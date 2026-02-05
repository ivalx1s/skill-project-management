package board

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Directory format: EPIC-260203-abc123_recording (distributed ID)
var dirPattern = regexp.MustCompile(`^(EPIC|STORY|TASK|BUG)-(\d{6})-([0-9a-z]{6})_(.+)$`)

// Legacy directory format: EPIC-01_recording (sequential ID)
var legacyDirPattern = regexp.MustCompile(`^(EPIC|STORY|TASK|BUG)-(\d+)_([^/]+)$`)

// ParseDirName parses a directory name into type, rawID, and name.
// Supported formats:
//   - TYPE-YYMMDD-xxxxxx_name (distributed ID)
//   - TYPE-NN_name (legacy sequential ID)
//
// Returns: type, number (legacy only), name, error
func ParseDirName(dirName string) (ElementType, int, string, error) {
	matches := dirPattern.FindStringSubmatch(dirName)
	if matches != nil {
		prefix := matches[1]
		elemType := prefixToType(prefix)
		if elemType == "" {
			return "", 0, "", fmt.Errorf("unknown prefix: %s", prefix)
		}

		// Return 0 for number (not used with distributed IDs)
		return elemType, 0, matches[4], nil
	}

	legacyMatches := legacyDirPattern.FindStringSubmatch(dirName)
	if legacyMatches == nil {
		return "", 0, "", fmt.Errorf("invalid element directory name: %s (expected TYPE-YYMMDD-xxxxxx_name or TYPE-NN_name)", dirName)
	}

	prefix := legacyMatches[1]
	elemType := prefixToType(prefix)
	if elemType == "" {
		return "", 0, "", fmt.Errorf("unknown prefix: %s", prefix)
	}

	num, err := strconv.Atoi(legacyMatches[2])
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid element number in directory name: %s", dirName)
	}

	return elemType, num, legacyMatches[3], nil
}

// ExtractRawID extracts the full ID from a directory name.
// E.g., "TASK-260101-aaaaaa_interface" -> "TASK-260101-aaaaaa"
func ExtractRawID(dirName string) string {
	matches := dirPattern.FindStringSubmatch(dirName)
	if matches != nil {
		return fmt.Sprintf("%s-%s-%s", matches[1], matches[2], matches[3])
	}

	legacyMatches := legacyDirPattern.FindStringSubmatch(dirName)
	if legacyMatches == nil {
		return ""
	}
	return fmt.Sprintf("%s-%s", legacyMatches[1], legacyMatches[2])
}

// ID format: TASK-260203-abc123 (case-insensitive matching, pattern uses lowercase)
var idPattern = regexp.MustCompile(`^(epic|story|task|bug)-(\d{6})-([0-9a-z]{6})$`)

// Legacy ID format: TASK-12
var legacyIDPattern = regexp.MustCompile(`^(epic|story|task|bug)-(\d+)$`)

// ParseID parses an element ID into type.
// Supported formats:
//   - TYPE-YYMMDD-xxxxxx (distributed ID)
//   - TYPE-NN (legacy sequential ID)
func ParseID(id string) (ElementType, int, error) {
	matches := idPattern.FindStringSubmatch(strings.ToLower(id))
	if matches != nil {
		elemType := prefixToType(matches[1])
		if elemType == "" {
			return "", 0, fmt.Errorf("unknown prefix in ID: %s", id)
		}

		// Return 0 for number (not used with distributed IDs)
		return elemType, 0, nil
	}

	legacyMatches := legacyIDPattern.FindStringSubmatch(strings.ToLower(id))
	if legacyMatches == nil {
		return "", 0, fmt.Errorf("invalid element ID: %s (expected TYPE-YYMMDD-xxxxxx or TYPE-NN)", id)
	}

	elemType := prefixToType(legacyMatches[1])
	if elemType == "" {
		return "", 0, fmt.Errorf("unknown prefix in ID: %s", id)
	}

	num, err := strconv.Atoi(legacyMatches[2])
	if err != nil {
		return "", 0, fmt.Errorf("invalid element number in ID: %s", id)
	}

	return elemType, num, nil
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
