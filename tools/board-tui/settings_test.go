package main

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDefaultRefreshOptions(t *testing.T) {
	options := DefaultRefreshOptions()

	if len(options) != 5 {
		t.Errorf("expected 5 refresh options, got %d", len(options))
	}

	// Verify expected durations
	expectedDurations := []time.Duration{
		5 * time.Second,
		10 * time.Second,
		30 * time.Second,
		60 * time.Second,
		0, // Off
	}

	for i, expected := range expectedDurations {
		if options[i].Duration != expected {
			t.Errorf("option %d: expected duration %v, got %v", i, expected, options[i].Duration)
		}
	}
}

func TestNewSettingsModel(t *testing.T) {
	tests := []struct {
		name             string
		currentInterval  time.Duration
		expectedSelected int
	}{
		{"5 seconds", 5 * time.Second, 0},
		{"10 seconds (default)", 10 * time.Second, 1},
		{"30 seconds", 30 * time.Second, 2},
		{"60 seconds", 60 * time.Second, 3},
		{"Off", 0, 4},
		{"Unknown defaults to 10s", 15 * time.Second, 1}, // Unknown interval defaults to index 1 (10s)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewSettingsModel(tc.currentInterval, nil)

			if m.selected != tc.expectedSelected {
				t.Errorf("expected selected=%d, got %d", tc.expectedSelected, m.selected)
			}
			if m.cursor != tc.expectedSelected {
				t.Errorf("expected cursor=%d, got %d", tc.expectedSelected, m.cursor)
			}
		})
	}
}

func TestSettingsModelNavigation(t *testing.T) {
	m := NewSettingsModel(10*time.Second, nil)

	// Initial state: cursor at index 1 (10 seconds)
	if m.cursor != 1 {
		t.Fatalf("expected initial cursor=1, got %d", m.cursor)
	}

	// Move down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 2 {
		t.Errorf("after down: expected cursor=2, got %d", m.cursor)
	}

	// Move down with 'j'
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 3 {
		t.Errorf("after j: expected cursor=3, got %d", m.cursor)
	}

	// Move up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 2 {
		t.Errorf("after up: expected cursor=2, got %d", m.cursor)
	}

	// Move up with 'k'
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 1 {
		t.Errorf("after k: expected cursor=1, got %d", m.cursor)
	}

	// Try to move up past top
	m.cursor = 0
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 0 {
		t.Errorf("at top, after up: expected cursor=0, got %d", m.cursor)
	}

	// Try to move down past bottom
	m.cursor = 4
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 4 {
		t.Errorf("at bottom, after down: expected cursor=4, got %d", m.cursor)
	}
}

func TestSettingsModelSelection(t *testing.T) {
	var savedDuration time.Duration
	onSave := func(d time.Duration) {
		savedDuration = d
	}

	m := NewSettingsModel(10*time.Second, onSave)

	// Move to 30 seconds option (index 2)
	m.cursor = 2

	// Select with Enter
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.selected != 2 {
		t.Errorf("expected selected=2, got %d", m.selected)
	}

	if savedDuration != 30*time.Second {
		t.Errorf("expected savedDuration=30s, got %v", savedDuration)
	}

	// Check that the command returns SettingsCloseMsg
	if cmd != nil {
		msg := cmd()
		closeMsg, ok := msg.(SettingsCloseMsg)
		if !ok {
			t.Errorf("expected SettingsCloseMsg, got %T", msg)
		} else {
			if closeMsg.NewInterval != 30*time.Second {
				t.Errorf("expected NewInterval=30s, got %v", closeMsg.NewInterval)
			}
			if !closeMsg.Changed {
				t.Errorf("expected Changed=true")
			}
		}
	}
}

func TestSettingsModelEsc(t *testing.T) {
	m := NewSettingsModel(10*time.Second, nil)

	// Change cursor position but don't select
	m.cursor = 3

	// Press Esc
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Selected should remain unchanged
	if m.selected != 1 {
		t.Errorf("expected selected=1 (unchanged), got %d", m.selected)
	}

	// Should return close message
	if cmd != nil {
		msg := cmd()
		closeMsg, ok := msg.(SettingsCloseMsg)
		if !ok {
			t.Errorf("expected SettingsCloseMsg, got %T", msg)
		} else {
			if closeMsg.Changed {
				t.Errorf("expected Changed=false for Esc")
			}
		}
	}
}

func TestSettingsModelView(t *testing.T) {
	m := NewSettingsModel(10*time.Second, nil)
	m.width = 80
	m.height = 24

	view := m.View()

	// Check that the view contains expected elements
	if view == "" {
		t.Error("view should not be empty")
	}

	// Should contain "Settings" title
	if !containsString(view, "Settings") {
		t.Error("view should contain 'Settings'")
	}

	// Should contain "Refresh Rate" header
	if !containsString(view, "Refresh Rate") {
		t.Error("view should contain 'Refresh Rate'")
	}

	// Should contain all options
	expectedLabels := []string{
		"5 seconds",
		"10 seconds",
		"30 seconds",
		"60 seconds",
		"Off",
	}
	for _, label := range expectedLabels {
		if !containsString(view, label) {
			t.Errorf("view should contain '%s'", label)
		}
	}

	// Should contain radio buttons (filled and empty)
	if !containsString(view, "●") {
		t.Error("view should contain filled radio button ●")
	}
	if !containsString(view, "○") {
		t.Error("view should contain empty radio button ○")
	}
}

func TestGetSelectedDuration(t *testing.T) {
	m := NewSettingsModel(30*time.Second, nil)

	dur := m.GetSelectedDuration()
	if dur != 30*time.Second {
		t.Errorf("expected 30s, got %v", dur)
	}

	// Test with invalid selected index
	m.selected = -1
	dur = m.GetSelectedDuration()
	if dur != defaultRefreshInterval {
		t.Errorf("expected default interval for invalid index, got %v", dur)
	}
}

func TestSetSize(t *testing.T) {
	m := NewSettingsModel(10*time.Second, nil)

	m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("expected width=100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("expected height=50, got %d", m.height)
	}
}

// containsString checks if s contains substr (helper for view tests)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
