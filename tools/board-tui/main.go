package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Default refresh interval for auto-refresh
const defaultRefreshInterval = 10 * time.Second

// Screen represents the current view
type Screen int

const (
	BoardScreen Screen = iota
	SettingsScreen
	DetailScreen
	AgentsScreen
)

// Styles
var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#6C5CE7")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	// Status colors for visual feedback
	statusStyles = map[string]lipgloss.Style{
		"backlog":     lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")),
		"ready":       lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF")),
		"development": lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")),
		"progress":    lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")),
		"review":      lipgloss.NewStyle().Foreground(lipgloss.Color("#9370DB")),
		"done":        lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32")),
		"blocked":     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4500")),
	}

	// Type indicators with visual symbols
	typeIndicators = map[string]string{
		"epic":  "◆", // Diamond - major initiative
		"story": "◇", // Empty diamond - feature
		"task":  "○", // Circle - work item
		"bug":   "●", // Filled circle - defect
	}

	// Type colors for visual distinction
	typeStyles = map[string]lipgloss.Style{
		"epic":  lipgloss.NewStyle().Foreground(lipgloss.Color("#E040FB")).Bold(true), // Purple/Magenta
		"story": lipgloss.NewStyle().Foreground(lipgloss.Color("#29B6F6")),            // Light Blue
		"task":  lipgloss.NewStyle().Foreground(lipgloss.Color("#66BB6A")),            // Green
		"bug":   lipgloss.NewStyle().Foreground(lipgloss.Color("#EF5350")),            // Red
	}
)

// Model is the bubbletea model
type model struct {
	tree             []*TreeNode // The tree data
	boardRows        []boardRow
	boardSelectedIdx int
	boardScrollOff   int
	quitting        bool
	err             error
	refreshInterval time.Duration // Auto-refresh interval
	lastUpdate      time.Time     // Time of last successful data refresh
	refreshing      bool          // True while refreshing data
	loadError       error         // Last load error (nil if online)
	width           int           // Terminal width for status bar
	height          int           // Terminal height
	currentScreen        Screen        // Current screen being displayed
	previousScreen       Screen        // Screen to return to from detail
	config               *Config       // Persisted configuration
	configLoaded         bool          // True after initial config has been applied to tree
	settingsModel        SettingsModel // Settings screen model
	detailModel          DetailModel   // Detail view model
	agentsModel          AgentsModel   // Agents dashboard model
	commandModel         CommandModel  // Command palette model
	logger               *Logger       // Session logger
	confirmQuit          bool          // Show quit confirmation dialog
	confirmSelection     int           // 0 = No (default), 1 = Yes
	agentsFilter         int           // Agents filter: 0=all, 1-5=stale minutes
}

func (m model) Init() tea.Cmd {
	return tea.Batch(loadTree, m.tickCmd(), m.updateTimeTickCmd())
}

// updateTimeTickCmd returns a tick for updating the status bar time display every second
func (m model) updateTimeTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return updateTimeMsg{}
	})
}

// updateTimeMsg is sent every second to refresh the "Updated Xs ago" display
type updateTimeMsg struct{}

// tickCmd returns a tea.Tick command for periodic refresh
func (m model) tickCmd() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg{time: t}
	})
}

