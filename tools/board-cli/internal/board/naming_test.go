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
		{"EPIC-01_recording", EpicType, 1, "recording", false},
		{"STORY-05_audio-capture", StoryType, 5, "audio-capture", false},
		{"TASK-12_audiorecorder-interface", TaskType, 12, "audiorecorder-interface", false},
		{"BUG-47_some-bug", BugType, 47, "some-bug", false},
		{"invalid", "", 0, "", true},
		{"EPIC-_nonum", "", 0, "", true},
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
		{"EPIC-01", EpicType, 1, false},
		{"STORY-05", StoryType, 5, false},
		{"TASK-12", TaskType, 12, false},
		{"BUG-47", BugType, 47, false},
		{"task-12", TaskType, 12, false}, // case insensitive
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
