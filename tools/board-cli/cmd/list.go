package cmd

import (
	"fmt"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <epics|stories|tasks|bugs>",
	Short: "List board elements",
	Args:  cobra.ExactArgs(1),
	RunE:  runList,
}

var (
	listStatus string
	listEpic   string
	listStory  string
)

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status")
	listCmd.Flags().StringVar(&listEpic, "epic", "", "Filter by epic ID")
	listCmd.Flags().StringVar(&listStory, "story", "", "Filter by story ID")
}

func runList(cmd *cobra.Command, args []string) error {
	elemType, err := board.ParseElementType(args[0])
	if err != nil {
		return err
	}

	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	elements := b.FindByType(elemType)

	// Apply filters
	if listStatus != "" {
		status, err := board.ParseStatus(listStatus)
		if err != nil {
			return err
		}
		elements = board.FilterByStatus(elements, status)
	}

	if listEpic != "" {
		elements = board.FilterByParent(elements, listEpic)
	}

	if listStory != "" {
		elements = board.FilterByParent(elements, listStory)
	}

	if len(elements) == 0 {
		fmt.Println("No elements found.")
		return nil
	}

	table := output.NewTable("ID", "STATUS", "NAME", "PARENT")
	for _, e := range elements {
		statusStr := output.ColorStatus(string(e.Status))
		// Show [BLOCKED by X] if element has active blockers
		if activeBlockers := b.ActiveBlockers(e); len(activeBlockers) > 0 {
			var blockerIDs []string
			for _, blocker := range activeBlockers {
				blockerIDs = append(blockerIDs, blocker.ID())
			}
			statusStr += fmt.Sprintf(" %s[BLOCKED by %s]%s", output.Red, blockerIDs[0], output.Reset)
			if len(blockerIDs) > 1 {
				statusStr = statusStr[:len(statusStr)-len(output.Reset)] + fmt.Sprintf("+%d%s", len(blockerIDs)-1, output.Reset)
			}
		}
		table.AddRow(
			e.ID(),
			statusStr,
			e.Name,
			e.ParentID,
		)
	}

	fmt.Print(table.String())
	return nil
}
