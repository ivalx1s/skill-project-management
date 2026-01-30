# SPEC: Project Management Skill

---

## Part 1: Planner

### Overview

The Planner is a set of CLI commands in `task-board` that build a dependency graph, group elements into phases, and render visualizations. The tool is "dumb" — it does not work with LLMs, only provides tooling. The agent itself analyzes tasks and invokes the CLI.

Gantt without dates — instead of a time axis, **phases** are used (phase 1, phase 2, ...). Parallel tasks land in the same phase.

### Users

- AI agent (primary consumer — analyzes and sets up dependencies via CLI)
- Human (reviews, adjusts, can run commands manually)

### Hierarchy (strict)

```
Epic → Story → Task/Bug
```

Orphans are impossible: a task must belong to a story, a story must belong to an epic. CLI enforces this.

### Requirements

#### R1: Hierarchical Dependency Graph

The dependency graph is built at **each level of the hierarchy** as an independent subgraph:

```
Project graph  ← epic subgraphs
  Epic graph   ← story subgraphs
    Story graph ← task/bug graph
```

**Properties:**
- Each subgraph is self-contained: you can view the plan at any level
- **Zoom in** — `task-board plan STORY-05` → phases of tasks within a story
- **Zoom out** — `task-board plan EPIC-01` → phases of stories within an epic
- **Full picture** — `task-board plan` → phases of epics for the entire project

**Dependency escalation:** when calling `task-board link`, CLI automatically escalates dependencies bottom-up **and writes them to the board**:
- `link TASK-05 --blocked-by TASK-02` (tasks from different stories) → automatically writes STORY-XX blocked-by STORY-YY
- If stories are from different epics → automatically writes EPIC-XX blocked-by EPIC-YY
- `unlink` — reverse process: removes escalated dependencies if no more cross-links exist at the lower level

#### R2: Topological Sort and Phases

At each subgraph level:
1. Reads dependencies from the board
2. Builds a topological sort
3. Groups elements into **phases**:
   - Phase 1 — elements with no incoming dependencies (can start immediately)
   - Phase 2 — elements depending only on Phase 1
   - Phase N — elements depending on Phase N-1
