package styles

import "github.com/charmbracelet/lipgloss"

// App-wide styles
var (
	App = lipgloss.NewStyle().Padding(1, 2)

	Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#6C5CE7")).
		Padding(0, 1)

	StatusBar = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	Help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF4500"))
)

// Status colors for elements
var Status = map[string]lipgloss.Style{
	"backlog":     lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")),
	"analysis":    lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF")),
	"to-dev":      lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF")),
	"development": lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")),
	"progress":    lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")),
	"to-review":   lipgloss.NewStyle().Foreground(lipgloss.Color("#9370DB")),
	"reviewing":   lipgloss.NewStyle().Foreground(lipgloss.Color("#9370DB")),
	"review":      lipgloss.NewStyle().Foreground(lipgloss.Color("#9370DB")),
	"done":        lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32")),
	"closed":      lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32")),
	"blocked":     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4500")),
}

// Type indicators with visual symbols
var TypeIndicator = map[string]string{
	"epic":  "◆", // Diamond - major initiative
	"story": "◇", // Empty diamond - feature
	"task":  "○", // Circle - work item
	"bug":   "●", // Filled circle - defect
}

// Type colors for visual distinction
var TypeStyle = map[string]lipgloss.Style{
	"epic":  lipgloss.NewStyle().Foreground(lipgloss.Color("#E040FB")).Bold(true), // Purple/Magenta
	"story": lipgloss.NewStyle().Foreground(lipgloss.Color("#29B6F6")),            // Light Blue
	"task":  lipgloss.NewStyle().Foreground(lipgloss.Color("#66BB6A")),            // Green
	"bug":   lipgloss.NewStyle().Foreground(lipgloss.Color("#EF5350")),            // Red
}

// Dialog styles
var (
	DialogBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 3)

	DialogTitle = lipgloss.NewStyle().
			Bold(true)

	ButtonSelected = lipgloss.NewStyle().
			Background(lipgloss.Color("#6C5CE7")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 2)

	ButtonNormal = lipgloss.NewStyle().
			Background(lipgloss.Color("#353533")).
			Foreground(lipgloss.Color("#AAAAAA")).
			Padding(0, 2)
)

// Command palette styles
var (
	CommandBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6C5CE7")).
			Padding(0, 1)

	CommandSelected = lipgloss.NewStyle().
			Background(lipgloss.Color("#6C5CE7")).
			Foreground(lipgloss.Color("#FFFFFF"))

	CommandNormal = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA"))

	CommandDesc = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

// Dependency styles
var (
	BlockedBy = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4500"))
	Blocks    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	Stale     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4500"))
)

// Agent styles
var (
	AgentName = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#29B6F6"))

	AgentHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#6C5CE7"))
)
