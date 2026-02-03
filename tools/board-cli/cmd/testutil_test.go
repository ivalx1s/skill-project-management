package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// Test IDs in new distributed format (YYMMDD-xxxxxx)
const (
	testEpic1ID   = "EPIC-260101-aaaaaa"
	testEpic2ID   = "EPIC-260101-bbbbbb"
	testStory1ID  = "STORY-260101-cccccc"
	testStory2ID  = "STORY-260101-dddddd"
	testStory3ID  = "STORY-260101-eeeeee"
	testTask1ID   = "TASK-260101-ffffff"
	testTask2ID   = "TASK-260101-gggggg"
	testTask3ID   = "TASK-260101-hhhhhh"
	testTask4ID   = "TASK-260101-iiiiii"
	testBug1ID    = "BUG-260101-jjjjjj"
)

// setupTestBoard creates a temp board with known structure for cmd tests.
// Returns the board directory path.
func setupTestBoard(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bd := filepath.Join(dir, ".task-board")
	os.MkdirAll(bd, 0755)

	// EPIC-260101-aaaaaa_recording
	epicDir := filepath.Join(bd, testEpic1ID+"_recording")
	os.MkdirAll(epicDir, 0755)
	os.WriteFile(filepath.Join(epicDir, "README.md"),
		[]byte("# "+testEpic1ID+": Recording\n\n## Description\nRecording epic\n\n## Scope\nrecording-lib\n\n## Acceptance Criteria\n- Works\n"), 0644)
	os.WriteFile(filepath.Join(epicDir, "progress.md"),
		[]byte("## Status\ndevelopment\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// EPIC-260101-bbbbbb_storage
	epic2Dir := filepath.Join(bd, testEpic2ID+"_storage")
	os.MkdirAll(epic2Dir, 0755)
	os.WriteFile(filepath.Join(epic2Dir, "README.md"),
		[]byte("# "+testEpic2ID+": Storage\n\n## Description\nStorage epic\n\n## Scope\nstorage-lib\n\n## Acceptance Criteria\n- Works\n"), 0644)
	os.WriteFile(filepath.Join(epic2Dir, "progress.md"),
		[]byte("## Status\nbacklog\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// STORY-260101-eeeeee_migration inside EPIC-02 (for cross-epic tests)
	story3Dir := filepath.Join(epic2Dir, testStory3ID+"_migration")
	os.MkdirAll(story3Dir, 0755)
	os.WriteFile(filepath.Join(story3Dir, "README.md"),
		[]byte("# "+testStory3ID+": Migration\n\n## Description\nDB migration\n\n## Scope\nstorage-lib\n\n## Acceptance Criteria\n- Migrated\n"), 0644)
	os.WriteFile(filepath.Join(story3Dir, "progress.md"),
		[]byte("## Status\nbacklog\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-260101-iiiiii_schema inside STORY-03 (for cross-story/epic tests)
	task4Dir := filepath.Join(story3Dir, testTask4ID+"_schema")
	os.MkdirAll(task4Dir, 0755)
	os.WriteFile(filepath.Join(task4Dir, "README.md"),
		[]byte("# "+testTask4ID+": Schema\n\n## Description\nDefine schema\n\n## Scope\nstorage-lib\n\n## Acceptance Criteria\n- Schema defined\n"), 0644)
	os.WriteFile(filepath.Join(task4Dir, "progress.md"),
		[]byte("## Status\nbacklog\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// STORY-260101-cccccc_audio-capture inside EPIC-01
	storyDir := filepath.Join(epicDir, testStory1ID+"_audio-capture")
	os.MkdirAll(storyDir, 0755)
	os.WriteFile(filepath.Join(storyDir, "README.md"),
		[]byte("# "+testStory1ID+": Audio Capture\n\n## Description\nCapture audio\n\n## Scope\nrecording-lib\n\n## Acceptance Criteria\n- Records\n"), 0644)
	os.WriteFile(filepath.Join(storyDir, "progress.md"),
		[]byte("## Status\nbacklog\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// STORY-260101-dddddd_amplitude inside EPIC-01
	story2Dir := filepath.Join(epicDir, testStory2ID+"_amplitude")
	os.MkdirAll(story2Dir, 0755)
	os.WriteFile(filepath.Join(story2Dir, "README.md"),
		[]byte("# "+testStory2ID+": Amplitude\n\n## Description\nAmplitude analysis\n\n## Scope\nrecording-lib\n\n## Acceptance Criteria\n- Analyzes\n"), 0644)
	os.WriteFile(filepath.Join(story2Dir, "progress.md"),
		[]byte("## Status\nbacklog\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-260101-ffffff_interface inside STORY-01
	taskDir := filepath.Join(storyDir, testTask1ID+"_interface")
	os.MkdirAll(taskDir, 0755)
	os.WriteFile(filepath.Join(taskDir, "README.md"),
		[]byte("# "+testTask1ID+": Interface\n\n## Description\nDefine interface\n\n## Scope\nrecording-lib\n\n## Acceptance Criteria\n- Interface defined\n"), 0644)
	os.WriteFile(filepath.Join(taskDir, "progress.md"),
		[]byte("## Status\nbacklog\n\n## Blocked By\n- (none)\n\n## Blocks\n- "+testTask2ID+"\n\n## Checklist\n- [ ] Step 1\n- [x] Step 2\n\n## Notes\nStarted work\n"), 0644)

	// TASK-260101-gggggg_impl inside STORY-01, blocked by TASK-01
	task2Dir := filepath.Join(storyDir, testTask2ID+"_impl")
	os.MkdirAll(task2Dir, 0755)
	os.WriteFile(filepath.Join(task2Dir, "README.md"),
		[]byte("# "+testTask2ID+": Implementation\n\n## Description\nImplement it\n\n## Scope\nrecording-lib\n\n## Acceptance Criteria\n- Implemented\n"), 0644)
	os.WriteFile(filepath.Join(task2Dir, "progress.md"),
		[]byte("## Status\nbacklog\n\n## Blocked By\n- "+testTask1ID+"\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// TASK-260101-hhhhhh_tests inside STORY-01
	task3Dir := filepath.Join(storyDir, testTask3ID+"_tests")
	os.MkdirAll(task3Dir, 0755)
	os.WriteFile(filepath.Join(task3Dir, "README.md"),
		[]byte("# "+testTask3ID+": Tests\n\n## Description\nWrite tests\n\n## Scope\nrecording-lib\n\n## Acceptance Criteria\n- Tests pass\n"), 0644)
	os.WriteFile(filepath.Join(task3Dir, "progress.md"),
		[]byte("## Status\nbacklog\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	// BUG-260101-jjjjjj_crash inside STORY-01
	bugDir := filepath.Join(storyDir, testBug1ID+"_crash")
	os.MkdirAll(bugDir, 0755)
	os.WriteFile(filepath.Join(bugDir, "README.md"),
		[]byte("# "+testBug1ID+": Crash on start\n\n## Description\nApp crashes\n\n## Scope\nrecording-lib\n\n## Acceptance Criteria\n- No crash\n"), 0644)
	os.WriteFile(filepath.Join(bugDir, "progress.md"),
		[]byte("## Status\nbacklog\n\n## Blocked By\n- (none)\n\n## Blocks\n- (none)\n\n## Checklist\n(empty)\n\n## Notes\n"), 0644)

	return bd
}
