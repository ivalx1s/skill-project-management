package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AgentInfo holds agent dashboard info
type AgentInfo struct {
	Name             string            `json:"name"`
	AssignedElements []AssignedElement `json:"assignedElements"`
	TotalAssigned    int               `json:"totalAssigned"`
	StaleCount       int               `json:"staleCount"`
}

// AssignedElement represents an element assigned to an agent
type AssignedElement struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Name       string         `json:"name"`
	Status     string         `json:"status"`
	UpdatedAt  string         `json:"updatedAt"`
	StaleSince *string        `json:"staleSince"`
	Children   []ChildElement `json:"-"`
}

// ChildElement represents a child of an assigned element
type ChildElement struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// AgentsResponse is the JSON response from task-board agents --json
type AgentsResponse struct {
	Agents      []AgentInfo `json:"agents"`
	TotalAgents int         `json:"totalAgents"`
}

// AgentsLoadedMsg carries loaded agent data
type AgentsLoadedMsg struct {
	Agents []AgentInfo
	Err    error
}

// AgentsCloseMsg signals returning to board
type AgentsCloseMsg struct{}

// OpenDetailMsg is a shared intent to open detail view for any element.
// Can be emitted from any screen; main.go handles it.
type OpenDetailMsg struct {
	ID string
}

// LoadAgents fetches agent data from CLI and enriches with children
func LoadAgents() tea.Msg {
	return loadAgentsWithStale(0)
}

// LoadAgentsWithFilter returns a command that loads agents with stale filter
func LoadAgentsWithFilter(staleMinutes int) tea.Cmd {
	return func() tea.Msg {
		return loadAgentsWithStale(staleMinutes)
	}
}

// loadAgentsWithStale fetches agent data with optional stale filter
func loadAgentsWithStale(staleMinutes int) tea.Msg {
	args := []string{"agents", "--json"}
	if staleMinutes > 0 {
		args = append(args, "--stale", fmt.Sprintf("%d", staleMinutes))
	} else {
		args = append(args, "--all")
	}
	cmd := exec.Command("task-board", args...)
	output, err := cmd.Output()
	if err != nil {
		return AgentsLoadedMsg{Err: err}
	}

	var response AgentsResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return AgentsLoadedMsg{Err: err}
	}

	childrenMap := loadChildrenMap()

	for i := range response.Agents {
		for j := range response.Agents[i].AssignedElements {
			elem := &response.Agents[i].AssignedElements[j]
			if elem.Type == "story" || elem.Type == "epic" {
				elem.Children = childrenMap[elem.ID]
			}
		}
	}

	return AgentsLoadedMsg{Agents: response.Agents}
}

// loadChildrenMap loads tree and builds map of id -> children
func loadChildrenMap() map[string][]ChildElement {
	result := make(map[string][]ChildElement)

	cmd := exec.Command("task-board", "tree", "--json")
	output, err := cmd.Output()
	if err != nil {
		return result
	}

	var treeResp TreeResponse
	if err := json.Unmarshal(output, &treeResp); err != nil {
		return result
	}

	var collectChildren func(node *TreeNode)
	collectChildren = func(node *TreeNode) {
		if len(node.Children) > 0 {
			children := make([]ChildElement, 0, len(node.Children))
			for _, child := range node.Children {
				children = append(children, ChildElement{
					ID:     child.ID,
					Type:   child.Type,
					Name:   child.Name,
					Status: child.Status,
				})
				collectChildren(child)
			}
			result[node.ID] = children
		}
	}

	for _, root := range treeResp.Tree {
		collectChildren(root)
	}

	return result
}

// --- Row data model ---

type rowKind int

const (
	rowHeader    rowKind = iota // agent name line
	rowSeparator                // ━━━ line
	rowElement                  // assigned element (selectable)
	rowChild                    // child of element (selectable)
	rowBlank                    // empty line between agents
)

type agentRow struct {
	kind      rowKind
	elementID string // set for rowElement and rowChild
	text      string // pre-rendered display text (without highlight)
}

