package command

// Reduce applies an action to the command palette state
func Reduce(state State, action Action) State {
	switch a := action.(type) {
	case Activate:
		state.Active = true
		state.Input = ""
		state.SelectedIdx = 0
		return state

	case Deactivate:
		state.Active = false
		state.Input = ""
		state.SelectedIdx = 0
		return state

	case SetInput:
		state.Input = a.Input
		state.SelectedIdx = 0 // Reset selection on input change
		return state

	case SelectUp:
		filtered := state.FilteredCommands()
		if len(filtered) > 0 {
			state.SelectedIdx--
			if state.SelectedIdx < 0 {
				state.SelectedIdx = len(filtered) - 1
			}
		}
		return state

	case SelectDown:
		filtered := state.FilteredCommands()
		if len(filtered) > 0 {
			state.SelectedIdx++
			if state.SelectedIdx >= len(filtered) {
				state.SelectedIdx = 0
			}
		}
		return state

	case SetSelection:
		state.SelectedIdx = a.Index
		return state

	case Autocomplete:
		if cmd := state.SelectedCommand(); cmd != nil {
			state.Input = cmd.Name + " "
		}
		return state
	}

	return state
}