// tickMsg is sent on each refresh tick
type tickMsg struct {
	time time.Time
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle quit confirmation dialog FIRST - block all other input
		if m.confirmQuit {
			key := msg.String()
			if m.logger != nil {
				m.logger.Debug("confirmQuit key: %q, selection: %d", key, m.confirmSelection)
			}
			switch key {
			case "left", "h", "right", "l", "tab":
				// Toggle selection
				m.confirmSelection = 1 - m.confirmSelection
				return m, nil
			case "y", "Y":
				m.saveConfigOnQuit()
				m.quitting = true
				return m, tea.Quit
			case "n", "N", "esc":
				m.confirmQuit = false
				m.confirmSelection = 0
				return m, nil
			case "enter":
				if m.confirmSelection == 1 {
					m.saveConfigOnQuit()
					m.quitting = true
					return m, tea.Quit
				}
				m.confirmQuit = false
				m.confirmSelection = 0
				return m, nil
			}
			// Ignore all other keys while confirming
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			m.saveConfigOnQuit()
			m.quitting = true
			return m, tea.Quit

		case "q":
			if m.currentScreen == BoardScreen {
				m.confirmQuit = true
				return m, nil
			}

		case "esc":
			// Esc returns to Board from Settings/Detail, or shows quit confirm from Board (but not when filtering)
			if m.currentScreen == AgentsScreen {
				m.currentScreen = BoardScreen
				return m, nil
			}
			if m.currentScreen == SettingsScreen {
				m.currentScreen = BoardScreen
				return m, nil
			}
			if m.currentScreen == DetailScreen {
				m.currentScreen = m.previousScreen
				return m, nil
			}
			if m.currentScreen == BoardScreen {
				m.confirmQuit = true
				return m, nil
			}

		}

		// Command palette input - must be checked BEFORE board handlers
		if m.commandModel.IsActive() {
			if m.logger != nil {
				m.logger.Key(msg.String())
				m.logger.State("command_palette active, input: %q", m.commandModel.input.Value())
			}
			var cmd tea.Cmd
			m.commandModel, cmd = m.commandModel.Update(msg)
			return m, cmd
		}

		// Board-specific key handlers
		if m.currentScreen == BoardScreen {
			switch msg.String() {
			case "enter", "o":
				if node := m.boardSelectedNode(); node != nil {
					id := node.ID
					return m, func() tea.Msg { return OpenDetailMsg{ID: id} }
				}
				return m, nil

			case " ":
				if node := m.boardSelectedNode(); node != nil && node.HasChildren() {
					node.Toggle()
					m.boardRebuildRows()
				}
				return m, nil

			case "e":
				for _, root := range m.tree {
					root.ExpandAll()
				}
				m.boardRebuildRows()
				return m, nil

			case "c":
				for _, root := range m.tree {
					root.CollapseAll()
				}
				m.boardRebuildRows()
				return m, nil

			case "right", "l":
				if node := m.boardSelectedNode(); node != nil && node.HasChildren() && !node.Expanded {
					node.Expand()
					m.boardRebuildRows()
				}
				return m, nil

			case "left", "h":
				if node := m.boardSelectedNode(); node != nil {
					if node.Expanded && node.HasChildren() {
						node.Collapse()
						m.boardRebuildRows()
					} else if node.Parent != nil {
						m.boardSelectNodeByID(node.Parent.ID)
					}
				}
				return m, nil

			case "r":
				if !m.refreshing {
					m.refreshing = true
					return m, loadTree
				}
				return m, nil

			case "g":
				m.boardGoTop()
				return m, nil

			case "G":
				m.boardGoBottom()
				return m, nil

			case "down":
				m.boardMoveDown()
				return m, nil

			case "up":
				m.boardMoveUp()
				return m, nil

			case "/", ".":
				if m.logger != nil {
					m.logger.Action("command_palette", "opened")
				}
				m.commandModel = NewCommandModel()
				m.commandModel.SetWidth(m.width)
				m.commandModel = m.commandModel.Activate()
				return m, nil
			}
		}

		// Settings-specific key handlers
		if m.currentScreen == SettingsScreen {
			var cmd tea.Cmd
			m.settingsModel, cmd = m.settingsModel.Update(msg)
			return m, cmd
		}

		// Detail-specific key handlers
		if m.currentScreen == DetailScreen {
			var cmd tea.Cmd
			m.detailModel, cmd = m.detailModel.Update(msg)
			return m, cmd
		}

		// Agents-specific key handlers
		if m.currentScreen == AgentsScreen {
			var cmd tea.Cmd
			m.agentsModel, cmd = m.agentsModel.Update(msg)
			return m, cmd
		}

	case tea.MouseMsg:
		// Handle mouse events
		if m.currentScreen == BoardScreen {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				m.boardMoveDown() // Natural scrolling (trackpad)
				return m, nil
			case tea.MouseButtonWheelDown:
				m.boardMoveUp() // Natural scrolling (trackpad)
				return m, nil
			case tea.MouseButtonLeft:
				// Tap opens detail for current cursor position
				if node := m.boardSelectedNode(); node != nil {
					id := node.ID
					return m, func() tea.Msg { return OpenDetailMsg{ID: id} }
				}
				return m, nil
			}
		}
		if m.currentScreen == AgentsScreen {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				m.agentsModel.moveDown() // Natural scrolling (trackpad)
				return m, nil
			case tea.MouseButtonWheelDown:
				m.agentsModel.moveUp() // Natural scrolling (trackpad)
				return m, nil
			case tea.MouseButtonLeft:
				// Tap opens detail for current cursor position
				if id := m.agentsModel.selectedElementID(); id != "" {
					return m, func() tea.Msg { return OpenDetailMsg{ID: id} }
				}
				return m, nil
			}
		}
		if m.currentScreen == DetailScreen {
			// Forward mouse wheel to viewport for trackpad scrolling
			var cmd tea.Cmd
			m.detailModel, cmd = m.detailModel.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.settingsModel.SetSize(msg.Width, msg.Height)
		m.detailModel.SetSize(msg.Width, msg.Height)
		m.agentsModel.SetSize(msg.Width, msg.Height)
		m.commandModel.SetWidth(msg.Width)

	case treeLoadedMsg:
		// Preserve expanded state from current tree before replacing
		var expandedIDs []string
		if len(m.tree) > 0 {
			expandedIDs = CollectExpandedNodes(m.tree)
		} else if !m.configLoaded && m.config != nil && len(m.config.ExpandedNodes) > 0 {
			// First load: use config
			expandedIDs = m.config.ExpandedNodes
			m.configLoaded = true
		}

		m.tree = msg.tree
		m.loadError = msg.loadErr
		m.refreshing = false

		// Only update lastUpdate time if load was successful
		if msg.loadErr == nil {
			m.lastUpdate = time.Now()
		}

		// Restore expanded state to new tree
		if len(expandedIDs) > 0 {
			ApplyExpandedNodes(m.tree, expandedIDs)
		}
		m.boardRebuildRows()

	case tickMsg:
		// Auto-refresh based on current screen
		if m.currentScreen == AgentsScreen {
			// Refresh agents data
			return m, tea.Batch(LoadAgentsWithFilter(m.getStaleMinutes()), m.tickCmd())
		}
		// Refresh board tree
		m.refreshing = true
		return m, tea.Batch(loadTree, m.tickCmd())

	case updateTimeMsg:
		// Just trigger a re-render to update the "Updated Xs ago" display
		return m, m.updateTimeTickCmd()

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case OpenDetailMsg:
		// Shared intent: open detail from any screen
		// Guard: trackpad tap can generate duplicate events — don't overwrite previousScreen
		if m.currentScreen != DetailScreen {
			m.previousScreen = m.currentScreen
		}
		m.detailModel = NewDetailModel()
		m.detailModel.SetSize(m.width, m.height)
		m.currentScreen = DetailScreen
		return m, m.detailModel.LoadElement(msg.ID)

	case DetailCloseMsg:
		// Return to the screen that opened detail
		m.currentScreen = m.previousScreen
		return m, nil

	case elementLoadedMsg:
		// Forward to detail model
		var cmd tea.Cmd
		m.detailModel, cmd = m.detailModel.Update(msg)
		return m, cmd

	case SettingsCloseMsg:
		// Return to board screen
		m.currentScreen = BoardScreen

		// Update refresh interval if changed
		if msg.RefreshChanged {
			if msg.NewInterval > 0 {
				m.refreshInterval = msg.NewInterval
				if m.config != nil {
					m.config.RefreshRate = int(msg.NewInterval.Seconds())
				}
			} else {
				// Refresh disabled - set a very long interval
				m.refreshInterval = 24 * time.Hour
				if m.config != nil {
					m.config.RefreshRate = 0
				}
			}
		}

		// Update agents filter if changed
		if msg.AgentsChanged {
			m.agentsFilter = msg.NewAgentsFilter
			if m.config != nil {
				m.config.AgentsFilter = msg.NewAgentsFilter
			}
		}

		// Restart tick if refresh changed
		if msg.RefreshChanged && msg.NewInterval > 0 {
			return m, m.tickCmd()
		}
		return m, nil

	case AgentsCloseMsg:
		m.currentScreen = BoardScreen
		return m, nil

	case AgentsLoadedMsg:
		var cmd tea.Cmd
		m.agentsModel, cmd = m.agentsModel.Update(msg)
		return m, cmd

	case CommandCancelMsg:
		// Command cancelled, nothing to do
		if m.logger != nil {
			m.logger.Action("command_palette", "cancelled")
		}
		return m, nil

	case CommandExecuteMsg:
		// Execute the command
		if m.logger != nil {
			m.logger.Info("CommandExecuteMsg received: cmd=%q args=%q", msg.Command, msg.Args)
		}
		return m.executeCommand(msg.Command, msg.Args)
	}

	return m, nil
}