func (r agentRow) selectable() bool {
	return r.kind == rowElement || r.kind == rowChild
}

// --- Styles ---

var (
	agentHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#6C5CE7"))

	agentNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#29B6F6"))

	staleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4500"))

	cursorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3B3B5C"))
)

// --- AgentsModel ---

// AgentsModel displays the agents dashboard with cursor navigation
type AgentsModel struct {
	agents      []AgentInfo
	rows        []agentRow
	selectedIdx int
	scrollOff   int
	loading     bool
	err         error
	width       int
	height      int
	lastUpdate  time.Time
}

// NewAgentsModel creates a new agents view
func NewAgentsModel() AgentsModel {
	return AgentsModel{
		loading: true,
	}
}

// SetSize sets the available area
func (m *AgentsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// visibleHeight returns how many content rows fit between title and footer
func (m *AgentsModel) visibleHeight() int {
	// appPadTop(1) + title(1) + blank(1) + footerBlank(1) + footer(1) + appPadBottom(1) = 6
	h := m.height - 6
	if h < 1 {
		h = 1
	}
	return h
}

// buildRows flattens agents into a list of rows
func (m *AgentsModel) buildRows() {
	m.rows = nil

	for i, agent := range m.agents {
		if i > 0 {
			m.rows = append(m.rows, agentRow{kind: rowBlank, text: ""})
		}

		// Separator
		m.rows = append(m.rows, agentRow{
			kind: rowSeparator,
			text: agentHeaderStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"),
		})

		// Agent name header
		staleInfo := ""
		if agent.StaleCount > 0 {
			staleInfo = staleStyle.Render(fmt.Sprintf(" (%d stale)", agent.StaleCount))
		}
		m.rows = append(m.rows, agentRow{
			kind: rowHeader,
			text: agentNameStyle.Render("@"+agent.Name) + fmt.Sprintf(" — %d tasks", agent.TotalAssigned) + staleInfo,
		})

		// Elements
		for _, elem := range agent.AssignedElements {
			typeInd := typeIndicators[elem.Type]
			if typeInd == "" {
				typeInd = "?"
			}
			statusSt, ok := statusStyles[elem.Status]
			if !ok {
				statusSt = lipgloss.NewStyle()
			}

			line := fmt.Sprintf("  %s %s %s %s",
				typeInd, elem.ID, elem.Name,
				statusSt.Render("["+elem.Status+"]"))

			if elem.StaleSince != nil {
				line += staleStyle.Render(" (stale)")
			}

			m.rows = append(m.rows, agentRow{
				kind:      rowElement,
				elementID: elem.ID,
				text:      line,
			})

			// Children
			for _, child := range elem.Children {
				childTypeInd := typeIndicators[child.Type]
				if childTypeInd == "" {
					childTypeInd = "?"
				}
				childStatusSt, ok := statusStyles[child.Status]
				if !ok {
					childStatusSt = lipgloss.NewStyle()
				}
				childLine := fmt.Sprintf("    └─ %s %s %s %s",
					childTypeInd, child.ID, child.Name,
					childStatusSt.Render("["+child.Status+"]"))

				m.rows = append(m.rows, agentRow{
					kind:      rowChild,
					elementID: child.ID,
					text:      childLine,
				})
			}
		}
	}

	// Ensure selectedIdx points to a selectable row
	if m.selectedIdx >= len(m.rows) {
		m.selectedIdx = len(m.rows) - 1
	}
	if m.selectedIdx < 0 {
		m.selectedIdx = 0
	}
	if len(m.rows) > 0 && !m.rows[m.selectedIdx].selectable() {
		m.moveDown()
	}
}

// --- Navigation ---

func (m *AgentsModel) moveDown() {
	for i := m.selectedIdx + 1; i < len(m.rows); i++ {
		if m.rows[i].selectable() {
			m.selectedIdx = i
			m.ensureVisible()
			return
		}
	}
}

func (m *AgentsModel) moveUp() {
	for i := m.selectedIdx - 1; i >= 0; i-- {
		if m.rows[i].selectable() {
			m.selectedIdx = i
			m.ensureVisible()
			return
		}
	}
}

func (m *AgentsModel) goTop() {
	for i := 0; i < len(m.rows); i++ {
		if m.rows[i].selectable() {
			m.selectedIdx = i
			m.scrollOff = 0
			return
		}
	}
}

func (m *AgentsModel) goBottom() {
	for i := len(m.rows) - 1; i >= 0; i-- {
		if m.rows[i].selectable() {
			m.selectedIdx = i
			m.ensureVisible()
			return
		}
	}
}

func (m *AgentsModel) ensureVisible() {
	vh := m.visibleHeight()
	if m.selectedIdx < m.scrollOff {
		m.scrollOff = m.selectedIdx
	}
	if m.selectedIdx >= m.scrollOff+vh {
		m.scrollOff = m.selectedIdx - vh + 1
	}
}

// selectedElementID returns the ID of the currently selected element, or ""
func (m *AgentsModel) selectedElementID() string {
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.rows) {
		return m.rows[m.selectedIdx].elementID
	}
	return ""
}

