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
task-board progress status TASK-12 development # backlog|analysis|to-dev|development|to-review|reviewing|done|closed|blocked
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

**Distributed IDs** — each element gets a unique ID based on date + random hash.

```
EPIC-260203-a1b2c3_recording/
  STORY-260203-d4e5f6_audio-capture/
    TASK-260203-g7h8i9_audiorecorder-interface/
    TASK-260203-j0k1l2_implementation/
    BUG-260203-m3n4o5_some-bug/
```

Format: `{TYPE}-{YYMMDD}-{hash}_{kebab-case-name}`
- `EPIC-YYMMDD-xxxxxx_name`
- `STORY-YYMMDD-xxxxxx_name`
- `TASK-YYMMDD-xxxxxx_name`
- `BUG-YYMMDD-xxxxxx_name`

Where:
- `YYMMDD` — creation date (year, month, day)
- `xxxxxx` — 6-character base36 hash (collision-resistant)

IDs are globally unique and work in parallel environments (no counter conflicts).

### Directory Layout

```
.task-board/
├── EPIC-260203-a1b2c3_recording/
│   ├── README.md                 # Epic description, scope, AC
│   ├── progress.md               # Status, blockedBy, checklist, notes
│   ├── STORY-260203-d4e5f6_audio-capture/
│   │   ├── README.md
│   │   ├── progress.md
│   │   ├── TASK-260203-g7h8i9_interface/
│   │   │   ├── README.md
│   │   │   └── progress.md
│   │   ├── TASK-260203-j0k1l2_implementation/
│   │   │   ├── README.md
│   │   │   └── progress.md
│   │   └── BUG-260203-m3n4o5_crash-on-start/
│   │       ├── README.md
│   │       └── progress.md
│   └── STORY-260203-p6q7r8_amplitude/
│       └── ...
└── EPIC-260203-s9t0u1_storage/
    └── ...
```

No `system.md` needed — IDs are self-contained and don't require global counters.

---

## Statuses

9 statuses organized in a workflow:

| Status | Meaning | Who sets |
|--------|---------|----------|
| **backlog** | Captured, no work yet | anyone |
| **analysis** | Researching, decomposing | anyone |
| **to-dev** | Ready for development | orchestrator |
| **development** | Writing code | sub-agent |
| **to-review** | Code ready for review | sub-agent |
| **reviewing** | Checking the work | orchestrator (auto) |
| **done** | Reviewed and accepted | orchestrator (auto) |
| **closed** | Won't do (reason in notes) | anyone |
| **blocked** | External block (reason in notes) | anyone |

### Status Flow

**Happy path:**
```
backlog → analysis → to-dev → development → to-review → reviewing → done
```

**Returns from reviewing:**
- `reviewing → analysis` — need more research
- `reviewing → to-dev` — code issues, needs rework

**Terminal from any status:**
- `any → closed` — won't do / cancelled
- `any → blocked` — external block

**Note on closed:** Use `closed` for cancelled, obsolete, or won't-do items. Don't delete — close with a note explaining why. History is preserved, and the item can be reopened if needed.

**Don't delete elements.** Use `closed` instead of `delete`. Deletion loses history and context. The only exception is removing test/junk data during board setup.

**Orchestrator auto-review:** When sub-agent sets `to-review`, orchestrator automatically picks it up for review (`reviewing`) and closes (`done`) without user confirmation. User only involved if issues found — then task returns to `to-dev` or `analysis`.

### blocked vs is_blocked

Two different concepts:

- **blocked** (status) — explicit status for external blocks outside the board
- **is_blocked** (computed) — from `blocked-by` dependencies; if dependency is not `done`/`closed`, element is blocked

CLI display:
```
TASK-05: to-dev [BLOCKED by TASK-02]     ← computed from dependency
TASK-07: blocked                          ← explicit status
```

### Auto-Promotion & Auto-Reopen

The CLI automatically manages parent statuses:

**Auto-promotion:** When all children of a story/epic are `done` or `closed`, the parent is automatically promoted to `done`. Cascades up the hierarchy.

```bash
task-board progress status TASK-12 done
# TASK-12 → done
# STORY-05 → done (auto-promoted)
# EPIC-01 → done (auto-promoted)
```

**Auto-reopen:** When a child becomes active and the parent is `done`/`closed`, the parent is reopened. Cascades up.

