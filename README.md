# Project Management Skill

File-based project management for AI agents. Go CLI tool (`task-board`) that manages hierarchical task boards stored as directories and markdown files. Everything lives in git, no external dependencies.

## Features

- **Hierarchical board:** Epic → Story → Task/Bug — strict hierarchy, no orphans
- **Dependency graph:** `link`/`unlink` with automatic escalation up the hierarchy
- **Planner:** topological sort into phases, critical path detection
- **Graph rendering:** Graphviz DOT → SVG/PNG with two layouts (hierarchy, phases), status colors, legend, `--active` filter
- **Agent tracking:** assign sub-agents, monitor progress via dashboard with freshness filtering
- **Full lifecycle journal:** `progress.md` with status, assignee, created/last-update timestamps, checklist, notes

## Setup

```bash
./scripts/setup.sh
```

Installs Go (via Homebrew), builds the `task-board` binary, symlinks to `~/.local/bin/task-board`.

### Manual Build

```bash
cd tools/board-cli
go build -o task-board .
```

### Requirements

- **Go** 1.21+
- **Graphviz** (for `--render` graph output) — `brew install graphviz`

## Quick Start

```bash
# Create board elements
task-board create epic --name "auth-system"
task-board create story --epic EPIC-01 --name "login-flow"
task-board create task --story STORY-01 --name "jwt-tokens" --description "Implement JWT auth"

# Set dependencies
task-board link TASK-02 --blocked-by TASK-01

# Track progress
task-board progress status TASK-01 progress
task-board progress status TASK-01 done

# View plan with phases
task-board plan EPIC-01

# Render dependency graph
task-board plan EPIC-01 --render --format png

# Render phases layout (only active elements)
task-board plan EPIC-01 --render --layout phases --active

# Assign sub-agents and monitor
task-board assign STORY-01 --agent "agent-auth"
task-board agents
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `create epic/story/task/bug` | Create board elements |
| `update ID` | Update README.md fields |
| `show ID` | Show full element details |
| `progress status ID STATUS` | Set status (open/progress/done/closed/blocked) |
| `progress checklist ID` | Show checklist |
| `progress check/uncheck ID N` | Check/uncheck checklist item |
| `progress add-item ID "text"` | Add checklist item |
| `progress notes ID "text"` | Append notes |
| `link ID --blocked-by ID` | Add dependency (auto-escalates cross-parent) |
| `unlink ID --blocked-by ID` | Remove dependency (auto-de-escalates) |
| `plan [ID]` | Show execution plan with phases |
| `plan [ID] --render` | Render Graphviz graph |
| `plan [ID] --render --layout phases` | Render with phase clusters |
| `plan [ID] --render --active` | Exclude done/closed from graph |
| `plan [ID] --save` | Save plan as plan.md |
| `plan [ID] --critical-path` | Show only critical path |
| `assign ID --agent "name"` | Assign agent to element |
| `unassign ID` | Remove assignment |
| `agents` | Show sub-agent dashboard |
| `agents --stale N` | Set freshness window (minutes) |
| `list epics/stories/tasks/bugs` | List elements (with `--status`, `--story` filters) |
| `summary` | Board overview |
| `search "regex"` | Search board content |
| `validate` | Check board structure |
| `move ID --to PARENT` | Move element to different parent |
| `delete ID` | Delete element |

## Board Structure

```
.task-board/
├── system.md                     # Global counters
├── EPIC-01_auth-system/
│   ├── README.md                 # Description, scope, acceptance criteria
│   ├── progress.md               # Status, assignee, timestamps, dependencies, checklist
│   └── STORY-01_login-flow/
│       ├── README.md
│       ├── progress.md
│       ├── TASK-01_jwt-tokens/
│       │   ├── README.md
│       │   └── progress.md
│       └── BUG-01_token-expiry/
│           ├── README.md
│           └── progress.md
└── .temp/                        # Rendered graphs (gitignored)
```

## Tools Used

| Tool | Purpose | Commands | Output |
|------|---------|----------|--------|
| Go | CLI binary build/test | `go build`, `go test ./...` | `tools/board-cli/task-board` |
| Graphviz | Graph rendering | `dot -Tpng -o out.png in.dot` | `.task-board/**/.temp/plan*.{svg,png}` |
| Homebrew | Dependency management | `brew install go graphviz` | — |

## Project Files

| File | Description |
|------|-------------|
| `SKILL.md` | Full skill specification — CLI reference, file formats, workflows, agent development cycle |
| `SPEC.md` | Product requirements — planner (R1-R5), agent tracking (R6-R7) |
| `CLAUDE.md` | Claude Code guidance — build/test commands, architecture |
| `scripts/setup.sh` | One-command setup script |
| `tools/board-cli/` | Go source code for `task-board` CLI |

## Architecture

```
tools/board-cli/
├── main.go              # Entry point
├── cmd/                 # Cobra commands (create, plan, agents, link, etc.)
├── internal/
│   ├── board/           # Core domain: board loader, elements, progress, dependencies
│   ├── plan/            # Planner: graph builder, toposort, DOT generator, renderer
│   └── output/          # Terminal formatting: colored tables, status badges
└── templates/           # Embedded Go templates for README.md and progress.md
```
