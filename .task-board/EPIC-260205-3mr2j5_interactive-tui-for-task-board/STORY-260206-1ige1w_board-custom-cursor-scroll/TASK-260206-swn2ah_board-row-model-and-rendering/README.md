# TASK-260206-swn2ah: board-row-model-and-rendering

## Description
Create boardRow struct (like agentRow) from FlattenTree output. Each row has: kind (tree node = selectable), treePrefix, node reference, pre-rendered text with tree prefix + expand indicator + type indicator + ID + name + status + assignee. buildBoardRows() flattens tree via FlattenTree, maps FlatNode to boardRow. Render function outputs rows with cursor highlight (cursorStyle background on selectedIdx). Replace m.list.View() with custom rendering. Keep FlattenTree/FlatNode as-is.

## Scope
(define task scope)

## Acceptance Criteria
(define acceptance criteria)
