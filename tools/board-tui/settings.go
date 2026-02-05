package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RefreshOption represents a selectable option with duration
type RefreshOption struct {
	Label    string
	Duration time.Duration // 0 means "Off"
}

// AgentsFilterOption represents an agents filter option
type AgentsFilterOption struct {
	Label       string
	StaleMinutes int // 0 = all
}

// SettingsGroup represents which settings group is focused
type SettingsGroup int

const (
	GroupRefresh SettingsGroup = iota
	GroupAgents
)

// SettingsModel is the bubbletea model for the settings screen
type SettingsModel struct {
	// Refresh rate settings
	refreshOptions  []RefreshOption
	refreshCursor   int
	refreshSelected int

	// Agents filter settings
	agentsOptions  []AgentsFilterOption
	agentsCursor   int
	agentsSelected int

	// UI state
	focusGroup SettingsGroup
	width      int
	height     int

	// Callback
	onSave func(time.Duration)
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

	settingsHeaderInactiveStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#808080")).
					Bold(true).
					MarginTop(1).
					MarginBottom(1)

	settingsOptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#C1C6B2")).
				PaddingLeft(2)

	settingsOptionInactiveStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#626262")).
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

// DefaultAgentsOptions returns the available agents filter options
func DefaultAgentsOptions() []AgentsFilterOption {
	return []AgentsFilterOption{
		{Label: "All agents", StaleMinutes: 0},
		{Label: "Stale > 5 min", StaleMinutes: 5},
		{Label: "Stale > 10 min", StaleMinutes: 10},
		{Label: "Stale > 15 min", StaleMinutes: 15},
		{Label: "Stale > 30 min", StaleMinutes: 30},
		{Label: "Stale > 60 min", StaleMinutes: 60},
	}
}

// NewSettingsModel creates a new settings model
func NewSettingsModel(currentInterval time.Duration, onSave func(time.Duration)) SettingsModel {
	return NewSettingsModelWithAgents(currentInterval, 0, onSave)
}

// NewSettingsModelWithAgents creates a new settings model with agents filter
func NewSettingsModelWithAgents(currentInterval time.Duration, agentsFilter int, onSave func(time.Duration)) SettingsModel {
	refreshOptions := DefaultRefreshOptions()
	agentsOptions := DefaultAgentsOptions()

	// Find selected refresh option
	refreshSelected := 1 // Default to "10 seconds"
	for i, opt := range refreshOptions {
		if opt.Duration == currentInterval {
			refreshSelected = i
			break
		}
	}

	// Agents filter selection
	agentsSelected := agentsFilter
	if agentsSelected >= len(agentsOptions) {
		agentsSelected = 0
	}

	return SettingsModel{
		refreshOptions:  refreshOptions,
		refreshCursor:   refreshSelected,
		refreshSelected: refreshSelected,
		agentsOptions:   agentsOptions,
		agentsCursor:    agentsSelected,
		agentsSelected:  agentsSelected,
		focusGroup:      GroupRefresh,
		onSave:          onSave,
	}
}

// Init initializes the settings model
func (m SettingsModel) Init() tea.Cmd {
	return nil
}

// SettingsCloseMsg is sent when the settings screen should close
type SettingsCloseMsg struct {
	NewInterval     time.Duration
	NewAgentsFilter int
	RefreshChanged  bool
	AgentsChanged   bool
}

// Update handles input for the settings model
func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.moveCursorUp()
		case "down", "j":
			m.moveCursorDown()
		case "tab", "right", "l":
			m.nextGroup()
		case "shift+tab", "left", "h":
			m.prevGroup()
		case "enter", " ":
			return m.selectCurrent()
		case "esc", "q":
			return m, func() tea.Msg {
				return SettingsCloseMsg{
					NewInterval:     m.refreshOptions[m.refreshSelected].Duration,
					NewAgentsFilter: m.agentsSelected,
					RefreshChanged:  false,
					AgentsChanged:   false,
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m *SettingsModel) moveCursorUp() {
	switch m.focusGroup {
	case GroupRefresh:
		if m.refreshCursor > 0 {
			m.refreshCursor--
		}
	case GroupAgents:
		if m.agentsCursor > 0 {
			m.agentsCursor--
		}
	}
}

func (m *SettingsModel) moveCursorDown() {
	switch m.focusGroup {
	case GroupRefresh:
		if m.refreshCursor < len(m.refreshOptions)-1 {
			m.refreshCursor++
		}
	case GroupAgents:
		if m.agentsCursor < len(m.agentsOptions)-1 {
			m.agentsCursor++
		}
	}
}

func (m *SettingsModel) nextGroup() {
	if m.focusGroup == GroupRefresh {
		m.focusGroup = GroupAgents
	}
}

func (m *SettingsModel) prevGroup() {
	if m.focusGroup == GroupAgents {
		m.focusGroup = GroupRefresh
	}
}

func (m SettingsModel) selectCurrent() (SettingsModel, tea.Cmd) {
	oldRefresh := m.refreshSelected
	oldAgents := m.agentsSelected

	switch m.focusGroup {
	case GroupRefresh:
		m.refreshSelected = m.refreshCursor
		if m.onSave != nil {
			m.onSave(m.refreshOptions[m.refreshSelected].Duration)
		}
	case GroupAgents:
		m.agentsSelected = m.agentsCursor
	}

	return m, func() tea.Msg {
		return SettingsCloseMsg{
			NewInterval:     m.refreshOptions[m.refreshSelected].Duration,
			NewAgentsFilter: m.agentsSelected,
			RefreshChanged:  oldRefresh != m.refreshSelected,
			AgentsChanged:   oldAgents != m.agentsSelected,
		}
	}
}

// View renders the settings screen
func (m SettingsModel) View() string {
	var s string

	s += settingsTitleStyle.Render("Settings") + "\n\n"

	// Render two groups side by side
	leftCol := m.renderRefreshGroup()
	rightCol := m.renderAgentsGroup()

	// Join columns horizontally with spacing
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, "    ", rightCol)
	s += columns

	s += settingsHelpStyle.Render("\n\n  ↑/↓ navigate • ←/→/Tab switch group • Enter select • Esc back")

	return settingsContainerStyle.Render(s)
}

func (m SettingsModel) renderRefreshGroup() string {
	var s string
	active := m.focusGroup == GroupRefresh

	if active {
		s += settingsHeaderStyle.Render("Refresh Rate") + "\n"
	} else {
		s += settingsHeaderInactiveStyle.Render("Refresh Rate") + "\n"
	}

	for i, opt := range m.refreshOptions {
		var radioBtn string
		if i == m.refreshSelected {
			radioBtn = "●"
		} else {
			radioBtn = "○"
		}

		var line string
		if active && i == m.refreshCursor {
			radioBtn = settingsCursorStyle.Render(radioBtn)
			line = settingsSelectedStyle.Render(radioBtn + " " + opt.Label)
		} else if i == m.refreshSelected {
			if active {
				line = settingsSelectedStyle.Render(radioBtn + " " + opt.Label)
			} else {
				line = settingsOptionInactiveStyle.Render(radioBtn + " " + opt.Label)
			}
		} else {
			if active {
				line = settingsOptionStyle.Render(radioBtn + " " + opt.Label)
			} else {
				line = settingsOptionInactiveStyle.Render(radioBtn + " " + opt.Label)
			}
		}

		s += line + "\n"
	}

	return s
}

func (m SettingsModel) renderAgentsGroup() string {
	var s string
	active := m.focusGroup == GroupAgents

	if active {
		s += settingsHeaderStyle.Render("Agents Display") + "\n"
	} else {
		s += settingsHeaderInactiveStyle.Render("Agents Display") + "\n"
	}

	for i, opt := range m.agentsOptions {
		var radioBtn string
		if i == m.agentsSelected {
			radioBtn = "●"
		} else {
			radioBtn = "○"
		}

		var line string
		if active && i == m.agentsCursor {
			radioBtn = settingsCursorStyle.Render(radioBtn)
			line = settingsSelectedStyle.Render(radioBtn + " " + opt.Label)
		} else if i == m.agentsSelected {
			if active {
				line = settingsSelectedStyle.Render(radioBtn + " " + opt.Label)
			} else {
				line = settingsOptionInactiveStyle.Render(radioBtn + " " + opt.Label)
			}
		} else {
			if active {
				line = settingsOptionStyle.Render(radioBtn + " " + opt.Label)
			} else {
				line = settingsOptionInactiveStyle.Render(radioBtn + " " + opt.Label)
			}
		}

		s += line + "\n"
	}

	return s
}

// GetSelectedDuration returns the currently selected refresh duration
func (m SettingsModel) GetSelectedDuration() time.Duration {
	if m.refreshSelected >= 0 && m.refreshSelected < len(m.refreshOptions) {
		return m.refreshOptions[m.refreshSelected].Duration
	}
	return defaultRefreshInterval
}

// GetSelectedAgentsFilter returns the selected agents filter index
func (m SettingsModel) GetSelectedAgentsFilter() int {
	return m.agentsSelected
}

// SetSize updates the terminal size for the settings view
func (m *SettingsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
