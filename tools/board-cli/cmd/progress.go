package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/spf13/cobra"
)

var progressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Manage element progress",
}

var progressStatusCmd = &cobra.Command{
	Use:   "status <ID> <status>",
	Short: "Set element status",
	Long: `Set element status.

Valid statuses: backlog, analysis, to-dev, development, to-review, reviewing, done, closed, blocked

Aliases: dev (development), review (reviewing), todev (to-dev), toreview (to-review)

Auto-promotion: When all children of a story/epic become done/closed,
the parent is automatically promoted to done. Cascades up the hierarchy.

Auto-reopen: When a child becomes active and the parent is done/closed,
the parent is automatically reopened to development.

Dependency blocking: Cannot start development if blocked by unfinished tasks.`,
	Args: cobra.ExactArgs(2),
	RunE: runProgressStatus,
}

var progressChecklistCmd = &cobra.Command{
	Use:   "checklist <ID>",
	Short: "Show element checklist",
	Args:  cobra.ExactArgs(1),
	RunE:  runProgressChecklist,
}

var progressCheckCmd = &cobra.Command{
	Use:   "check <ID> <item-number>",
	Short: "Check a checklist item",
	Args:  cobra.ExactArgs(2),
	RunE:  runProgressCheck,
}

var progressUncheckCmd = &cobra.Command{
	Use:   "uncheck <ID> <item-number>",
	Short: "Uncheck a checklist item",
	Args:  cobra.ExactArgs(2),
	RunE:  runProgressUncheck,
}

var progressAddItemCmd = &cobra.Command{
	Use:   "add-item <ID> <text>",
	Short: "Add a checklist item",
	Args:  cobra.ExactArgs(2),
	RunE:  runProgressAddItem,
}

var progressNotesCmd = &cobra.Command{
	Use:   "notes <ID> <text>",
	Short: "Add or set notes on an element",
	Args:  cobra.ExactArgs(2),
	RunE:  runProgressNotes,
}

var progressNotesSet bool

func init() {
	rootCmd.AddCommand(progressCmd)
	progressCmd.AddCommand(progressStatusCmd)
	progressCmd.AddCommand(progressChecklistCmd)
	progressCmd.AddCommand(progressCheckCmd)
	progressCmd.AddCommand(progressUncheckCmd)
	progressCmd.AddCommand(progressAddItemCmd)
	progressCmd.AddCommand(progressNotesCmd)

	progressNotesCmd.Flags().BoolVar(&progressNotesSet, "set", false, "Replace all notes (default: append)")
}

func runProgressStatus(cmd *cobra.Command, args []string) error {
	id := args[0]
	newStatus, err := board.ParseStatus(args[1])
	if err != nil {
		return err
	}

	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	elem := b.FindByID(id)
	if elem == nil {
		return fmt.Errorf("element %s not found", id)
	}

	pd, err := board.ParseProgressFile(elem.ProgressPath())
	if err != nil {
		return fmt.Errorf("reading progress: %w", err)
	}

	// Check blockedBy before allowing development or later
	if newStatus == board.StatusDevelopment || newStatus == board.StatusToReview ||
		newStatus == board.StatusReviewing || newStatus == board.StatusDone {
		activeBlockers := b.ActiveBlockers(elem)
		if len(activeBlockers) > 0 {
			var blockerDescs []string
			for _, blocker := range activeBlockers {
				blockerDescs = append(blockerDescs, fmt.Sprintf("%s (status: %s)", blocker.ID(), blocker.Status))
			}
			return fmt.Errorf("cannot set %s to %s — blocked by:\n  %s",
				id, newStatus, strings.Join(blockerDescs, "\n  "))
		}
	}

	pd.Status = newStatus
	if err := board.WriteProgressFile(elem.ProgressPath(), pd); err != nil {
		return fmt.Errorf("writing progress: %w", err)
	}

	// Update in-memory status for auto-promotion check
	elem.Status = newStatus

	fmt.Printf("%s → %s\n", id, newStatus)

	// Auto-promote parent if all children are done
	if newStatus == board.StatusDone {
		promoteParentIfAllChildrenDone(b, elem)
	}

	// Auto-reopen parent if child becomes active (not done/closed/blocked)
	if newStatus != board.StatusDone && newStatus != board.StatusClosed && newStatus != board.StatusBlocked {
		reopenParentIfNeeded(b, elem)
	}

	return nil
}

// promoteParentIfAllChildrenDone checks if all children of the parent are done/closed,
// and if so, automatically promotes the parent to done. Recurses up the hierarchy.
func promoteParentIfAllChildrenDone(b *board.Board, elem *board.Element) {
	parent := b.ParentOf(elem)
	if parent == nil {
		return // top-level element, nothing to promote
	}

	// Check if all children of parent are done or closed
	children := b.Children(parent)
	for _, child := range children {
		if child.Status != board.StatusDone && child.Status != board.StatusClosed {
			return // not all children done, don't promote
		}
	}

	// All children done → promote parent to done
	parentPd, err := board.ParseProgressFile(parent.ProgressPath())
	if err != nil {
		return // silently skip on error
	}

	if parentPd.Status == board.StatusDone || parentPd.Status == board.StatusClosed {
		return // already done/closed
	}

	parentPd.Status = board.StatusDone
	if err := board.WriteProgressFile(parent.ProgressPath(), parentPd); err != nil {
		return // silently skip on error
	}

	// Update in-memory status for recursive check
	parent.Status = board.StatusDone

	fmt.Printf("%s → %s (auto-promoted: all children done)\n", parent.ID(), board.StatusDone)

	// Recursively check grandparent
	promoteParentIfAllChildrenDone(b, parent)
}

