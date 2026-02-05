package agents

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"board-tui/internal/styles"
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

type agentsResponse struct {
	Agents      []AgentInfo `json:"agents"`
	TotalAgents int         `json:"totalAgents"`
}

// CloseMsg signals returning to board
type CloseMsg struct{}

// LoadedMsg carries loaded agent data
type LoadedMsg struct {
	Agents []AgentInfo
	Err    error
}

// Model displays the agents dashboard
type Model struct {
	viewport   viewport.Model
	agents     []AgentInfo
	loading    bool
	err        error
	width      int
	height     int
	ready      bool
	lastUpdate time.Time
}

// New creates a new agents view
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

	if len(m.agents) > 0 {
		m.viewport.SetContent(m.renderContent())
	}
}

// Load fetches agent data from CLI
func Load() tea.Msg {
	cmd := exec.Command("task-board", "agents", "--json")
	output, err := cmd.Output()
	if err != nil {
		return LoadedMsg{Err: err}
	}

	var response agentsResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return LoadedMsg{Err: err}
	}

	// Load tree to get children for stories/epics
	childrenMap := loadChildrenMap()

	// Enrich assigned elements with children
	for i := range response.Agents {
		for j := range response.Agents[i].AssignedElements {
			elem := &response.Agents[i].AssignedElements[j]
			if elem.Type == "story" || elem.Type == "epic" {
				elem.Children = childrenMap[elem.ID]
			}
		}
	}

	return LoadedMsg{Agents: response.Agents}
}

// Tree structures for loading children
type treeNode struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Name     string      `json:"name"`
	Status   string      `json:"status"`
	Children []*treeNode `json:"children"`
}

type treeResponse struct {
	Tree []*treeNode `json:"tree"`
}

func loadChildrenMap() map[string][]ChildElement {
	result := make(map[string][]ChildElement)

	cmd := exec.Command("task-board", "tree", "--json")
	output, err := cmd.Output()
	if err != nil {
		return result
	}

	var treeResp treeResponse
	if err := json.Unmarshal(output, &treeResp); err != nil {
		return result
	}

	var collectChildren func(node *treeNode)
	collectChildren = func(node *treeNode) {
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

// Update handles messages
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "enter":
			return m, func() tea.Msg { return CloseMsg{} }
		}

	case LoadedMsg:
		m.loading = false
		m.err = msg.Err
		m.agents = msg.Agents
		sort.Slice(m.agents, func(i, j int) bool {
			return m.agents[i].Name < m.agents[j].Name
		})
		if msg.Err == nil {
			m.lastUpdate = time.Now()
		}
		if m.ready {
			m.viewport.SetContent(m.renderContent())
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) renderContent() string {
	if len(m.agents) == 0 {
		return "\n  No agents with assigned tasks."
	}

	var sb strings.Builder

	for _, agent := range m.agents {
		staleInfo := ""
		if agent.StaleCount > 0 {
			staleInfo = styles.Stale.Render(fmt.Sprintf(" (%d stale)", agent.StaleCount))
		}
		sb.WriteString(styles.AgentHeader.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━") + "\n")
		sb.WriteString(styles.AgentName.Render("@"+agent.Name) + fmt.Sprintf(" - %d tasks", agent.TotalAssigned) + staleInfo + "\n")
		sb.WriteString(styles.AgentHeader.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━") + "\n\n")

		for _, elem := range agent.AssignedElements {
			typeInd := styles.TypeIndicator[elem.Type]
			if typeInd == "" {
				typeInd = "?"
			}

			statusStyle := styles.Status[elem.Status]

			line := fmt.Sprintf("  %s %s %s %s",
				typeInd,
				elem.ID,
				elem.Name,
				statusStyle.Render("["+elem.Status+"]"))

			if elem.StaleSince != nil {
				line += styles.Stale.Render(" (stale)")
			}

			sb.WriteString(line + "\n")

			for _, child := range elem.Children {
				childTypeInd := styles.TypeIndicator[child.Type]
				if childTypeInd == "" {
					childTypeInd = "?"
				}
				childStatusStyle := styles.Status[child.Status]
				childLine := fmt.Sprintf("    └─ %s %s %s %s",
					childTypeInd,
					child.ID,
					child.Name,
					childStatusStyle.Render("["+child.Status+"]"))
				sb.WriteString(childLine + "\n")
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// View renders the agents screen
func (m Model) View() string {
	if m.loading && len(m.agents) == 0 {
		return "\n  Loading agents..."
	}

	if m.err != nil && len(m.agents) == 0 {
		return fmt.Sprintf("\n  Error: %v\n\n  Press esc to go back", m.err)
	}

	title := styles.Title.Render(" Agents Dashboard ")

	var updateInfo string
	if !m.lastUpdate.IsZero() {
		ago := time.Since(m.lastUpdate)
		if ago < time.Minute {
			updateInfo = fmt.Sprintf(" Updated %ds ago", int(ago.Seconds()))
		} else {
			updateInfo = fmt.Sprintf(" Updated %dm ago", int(ago.Minutes()))
		}
	}
	status := styles.Help.Render(updateInfo)

	help := styles.Help.Render("  j/k: scroll | esc/q/enter: back")

	return fmt.Sprintf("%s%s\n\n%s\n\n%s", title, status, m.viewport.View(), help)
}
