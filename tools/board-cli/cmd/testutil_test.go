package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aagrigore/task-board/internal/board"
)

// setupTestBoard creates a temp board with known structure for cmd tests.
// Returns the board directory path.
func setupTestBoard(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bd := filepath.Join(dir, ".task-board")
	os.MkdirAll(bd, 0755)

	board.WriteCounters(bd, &board.Counters{Epic: 2, Story: 3, Task: 4, Bug: 1})

	// EPIC-01_recording
	epicDir := filepath.Join(bd, "EPIC-01_recording")
	os.MkdirAll(epicDir, 0755)
	os.WriteFile(filepath.Join(epicDir, "README.md"),
		[]byte("# EPIC-01: Recording\n\n## Description\nRecording epic\n\n## Scope\nrecording-lib\n\n## Acceptance Criteria\n- Works\n"), 0644)
	os.WriteFile(filepath.Join(epicDir, "progress.md"),
		[]byte("## Status\nprogress\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// EPIC-02_storage
	epic2Dir := filepath.Join(bd, "EPIC-02_storage")
	os.MkdirAll(epic2Dir, 0755)
	os.WriteFile(filepath.Join(epic2Dir, "README.md"),
		[]byte("# EPIC-02: Storage\n\n## Description\nStorage epic\n\n## Scope\nstorage-lib\n\n## Acceptance Criteria\n- Works\n"), 0644)
	os.WriteFile(filepath.Join(epic2Dir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// STORY-03_migration inside EPIC-02 (for cross-epic tests)
	story3Dir := filepath.Join(epic2Dir, "STORY-03_migration")
	os.MkdirAll(story3Dir, 0755)
	os.WriteFile(filepath.Join(story3Dir, "README.md"),
		[]byte("# STORY-03: Migration\n\n## Description\nDB migration\n\n## Scope\nstorage-lib\n\n## Acceptance Criteria\n- Migrated\n"), 0644)
	os.WriteFile(filepath.Join(story3Dir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-04_schema inside STORY-03 (for cross-story/epic tests)
	task4Dir := filepath.Join(story3Dir, "TASK-04_schema")
	os.MkdirAll(task4Dir, 0755)
	os.WriteFile(filepath.Join(task4Dir, "README.md"),
		[]byte("# TASK-04: Schema\n\n## Description\nDefine schema\n\n## Scope\nstorage-lib\n\n## Acceptance Criteria\n- Schema defined\n"), 0644)
	os.WriteFile(filepath.Join(task4Dir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// STORY-01_audio-capture inside EPIC-01
	storyDir := filepath.Join(epicDir, "STORY-01_audio-capture")
	os.MkdirAll(storyDir, 0755)
	os.WriteFile(filepath.Join(storyDir, "README.md"),
		[]byte("# STORY-01: Audio Capture\n\n## Description\nCapture audio\n\n## Scope\nrecording-lib\n\n## Acceptance Criteria\n- Records\n"), 0644)
	os.WriteFile(filepath.Join(storyDir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// STORY-02_amplitude inside EPIC-01
	story2Dir := filepath.Join(epicDir, "STORY-02_amplitude")
	os.MkdirAll(story2Dir, 0755)
	os.WriteFile(filepath.Join(story2Dir, "README.md"),
		[]byte("# STORY-02: Amplitude\n\n## Description\nAmplitude analysis\n\n## Scope\nrecording-lib\n\n## Acceptance Criteria\n- Analyzes\n"), 0644)
	os.WriteFile(filepath.Join(story2Dir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-01_interface inside STORY-01
	taskDir := filepath.Join(storyDir, "TASK-01_interface")
	os.MkdirAll(taskDir, 0755)
	os.WriteFile(filepath.Join(taskDir, "README.md"),
		[]byte("# TASK-01: Interface\n\n## Description\nDefine interface\n\n## Scope\nrecording-lib\n\n## Acceptance Criteria\n- Interface defined\n"), 0644)
	os.WriteFile(filepath.Join(taskDir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Blocks\n- TASK-02\n\n## Checklist\n- [ ] Step 1\n- [x] Step 2\n\n## Notes\nStarted work\n"), 0644)

	// TASK-02_impl inside STORY-01, blocked by TASK-01
	task2Dir := filepath.Join(storyDir, "TASK-02_impl")
	os.MkdirAll(task2Dir, 0755)
	os.WriteFile(filepath.Join(task2Dir, "README.md"),
		[]byte("# TASK-02: Implementation\n\n## Description\nImplement it\n\n## Scope\nrecording-lib\n\n## Acceptance Criteria\n- Implemented\n"), 0644)
	os.WriteFile(filepath.Join(task2Dir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- TASK-01\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-03_tests inside STORY-01
	task3Dir := filepath.Join(storyDir, "TASK-03_tests")
	os.MkdirAll(task3Dir, 0755)
	os.WriteFile(filepath.Join(task3Dir, "README.md"),
		[]byte("# TASK-03: Tests\n\n## Description\nWrite tests\n\n## Scope\nrecording-lib\n\n## Acceptance Criteria\n- Tests pass\n"), 0644)
	os.WriteFile(filepath.Join(task3Dir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// BUG-01_crash inside STORY-01
	bugDir := filepath.Join(storyDir, "BUG-01_crash")
	os.MkdirAll(bugDir, 0755)
	os.WriteFile(filepath.Join(bugDir, "README.md"),
		[]byte("# BUG-01: Crash on start\n\n## Description\nApp crashes\n\n## Scope\nrecording-lib\n\n## Acceptance Criteria\n- No crash\n"), 0644)
	os.WriteFile(filepath.Join(bugDir, "progress.md"),
		[]byte("## Status\nopen\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	return bd
}
