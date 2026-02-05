package settings

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"board-tui/internal/styles"
)

// Option represents a selectable option with duration
type Option struct {
	Label    string
	Duration time.Duration // 0 means "Off" or "All"
}

// AgentsFilter represents the agents display filter setting
type AgentsFilter int

const (
	AgentsFilterAll AgentsFilter = iota
	AgentsFilterStale5
	AgentsFilterStale10
	AgentsFilterStale15
	AgentsFilterStale30
	AgentsFilterStale60
)

// StaleMinutes returns the stale threshold in minutes (0 = all)
func (f AgentsFilter) StaleMinutes() int {
	switch f {
	case AgentsFilterStale5:
		return 5
	case AgentsFilterStale10:
		return 10
	case AgentsFilterStale15:
		return 15
	case AgentsFilterStale30:
		return 30
	case AgentsFilterStale60:
		return 60
	default:
		return 0
	}
}

// CloseMsg is sent when the settings screen should close
type CloseMsg struct {
	NewInterval       time.Duration
	NewAgentsFilter   AgentsFilter
	RefreshChanged    bool
	AgentsChanged     bool
}

// SettingsGroup represents which settings group is focused
type SettingsGroup int

const (
	GroupRefresh SettingsGroup = iota
	GroupAgents
)

// Model is the bubbletea model for the settings screen
type Model struct {
	// Refresh rate settings
	refreshOptions  []Option
	refreshCursor   int
	refreshSelected int

	// Agents filter settings
	agentsOptions  []Option
	agentsCursor   int
	agentsSelected int

	// UI state
	focusGroup SettingsGroup
	width      int
	height     int

	// Callback
	onSave func(time.Duration)
}

// Local styles
var (
	titleStyle = styles.Title.MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

	headerInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#808080")).
				Bold(true).
				MarginTop(1).
				MarginBottom(1)

	optionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C1C6B2")).
			PaddingLeft(2)

	optionInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262")).
				PaddingLeft(2)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C5CE7")).
			Bold(true).
			PaddingLeft(2)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Bold(true)

	containerStyle = lipgloss.NewStyle().
			Padding(1, 2)
)

// DefaultRefreshOptions returns the available refresh rate options
func DefaultRefreshOptions() []Option {
	return []Option{
		{Label: "5 seconds", Duration: 5 * time.Second},
		{Label: "10 seconds (default)", Duration: 10 * time.Second},
		{Label: "30 seconds", Duration: 30 * time.Second},
		{Label: "60 seconds", Duration: 60 * time.Second},
		{Label: "Off (manual only)", Duration: 0},
	}
}

// DefaultAgentsOptions returns the available agents filter options
func DefaultAgentsOptions() []Option {
	return []Option{
		{Label: "All agents", Duration: 0},
		{Label: "Stale > 5 min", Duration: 5 * time.Minute},
		{Label: "Stale > 10 min", Duration: 10 * time.Minute},
		{Label: "Stale > 15 min", Duration: 15 * time.Minute},
		{Label: "Stale > 30 min", Duration: 30 * time.Minute},
		{Label: "Stale > 60 min", Duration: 60 * time.Minute},
	}
}

// New creates a new settings model
func New(currentInterval time.Duration, onSave func(time.Duration)) Model {
	return NewWithAgentsFilter(currentInterval, AgentsFilterAll, onSave)
}

// NewWithAgentsFilter creates a new settings model with agents filter
func NewWithAgentsFilter(currentInterval time.Duration, agentsFilter AgentsFilter, onSave func(time.Duration)) Model {
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
	agentsSelected := int(agentsFilter)
	if agentsSelected >= len(agentsOptions) {
		agentsSelected = 0
	}

	return Model{
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

// SetSize updates the terminal size
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetSelectedDuration returns the currently selected refresh duration
func (m Model) GetSelectedDuration() time.Duration {
	if m.refreshSelected >= 0 && m.refreshSelected < len(m.refreshOptions) {
		return m.refreshOptions[m.refreshSelected].Duration
	}
	return 10 * time.Second
}

// GetSelectedAgentsFilter returns the currently selected agents filter
func (m Model) GetSelectedAgentsFilter() AgentsFilter {
	return AgentsFilter(m.agentsSelected)
}

// Update handles input
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
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
				return CloseMsg{
					NewInterval:     m.refreshOptions[m.refreshSelected].Duration,
					NewAgentsFilter: AgentsFilter(m.agentsSelected),
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

func (m *Model) moveCursorUp() {
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

func (m *Model) moveCursorDown() {
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

func (m *Model) nextGroup() {
	if m.focusGroup == GroupRefresh {
		m.focusGroup = GroupAgents
	}
}

func (m *Model) prevGroup() {
	if m.focusGroup == GroupAgents {
		m.focusGroup = GroupRefresh
	}
}

func (m Model) selectCurrent() (Model, tea.Cmd) {
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
		return CloseMsg{
			NewInterval:     m.refreshOptions[m.refreshSelected].Duration,
			NewAgentsFilter: AgentsFilter(m.agentsSelected),
			RefreshChanged:  oldRefresh != m.refreshSelected,
			AgentsChanged:   oldAgents != m.agentsSelected,
		}
	}
}

// View renders the settings screen
func (m Model) View() string {
	var s string

	s += titleStyle.Render("Settings") + "\n\n"

	// Refresh Rate section
	s += m.renderGroup("Refresh Rate", m.refreshOptions, m.refreshCursor, m.refreshSelected, m.focusGroup == GroupRefresh)
	s += "\n"

	// Agents Filter section
	s += m.renderGroup("Agents Display", m.agentsOptions, m.agentsCursor, m.agentsSelected, m.focusGroup == GroupAgents)

	s += styles.Help.Render("\n\n  ↑/↓ navigate • ←/→/Tab switch group • Enter select • Esc back")

	return containerStyle.Render(s)
}

func (m Model) renderGroup(title string, options []Option, cursor, selected int, active bool) string {
	var s string

	if active {
		s += headerStyle.Render(title) + "\n"
	} else {
		s += headerInactiveStyle.Render(title) + "\n"
	}

	for i, opt := range options {
		var radioBtn string
		if i == selected {
			radioBtn = "●"
		} else {
			radioBtn = "○"
		}

		var line string
		if active && i == cursor {
			radioBtn = cursorStyle.Render(radioBtn)
			line = selectedStyle.Render(radioBtn + " " + opt.Label)
		} else if i == selected {
			if active {
				line = selectedStyle.Render(radioBtn + " " + opt.Label)
			} else {
				line = optionInactiveStyle.Render(radioBtn + " " + opt.Label)
			}
		} else {
			if active {
				line = optionStyle.Render(radioBtn + " " + opt.Label)
			} else {
				line = optionInactiveStyle.Render(radioBtn + " " + opt.Label)
			}
		}

		s += line + "\n"
	}

	return s
}
