package board

import (
	"board-tui/internal/ui/components/command"
	"board-tui/internal/ui/components/dialog"
	"board-tui/internal/ui/components/filter"
)

// Reduce applies an action to the board state
func Reduce(state State, action Action) State {
	switch a := action.(type) {

	// Delegated to children
	case FilterAction:
		state.Filter = filter.Reduce(state.Filter, a.Action)
		return state

	case CommandAction:
		state.Command = command.Reduce(state.Command, a.Action)
		return state

	case DialogAction:
		state.Dialog = dialog.Reduce(state.Dialog, a.Action)
		return state

	// Board-specific
	case SetSelected:
		state.SelectedID = a.ID
		return state

	case SetRefreshing:
		state.Refreshing = a.Refreshing
		return state

	case RequestQuit:
		state.Dialog = dialog.Reduce(state.Dialog, dialog.ShowQuit())
		return state

	case ConfirmQuit:
		// Handled at app level
		return state

	case CancelQuit:
		state.Dialog = dialog.Reduce(state.Dialog, dialog.Hide{})
		return state
	}

	return state
}
