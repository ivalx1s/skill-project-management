package board

import (
	"testing"

	"board-tui/internal/ui/components/command"
	"board-tui/internal/ui/components/dialog"
	"board-tui/internal/ui/components/filter"
)

func TestBoardInitial(t *testing.T) {
	commands := []command.Command{
		{Name: "help", Description: "Show help"},
	}
	s := Initial(commands)

	if s.Filter.IsActive() {
		t.Error("Filter should not be active initially")
	}
	if s.Command.IsActive() {
		t.Error("Command should not be active initially")
	}
	if s.Dialog.IsOpen() {
		t.Error("Dialog should not be open initially")
	}
}

func TestBoardReduceFilterAction(t *testing.T) {
	s := Initial(nil)

	s = Reduce(s, FilterAction{Action: filter.Activate{}})

	if !s.Filter.IsActive() {
		t.Error("Filter should be active after FilterAction(Activate)")
	}
}

func TestBoardReduceCommandAction(t *testing.T) {
	s := Initial([]command.Command{{Name: "help"}})

	s = Reduce(s, CommandAction{Action: command.Activate{}})

	if !s.Command.IsActive() {
		t.Error("Command should be active after CommandAction(Activate)")
	}
}

func TestBoardReduceDialogAction(t *testing.T) {
	s := Initial(nil)

	s = Reduce(s, DialogAction{Action: dialog.ShowQuit()})

	if !s.Dialog.IsOpen() {
		t.Error("Dialog should be open after DialogAction(ShowQuit)")
	}
}

func TestBoardReduceSetSelected(t *testing.T) {
	s := Initial(nil)

	s = Reduce(s, SetSelected{ID: "TASK-01"})

	if s.SelectedID != "TASK-01" {
		t.Errorf("Expected SelectedID 'TASK-01', got '%s'", s.SelectedID)
	}
}

func TestBoardReduceSetRefreshing(t *testing.T) {
	s := Initial(nil)

	s = Reduce(s, SetRefreshing{Refreshing: true})

	if !s.Refreshing {
		t.Error("Should be refreshing")
	}

	s = Reduce(s, SetRefreshing{Refreshing: false})

	if s.Refreshing {
		t.Error("Should not be refreshing")
	}
}

func TestBoardReduceRequestQuit(t *testing.T) {
	s := Initial(nil)

	s = Reduce(s, RequestQuit{})

	if !s.Dialog.IsOpen() {
		t.Error("Dialog should be open after RequestQuit")
	}
	if s.Dialog.Type != dialog.TypeQuit {
		t.Error("Dialog type should be TypeQuit")
	}
}

func TestBoardReduceCancelQuit(t *testing.T) {
	s := Initial(nil)
	s = Reduce(s, RequestQuit{})

	s = Reduce(s, CancelQuit{})

	if s.Dialog.IsOpen() {
		t.Error("Dialog should be closed after CancelQuit")
	}
}

func TestBoardIsBlocked(t *testing.T) {
	s := Initial(nil)

	if s.IsBlocked() {
		t.Error("Should not be blocked initially")
	}

	s = Reduce(s, RequestQuit{})

	if !s.IsBlocked() {
		t.Error("Should be blocked when dialog is open")
	}
}

func TestBoardCanNavigate(t *testing.T) {
	s := Initial(nil)

	if !s.CanNavigate() {
		t.Error("Should be able to navigate initially")
	}

	s = Reduce(s, RequestQuit{})

	if s.CanNavigate() {
		t.Error("Should not be able to navigate when dialog is open")
	}
}

func TestWrapFilter(t *testing.T) {
	action := WrapFilter(filter.Activate{})

	if _, ok := action.Action.(filter.Activate); !ok {
		t.Error("WrapFilter should wrap filter.Activate")
	}
}

func TestWrapCommand(t *testing.T) {
	action := WrapCommand(command.Activate{})

	if _, ok := action.Action.(command.Activate); !ok {
		t.Error("WrapCommand should wrap command.Activate")
	}
}

func TestWrapDialog(t *testing.T) {
	action := WrapDialog(dialog.Hide{})

	if _, ok := action.Action.(dialog.Hide); !ok {
		t.Error("WrapDialog should wrap dialog.Hide")
	}
}
