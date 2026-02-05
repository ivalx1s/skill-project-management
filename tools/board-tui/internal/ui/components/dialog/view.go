package dialog

import (
	"github.com/charmbracelet/lipgloss"

	"board-tui/internal/styles"
)

// View renders the dialog
func View(state State, width int) string {
	if !state.IsOpen() {
		return ""
	}

	dialogStyle := styles.DialogBorder.Width(width - 10)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF4500")).
		Render(state.Title)

	// Render buttons
	var buttons string
	for i, opt := range state.Options {
		var btn string
		if i == state.Selection {
			btn = styles.ButtonSelected.Render(" " + opt + " ")
		} else {
			btn = styles.ButtonNormal.Render(" " + opt + " ")
		}
		if i > 0 {
			buttons += "  "
		}
		buttons += btn
	}

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("←/→: select • enter: confirm • esc: cancel")

	content := title
	if state.Message != "" {
		content += "\n\n" + state.Message
	}
	content += "\n\n" + buttons + "\n\n" + hint

	return dialogStyle.Render(content)
}
