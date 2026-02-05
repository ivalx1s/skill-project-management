package cmd

import (
	"fmt"
	"os"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

var treeEpic string

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Show board as hierarchical tree",
	Long:  "Display the full board structure as a hierarchical tree. Use --json for machine-readable output.",
	Args:  cobra.NoArgs,
	RunE:  runTree,
}

func init() {
	rootCmd.AddCommand(treeCmd)
	treeCmd.Flags().StringVar(&treeEpic, "epic", "", "Filter by epic ID (show only this epic and its children)")
}

// TreeNode represents a node in the hierarchical tree output
type TreeNode struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Name      string      `json:"name"`
	Status    string      `json:"status"`
	Assignee  *string     `json:"assignee"` // nil if not assigned
	UpdatedAt string      `json:"updatedAt"`
	Children  []*TreeNode `json:"children"`
}

// TreeResponse is the JSON response for the tree command
type TreeResponse struct {
	Tree []*TreeNode `json:"tree"`
}

func runTree(cmd *cobra.Command, args []string) error {
	b, err := board.Load(boardDir)
	if err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("loading board: %s", err.Error()), nil)
			return nil
		}
		return fmt.Errorf("loading board: %w", err)
	}

	// Get epics (filtered if --epic flag is set)
	epics := b.FindByType(board.EpicType)
	if treeEpic != "" {
		epic := b.FindByID(treeEpic)
		if epic == nil {
			if JSONEnabled() {
				output.PrintError(os.Stderr, output.NotFound, fmt.Sprintf("epic %s not found", treeEpic), nil)
				return nil
			}
			return fmt.Errorf("epic %s not found", treeEpic)
		}
		if epic.Type != board.EpicType {
			if JSONEnabled() {
				output.PrintError(os.Stderr, output.InvalidID, fmt.Sprintf("%s is not an epic", treeEpic), nil)
				return nil
			}
			return fmt.Errorf("%s is not an epic", treeEpic)
		}
		epics = []*board.Element{epic}
	}

	// Build tree
	tree := buildTree(b, epics)

	if JSONEnabled() {
		response := TreeResponse{Tree: tree}
		return output.PrintJSON(os.Stdout, response)
	}

	// Text output
	if len(tree) == 0 {
		fmt.Println("Board is empty.")
		return nil
	}

	printTreeText(tree, "")
	return nil
}

func buildTree(b *board.Board, epics []*board.Element) []*TreeNode {
	var nodes []*TreeNode
	for _, epic := range epics {
		node := elementToNode(epic)
		// Add stories as children
		stories := b.Children(epic)
		for _, story := range stories {
			storyNode := elementToNode(story)
			// Add tasks/bugs as children of story
			tasks := b.Children(story)
			for _, task := range tasks {
				taskNode := elementToNode(task)
				storyNode.Children = append(storyNode.Children, taskNode)
			}
			node.Children = append(node.Children, storyNode)
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func elementToNode(e *board.Element) *TreeNode {
	node := &TreeNode{
		ID:       e.ID(),
		Type:     string(e.Type),
		Name:     e.Name,
		Status:   string(e.Status),
		Children: []*TreeNode{},
	}

	// Set assignee (nil if empty)
	if e.AssignedTo != "" {
		node.Assignee = &e.AssignedTo
	}

	// Format updatedAt in RFC3339
	if !e.LastUpdate.IsZero() {
		node.UpdatedAt = e.LastUpdate.Format("2006-01-02T15:04:05Z")
	}

	return node
}

func printTreeText(nodes []*TreeNode, prefix string) {
	for i, node := range nodes {
		isLast := i == len(nodes)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		// Status with color
		statusStr := output.ColorStatus(node.Status)

		// Assignee indicator
		assigneeStr := ""
		if node.Assignee != nil {
			assigneeStr = fmt.Sprintf(" %s[@%s]%s", output.Cyan, *node.Assignee, output.Reset)
		}

		fmt.Printf("%s%s%s%s %s [%s]%s\n", prefix, connector, output.Bold, node.ID, output.Reset+node.Name, statusStr, assigneeStr)

		// Recursively print children
		childPrefix := prefix
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
		printTreeText(node.Children, childPrefix)
	}
}
