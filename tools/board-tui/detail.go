package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Global markdown renderer (singleton)
var (
	mdRenderer     *glamour.TermRenderer
	mdRendererOnce sync.Once
)

func getMarkdownRenderer() *glamour.TermRenderer {
	mdRendererOnce.Do(func() {
		var err error
		mdRenderer, err = glamour.NewTermRenderer(
			glamour.WithStylePath("dark"),
			glamour.WithWordWrap(80),
		)
		if err != nil {
			mdRenderer = nil
		}
	})
	return mdRenderer
}

// InitMarkdownRenderer initializes the renderer at startup (call from main)
func InitMarkdownRenderer() {
	getMarkdownRenderer() // warm up synchronously
}

// ElementDetail holds the full detail from task-board show --json
type ElementDetail struct {
	ID                 string          `json:"id"`
	Type               string          `json:"type"`
	Name               string          `json:"name"`
	Status             string          `json:"status"`
	Assignee           string          `json:"assignee"`
	Parent             string          `json:"parent"`
	Path               string          `json:"path"`
	CreatedAt          string          `json:"createdAt"`
	UpdatedAt          string          `json:"updatedAt"`
	BlockedBy          []string        `json:"blockedBy"`
	Blocks             []string        `json:"blocks"`
	Description        string          `json:"description"`
	AcceptanceCriteria string          `json:"acceptanceCriteria"`
	Checklist          []ChecklistItem `json:"checklist"`
	Notes              []NoteItem      `json:"notes"`
}

type ChecklistItem struct {
	Text string `json:"text"`
	Done bool   `json:"done"`
}

type NoteItem struct {
	Timestamp string `json:"timestamp"`
	Text      string `json:"text"`
}

// ShowResponse is the JSON response from task-board show --json
type ShowResponse struct {
	Element ElementDetail `json:"element"`
}

// DetailModel is the model for the detail view
type DetailModel struct {
	viewport    viewport.Model
	element     *ElementDetail
	loading     bool
	err         error
	width       int
	height      int
	ready       bool
	title       string // Custom title (for help screen)
	helpContent string // Static help content (bypasses element loading)
}

// DetailCloseMsg signals returning to the board
type DetailCloseMsg struct{}

// elementLoadedMsg carries loaded element data
type elementLoadedMsg struct {
	element *ElementDetail
	err     error
}

// NewDetailModel creates a new detail view model
func NewDetailModel() DetailModel {
	return DetailModel{
		loading: true,
	}
}

// LoadElement starts loading element details
func (m *DetailModel) LoadElement(id string) tea.Cmd {
	m.loading = true
	m.err = nil
	return func() tea.Msg {
		element, err := loadElementFromCLI(id)
		return elementLoadedMsg{element: element, err: err}
	}
}

// SetHelpContent sets static help content (for /help command)
func (m *DetailModel) SetHelpContent(commands []Command) {
	m.loading = false
	m.title = "Help - Available Commands"

	var sb strings.Builder
	sb.WriteString("# Available Commands\n\n")
	sb.WriteString("Press `/` or `.` to open the command palette.\n\n")

	for _, cmd := range commands {
		sb.WriteString(fmt.Sprintf("## /%s\n%s\n\n", cmd.Name, cmd.Description))
	}

	sb.WriteString("---\n\n")
	sb.WriteString("# Keyboard Shortcuts\n\n")
	sb.WriteString("| Key | Action |\n")
	sb.WriteString("|-----|--------|\n")
	sb.WriteString("| `j/k` or `↑/↓` | Navigate up/down |\n")
	sb.WriteString("| `space` | Toggle expand/collapse |\n")
	sb.WriteString("| `enter` or `o` | Open detail view |\n")
	sb.WriteString("| `e` | Expand all |\n")
	sb.WriteString("| `c` | Collapse all |\n")
	sb.WriteString("| `g` | Go to top |\n")
	sb.WriteString("| `G` | Go to bottom |\n")
	sb.WriteString("| `r` | Refresh data |\n")
	sb.WriteString("| `q` or `esc` | Quit / Go back |\n")

	m.helpContent = sb.String()
}

// loadElementFromCLI calls task-board show ID --json
func loadElementFromCLI(id string) (*ElementDetail, error) {
	cmd := exec.Command("task-board", "show", id, "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var response ShowResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, err
	}

	return &response.Element, nil
}

// SetSize sets the viewport size
func (m *DetailModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	headerHeight := 3 // Title bar
	footerHeight := 2 // Help line

	if !m.ready {
		m.viewport = viewport.New(width-4, height-headerHeight-footerHeight)
		m.viewport.YPosition = headerHeight
		m.ready = true
	} else {
		m.viewport.Width = width - 4
		m.viewport.Height = height - headerHeight - footerHeight
	}

	// Render help content if present
	if m.helpContent != "" {
		rendered := m.helpContent
		if r := getMarkdownRenderer(); r != nil {
			if out, err := r.Render(m.helpContent); err == nil {
				rendered = out
			}
		}
		m.viewport.SetContent(rendered)
		return
	}

	// Re-render content if we have element
	if m.element != nil {
		m.viewport.SetContent(m.renderContent())
	}
}

