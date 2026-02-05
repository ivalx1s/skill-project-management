package dialog

// Reduce applies an action to the dialog state
func Reduce(state State, action Action) State {
	switch a := action.(type) {
	case Show:
		return State{
			Type:      a.Type,
			Title:     a.Title,
			Message:   a.Message,
			Options:   a.Options,
			Selection: 0,
		}

	case Hide:
		return Initial()

	case SelectNext:
		if len(state.Options) > 0 {
			state.Selection = (state.Selection + 1) % len(state.Options)
		}
		return state

	case SelectPrev:
		if len(state.Options) > 0 {
			state.Selection--
			if state.Selection < 0 {
				state.Selection = len(state.Options) - 1
			}
		}
		return state

	case SetSelection:
		if a.Index >= 0 && a.Index < len(state.Options) {
			state.Selection = a.Index
		}
		return state

	case Toggle:
		if len(state.Options) == 2 {
			state.Selection = 1 - state.Selection
		}
		return state
	}

	return state
}

// ShowQuit is a convenience function to create a quit dialog
func ShowQuit() Show {
	return Show{
		Type:    TypeQuit,
		Title:   "Quit?",
		Options: []string{"No", "Yes"},
	}
}
