# CLI JSON API Specification

## Overview

Machine-readable JSON output for task-board CLI commands. Enables integration with TUI, scripts, and external tools.

## Global Flag

All commands supporting JSON output use the `--json` flag:

```bash
task-board <command> --json
```

When `--json` is set:
- Output is valid JSON to stdout
- Errors are JSON objects to stderr
- Exit codes remain the same (0 = success, 1 = error)

---

## Error Format

All errors follow the same schema:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Element TASK-999 not found",
    "details": {}
  }
}
```

**Error Codes:**
- `NOT_FOUND` — element doesn't exist
- `INVALID_ID` — malformed ID format
- `INVALID_STATUS` — unknown status value
- `CYCLE_DETECTED` — dependency would create cycle
- `VALIDATION_ERROR` — board structure invalid
- `INTERNAL_ERROR` — unexpected error

---

## Commands

### list

List board elements by type.

```bash
task-board list epics --json
task-board list stories --epic EPIC-001 --json
task-board list tasks --story STORY-001 --status development --json
```

**Response:**

```json
{
  "elements": [
    {
      "id": "TASK-260205-abc123",
      "type": "task",
      "name": "Implement feature X",
      "status": "development",
      "assignee": "agent-builder",
      "parent": "STORY-260205-xyz789",
      "path": "EPIC-260205-foo/STORY-260205-xyz/TASK-260205-abc",
      "createdAt": "2025-02-05T10:00:00Z",
      "updatedAt": "2025-02-05T13:30:00Z",
      "blockedBy": ["TASK-260205-other"],
      "blocks": []
    }
  ],
  "count": 1,
  "filters": {
    "type": "task",
    "story": "STORY-260205-xyz789",
    "status": "development"
  }
}
```

---

### show

Show full element details.

```bash
task-board show TASK-260205-abc123 --json
```

**Response:**

```json
{
  "element": {
    "id": "TASK-260205-abc123",
    "type": "task",
    "name": "Implement feature X",
    "status": "development",
    "assignee": "agent-builder",
    "parent": "STORY-260205-xyz789",
    "path": "EPIC-260205-foo/STORY-260205-xyz/TASK-260205-abc",
    "createdAt": "2025-02-05T10:00:00Z",
    "updatedAt": "2025-02-05T13:30:00Z",
    "blockedBy": ["TASK-260205-other"],
    "blocks": [],
    "description": "Full markdown description...",
    "acceptanceCriteria": "- [ ] Criterion 1\n- [ ] Criterion 2",
    "checklist": [
      {"text": "Step 1", "done": true},
      {"text": "Step 2", "done": false}
    ],
    "notes": [
      {"timestamp": "2025-02-05T12:00:00Z", "text": "Started work"}
    ]
  }
}
```

---

### summary

Show board summary statistics.

```bash
task-board summary --json
```

**Response:**

```json
{
  "summary": {
    "byType": {
      "epic": {"total": 9, "todo": 2, "active": 0, "done": 7, "closed": 0, "blocked": 0},
      "story": {"total": 24, "todo": 5, "active": 0, "done": 19, "closed": 0, "blocked": 0},
      "task": {"total": 77, "todo": 13, "active": 0, "done": 64, "closed": 0, "blocked": 0},
      "bug": {"total": 0, "todo": 0, "active": 0, "done": 0, "closed": 0, "blocked": 0}
    },
    "active": [
      {"id": "TASK-001", "name": "...", "status": "development", "ancestry": "EPIC-001 > STORY-001"}
    ],
    "blocked": [
      {"id": "STORY-002", "name": "...", "blockedBy": ["STORY-001"], "ancestry": "EPIC-001"}
    ]
  }
}
```

---

### search

Search board content.

```bash
task-board search "authentication" --json
```

**Response:**

```json
{
  "results": [
    {
      "id": "TASK-260205-abc123",
      "type": "task",
      "name": "Add authentication",
      "status": "backlog",
      "matchField": "name",
      "matchContext": "Add **authentication** to API"
    }
  ],
  "count": 1,
  "query": "authentication"
}
```

---

### plan

Show execution plan for an epic.

```bash
task-board plan EPIC-260205-abc123 --json
```

**Response:**

```json
{
  "plan": {
    "epicId": "EPIC-260205-abc123",
    "epicName": "Interactive TUI",
    "phases": [
      {
        "phase": 1,
        "description": "no dependencies",
        "elements": [
          {"id": "STORY-001", "name": "CLI JSON Output", "status": "backlog", "blockedBy": []}
        ]
      },
      {
        "phase": 2,
        "description": null,
        "elements": [
          {"id": "STORY-002", "name": "Board View", "status": "backlog", "blockedBy": ["STORY-001"]}
        ]
      }
    ],
    "criticalPath": ["STORY-001", "STORY-002", "STORY-003"],
    "criticalPathLength": 3
  }
}
```

---

### agents

Show agent dashboard.

```bash
task-board agents --json
task-board agents --stale 30 --json
```

**Response:**

```json
{
  "agents": [
    {
      "name": "agent-builder",
      "assignedElements": [
        {
          "id": "TASK-260205-abc123",
          "type": "task",
          "name": "Build feature",
          "status": "development",
          "updatedAt": "2025-02-05T13:00:00Z",
          "staleSince": null
        }
      ],
      "totalAssigned": 1,
      "staleCount": 0
    }
  ],
  "totalAgents": 1,
  "filters": {
    "staleMinutes": 30
  }
}
```

---

### validate

Validate board structure.

```bash
task-board validate --json
```

**Response (valid):**

```json
{
  "valid": true,
  "errors": [],
  "warnings": []
}
```

**Response (invalid):**

```json
{
  "valid": false,
  "errors": [
    {"code": "ORPHAN_ELEMENT", "message": "TASK-001 has no parent", "elementId": "TASK-001"},
    {"code": "CYCLE_DETECTED", "message": "Dependency cycle: A -> B -> A", "elementIds": ["A", "B"]}
  ],
  "warnings": [
    {"code": "STALE_ELEMENT", "message": "TASK-002 not updated in 7 days", "elementId": "TASK-002"}
  ]
}
```

---

### tree

New command: get full board as tree structure (optimized for TUI).

```bash
task-board tree --json
task-board tree --epic EPIC-001 --json
```

**Response:**

```json
{
  "tree": [
    {
      "id": "EPIC-260205-abc123",
      "type": "epic",
      "name": "Interactive TUI",
      "status": "backlog",
      "updatedAt": "2025-02-05T13:00:00Z",
      "children": [
        {
          "id": "STORY-260205-xyz789",
          "type": "story",
          "name": "Board View",
          "status": "backlog",
          "updatedAt": "2025-02-05T12:00:00Z",
          "children": [
            {
              "id": "TASK-260205-task01",
              "type": "task",
              "name": "Implement tree model",
              "status": "backlog",
              "assignee": null,
              "updatedAt": "2025-02-05T11:00:00Z",
              "children": []
            }
          ]
        }
      ]
    }
  ]
}
```

---

## Write Commands

Write commands return the created/modified element on success.

### create

```bash
task-board create task --name "New task" --story STORY-001 --json
```

**Response:**

```json
{
  "created": {
    "id": "TASK-260205-new123",
    "type": "task",
    "name": "New task",
    "status": "backlog",
    "parent": "STORY-001",
    "path": "..."
  }
}
```

### update, assign, progress, link, etc.

Similar pattern — return affected element(s):

```json
{
  "updated": { ... },
  "message": "Status changed to development"
}
```

---

## Implementation Priority

1. **Phase 1** (TUI MVP)
   - `list --json`
   - `tree --json` (new command)
   - `show --json`

2. **Phase 2** (Full read API)
   - `summary --json`
   - `plan --json`
   - `agents --json`
   - `search --json`
   - `validate --json`

3. **Phase 3** (Write API)
   - `create --json`
   - `update --json`
   - `assign --json`
   - `progress --json`
   - `link --json`
   - `delete --json`