```bash
task-board progress status TASK-12 to-dev
# TASK-12 → to-dev
# STORY-05 → development (auto-reopened)
# EPIC-01 → development (auto-reopened)
```

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
- **Status** — `backlog` | `analysis` | `to-dev` | `development` | `to-review` | `reviewing` | `done` | `closed` | `blocked`
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
task-board progress status BUG-01 development
task-board progress add-item BUG-01 "Reproduce the issue"
task-board progress add-item BUG-01 "Find root cause"
task-board progress add-item BUG-01 "Write fix"
task-board progress add-item BUG-01 "Add regression test"
```

---

## Workflow

### Full Development Cycle (MANDATORY)

When the user asks to build/implement something, the agent MUST follow this end-to-end flow:

#### Spec
- Capture the request in `SPEC.md` (create or update)
- Clarify requirements with the user if anything is ambiguous
- SPEC is the source of truth — all work traces back to it

#### Plan on the Board
- Create epics and stories on the board via `task-board create`
- Refine each element: read the generated README.md, clarify gaps with user, update via `task-board update`
- No element should have placeholder text — description and AC must be complete and unambiguous

#### Story Detailing (MANDATORY before planning)
- **Delegate to sub-agents:** spawn sub-agents to decompose each story
- Sub-agent workflow:
  - Set story status to `analysis`: `task-board progress status STORY-XX analysis`
  - Read the story README.md
  - Explore relevant codebase areas
  - Create tasks that break down the implementation work
  - Each task should be atomic — one clear deliverable
  - Set task dependencies within the story via `task-board link`
  - **IMPORTANT:** After finishing, set story back to `to-dev` (NOT `done`!)
    - `done` means implementation complete, not decomposition complete
    - `task-board progress status STORY-XX to-dev`
- **Orchestrator reviews:** after sub-agents finish, review the decomposition
  - Are tasks granular enough?
  - Are descriptions and AC clear?
  - Any missing tasks?
- **Do NOT proceed to Phase Planning until all stories have tasks**
- This ensures the plan shows the real scope of work

#### Phase Planning
- **Review cross-story dependencies:** sub-agents set dependencies within their story, but orchestrator must link dependencies BETWEEN stories
- Add missing cross-story/cross-epic links via `task-board link` (CLI auto-escalates to parent level)
- Run `task-board plan` to build phases from the dependency graph
- Verify phases make sense — adjust dependencies if needed
- Render graphs for visual overview:
  ```bash
  task-board plan EPIC-XX --render --format png                  # hierarchy view
  task-board plan EPIC-XX --render --format png --layout phases  # phases view
  ```

#### User Approval (MANDATORY)
- **ALWAYS show the plan to the user and wait for explicit approval**
- Display:
  - Phase breakdown: `task-board plan EPIC-XX`
  - Task-level details: `task-board plan STORY-XX`
  - Summary of what will be built
- Ask: "Plan ready. Proceed?" (or similar)
- **DO NOT start execution until user confirms**
- If user has concerns → adjust plan, re-show, re-confirm

#### Pre-Launch Checklist (MANDATORY)

Before spawning ANY implementation agent, the coordinator MUST verify:

- [ ] Story status is `to-dev` (not backlog!) — run `task-board progress status STORY-XX to-dev`
- [ ] Agent assigned via `task-board assign STORY-XX --agent "agent-name"`
- [ ] Prompt includes CONCRETE task IDs (e.g., `TASK-260203-ae3fcd`), NOT placeholders like `TASK-XXX`
- [ ] Prompt starts with the FIRST command agent must run (status update)
- [ ] First task ID in prompt matches the Phase 1 unblocked task

**If you skip this checklist, agents WILL ignore task-board updates.**

#### Execution via Sub-Agents
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

#### Supervision
- Monitor progress: `task-board agents --stale 60`
- When a sub-agent finishes — **review its work**:
  - Run tests: do they pass?
  - Check code quality: does it match the AC?
  - Validate integration: does it work with other completed stories?
- If something is incomplete or broken — return the task to the sub-agent (resume it or spawn a new one)
- If the work is solid — move tasks to done
- **Never do the sub-agent's work** — always delegate back

#### Board Hygiene
- Keep the board in sync with reality at all times
- Move tasks through statuses as work progresses: `to-dev` → `development` → `to-review` → `reviewing` → `done`
- Close stories when all tasks are done
- Close epics when all stories are done
- Run `task-board validate` periodically

#### Visualization
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

- Create `.task-board/` via first `task-board create` (auto-creates)
- Break SPEC into epics: `task-board create epic --name "..."`
- Create stories for each epic: `task-board create story --epic EPIC-01 --name "..."`
- Detail tasks for each story: `task-board create task --story STORY-XX --name "..."`
- Run phase planning only after all stories have tasks

### Working on a Task

1. Choose task → `task-board progress status TASK-12 development`
2. Work (write code, tests)
3. Ready for review → `task-board progress status TASK-12 to-review`
4. Orchestrator reviews → `task-board progress status TASK-12 reviewing`
5. If OK → `task-board progress status TASK-12 done`
6. If issues → `task-board progress status TASK-12 to-dev` (back to sub-agent)
7. Update checklist along the way

### Dependencies

```bash
# TASK-13 depends on TASK-12
task-board link TASK-13 --blocked-by TASK-12