// executeCommand handles slash commands
func (m *model) executeCommand(cmd, args string) (tea.Model, tea.Cmd) {
	if m.logger != nil {
		m.logger.Info("executeCommand: cmd=%q args=%q", cmd, args)
	}

	switch cmd {
	case "filter":
		if m.logger != nil {
			m.logger.Command("filter", args, "filter not yet implemented")
		}
		// TODO: implement custom filter (TASK-260206-2tovz9)
		return m, nil

	case "agents":
		if m.logger != nil {
			m.logger.Command("agents", "", "opening agents screen")
		}
		m.agentsModel = NewAgentsModel()
		m.agentsModel.SetSize(m.width, m.height)
		m.currentScreen = AgentsScreen
		return m, LoadAgentsWithFilter(m.getStaleMinutes())

	case "help":
		if m.logger != nil {
			m.logger.Command("help", "", "showing help")
		}
		// Show help in detail view
		m.previousScreen = m.currentScreen
		m.detailModel = NewDetailModel()
		m.detailModel.SetSize(m.width, m.height)
		m.detailModel.SetHelpContent(m.commandModel.GetCommands())
		m.currentScreen = DetailScreen
		return m, nil

	case "refresh":
		if !m.refreshing {
			if m.logger != nil {
				m.logger.Command("refresh", "", "starting refresh")
			}
			m.refreshing = true
			return m, loadTree
		}
		return m, nil

	case "settings":
		if m.logger != nil {
			m.logger.Command("settings", "", "opening settings screen")
		}
		m.settingsModel = NewSettingsModelWithAgents(m.refreshInterval, m.agentsFilter, nil)
		m.settingsModel.SetSize(m.width, m.height)
		m.currentScreen = SettingsScreen
		return m, nil

	case "expand":
		if m.logger != nil {
			m.logger.Command("expand", "", "expanding all")
		}
		for _, root := range m.tree {
			root.ExpandAll()
		}
		m.boardRebuildRows()
		return m, nil

	case "collapse":
		if m.logger != nil {
			m.logger.Command("collapse", "", "collapsing all")
		}
		for _, root := range m.tree {
			root.CollapseAll()
		}
		m.boardRebuildRows()
		return m, nil

	default:
		if m.logger != nil {
			m.logger.Warn("Unknown command: %q", cmd)
		}
	}

	return m, nil
}

