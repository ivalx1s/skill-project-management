package board

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"board-tui/internal/styles"
)

// TreeNode represents a node in the task board tree
type TreeNode struct {
	ID        string
	Type      string
	Name      string
	Status    string
	Assignee  *string
	UpdatedAt string
	Children  []*TreeNode
	Expanded  bool
	Depth     int
	Parent    *TreeNode
	BlockedBy []string
	Blocks    []string
}

// HasChildren returns true if the node has children
func (n *TreeNode) HasChildren() bool {
	return len(n.Children) > 0
}

// Toggle flips the expanded state
func (n *TreeNode) Toggle() {
	n.Expanded = !n.Expanded
}

// Expand expands the node
func (n *TreeNode) Expand() {
	n.Expanded = true
}

// Collapse collapses the node
func (n *TreeNode) Collapse() {
	n.Expanded = false
}

// ExpandAll recursively expands this node and all descendants
func (n *TreeNode) ExpandAll() {
	n.Expanded = true
	for _, child := range n.Children {
		child.ExpandAll()
	}
}

// CollapseAll recursively collapses this node and all descendants
func (n *TreeNode) CollapseAll() {
	n.Expanded = false
	for _, child := range n.Children {
		child.CollapseAll()
	}
}

// GetAssignee returns the assignee or empty string if nil
func (n *TreeNode) GetAssignee() string {
	if n.Assignee == nil {
		return ""
	}
	return *n.Assignee
}

// Item represents a task board element for the list view
type Item struct {
	Node       *TreeNode
	TreePrefix string
	Depth      int
}

func (i Item) Title() string {
	var expandIndicator string
	if i.Node.HasChildren() {
		if i.Node.Expanded {
			expandIndicator = "▼ "
		} else {
			expandIndicator = "▶ "
		}
	} else {
		expandIndicator = "  "
	}

	typeInd := styles.TypeIndicator[i.Node.Type]
	if typeInd == "" {
		typeInd = "?"
	}

	typeStyle := styles.TypeStyle[i.Node.Type]
	statusStyle := styles.Status[i.Node.Status]

	id := typeStyle.Render(i.Node.ID)
	name := i.Node.Name
	status := statusStyle.Render("[" + i.Node.Status + "]")

	assignee := ""
	if i.Node.GetAssignee() != "" {
		assigneeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#29B6F6"))
		assignee = " " + assigneeStyle.Render("@"+i.Node.GetAssignee())
	}

	return fmt.Sprintf("%s%s%s %s %s %s%s",
		i.TreePrefix,
		expandIndicator,
		typeStyle.Render(typeInd),
		id,
		name,
		status,
		assignee)
}

func (i Item) Description() string {
	var parts []string

	if len(i.Node.BlockedBy) > 0 {
		parts = append(parts, styles.BlockedBy.Render("← blocked by: ")+strings.Join(i.Node.BlockedBy, ", "))
	}

	if len(i.Node.Blocks) > 0 {
		parts = append(parts, styles.Blocks.Render("→ blocks: ")+strings.Join(i.Node.Blocks, ", "))
	}

	if len(parts) == 0 {
		return ""
	}

	indent := strings.Repeat(" ", len(i.TreePrefix)+2)
	return indent + strings.Join(parts, "  ")
}

func (i Item) FilterValue() string {
	return i.Node.ID + " " + i.Node.Name + " " + i.Node.Status
}

// FlatNode represents a flattened tree node for display
type FlatNode struct {
	Node       *TreeNode
	Depth      int
	IsLast     bool
	TreePrefix string
}

// FlattenTree flattens the tree into a slice respecting expanded state
func FlattenTree(roots []*TreeNode) []FlatNode {
	var result []FlatNode
	for i, root := range roots {
		isLast := i == len(roots)-1
		flattenNode(root, 0, isLast, "", &result)
	}
	return result
}

func flattenNode(node *TreeNode, depth int, isLast bool, parentPrefix string, result *[]FlatNode) {
	var prefix string
	if depth == 0 {
		prefix = ""
	} else {
		if isLast {
			prefix = parentPrefix + "└── "
		} else {
			prefix = parentPrefix + "├── "
		}
	}

	flat := FlatNode{
		Node:       node,
		Depth:      depth,
		IsLast:     isLast,
		TreePrefix: prefix,
	}
	*result = append(*result, flat)

	if node.Expanded && len(node.Children) > 0 {
		var childPrefix string
		if depth == 0 {
			childPrefix = "    "
		} else {
			if isLast {
				childPrefix = parentPrefix + "    "
			} else {
				childPrefix = parentPrefix + "│   "
			}
		}

		for i, child := range node.Children {
			childIsLast := i == len(node.Children)-1
			flattenNode(child, depth+1, childIsLast, childPrefix, result)
		}
	}
}

// Model is the board screen model
type Model struct {
	List            list.Model
	Tree            []*TreeNode
	Width           int
	Height          int
	PreFilterExpIDs []string
	WasFiltering    bool
}

// New creates a new board model
func New() Model {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Task Board"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.InfiniteScrolling = false
	l.Styles.Title = styles.Title

	return Model{
		List: l,
	}
}

// SetSize sets the terminal size
func (m *Model) SetSize(width, height int) {
	m.Width = width
	m.Height = height
	m.List.SetSize(width-4, height-6)
}

// SetTree sets the tree data and refreshes the list
func (m *Model) SetTree(tree []*TreeNode) {
	m.Tree = tree
	m.RefreshList()
}

// RefreshList updates the list items from the tree
func (m *Model) RefreshList() {
	if len(m.Tree) == 0 {
		return
	}

	selectedID := ""
	if item, ok := m.List.SelectedItem().(Item); ok {
		selectedID = item.Node.ID
	}

	flatNodes := FlattenTree(m.Tree)
	items := make([]list.Item, len(flatNodes))
	selectedIdx := 0

	for i, fn := range flatNodes {
		items[i] = Item{
			Node:       fn.Node,
			TreePrefix: fn.TreePrefix,
			Depth:      fn.Depth,
		}
		if fn.Node.ID == selectedID {
			selectedIdx = i
		}
	}

	m.List.SetItems(items)
	m.List.Select(selectedIdx)
}

// SelectedNode returns the currently selected tree node
func (m *Model) SelectedNode() *TreeNode {
	if item, ok := m.List.SelectedItem().(Item); ok {
		return item.Node
	}
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

// View renders the board
func (m Model) View() string {
	return m.List.View()
}

// CollectExpandedNodes returns IDs of all expanded nodes
func CollectExpandedNodes(roots []*TreeNode) []string {
	var ids []string
	for _, root := range roots {
		collectExpanded(root, &ids)
	}
	return ids
}

func collectExpanded(node *TreeNode, ids *[]string) {
	if node.Expanded {
		*ids = append(*ids, node.ID)
	}
	for _, child := range node.Children {
		collectExpanded(child, ids)
	}
}

// ApplyExpandedNodes expands nodes with the given IDs
func ApplyExpandedNodes(roots []*TreeNode, ids []string) {
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}
	for _, root := range roots {
		applyExpanded(root, idSet)
	}
}

func applyExpanded(node *TreeNode, idSet map[string]bool) {
	node.Expanded = idSet[node.ID]
	for _, child := range node.Children {
		applyExpanded(child, idSet)
	}
}
