---
name: project-management
description: Create and manage file-based task boards from project specs. Epic → Story → Task hierarchy with progress tracking. Includes planning capabilities.
triggers:
  - creating task board from spec
  - breaking down spec into tasks
  - project planning
  - task management
  - creating bugs
  - bug tracking
---

# Project Management Skill

File-based project management via task board + CLI tool. No external tools — everything in git.

---

# Part 1: Task Management

---

## When to Use

- New project with SPEC/requirements → break down into epics/stories/tasks
- Need progress tracking in files
- Working with AI agent on a project
- Tracking bugs

---

## CLI Tool: `task-board`

All board operations go through the `task-board` CLI. No manual file editing.

**Location:** `~/.agents/skills/project-management/tools/board-cli/task-board`
**In PATH:** via symlink in `~/.local/bin/task-board`

### Quick Reference

```bash
# Create elements
task-board create epic --name "recording"
task-board create story --epic EPIC-01 --name "audio-capture"
task-board create task --story STORY-05 --name "interface" --description "..."
task-board create bug --story STORY-05 --name "crash" --description "..."

# Update README fields
task-board update TASK-12 --title "new title" --description "..." --scope "..." --ac "..."

# Show full element details
task-board show TASK-12

# Progress & status
task-board progress status TASK-12 progress    # open|progress|done|closed|blocked
task-board progress checklist TASK-12           # show checklist
task-board progress check TASK-12 3            # check item
task-board progress uncheck TASK-12 2          # uncheck item
task-board progress add-item TASK-12 "Write tests"  # add checklist item
task-board progress notes TASK-12 "Some notes"       # append notes
task-board progress notes TASK-12 "Replace" --set    # replace all notes

# Dependencies
task-board link TASK-13 --blocked-by TASK-12   # add dependency
task-board unlink TASK-13 --blocked-by TASK-12 # remove dependency

# Move elements
task-board move TASK-13 --to STORY-02          # move task to different story
task-board move STORY-05 --to EPIC-02          # move story to different epic

# Delete elements
task-board delete TASK-13                      # delete leaf element
task-board delete EPIC-01 --force              # delete with children

# View
task-board list epics                          # list all epics
task-board list tasks --status open            # filter by status
task-board list tasks --story STORY-05         # filter by parent
task-board list bugs --status open             # list open bugs
task-board summary                             # board overview

# Search & validate
task-board search "AudioRecorder"              # regex search
task-board validate                            # check board structure

# Custom board directory
task-board --board-dir /path/to/.task-board create epic --name "test"
```

---

## Task Board Structure

### Hierarchy

```
Epic → Story → Task/Bug
```

- **Epic** = large feature / module
- **Story** = specific functionality within epic
- **Task** = atomic unit of work
- **Bug** = defect tracked within a story

### Naming Convention

**Skvoznaya (global) numbering** — each type has its own global auto-increment counter.

```
EPIC-01_recording/
  STORY-05_audio-capture/
    TASK-12_audiorecorder-interface/
    TASK-13_implementation/
    BUG-47_some-bug/
```

Format: `{TYPE}-{NN}_{kebab-case-name}`
- `EPIC-NN_name`
- `STORY-NN_name`
- `TASK-NN_name`
- `BUG-NN_name`

IDs are globally unique: `TASK-12` exists only once across the entire board.

### Directory Layout

```
.task-board/
├── system.md                     # Global counters
├── EPIC-01_recording/
│   ├── README.md                 # Epic description, scope, AC
│   ├── progress.md               # Status, blockedBy, checklist, notes
│   ├── STORY-01_audio-capture/
│   │   ├── README.md
│   │   ├── progress.md
│   │   ├── TASK-01_interface/
│   │   │   ├── README.md
│   │   │   └── progress.md
│   │   ├── TASK-02_implementation/
│   │   │   ├── README.md
│   │   │   └── progress.md
│   │   └── BUG-01_crash-on-start/
│   │       ├── README.md
│   │       └── progress.md
│   └── STORY-02_amplitude/
│       └── ...
└── EPIC-02_storage/
    └── ...
```

### system.md

Tracks global counters. Managed automatically by CLI.

```markdown
## Counters
- epic: 9
- story: 43
- task: 156
- bug: 11
```

---

## Statuses

`open` | `progress` | `done` | `closed` | `blocked`

- **open** — not started
- **progress** — in work
- **done** — completed
- **closed** — archived/won't do
- **blocked** — waiting on dependency

---

## File Formats

### README.md — descriptive part

Static description only — what the element is, its scope, and acceptance criteria. Does not change during execution. This is the "what" of the element.

```markdown
# TASK-12: AudioRecorder Interface

## Description
Create interface AudioRecorder with start/stop/pause methods.

## Scope
Define in recording-lib module.

## Acceptance Criteria
- Interface defined
- Flows for state and data
```