// getStaleMinutes converts agentsFilter index to stale minutes
func (m *model) getStaleMinutes() int {
	opts := DefaultAgentsOptions()
	if m.agentsFilter >= 0 && m.agentsFilter < len(opts) {
		return opts[m.agentsFilter].StaleMinutes
	}
	return 0
}


// saveConfigOnQuit saves the current configuration before quitting
func (m *model) saveConfigOnQuit() {
	// Close logger
	if m.logger != nil {
		m.logger.Info("Saving config and closing session")
		m.logger.Close()
	}

	if m.config == nil {
		return
	}
	// Update expanded nodes from current tree state
	if len(m.tree) > 0 {
		m.config.SetExpandedNodes(CollectExpandedNodes(m.tree))
	}
	// Update refresh rate
	m.config.RefreshRate = int(m.refreshInterval.Seconds())
	// Save to file (ignore errors on quit)
	_ = m.config.SaveConfig()
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}
	if m.quitting {
		return ""
	}

	switch m.currentScreen {
	case SettingsScreen:
		return m.viewSettings()
	case DetailScreen:
		return m.detailModel.View()
	case AgentsScreen:
		return m.agentsModel.View()
	default:
		return m.viewBoard()
	}
}

// viewBoard renders the main board view
func (m model) viewBoard() string {
	// Title
	title := titleStyle.Render(" Task Board ")

	// Status indicator
	var statusInfo string
	if m.refreshing {
		statusInfo = statusBarStyle.Render(" Refreshing... ")
	} else if m.loadError != nil {
		statusInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4500")).
			Render(fmt.Sprintf(" Offline (last update: %s) ", m.formatTimeSince()))
	} else if !m.lastUpdate.IsZero() {
		statusInfo = statusBarStyle.Render(fmt.Sprintf(" Updated %s ", m.formatTimeSince()))
	}

	// Content rows
	vh := m.boardVisibleHeight()
	var content strings.Builder

	if len(m.boardRows) == 0 {
		content.WriteString("  Loading board...\n")
		for i := 1; i < vh; i++ {
			content.WriteByte('\n')
		}
	} else {
		end := m.boardScrollOff + vh
		if end > len(m.boardRows) {
			end = len(m.boardRows)
		}
		for i := m.boardScrollOff; i < end; i++ {
			row := m.boardRows[i]
			line := row.text
			if i == m.boardSelectedIdx && row.selectable() {
				plain := lipgloss.NewStyle().Width(m.width - 4).Render(line)
				line = cursorStyle.Render(plain)
			}
			content.WriteString(line)
			content.WriteByte('\n')
		}
		// Pad remaining lines
		rendered := end - m.boardScrollOff
		for i := rendered; i < vh; i++ {
			content.WriteByte('\n')
		}
	}

	// Help
	help := helpStyle.Render("  ↑↓: navigate | enter/o: open | space: toggle | /: commands | q: quit")

	// Command palette overlay
	if m.commandModel.IsActive() {
		return appStyle.Render(title + statusInfo + "\n\n" + content.String() + "\n" + m.commandModel.View())
	}

	// Quit confirmation overlay
	if m.confirmQuit {
		return appStyle.Render(title + statusInfo + "\n\n" + content.String() + "\n" + m.renderQuitDialog())
	}

	return appStyle.Render(title + statusInfo + "\n\n" + content.String() + help)
}

