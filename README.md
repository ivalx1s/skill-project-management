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

Installs Go (via Homebrew), builds both binaries, symlinks to `~/.local/bin/`:
- `task-board` — CLI tool
- `task-board-tui` — interactive TUI dashboard

### Manual Build

```bash
# CLI
cd tools/board-cli
go build -o task-board .

# TUI
cd tools/board-tui
go build -o task-board-tui .
```

### Requirements

- **Go** 1.21+
- **Graphviz** (for `--render` graph output) — `brew install graphviz`

## AI Agent Skill Setup

This repo is an AI agent skill compatible with coding agents (Claude Code, Codex CLI, and similar tools).

### With `~/.agents/` infrastructure

If you use [alexis-agents-infra](https://github.com/anthropics/alexis-agents-infra) for managing global instructions and skills:

```bash
# Clone this repo
git clone <repo-url> ~/src/skill-project-management

# Create external skills directory (not tracked by agents-infra)
mkdir -p ~/.agents/skills

# Symlink into external skills
ln -s ~/src/skill-project-management ~/.agents/skills/project-management

# Symlink to agent tools
ln -s ~/.agents/skills/project-management ~/.claude/skills/project-management  # Claude Code (CLAUDE.md)
ln -s ~/.agents/skills/project-management ~/.codex/skills/project-management   # Codex CLI (AGENTS.md)
```

The `~/.agents/` pattern:
- `~/.agents/.skills/` — skills bundled with agents-infra (auto-symlinked by `setup-symlinks.sh`)
- `~/.agents/skills/` — external skills from separate repos (manual symlinks, gitignored)

### Direct setup (without agents-infra)

```bash
# Clone and symlink directly
git clone <repo-url> ~/src/skill-project-management
mkdir -p ~/.claude/skills ~/.codex/skills
ln -s ~/src/skill-project-management ~/.claude/skills/project-management
ln -s ~/src/skill-project-management ~/.codex/skills/project-management
```

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
| `tui` | Launch interactive TUI dashboard |
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
| Go | CLI + TUI build/test | `go build`, `go test ./...` | `tools/board-cli/task-board`, `tools/board-tui/task-board-tui` |
| Graphviz | Graph rendering | `dot -Tpng -o out.png in.dot` | `.task-board/**/.temp/plan*.{svg,png}` |
| Homebrew | Dependency management | `brew install go graphviz` | — |

## Project Files

| File | Description |
|------|-------------|
| `SKILL.md` | Full skill specification — CLI reference, file formats, workflows, agent development cycle |
| `SPEC.md` | Product requirements — planner (R1-R5), agent tracking (R6-R7) |
| `CLAUDE.md` | Agent guidance — build/test commands, architecture |
| `scripts/setup.sh` | One-command setup script |
| `tools/board-cli/` | Go source code for `task-board` CLI |
| `tools/board-tui/` | Go source code for `task-board-tui` interactive dashboard |

## Architecture

```
tools/board-cli/                 # CLI (task-board)
├── main.go              # Entry point
├── cmd/                 # Cobra commands (create, plan, agents, tui, link, etc.)
├── internal/
│   ├── board/           # Core domain: board loader, elements, progress, dependencies
│   ├── plan/            # Planner: graph builder, toposort, DOT generator, renderer
│   └── output/          # Terminal formatting: colored tables, status badges
└── templates/           # Embedded Go templates for README.md and progress.md

tools/board-tui/                 # TUI dashboard (task-board-tui)
├── main.go              # Entry point, bubbletea model, screen routing
├── board.go             # Board screen: row model, navigation, rendering
├── agents.go            # Agents dashboard screen
├── detail.go            # Element detail view with markdown rendering
├── settings.go          # Settings screen (refresh rate, agents filter, scroll sensitivity)
├── command.go           # Command palette (slash commands)
├── mouse.go             # Mouse/trackpad scroll with sensitivity accumulator
├── config.go            # Persisted config (~/.config/board-tui/config.json)
├── tree.go              # Tree data model, expand/collapse, flatten
└── logger.go            # Session logger
```
