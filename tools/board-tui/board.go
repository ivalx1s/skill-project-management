package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// boardRow represents a single row in the board view.
// Rows with a non-nil node are selectable (tree nodes);
// rows with node == nil are decoration (dependency info).
type boardRow struct {
	node *TreeNode // non-nil for selectable rows
	text string    // pre-rendered display text
}

func (r boardRow) selectable() bool {
	return r.node != nil
}

// --- Row building ---

// buildBoardRows flattens the tree into renderable rows
func buildBoardRows(tree []*TreeNode) []boardRow {
	flatNodes := FlattenTree(tree)
	rows := make([]boardRow, 0, len(flatNodes)*2)

	for idx, fn := range flatNodes {
		node := fn.Node

		// Visual spacing: blank line before epics and stories
		if idx > 0 {
			if node.Type == "epic" || node.Type == "story" {
				rows = append(rows, boardRow{text: ""})
			}
		}

		// Expand indicator
		var expandInd string
		if node.HasChildren() {
			if node.Expanded {
				expandInd = "▼"
			} else {
				expandInd = "▶"
			}
		} else {
			expandInd = " "
		}

		// Type indicator with color
		typeInd := typeIndicators[node.Type]
		if typeInd == "" {
			typeInd = "?"
		}
		typeStyle, ok := typeStyles[node.Type]
		if !ok {
			typeStyle = lipgloss.NewStyle()
		}

		// Status
		statusStyle, ok := statusStyles[node.Status]
		if !ok {
			statusStyle = lipgloss.NewStyle()
		}

		assignee := ""
		if a := node.GetAssignee(); a != "" {
			assignee = fmt.Sprintf(" @%s", a)
		}

		text := fmt.Sprintf("%s%s %s %s %s %s%s",
			fn.TreePrefix, expandInd, typeStyle.Render(typeInd), node.ID,
			node.Name, statusStyle.Render("["+node.Status+"]"), assignee)

		rows = append(rows, boardRow{node: node, text: text})

		// Dependency description line (non-selectable)
		if desc := buildDescLine(node, fn.TreePrefix); desc != "" {
			rows = append(rows, boardRow{text: desc})
		}

		// Inter-row spacing after each node (except last)
		if idx < len(flatNodes)-1 {
			rows = append(rows, boardRow{text: ""})
		}
	}

	return rows
}

func buildDescLine(node *TreeNode, treePrefix string) string {
	var parts []string
	if len(node.BlockedBy) > 0 {
		s := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4500"))
		parts = append(parts, s.Render("← blocked by: ")+strings.Join(node.BlockedBy, ", "))
	}
	if len(node.Blocks) > 0 {
		s := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
		parts = append(parts, s.Render("→ blocks: ")+strings.Join(node.Blocks, ", "))
	}
	if len(parts) == 0 {
		return ""
	}
	indent := strings.Repeat(" ", len(treePrefix)+2)
	return indent + strings.Join(parts, "  ")
}

// --- Board row management ---

// boardRebuildRows rebuilds the flat row list from the tree, preserving selection
func (m *model) boardRebuildRows() {
	var selectedID string
	if m.boardSelectedIdx >= 0 && m.boardSelectedIdx < len(m.boardRows) {
		if node := m.boardRows[m.boardSelectedIdx].node; node != nil {
			selectedID = node.ID
		}
	}

	m.boardRows = buildBoardRows(m.tree)

	// Restore selection by ID
	if selectedID != "" {
		for i, row := range m.boardRows {
			if row.node != nil && row.node.ID == selectedID {
				m.boardSelectedIdx = i
				m.boardEnsureVisible()
				return
			}
		}
	}

	// Clamp index
	if m.boardSelectedIdx >= len(m.boardRows) {
		m.boardSelectedIdx = len(m.boardRows) - 1
	}
	if m.boardSelectedIdx < 0 {
		m.boardSelectedIdx = 0
	}
	if len(m.boardRows) > 0 && !m.boardRows[m.boardSelectedIdx].selectable() {
		m.boardMoveDown()
	}
}

// --- Navigation ---

func (m *model) boardMoveDown() {
	for i := m.boardSelectedIdx + 1; i < len(m.boardRows); i++ {
		if m.boardRows[i].selectable() {
			m.boardSelectedIdx = i
			m.boardEnsureVisible()
			return
		}
	}
}

func (m *model) boardMoveUp() {
	for i := m.boardSelectedIdx - 1; i >= 0; i-- {
		if m.boardRows[i].selectable() {
			m.boardSelectedIdx = i
			m.boardEnsureVisible()
			return
		}
	}
}

func (m *model) boardGoTop() {
	for i := 0; i < len(m.boardRows); i++ {
		if m.boardRows[i].selectable() {
			m.boardSelectedIdx = i
			m.boardScrollOff = 0
			return
		}
	}
}

func (m *model) boardGoBottom() {
	for i := len(m.boardRows) - 1; i >= 0; i-- {
		if m.boardRows[i].selectable() {
			m.boardSelectedIdx = i
			m.boardEnsureVisible()
			return
		}
	}
}

func (m *model) boardEnsureVisible() {
	vh := m.boardVisibleHeight()
	if m.boardSelectedIdx < m.boardScrollOff {
		m.boardScrollOff = m.boardSelectedIdx
	}
	if m.boardSelectedIdx >= m.boardScrollOff+vh {
		m.boardScrollOff = m.boardSelectedIdx - vh + 1
	}
}

func (m *model) boardVisibleHeight() int {
	// Layout: appPad(1) + title(1) + blank(1) + [rows] + help(1) + appPad(1) = 5
	h := m.height - 5
	if h < 1 {
		h = 1
	}
	return h
}

func (m *model) boardSelectedNode() *TreeNode {
	if m.boardSelectedIdx >= 0 && m.boardSelectedIdx < len(m.boardRows) {
		return m.boardRows[m.boardSelectedIdx].node
	}
	return nil
}

func (m *model) boardSelectNodeByID(id string) {
	for i, row := range m.boardRows {
		if row.node != nil && row.node.ID == id {
			m.boardSelectedIdx = i
			m.boardEnsureVisible()
			return
		}
	}
}
