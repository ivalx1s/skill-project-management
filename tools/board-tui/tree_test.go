package main

import (
	"testing"
)

func TestParseTreeJSON(t *testing.T) {
	jsonData := []byte(`{
		"tree": [
			{
				"id": "EPIC-01",
				"type": "epic",
				"name": "test-epic",
				"status": "backlog",
				"assignee": null,
				"updatedAt": "2026-02-05T13:00:00Z",
				"children": [
					{
						"id": "STORY-01",
						"type": "story",
						"name": "test-story",
						"status": "development",
						"assignee": "agent-1",
						"updatedAt": "2026-02-05T14:00:00Z",
						"children": [
							{
								"id": "TASK-01",
								"type": "task",
								"name": "test-task",
								"status": "done",
								"assignee": null,
								"updatedAt": "",
								"children": []
							}
						]
					}
				]
			}
		]
	}`)

	tree, err := ParseTreeJSON(jsonData)
	if err != nil {
		t.Fatalf("ParseTreeJSON failed: %v", err)
	}

	if len(tree) != 1 {
		t.Fatalf("Expected 1 root node, got %d", len(tree))
	}

	epic := tree[0]
	if epic.ID != "EPIC-01" {
		t.Errorf("Expected epic ID EPIC-01, got %s", epic.ID)
	}
	if epic.Type != "epic" {
		t.Errorf("Expected type epic, got %s", epic.Type)
	}
	if epic.Name != "test-epic" {
		t.Errorf("Expected name test-epic, got %s", epic.Name)
	}
	// Nodes are collapsed by default, expanded state is managed by config
	if epic.Expanded {
		t.Error("Epic should be collapsed by default")
	}

	if len(epic.Children) != 1 {
		t.Fatalf("Expected 1 story, got %d", len(epic.Children))
	}

	story := epic.Children[0]
	if story.ID != "STORY-01" {
		t.Errorf("Expected story ID STORY-01, got %s", story.ID)
	}
	if story.Parent != epic {
		t.Error("Story parent should be epic")
	}
	if story.Expanded {
		t.Error("Story should not be expanded by default")
	}

	if len(story.Children) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(story.Children))
	}

	task := story.Children[0]
	if task.ID != "TASK-01" {
		t.Errorf("Expected task ID TASK-01, got %s", task.ID)
	}
	if task.Parent != story {
		t.Error("Task parent should be story")
	}
}

func TestTreeNodeToggle(t *testing.T) {
	node := &TreeNode{Expanded: false}

	node.Toggle()
	if !node.Expanded {
		t.Error("Toggle should have set Expanded to true")
	}

	node.Toggle()
	if node.Expanded {
		t.Error("Toggle should have set Expanded to false")
	}
}

func TestTreeNodeExpandAll(t *testing.T) {
	tree := GetDemoTree()
	epic := tree[0]

	// Collapse everything first
	epic.CollapseAll()
	if epic.Expanded {
		t.Error("CollapseAll should collapse epic")
	}

	// Expand all
	epic.ExpandAll()
	if !epic.Expanded {
		t.Error("ExpandAll should expand epic")
	}

	story := epic.Children[0]
	if !story.Expanded {
		t.Error("ExpandAll should expand story")
	}
}

func TestFlattenTree(t *testing.T) {
	tree := GetDemoTree()
	epic := tree[0]

	// With epic expanded, story collapsed
	epic.Expanded = true
	epic.Children[0].Expanded = false

	flat := FlattenTree(tree)

	// Should see epic + story only (story's children hidden)
	if len(flat) != 2 {
		t.Errorf("Expected 2 visible nodes, got %d", len(flat))
	}

	if flat[0].Node.ID != "EPIC-001" {
		t.Errorf("Expected first node EPIC-001, got %s", flat[0].Node.ID)
	}
	if flat[1].Node.ID != "STORY-001" {
		t.Errorf("Expected second node STORY-001, got %s", flat[1].Node.ID)
	}

	// Now expand story
	epic.Children[0].Expanded = true
	flat = FlattenTree(tree)

	// Should see epic + story + 3 tasks
	if len(flat) != 5 {
		t.Errorf("Expected 5 visible nodes, got %d", len(flat))
	}
}

func TestFlattenTreePrefixes(t *testing.T) {
	tree := GetDemoTree()
	epic := tree[0]
	epic.Expanded = true
	epic.Children[0].Expanded = true

	flat := FlattenTree(tree)

	// Check tree prefixes
	if flat[0].TreePrefix != "" {
		t.Errorf("Root should have empty prefix, got '%s'", flat[0].TreePrefix)
	}

	// Story should have tree prefix
	if flat[1].Depth != 1 {
		t.Errorf("Story should have depth 1, got %d", flat[1].Depth)
	}
}

func TestFindNodeByID(t *testing.T) {
	tree := GetDemoTree()

	// Find epic
	found := FindNodeByID(tree, "EPIC-001")
	if found == nil {
		t.Fatal("Should find EPIC-001")
	}
	if found.ID != "EPIC-001" {
		t.Errorf("Found wrong node: %s", found.ID)
	}

	// Find nested task
	found = FindNodeByID(tree, "TASK-002")
	if found == nil {
		t.Fatal("Should find TASK-002")
	}
	if found.ID != "TASK-002" {
		t.Errorf("Found wrong node: %s", found.ID)
	}

	// Not found
	found = FindNodeByID(tree, "NONEXISTENT")
	if found != nil {
		t.Error("Should not find NONEXISTENT")
	}
}

func TestGetVisibleCount(t *testing.T) {
	tree := GetDemoTree()
	epic := tree[0]

	// Collapse all
	epic.CollapseAll()
	count := GetVisibleCount(tree)
	if count != 1 {
		t.Errorf("With everything collapsed, expected 1 visible, got %d", count)
	}

	// Expand epic only
	epic.Expanded = true
	count = GetVisibleCount(tree)
	if count != 2 { // epic + story
		t.Errorf("With epic expanded, expected 2 visible, got %d", count)
	}

	// Expand all
	epic.ExpandAll()
	count = GetVisibleCount(tree)
	if count != 5 { // epic + story + 3 tasks
		t.Errorf("With all expanded, expected 5 visible, got %d", count)
	}
}

func TestExpandToNode(t *testing.T) {
	tree := GetDemoTree()
	epic := tree[0]

	// Collapse everything
	epic.CollapseAll()

	// Find a deeply nested task
	task := FindNodeByID(tree, "TASK-001")
	if task == nil {
		t.Fatal("Should find TASK-001")
	}

	// Expand path to task
	ExpandToNode(task)

	// Check that ancestors are expanded
	if !epic.Expanded {
		t.Error("Epic should be expanded after ExpandToNode")
	}
	if !epic.Children[0].Expanded {
		t.Error("Story should be expanded after ExpandToNode")
	}
}

func TestHasChildren(t *testing.T) {
	epic := &TreeNode{
		Children: []*TreeNode{{}},
	}
	if !epic.HasChildren() {
		t.Error("Node with children should return true for HasChildren")
	}

	task := &TreeNode{
		Children: []*TreeNode{},
	}
	if task.HasChildren() {
		t.Error("Node without children should return false for HasChildren")
	}
}

func TestGetAssignee(t *testing.T) {
	// Nil assignee
	node := &TreeNode{Assignee: nil}
	if node.GetAssignee() != "" {
		t.Error("Nil assignee should return empty string")
	}

	// With assignee
	assignee := "agent-1"
	node.Assignee = &assignee
	if node.GetAssignee() != "agent-1" {
		t.Errorf("Expected agent-1, got %s", node.GetAssignee())
	}
}
