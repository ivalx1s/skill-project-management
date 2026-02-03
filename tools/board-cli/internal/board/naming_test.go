package board

import (
	"testing"
)

func TestParseDirName(t *testing.T) {
	tests := []struct {
		input    string
		wantType ElementType
		wantNum  int
		wantName string
		wantErr  bool
	}{
		{"EPIC-260203-a1b2c3_recording", EpicType, 0, "recording", false},
		{"STORY-260203-d4e5f6_audio-capture", StoryType, 0, "audio-capture", false},
		{"TASK-260203-g7h8i9_audiorecorder-interface", TaskType, 0, "audiorecorder-interface", false},
		{"BUG-260203-m3n4o5_some-bug", BugType, 0, "some-bug", false},
		{"EPIC-01_recording", EpicType, 1, "recording", false},
		{"TASK-12_name", TaskType, 12, "name", false},
		{"invalid", "", 0, "", true},
		{"README.md", "", 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotType, gotNum, gotName, err := ParseDirName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDirName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotType != tt.wantType {
					t.Errorf("type = %v, want %v", gotType, tt.wantType)
				}
				if gotNum != tt.wantNum {
					t.Errorf("num = %v, want %v", gotNum, tt.wantNum)
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
		wantNum  int
		wantErr  bool
	}{
		{"EPIC-260203-a1b2c3", EpicType, 0, false},
		{"STORY-260203-d4e5f6", StoryType, 0, false},
		{"TASK-260203-g7h8i9", TaskType, 0, false},
		{"BUG-260203-m3n4o5", BugType, 0, false},
		{"task-260203-g7h8i9", TaskType, 0, false}, // case insensitive
		{"EPIC-01", EpicType, 1, false},
		{"TASK-12", TaskType, 12, false},
		{"invalid", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotType, gotNum, err := ParseID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotType != tt.wantType {
					t.Errorf("type = %v, want %v", gotType, tt.wantType)
				}
				if gotNum != tt.wantNum {
					t.Errorf("num = %v, want %v", gotNum, tt.wantNum)
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
