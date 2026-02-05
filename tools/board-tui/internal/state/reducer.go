package state

import (
	"time"

	boardpkg "board-tui/internal/ui/screens/board"
)

// Reduce applies an action to the state and returns the new state
// This is a pure function - same input always gives same output
func Reduce(state *AppState, action Action) *AppState {
	if state == nil {
		state = NewAppState()
	}

	// Create a copy for immutable update
	newState := *state

	switch a := action.(type) {

	// =========================================================================
	// Screen Navigation
	// =========================================================================

	case SetScreen:
		newState.Screen = a.Screen

	// =========================================================================
	// Board Actions (delegated)
	// =========================================================================

	case BoardAction:
		newState.Board = boardpkg.Reduce(newState.Board, a.Action)

	// =========================================================================
	// Tree/Shared Data
	// =========================================================================

	case SetTree:
		newState.Shared.Tree = a.Tree
		newState.Shared.LoadError = a.LoadError
		newState.Board.Refreshing = false
		if a.LoadError == nil {
			newState.Shared.LastUpdate = time.Now()
		}

	case ToggleNode:
		if node := findNodeByID(newState.Shared.Tree, a.NodeID); node != nil {
			node.Toggle()
		}

	case ExpandNode:
		if node := findNodeByID(newState.Shared.Tree, a.NodeID); node != nil {
			node.Expand()
		}

	case CollapseNode:
		if node := findNodeByID(newState.Shared.Tree, a.NodeID); node != nil {
			node.Collapse()
		}

	case ExpandAll:
		for _, root := range newState.Shared.Tree {
			root.ExpandAll()
		}

	case CollapseAll:
		for _, root := range newState.Shared.Tree {
			root.CollapseAll()
		}

	case ApplyExpandedNodes:
		boardpkg.ApplyExpandedNodes(newState.Shared.Tree, a.IDs)

	// =========================================================================
	// Timing
	// =========================================================================

	case SetLastUpdate:
		newState.Shared.LastUpdate = a.Time

	case SetRefreshInterval:
		newState.Shared.RefreshInterval = a.Interval
		if newState.Shared.Config != nil {
			newState.Shared.Config.RefreshRate = int(a.Interval.Seconds())
		}

	// =========================================================================
	// Window
	// =========================================================================

	case SetWindowSize:
		newState.Shared.Width = a.Width
		newState.Shared.Height = a.Height

	// =========================================================================
	// Config
	// =========================================================================

	case SetConfig:
		newState.Shared.Config = a.Config

	case SetConfigLoaded:
		newState.Shared.ConfigLoaded = true

	// =========================================================================
	// Detail Screen
	// =========================================================================

	case ShowDetail:
		newState.Screen = ScreenDetail
		newState.Detail = DetailState{
			ElementID: a.ElementID,
			Loading:   true,
		}

	case ShowHelp:
		newState.Screen = ScreenDetail
		newState.Detail = DetailState{
			HelpMode: true,
		}

	case SetDetailLoading:
		newState.Detail.Loading = a.Loading

	case SetDetailError:
		newState.Detail.Error = a.Error
		newState.Detail.Loading = false

	// =========================================================================
	// Settings Screen
	// =========================================================================

	case ShowSettings:
		newState.Screen = ScreenSettings

	case SetSettingsCursor:
		newState.Settings.Cursor = a.Cursor

	case SetSettingsSelected:
		newState.Settings.Selected = a.Selected

	// =========================================================================
	// Agents Screen
	// =========================================================================

	case ShowAgents:
		newState.Screen = ScreenAgents
		newState.Agents = AgentsState{Loading: true}

	case SetAgentsLoading:
		newState.Agents.Loading = a.Loading

	case SetAgentsError:
		newState.Agents.Error = a.Error
		newState.Agents.Loading = false

	case SetAgentsLastUpdate:
		newState.Agents.LastUpdate = a.Time

	// =========================================================================
	// App Lifecycle
	// =========================================================================

	case ForceQuit:
		newState.Quitting = true

	case ConfirmQuit:
		newState.Quitting = true
	}

	return &newState
}
