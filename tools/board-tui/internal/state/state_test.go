package state

import (
	"testing"
	"time"

	boardpkg "board-tui/internal/ui/screens/board"
)

func TestNewAppState(t *testing.T) {
	s := NewAppState()

	if s.Screen != ScreenBoard {
		t.Error("Initial screen should be ScreenBoard")
	}
	if s.Shared.RefreshInterval != 10*time.Second {
		t.Error("Default refresh interval should be 10 seconds")
	}
	if s.Quitting {
		t.Error("Initial state should not be quitting")
	}
}

func TestReduceSetScreen(t *testing.T) {
	s := NewAppState()

	s = Reduce(s, SetScreen{Screen: ScreenSettings})

	if s.Screen != ScreenSettings {
		t.Error("Screen should be ScreenSettings after SetScreen")
	}
}

func TestReduceSetTree(t *testing.T) {
	s := NewAppState()
	tree := []*boardpkg.TreeNode{
		{ID: "EPIC-01", Type: "epic"},
	}

	s = Reduce(s, SetTree{Tree: tree})

	if len(s.Shared.Tree) != 1 {
		t.Error("Tree should have 1 node")
	}
	if s.Shared.Tree[0].ID != "EPIC-01" {
		t.Error("Tree node ID should be EPIC-01")
	}
	if s.Shared.LastUpdate.IsZero() {
		t.Error("LastUpdate should be set")
	}
}

func TestReduceSetTreeWithError(t *testing.T) {
	s := NewAppState()
	err := &testError{"load failed"}

	s = Reduce(s, SetTree{Tree: nil, LoadError: err})

	if s.Shared.LoadError == nil {
		t.Error("LoadError should be set")
	}
}

func TestReduceToggleNode(t *testing.T) {
	node := &boardpkg.TreeNode{ID: "EPIC-01", Expanded: false}
	s := NewAppState()
	s.Shared.Tree = []*boardpkg.TreeNode{node}

	s = Reduce(s, ToggleNode{NodeID: "EPIC-01"})

	if !s.Shared.Tree[0].Expanded {
		t.Error("Node should be expanded after toggle")
	}

	s = Reduce(s, ToggleNode{NodeID: "EPIC-01"})

	if s.Shared.Tree[0].Expanded {
		t.Error("Node should be collapsed after second toggle")
	}
}

func TestReduceExpandAll(t *testing.T) {
	child := &boardpkg.TreeNode{ID: "STORY-01", Expanded: false}
	node := &boardpkg.TreeNode{ID: "EPIC-01", Expanded: false, Children: []*boardpkg.TreeNode{child}}
	s := NewAppState()
	s.Shared.Tree = []*boardpkg.TreeNode{node}

	s = Reduce(s, ExpandAll{})

	if !s.Shared.Tree[0].Expanded {
		t.Error("Root should be expanded")
	}
	if !s.Shared.Tree[0].Children[0].Expanded {
		t.Error("Child should be expanded")
	}
}

func TestReduceCollapseAll(t *testing.T) {
	child := &boardpkg.TreeNode{ID: "STORY-01", Expanded: true}
	node := &boardpkg.TreeNode{ID: "EPIC-01", Expanded: true, Children: []*boardpkg.TreeNode{child}}
	s := NewAppState()
	s.Shared.Tree = []*boardpkg.TreeNode{node}

	s = Reduce(s, CollapseAll{})

	if s.Shared.Tree[0].Expanded {
		t.Error("Root should be collapsed")
	}
	if s.Shared.Tree[0].Children[0].Expanded {
		t.Error("Child should be collapsed")
	}
}

func TestReduceSetWindowSize(t *testing.T) {
	s := NewAppState()

	s = Reduce(s, SetWindowSize{Width: 100, Height: 50})

	if s.Shared.Width != 100 {
		t.Errorf("Expected width 100, got %d", s.Shared.Width)
	}
	if s.Shared.Height != 50 {
		t.Errorf("Expected height 50, got %d", s.Shared.Height)
	}
}

func TestReduceShowDetail(t *testing.T) {
	s := NewAppState()

	s = Reduce(s, ShowDetail{ElementID: "TASK-01"})

	if s.Screen != ScreenDetail {
		t.Error("Screen should be ScreenDetail")
	}
	if s.Detail.ElementID != "TASK-01" {
		t.Error("ElementID should be TASK-01")
	}
	if !s.Detail.Loading {
		t.Error("Should be loading")
	}
}

func TestReduceShowHelp(t *testing.T) {
	s := NewAppState()

	s = Reduce(s, ShowHelp{})

	if s.Screen != ScreenDetail {
		t.Error("Screen should be ScreenDetail")
	}
	if !s.Detail.HelpMode {
		t.Error("HelpMode should be true")
	}
}

func TestReduceShowAgents(t *testing.T) {
	s := NewAppState()

	s = Reduce(s, ShowAgents{})

	if s.Screen != ScreenAgents {
		t.Error("Screen should be ScreenAgents")
	}
	if !s.Agents.Loading {
		t.Error("Agents should be loading")
	}
}

func TestReduceShowSettings(t *testing.T) {
	s := NewAppState()

	s = Reduce(s, ShowSettings{})

	if s.Screen != ScreenSettings {
		t.Error("Screen should be ScreenSettings")
	}
}

func TestReduceForceQuit(t *testing.T) {
	s := NewAppState()

	s = Reduce(s, ForceQuit{})

	if !s.Quitting {
		t.Error("Should be quitting after ForceQuit")
	}
}

func TestReduceConfirmQuit(t *testing.T) {
	s := NewAppState()

	s = Reduce(s, ConfirmQuit{})

	if !s.Quitting {
		t.Error("Should be quitting after ConfirmQuit")
	}
}

func TestIsDialogOpen(t *testing.T) {
	s := NewAppState()

	if s.IsDialogOpen() {
		t.Error("Dialog should not be open initially")
	}
}

func TestIsCommandActive(t *testing.T) {
	s := NewAppState()

	if s.IsCommandActive() {
		t.Error("Command should not be active initially")
	}
}

func TestIsFiltering(t *testing.T) {
	s := NewAppState()

	if s.IsFiltering() {
		t.Error("Filter should not be active initially")
	}
}

// testError is a simple error for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
