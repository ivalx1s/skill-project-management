package filter

// Action represents filter actions
type Action interface {
	filterAction()
}

type action struct{}

func (action) filterAction() {}

// Activate activates the filter with saved expanded state
type Activate struct {
	action
	PreFilterExpanded []string
}

// Deactivate deactivates the filter
type Deactivate struct {
	action
}

// SetQuery updates the filter query
type SetQuery struct {
	action
	Query string
}

// Clear clears the filter query but keeps it active
type Clear struct {
	action
}
