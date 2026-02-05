package dialog

import "testing"

func TestDialogInitial(t *testing.T) {
	s := Initial()
	if s.IsOpen() {
		t.Error("Initial dialog should not be open")
	}
	if s.Type != TypeNone {
		t.Error("Initial dialog type should be TypeNone")
	}
}

func TestDialogShow(t *testing.T) {
	s := Initial()

	s = Reduce(s, Show{
		Type:    TypeQuit,
		Title:   "Quit?",
		Options: []string{"No", "Yes"},
	})

	if !s.IsOpen() {
		t.Error("Dialog should be open after Show")
	}
	if s.Type != TypeQuit {
		t.Errorf("Expected TypeQuit, got %d", s.Type)
	}
	if s.Title != "Quit?" {
		t.Errorf("Expected title 'Quit?', got '%s'", s.Title)
	}
	if len(s.Options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(s.Options))
	}
	if s.Selection != 0 {
		t.Error("Selection should be 0 after Show")
	}
}

func TestDialogHide(t *testing.T) {
	s := State{
		Type:      TypeQuit,
		Title:     "Quit?",
		Selection: 1,
	}

	s = Reduce(s, Hide{})

	if s.IsOpen() {
		t.Error("Dialog should not be open after Hide")
	}
}

func TestDialogSelectNext(t *testing.T) {
	s := State{
		Type:      TypeQuit,
		Options:   []string{"No", "Yes"},
		Selection: 0,
	}

	s = Reduce(s, SelectNext{})
	if s.Selection != 1 {
		t.Errorf("Expected selection 1, got %d", s.Selection)
	}

	// Wrap around
	s = Reduce(s, SelectNext{})
	if s.Selection != 0 {
		t.Errorf("Expected selection 0 (wrap), got %d", s.Selection)
	}
}

func TestDialogSelectPrev(t *testing.T) {
	s := State{
		Type:      TypeQuit,
		Options:   []string{"No", "Yes"},
		Selection: 1,
	}

	s = Reduce(s, SelectPrev{})
	if s.Selection != 0 {
		t.Errorf("Expected selection 0, got %d", s.Selection)
	}

	// Wrap around
	s = Reduce(s, SelectPrev{})
	if s.Selection != 1 {
		t.Errorf("Expected selection 1 (wrap), got %d", s.Selection)
	}
}

func TestDialogToggle(t *testing.T) {
	s := State{
		Type:      TypeQuit,
		Options:   []string{"No", "Yes"},
		Selection: 0,
	}

	s = Reduce(s, Toggle{})
	if s.Selection != 1 {
		t.Errorf("Expected selection 1 after toggle, got %d", s.Selection)
	}

	s = Reduce(s, Toggle{})
	if s.Selection != 0 {
		t.Errorf("Expected selection 0 after toggle, got %d", s.Selection)
	}
}

func TestDialogIsConfirmed(t *testing.T) {
	s := State{
		Type:      TypeQuit,
		Options:   []string{"No", "Yes"},
		Selection: 0,
	}

	if s.IsConfirmed() {
		t.Error("Selection 0 should not be confirmed")
	}

	s.Selection = 1
	if !s.IsConfirmed() {
		t.Error("Selection 1 (last option) should be confirmed")
	}
}

func TestShowQuit(t *testing.T) {
	action := ShowQuit()

	if action.Type != TypeQuit {
		t.Error("ShowQuit should create TypeQuit action")
	}
	if len(action.Options) != 2 {
		t.Error("ShowQuit should have 2 options")
	}
}
