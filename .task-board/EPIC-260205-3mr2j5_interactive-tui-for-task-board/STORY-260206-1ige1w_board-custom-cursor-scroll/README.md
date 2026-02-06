# STORY-260206-1ige1w: board-custom-cursor-scroll

## Description
Replace bubbles/list on the main board screen with the same custom cursor+scroll approach used in agents screen. Current board uses bubbles/list which intercepts j/k keys and has its own scroll logic. Custom approach gives: row data model (flat rows from tree), manual cursor tracking (selectedIdx), viewport scroll that follows cursor (scrollOff + ensureVisible), highlight on selected row, arrow keys + trackpad. This unifies navigation UX across board and agents screens and removes bubbles/list dependency for the board view.

## Scope
(define story scope)

## Acceptance Criteria
(define acceptance criteria)
