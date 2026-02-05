package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RefreshOption represents a selectable refresh rate option
type RefreshOption struct {
	Label    string
	Duration time.Duration // 0 means "Off"
}

// SettingsModel is the bubbletea model for the settings screen
type SettingsModel struct {
	options  []RefreshOption
	cursor   int                 // Current selection cursor
	selected int                 // Currently selected/applied option
	width    int                 // Terminal width
	height   int                 // Terminal height
	onSave   func(time.Duration) // Callback when settings are saved
}

// Styles for settings screen
var (
	settingsTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#6C5CE7")).
				Padding(0, 1).
				MarginBottom(1)

	settingsHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Bold(true).
				MarginTop(1).
				MarginBottom(1)

	settingsOptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#C1C6B2")).
				PaddingLeft(2)

	settingsSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6C5CE7")).
				Bold(true).
				PaddingLeft(2)

	settingsCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Bold(true)

	settingsHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262")).
				MarginTop(2)

	settingsContainerStyle = lipgloss.NewStyle().
				Padding(1, 2)
)

// DefaultRefreshOptions returns the available refresh rate options
func DefaultRefreshOptions() []RefreshOption {
	return []RefreshOption{
		{Label: "5 seconds", Duration: 5 * time.Second},
		{Label: "10 seconds (default)", Duration: 10 * time.Second},
		{Label: "30 seconds", Duration: 30 * time.Second},
		{Label: "60 seconds", Duration: 60 * time.Second},
		{Label: "Off (manual only)", Duration: 0},
	}
}

// NewSettingsModel creates a new settings model
func NewSettingsModel(currentInterval time.Duration, onSave func(time.Duration)) SettingsModel {
	options := DefaultRefreshOptions()

	// Find the currently selected option based on the interval
	selected := 1 // Default to "10 seconds"
	for i, opt := range options {
		if opt.Duration == currentInterval {
			selected = i
			break
		}
	}

	return SettingsModel{
		options:  options,
		cursor:   selected,
		selected: selected,
		onSave:   onSave,
	}
}

// Init initializes the settings model
func (m SettingsModel) Init() tea.Cmd {
	return nil
}

// SettingsCloseMsg is sent when the settings screen should close
type SettingsCloseMsg struct {
	NewInterval time.Duration
	Changed     bool
}

// Update handles input for the settings model
func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter", " ":
			// Select the option under cursor
			oldSelected := m.selected
			m.selected = m.cursor
			if m.onSave != nil {
				m.onSave(m.options[m.selected].Duration)
			}
			// Return close message with the new interval
			return m, func() tea.Msg {
				return SettingsCloseMsg{
					NewInterval: m.options[m.selected].Duration,
					Changed:     oldSelected != m.selected,
				}
			}
		case "esc":
			// Close without changing (return current selection)
			return m, func() tea.Msg {
				return SettingsCloseMsg{
					NewInterval: m.options[m.selected].Duration,
					Changed:     false,
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the settings screen
func (m SettingsModel) View() string {
	var s string

	// Title
	s += settingsTitleStyle.Render("Settings") + "\n\n"

	// Section header
	s += settingsHeaderStyle.Render("Refresh Rate") + "\n"

	// Radio button options
	for i, opt := range m.options {
		// Determine the radio button state
		var radioBtn string
		if i == m.selected {
			radioBtn = "●" // Selected (filled)
		} else {
			radioBtn = "○" // Not selected (empty)
		}

		// Build the line
		var line string
		if i == m.cursor {
			// Cursor is on this line - highlight it
			radioBtn = settingsCursorStyle.Render(radioBtn)
			line = settingsSelectedStyle.Render(radioBtn + " " + opt.Label)
		} else if i == m.selected {
			// This is the selected option but cursor isn't here
			line = settingsSelectedStyle.Render(radioBtn + " " + opt.Label)
		} else {
			// Regular unselected option
			line = settingsOptionStyle.Render(radioBtn + " " + opt.Label)
		}

		s += line + "\n"
	}

	// Help text
	s += settingsHelpStyle.Render("\n  ↑/↓ navigate • Enter/Space select • Esc back")

	return settingsContainerStyle.Render(s)
}

// GetSelectedDuration returns the currently selected refresh duration
func (m SettingsModel) GetSelectedDuration() time.Duration {
	if m.selected >= 0 && m.selected < len(m.options) {
		return m.options[m.selected].Duration
	}
	return defaultRefreshInterval
}

// SetSize updates the terminal size for the settings view
func (m *SettingsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
