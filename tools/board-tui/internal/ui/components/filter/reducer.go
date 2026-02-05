package filter

// Reduce applies an action to the filter state
func Reduce(state State, action Action) State {
	switch a := action.(type) {
	case Activate:
		return State{
			Active:            true,
			Query:             "",
			PreFilterExpanded: a.PreFilterExpanded,
		}

	case Deactivate:
		return State{
			Active:            false,
			Query:             "",
			PreFilterExpanded: nil,
		}

	case SetQuery:
		state.Query = a.Query
		return state

	case Clear:
		state.Query = ""
		return state
	}

	return state
}
