package command

// Command represents a slash command definition
type Command struct {
	Name        string
	Description string
}

// State holds the command palette state
type State struct {
	Active      bool
	Input       string
	SelectedIdx int
	Commands    []Command // Available commands
}

// Initial returns the initial command palette state
func Initial(commands []Command) State {
	return State{
		Active:   false,
		Commands: commands,
	}
}

// IsActive returns true if command palette is active
func (s State) IsActive() bool {
	return s.Active
}

// FilteredCommands returns commands matching the current input
func (s State) FilteredCommands() []Command {
	if s.Input == "" {
		return s.Commands
	}

	var filtered []Command
	for _, cmd := range s.Commands {
		if len(cmd.Name) >= len(s.Input) && cmd.Name[:len(s.Input)] == s.Input {
			filtered = append(filtered, cmd)
		}
	}
	return filtered
}

// SelectedCommand returns the currently selected command or nil
func (s State) SelectedCommand() *Command {
	filtered := s.FilteredCommands()
	if s.SelectedIdx >= 0 && s.SelectedIdx < len(filtered) {
		return &filtered[s.SelectedIdx]
	}
	return nil
}