### progress.md — full lifecycle journal

Dynamic tracking — the complete lifecycle of an element. Contains status, assignment, timestamps, dependencies, checklist, and notes. This is the "how" and "when" of the element.

```markdown
## Status
progress

## Assigned To
agent-1

## Created
2026-01-30T14:00:00Z

## Last Update
2026-01-30T15:03:35Z

## Blocked By
- TASK-11

## Blocks
- TASK-13

## Checklist
- [x] Define interface
- [ ] Add Flow types
- [ ] Write tests

## Notes
Started implementation
```

**Fields:**
- **Status** — `open` | `progress` | `done` | `closed` | `blocked`
- **Assigned To** — agent or person working on this element (set via `task-board assign`)
- **Created** — ISO 8601 timestamp, set once at creation
- **Last Update** — ISO 8601 timestamp, auto-updated on every progress.md write
- **Blocked By / Blocks** — bidirectional dependencies
- **Checklist** — sub-items tracking
- **Notes** — free-form notes

**Dependencies are bidirectional.** When you run `task-board link TASK-13 --blocked-by TASK-12`:
- TASK-13 gets `Blocked By: TASK-12`
- TASK-12 gets `Blocks: TASK-13`

This allows traversal in both directions — who blocks me and who I block.

**Dependency escalation.** When linking elements from different parents, CLI automatically creates implied dependencies up the hierarchy:
- `link TASK-05 --blocked-by TASK-02` (different stories) → STORY blocked by STORY
- Different epics → EPIC blocked by EPIC
- `unlink` reverses escalation if no cross-links remain.

---

## Bugs

Bugs are elements of the board, living inside stories alongside tasks.

```bash
# Create bug
task-board create bug --story STORY-05 --name "crash-on-start" --description "App crashes when starting recording"

# Track like any element
task-board progress status BUG-01 progress
task-board progress add-item BUG-01 "Reproduce the issue"
task-board progress add-item BUG-01 "Find root cause"
task-board progress add-item BUG-01 "Write fix"
task-board progress add-item BUG-01 "Add regression test"
```

---

## Workflow

### Full Development Cycle (MANDATORY)

When the user asks to build/implement something, the agent MUST follow this end-to-end flow:

#### 1. Spec
- Capture the request in `SPEC.md` (create or update)
- Clarify requirements with the user if anything is ambiguous
- SPEC is the source of truth — all work traces back to it

#### 2. Plan on the Board
- Create epics, stories, and tasks on the board via `task-board create`
- Refine each element: read the generated README.md, clarify gaps with user, update via `task-board update`
- No element should have placeholder text — description and AC must be complete and unambiguous
- Set up dependencies via `task-board link` (CLI auto-escalates cross-parent links)

#### 3. Phase Planning
- Run `task-board plan` to build phases from the dependency graph
- Review phases with the user — which stories are parallel, which are sequential
- Render graphs for visual overview:
  ```bash
  task-board plan EPIC-XX --render --format png                  # hierarchy view
  task-board plan EPIC-XX --render --format png --layout phases  # phases view
  ```

#### 4. Parallel Execution via Sub-Agents
- **The coordinator is a SUPERVISOR — it does NOT write code itself**
- ALL implementation work (code, tests, docs updates) goes to sub-agents
- Assign agents to stories/tasks: `task-board assign STORY-XX --agent "agent-name"`
- **Agent naming:** give sub-agents meaningful names that reflect their scope — e.g. `agent-auth`, `agent-api-tests`, `agent-schema`. Not `agent-1`, `agent-2`. The name appears in the dashboard and should be instantly recognizable.
- Spawn sub-agents in parallel for independent stories (same phase)
- Each sub-agent:
  - Works on its story/task scope
  - Writes code AND tests
  - Updates task statuses as it progresses
- Wait for Phase N to complete before starting Phase N+1
- Even small tasks go to sub-agents — the coordinator only plans, delegates, and reviews

#### 5. Supervision
- Monitor progress: `task-board agents --stale 60`
- When a sub-agent finishes — **review its work**:
  - Run tests: do they pass?
  - Check code quality: does it match the AC?
  - Validate integration: does it work with other completed stories?
- If something is incomplete or broken — return the task to the sub-agent (resume it or spawn a new one)
- If the work is solid — move tasks to done
- **Never do the sub-agent's work** — always delegate back

#### 6. Board Hygiene
- Keep the board in sync with reality at all times
- Move tasks through statuses as work progresses: `open` → `progress` → `done`
- Close stories when all tasks are done
- Close epics when all stories are done
- Run `task-board validate` periodically

#### 7. Visualization
- Render graphs at key milestones for the user:
  - Before starting: show the plan and phases
  - During execution: show progress (colors reflect statuses)
  - After completion: show the final state
