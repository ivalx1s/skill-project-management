package detail

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
	"board-tui/internal/styles"
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

// InitMarkdownRenderer initializes the renderer at startup
func InitMarkdownRenderer() {
	getMarkdownRenderer()
}

// Element holds the full detail from task-board show --json
type Element struct {
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

type showResponse struct {
	Element Element `json:"element"`
}

// CloseMsg signals returning to the board
type CloseMsg struct{}

// elementLoadedMsg carries loaded element data
type elementLoadedMsg struct {
	element *Element
	err     error
}

// HelpCommand represents a command for help display
type HelpCommand struct {
	Name        string
	Description string
}

// Model is the model for the detail view
type Model struct {
	viewport    viewport.Model
	element     *Element
	loading     bool
	err         error
	width       int
	height      int
	ready       bool
	title       string
	helpContent string
}

// New creates a new detail view model
func New() Model {
	return Model{
		loading: true,
	}
}

// SetSize sets the viewport size
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	headerHeight := 3
	footerHeight := 2

	if !m.ready {
		m.viewport = viewport.New(width-4, height-headerHeight-footerHeight)
		m.viewport.YPosition = headerHeight
		m.ready = true
	} else {
		m.viewport.Width = width - 4
		m.viewport.Height = height - headerHeight - footerHeight
	}

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

	if m.element != nil {
		m.viewport.SetContent(m.renderContent())
	}
}

// LoadElement starts loading element details
func (m *Model) LoadElement(id string) tea.Cmd {
	m.loading = true
	m.err = nil
	return func() tea.Msg {
		element, err := loadElementFromCLI(id)
		return elementLoadedMsg{element: element, err: err}
	}
}

// SetHelpContent sets static help content
func (m *Model) SetHelpContent(commands []HelpCommand) {
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

func loadElementFromCLI(id string) (*Element, error) {
	cmd := exec.Command("task-board", "show", id, "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var response showResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, err
	}

	return &response.Element, nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "enter", "o":
			return m, func() tea.Msg { return CloseMsg{} }
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

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) renderContent() string {
	if m.element == nil {
		return ""
	}

	e := m.element
	var sb strings.Builder

	typeInd := styles.TypeIndicator[e.Type]
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

	if len(e.BlockedBy) > 0 {
		sb.WriteString(fmt.Sprintf("**Blocked by:** %s\n\n", strings.Join(e.BlockedBy, ", ")))
	}
	if len(e.Blocks) > 0 {
		sb.WriteString(fmt.Sprintf("**Blocks:** %s\n\n", strings.Join(e.Blocks, ", ")))
	}

	sb.WriteString("---\n\n")

	if e.Description != "" && e.Description != "(no description)" {
		sb.WriteString("## Description\n\n")
		sb.WriteString(e.Description)
		sb.WriteString("\n\n")
	}

	if e.AcceptanceCriteria != "" && e.AcceptanceCriteria != "(define acceptance criteria)" {
		sb.WriteString("## Acceptance Criteria\n\n")
		sb.WriteString(e.AcceptanceCriteria)
		sb.WriteString("\n\n")
	}

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

	renderer := getMarkdownRenderer()
	if renderer == nil {
		return sb.String()
	}
	rendered, err := renderer.Render(sb.String())
	if err != nil {
		return sb.String()
	}

	return rendered
}

// View renders the detail screen
func (m Model) View() string {
	if m.loading {
		return "\n  Loading..."
	}

	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n\n  Press q or esc to go back", m.err)
	}

	if m.helpContent != "" {
		if !m.ready {
			return "\n  Preparing help..."
		}

		title := styles.Title.Render(fmt.Sprintf(" %s ", m.title))
		scrollPercent := fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)
		help := styles.Help.Render("  j/k: scroll | esc/q: back")
		footer := lipgloss.JoinHorizontal(lipgloss.Left, help, "  ", scrollPercent)

		return fmt.Sprintf("%s\n\n%s\n\n%s", title, m.viewport.View(), footer)
	}

	if m.element == nil {
		return "\n  No element loaded"
	}

	titleText := m.element.ID
	if m.title != "" {
		titleText = m.title
	}
	title := styles.Title.Render(fmt.Sprintf(" %s ", titleText))

	scrollPercent := fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)
	help := styles.Help.Render("  j/k: scroll | esc/q: back")
	footer := lipgloss.JoinHorizontal(lipgloss.Left, help, "  ", scrollPercent)

	return fmt.Sprintf("%s\n\n%s\n\n%s", title, m.viewport.View(), footer)
}
