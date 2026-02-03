package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aagrigore/task-board/internal/board"
)

func TestRenderOutputPathProjectLevel(t *testing.T) {
	tmpDir := t.TempDir()
	b := &board.Board{Dir: tmpDir}

	path, err := RenderOutputPath(b, "", "plan", "svg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(tmpDir, ".temp", "plan.svg")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}

	// .temp directory should have been created.
	info, err := os.Stat(filepath.Join(tmpDir, ".temp"))
	if err != nil {
		t.Fatalf(".temp dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error(".temp should be a directory")
	}
}

func TestRenderOutputPathElementLevel(t *testing.T) {
	tmpDir := t.TempDir()
	epicDir := filepath.Join(tmpDir, "EPIC-01_recording")
	if err := os.MkdirAll(epicDir, 0o755); err != nil {
		t.Fatal(err)
	}

	elem := &board.Element{
		Type:   board.EpicType,
		Number: 1,
		Name:   "recording",
		Path:   epicDir,
	}

	b := &board.Board{
		Dir:      tmpDir,
		Elements: []*board.Element{elem},
	}

	path, err := RenderOutputPath(b, "EPIC-01", "plan", "png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(epicDir, ".temp", "plan.png")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}

	// .temp directory should have been created inside epic dir.
	info, err := os.Stat(filepath.Join(epicDir, ".temp"))
	if err != nil {
		t.Fatalf(".temp dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error(".temp should be a directory")
	}
}

func TestRenderOutputPathStoryLevel(t *testing.T) {
	tmpDir := t.TempDir()
	storyDir := filepath.Join(tmpDir, "EPIC-01_recording", "STORY-05_auth")
	if err := os.MkdirAll(storyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	story := &board.Element{
		Type:     board.StoryType,
		Number:   5,
		Name:     "auth",
		Path:     storyDir,
		ParentID: "EPIC-01",
	}

	b := &board.Board{
		Dir:      tmpDir,
		Elements: []*board.Element{story},
	}

	path, err := RenderOutputPath(b, "STORY-05", "plan", "svg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(storyDir, ".temp", "plan.svg")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}
}

func TestRenderOutputPathDefaultFormat(t *testing.T) {
	tmpDir := t.TempDir()
	b := &board.Board{Dir: tmpDir}

	path, err := RenderOutputPath(b, "", "plan", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasSuffix(path, "plan.svg") {
		t.Errorf("default format should be svg, got path: %s", path)
	}
}

func TestRenderOutputPathNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	b := &board.Board{Dir: tmpDir}

	_, err := RenderOutputPath(b, "TASK-99", "plan", "svg")
	if err == nil {
		t.Error("expected error for nonexistent element")
	}
	if !strings.Contains(err.Error(), "TASK-99") {
		t.Errorf("error should mention TASK-99: %v", err)
	}
}

func TestRenderDOTMissingBinary(t *testing.T) {
	// Temporarily override PATH so dot is not found.
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "")
	defer os.Setenv("PATH", origPath)

	err := RenderDOT("digraph{}", filepath.Join(t.TempDir(), "out.svg"), "svg", "")
	if err == nil {
		t.Fatal("expected error when dot binary is missing")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
	if !strings.Contains(err.Error(), "graphviz") {
		t.Errorf("error should mention 'graphviz': %v", err)
	}
}

func TestRenderDOTDefaultFormat(t *testing.T) {
	// This test just checks that empty format defaults to "svg" by verifying
	// the function doesn't panic with empty format. The actual binary may not
	// be installed, so we allow the error from missing binary.
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "")
	defer os.Setenv("PATH", origPath)

	err := RenderDOT("digraph{}", filepath.Join(t.TempDir(), "out.svg"), "", "")
	// We expect an error (no dot binary), just checking no panic.
	if err == nil {
		t.Log("dot binary unexpectedly available — that's fine")
	}
}
