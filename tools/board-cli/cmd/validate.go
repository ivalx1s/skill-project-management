package cmd

import (
	"fmt"
	"os"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

// ValidateResponse is the JSON response structure for validate command
type ValidateResponse struct {
	Valid    bool            `json:"valid"`
	Errors   []ValidateIssue `json:"errors"`
	Warnings []ValidateIssue `json:"warnings"`
}

// ValidateIssue represents a validation error or warning
type ValidateIssue struct {
	Code       string   `json:"code"`
	Message    string   `json:"message"`
	ElementID  string   `json:"elementId,omitempty"`
	ElementIDs []string `json:"elementIds,omitempty"`
}

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
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("loading board: %v", err), nil)
		}
		return fmt.Errorf("loading board: %w", err)
	}

	var errors []ValidateIssue
	var warnings []ValidateIssue

	for _, e := range b.Elements {
		// Check README.md exists
		if _, err := os.Stat(e.ReadmePath()); os.IsNotExist(err) {
			issue := ValidateIssue{
				Code:      "MISSING_README",
				Message:   fmt.Sprintf("%s: README.md missing", e.ID()),
				ElementID: e.ID(),
			}
			errors = append(errors, issue)
			if !JSONEnabled() {
				fmt.Printf("%s[MISSING]%s %s: README.md\n", output.Red, output.Reset, e.ID())
			}
		}

		// Check progress.md exists
		if _, err := os.Stat(e.ProgressPath()); os.IsNotExist(err) {
			issue := ValidateIssue{
				Code:      "MISSING_PROGRESS",
				Message:   fmt.Sprintf("%s: progress.md missing", e.ID()),
				ElementID: e.ID(),
			}
			errors = append(errors, issue)
			if !JSONEnabled() {
				fmt.Printf("%s[MISSING]%s %s: progress.md\n", output.Red, output.Reset, e.ID())
			}
		}

		// Validate naming convention
		_, _, _, err := board.ParseDirName(e.DirName())
		if err != nil {
			issue := ValidateIssue{
				Code:      "NAMING_ERROR",
				Message:   fmt.Sprintf("%s: %v", e.ID(), err),
				ElementID: e.ID(),
			}
			warnings = append(warnings, issue)
			if !JSONEnabled() {
				fmt.Printf("%s[NAMING]%s %s: %v\n", output.Yellow, output.Reset, e.ID(), err)
			}
		}

		// Validate blockedBy references
		for _, blockerID := range e.BlockedBy {
			if blockerID == "(none)" {
				continue
			}
			if b.FindByID(blockerID) == nil {
				issue := ValidateIssue{
					Code:       "BROKEN_LINK",
					Message:    fmt.Sprintf("%s: blockedBy %s (not found)", e.ID(), blockerID),
					ElementID:  e.ID(),
					ElementIDs: []string{e.ID(), blockerID},
				}
				errors = append(errors, issue)
				if !JSONEnabled() {
					fmt.Printf("%s[BROKEN LINK]%s %s: blockedBy %s (not found)\n",
						output.Red, output.Reset, e.ID(), blockerID)
				}
			}
		}

		// Check orphans
		switch e.Type {
		case board.StoryType:
			if e.ParentID == "" {
				issue := ValidateIssue{
					Code:      "ORPHAN_ELEMENT",
					Message:   fmt.Sprintf("%s: story without epic", e.ID()),
					ElementID: e.ID(),
				}
				warnings = append(warnings, issue)
				if !JSONEnabled() {
					fmt.Printf("%s[ORPHAN]%s %s: story without epic\n", output.Yellow, output.Reset, e.ID())
				}
			}
		case board.TaskType, board.BugType:
			if e.ParentID == "" {
				issue := ValidateIssue{
					Code:      "ORPHAN_ELEMENT",
					Message:   fmt.Sprintf("%s: %s without story", e.ID(), e.Type),
					ElementID: e.ID(),
				}
				warnings = append(warnings, issue)
				if !JSONEnabled() {
					fmt.Printf("%s[ORPHAN]%s %s: %s without story\n", output.Yellow, output.Reset, e.ID(), e.Type)
				}
			}
		}
	}

	// JSON output
	if JSONEnabled() {
		// Ensure slices are not nil for JSON
		if errors == nil {
			errors = []ValidateIssue{}
		}
		if warnings == nil {
			warnings = []ValidateIssue{}
		}

		response := ValidateResponse{
			Valid:    len(errors) == 0,
			Errors:   errors,
			Warnings: warnings,
		}
		return output.PrintJSON(os.Stdout, response)
	}

	// Text output summary
	issues := len(errors) + len(warnings)
	if issues == 0 {
		fmt.Println("Board is valid. No issues found.")
	} else {
		fmt.Printf("\n%d issue(s) found.\n", issues)
	}

	return nil
}
