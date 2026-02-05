package filter

// State holds the filter component state
type State struct {
	Active            bool
	Query             string
	PreFilterExpanded []string // Expanded node IDs to restore after filtering
}

// Initial returns the initial filter state
func Initial() State {
	return State{
		Active: false,
		Query:  "",
	}
}

// IsActive returns true if filter is active
func (s State) IsActive() bool {
	return s.Active
}
