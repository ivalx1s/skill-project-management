package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateCleanBoard(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	err := runValidate(validateCmd, nil)
	if err != nil {
		t.Fatalf("runValidate: %v", err)
	}
}

func TestValidateMissingReadme(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	// Remove a README.md
	readmePath := filepath.Join(bd, "EPIC-01_recording", "STORY-01_audio-capture", "TASK-03_tests", "README.md")
	os.Remove(readmePath)

	// Should not error, just report issues
	err := runValidate(validateCmd, nil)
	if err != nil {
		t.Fatalf("runValidate: %v", err)
	}
}

func TestValidateBrokenLink(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	// Add a broken blockedBy reference
	taskDir := filepath.Join(bd, "EPIC-01_recording", "STORY-01_audio-capture", "TASK-03_tests")
	os.WriteFile(filepath.Join(taskDir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- TASK-999\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	err := runValidate(validateCmd, nil)
	if err != nil {
		t.Fatalf("runValidate: %v", err)
	}
}
