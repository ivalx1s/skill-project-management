package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/aagrigore/task-board/templates"
	"github.com/spf13/cobra"
)

// CreateResponse is the JSON response for create commands
type CreateResponse struct {
	Created CreatedElement `json:"created"`
}

// CreatedElement represents a newly created element
type CreatedElement struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Parent string `json:"parent,omitempty"`
	Path   string `json:"path"`
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a board element",
}

var createEpicCmd = &cobra.Command{
	Use:   "epic",
	Short: "Create a new epic",
	RunE:  runCreateEpic,
}

var createStoryCmd = &cobra.Command{
	Use:   "story",
	Short: "Create a new story",
	RunE:  runCreateStory,
}

var createTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Create a new task",
	RunE:  runCreateTask,
}

var createBugCmd = &cobra.Command{
	Use:   "bug",
	Short: "Create a new bug",
	RunE:  runCreateBug,
}

var (
	createName        string
	createDescription string
	createEpicFlag    string
	createStoryFlag   string
)

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.AddCommand(createEpicCmd)
	createCmd.AddCommand(createStoryCmd)
	createCmd.AddCommand(createTaskCmd)
	createCmd.AddCommand(createBugCmd)

	// Common flags
	createEpicCmd.Flags().StringVar(&createName, "name", "", "Element name (required)")
	createEpicCmd.Flags().StringVar(&createDescription, "description", "", "Element description")
	createEpicCmd.MarkFlagRequired("name")

	createStoryCmd.Flags().StringVar(&createName, "name", "", "Element name (required)")
	createStoryCmd.Flags().StringVar(&createDescription, "description", "", "Element description")
	createStoryCmd.Flags().StringVar(&createEpicFlag, "epic", "", "Parent epic ID (required)")
	createStoryCmd.MarkFlagRequired("name")
	createStoryCmd.MarkFlagRequired("epic")

	createTaskCmd.Flags().StringVar(&createName, "name", "", "Element name (required)")
	createTaskCmd.Flags().StringVar(&createDescription, "description", "", "Element description")
	createTaskCmd.Flags().StringVar(&createStoryFlag, "story", "", "Parent story ID (required)")
	createTaskCmd.MarkFlagRequired("name")
	createTaskCmd.MarkFlagRequired("story")

	createBugCmd.Flags().StringVar(&createName, "name", "", "Element name (required)")
	createBugCmd.Flags().StringVar(&createDescription, "description", "", "Element description")
	createBugCmd.Flags().StringVar(&createStoryFlag, "story", "", "Parent story ID (required)")
	createBugCmd.MarkFlagRequired("name")
	createBugCmd.MarkFlagRequired("story")
}

func runCreateEpic(cmd *cobra.Command, args []string) error {
	return createElement(board.EpicType, createName, createDescription, "")
}

func runCreateStory(cmd *cobra.Command, args []string) error {
	return createElement(board.StoryType, createName, createDescription, createEpicFlag)
}

func runCreateTask(cmd *cobra.Command, args []string) error {
	return createElement(board.TaskType, createName, createDescription, createStoryFlag)
}

func runCreateBug(cmd *cobra.Command, args []string) error {
	return createElement(board.BugType, createName, createDescription, createStoryFlag)
}

func createElement(elemType board.ElementType, name, description, parentID string) error {
	if err := board.EnsureBoardDir(boardDir); err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, err.Error(), nil)
			return nil
		}
		return err
	}

	// Find parent directory
	parentDir := boardDir
	if parentID != "" {
		b, err := board.Load(boardDir)
		if err != nil {
			if JSONEnabled() {
				output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("loading board: %s", err.Error()), nil)
				return nil
			}
			return fmt.Errorf("loading board: %w", err)
		}
		parent := b.FindByID(parentID)
		if parent == nil {
			if JSONEnabled() {
				output.PrintError(os.Stderr, output.NotFound, fmt.Sprintf("parent %s not found", parentID), map[string]interface{}{
					"parent_id": parentID,
				})
				return nil
			}
			return fmt.Errorf("parent %s not found", parentID)
		}
		parentDir = parent.Path
	}

	// Generate distributed ID (YYMMDD-xxxxxx format)
	id := board.GenerateID(elemType)

	// Create directory
	sanitized := board.SanitizeName(name)
	dirName := fmt.Sprintf("%s_%s", id, sanitized)
	elemPath := filepath.Join(parentDir, dirName)

	if err := os.MkdirAll(elemPath, 0755); err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("creating directory: %s", err.Error()), nil)
			return nil
		}
		return fmt.Errorf("creating directory: %w", err)
	}

	// Render and write README.md
	title := fmt.Sprintf("%s: %s", id, name)
	readmeContent, err := templates.RenderReadme(string(elemType), templates.TemplateData{
		Title:       title,
		Description: description,
	})
	if err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("rendering readme template: %s", err.Error()), nil)
			return nil
		}
		return fmt.Errorf("rendering readme template: %w", err)
	}
	if err := os.WriteFile(filepath.Join(elemPath, "README.md"), []byte(readmeContent), 0644); err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("writing README.md: %s", err.Error()), nil)
			return nil
		}
		return fmt.Errorf("writing README.md: %w", err)
	}

	// Render and write progress.md
	progressContent, err := templates.RenderProgress(string(elemType))
	if err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("rendering progress template: %s", err.Error()), nil)
			return nil
		}
		return fmt.Errorf("rendering progress template: %w", err)
	}
	if err := os.WriteFile(filepath.Join(elemPath, "progress.md"), []byte(progressContent), 0644); err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("writing progress.md: %s", err.Error()), nil)
			return nil
		}
		return fmt.Errorf("writing progress.md: %w", err)
	}

	// Set CreatedAt timestamp
	progressPath := filepath.Join(elemPath, "progress.md")
	pd, err := board.ParseProgressFile(progressPath)
	if err == nil {
		pd.CreatedAt = time.Now().UTC()
		board.WriteProgressFile(progressPath, pd)
	}

	if JSONEnabled() {
		// Compute relative path from board root
		relPath := computeRelativePath(boardDir, elemPath)

		resp := CreateResponse{
			Created: CreatedElement{
				ID:     id,
				Type:   string(elemType),
				Name:   name,
				Status: string(board.StatusBacklog),
				Parent: parentID,
				Path:   relPath,
			},
		}
		output.PrintJSON(os.Stdout, resp)
	} else {
		fmt.Printf("Created %s: %s\n", id, name)
		fmt.Printf("  Path: %s\n", elemPath)
	}
	return nil
}

// computeRelativePath computes a hierarchical path like "EPIC-X/STORY-Y/TASK-Z"
// from the board directory to the element path
func computeRelativePath(boardDir, elemPath string) string {
	rel, err := filepath.Rel(boardDir, elemPath)
	if err != nil {
		return filepath.Base(elemPath)
	}
	// Convert path separators to forward slashes for consistent output
	return strings.ReplaceAll(rel, string(filepath.Separator), "/")
}
