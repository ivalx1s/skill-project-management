package command

import "testing"

func TestCommandInitial(t *testing.T) {
	commands := []Command{
		{Name: "help", Description: "Show help"},
		{Name: "quit", Description: "Quit"},
	}
	s := Initial(commands)

	if s.IsActive() {
		t.Error("Initial command palette should not be active")
	}
	if len(s.Commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(s.Commands))
	}
}

func TestCommandActivate(t *testing.T) {
	s := State{
		Active: false,
		Input:  "old",
	}

	s = Reduce(s, Activate{})

	if !s.Active {
		t.Error("Should be active after Activate")
	}
	if s.Input != "" {
		t.Error("Input should be cleared after Activate")
	}
	if s.SelectedIdx != 0 {
		t.Error("SelectedIdx should be 0 after Activate")
	}
}

func TestCommandDeactivate(t *testing.T) {
	s := State{
		Active:      true,
		Input:       "test",
		SelectedIdx: 2,
	}

	s = Reduce(s, Deactivate{})

	if s.Active {
		t.Error("Should not be active after Deactivate")
	}
	if s.Input != "" {
		t.Error("Input should be cleared after Deactivate")
	}
}

func TestCommandSetInput(t *testing.T) {
	s := State{
		Active:      true,
		Input:       "old",
		SelectedIdx: 2,
	}

	s = Reduce(s, SetInput{Input: "new"})

	if s.Input != "new" {
		t.Errorf("Expected input 'new', got '%s'", s.Input)
	}
	if s.SelectedIdx != 0 {
		t.Error("SelectedIdx should reset to 0 on input change")
	}
}

func TestCommandSelectUpDown(t *testing.T) {
	s := State{
		Active: true,
		Commands: []Command{
			{Name: "a"},
			{Name: "b"},
			{Name: "c"},
		},
		SelectedIdx: 0,
	}

	s = Reduce(s, SelectDown{})
	if s.SelectedIdx != 1 {
		t.Errorf("Expected idx 1, got %d", s.SelectedIdx)
	}

	s = Reduce(s, SelectDown{})
	if s.SelectedIdx != 2 {
		t.Errorf("Expected idx 2, got %d", s.SelectedIdx)
	}

	// Wrap around
	s = Reduce(s, SelectDown{})
	if s.SelectedIdx != 0 {
		t.Errorf("Expected idx 0 (wrap), got %d", s.SelectedIdx)
	}

	// Select up with wrap
	s = Reduce(s, SelectUp{})
	if s.SelectedIdx != 2 {
		t.Errorf("Expected idx 2 (wrap up), got %d", s.SelectedIdx)
	}
}

func TestCommandFilteredCommands(t *testing.T) {
	s := State{
		Commands: []Command{
			{Name: "help"},
			{Name: "hello"},
			{Name: "quit"},
		},
		Input: "",
	}

	// Empty input returns all
	filtered := s.FilteredCommands()
	if len(filtered) != 3 {
		t.Errorf("Empty input should return all commands, got %d", len(filtered))
	}

	// Filter by prefix
	s.Input = "hel"
	filtered = s.FilteredCommands()
	if len(filtered) != 2 {
		t.Errorf("'hel' should match 2 commands, got %d", len(filtered))
	}

	// No match
	s.Input = "xyz"
	filtered = s.FilteredCommands()
	if len(filtered) != 0 {
		t.Errorf("'xyz' should match 0 commands, got %d", len(filtered))
	}
}

func TestCommandSelectedCommand(t *testing.T) {
	s := State{
		Commands: []Command{
			{Name: "help"},
			{Name: "quit"},
		},
		SelectedIdx: 0,
	}

	cmd := s.SelectedCommand()
	if cmd == nil {
		t.Fatal("SelectedCommand should not be nil")
	}
	if cmd.Name != "help" {
		t.Errorf("Expected 'help', got '%s'", cmd.Name)
	}

	s.SelectedIdx = 1
	cmd = s.SelectedCommand()
	if cmd.Name != "quit" {
		t.Errorf("Expected 'quit', got '%s'", cmd.Name)
	}
}

func TestCommandAutocomplete(t *testing.T) {
	s := State{
		Commands: []Command{
			{Name: "help"},
			{Name: "quit"},
		},
		Input:       "h",
		SelectedIdx: 0,
	}

	s = Reduce(s, Autocomplete{})

	if s.Input != "help " {
		t.Errorf("Expected 'help ', got '%s'", s.Input)
	}
}
