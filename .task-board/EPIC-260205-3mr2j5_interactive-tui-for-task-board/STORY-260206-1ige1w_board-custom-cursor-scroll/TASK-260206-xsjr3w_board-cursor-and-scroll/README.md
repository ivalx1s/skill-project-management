# TASK-260206-xsjr3w: board-cursor-and-scroll

## Description
Replace bubbles/list cursor with custom selectedIdx + scrollOff. Implement moveUp/moveDown (skip non-selectable rows), goTop (g), goBottom (G), ensureVisible, visibleHeight. Mouse wheel: CursorUp/Down → custom moveUp/moveDown. Replace m.list.Select() calls with direct selectedIdx manipulation. Keep selected node ID across rebuilds (same pattern as agents buildRows).

## Scope
(define task scope)

## Acceptance Criteria
(define acceptance criteria)
