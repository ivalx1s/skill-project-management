package cmd

import (
	"fmt"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <ID>",
	Short: "Update element README.md fields",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

var (
	updateTitle       string
	updateDescription string
	updateScope       string
	updateAC          string
)

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New title")
	updateCmd.Flags().StringVar(&updateDescription, "description", "", "New description")
	updateCmd.Flags().StringVar(&updateScope, "scope", "", "New scope")
	updateCmd.Flags().StringVar(&updateAC, "ac", "", "New acceptance criteria")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	id := args[0]

	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	elem := b.FindByID(id)
	if elem == nil {
		return fmt.Errorf("element %s not found", id)
	}

	rd, err := board.ParseReadmeFile(elem.ReadmePath())
	if err != nil {
		return fmt.Errorf("reading README.md: %w", err)
	}

	changed := false
	if cmd.Flags().Changed("title") {
		rd.Title = updateTitle
		changed = true
	}
	if cmd.Flags().Changed("description") {
		rd.Description = updateDescription
		changed = true
	}
	if cmd.Flags().Changed("scope") {
		rd.Scope = updateScope
		changed = true
	}
	if cmd.Flags().Changed("ac") {
		rd.AC = updateAC
		changed = true
	}

	if !changed {
		fmt.Println("No changes specified. Use --title, --description, --scope, or --ac flags.")
		return nil
	}

	if err := board.WriteReadmeFile(elem.ReadmePath(), rd); err != nil {
		return fmt.Errorf("writing README.md: %w", err)
	}

	fmt.Printf("Updated %s\n", id)
	return nil
}
