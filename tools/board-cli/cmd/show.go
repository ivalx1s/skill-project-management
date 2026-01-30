package cmd

import (
	"fmt"
	"strings"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <ID>",
	Short: "Show full element details",
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) error {
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

	rd, err := board.ParseReadmeFile(elem.ReadmePath())
	if err != nil {
		return fmt.Errorf("reading README: %w", err)
	}

	// Header
	fmt.Printf("%s%s: %s%s\n", output.Bold, elem.ID(), rd.Title, output.Reset)
	fmt.Printf("Path: %s\n", b.Ancestry(elem))
	fmt.Printf("Status: %s\n", output.ColorStatus(string(pd.Status)))
	fmt.Println()

	// Description
	if rd.Description != "" {
		fmt.Println("Description:")
		for _, line := range strings.Split(rd.Description, "\n") {
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}

	// Scope
	if rd.Scope != "" {
		fmt.Println("Scope:")
		for _, line := range strings.Split(rd.Scope, "\n") {
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}

	// Acceptance Criteria
	if rd.AC != "" {
		fmt.Println("Acceptance Criteria:")
		for _, line := range strings.Split(rd.AC, "\n") {
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}

	// Checklist
	if len(pd.Checklist) > 0 {
		fmt.Println("Checklist:")
		for i, item := range pd.Checklist {
			mark := "[ ]"
			if item.Checked {
				mark = "[x]"
			}
			fmt.Printf("  %d. %s %s\n", i+1, mark, item.Text)
		}
		fmt.Println()
	}

	// Blocked By
	if len(pd.BlockedBy) > 0 {
		fmt.Printf("Blocked By: %s\n", strings.Join(pd.BlockedBy, ", "))
	} else {
		fmt.Println("Blocked By: (none)")
	}

	// Blocks
	if len(pd.Blocks) > 0 {
		fmt.Printf("Blocks: %s\n", strings.Join(pd.Blocks, ", "))
	} else {
		fmt.Println("Blocks: (none)")
	}

	// Notes
	if pd.Notes != "" {
		fmt.Println()
		fmt.Println("Notes:")
		for _, line := range strings.Split(pd.Notes, "\n") {
			fmt.Printf("  %s\n", line)
		}
	}

	return nil
}
