package board

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Board represents the full task board loaded from disk.
type Board struct {
	Dir      string
	Counters *Counters
	Elements []*Element
}

// Load reads the entire board from the given directory.
func Load(boardDir string) (*Board, error) {
	b := &Board{Dir: boardDir}

	counters, err := ReadCounters(boardDir)
	if err != nil {
		return nil, err
	}
	b.Counters = counters

	// Walk epics
	entries, err := os.ReadDir(boardDir)
	if err != nil {
		return nil, fmt.Errorf("reading board directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		elemType, num, name, err := ParseDirName(entry.Name())
		if err != nil {
			continue // skip non-element directories
		}
		if elemType != EpicType {
			continue
		}

		epic := &Element{
			Type:   EpicType,
			Number: num,
			Name:   name,
			Path:   filepath.Join(boardDir, entry.Name()),
		}
		loadElementDetails(epic)
		b.Elements = append(b.Elements, epic)

		// Walk stories inside epic
		storyEntries, err := os.ReadDir(epic.Path)
		if err != nil {
			continue
		}
		for _, se := range storyEntries {
			if !se.IsDir() {
				continue
			}
			st, sn, sname, err := ParseDirName(se.Name())
			if err != nil {
				continue
			}
			if st != StoryType {
				continue
			}

			story := &Element{
				Type:     StoryType,
				Number:   sn,
				Name:     sname,
				Path:     filepath.Join(epic.Path, se.Name()),
				ParentID: epic.ID(),
			}
			loadElementDetails(story)
			b.Elements = append(b.Elements, story)

			// Walk tasks/bugs inside story
			taskEntries, err := os.ReadDir(story.Path)
			if err != nil {
				continue
			}
			for _, te := range taskEntries {
				if !te.IsDir() {
					continue
				}
				tt, tn, tname, err := ParseDirName(te.Name())
				if err != nil {
					continue
				}
				if tt != TaskType && tt != BugType {
					continue
				}

				task := &Element{
					Type:     tt,
					Number:   tn,
					Name:     tname,
					Path:     filepath.Join(story.Path, te.Name()),
					ParentID: story.ID(),
				}
				loadElementDetails(task)
				b.Elements = append(b.Elements, task)
			}
		}
	}

	return b, nil
}

func loadElementDetails(e *Element) {
	// Load progress
	pd, err := ParseProgressFile(e.ProgressPath())
	if err == nil {
		e.Status = pd.Status
		e.AssignedTo = pd.AssignedTo
		e.CreatedAt = pd.CreatedAt
		e.LastUpdate = pd.LastUpdate
		e.BlockedBy = pd.BlockedBy
		e.Blocks = pd.Blocks
		e.Checklist = pd.Checklist
	} else {
		e.Status = StatusOpen
	}

	// Load readme
	rd, err := ParseReadmeFile(e.ReadmePath())
	if err == nil {
		e.Title = rd.Title
		e.Description = rd.Description
		e.Scope = rd.Scope
		e.AC = rd.AC
	}
}

// FindByID finds an element by its ID (e.g. "TASK-12").
func (b *Board) FindByID(id string) *Element {
	id = strings.ToUpper(id)
	for _, e := range b.Elements {
		if e.ID() == id {
			return e
		}
	}
	return nil
}

// FindByType returns all elements of the given type.
func (b *Board) FindByType(t ElementType) []*Element {
	var result []*Element
	for _, e := range b.Elements {
		if e.Type == t {
			result = append(result, e)
		}
	}
	return result
}

// FilterByStatus filters elements by status.
func FilterByStatus(elements []*Element, status Status) []*Element {
	var result []*Element
	for _, e := range elements {
		if e.Status == status {
			result = append(result, e)
		}
	}
	return result
}

// FilterByParent filters elements by parent ID.
func FilterByParent(elements []*Element, parentID string) []*Element {
	parentID = strings.ToUpper(parentID)
	var result []*Element
	for _, e := range elements {
		if e.ParentID == parentID {
			result = append(result, e)
		}
	}
	return result
}

// Children returns direct children of an element.
func (b *Board) Children(e *Element) []*Element {
	var result []*Element
	for _, el := range b.Elements {
		if el.ParentID == e.ID() {
			result = append(result, el)
		}
	}
	return result
}

// ParentOf returns the parent element, or nil if the element is a top-level epic.
func (b *Board) ParentOf(e *Element) *Element {
	if e.ParentID == "" {
		return nil
	}
	return b.FindByID(e.ParentID)
}

// HasCrossChildDependency checks if any child of `from` has a blocked-by dependency
// on any child of `to`. Used to determine if an escalated dependency should remain.
func (b *Board) HasCrossChildDependency(from, to *Element) bool {
	fromChildren := b.Children(from)
	toChildren := b.Children(to)

	toChildIDs := make(map[string]bool)
	for _, tc := range toChildren {
		toChildIDs[tc.ID()] = true
	}

	for _, fc := range fromChildren {
		for _, blockerID := range fc.BlockedBy {
			if toChildIDs[blockerID] {
				return true
			}
		}
	}
	return false
}

// Ancestry returns the path from root to element (e.g. "EPIC-01 > STORY-05 > TASK-12").
func (b *Board) Ancestry(e *Element) string {
	var parts []string
	current := e
	for current != nil {
		parts = append([]string{current.ID()}, parts...)
		if current.ParentID == "" {
			break
		}
		current = b.FindByID(current.ParentID)
	}
	return strings.Join(parts, " > ")
}
