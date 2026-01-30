package cmd

import (
	"fmt"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link <ID>",
	Short: "Link element dependencies",
	Args:  cobra.ExactArgs(1),
	RunE:  runLink,
}

var linkBlockedBy string

func init() {
	rootCmd.AddCommand(linkCmd)
	linkCmd.Flags().StringVar(&linkBlockedBy, "blocked-by", "", "ID of blocking element (required)")
	linkCmd.MarkFlagRequired("blocked-by")
}

func runLink(cmd *cobra.Command, args []string) error {
	id := args[0]

	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	elem := b.FindByID(id)
	if elem == nil {
		return fmt.Errorf("element %s not found", id)
	}

	blocker := b.FindByID(linkBlockedBy)
	if blocker == nil {
		return fmt.Errorf("blocker %s not found", linkBlockedBy)
	}

	// --- Update blocked element: add blockedBy ---
	pd, err := board.ParseProgressFile(elem.ProgressPath())
	if err != nil {
		return fmt.Errorf("reading progress for %s: %w", id, err)
	}

	// Check for duplicates
	for _, existing := range pd.BlockedBy {
		if existing == blocker.ID() {
			fmt.Printf("%s is already blocked by %s\n", id, blocker.ID())
			return nil
		}
	}

	// Remove "(none)" placeholder if present
	var cleanedBy []string
	for _, item := range pd.BlockedBy {
		if item != "(none)" {
			cleanedBy = append(cleanedBy, item)
		}
	}
	pd.BlockedBy = append(cleanedBy, blocker.ID())

	if err := board.WriteProgressFile(elem.ProgressPath(), pd); err != nil {
		return fmt.Errorf("writing progress for %s: %w", id, err)
	}

	// --- Update blocker element: add blocks ---
	blockerPd, err := board.ParseProgressFile(blocker.ProgressPath())
	if err != nil {
		return fmt.Errorf("reading progress for %s: %w", blocker.ID(), err)
	}

	// Check for duplicates in blocks
	alreadyBlocks := false
	for _, existing := range blockerPd.Blocks {
		if existing == elem.ID() {
			alreadyBlocks = true
			break
		}
	}

	if !alreadyBlocks {
		var cleanedBlocks []string
		for _, item := range blockerPd.Blocks {
			if item != "(none)" {
				cleanedBlocks = append(cleanedBlocks, item)
			}
		}
		blockerPd.Blocks = append(cleanedBlocks, elem.ID())

		if err := board.WriteProgressFile(blocker.ProgressPath(), blockerPd); err != nil {
			return fmt.Errorf("writing progress for %s: %w", blocker.ID(), err)
		}
	}

	fmt.Printf("%s → blocked by %s\n", id, blocker.ID())
	fmt.Printf("%s → blocks %s\n", blocker.ID(), id)

	// --- Escalate dependencies up the hierarchy ---
	if err := escalateDependency(b, elem, blocker); err != nil {
		return fmt.Errorf("escalating dependency: %w", err)
	}

	return nil
}

// escalateDependency checks if elem and blocker have different parents,
// and if so, creates an implied dependency between those parents.
// Recurses up the hierarchy (story→epic→project).
func escalateDependency(b *board.Board, elem, blocker *board.Element) error {
	elemParent := b.ParentOf(elem)
	blockerParent := b.ParentOf(blocker)

	// Both must have parents and they must differ
	if elemParent == nil || blockerParent == nil {
		return nil
	}
	if elemParent.ID() == blockerParent.ID() {
		return nil
	}

	// Check if dependency already exists
	pd, err := board.ParseProgressFile(elemParent.ProgressPath())
	if err != nil {
		return fmt.Errorf("reading progress for %s: %w", elemParent.ID(), err)
	}
	for _, existing := range pd.BlockedBy {
		if existing == blockerParent.ID() {
			// Already linked, but still recurse up
			return escalateDependency(b, elemParent, blockerParent)
		}
	}

	// Add blocked-by to parent
	var cleanedBy []string
	for _, item := range pd.BlockedBy {
		if item != "(none)" {
			cleanedBy = append(cleanedBy, item)
		}
	}
	pd.BlockedBy = append(cleanedBy, blockerParent.ID())
	if err := board.WriteProgressFile(elemParent.ProgressPath(), pd); err != nil {
		return fmt.Errorf("writing progress for %s: %w", elemParent.ID(), err)
	}

	// Add blocks to blocker's parent
	blockerParentPd, err := board.ParseProgressFile(blockerParent.ProgressPath())
	if err != nil {
		return fmt.Errorf("reading progress for %s: %w", blockerParent.ID(), err)
	}
	alreadyBlocks := false
	for _, existing := range blockerParentPd.Blocks {
		if existing == elemParent.ID() {
			alreadyBlocks = true
			break
		}
	}
	if !alreadyBlocks {
		var cleanedBlocks []string
		for _, item := range blockerParentPd.Blocks {
			if item != "(none)" {
				cleanedBlocks = append(cleanedBlocks, item)
			}
		}
		blockerParentPd.Blocks = append(cleanedBlocks, elemParent.ID())
		if err := board.WriteProgressFile(blockerParent.ProgressPath(), blockerParentPd); err != nil {
			return fmt.Errorf("writing progress for %s: %w", blockerParent.ID(), err)
		}
	}

	fmt.Printf("  ↳ escalated: %s → blocked by %s\n", elemParent.ID(), blockerParent.ID())

	// Recurse up
	return escalateDependency(b, elemParent, blockerParent)
}
