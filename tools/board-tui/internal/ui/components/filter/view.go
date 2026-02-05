package filter

// View renders the filter indicator (if needed)
// The actual filtering is handled by bubbles/list
func View(state State) string {
	if !state.IsActive() {
		return ""
	}

	// Filter UI is handled by the list component
	// This is just for custom rendering if needed
	return ""
}
