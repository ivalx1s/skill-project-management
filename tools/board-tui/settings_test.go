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

func TestDefaultAgentsOptions(t *testing.T) {
	options := DefaultAgentsOptions()

	if len(options) != 6 {
		t.Errorf("expected 6 agents options, got %d", len(options))
	}

	expectedStale := []int{0, 5, 10, 15, 30, 60}
	for i, expected := range expectedStale {
		if options[i].StaleMinutes != expected {
			t.Errorf("option %d: expected stale %d, got %d", i, expected, options[i].StaleMinutes)
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
		{"Unknown defaults to 10s", 15 * time.Second, 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewSettingsModel(tc.currentInterval, nil)

			if m.refreshSelected != tc.expectedSelected {
				t.Errorf("expected refreshSelected=%d, got %d", tc.expectedSelected, m.refreshSelected)
			}
			if m.refreshCursor != tc.expectedSelected {
				t.Errorf("expected refreshCursor=%d, got %d", tc.expectedSelected, m.refreshCursor)
			}
		})
	}
}

func TestNewSettingsModelWithAgents(t *testing.T) {
	m := NewSettingsModelWithAgents(10*time.Second, 3, nil)

	if m.agentsSelected != 3 {
		t.Errorf("expected agentsSelected=3, got %d", m.agentsSelected)
	}
	if m.agentsCursor != 3 {
		t.Errorf("expected agentsCursor=3, got %d", m.agentsCursor)
	}
}

func TestSettingsModelNavigation(t *testing.T) {
	m := NewSettingsModel(10*time.Second, nil)

	// Initial state: cursor at index 1 (10 seconds)
	if m.refreshCursor != 1 {
		t.Fatalf("expected initial refreshCursor=1, got %d", m.refreshCursor)
	}

	// Move down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.refreshCursor != 2 {
		t.Errorf("after down: expected refreshCursor=2, got %d", m.refreshCursor)
	}

	// Move down with 'j'
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.refreshCursor != 3 {
		t.Errorf("after j: expected refreshCursor=3, got %d", m.refreshCursor)
	}

	// Move up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.refreshCursor != 2 {
		t.Errorf("after up: expected refreshCursor=2, got %d", m.refreshCursor)
	}

	// Move up with 'k'
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.refreshCursor != 1 {
		t.Errorf("after k: expected refreshCursor=1, got %d", m.refreshCursor)
	}

	// Try to move up past top
	m.refreshCursor = 0
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.refreshCursor != 0 {
		t.Errorf("at top, after up: expected refreshCursor=0, got %d", m.refreshCursor)
	}

	// Try to move down past bottom
	m.refreshCursor = 4
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.refreshCursor != 4 {
		t.Errorf("at bottom, after down: expected refreshCursor=4, got %d", m.refreshCursor)
	}
}

func TestSettingsModelGroupSwitch(t *testing.T) {
	m := NewSettingsModel(10*time.Second, nil)

	if m.focusGroup != GroupRefresh {
		t.Error("initial focus should be GroupRefresh")
	}

	// Switch to agents group with Tab
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focusGroup != GroupAgents {
		t.Error("after Tab: focus should be GroupAgents")
	}

	// Switch back with shift+tab
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.focusGroup != GroupRefresh {
		t.Error("after shift+Tab: focus should be GroupRefresh")
	}

	// Switch with 'l'
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if m.focusGroup != GroupAgents {
		t.Error("after l: focus should be GroupAgents")
	}

	// Switch with 'h'
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if m.focusGroup != GroupRefresh {
		t.Error("after h: focus should be GroupRefresh")
	}
}

func TestSettingsModelSelection(t *testing.T) {
	var savedDuration time.Duration
	onSave := func(d time.Duration) {
		savedDuration = d
	}

	m := NewSettingsModel(10*time.Second, onSave)

	// Move to 30 seconds option (index 2)
	m.refreshCursor = 2

	// Select with Enter
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.refreshSelected != 2 {
		t.Errorf("expected refreshSelected=2, got %d", m.refreshSelected)
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
			if !closeMsg.RefreshChanged {
				t.Errorf("expected RefreshChanged=true")
			}
		}
	}
}

func TestSettingsModelAgentsSelection(t *testing.T) {
	m := NewSettingsModelWithAgents(10*time.Second, 0, nil)

	// Switch to agents group
	m.focusGroup = GroupAgents
	m.agentsCursor = 3 // Stale > 15 min

	// Select with Enter
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.agentsSelected != 3 {
		t.Errorf("expected agentsSelected=3, got %d", m.agentsSelected)
	}

	if cmd != nil {
		msg := cmd()
		closeMsg, ok := msg.(SettingsCloseMsg)
		if !ok {
			t.Errorf("expected SettingsCloseMsg, got %T", msg)
		} else {
			if closeMsg.NewAgentsFilter != 3 {
				t.Errorf("expected NewAgentsFilter=3, got %d", closeMsg.NewAgentsFilter)
			}
			if !closeMsg.AgentsChanged {
				t.Errorf("expected AgentsChanged=true")
			}
		}
	}
}

func TestSettingsModelEsc(t *testing.T) {
	m := NewSettingsModel(10*time.Second, nil)

	// Change cursor position but don't select
	m.refreshCursor = 3

	// Press Esc
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Selected should remain unchanged
	if m.refreshSelected != 1 {
		t.Errorf("expected refreshSelected=1 (unchanged), got %d", m.refreshSelected)
	}

	// Should return close message
	if cmd != nil {
		msg := cmd()
		closeMsg, ok := msg.(SettingsCloseMsg)
		if !ok {
			t.Errorf("expected SettingsCloseMsg, got %T", msg)
		} else {
			if closeMsg.RefreshChanged {
				t.Errorf("expected RefreshChanged=false for Esc")
			}
			if closeMsg.AgentsChanged {
				t.Errorf("expected AgentsChanged=false for Esc")
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

	// Should contain "Agents Display" header
	if !containsString(view, "Agents Display") {
		t.Error("view should contain 'Agents Display'")
	}

	// Should contain refresh options
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

	// Should contain agents options
	if !containsString(view, "All agents") {
		t.Error("view should contain 'All agents'")
	}
	if !containsString(view, "Stale > 5 min") {
		t.Error("view should contain 'Stale > 5 min'")
	}

	// Should contain radio buttons
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
	m.refreshSelected = -1
	dur = m.GetSelectedDuration()
	if dur != defaultRefreshInterval {
		t.Errorf("expected default interval for invalid index, got %v", dur)
	}
}

func TestGetSelectedAgentsFilter(t *testing.T) {
	m := NewSettingsModelWithAgents(10*time.Second, 4, nil)

	filter := m.GetSelectedAgentsFilter()
	if filter != 4 {
		t.Errorf("expected 4, got %d", filter)
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
