package board

import (
	"regexp"
	"testing"
	"time"
)

func TestGenerateIDFormat(t *testing.T) {
	tests := []struct {
		typ    ElementType
		prefix string
	}{
		{EpicType, "EPIC"},
		{StoryType, "STORY"},
		{TaskType, "TASK"},
		{BugType, "BUG"},
	}

	// Format: TYPE-YYMMDD-xxxxxx
	pattern := regexp.MustCompile(`^[A-Z]+-\d{6}-[0-9a-z]{6}$`)

	for _, tt := range tests {
		t.Run(string(tt.typ), func(t *testing.T) {
			id := GenerateID(tt.typ)

			if !pattern.MatchString(id) {
				t.Errorf("GenerateID(%s) = %q, doesn't match pattern TYPE-YYMMDD-xxxx", tt.typ, id)
			}

			// Check prefix
			if id[:len(tt.prefix)] != tt.prefix {
				t.Errorf("GenerateID(%s) = %q, expected prefix %s", tt.typ, id, tt.prefix)
			}
		})
	}
}

func TestGenerateIDDatePart(t *testing.T) {
	id := GenerateID(TaskType)

	// Extract date part (positions 5-10 after "TASK-")
	datePart := id[5:11]

	// Should match today's date in YYMMDD format
	expected := time.Now().Format("060102")
	if datePart != expected {
		t.Errorf("date part = %q, want %q", datePart, expected)
	}
}

func TestGenerateIDHashPartBase36(t *testing.T) {
	id := GenerateID(TaskType)

	// Extract hash part (last 6 characters)
	hashPart := id[len(id)-6:]

	// Should only contain base36 characters (0-9, a-z)
	for _, c := range hashPart {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z')) {
			t.Errorf("hash part %q contains invalid character %c", hashPart, c)
		}
	}
}

func TestGenerateIDUniqueness(t *testing.T) {
	// Generate many IDs rapidly and check for uniqueness
	ids := make(map[string]bool)
	const count = 1000

	for i := 0; i < count; i++ {
		id := GenerateID(TaskType)
		if ids[id] {
			t.Errorf("duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestToBase36(t *testing.T) {
	tests := []struct {
		input uint64
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{35, "z"},
		{36, "10"},
		{1295, "zz"},
		{1296, "100"},
	}

	for _, tt := range tests {
		got := toBase36(tt.input)
		if got != tt.want {
			t.Errorf("toBase36(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGenerateHashSuffixLength(t *testing.T) {
	// Should always return exactly 6 characters
	for i := 0; i < 100; i++ {
		suffix := generateHashSuffix()
		if len(suffix) != 6 {
			t.Errorf("generateHashSuffix returned %q (len=%d), want length 6", suffix, len(suffix))
		}
	}
}

func TestTypePrefixUnknown(t *testing.T) {
	prefix := typePrefix(ElementType("unknown"))
	if prefix != "ITEM" {
		t.Errorf("typePrefix(unknown) = %q, want \"ITEM\"", prefix)
	}
}
