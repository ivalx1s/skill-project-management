package cmd

import (
	"fmt"
	"os"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

// UnlinkResponse represents the JSON response for unlink command
type UnlinkResponse struct {
	Updated UnlinkUpdate `json:"updated"`
	Message string       `json:"message"`
}

// UnlinkUpdate contains the unlink details
type UnlinkUpdate struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Relation string `json:"relation"`
}

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
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("loading board: %v", err), nil)
			return nil
		}
		return fmt.Errorf("loading board: %w", err)
	}

	elem := b.FindByID(id)
	if elem == nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.NotFound, fmt.Sprintf("element %s not found", id), map[string]interface{}{
				"id": id,
			})
			return nil
		}
		return fmt.Errorf("element %s not found", id)
	}

	blocker := b.FindByID(unlinkBlockedBy)
	if blocker == nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.NotFound, fmt.Sprintf("blocker %s not found", unlinkBlockedBy), map[string]interface{}{
				"id": unlinkBlockedBy,
			})
			return nil
		}
		return fmt.Errorf("blocker %s not found", unlinkBlockedBy)
	}

	// Remove from element's BlockedBy
	pd, err := board.ParseProgressFile(elem.ProgressPath())
	if err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("reading progress for %s: %v", id, err), nil)
			return nil
		}
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
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.ValidationError, fmt.Sprintf("%s is not blocked by %s", id, blocker.ID()), map[string]interface{}{
				"source": id,
				"target": blocker.ID(),
			})
			return nil
		}
		return fmt.Errorf("%s is not blocked by %s", id, blocker.ID())
	}
	pd.BlockedBy = newBlockedBy
	if err := board.WriteProgressFile(elem.ProgressPath(), pd); err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("writing progress for %s: %v", id, err), nil)
			return nil
		}
		return fmt.Errorf("writing progress for %s: %w", id, err)
	}

	// Remove from blocker's Blocks
	blockerPd, err := board.ParseProgressFile(blocker.ProgressPath())
	if err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("reading progress for %s: %v", blocker.ID(), err), nil)
			return nil
		}
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
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("writing progress for %s: %v", blocker.ID(), err), nil)
			return nil
		}
		return fmt.Errorf("writing progress for %s: %w", blocker.ID(), err)
	}

	// --- De-escalate dependencies up the hierarchy ---
	if err := deescalateDependency(b, elem, blocker); err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("de-escalating dependency: %v", err), nil)
			return nil
		}
		return fmt.Errorf("de-escalating dependency: %w", err)
	}

	// Output result
	if JSONEnabled() {
		response := UnlinkResponse{
			Updated: UnlinkUpdate{
				Source:   id,
				Target:   blocker.ID(),
				Relation: "blocked-by",
			},
			Message: fmt.Sprintf("%s no longer blocked by %s", id, blocker.ID()),
		}
		return output.PrintJSON(os.Stdout, response)
	}

	fmt.Printf("%s: removed blocked-by %s\n", id, blocker.ID())
	fmt.Printf("%s: removed blocks %s\n", blocker.ID(), id)

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
