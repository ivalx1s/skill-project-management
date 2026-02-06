package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.RefreshRate != DefaultRefreshRate {
		t.Errorf("expected RefreshRate=%d, got %d", DefaultRefreshRate, cfg.RefreshRate)
	}

	if cfg.ExpandedNodes == nil {
		t.Error("expected ExpandedNodes to be initialized, got nil")
	}

	if len(cfg.ExpandedNodes) != 0 {
		t.Errorf("expected empty ExpandedNodes, got %v", cfg.ExpandedNodes)
	}

	if cfg.ScrollSensitivity != DefaultScrollSensitivity {
		t.Errorf("expected ScrollSensitivity=%v, got %v", DefaultScrollSensitivity, cfg.ScrollSensitivity)
	}
}

func TestGetRefreshDuration(t *testing.T) {
	tests := []struct {
		name         string
		refreshRate  int
		wantDuration time.Duration
	}{
		{"default rate", 10, 10 * time.Second},
		{"5 seconds", 5, 5 * time.Second},
		{"30 seconds", 30, 30 * time.Second},
		{"disabled (0)", 0, 0},
		{"negative", -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{RefreshRate: tt.refreshRate}
			got := cfg.GetRefreshDuration()
			if got != tt.wantDuration {
				t.Errorf("GetRefreshDuration() = %v, want %v", got, tt.wantDuration)
			}
		})
	}
}

