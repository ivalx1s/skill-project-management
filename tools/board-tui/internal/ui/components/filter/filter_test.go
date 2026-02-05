package filter

import "testing"

func TestFilterInitial(t *testing.T) {
	s := Initial()
	if s.Active {
		t.Error("Initial filter should not be active")
	}
	if s.Query != "" {
		t.Error("Initial filter query should be empty")
	}
}

func TestFilterActivate(t *testing.T) {
	s := Initial()
	expanded := []string{"EPIC-01", "STORY-01"}

	s = Reduce(s, Activate{PreFilterExpanded: expanded})

	if !s.Active {
		t.Error("Filter should be active after Activate")
	}
	if len(s.PreFilterExpanded) != 2 {
		t.Errorf("Expected 2 pre-filter expanded IDs, got %d", len(s.PreFilterExpanded))
	}
}

func TestFilterDeactivate(t *testing.T) {
	s := State{
		Active:            true,
		Query:             "test",
		PreFilterExpanded: []string{"EPIC-01"},
	}

	s = Reduce(s, Deactivate{})

	if s.Active {
		t.Error("Filter should not be active after Deactivate")
	}
	if s.Query != "" {
		t.Error("Query should be cleared after Deactivate")
	}
	if s.PreFilterExpanded != nil {
		t.Error("PreFilterExpanded should be nil after Deactivate")
	}
}

func TestFilterSetQuery(t *testing.T) {
	s := State{Active: true}

	s = Reduce(s, SetQuery{Query: "test query"})

	if s.Query != "test query" {
		t.Errorf("Expected query 'test query', got '%s'", s.Query)
	}
}

func TestFilterClear(t *testing.T) {
	s := State{Active: true, Query: "something"}

	s = Reduce(s, Clear{})

	if s.Query != "" {
		t.Error("Clear should set query to empty string")
	}
	if !s.Active {
		t.Error("Clear should not deactivate filter")
	}
}

func TestFilterIsActive(t *testing.T) {
	s := State{Active: false}
	if s.IsActive() {
		t.Error("IsActive should return false when not active")
	}

	s.Active = true
	if !s.IsActive() {
		t.Error("IsActive should return true when active")
	}
}
