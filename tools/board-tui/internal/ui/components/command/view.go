package command

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"board-tui/internal/styles"
)

// View renders the command palette
func View(state State, input string, width int) string {
	if !state.IsActive() {
		return ""
	}

	style := styles.CommandBorder.Width(width - 4)

	// Build suggestions list
	filtered := state.FilteredCommands()
	var suggestions strings.Builder

	for i, cmd := range filtered {
		prefix := "  "
		var name string
		if i == state.SelectedIdx {
			prefix = "> "
			name = styles.CommandSelected.Render(cmd.Name)
		} else {
			name = styles.CommandNormal.Render(cmd.Name)
		}
		desc := styles.CommandDesc.Render(" - " + cmd.Description)
		suggestions.WriteString(prefix + name + desc + "\n")
	}

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("  ↑↓ navigate • tab complete • enter execute • esc cancel")

	// Input field rendering
	prompt := "> "
	inputDisplay := prompt + input + "█" // Cursor

	return style.Render(inputDisplay + "\n\n" + suggestions.String() + "\n" + hint)
}