// Update handles messages
func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "enter", "o":
			return m, func() tea.Msg { return DetailCloseMsg{} }
		}

	case elementLoadedMsg:
		m.loading = false
		m.err = msg.err
		m.element = msg.element
		if m.ready && m.element != nil {
			m.viewport.SetContent(m.renderContent())
		}
		return m, nil
	}

	// Handle viewport scrolling
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// renderContent builds the markdown content for the viewport
func (m *DetailModel) renderContent() string {
	if m.element == nil {
		return ""
	}

	e := m.element
	var sb strings.Builder

	// Header with type indicator and status
	typeInd := typeIndicators[e.Type]
	if typeInd == "" {
		typeInd = "?"
	}

	sb.WriteString(fmt.Sprintf("# %s %s\n\n", typeInd, e.Name))
	sb.WriteString(fmt.Sprintf("**Status:** %s", e.Status))
	if e.Assignee != "" {
		sb.WriteString(fmt.Sprintf("  |  **Assignee:** @%s", e.Assignee))
	}
	sb.WriteString("\n\n")

	if e.Path != "" {
		sb.WriteString(fmt.Sprintf("**Path:** `%s`\n\n", e.Path))
	}

	// Dependencies
	if len(e.BlockedBy) > 0 {
		sb.WriteString(fmt.Sprintf("**Blocked by:** %s\n\n", strings.Join(e.BlockedBy, ", ")))
	}
	if len(e.Blocks) > 0 {
		sb.WriteString(fmt.Sprintf("**Blocks:** %s\n\n", strings.Join(e.Blocks, ", ")))
	}

	sb.WriteString("---\n\n")

	// Description
	if e.Description != "" && e.Description != "(no description)" {
		sb.WriteString("## Description\n\n")
		sb.WriteString(e.Description)
		sb.WriteString("\n\n")
	}

	// Acceptance Criteria
	if e.AcceptanceCriteria != "" && e.AcceptanceCriteria != "(define acceptance criteria)" {
		sb.WriteString("## Acceptance Criteria\n\n")
		sb.WriteString(e.AcceptanceCriteria)
		sb.WriteString("\n\n")
	}

	// Checklist
	if len(e.Checklist) > 0 {
		sb.WriteString("## Checklist\n\n")
		for _, item := range e.Checklist {
			if item.Done {
				sb.WriteString(fmt.Sprintf("- [x] %s\n", item.Text))
			} else {
				sb.WriteString(fmt.Sprintf("- [ ] %s\n", item.Text))
			}
		}
		sb.WriteString("\n")
	}

	// Notes
	if len(e.Notes) > 0 {
		sb.WriteString("## Notes\n\n")
		for _, note := range e.Notes {
			if note.Timestamp != "" {
				sb.WriteString(fmt.Sprintf("**%s**\n\n", note.Timestamp))
			}
			sb.WriteString(note.Text)
			sb.WriteString("\n\n---\n\n")
		}
	}

	// Render markdown
	renderer := getMarkdownRenderer()
	if renderer == nil {
		return sb.String() // fallback to plain text
	}
	rendered, err := renderer.Render(sb.String())
	if err != nil {
		return sb.String() // fallback to plain text
	}

	return rendered
}

// View renders the detail screen
func (m DetailModel) View() string {
	if m.loading {
		return "\n  Loading..."
	}

	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n\n  Press q or esc to go back", m.err)
	}

	// Check for help content first
	if m.helpContent != "" {
		// Render help content if not already done
		if !m.ready {
			return "\n  Preparing help..."
		}

		// Title bar
		title := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#6C5CE7")).
			Padding(0, 1).
			Render(fmt.Sprintf(" %s ", m.title))

		// Footer
		scrollPercent := fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)
		help := helpStyle.Render("  j/k: scroll | esc/q: back")
		footer := lipgloss.JoinHorizontal(lipgloss.Left, help, "  ", scrollPercent)

		return fmt.Sprintf("%s\n\n%s\n\n%s", title, m.viewport.View(), footer)
	}

	if m.element == nil {
		return "\n  No element loaded"
	}

	// Title bar
	titleText := m.element.ID
	if m.title != "" {
		titleText = m.title
	}
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#6C5CE7")).
		Padding(0, 1).
		Render(fmt.Sprintf(" %s ", titleText))

	// Footer with scroll percentage and help
	scrollPercent := fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)
	help := helpStyle.Render("  j/k: scroll | esc/q: back")
	footer := lipgloss.JoinHorizontal(lipgloss.Left, help, "  ", scrollPercent)

	return fmt.Sprintf("%s\n\n%s\n\n%s", title, m.viewport.View(), footer)
}
