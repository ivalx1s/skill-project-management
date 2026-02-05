// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"board-tui/internal/app"
	"board-tui/internal/state"
	"board-tui/internal/ui/screens/board"
	"board-tui/internal/ui/screens/detail"
)

// =============================================================================
// Effect implementations (IO operations)
// =============================================================================

func loadTreeEffect() tea.Msg {
	tree, err := loadTreeFromCLI()
	return state.TreeLoadedMsg{Tree: tree, LoadErr: err}
}

func loadTreeFromCLI() ([]*board.TreeNode, error) {
	cmd := exec.Command("task-board", "tree", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var response struct {
		Tree []*board.TreeNode `json:"tree"`
	}
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, err
	}

	// Build parent references and set depth
	for _, root := range response.Tree {
		setParentRefs(root, nil, 0)
	}

	return response.Tree, nil
}

func setParentRefs(node *board.TreeNode, parent *board.TreeNode, depth int) {
	node.Parent = parent
	node.Depth = depth
	for _, child := range node.Children {
		setParentRefs(child, node, depth+1)
	}
}

func loadElementEffect(id string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("task-board", "show", id, "--json")
		output, err := cmd.Output()
		if err != nil {
			return state.ElementLoadedMsg{Err: err}
		}

		var response struct {
			Element detail.Element `json:"element"`
		}
		if err := json.Unmarshal(output, &response); err != nil {
			return state.ElementLoadedMsg{Err: err}
		}

		return state.ElementLoadedMsg{Element: &response.Element}
	}
}

// =============================================================================
// Configuration
// =============================================================================

type fileConfig struct {
	RefreshRate   int      `json:"refresh_rate"`
	ExpandedNodes []string `json:"expanded_nodes"`
}

func loadConfig() (*app.Config, error) {
	data, err := os.ReadFile(".task-board/config.json")
	if err != nil {
		if os.IsNotExist(err) {
			return &app.Config{RefreshRate: 10}, nil
		}
		return nil, err
	}

	var cfg fileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &app.Config{
		RefreshRate:   cfg.RefreshRate,
		ExpandedNodes: cfg.ExpandedNodes,
	}, nil
}

func saveConfig(config *state.Config) tea.Msg {
	if config == nil {
		return state.ConfigSavedMsg{}
	}

	cfg := fileConfig{
		RefreshRate:   config.RefreshRate,
		ExpandedNodes: config.ExpandedNodes,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return state.ConfigSavedMsg{Err: err}
	}

	err = os.WriteFile(".task-board/config.json", data, 0644)
	return state.ConfigSavedMsg{Err: err}
}

// =============================================================================
// Logger implementation
// =============================================================================

type fileLogger struct {
	file *os.File
}

func newFileLogger(dir string) (*fileLogger, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%s/session_%s.log", dir, time.Now().Format("20060102_150405"))
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return &fileLogger{file: f}, nil
}

func (l *fileLogger) log(level, format string, args ...interface{}) {
	if l.file == nil {
		return
	}
	timestamp := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.file, "%s [%s] %s\n", timestamp, level, msg)
}

func (l *fileLogger) Debug(format string, args ...interface{}) { l.log("DEBUG", format, args...) }
func (l *fileLogger) Info(format string, args ...interface{})  { l.log("INFO", format, args...) }
func (l *fileLogger) Warn(format string, args ...interface{})  { l.log("WARN", format, args...) }
func (l *fileLogger) Key(key string)                           { l.log("KEY", "pressed: %s", key) }
func (l *fileLogger) Action(action, detail string)             { l.log("ACTION", "%s: %s", action, detail) }
func (l *fileLogger) Command(cmd, args, result string) {
	l.log("CMD", "/%s %s -> %s", cmd, args, result)
}
func (l *fileLogger) State(format string, args ...interface{}) { l.log("STATE", format, args...) }
func (l *fileLogger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

// =============================================================================
// Main
// =============================================================================

func main() {
	// Warm up markdown renderer
	detail.InitMarkdownRenderer()

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
	}
	if cfg == nil {
		cfg = &app.Config{RefreshRate: 10}
	}

	// Create logger
	logger, err := newFileLogger(".task-board/logs")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create logger: %v\n", err)
	}

	// Create effects
	effects := &state.Effects{
		LoadTree: loadTreeEffect,
		LoadElement: loadElementEffect,
		SaveConfig: saveConfig,
	}

	// Create model
	model := app.New(app.Options{
		Effects: effects,
		Logger:  logger,
		Config:  cfg,
	})

	// Run program
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
