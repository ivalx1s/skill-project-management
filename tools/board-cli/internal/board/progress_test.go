package board

import (
	"testing"
)

func TestParseProgress(t *testing.T) {
	content := `## Status
development

## Blocked By
- TASK-12
- TASK-15

## Blocks
- TASK-20
- TASK-21

## Checklist
- [x] Write unit tests
- [ ] Write integration tests
- [x] Code review

## Notes
Some notes here
`
	pd, err := ParseProgress(content)
	if err != nil {
		t.Fatal(err)
	}

	if pd.Status != StatusDevelopment {
		t.Errorf("Status = %v, want development", pd.Status)
	}

	if len(pd.BlockedBy) != 2 {
		t.Fatalf("BlockedBy len = %d, want 2", len(pd.BlockedBy))
	}
	if pd.BlockedBy[0] != "TASK-12" || pd.BlockedBy[1] != "TASK-15" {
		t.Errorf("BlockedBy = %v", pd.BlockedBy)
	}

	if len(pd.Blocks) != 2 {
		t.Fatalf("Blocks len = %d, want 2", len(pd.Blocks))
	}
	if pd.Blocks[0] != "TASK-20" || pd.Blocks[1] != "TASK-21" {
		t.Errorf("Blocks = %v", pd.Blocks)
	}

	if len(pd.Checklist) != 3 {
		t.Fatalf("Checklist len = %d, want 3", len(pd.Checklist))
	}
	if !pd.Checklist[0].Checked || pd.Checklist[0].Text != "Write unit tests" {
		t.Errorf("Checklist[0] = %+v", pd.Checklist[0])
	}
	if pd.Checklist[1].Checked || pd.Checklist[1].Text != "Write integration tests" {
		t.Errorf("Checklist[1] = %+v", pd.Checklist[1])
	}

	if pd.Notes != "Some notes here" {
		t.Errorf("Notes = %q", pd.Notes)
	}
}

func TestParseProgressEmpty(t *testing.T) {
	content := `## Status
to-dev

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
(empty)

## Notes
`
	pd, err := ParseProgress(content)
	if err != nil {
		t.Fatal(err)
	}

	if pd.Status != StatusToDev {
		t.Errorf("Status = %v, want to-dev", pd.Status)
	}
	if len(pd.BlockedBy) != 0 {
		t.Errorf("BlockedBy = %v, want empty", pd.BlockedBy)
	}
	if len(pd.Blocks) != 0 {
		t.Errorf("Blocks = %v, want empty", pd.Blocks)
	}
	if len(pd.Checklist) != 0 {
		t.Errorf("Checklist = %v, want empty", pd.Checklist)
	}
}

func TestWriteProgress(t *testing.T) {
	pd := &ProgressData{
		Status:    StatusDevelopment,
		BlockedBy: []string{"TASK-12"},
		Blocks:    []string{"TASK-20"},
		Checklist: []ChecklistItem{
			{Text: "Step 1", Checked: true},
			{Text: "Step 2", Checked: false},
		},
		Notes: "WIP",
	}

	content := WriteProgress(pd)

	// Re-parse to verify roundtrip
	pd2, err := ParseProgress(content)
	if err != nil {
		t.Fatal(err)
	}

	if pd2.Status != StatusDevelopment {
		t.Errorf("Status = %v", pd2.Status)
	}
	if len(pd2.BlockedBy) != 1 || pd2.BlockedBy[0] != "TASK-12" {
		t.Errorf("BlockedBy = %v", pd2.BlockedBy)
	}
	if len(pd2.Blocks) != 1 || pd2.Blocks[0] != "TASK-20" {
		t.Errorf("Blocks = %v", pd2.Blocks)
	}
	if len(pd2.Checklist) != 2 {
		t.Errorf("Checklist len = %d", len(pd2.Checklist))
	}
}