- Use both layouts as needed:
  - `--layout hierarchy` — organizational structure (who owns what)
  - `--layout phases` — execution order (what runs when)
- Use `--active` to hide done/closed elements and focus on remaining work

### Iterative Refinement (for individual elements)

When creating ANY element (epic/story/task/bug), the agent MUST:

1. **Create** element via `task-board create`
2. **Read** the generated README.md
3. **Clarify** — if description is vague or has gaps, ask the user
4. **Update** via `task-board update` with refined content
5. **Repeat** steps 2-4 until description is exhaustive
6. Element is **ready** only when description and AC are complete and unambiguous

### Project Start

1. Create `.task-board/` via first `task-board create` (auto-creates)
2. Break SPEC into epics: `task-board create epic --name "..."`
3. Create stories for each epic: `task-board create story --epic EPIC-01 --name "..."`
4. Detail tasks when taking story into work

### Working on a Task

1. Choose task → `task-board progress status TASK-12 progress`
2. Work
3. Done → `task-board progress status TASK-12 done`
4. Update checklist along the way

### Dependencies

```bash
# TASK-13 depends on TASK-12
task-board link TASK-13 --blocked-by TASK-12

# CLI will refuse to move TASK-13 to progress/done while TASK-12 is not done
task-board progress status TASK-13 progress
# Error: cannot set TASK-13 to progress — blocked by: TASK-12 (status: open)

# Remove dependency
task-board unlink TASK-13 --blocked-by TASK-12
```

### Board Review

```bash
# Quick overview
task-board summary

# Full element details
task-board show TASK-12

# What's in progress
task-board list tasks --status progress

# What's blocked
task-board list tasks --status blocked

# Full validation
task-board validate
```

---

## Tips

- **Don't create all tasks upfront** — detail when taking story into work
- **AC must be verifiable** — "works" is bad, "test passes" is good
- **Update progress immediately** — don't defer
- **Use checklist** for tracking subtasks within a task
- **Bugs live in stories** — always attach to a relevant story
- **Sub-agents write tests** — every sub-agent is responsible for tests in its scope
- **Supervisor reviews** — coordinator must verify sub-agent work before marking done
- **Visualize at milestones** — render graphs before, during, and after execution

---

## Setup

Run the setup script to install dependencies, build the CLI, and symlink to PATH:

```bash
./scripts/setup.sh
```

The script will:
1. **Install Go** via Homebrew (if not present)
2. **Build** `task-board` binary from source
3. **Symlink** to `~/.local/bin/task-board`
4. **Verify** PATH and installation

### Manual Build

```bash
cd tools/board-cli
go build -o task-board .
```

---
---

# Part 2: Planning

## Plan Command

Build and visualize dependency graph with phases.

```bash
# Full project plan (epic-level)
task-board plan

# Epic plan (story-level)
task-board plan EPIC-01

# Story plan (task-level)
task-board plan STORY-05

# Save to file
task-board plan --save
task-board plan EPIC-01 --save

# Critical path only
task-board plan --critical-path

# Specific phase
task-board plan --phase 2

# Render graph (Graphviz) — hierarchy layout (default)
task-board plan --render
task-board plan EPIC-01 --render --format png

# Render graph — phases layout
task-board plan --render --layout phases
task-board plan EPIC-01 --render --layout phases --format png

# Render only active elements (exclude done/closed)
task-board plan --render --active
task-board plan EPIC-01 --render --active --format png
task-board plan --render --layout phases --active
```

**`--active` flag:** Filters out `done` and `closed` elements from rendered graphs, showing only `open`, `progress`, and `blocked` elements. Works with both `--layout hierarchy` (default) and `--layout phases`. Useful during execution to focus on remaining work without visual clutter from completed items.

Rendered graphs go to `.temp/` inside the scope element's directory.

---

# Part 3: Agent Tracking

## Assigning Agents

```bash
# Assign agent to element
task-board assign STORY-03 --agent "agent-1"

# Remove assignment
task-board unassign STORY-03
```

## Agent Dashboard

```bash
# Show active agents (default 30 min freshness window)
task-board agents

# Custom freshness window
task-board agents --stale 60

# Show all (including done/stale)
task-board agents --all
```

Dashboard shows: agent name, scope, status, child progress (done/total), last update time. Use `--stale N` to set freshness window in minutes (default 30). Stale entries (done + old timestamp) auto-filter from default view.

## Sub-Agent Workflow

1. Coordinator breaks work into stories, assigns agents:
   ```bash
   task-board assign STORY-03 --agent "agent-1"
   task-board assign STORY-04 --agent "agent-2"
   ```
2. Each sub-agent works on its scope, updating task statuses
3. Coordinator monitors: `task-board agents`
4. When all done — agents auto-disappear from default dashboard view

---
