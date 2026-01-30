# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

A file-based project management skill for AI agents. Provides a Go CLI tool (`task-board`) that manages hierarchical task boards (Epic → Story → Task/Bug) stored as directories and markdown files in `.task-board/`. Includes a planner (dependency graph, phases, Graphviz rendering) and agent tracking (assign, dashboard). Everything lives in git, no external dependencies.

## Setup

```bash
./scripts/setup.sh
```

Installs Go (via Homebrew), builds the binary, symlinks `task-board` to `~/.local/bin/`.

## Build & Test

All commands run from `tools/board-cli/`:

```bash
# Build
go build -o task-board .

# Run all tests
go test ./...

# Run tests verbose
go test ./... -v

# Run a single test
go test ./cmd -run TestCreateEpic -v

# Lint
go fmt ./...
go vet ./...
```

## Architecture

```
tools/board-cli/
├── main.go              # Entry point, calls cmd.Execute()
├── cmd/                 # Cobra command handlers (one file per command)
│   ├── create.go        # create epic/story/task/bug
│   ├── plan.go          # plan [ID] with --render, --layout, --active, --save
│   ├── link.go          # link with dependency escalation
│   ├── unlink.go        # unlink with de-escalation
│   ├── assign.go        # assign element to agent
│   ├── unassign.go      # unassign element
│   ├── agents.go        # agents dashboard with --stale, --all
│   ├── progress.go      # status, checklist, notes
│   └── ...              # list, show, summary, search, validate, move, delete, update
├── internal/
│   ├── board/           # Core domain logic
│   │   ├── board.go     # Board loader — walks .task-board/ filesystem, builds element tree
│   │   ├── element.go   # Element type (Epic/Story/Task/Bug), status, hierarchy
│   │   ├── naming.go    # Directory name parser (TYPE-NN_kebab-name format)
│   │   ├── system.go    # Global auto-increment counters (system.md)
│   │   ├── readme.go    # README.md parser/generator (title, description, scope, AC)
│   │   └── progress.go  # progress.md parser (status, assignee, timestamps, dependencies, checklist, notes)
│   ├── plan/            # Planner engine
│   │   ├── graph.go     # BuildGraph, BuildPlan (Kahn's toposort), critical path, AllDescendants
│   │   ├── dot.go       # GenerateDOT (phases), GenerateFullDOT (hierarchy), legend, status colors
│   │   ├── render.go    # RenderDOT (Graphviz invocation), RenderOutputPath (.temp/ placement)
│   │   └── problems.go  # Cycle detection, critical path formatting
│   └── output/
│       └── format.go    # Colorized terminal table rendering
└── templates/           # Embedded Go templates for README.md and progress.md per element type
```

**Key design decisions:**
- All state is file-based — the board is the `.task-board/` directory tree, read fresh on every command
- IDs are globally unique per type (`TASK-12` exists once across the entire board), tracked via `system.md` counters
- Dependencies are bidirectional — `link TASK-13 --blocked-by TASK-12` updates both elements' `progress.md`
- Dependency escalation — cross-parent links auto-create parent-level dependencies
- Each element is a directory (`TYPE-NN_kebab-name/`) containing `README.md` (static description) and `progress.md` (dynamic lifecycle journal)
- Templates use Go's `embed` to bundle markdown templates into the binary
- `progress.md` auto-updates `Last Update` timestamp on every write
- Graph rendering supports two layouts: `hierarchy` (organizational) and `phases` (execution order)

## Skill Documentation

`SKILL.md` — full skill specification: CLI reference, file formats, workflows, agent development cycle (spec → board → phases → sub-agents → supervision).

`SPEC.md` — product requirements: planner (R1-R5), agent tracking (R6-R7).
