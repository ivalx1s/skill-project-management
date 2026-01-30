package board

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// ElementType represents the type of board element.
type ElementType string

const (
	EpicType  ElementType = "epic"
	StoryType ElementType = "story"
	TaskType  ElementType = "task"
	BugType   ElementType = "bug"
)

// Status represents the status of an element.
type Status string

const (
	StatusOpen     Status = "open"
	StatusProgress Status = "progress"
	StatusDone     Status = "done"
	StatusClosed   Status = "closed"
	StatusBlocked  Status = "blocked"
)

func ParseStatus(s string) (Status, error) {
	switch strings.ToLower(s) {
	case "open":
		return StatusOpen, nil
	case "progress":
		return StatusProgress, nil
	case "done":
		return StatusDone, nil
	case "closed":
		return StatusClosed, nil
	case "blocked":
		return StatusBlocked, nil
	default:
		return "", fmt.Errorf("unknown status: %s (valid: open, progress, done, closed, blocked)", s)
	}
}

func ParseElementType(s string) (ElementType, error) {
	switch strings.ToLower(s) {
	case "epic", "epics":
		return EpicType, nil
	case "story", "stories":
		return StoryType, nil
	case "task", "tasks":
		return TaskType, nil
	case "bug", "bugs":
		return BugType, nil
	default:
		return "", fmt.Errorf("unknown element type: %s (valid: epic, story, task, bug)", s)
	}
}

func (t ElementType) Prefix() string {
	switch t {
	case EpicType:
		return "EPIC"
	case StoryType:
		return "STORY"
	case TaskType:
		return "TASK"
	case BugType:
		return "BUG"
	default:
		return ""
	}
}

func (t ElementType) CounterKey() string {
	return string(t)
}

// Element represents a board element (epic, story, task, or bug).
type Element struct {
	Type      ElementType
	Number    int
	Name      string
	Path      string // absolute path to element directory
	ParentID  string // e.g. "EPIC-01" for a story, "STORY-05" for a task
	Status     Status
	AssignedTo string
	CreatedAt  time.Time
	LastUpdate time.Time
	BlockedBy  []string
	Blocks     []string
	Checklist  []ChecklistItem
	// README fields
	Title       string
	Description string
	Scope       string
	AC          string // acceptance criteria
}

type ChecklistItem struct {
	Text    string
	Checked bool
}

// ID returns the element ID, e.g. "EPIC-01", "TASK-12".
func (e *Element) ID() string {
	return fmt.Sprintf("%s-%02d", e.Type.Prefix(), e.Number)
}

// DirName returns the directory name, e.g. "EPIC-01_recording".
func (e *Element) DirName() string {
	return fmt.Sprintf("%s-%02d_%s", e.Type.Prefix(), e.Number, e.Name)
}

// ReadmePath returns the path to README.md.
func (e *Element) ReadmePath() string {
	return filepath.Join(e.Path, "README.md")
}

// ProgressPath returns the path to progress.md.
func (e *Element) ProgressPath() string {
	return filepath.Join(e.Path, "progress.md")
}
