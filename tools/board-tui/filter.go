package main

import (
	"regexp"
	"strings"
)

// FilterExpression represents a parsed filter with logical operators
type FilterExpression interface {
	Match(node *TreeNode) bool
}

// TermExpr matches a simple term against node fields
type TermExpr struct {
	term    string
	negate  bool
}

func (e *TermExpr) Match(node *TreeNode) bool {
	term := strings.ToLower(e.term)
	searchable := strings.ToLower(node.ID + " " + node.Name + " " + node.Status + " " + node.Type + " " + node.GetAssignee())
	result := strings.Contains(searchable, term)
	if e.negate {
		return !result
	}
	return result
}

// AndExpr matches if all sub-expressions match
type AndExpr struct {
	exprs []FilterExpression
}

func (e *AndExpr) Match(node *TreeNode) bool {
	for _, expr := range e.exprs {
		if !expr.Match(node) {
			return false
		}
	}
	return true
}

// OrExpr matches if any sub-expression matches
type OrExpr struct {
	exprs []FilterExpression
}

func (e *OrExpr) Match(node *TreeNode) bool {
	for _, expr := range e.exprs {
		if expr.Match(node) {
			return true
		}
	}
	return false
}

// ParseFilter parses a filter string with && || ! and parentheses
// Examples:
//   - "auth" - simple term
//   - "!done" - negation
//   - "auth && done" - AND
//   - "auth || login" - OR
//   - "(auth || login) && done" - grouped
func ParseFilter(input string) FilterExpression {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}
	return parseOr(input)
}

// parseOr handles || operator (lowest precedence)
func parseOr(input string) FilterExpression {
	parts := splitByOperator(input, "||")
	if len(parts) == 1 {
		return parseAnd(parts[0])
	}

	exprs := make([]FilterExpression, 0, len(parts))
	for _, part := range parts {
		if expr := parseAnd(part); expr != nil {
			exprs = append(exprs, expr)
		}
	}
	if len(exprs) == 0 {
		return nil
	}
	if len(exprs) == 1 {
		return exprs[0]
	}
	return &OrExpr{exprs: exprs}
}

// parseAnd handles && operator (higher precedence than ||)
func parseAnd(input string) FilterExpression {
	parts := splitByOperator(input, "&&")
	if len(parts) == 1 {
		return parseTerm(parts[0])
	}

	exprs := make([]FilterExpression, 0, len(parts))
	for _, part := range parts {
		if expr := parseTerm(part); expr != nil {
			exprs = append(exprs, expr)
		}
	}
	if len(exprs) == 0 {
		return nil
	}
	if len(exprs) == 1 {
		return exprs[0]
	}
	return &AndExpr{exprs: exprs}
}

// parseTerm handles individual terms, negation, and parentheses
func parseTerm(input string) FilterExpression {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	// Handle negation
	if strings.HasPrefix(input, "!") {
		inner := parseTerm(input[1:])
		if inner == nil {
			return nil
		}
		// If it's already a TermExpr, just flip negate
		if term, ok := inner.(*TermExpr); ok {
			term.negate = !term.negate
			return term
		}
		// Otherwise wrap in a negating term (simplified)
		return &TermExpr{term: input[1:], negate: true}
	}

	// Handle parentheses
	if strings.HasPrefix(input, "(") && strings.HasSuffix(input, ")") {
		// Find matching closing paren
		depth := 0
		for i, c := range input {
			if c == '(' {
				depth++
			} else if c == ')' {
				depth--
				if depth == 0 && i == len(input)-1 {
					return parseOr(input[1 : len(input)-1])
				}
			}
		}
	}

	// Simple term
	return &TermExpr{term: input, negate: false}
}

// splitByOperator splits input by operator, respecting parentheses
func splitByOperator(input string, op string) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for i := 0; i < len(input); i++ {
		c := input[i]

		if c == '(' {
			depth++
			current.WriteByte(c)
		} else if c == ')' {
			depth--
			current.WriteByte(c)
		} else if depth == 0 && i+len(op) <= len(input) && input[i:i+len(op)] == op {
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
			i += len(op) - 1
		} else {
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}

	return parts
}

// FilterTree filters the tree and returns matching nodes with their ancestors expanded
// Returns a new tree with only matching nodes visible
func FilterTree(roots []*TreeNode, filter FilterExpression) []*TreeNode {
	if filter == nil {
		return roots
	}

	// Find all matching nodes
	matchingIDs := make(map[string]bool)
	for _, root := range roots {
		findMatches(root, filter, matchingIDs)
	}

	if len(matchingIDs) == 0 {
		return nil
	}

	// Mark ancestors of matching nodes
	ancestorIDs := make(map[string]bool)
	for _, root := range roots {
		markAncestors(root, matchingIDs, ancestorIDs)
	}

	// Build filtered tree
	return buildFilteredTree(roots, matchingIDs, ancestorIDs)
}

// findMatches recursively finds all nodes matching the filter
func findMatches(node *TreeNode, filter FilterExpression, matches map[string]bool) {
	if filter.Match(node) {
		matches[node.ID] = true
	}
	for _, child := range node.Children {
		findMatches(child, filter, matches)
	}
}

// markAncestors marks all ancestors of matching nodes
func markAncestors(node *TreeNode, matchingIDs, ancestorIDs map[string]bool) bool {
	hasMatchingDescendant := matchingIDs[node.ID]

	for _, child := range node.Children {
		if markAncestors(child, matchingIDs, ancestorIDs) {
			hasMatchingDescendant = true
		}
	}

	if hasMatchingDescendant && !matchingIDs[node.ID] {
		ancestorIDs[node.ID] = true
	}

	return hasMatchingDescendant
}

// buildFilteredTree builds a new tree with only matching nodes and their ancestors
func buildFilteredTree(roots []*TreeNode, matchingIDs, ancestorIDs map[string]bool) []*TreeNode {
	var result []*TreeNode

	for _, root := range roots {
		if filtered := filterNode(root, matchingIDs, ancestorIDs); filtered != nil {
			result = append(result, filtered)
		}
	}

	return result
}

// filterNode recursively filters a node
func filterNode(node *TreeNode, matchingIDs, ancestorIDs map[string]bool) *TreeNode {
	isMatch := matchingIDs[node.ID]
	isAncestor := ancestorIDs[node.ID]

	if !isMatch && !isAncestor {
		return nil
	}

	// Create a copy of the node
	filtered := &TreeNode{
		ID:        node.ID,
		Type:      node.Type,
		Name:      node.Name,
		Status:    node.Status,
		Assignee:  node.Assignee,
		UpdatedAt: node.UpdatedAt,
		Expanded:  true, // Auto-expand to show matches
		Depth:     node.Depth,
		Parent:    node.Parent,
	}

	// Filter children
	for _, child := range node.Children {
		if filteredChild := filterNode(child, matchingIDs, ancestorIDs); filteredChild != nil {
			filteredChild.Parent = filtered
			filtered.Children = append(filtered.Children, filteredChild)
		}
	}

	return filtered
}

// HighlightMatches wraps matching terms in ANSI highlight codes
func HighlightMatches(text, term string) string {
	if term == "" {
		return text
	}

	// Case-insensitive replace with highlight
	re, err := regexp.Compile("(?i)(" + regexp.QuoteMeta(term) + ")")
	if err != nil {
		return text
	}

	// Yellow background highlight
	return re.ReplaceAllString(text, "\033[43m\033[30m$1\033[0m")
}
