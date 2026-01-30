package templates

import (
	"strings"
	"testing"
)

func TestRenderReadmeEpic(t *testing.T) {
	content, err := RenderReadme("epic", TemplateData{
		Title:       "EPIC-01: Recording",
		Description: "Recording epic description",
	})
	if err != nil {
		t.Fatalf("RenderReadme epic: %v", err)
	}
	if !strings.Contains(content, "EPIC-01: Recording") {
		t.Error("missing title")
	}
	if !strings.Contains(content, "Recording epic description") {
		t.Error("missing description")
	}
	if !strings.Contains(content, "## Description") {
		t.Error("missing Description section")
	}
}

func TestRenderReadmeStory(t *testing.T) {
	content, err := RenderReadme("story", TemplateData{
		Title:       "STORY-05: Audio Capture",
		Description: "Capture audio from microphone",
	})
	if err != nil {
		t.Fatalf("RenderReadme story: %v", err)
	}
	if !strings.Contains(content, "STORY-05: Audio Capture") {
		t.Error("missing title")
	}
}

func TestRenderReadmeTask(t *testing.T) {
	content, err := RenderReadme("task", TemplateData{
		Title:       "TASK-12: Interface",
		Description: "Define interface",
	})
	if err != nil {
		t.Fatalf("RenderReadme task: %v", err)
	}
	if !strings.Contains(content, "TASK-12: Interface") {
		t.Error("missing title")
	}
}

func TestRenderReadmeBug(t *testing.T) {
	content, err := RenderReadme("bug", TemplateData{
		Title:       "BUG-01: Crash",
		Description: "App crashes on start",
	})
	if err != nil {
		t.Fatalf("RenderReadme bug: %v", err)
	}
	if !strings.Contains(content, "BUG-01: Crash") {
		t.Error("missing title")
	}
}

func TestRenderReadmeInvalidType(t *testing.T) {
	_, err := RenderReadme("widget", TemplateData{Title: "test"})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestRenderProgressEpic(t *testing.T) {
	content, err := RenderProgress("epic")
	if err != nil {
		t.Fatalf("RenderProgress epic: %v", err)
	}
	if !strings.Contains(content, "## Status") {
		t.Error("missing Status section")
	}
	if !strings.Contains(content, "## Blocked By") {
		t.Error("missing Blocked By section")
	}
	if !strings.Contains(content, "## Checklist") {
		t.Error("missing Checklist section")
	}
	if !strings.Contains(content, "## Notes") {
		t.Error("missing Notes section")
	}
}

func TestRenderProgressStory(t *testing.T) {
	content, err := RenderProgress("story")
	if err != nil {
		t.Fatalf("RenderProgress story: %v", err)
	}
	if !strings.Contains(content, "## Status") {
		t.Error("missing Status section")
	}
}

func TestRenderProgressTask(t *testing.T) {
	content, err := RenderProgress("task")
	if err != nil {
		t.Fatalf("RenderProgress task: %v", err)
	}
	if !strings.Contains(content, "open") {
		t.Error("missing default open status")
	}
}

func TestRenderProgressBug(t *testing.T) {
	content, err := RenderProgress("bug")
	if err != nil {
		t.Fatalf("RenderProgress bug: %v", err)
	}
	if !strings.Contains(content, "## Status") {
		t.Error("missing Status section")
	}
}

func TestRenderProgressInvalidType(t *testing.T) {
	_, err := RenderProgress("widget")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}
