package board

import (
	"board-tui/internal/ui/components/command"
	"board-tui/internal/ui/components/dialog"
	"board-tui/internal/ui/components/filter"
)

// Action represents board screen actions
type Action interface {
	boardAction()
}

type action struct{}

func (action) boardAction() {}

// =============================================================================
// Delegated actions (forwarded to child reducers)
// =============================================================================

// FilterAction wraps a filter action
type FilterAction struct {
	action
	Action filter.Action
}

// CommandAction wraps a command action
type CommandAction struct {
	action
	Action command.Action
}

// DialogAction wraps a dialog action
type DialogAction struct {
	action
	Action dialog.Action
}

// =============================================================================
// Board-specific actions
// =============================================================================

// SetSelected sets the selected node ID
type SetSelected struct {
	action
	ID string
}

// SetRefreshing sets the refreshing state
type SetRefreshing struct {
	action
	Refreshing bool
}

// RequestQuit shows the quit confirmation dialog
type RequestQuit struct {
	action
}

// ConfirmQuit is sent when quit is confirmed
type ConfirmQuit struct {
	action
}

// CancelQuit hides the quit dialog
type CancelQuit struct {
	action
}

// =============================================================================
// Convenience constructors
// =============================================================================

// WrapFilter wraps a filter action
func WrapFilter(a filter.Action) FilterAction {
	return FilterAction{Action: a}
}

// WrapCommand wraps a command action
func WrapCommand(a command.Action) CommandAction {
	return CommandAction{Action: a}
}

// WrapDialog wraps a dialog action
func WrapDialog(a dialog.Action) DialogAction {
	return DialogAction{Action: a}
}