// Update handles messages
func (m AgentsModel) Update(msg tea.Msg) (AgentsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg { return AgentsCloseMsg{} }

		case "enter":
			if id := m.selectedElementID(); id != "" {
				return m, func() tea.Msg { return OpenDetailMsg{ID: id} }
			}
			return m, func() tea.Msg { return AgentsCloseMsg{} }

		case "down":
			m.moveDown()
			return m, nil

		case "up":
			m.moveUp()
			return m, nil
		}

	case AgentsLoadedMsg:
		m.loading = false
		m.err = msg.Err
		m.agents = msg.Agents
		sort.Slice(m.agents, func(i, j int) bool {
			return m.agents[i].Name < m.agents[j].Name
		})
		if msg.Err == nil {
			m.lastUpdate = time.Now()
		}
		m.buildRows()
		return m, nil
	}

	return m, nil
}

// View renders the agents screen
func (m AgentsModel) View() string {
	// Title
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#6C5CE7")).
		Padding(0, 1).
		Render(" Agents Dashboard ")

	var updateInfo string
	if !m.lastUpdate.IsZero() {
		ago := time.Since(m.lastUpdate)
		if ago < time.Minute {
			updateInfo = fmt.Sprintf(" Updated %ds ago", int(ago.Seconds()))
		} else {
			updateInfo = fmt.Sprintf(" Updated %dm ago", int(ago.Minutes()))
		}
	}
	status := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render(updateInfo)

	help := helpStyle.Render("  ↑↓: navigate | enter: detail | esc/q: back")

	if m.loading && len(m.agents) == 0 {
		return appStyle.Render(fmt.Sprintf("%s%s\n\n  Loading agents...\n\n%s", title, status, help))
	}

	if m.err != nil && len(m.agents) == 0 {
		return appStyle.Render(fmt.Sprintf("%s%s\n\n  Error: %v\n\n%s", title, status, m.err, help))
	}

	// Content rows
	vh := m.visibleHeight()
	var content strings.Builder

	if len(m.rows) == 0 {
		content.WriteString("\n  No agents with assigned tasks.")
	} else {
		end := m.scrollOff + vh
		if end > len(m.rows) {
			end = len(m.rows)
		}
		for i := m.scrollOff; i < end; i++ {
			row := m.rows[i]
			line := row.text
			if i == m.selectedIdx && row.selectable() {
				// Pad to full width for consistent highlight
				plain := lipgloss.NewStyle().Width(m.width - 4).Render(line)
				line = cursorStyle.Render(plain)
			}
			content.WriteString(line)
			content.WriteByte('\n')
		}
		// Pad remaining lines
		rendered := end - m.scrollOff
		for i := rendered; i < vh; i++ {
			content.WriteByte('\n')
		}
	}

	return appStyle.Render(fmt.Sprintf("%s%s\n\n%s\n%s", title, status, content.String(), help))
}
