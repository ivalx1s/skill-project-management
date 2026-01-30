package board

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Counters holds global auto-increment counters for each element type.
type Counters struct {
	Epic  int
	Story int
	Task  int
	Bug   int
}

func (c *Counters) Get(t ElementType) int {
	switch t {
	case EpicType:
		return c.Epic
	case StoryType:
		return c.Story
	case TaskType:
		return c.Task
	case BugType:
		return c.Bug
	default:
		return 0
	}
}

func (c *Counters) Set(t ElementType, val int) {
	switch t {
	case EpicType:
		c.Epic = val
	case StoryType:
		c.Story = val
	case TaskType:
		c.Task = val
	case BugType:
		c.Bug = val
	}
}

func (c *Counters) Increment(t ElementType) int {
	next := c.Get(t) + 1
	c.Set(t, next)
	return next
}

const systemFileName = "system.md"

// SystemPath returns the path to system.md in the board directory.
func SystemPath(boardDir string) string {
	return filepath.Join(boardDir, systemFileName)
}

// ReadCounters reads counters from system.md.
func ReadCounters(boardDir string) (*Counters, error) {
	path := SystemPath(boardDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Counters{}, nil
		}
		return nil, fmt.Errorf("reading system.md: %w", err)
	}
	return parseCounters(string(data))
}

func parseCounters(content string) (*Counters, error) {
	c := &Counters{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "- ") {
			continue
		}
		line = strings.TrimPrefix(line, "- ")
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			continue
		}
		switch key {
		case "epic":
			c.Epic = val
		case "story":
			c.Story = val
		case "task":
			c.Task = val
		case "bug":
			c.Bug = val
		}
	}
	return c, scanner.Err()
}

// WriteCounters writes counters to system.md.
func WriteCounters(boardDir string, c *Counters) error {
	path := SystemPath(boardDir)
	content := fmt.Sprintf(`## Counters
- epic: %d
- story: %d
- task: %d
- bug: %d
`, c.Epic, c.Story, c.Task, c.Bug)
	return os.WriteFile(path, []byte(content), 0644)
}

// EnsureBoardDir creates the board directory and system.md if they don't exist.
func EnsureBoardDir(boardDir string) error {
	if err := os.MkdirAll(boardDir, 0755); err != nil {
		return fmt.Errorf("creating board directory: %w", err)
	}
	path := SystemPath(boardDir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return WriteCounters(boardDir, &Counters{})
	}
	return nil
}
