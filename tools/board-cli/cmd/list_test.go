package cmd

import (
	"testing"
)

func TestListEpics(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	listStatus = ""
	listEpic = ""
	listStory = ""

	err := runList(listCmd, []string{"epics"})
	if err != nil {
		t.Fatalf("runList epics: %v", err)
	}
}

func TestListTasksFilterStatus(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	listStatus = "to-dev"
	listEpic = ""
	listStory = ""

	err := runList(listCmd, []string{"tasks"})
	if err != nil {
		t.Fatalf("runList tasks --status to-dev: %v", err)
	}
}

func TestListTasksFilterStory(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	listStatus = ""
	listEpic = ""
	listStory = testStory1ID

	err := runList(listCmd, []string{"tasks"})
	if err != nil {
		t.Fatalf("runList tasks --story STORY-01: %v", err)
	}
}

func TestListStoriesFilterEpic(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	listStatus = ""
	listEpic = testEpic1ID
	listStory = ""

	err := runList(listCmd, []string{"stories"})
	if err != nil {
		t.Fatalf("runList stories --epic EPIC-01: %v", err)
	}
}

func TestListBugs(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	listStatus = ""
	listEpic = ""
	listStory = ""

	err := runList(listCmd, []string{"bugs"})
	if err != nil {
		t.Fatalf("runList bugs: %v", err)
	}
}

func TestListInvalidType(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	err := runList(listCmd, []string{"widgets"})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestListEmpty(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd
	listStatus = "done"
	listEpic = ""
	listStory = ""

	// No done elements in test board
	err := runList(listCmd, []string{"tasks"})
	if err != nil {
		t.Fatalf("runList: %v", err)
	}
}
