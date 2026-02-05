package board

import (
	"board-tui/internal/ui/components/command"
	"board-tui/internal/ui/components/dialog"
	"board-tui/internal/ui/components/filter"
)

// State holds the board screen state
type State struct {
	// Child component states
	Filter  filter.State
	Command command.State
	Dialog  dialog.State

	// Board-specific state
	SelectedID string // Currently selected node ID
	Refreshing bool
}

// Initial returns the initial board state
func Initial(commands []command.Command) State {
	return State{
		Filter:  filter.Initial(),
		Command: command.Initial(commands),
		Dialog:  dialog.Initial(),
	}
}

// IsBlocked returns true if any modal is open (dialog, command, filter)
func (s State) IsBlocked() bool {
	return s.Dialog.IsOpen() || s.Command.IsActive()
}

// CanNavigate returns true if tree navigation is allowed
func (s State) CanNavigate() bool {
	return !s.Dialog.IsOpen() && !s.Command.IsActive() && !s.Filter.IsActive()
}
