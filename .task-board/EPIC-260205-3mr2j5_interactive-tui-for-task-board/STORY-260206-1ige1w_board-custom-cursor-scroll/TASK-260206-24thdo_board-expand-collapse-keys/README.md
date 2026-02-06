# TASK-260206-24thdo: board-expand-collapse-keys

## Description
Migrate expand/collapse keybindings from bubbles/list to custom rows. Space: toggle selected node. Right/l: expand if collapsed. Left/h: collapse if expanded, else go to parent. e: expand all. c: collapse all. On any expand/collapse, rebuild rows preserving selected node ID. Replace m.list.SelectedItem().(BoardItem) with direct node access via boardRow.

## Scope
(define task scope)

## Acceptance Criteria
(define acceptance criteria)
