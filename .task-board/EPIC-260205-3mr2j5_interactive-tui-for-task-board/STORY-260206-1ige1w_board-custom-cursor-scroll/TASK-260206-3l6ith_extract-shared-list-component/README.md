# TASK-260206-3l6ith: extract-shared-list-component

## Description
Refactoring: extract shared list UI component from board and agents screens. Both use the same pattern — flat rows, cursor (selectedIdx), scroll (scrollOff), ensureVisible, moveUp/moveDown, highlight on selected. Extract a reusable ListModel (or RowList) struct with: rows []Row interface, selectedIdx, scrollOff, width, height, moveUp/moveDown/goTop/goBottom/ensureVisible, visibleHeight, Render(). Board and agents embed this component and provide their own row building and row rendering.

## Scope
(define task scope)

## Acceptance Criteria
(define acceptance criteria)
