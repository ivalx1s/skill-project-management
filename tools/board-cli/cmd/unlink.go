package cmd

import (
	"fmt"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/spf13/cobra"
)

var unlinkCmd = &cobra.Command{
	Use:   "unlink <ID>",
	Short: "Remove a dependency link",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnlink,
}

var unlinkBlockedBy string

func init() {
	rootCmd.AddCommand(unlinkCmd)
	unlinkCmd.Flags().StringVar(&unlinkBlockedBy, "blocked-by", "", "ID of blocker to remove (required)")
	unlinkCmd.MarkFlagRequired("blocked-by")
}

func runUnlink(cmd *cobra.Command, args []string) error {
	id := args[0]

	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	elem := b.FindByID(id)
	if elem == nil {
		return fmt.Errorf("element %s not found", id)
	}

	blocker := b.FindByID(unlinkBlockedBy)
	if blocker == nil {
		return fmt.Errorf("blocker %s not found", unlinkBlockedBy)
	}

	// Remove from element's BlockedBy
	pd, err := board.ParseProgressFile(elem.ProgressPath())
	if err != nil {
		return fmt.Errorf("reading progress for %s: %w", id, err)
	}

	var newBlockedBy []string
	found := false
	for _, bid := range pd.BlockedBy {
		if bid == blocker.ID() {
			found = true
			continue
		}
		newBlockedBy = append(newBlockedBy, bid)
	}
	if !found {
		return fmt.Errorf("%s is not blocked by %s", id, blocker.ID())
	}
	pd.BlockedBy = newBlockedBy
	if err := board.WriteProgressFile(elem.ProgressPath(), pd); err != nil {
		return fmt.Errorf("writing progress for %s: %w", id, err)
	}

	// Remove from blocker's Blocks
	blockerPd, err := board.ParseProgressFile(blocker.ProgressPath())
	if err != nil {
		return fmt.Errorf("reading progress for %s: %w", blocker.ID(), err)
	}

	var newBlocks []string
	for _, bid := range blockerPd.Blocks {
		if bid == elem.ID() {
			continue
		}
		newBlocks = append(newBlocks, bid)
	}
	blockerPd.Blocks = newBlocks
	if err := board.WriteProgressFile(blocker.ProgressPath(), blockerPd); err != nil {
		return fmt.Errorf("writing progress for %s: %w", blocker.ID(), err)
	}

	fmt.Printf("%s: removed blocked-by %s\n", id, blocker.ID())
	fmt.Printf("%s: removed blocks %s\n", blocker.ID(), id)

	// --- De-escalate dependencies up the hierarchy ---
	if err := deescalateDependency(b, elem, blocker); err != nil {
		return fmt.Errorf("de-escalating dependency: %w", err)
	}

	return nil
}

// deescalateDependency checks if removing this link means the parent-level
// dependency should also be removed (no more cross-child links exist).
func deescalateDependency(b *board.Board, elem, blocker *board.Element) error {
	elemParent := b.ParentOf(elem)
	blockerParent := b.ParentOf(blocker)

	if elemParent == nil || blockerParent == nil {
		return nil
	}
	if elemParent.ID() == blockerParent.ID() {
		return nil
	}

	// Re-load board to get fresh state after unlink
	freshBoard, err := board.Load(b.Dir)
	if err != nil {
		return fmt.Errorf("reloading board: %w", err)
	}

	freshElemParent := freshBoard.FindByID(elemParent.ID())
	freshBlockerParent := freshBoard.FindByID(blockerParent.ID())
	if freshElemParent == nil || freshBlockerParent == nil {
		return nil
	}

	// Check if any cross-child dependency still exists
	if freshBoard.HasCrossChildDependency(freshElemParent, freshBlockerParent) {
		return nil // still has cross-links, keep escalated dependency
	}

	// Remove escalated blocked-by from elemParent
	pd, err := board.ParseProgressFile(freshElemParent.ProgressPath())
	if err != nil {
		return fmt.Errorf("reading progress for %s: %w", freshElemParent.ID(), err)
	}
	var newBlockedBy []string
	for _, bid := range pd.BlockedBy {
		if bid != freshBlockerParent.ID() {
			newBlockedBy = append(newBlockedBy, bid)
		}
	}
	pd.BlockedBy = newBlockedBy
	if err := board.WriteProgressFile(freshElemParent.ProgressPath(), pd); err != nil {
		return fmt.Errorf("writing progress for %s: %w", freshElemParent.ID(), err)
	}

	// Remove escalated blocks from blockerParent
	blockerParentPd, err := board.ParseProgressFile(freshBlockerParent.ProgressPath())
	if err != nil {
		return fmt.Errorf("reading progress for %s: %w", freshBlockerParent.ID(), err)
	}
	var newBlocks []string
	for _, bid := range blockerParentPd.Blocks {
		if bid != freshElemParent.ID() {
			newBlocks = append(newBlocks, bid)
		}
	}
	blockerParentPd.Blocks = newBlocks
	if err := board.WriteProgressFile(freshBlockerParent.ProgressPath(), blockerParentPd); err != nil {
		return fmt.Errorf("writing progress for %s: %w", freshBlockerParent.ID(), err)
	}

	fmt.Printf("  ↳ de-escalated: %s no longer blocked by %s\n", freshElemParent.ID(), freshBlockerParent.ID())

	// Recurse up
	return deescalateDependency(freshBoard, freshElemParent, freshBlockerParent)
}
