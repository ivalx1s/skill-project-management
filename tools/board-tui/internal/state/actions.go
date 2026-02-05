package state

import (
	"time"

	boardpkg "board-tui/internal/ui/screens/board"
)

// Action represents any action that can modify the application state
type Action interface {
	actionMarker()
}

type action struct{}

func (action) actionMarker() {}

// =============================================================================
// Screen Navigation
// =============================================================================

// SetScreen changes the current screen
type SetScreen struct {
	action
	Screen Screen
}

// =============================================================================
// Board Actions (delegated to board reducer)
// =============================================================================

// BoardAction wraps a board action for delegation
type BoardAction struct {
	action
	Action boardpkg.Action
}

// WrapBoard wraps a board action
func WrapBoard(a boardpkg.Action) BoardAction {
	return BoardAction{Action: a}
}

// =============================================================================
// Tree/Shared Data Actions
// =============================================================================

// SetTree replaces the tree data
type SetTree struct {
	action
	Tree      []*boardpkg.TreeNode
	LoadError error
}

// ToggleNode toggles expansion of a specific node
type ToggleNode struct {
	action
	NodeID string
}

// ExpandNode expands a specific node
type ExpandNode struct {
	action
	NodeID string
}

// CollapseNode collapses a specific node
type CollapseNode struct {
	action
	NodeID string
}

// ExpandAll expands all nodes
type ExpandAll struct {
	action
}

// CollapseAll collapses all nodes
type CollapseAll struct {
	action
}

// ApplyExpandedNodes applies a list of expanded node IDs
type ApplyExpandedNodes struct {
	action
	IDs []string
}

// =============================================================================
// Timing Actions
// =============================================================================

// SetLastUpdate sets the last update time
type SetLastUpdate struct {
	action
	Time time.Time
}

// SetRefreshInterval sets the refresh interval
type SetRefreshInterval struct {
	action
	Interval time.Duration
}

// =============================================================================
// Window Actions
// =============================================================================

// SetWindowSize sets the terminal dimensions
type SetWindowSize struct {
	action
	Width  int
	Height int
}

// =============================================================================
// Config Actions
// =============================================================================

// SetConfig sets the configuration
type SetConfig struct {
	action
	Config *Config
}

// SetConfigLoaded marks config as loaded
type SetConfigLoaded struct {
	action
}

// =============================================================================
// Detail Screen Actions
// =============================================================================

// ShowDetail shows the detail view
type ShowDetail struct {
	action
	ElementID string
}

// ShowHelp shows the help view
type ShowHelp struct {
	action
}

// SetDetailLoading sets detail loading state
type SetDetailLoading struct {
	action
	Loading bool
}

// SetDetailError sets detail error
type SetDetailError struct {
	action
	Error error
}

// =============================================================================
// Settings Screen Actions
// =============================================================================

// ShowSettings shows the settings screen
type ShowSettings struct {
	action
}

// SetSettingsCursor sets the settings cursor
type SetSettingsCursor struct {
	action
	Cursor int
}

// SetSettingsSelected sets the selected setting
type SetSettingsSelected struct {
	action
	Selected int
}

// =============================================================================
// Agents Screen Actions
// =============================================================================

// ShowAgents shows the agents dashboard
type ShowAgents struct {
	action
}

// SetAgentsLoading sets agents loading state
type SetAgentsLoading struct {
	action
	Loading bool
}

// SetAgentsError sets agents error
type SetAgentsError struct {
	action
	Error error
}

// SetAgentsLastUpdate sets agents last update time
type SetAgentsLastUpdate struct {
	action
	Time time.Time
}

// =============================================================================
// App Lifecycle Actions
// =============================================================================

// ForceQuit immediately quits
type ForceQuit struct {
	action
}

// ConfirmQuit confirms quit (from dialog)
type ConfirmQuit struct {
	action
}
