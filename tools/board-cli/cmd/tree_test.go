package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aagrigore/task-board/internal/board"
)

func TestTreeCommand(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	err := runTree(treeCmd, []string{})
	if err != nil {
		t.Fatalf("runTree: %v", err)
	}
}

func TestTreeCommandJSON(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	jsonOutput = true
	defer func() { jsonOutput = false }()

	err := runTree(treeCmd, []string{})
	if err != nil {
		t.Fatalf("runTree: %v", err)
	}
}

func TestTreeStructure(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	// Build the tree directly
	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("loadBoard: %v", err)
	}

	epics := b.FindByType(board.EpicType)
	tree := buildTree(b, epics)

	// Should have 2 epics
	if len(tree) != 2 {
		t.Errorf("expected 2 epics, got %d", len(tree))
	}

	// Find EPIC-260101-aaaaaa_recording
	var epic1 *TreeNode
	for _, node := range tree {
		if node.ID == testEpic1ID {
			epic1 = node
			break
		}
	}
	if epic1 == nil {
		t.Fatalf("EPIC-01 not found in tree")
	}

	// EPIC-01 should have 2 stories (STORY-01 and STORY-02)
	if len(epic1.Children) != 2 {
		t.Errorf("EPIC-01 should have 2 stories, got %d", len(epic1.Children))
	}

	// Find STORY-01 in epic1's children
	var story1 *TreeNode
	for _, node := range epic1.Children {
		if node.ID == testStory1ID {
			story1 = node
			break
		}
	}
	if story1 == nil {
		t.Fatalf("STORY-01 not found in EPIC-01's children")
	}

	// STORY-01 should have 4 children (3 tasks + 1 bug)
	if len(story1.Children) != 4 {
		t.Errorf("STORY-01 should have 4 children, got %d", len(story1.Children))
	}
}

func TestTreeJSONFormat(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	jsonOutput = true
	defer func() { jsonOutput = false }()

	// Run tree command with JSON output
	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("loadBoard: %v", err)
	}

	epics := b.FindByType(board.EpicType)
	tree := buildTree(b, epics)
	response := TreeResponse{Tree: tree}

	// Marshal to JSON and verify structure
	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}

	// Parse back to verify
	var parsed TreeResponse
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}

	if len(parsed.Tree) != 2 {
		t.Errorf("expected 2 epics in JSON, got %d", len(parsed.Tree))
	}

	// Verify field names in JSON output
	jsonStr := string(jsonBytes)
	requiredFields := []string{"id", "type", "name", "status", "updatedAt", "children"}
	for _, field := range requiredFields {
		if !strings.Contains(jsonStr, `"`+field+`"`) {
			t.Errorf("JSON output missing field: %s", field)
		}
	}
}

func TestTreeEpicFilter(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	treeEpic = testEpic1ID
	defer func() { treeEpic = "" }()

	b, err := board.Load(bd)
	if err != nil {
		t.Fatalf("loadBoard: %v", err)
	}

	// Filter to single epic
	epic := b.FindByID(testEpic1ID)
	if epic == nil {
		t.Fatalf("epic not found")
	}

	tree := buildTree(b, []*board.Element{epic})

	// Should have only 1 epic
	if len(tree) != 1 {
		t.Errorf("expected 1 epic with filter, got %d", len(tree))
	}

	if tree[0].ID != testEpic1ID {
		t.Errorf("expected %s, got %s", testEpic1ID, tree[0].ID)
	}
}

func TestTreeNotFoundEpic(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	treeEpic = "EPIC-999"
	defer func() { treeEpic = "" }()

	err := runTree(treeCmd, []string{})
	if err == nil {
		t.Fatal("expected error for missing epic")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want 'not found'", err.Error())
	}
}
