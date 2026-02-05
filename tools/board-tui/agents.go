package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AgentsModel displays the agents dashboard
type AgentsModel struct {
	viewport   viewport.Model
	agents     []AgentInfo
	loading    bool
	err        error
	width      int
	height     int
	ready      bool
	lastUpdate time.Time
}

// AgentsCloseMsg signals returning to board
type AgentsCloseMsg struct{}

// NewAgentsModel creates a new agents view
func NewAgentsModel() AgentsModel {
	return AgentsModel{
		loading: true,
	}
}

// SetSize sets the viewport size
func (m *AgentsModel) SetSize(width, height int) {
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

// Update handles messages
func (m AgentsModel) Update(msg tea.Msg) (AgentsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "enter":
			return m, func() tea.Msg { return AgentsCloseMsg{} }
		}

	case AgentsLoadedMsg:
		m.loading = false
		m.err = msg.Err
		m.agents = msg.Agents
		// Sort agents by name for consistent ordering
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

// renderContent builds the agents dashboard content
func (m *AgentsModel) renderContent() string {
	if len(m.agents) == 0 {
		return "\n  No agents with assigned tasks."
	}

	var sb strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#6C5CE7"))

	agentStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#29B6F6"))

	staleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF4500"))

	for _, agent := range m.agents {
		// Agent header
		staleInfo := ""
		if agent.StaleCount > 0 {
			staleInfo = staleStyle.Render(fmt.Sprintf(" (%d stale)", agent.StaleCount))
		}
		sb.WriteString(headerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━") + "\n")
		sb.WriteString(agentStyle.Render("@"+agent.Name) + fmt.Sprintf(" - %d tasks", agent.TotalAssigned) + staleInfo + "\n")
		sb.WriteString(headerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━") + "\n\n")

		// Assigned elements
		for _, elem := range agent.AssignedElements {
			typeInd := typeIndicators[elem.Type]
			if typeInd == "" {
				typeInd = "?"
			}

			statusStyle, ok := statusStyles[elem.Status]
			if !ok {
				statusStyle = lipgloss.NewStyle()
			}

			line := fmt.Sprintf("  %s %s %s %s",
				typeInd,
				elem.ID,
				elem.Name,
				statusStyle.Render("["+elem.Status+"]"))

			if elem.StaleSince != nil {
				line += staleStyle.Render(" (stale)")
			}

			sb.WriteString(line + "\n")

			// Show children for stories/epics
			for _, child := range elem.Children {
				childTypeInd := typeIndicators[child.Type]
				if childTypeInd == "" {
					childTypeInd = "?"
				}
				childStatusStyle, ok := statusStyles[child.Status]
				if !ok {
					childStatusStyle = lipgloss.NewStyle()
				}
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
func (m AgentsModel) View() string {
	if m.loading && len(m.agents) == 0 {
		return "\n  Loading agents..."
	}

	if m.err != nil && len(m.agents) == 0 {
		return fmt.Sprintf("\n  Error: %v\n\n  Press esc to go back", m.err)
	}

	// Title bar
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#6C5CE7")).
		Padding(0, 1).
		Render(" Agents Dashboard ")

	// Update indicator
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

	// Footer
	help := helpStyle.Render("  j/k: scroll | esc/q/enter: back")

	return fmt.Sprintf("%s%s\n\n%s\n\n%s", title, status, m.viewport.View(), help)
}
