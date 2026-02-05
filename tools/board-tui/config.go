package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Default configuration values
const (
	DefaultRefreshRate = 10 // seconds
)

// Config holds the TUI configuration
type Config struct {
	// RefreshRate is the auto-refresh interval in seconds (0 = off)
	RefreshRate int `json:"refreshRate"`
	// ExpandedNodes is a list of node IDs to expand on startup
	ExpandedNodes []string `json:"expandedNodes"`
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		RefreshRate:   DefaultRefreshRate,
		ExpandedNodes: []string{},
	}
}

// DefaultConfigPath returns the default config file path: ~/.config/board-tui/config.json
func DefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "board-tui", "config.json"), nil
}

// LoadConfig loads configuration from the default path.
// Returns default config if the file doesn't exist.
func LoadConfig() (*Config, error) {
	configPath, err := DefaultConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	return LoadConfigFromPath(configPath)
}

// LoadConfigFromPath loads configuration from a specific path.
// Returns default config if the file doesn't exist.
func LoadConfigFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist - return defaults
			return DefaultConfig(), nil
		}
		return DefaultConfig(), err
	}

	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return DefaultConfig(), err
	}

	return config, nil
}

// SaveConfig saves configuration to the default path.
// Creates the config directory if it doesn't exist.
func (c *Config) SaveConfig() error {
	configPath, err := DefaultConfigPath()
	if err != nil {
		return err
	}

	return c.SaveConfigToPath(configPath)
}

// SaveConfigToPath saves configuration to a specific path.
// Creates the parent directory if it doesn't exist.
func (c *Config) SaveConfigToPath(path string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Write with newline at end
	data = append(data, '\n')

	return os.WriteFile(path, data, 0644)
}

// GetRefreshDuration returns the refresh rate as a time.Duration.
// Returns 0 if auto-refresh is disabled (RefreshRate == 0).
func (c *Config) GetRefreshDuration() time.Duration {
	if c.RefreshRate <= 0 {
		return 0
	}
	return time.Duration(c.RefreshRate) * time.Second
}

// SetExpandedNodes updates the list of expanded node IDs
func (c *Config) SetExpandedNodes(nodes []string) {
	c.ExpandedNodes = nodes
}

// AddExpandedNode adds a node ID to the expanded list if not already present
func (c *Config) AddExpandedNode(id string) {
	for _, existing := range c.ExpandedNodes {
		if existing == id {
			return // Already in list
		}
	}
	c.ExpandedNodes = append(c.ExpandedNodes, id)
}

// RemoveExpandedNode removes a node ID from the expanded list
func (c *Config) RemoveExpandedNode(id string) {
	for i, existing := range c.ExpandedNodes {
		if existing == id {
			c.ExpandedNodes = append(c.ExpandedNodes[:i], c.ExpandedNodes[i+1:]...)
			return
		}
	}
}

// IsExpanded checks if a node ID is in the expanded list
func (c *Config) IsExpanded(id string) bool {
	for _, existing := range c.ExpandedNodes {
		if existing == id {
			return true
		}
	}
	return false
}

// CollectExpandedNodes collects IDs of all expanded nodes from a tree
func CollectExpandedNodes(roots []*TreeNode) []string {
	var expanded []string
	for _, root := range roots {
		collectExpandedRecursive(root, &expanded)
	}
	return expanded
}

// collectExpandedRecursive recursively collects expanded node IDs
func collectExpandedRecursive(node *TreeNode, expanded *[]string) {
	if node.Expanded {
		*expanded = append(*expanded, node.ID)
	}
	for _, child := range node.Children {
		collectExpandedRecursive(child, expanded)
	}
}

// ApplyExpandedNodes applies the expanded state from config to a tree
func ApplyExpandedNodes(roots []*TreeNode, expandedIDs []string) {
	// Build a set for O(1) lookup
	expandedSet := make(map[string]bool)
	for _, id := range expandedIDs {
		expandedSet[id] = true
	}

	for _, root := range roots {
		applyExpandedRecursive(root, expandedSet)
	}
}

// applyExpandedRecursive recursively applies expanded state
func applyExpandedRecursive(node *TreeNode, expandedSet map[string]bool) {
	// Set expanded state based on config (overrides defaults)
	node.Expanded = expandedSet[node.ID]
	for _, child := range node.Children {
		applyExpandedRecursive(child, expandedSet)
	}
}
