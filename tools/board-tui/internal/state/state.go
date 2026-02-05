package state

import (
	"time"

	boardpkg "board-tui/internal/ui/screens/board"
)

// Screen represents the current view
type Screen int

const (
	ScreenBoard Screen = iota
	ScreenSettings
	ScreenDetail
	ScreenAgents
)

// =============================================================================
// Shared State (data shared across all screens)
// =============================================================================

// SharedState holds data shared across all screens
type SharedState struct {
	// Tree data
	Tree []*boardpkg.TreeNode

	// Timing
	LastUpdate      time.Time
	RefreshInterval time.Duration

	// Terminal dimensions
	Width  int
	Height int

	// Configuration
	Config       *Config
	ConfigLoaded bool

	// Error
	LoadError error
}

// Config holds persisted configuration
type Config struct {
	RefreshRate   int      // Refresh interval in seconds (0 = disabled)
	ExpandedNodes []string // IDs of expanded nodes
}

// =============================================================================
// Screen States (each screen has its own state)
// =============================================================================

// DetailState holds detail screen state
type DetailState struct {
	ElementID string
	Loading   bool
	Error     error
	HelpMode  bool
}

// SettingsState holds settings screen state
type SettingsState struct {
	Cursor   int
	Selected int
}

// AgentsState holds agents screen state
type AgentsState struct {
	Loading    bool
	Error      error
	LastUpdate time.Time
}

// =============================================================================
// App State (composition of all states)
// =============================================================================

// AppState is the single source of truth
type AppState struct {
	// Current screen
	Screen Screen

	// Shared data
	Shared SharedState

	// Screen-specific states (only active screen's state is used)
	Board    boardpkg.State
	Detail   DetailState
	Settings SettingsState
	Agents   AgentsState

	// App lifecycle
	Quitting bool
}

// NewAppState creates a new application state with defaults
func NewAppState() *AppState {
	return &AppState{
		Screen: ScreenBoard,
		Shared: SharedState{
			RefreshInterval: 10 * time.Second,
		},
		Board:    boardpkg.State{}, // Will be initialized with commands later
		Settings: SettingsState{Selected: 1},
	}
}

// =============================================================================
// Convenience accessors
// =============================================================================

// IsDialogOpen returns true if any dialog is open
func (s *AppState) IsDialogOpen() bool {
	if s.Screen == ScreenBoard {
		return s.Board.Dialog.IsOpen()
	}
	return false
}

// IsCommandActive returns true if command palette is active
func (s *AppState) IsCommandActive() bool {
	if s.Screen == ScreenBoard {
		return s.Board.Command.IsActive()
	}
	return false
}

// IsFiltering returns true if filter is active
func (s *AppState) IsFiltering() bool {
	if s.Screen == ScreenBoard {
		return s.Board.Filter.IsActive()
	}
	return false
}

// SelectedNode returns the currently selected tree node
func (s *AppState) SelectedNode() *boardpkg.TreeNode {
	return findNodeByID(s.Shared.Tree, s.Board.SelectedID)
}

func findNodeByID(nodes []*boardpkg.TreeNode, id string) *boardpkg.TreeNode {
	for _, node := range nodes {
		if node.ID == id {
			return node
		}
		if found := findNodeByID(node.Children, id); found != nil {
			return found
		}
	}
	return nil
}
