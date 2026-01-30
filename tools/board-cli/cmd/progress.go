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
	Short: "Set element status (open|progress|done|closed|blocked)",
	Args:  cobra.ExactArgs(2),
	RunE:  runProgressStatus,
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

	// Check blockedBy before allowing progress or done
	if (newStatus == board.StatusProgress || newStatus == board.StatusDone) && len(pd.BlockedBy) > 0 {
		// Filter out resolved blockers
		var activeBlockers []string
		for _, blockerID := range pd.BlockedBy {
			blocker := b.FindByID(blockerID)
			if blocker == nil {
				activeBlockers = append(activeBlockers, blockerID+" (not found)")
				continue
			}
			if blocker.Status != board.StatusDone && blocker.Status != board.StatusClosed {
				activeBlockers = append(activeBlockers, fmt.Sprintf("%s (status: %s)", blockerID, blocker.Status))
			}
		}
		if len(activeBlockers) > 0 {
			return fmt.Errorf("cannot set %s to %s — blocked by:\n  %s",
				id, newStatus, strings.Join(activeBlockers, "\n  "))
		}
	}

	pd.Status = newStatus
	if err := board.WriteProgressFile(elem.ProgressPath(), pd); err != nil {
		return fmt.Errorf("writing progress: %w", err)
	}

	fmt.Printf("%s → %s\n", id, newStatus)
	return nil
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
