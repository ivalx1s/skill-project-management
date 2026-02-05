package main

import (
	"encoding/json"
	"os/exec"
	"time"
)

// TreeNode represents a node in the task board tree with expand/collapse state
type TreeNode struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Name      string      `json:"name"`
	Status    string      `json:"status"`
	Assignee  *string     `json:"assignee"`
	UpdatedAt string      `json:"updatedAt"`
	Children  []*TreeNode `json:"children"`
	Expanded  bool        `json:"-"` // Not from JSON, managed locally
	Depth     int         `json:"-"` // Calculated during flattening
	Parent    *TreeNode   `json:"-"` // Back-reference to parent
	BlockedBy []string    `json:"-"` // Dependencies - loaded separately
	Blocks    []string    `json:"-"` // What this blocks - loaded separately
}

// TreeResponse is the JSON response from `task-board tree --json`
type TreeResponse struct {
	Tree []*TreeNode `json:"tree"`
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
	n.Collapsed()
}

// Collapsed collapses the node (named differently to avoid confusion)
func (n *TreeNode) Collapsed() {
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

// GetParsedUpdatedAt parses the UpdatedAt string into time.Time
func (n *TreeNode) GetParsedUpdatedAt() (time.Time, error) {
	if n.UpdatedAt == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, n.UpdatedAt)
}

// GetAssignee returns the assignee or empty string if nil
func (n *TreeNode) GetAssignee() string {
	if n.Assignee == nil {
		return ""
	}
	return *n.Assignee
}

// FlatNode represents a flattened tree node for display purposes
type FlatNode struct {
	Node       *TreeNode
	Depth      int
	IsLast     bool   // Is this the last child of its parent
	TreePrefix string // Visual tree prefix (e.g., "├── " or "└── ")
}

// ListElement represents an element from task-board list --json
type ListElement struct {
	ID        string   `json:"id"`
	BlockedBy []string `json:"blockedBy"`
	Blocks    []string `json:"blocks"`
}

// ListResponse is the JSON response from task-board list --json
type ListResponse struct {
	Elements []ListElement `json:"elements"`
}

// LoadTreeFromCLI calls `task-board tree --json` and parses the response
func LoadTreeFromCLI() ([]*TreeNode, error) {
	cmd := exec.Command("task-board", "tree", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	tree, err := ParseTreeJSON(output)
	if err != nil {
		return nil, err
	}

	// Load dependencies from list and apply to tree
	LoadDependencies(tree)

	return tree, nil
}

// LoadDependencies loads blockedBy/blocks from list command and applies to tree nodes
func LoadDependencies(roots []*TreeNode) {
	// Get all elements with dependencies
	cmd := exec.Command("task-board", "list", "tasks", "--json")
	output, err := cmd.Output()
	if err != nil {
		return // Silently fail - dependencies are optional
	}

	var response ListResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return
	}

	// Build map of dependencies
	depsMap := make(map[string]ListElement)
	for _, elem := range response.Elements {
		depsMap[elem.ID] = elem
	}

	// Also load stories
	cmd = exec.Command("task-board", "list", "stories", "--json")
	output, err = cmd.Output()
	if err == nil {
		var storyResponse ListResponse
		if json.Unmarshal(output, &storyResponse) == nil {
			for _, elem := range storyResponse.Elements {
				depsMap[elem.ID] = elem
			}
		}
	}

	// Apply to tree nodes
	for _, root := range roots {
		applyDependencies(root, depsMap)
	}
}

// applyDependencies recursively applies dependencies to tree nodes
func applyDependencies(node *TreeNode, depsMap map[string]ListElement) {
	if elem, ok := depsMap[node.ID]; ok {
		node.BlockedBy = elem.BlockedBy
		node.Blocks = elem.Blocks
	}
	for _, child := range node.Children {
		applyDependencies(child, depsMap)
	}
}

// LoadTreeFromCLIWithEpic calls `task-board tree --json --epic EPIC-XX` for a specific epic
func LoadTreeFromCLIWithEpic(epicID string) ([]*TreeNode, error) {
	cmd := exec.Command("task-board", "tree", "--json", "--epic", epicID)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return ParseTreeJSON(output)
}

