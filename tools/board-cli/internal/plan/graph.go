package plan

import (
	"fmt"

	"github.com/aagrigore/task-board/internal/board"
)

// Graph represents a dependency graph for a set of elements.
type Graph struct {
	Nodes    []*board.Element
	AdjList  map[string][]string // ID -> list of IDs this element blocks
	InDegree map[string]int      // ID -> number of elements blocking this one
}

// BuildGraph creates a dependency graph from a set of elements.
// Only considers dependencies between elements within the given set.
func BuildGraph(elements []*board.Element) *Graph {
	g := &Graph{
		Nodes:    elements,
		AdjList:  make(map[string][]string),
		InDegree: make(map[string]int),
	}

	// Build a set of IDs in scope
	inScope := make(map[string]bool)
	for _, e := range elements {
		inScope[e.ID()] = true
		g.InDegree[e.ID()] = 0
	}

	// Build adjacency list from blocked-by relationships
	for _, e := range elements {
		for _, blockerID := range e.BlockedBy {
			if !inScope[blockerID] {
				continue // ignore out-of-scope dependencies
			}
			g.AdjList[blockerID] = append(g.AdjList[blockerID], e.ID())
			g.InDegree[e.ID()]++
		}
	}

	return g
}

// Phase represents a group of elements that can be executed in parallel.
type Phase struct {
	Number   int
	Elements []*board.Element
}

// Plan is the result of planning: phases + metadata.
type Plan struct {
	Phases       []Phase
	CriticalPath []*board.Element
	HasCycle     bool
	CycleNodes   []string // IDs involved in cycle
}

// BuildPlan runs topological sort and groups elements into phases.
func BuildPlan(elements []*board.Element) *Plan {
	g := BuildGraph(elements)
	plan := &Plan{}

	// Kahn's algorithm with level tracking for phases
	elemByID := make(map[string]*board.Element)
	for _, e := range elements {
		elemByID[e.ID()] = e
	}

	// Copy in-degree map (we'll modify it)
	inDeg := make(map[string]int)
	for id, deg := range g.InDegree {
		inDeg[id] = deg
	}

	// Find initial queue (in-degree 0)
	var queue []string
	for _, e := range elements {
		if inDeg[e.ID()] == 0 {
			queue = append(queue, e.ID())
		}
	}

	processed := 0
	phaseNum := 1

	for len(queue) > 0 {
		phase := Phase{Number: phaseNum}
		var nextQueue []string

		for _, id := range queue {
			phase.Elements = append(phase.Elements, elemByID[id])
			processed++

			for _, neighborID := range g.AdjList[id] {
				inDeg[neighborID]--
				if inDeg[neighborID] == 0 {
					nextQueue = append(nextQueue, neighborID)
				}
			}
		}

		plan.Phases = append(plan.Phases, phase)
		queue = nextQueue
		phaseNum++
	}

	// Cycle detection: if not all nodes processed, there's a cycle
	if processed < len(elements) {
		plan.HasCycle = true
		for id, deg := range inDeg {
			if deg > 0 {
				plan.CycleNodes = append(plan.CycleNodes, id)
			}
		}
		return plan
	}

	// Compute critical path (longest path in DAG)
	plan.CriticalPath = computeCriticalPath(elements, g, elemByID)

	return plan
}

// computeCriticalPath finds the longest path in the DAG.
func computeCriticalPath(elements []*board.Element, g *Graph, elemByID map[string]*board.Element) []*board.Element {
	// Longest distance from any source
	dist := make(map[string]int)
	prev := make(map[string]string)

	for _, e := range elements {
		dist[e.ID()] = 0
		prev[e.ID()] = ""
	}

	// Process in topological order (phase by phase)
	inDeg := make(map[string]int)
	for id, deg := range g.InDegree {
		inDeg[id] = deg
	}

	var queue []string
	for _, e := range elements {
		if inDeg[e.ID()] == 0 {
			queue = append(queue, e.ID())
		}
	}

	for len(queue) > 0 {
		var nextQueue []string
		for _, id := range queue {
			for _, neighborID := range g.AdjList[id] {
				if dist[id]+1 > dist[neighborID] {
					dist[neighborID] = dist[id] + 1
					prev[neighborID] = id
				}
				inDeg[neighborID]--
				if inDeg[neighborID] == 0 {
					nextQueue = append(nextQueue, neighborID)
				}
			}
		}
		queue = nextQueue
	}

	// Find the node with max distance
	maxDist := 0
	maxID := ""
	for _, e := range elements {
		if dist[e.ID()] > maxDist {
			maxDist = dist[e.ID()]
			maxID = e.ID()
		}
	}

	if maxID == "" {
		return nil
	}

	// Trace back
	var path []*board.Element
	current := maxID
	for current != "" {
		path = append([]*board.Element{elemByID[current]}, path...)
		current = prev[current]
	}

	return path
}

// ScopeElements returns the direct children for the given scope.
// If scopeID is empty, returns all epics (project-level).
// If scopeID is an epic, returns its stories.
// If scopeID is a story, returns its tasks/bugs.
func ScopeElements(b *board.Board, scopeID string) ([]*board.Element, error) {
	if scopeID == "" {
		return b.FindByType(board.EpicType), nil
	}

	elem := b.FindByID(scopeID)
	if elem == nil {
		return nil, fmt.Errorf("element %s not found", scopeID)
	}

	children := b.Children(elem)
	if len(children) == 0 {
		return nil, fmt.Errorf("%s has no children", scopeID)
	}

	return children, nil
}

// AllDescendants returns all elements recursively from the given scope downward.
// If scopeID is empty, returns everything on the board.
// Results are returned grouped: parent first, then its children depth-first.
func AllDescendants(b *board.Board, scopeID string) ([]*board.Element, error) {
	if scopeID == "" {
		return b.Elements, nil
	}

	root := b.FindByID(scopeID)
	if root == nil {
		return nil, fmt.Errorf("element %s not found", scopeID)
	}

	var result []*board.Element
	result = append(result, root)
	collectDescendants(b, root, &result)
	return result, nil
}

func collectDescendants(b *board.Board, parent *board.Element, result *[]*board.Element) {
	children := b.Children(parent)
	for _, c := range children {
		*result = append(*result, c)
		collectDescendants(b, c, result)
	}
}