func TestLoadConfigFromPath_NonExistent(t *testing.T) {
	// Load from non-existent path should return defaults
	cfg, err := LoadConfigFromPath("/nonexistent/path/config.json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if cfg.RefreshRate != DefaultRefreshRate {
		t.Errorf("expected default RefreshRate=%d, got %d", DefaultRefreshRate, cfg.RefreshRate)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "board-tui-config-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	// Create config with custom values
	original := &Config{
		RefreshRate:   30,
		ExpandedNodes: []string{"EPIC-001", "STORY-002", "TASK-003"},
	}

	// Save
	if err := original.SaveConfigToPath(configPath); err != nil {
		t.Fatalf("SaveConfigToPath failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load
	loaded, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFromPath failed: %v", err)
	}

	// Verify values
	if loaded.RefreshRate != original.RefreshRate {
		t.Errorf("RefreshRate: got %d, want %d", loaded.RefreshRate, original.RefreshRate)
	}

	if len(loaded.ExpandedNodes) != len(original.ExpandedNodes) {
		t.Errorf("ExpandedNodes length: got %d, want %d", len(loaded.ExpandedNodes), len(original.ExpandedNodes))
	}

	for i, id := range original.ExpandedNodes {
		if loaded.ExpandedNodes[i] != id {
			t.Errorf("ExpandedNodes[%d]: got %s, want %s", i, loaded.ExpandedNodes[i], id)
		}
	}
}

func TestSaveConfig_CreatesDirectory(t *testing.T) {
	// Create temp directory with nested path that doesn't exist
	tmpDir, err := os.MkdirTemp("", "board-tui-config-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use a nested path that doesn't exist
	configPath := filepath.Join(tmpDir, "nested", "dir", "config.json")

	cfg := DefaultConfig()
	if err := cfg.SaveConfigToPath(configPath); err != nil {
		t.Fatalf("SaveConfigToPath failed to create nested directory: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created in nested directory")
	}
}

func TestExpandedNodeOperations(t *testing.T) {
	cfg := DefaultConfig()

	// Test AddExpandedNode
	cfg.AddExpandedNode("EPIC-001")
	if !cfg.IsExpanded("EPIC-001") {
		t.Error("EPIC-001 should be expanded after AddExpandedNode")
	}

	// Test adding duplicate (should not add twice)
	cfg.AddExpandedNode("EPIC-001")
	count := 0
	for _, id := range cfg.ExpandedNodes {
		if id == "EPIC-001" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("EPIC-001 should appear only once, got %d", count)
	}

	// Test adding another node
	cfg.AddExpandedNode("STORY-001")
	if !cfg.IsExpanded("STORY-001") {
		t.Error("STORY-001 should be expanded after AddExpandedNode")
	}

	// Test RemoveExpandedNode
	cfg.RemoveExpandedNode("EPIC-001")
	if cfg.IsExpanded("EPIC-001") {
		t.Error("EPIC-001 should not be expanded after RemoveExpandedNode")
	}

	// STORY-001 should still be expanded
	if !cfg.IsExpanded("STORY-001") {
		t.Error("STORY-001 should still be expanded")
	}

	// Test removing non-existent node (should not panic)
	cfg.RemoveExpandedNode("NONEXISTENT")
}

func TestSetExpandedNodes(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AddExpandedNode("OLD-001")

	newNodes := []string{"NEW-001", "NEW-002"}
	cfg.SetExpandedNodes(newNodes)

	if len(cfg.ExpandedNodes) != 2 {
		t.Errorf("expected 2 expanded nodes, got %d", len(cfg.ExpandedNodes))
	}

	if cfg.IsExpanded("OLD-001") {
		t.Error("OLD-001 should not be expanded after SetExpandedNodes")
	}

	if !cfg.IsExpanded("NEW-001") || !cfg.IsExpanded("NEW-002") {
		t.Error("NEW-001 and NEW-002 should be expanded")
	}
}

func TestCollectExpandedNodes(t *testing.T) {
	// Build a test tree
	task1 := &TreeNode{ID: "TASK-001", Expanded: false}
	task2 := &TreeNode{ID: "TASK-002", Expanded: true} // Unusual but valid
	story1 := &TreeNode{
		ID:       "STORY-001",
		Expanded: true,
		Children: []*TreeNode{task1, task2},
	}
	epic1 := &TreeNode{
		ID:       "EPIC-001",
		Expanded: true,
		Children: []*TreeNode{story1},
	}
	epic2 := &TreeNode{
		ID:       "EPIC-002",
		Expanded: false,
	}

	roots := []*TreeNode{epic1, epic2}
	expanded := CollectExpandedNodes(roots)

	// Should have EPIC-001, STORY-001, TASK-002
	expectedSet := map[string]bool{
		"EPIC-001":  true,
		"STORY-001": true,
		"TASK-002":  true,
	}

	if len(expanded) != len(expectedSet) {
		t.Errorf("expected %d expanded nodes, got %d: %v", len(expectedSet), len(expanded), expanded)
	}

	for _, id := range expanded {
		if !expectedSet[id] {
			t.Errorf("unexpected expanded node: %s", id)
		}
	}
}

func TestApplyExpandedNodes(t *testing.T) {
	// Build a test tree (all default to collapsed)
	task1 := &TreeNode{ID: "TASK-001", Expanded: false}
	task2 := &TreeNode{ID: "TASK-002", Expanded: false}
	story1 := &TreeNode{
		ID:       "STORY-001",
		Expanded: false,
		Children: []*TreeNode{task1, task2},
	}
	epic1 := &TreeNode{
		ID:       "EPIC-001",
		Expanded: false, // Will be set by initializeNode normally, but testing override
		Children: []*TreeNode{story1},
	}

	roots := []*TreeNode{epic1}
	expandedIDs := []string{"EPIC-001", "TASK-002"} // Only these should be expanded

	ApplyExpandedNodes(roots, expandedIDs)

	if !epic1.Expanded {
		t.Error("EPIC-001 should be expanded")
	}
	if story1.Expanded {
		t.Error("STORY-001 should not be expanded")
	}
	if task1.Expanded {
		t.Error("TASK-001 should not be expanded")
	}
	if !task2.Expanded {
		t.Error("TASK-002 should be expanded")
	}
}

func TestLoadConfigFromPath_InvalidJSON(t *testing.T) {
	// Create temp file with invalid JSON
	tmpDir, err := os.MkdirTemp("", "board-tui-config-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configPath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg, err := LoadConfigFromPath(configPath)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}

	// Should still return default config
	if cfg.RefreshRate != DefaultRefreshRate {
		t.Errorf("expected default RefreshRate=%d on error, got %d", DefaultRefreshRate, cfg.RefreshRate)
	}
}

func TestClampScrollSensitivity(t *testing.T) {
	if ClampScrollSensitivity(0.01) != MinScrollSensitivity {
		t.Fatalf("expected lower clamp to %v", MinScrollSensitivity)
	}
	if ClampScrollSensitivity(2.0) != MaxScrollSensitivity {
		t.Fatalf("expected upper clamp to %v", MaxScrollSensitivity)
	}
	if ClampScrollSensitivity(0.85) != 0.85 {
		t.Fatalf("expected in-range sensitivity to be unchanged")
	}
}