// ParseTreeJSON parses the JSON output from task-board tree --json
func ParseTreeJSON(data []byte) ([]*TreeNode, error) {
	var response TreeResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	// Set parent references and default expanded state
	for _, node := range response.Tree {
		initializeNode(node, nil, 0)
	}

	return response.Tree, nil
}

// initializeNode recursively initializes parent references and depth
func initializeNode(node *TreeNode, parent *TreeNode, depth int) {
	node.Parent = parent
	node.Depth = depth
	// Default: everything collapsed
	node.Expanded = false

	for _, child := range node.Children {
		initializeNode(child, node, depth+1)
	}
}

// FlattenTree flattens the tree into a slice respecting expanded state
// Only expanded nodes' children are included
func FlattenTree(roots []*TreeNode) []FlatNode {
	var result []FlatNode
	for i, root := range roots {
		isLast := i == len(roots)-1
		flattenNode(root, 0, isLast, "", &result)
	}
	return result
}

// flattenNode recursively flattens a node and its visible children
func flattenNode(node *TreeNode, depth int, isLast bool, parentPrefix string, result *[]FlatNode) {
	// Build tree prefix for visual hierarchy
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

	// Only recurse into children if expanded
	if node.Expanded && len(node.Children) > 0 {
		// Calculate the prefix for children
		var childPrefix string
		if depth == 0 {
			childPrefix = "    " // Indent children of root elements
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

// FindNodeByID searches for a node by ID in the tree
func FindNodeByID(roots []*TreeNode, id string) *TreeNode {
	for _, root := range roots {
		if found := findNodeRecursive(root, id); found != nil {
			return found
		}
	}
	return nil
}

// findNodeRecursive recursively searches for a node by ID
func findNodeRecursive(node *TreeNode, id string) *TreeNode {
	if node.ID == id {
		return node
	}
	for _, child := range node.Children {
		if found := findNodeRecursive(child, id); found != nil {
			return found
		}
	}
	return nil
}

// GetVisibleCount returns the number of visible nodes (respecting expanded state)
func GetVisibleCount(roots []*TreeNode) int {
	count := 0
	for _, root := range roots {
		count += countVisible(root)
	}
	return count
}

// countVisible recursively counts visible nodes
func countVisible(node *TreeNode) int {
	count := 1 // This node
	if node.Expanded {
		for _, child := range node.Children {
			count += countVisible(child)
		}
	}
	return count
}

// ExpandToNode expands all ancestors of the given node so it becomes visible
func ExpandToNode(node *TreeNode) {
	parent := node.Parent
	for parent != nil {
		parent.Expanded = true
		parent = parent.Parent
	}
}

// GetDemoTree returns demo tree data for testing/fallback
func GetDemoTree() []*TreeNode {
	task1 := &TreeNode{
		ID:       "TASK-001",
		Type:     "task",
		Name:     "setup-go-module",
		Status:   "done",
		Children: []*TreeNode{},
		Expanded: false,
	}
	task2 := &TreeNode{
		ID:       "TASK-002",
		Type:     "task",
		Name:     "implement-board-loader",
		Status:   "development",
		Children: []*TreeNode{},
		Expanded: false,
	}
	task3 := &TreeNode{
		ID:       "TASK-003",
		Type:     "task",
		Name:     "build-tree-navigation",
		Status:   "backlog",
		Children: []*TreeNode{},
		Expanded: false,
	}

	story1 := &TreeNode{
		ID:       "STORY-001",
		Type:     "story",
		Name:     "bubbletea-viewer-prototype",
		Status:   "development",
		Children: []*TreeNode{task1, task2, task3},
		Expanded: false,
	}

	epic1 := &TreeNode{
		ID:       "EPIC-001",
		Type:     "epic",
		Name:     "interactive-tui-for-task-board",
		Status:   "backlog",
		Children: []*TreeNode{story1},
		Expanded: true, // Epics start expanded
	}

	// Set parent references
	task1.Parent = story1
	task2.Parent = story1
	task3.Parent = story1
	story1.Parent = epic1

	return []*TreeNode{epic1}
}