// renderQuitDialog renders the quit confirmation dialog
func (m model) renderQuitDialog() string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF4500")).
		Padding(1, 3)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF4500")).
		Render("Quit?")

	// Button styles
	selectedBtn := lipgloss.NewStyle().
		Background(lipgloss.Color("#6C5CE7")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 2)

	normalBtn := lipgloss.NewStyle().
		Background(lipgloss.Color("#353533")).
		Foreground(lipgloss.Color("#AAAAAA")).
		Padding(0, 2)

	var noBtn, yesBtn string
	if m.confirmSelection == 0 {
		noBtn = selectedBtn.Render(" No ")
		yesBtn = normalBtn.Render(" Yes ")
	} else {
		noBtn = normalBtn.Render(" No ")
		yesBtn = selectedBtn.Render(" Yes ")
	}

	buttons := noBtn + "  " + yesBtn

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("←/→: select • enter: confirm • esc: cancel")

	return dialogStyle.Render(title + "\n\n" + buttons + "\n\n" + hint)
}

// viewSettings renders the settings screen using the SettingsModel
func (m model) viewSettings() string {
	return m.settingsModel.View()
}

// formatTimeSince returns a human-readable string for time since last update
func (m model) formatTimeSince() string {
	if m.lastUpdate.IsZero() {
		return "never"
	}
	d := time.Since(m.lastUpdate)
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh ago", int(d.Hours()))
}

