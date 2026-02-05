package command

// Action represents command palette actions
type Action interface {
	commandAction()
}

type action struct{}

func (action) commandAction() {}

// Activate opens the command palette
type Activate struct {
	action
}

// Deactivate closes the command palette
type Deactivate struct {
	action
}

// SetInput updates the input text
type SetInput struct {
	action
	Input string
}

// SelectUp moves selection up
type SelectUp struct {
	action
}

// SelectDown moves selection down
type SelectDown struct {
	action
}

// SetSelection sets selection to specific index
type SetSelection struct {
	action
	Index int
}

// Autocomplete fills input with selected command
type Autocomplete struct {
	action
}
