package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Command represents a slash command
type Command struct {
	Name        string
	Description string
	Handler     func(args string) tea.Cmd
}

// CommandModel handles the command input mode
type CommandModel struct {
	input       textinput.Model
	active      bool
	commands    []Command
	width       int
	selectedIdx int // Selected suggestion index
}

// Messages
type CommandExecuteMsg struct {
	Command string
	Args    string
}

type CommandCancelMsg struct{}

// NewCommandModel creates a new command input model
func NewCommandModel() CommandModel {
	ti := textinput.New()
	ti.Placeholder = "Type command name..."
	ti.Prompt = "> "
	ti.CharLimit = 100

	return CommandModel{
		input:  ti,
		active: false,
		commands: []Command{
			{Name: "filter", Description: "Filter items (e.g., /filter done)"},
			{Name: "agents", Description: "Show agent assignments"},
			{Name: "arkanoid", Description: "Open Arkanoid mini-game"},
			{Name: "settings", Description: "Open settings screen"},
			{Name: "refresh", Description: "Force refresh data"},
			{Name: "expand", Description: "Expand all nodes"},
			{Name: "collapse", Description: "Collapse all nodes"},
			{Name: "help", Description: "Show available commands"},
		},
	}
}

// Activate shows the command input and returns updated model
func (m CommandModel) Activate() CommandModel {
	m.active = true
	m.input.SetValue("")
	m.input.Focus()
	m.selectedIdx = 0
	return m
}

// Deactivate hides the command input and returns updated model
func (m CommandModel) Deactivate() CommandModel {
	m.active = false
	m.input.Blur()
	return m
}

// IsActive returns true if command mode is active
func (m *CommandModel) IsActive() bool {
	return m.active
}

// SetWidth sets the input width
func (m *CommandModel) SetWidth(w int) {
	m.width = w
	m.input.Width = w - 10
}

// getFilteredCommands returns commands matching current input
func (m *CommandModel) getFilteredCommands() []Command {
	value := strings.ToLower(strings.TrimSpace(m.input.Value()))
	// Get just the command part (before any space)
	parts := strings.SplitN(value, " ", 2)
	query := parts[0]

	if query == "" {
		return m.commands
	}

	var filtered []Command
	for _, cmd := range m.commands {
		if strings.HasPrefix(strings.ToLower(cmd.Name), query) {
			filtered = append(filtered, cmd)
		}
	}
	return filtered
}

// Update handles input
func (m CommandModel) Update(msg tea.Msg) (CommandModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "tab":
			filtered := m.getFilteredCommands()
			value := strings.TrimSpace(m.input.Value())

			// If tab and we have suggestions, autocomplete
			if msg.String() == "tab" && len(filtered) > 0 {
				// Complete to selected command
				if m.selectedIdx < len(filtered) {
					m.input.SetValue(filtered[m.selectedIdx].Name + " ")
					m.input.CursorEnd()
				}
				return m, nil
			}

			// Enter - execute command
			m = m.Deactivate()

			// If no input but has selected suggestion, use that
			if value == "" && len(filtered) > 0 && m.selectedIdx < len(filtered) {
				value = filtered[m.selectedIdx].Name
			}

			if value == "" {
				return m, func() tea.Msg { return CommandCancelMsg{} }
			}

			// If partial match with suggestion, use selected command
			if len(filtered) > 0 && m.selectedIdx < len(filtered) {
				parts := strings.SplitN(value, " ", 2)
				cmdPart := parts[0]
				// Check if input is prefix of selected command
				if strings.HasPrefix(filtered[m.selectedIdx].Name, cmdPart) && !strings.Contains(value, " ") {
					value = filtered[m.selectedIdx].Name
				}
			}

			parts := strings.SplitN(value, " ", 2)
			cmd := strings.ToLower(parts[0])
			args := ""
			if len(parts) > 1 {
				args = parts[1]
			}

			return m, func() tea.Msg {
				return CommandExecuteMsg{Command: cmd, Args: args}
			}

		case "esc":
			m = m.Deactivate()
			return m, func() tea.Msg { return CommandCancelMsg{} }

		case "up":
			filtered := m.getFilteredCommands()
			if len(filtered) > 0 {
				m.selectedIdx--
				if m.selectedIdx < 0 {
					m.selectedIdx = len(filtered) - 1
				}
			}
			return m, nil

		case "down":
			filtered := m.getFilteredCommands()
			if len(filtered) > 0 {
				m.selectedIdx++
				if m.selectedIdx >= len(filtered) {
					m.selectedIdx = 0
				}
			}
			return m, nil
		}
	}

	// Reset selection when input changes
	oldValue := m.input.Value()
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != oldValue {
		m.selectedIdx = 0
	}

	return m, cmd
}

// View renders the command input with suggestions
func (m CommandModel) View() string {
	if !m.active {
		return ""
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6C5CE7")).
		Padding(0, 1).
		Width(m.width - 4)

	// Build suggestions list
	filtered := m.getFilteredCommands()
	var suggestions strings.Builder

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#6C5CE7")).
		Foreground(lipgloss.Color("#FFFFFF"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	for i, cmd := range filtered {
		prefix := "  "
		var name string
		if i == m.selectedIdx {
			prefix = "> "
			name = selectedStyle.Render(cmd.Name)
		} else {
			name = normalStyle.Render(cmd.Name)
		}
		desc := descStyle.Render(" - " + cmd.Description)
		suggestions.WriteString(prefix + name + desc + "\n")
	}

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("  ↑↓ navigate • tab complete • enter execute • esc cancel")

	return style.Render(m.input.View() + "\n\n" + suggestions.String() + "\n" + hint)
}

// GetCommands returns available commands for help
func (m *CommandModel) GetCommands() []Command {
	return m.commands
}
