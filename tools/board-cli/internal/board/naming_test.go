package board

import (
	"testing"
)

func TestParseDirName(t *testing.T) {
	tests := []struct {
		input    string
		wantType ElementType
		wantName string
		wantErr  bool
	}{
		{"EPIC-260203-a1b2c3_recording", EpicType, "recording", false},
		{"STORY-260203-d4e5f6_audio-capture", StoryType, "audio-capture", false},
		{"TASK-260203-g7h8i9_audiorecorder-interface", TaskType, "audiorecorder-interface", false},
		{"BUG-260203-m3n4o5_some-bug", BugType, "some-bug", false},
		{"invalid", "", "", true},
		{"EPIC-01_recording", "", "", true},          // old format not supported
		{"TASK-12_name", "", "", true},               // old format not supported
		{"README.md", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotType, _, gotName, err := ParseDirName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDirName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotType != tt.wantType {
					t.Errorf("type = %v, want %v", gotType, tt.wantType)
				}
				if gotName != tt.wantName {
					t.Errorf("name = %q, want %q", gotName, tt.wantName)
				}
			}
		})
	}
}

func TestParseID(t *testing.T) {
	tests := []struct {
		input    string
		wantType ElementType
		wantErr  bool
	}{
		{"EPIC-260203-a1b2c3", EpicType, false},
		{"STORY-260203-d4e5f6", StoryType, false},
		{"TASK-260203-g7h8i9", TaskType, false},
		{"BUG-260203-m3n4o5", BugType, false},
		{"task-260203-g7h8i9", TaskType, false}, // case insensitive
		{"EPIC-01", "", true},                    // old format not supported
		{"TASK-12", "", true},                    // old format not supported
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotType, _, err := ParseID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotType != tt.wantType {
					t.Errorf("type = %v, want %v", gotType, tt.wantType)
				}
			}
		})
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Audio Capture", "audio-capture"},
		{"some_thing", "some-thing"},
		{"Hello World!", "hello-world"},
		{"multiple---hyphens", "multiple-hyphens"},
		{" leading trailing ", "leading-trailing"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SanitizeName(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