4. Elements within the same phase are parallel (order doesn't matter)

**Parallelism criterion:** tasks are considered parallel only if they can be executed by different sub-agents **without overlapping activity**. This means:
- Tasks don't touch the same files / modules / APIs
- The result of one task is not the input for another
- No shared resources that create conflicts during simultaneous work

The agent must account for isolation when designing tasks: if two tasks touch the same module — they are sequential, even if not formally linked via `blocked-by`.

**Sub-agent mapping:**
- One sub-agent = one story (works on the task subgraph)
- Large sub-agent = epic (works on the story subgraph)
- Small sub-agent = one task

#### R3: Plan Output

The planner outputs results:
- **To terminal** — human-readable view with phases, elements, dependencies
- **To file** — generates `plan.md` at the element level for storage and review

#### R4: Graph Rendering (Graphviz)

CLI generates a DOT file and renders to SVG/PNG/PDF via `dot`. Renders the **full hierarchy** from any scope downward on a single graph — e.g. `plan EPIC-01 --render` shows the epic with all its stories and tasks in one picture.

**Graph node:**
```
┌──────────────┐
│  TASK-12     │
│  Interface   │
└──────────────┘
```
Format: type + ID on top, name on the bottom.

**Node shapes:**
- Epic = box3d (3D box)
- Task/Story = box
- Bug = octagon

**Node colors (status):**
- white = open
- yellow = progress
- green = done
- red = blocked
- grey = closed

**Legend:** every rendered graph includes a legend block explaining the color scheme.

**Two layout modes** (`--layout`):
- `hierarchy` (default) — clusters by hierarchy: epics contain stories, stories contain tasks/bugs. Shows organizational structure with dependency arrows.
- `phases` — clusters by execution phases (Phase 1, Phase 2, ...). Shows parallelism and sequential ordering. All elements from the scope downward are grouped into phases.

**Filtering:**
- `--active` flag — exclude done and closed elements from the graph. Only open, progress, and blocked elements are rendered. Useful during active development to focus on remaining work.

**Graph structure:**
- Arrows = blocked-by (direction: dependency → dependent)
- Cluster borders: green (done), yellow (progress), red (blocked), grey (open)

**Render output location:**
- Render result goes into `.temp/` of **the element whose graph is being rendered**:
  - `task-board plan --render` → `.task-board/.temp/plan.svg`
  - `task-board plan --render --layout phases` → `.task-board/.temp/plan-phases.svg`
  - `task-board plan EPIC-01 --render` → `.task-board/EPIC-01_name/.temp/plan.svg`
  - `task-board plan STORY-05 --render` → `.task-board/EPIC-XX_name/STORY-05_name/.temp/plan.svg`

#### R5: Problem Detection

The planner must detect:
- **Cycles** — A blocks B, B blocks A (error, plan cannot be built)
- **Critical path** — the longest chain of dependencies (info, not a warning)

### CLI Interface (preliminary)

```bash
# Full project plan (epic graph)
task-board plan

# Specific epic plan (story graph)
task-board plan EPIC-01

# Specific story plan (task graph)
task-board plan STORY-05

# Save plan to file
task-board plan --save
task-board plan EPIC-01 --save

# Show only critical path
task-board plan --critical-path

# Show specific phase
task-board plan --phase 2
task-board plan STORY-05 --phase 1

# Render graph (hierarchy layout — epic/story/task clusters)
task-board plan --render
task-board plan EPIC-01 --render
task-board plan STORY-05 --render
task-board plan --render --format png

# Render graph (phases layout — phase clusters)
task-board plan --render --layout phases
task-board plan EPIC-01 --render --layout phases --format png

# Render only active elements (exclude done/closed)
task-board plan --render --active
task-board plan EPIC-01 --render --layout phases --active
```

### plan.md Format (preliminary)

```markdown
# Plan: EPIC-01 Recording

Generated: 2026-01-30

## Phase 1 (no dependencies)
- STORY-01: Audio Capture
- STORY-03: Settings UI

## Phase 2
- STORY-02: Amplitude Display (blocked by: STORY-01)

## Critical Path
STORY-01 → STORY-02 (2 phases)

## Warnings
- No issues found
```

---

## Part 2: Sub-Agents (Agent Tracking)

### Overview

Tracking which sub-agent is working on what, how many there are, what's their progress. Dashboard right in CLI. Data is never lost — filtering by freshness, no manual cleanup needed.

### Requirements

#### R6: Assignee, Created and Last Update in progress.md

Three fields in `progress.md`:

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
...
```

**Assigned To** — arbitrary string (agent name, session ID, person name). Can be assigned at any level: epic, story, task.

**Created** — ISO 8601 timestamp. Set once when element is created via `task-board create`. Never changes.

**Last Update** — ISO 8601 timestamp. Auto-updated on **any** change to progress.md:
- `progress status`
- `progress check` / `uncheck`
- `progress add-item`
- `progress notes`
- `assign` / `unassign`

CLI commands:

```bash
task-board assign STORY-03 --agent "agent-1"
task-board unassign STORY-03
```

#### R7: Agent Dashboard

CLI command `task-board agents` — shows who is doing what:

```
Sub-Agent Dashboard

AGENT    SCOPE                        STATUS    PROGRESS  LAST UPDATE
agent-1  STORY-03: plan-output        progress  3/5 done  30 sec ago
agent-2  STORY-04: graphviz-render    progress  2/4 done  1 min ago
agent-3  STORY-05: problem-detection  done      3/3 done  5 min ago

Total: 3 agents, 2 active, 1 done
```

**Logic:**
- Collects all elements with non-empty `Assigned To`
- By default shows only **fresh** entries (status != done OR last update < `--stale` minutes ago, default 30)
- For each: agent name, scope, status, child progress (done/total), last update time (human-readable: "30 sec ago", "5 min ago")
- Footer — totals: how many agents, how many active, how many done
- Old entries (done + stale timestamp) don't appear in default view — they "expire" on their own
- `--all` shows everyone including stale

**Liveness monitoring:** if an agent hasn't updated for a long time but status is not done — visible that it's stuck or crashed.

### CLI Interface

```bash
# Assign agent
task-board assign STORY-03 --agent "agent-1"

# Remove assignment
task-board unassign STORY-03

# Dashboard (active and fresh only, default 30 min window)
task-board agents

# Custom freshness window (in minutes)
task-board agents --stale 60

# All including stale
task-board agents --all
```

---