// reopenParentIfNeeded checks if parent is done/closed and reopens it
// when a child is set to open/progress. Recurses up the hierarchy.
func reopenParentIfNeeded(b *board.Board, elem *board.Element) {
	parent := b.ParentOf(elem)
	if parent == nil {
		return // top-level element
	}

	// Only reopen if parent is currently done or closed
	if parent.Status != board.StatusDone && parent.Status != board.StatusClosed {
		return // parent is already open/progress/blocked
	}

	// Reopen parent to development (since it has work in progress)
	parentPd, err := board.ParseProgressFile(parent.ProgressPath())
	if err != nil {
		return
	}

	parentPd.Status = board.StatusDevelopment
	if err := board.WriteProgressFile(parent.ProgressPath(), parentPd); err != nil {
		return
	}

	parent.Status = board.StatusDevelopment
	fmt.Printf("%s → %s (auto-reopened: child active)\n", parent.ID(), board.StatusDevelopment)

	// Recursively check grandparent
	reopenParentIfNeeded(b, parent)
}

func runProgressChecklist(cmd *cobra.Command, args []string) error {
	id := args[0]

	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	elem := b.FindByID(id)
	if elem == nil {
		return fmt.Errorf("element %s not found", id)
	}

	pd, err := board.ParseProgressFile(elem.ProgressPath())
	if err != nil {
		return fmt.Errorf("reading progress: %w", err)
	}

	if len(pd.Checklist) == 0 {
		fmt.Printf("%s: checklist is empty\n", id)
		return nil
	}

	fmt.Printf("%s checklist:\n", id)
	for i, item := range pd.Checklist {
		mark := "[ ]"
		if item.Checked {
			mark = "[x]"
		}
		fmt.Printf("  %d. %s %s\n", i+1, mark, item.Text)
	}
	return nil
}

func runProgressCheck(cmd *cobra.Command, args []string) error {
	return toggleChecklistItem(args[0], args[1], true)
}

func runProgressUncheck(cmd *cobra.Command, args []string) error {
	return toggleChecklistItem(args[0], args[1], false)
}

func toggleChecklistItem(id, numStr string, checked bool) error {
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return fmt.Errorf("invalid item number: %s", numStr)
	}

	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	elem := b.FindByID(id)
	if elem == nil {
		return fmt.Errorf("element %s not found", id)
	}

	pd, err := board.ParseProgressFile(elem.ProgressPath())
	if err != nil {
		return fmt.Errorf("reading progress: %w", err)
	}

	if num < 1 || num > len(pd.Checklist) {
		return fmt.Errorf("item %d out of range (1-%d)", num, len(pd.Checklist))
	}

	pd.Checklist[num-1].Checked = checked
	if err := board.WriteProgressFile(elem.ProgressPath(), pd); err != nil {
		return fmt.Errorf("writing progress: %w", err)
	}

	action := "checked"
	if !checked {
		action = "unchecked"
	}
	fmt.Printf("%s item %d %s: %s\n", id, num, action, pd.Checklist[num-1].Text)
	return nil
}

func runProgressNotes(cmd *cobra.Command, args []string) error {
	id := args[0]
	text := args[1]

	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	elem := b.FindByID(id)
	if elem == nil {
		return fmt.Errorf("element %s not found", id)
	}

	pd, err := board.ParseProgressFile(elem.ProgressPath())
	if err != nil {
		return fmt.Errorf("reading progress: %w", err)
	}

	if progressNotesSet {
		pd.Notes = text
	} else {
		if pd.Notes != "" {
			pd.Notes += "\n" + text
		} else {
			pd.Notes = text
		}
	}

	if err := board.WriteProgressFile(elem.ProgressPath(), pd); err != nil {
		return fmt.Errorf("writing progress: %w", err)
	}

	action := "appended"
	if progressNotesSet {
		action = "set"
	}
	fmt.Printf("%s: notes %s\n", id, action)
	return nil
}

func runProgressAddItem(cmd *cobra.Command, args []string) error {
	id := args[0]
	text := args[1]

	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	elem := b.FindByID(id)
	if elem == nil {
		return fmt.Errorf("element %s not found", id)
	}

	pd, err := board.ParseProgressFile(elem.ProgressPath())
	if err != nil {
		return fmt.Errorf("reading progress: %w", err)
	}

	pd.Checklist = append(pd.Checklist, board.ChecklistItem{Text: text, Checked: false})
	if err := board.WriteProgressFile(elem.ProgressPath(), pd); err != nil {
		return fmt.Errorf("writing progress: %w", err)
	}

	fmt.Printf("%s: added item %d — %s\n", id, len(pd.Checklist), text)
	return nil
}
