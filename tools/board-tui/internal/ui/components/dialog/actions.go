package dialog

// Action represents dialog actions
type Action interface {
	dialogAction()
}

type action struct{}

func (action) dialogAction() {}

// Show displays a dialog
type Show struct {
	action
	Type    Type
	Title   string
	Message string
	Options []string
}

// Hide hides the dialog
type Hide struct {
	action
}

// SelectNext moves selection to next option
type SelectNext struct {
	action
}

// SelectPrev moves selection to previous option
type SelectPrev struct {
	action
}

// SetSelection sets selection to specific index
type SetSelection struct {
	action
	Index int
}

// Toggle toggles between two options (for binary dialogs)
type Toggle struct {
	action
}