# CLI will refuse to move TASK-13 to development while TASK-12 is not done
task-board progress status TASK-13 development
# Error: cannot start TASK-13 — blocked by: TASK-12 (status: to-dev)

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

- **Detail stories before planning** — create tasks for all stories before phase planning, so the plan shows real scope
- **AC must be verifiable** — "works" is bad, "test passes" is good
- **Update progress immediately** — don't defer
- **Use checklist** for tracking subtasks within a task
- **Bugs live in stories** — always attach to a relevant story
- **Sub-agents write tests** — every sub-agent is responsible for tests in its scope
- **Supervisor reviews** — coordinator must verify sub-agent work before marking done
- **Visualize at milestones** — render graphs before, during, and after execution
- **Verify agents early** — run `task-board agents` within 2 minutes of launch; if status not updated, stop and restart with better prompt
- **Concrete IDs only** — never use placeholders like `TASK-XXX` in prompts; always use real IDs like `TASK-260203-ae3fcd`

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

## Sub-Agent Prompt Templates (CRITICAL)

### ⚠️ Sub-Agent Reality Check

Sub-agents often ignore task-board CLI instructions when focused on writing code. This is a known limitation. To maximize compliance:

1. **Put the CLI command FIRST** in the prompt — before any context or task description
2. **Use CONCRETE IDs** — never use placeholders like `TASK-XXX`, always use real IDs like `TASK-260203-ae3fcd`
3. **Repeat the instruction** at the END of the prompt as a reminder
4. **Verify within 2 minutes** — run `task-board agents` shortly after launch to confirm status updates
5. **If no updates** — stop the agent, fix the prompt, restart

**Common failure modes:**
- Agent reads code, starts implementing, forgets to update status
- Agent uses placeholder IDs from template instead of real IDs
- Agent treats task-board instructions as "optional"

### Decomposition Agent (for Story Detailing phase)

When spawning a sub-agent to decompose a story into tasks:

```
## Task Board Protocol (MANDATORY)

You are decomposing a story into tasks. This is PLANNING, not implementation.

BEFORE starting:
  task-board progress status STORY-XX analysis

Create tasks:
  task-board create task --story STORY-XX --name "task-name" --description "..."
  task-board update TASK-YY --ac "- criterion 1\n- criterion 2"
  task-board link TASK-YY --blocked-by TASK-ZZ  (if dependencies exist)

AFTER finishing decomposition:
  task-board progress status STORY-XX to-dev   # NOT done! Decomposition != implementation
```

### Implementation Agent (for Execution phase)

When spawning a sub-agent to implement tasks:

```
## ⚠️ FIRST COMMAND — RUN IMMEDIATELY

Before reading ANY code, run this command RIGHT NOW:

  task-board progress status TASK-260203-ae3fcd development

This is NOT optional. The coordinator is monitoring via `task-board agents`.
If you don't run this command, you will be stopped and restarted.

---

## Task Board Protocol (MANDATORY)

You MUST use task-board CLI to track your progress.

For EACH task:
1. BEFORE starting: `task-board progress status TASK-260203-XXXXXX development`
2. Do the work (write code, tests)
3. AFTER completing: `task-board progress status TASK-260203-XXXXXX to-review`
4. Move to next task

If blocked:
  task-board progress status TASK-260203-XXXXXX blocked
  task-board progress notes TASK-260203-XXXXXX "Reason for block"

---

## Your Tasks (in order):

1. TASK-260203-ae3fcd: [task name] — [description]
2. TASK-260203-xxxxxx: [task name] — [description] (blocked by above)
...

---

## REMINDER: Update task-board status before and after EACH task!
```

**Key differences from naive template:**
- First command is FIRST, before any context
- Concrete task IDs, not placeholders
- Warning about being stopped if not compliant
- Reminder repeated at the end

**Why this matters:**
- `task-board agents` shows real-time progress based on statuses
- If sub-agent doesn't update statuses, coordinator sees stale data
- Board becomes useless for tracking

**Verification after sub-agent completes:**
```bash
task-board list tasks --story STORY-XX   # All should be to-review
task-board agents                         # Agent should show N/N to-review
```

---
