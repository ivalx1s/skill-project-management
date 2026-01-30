package plan

import (
	"fmt"
	"strings"

	"github.com/aagrigore/task-board/internal/board"
)

// Problem represents an issue found in the dependency graph.
type Problem struct {
	Severity string   // "error" or "warning" or "info"
	Type     string   // "cycle", "critical-path"
	Message  string
	Elements []string // IDs involved
}

// DetectProblems analyzes a plan and its source elements, returning all problems found.
func DetectProblems(p *Plan, elements []*board.Element) []Problem {
	var problems []Problem

	if p.HasCycle {
		problems = append(problems, detectCycleProblem(p, elements))
	}

	if len(p.CriticalPath) > 0 {
		problems = append(problems, Problem{
			Severity: "info",
			Type:     "critical-path",
			Message:  FormatCriticalPath(p.CriticalPath),
			Elements: criticalPathIDs(p.CriticalPath),
		})
	}

	return problems
}

// detectCycleProblem reconstructs the cycle path from CycleNodes and returns a Problem.
func detectCycleProblem(p *Plan, elements []*board.Element) Problem {
	cycleSet := make(map[string]bool, len(p.CycleNodes))
	for _, id := range p.CycleNodes {
		cycleSet[id] = true
	}

	// Build blockedBy adjacency restricted to cycle nodes.
	blockedByMap := make(map[string][]string)
	for _, e := range elements {
		if !cycleSet[e.ID()] {
			continue
		}
		for _, dep := range e.BlockedBy {
			if cycleSet[dep] {
				blockedByMap[e.ID()] = append(blockedByMap[e.ID()], dep)
			}
		}
	}

	// Try to trace a cycle starting from the first CycleNode using DFS.
	cyclePath := traceCycle(p.CycleNodes[0], blockedByMap)

	var msg string
	if len(cyclePath) > 0 {
		msg = "Dependency cycle detected: " + strings.Join(cyclePath, " -> ")
	} else {
		msg = "Dependency cycle detected involving: " + strings.Join(p.CycleNodes, ", ")
	}

	return Problem{
		Severity: "error",
		Type:     "cycle",
		Message:  msg,
		Elements: p.CycleNodes,
	}
}

// traceCycle attempts to find a cycle path starting from startID using DFS
// through the blockedBy relationships. Returns the cycle as a slice of IDs
// ending with the repeated start node (e.g. [A, B, C, A]).
func traceCycle(startID string, blockedByMap map[string][]string) []string {
	visited := make(map[string]bool)
	var path []string

	var dfs func(id string) bool
	dfs = func(id string) bool {
		if visited[id] {
			if id == startID {
				path = append(path, id)
				return true
			}
			return false
		}
		visited[id] = true
		path = append(path, id)

		for _, dep := range blockedByMap[id] {
			if dfs(dep) {
				return true
			}
		}

		// Backtrack
		path = path[:len(path)-1]
		visited[id] = false
		return false
	}

	visited[startID] = true
	path = append(path, startID)

	for _, dep := range blockedByMap[startID] {
		if dfs(dep) {
			return path
		}
	}

	return nil
}

// FormatCriticalPath returns a human-readable string for the critical path.
// Example: "TASK-01 -> TASK-02 -> TASK-06 (3 phases)"
func FormatCriticalPath(path []*board.Element) string {
	if len(path) == 0 {
		return ""
	}

	ids := make([]string, len(path))
	for i, e := range path {
		ids[i] = e.ID()
	}

	return fmt.Sprintf("%s (%d elements)", strings.Join(ids, " -> "), len(path))
}

// criticalPathIDs extracts IDs from a critical path.
func criticalPathIDs(path []*board.Element) []string {
	ids := make([]string, len(path))
	for i, e := range path {
		ids[i] = e.ID()
	}
	return ids
}
