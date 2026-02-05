package dialog

// Type represents dialog types
type Type int

const (
	TypeNone Type = iota
	TypeQuit
	TypeConfirm
	TypeError
)

// State holds the dialog component state
type State struct {
	Type      Type
	Title     string
	Message   string
	Selection int      // Selected button index
	Options   []string // Button labels
}

// Initial returns the initial dialog state (hidden)
func Initial() State {
	return State{
		Type: TypeNone,
	}
}

// IsOpen returns true if dialog is visible
func (s State) IsOpen() bool {
	return s.Type != TypeNone
}

// IsConfirmed returns true if the confirm option is selected
func (s State) IsConfirmed() bool {
	return s.Selection == len(s.Options)-1
}
