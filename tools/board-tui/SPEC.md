# Board TUI Specification

## Overview

Terminal User Interface (TUI) for task-board. Provides interactive view of the board for human supervisors while CLI remains the primary interface for agents and automation.

## Target User

Human supervisor who:
- Monitors board state and progress
- Observes agent activity
- Needs quick overview without running multiple CLI commands

## Technology

- **Language**: Go
- **Framework**: Bubbletea (Charm)
- **Styling**: Lipgloss
- **Data Source**: Calls `task-board` CLI for data

---

## Screens

### 1. Board View (Main)

Primary screen showing the task board hierarchy.

**Layout:**
```
┌─────────────────────────────────────────────────────────┐
│  Task Board                                             │
├─────────────────────────────────────────────────────────┤
│  ▼ EPIC-001 Interactive TUI for task-board    backlog   │
│    ▼ STORY-001 Bubbletea prototype          development │
│        TASK-001 Setup Go module                   done  │
│        TASK-002 Implement board loader     development  │
│        TASK-003 Build tree navigation          backlog  │
│    ▶ STORY-002 Settings screen                 backlog  │
│  ▶ EPIC-002 Another epic                       backlog  │
│                                                         │
├─────────────────────────────────────────────────────────┤
│  Board │ ↑↓ nav │ ←→ collapse/expand │ s settings │ q quit │ Updated 5s ago │
└─────────────────────────────────────────────────────────┘
```

**Features:**
- Disclosure list (tree with expand/collapse)
- Hierarchy: Epic → Story → Task/Bug
- Sorting: newest first (by Last Update timestamp)
- Columns: ID, Name, Status, Assignee (if set)

**Interactions:**
- `▼` expanded node (children visible)
- `▶` collapsed node (children hidden)
- Enter/→/l expands or focuses item
- ←/h collapses or goes to parent

### 2. Settings Screen

Configuration screen inspired by Claude Code settings.

**Layout:**
```
┌─────────────────────────────────────────────────────────┐
│  Settings                                               │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Refresh Rate                                           │
│  ● 5 seconds                                            │
│  ○ 10 seconds (default)                                 │
│  ○ 30 seconds                                           │
│  ○ 60 seconds                                           │
│  ○ Off (manual only)                                    │
│                                                         │
├─────────────────────────────────────────────────────────┤
│  Settings │ ↑↓ nav │ Enter select │ Esc back │          │
└─────────────────────────────────────────────────────────┘
```

**Settings:**
- **Refresh Rate**: Auto-refresh interval for board data
  - Options: 5s, 10s (default), 30s, 60s, Off
  - Persisted to config file (`~/.config/board-tui/config.json`)

---

## Live Watch

Board data automatically refreshes at configured interval.

**Behavior:**
- Timer runs in background
- On tick: re-fetch data from CLI, update view
- Status bar shows time since last update
- Manual refresh available via `r` key

**Implementation:**
- Use bubbletea's `tea.Tick` for interval
- Call `task-board list --json` (requires CLI support)
- Gracefully handle CLI errors (show stale data with warning)

---

## Keyboard Navigation

### Global

| Key | Action |
|-----|--------|
| `q` | Quit application |
| `Esc` | Back / Quit |
| `Ctrl+C` | Force quit |
| `s` | Open Settings |
| `r` | Force refresh |
| `?` | Show help |

### Board View

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `Enter` / `l` / `→` | Expand node / Enter details |
| `h` / `←` | Collapse node / Go to parent |
| `g` | Go to top |
| `G` | Go to bottom |
| `/` | Search / Filter |

### Settings View

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `Enter` / `Space` | Select option |
| `Esc` | Back to Board |

---

## Status Bar

Bottom bar showing context and controls.

**Sections:**
- **Left**: Current screen name ("Board" / "Settings")
- **Center**: Key hints for current context
- **Right**: Update status ("Updated 5s ago" / "Refreshing..." / "Offline")

---

## Data Flow

```
┌─────────────┐      exec       ┌─────────────┐      read       ┌─────────────┐
│  board-tui  │ ──────────────► │ task-board  │ ──────────────► │ .task-board │
│   (TUI)     │ ◄────────────── │   (CLI)     │                 │  (files)    │
└─────────────┘      JSON       └─────────────┘                 └─────────────┘
```

**CLI Requirements:**
- `task-board list --json` — returns all elements as JSON array
- JSON schema per element:
  ```json
  {
    "id": "TASK-001",
    "type": "task",
    "name": "Setup Go module",
    "status": "done",
    "assignee": "agent-builder",
    "parent": "STORY-001",
    "lastUpdate": "2025-02-05T13:00:00Z",
    "children": []
  }
  ```

---

## Configuration

**File**: `~/.config/board-tui/config.json`

```json
{
  "refreshRate": 10,
  "expandedNodes": ["EPIC-001", "STORY-001"]
}
```

**Fields:**
- `refreshRate`: Refresh interval in seconds (0 = off)
- `expandedNodes`: List of node IDs to expand on startup (preserves state)

---

## Future Considerations (Out of Scope v1)

- Detail view for single element (README, progress, notes)
- Quick actions (change status, add note)
- Agent dashboard (who's working on what)
- Dependency graph visualization
- Themes / color schemes
- Mouse support
