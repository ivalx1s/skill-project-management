package cmd

import (
	"fmt"
	"os"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate board structure",
	RunE:  runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	issues := 0

	for _, e := range b.Elements {
		// Check README.md exists
		if _, err := os.Stat(e.ReadmePath()); os.IsNotExist(err) {
			fmt.Printf("%s[MISSING]%s %s: README.md\n", output.Red, output.Reset, e.ID())
			issues++
		}

		// Check progress.md exists
		if _, err := os.Stat(e.ProgressPath()); os.IsNotExist(err) {
			fmt.Printf("%s[MISSING]%s %s: progress.md\n", output.Red, output.Reset, e.ID())
			issues++
		}

		// Validate naming convention
		_, _, _, err := board.ParseDirName(e.DirName())
		if err != nil {
			fmt.Printf("%s[NAMING]%s %s: %v\n", output.Yellow, output.Reset, e.ID(), err)
			issues++
		}

		// Validate blockedBy references
		for _, blockerID := range e.BlockedBy {
			if blockerID == "(none)" {
				continue
			}
			if b.FindByID(blockerID) == nil {
				fmt.Printf("%s[BROKEN LINK]%s %s: blockedBy %s (not found)\n",
					output.Red, output.Reset, e.ID(), blockerID)
				issues++
			}
		}

		// Check orphans
		switch e.Type {
		case board.StoryType:
			if e.ParentID == "" {
				fmt.Printf("%s[ORPHAN]%s %s: story without epic\n", output.Yellow, output.Reset, e.ID())
				issues++
			}
		case board.TaskType, board.BugType:
			if e.ParentID == "" {
				fmt.Printf("%s[ORPHAN]%s %s: %s without story\n", output.Yellow, output.Reset, e.ID(), e.Type)
				issues++
			}
		}
	}

	if issues == 0 {
		fmt.Println("Board is valid. No issues found.")
	} else {
		fmt.Printf("\n%d issue(s) found.\n", issues)
	}

	return nil
}