// Messages
type treeLoadedMsg struct {
	tree    []*TreeNode
	loadErr error // Non-nil if load failed (for offline state)
}

type errMsg struct {
	err error
}

// Commands
func loadTree() tea.Msg {
	tree, err := LoadTreeFromCLI()
	if err != nil {
		// Fallback to demo data if CLI not available, but record the error
		return treeLoadedMsg{tree: GetDemoTree(), loadErr: err}
	}
	return treeLoadedMsg{tree: tree, loadErr: nil}
}

// getStatusSymbol returns a visual symbol for status (used in compact views)
func getStatusSymbol(status string) string {
	switch status {
	case "done":
		return "[x]"
	case "development", "progress":
		return "[~]"
	case "blocked":
		return "[!]"
	case "review":
		return "[?]"
	default:
		return "[ ]"
	}
}

// formatTreeLine formats a single tree line with proper indentation
func formatTreeLine(prefix, id, name, status string, hasChildren, expanded bool) string {
	expandChar := " "
	if hasChildren {
		if expanded {
			expandChar = "v"
		} else {
			expandChar = ">"
		}
	}

	statusSym := getStatusSymbol(status)
	return fmt.Sprintf("%s%s %s %s %s", prefix, expandChar, statusSym, id, truncate(name, 40))
}

// truncate truncates a string to max length with ellipsis
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// countByStatus counts items by status in the tree
func countByStatus(roots []*TreeNode) map[string]int {
	counts := make(map[string]int)
	for _, root := range roots {
		countStatusRecursive(root, counts)
	}
	return counts
}

func countStatusRecursive(node *TreeNode, counts map[string]int) {
	counts[node.Status]++
	for _, child := range node.Children {
		countStatusRecursive(child, counts)
	}
}

// buildStatusBar creates a status summary bar
func buildStatusBar(roots []*TreeNode) string {
	counts := countByStatus(roots)
	total := 0
	for _, c := range counts {
		total += c
	}

	parts := []string{}
	for status, count := range counts {
		if count > 0 {
			style, ok := statusStyles[status]
			if !ok {
				style = lipgloss.NewStyle()
			}
			parts = append(parts, style.Render(fmt.Sprintf("%s:%d", status, count)))
		}
	}

	return fmt.Sprintf("Total: %d | %s", total, strings.Join(parts, " | "))
}

func main() {
	// Warm up markdown renderer in background
	InitMarkdownRenderer()

	// Initialize logger
	logger := NewLogger(".task-board/logs")
	if err := logger.Open(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to open logger: %v\n", err)
	}

	// Load configuration (uses defaults if file doesn't exist)
	cfg, err := LoadConfig()
	if err != nil {
		// Log error but continue with defaults
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		logger.Warn("Failed to load config: %v", err)
	}

	// Determine refresh interval from config
	refreshInterval := cfg.GetRefreshDuration()
	if refreshInterval == 0 {
		// If "Off", use a very long interval (effectively disabled)
		// We still need a tick for the time display update
		refreshInterval = 24 * time.Hour
	}

	logger.Info("Config loaded, refresh interval: %v", refreshInterval)

	m := model{
		refreshInterval: refreshInterval,
		config:          cfg,
		logger:          logger,
		agentsFilter:    cfg.AgentsFilter,
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
