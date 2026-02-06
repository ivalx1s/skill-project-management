# TASK-260206-a9sxhx: show-all-assigned-statuses

## Description
TUI agents screen should show ALL assigned elements regardless of status — backlog, analysis, to-dev, development, to-review, reviewing, done, closed, blocked. Currently uses default CLI behavior which hides done/closed older than 30min. Fix: always pass --all flag when loading agents data. User wants to see everything that has an assignee, in any status (including analysis).

## Scope
(define task scope)

## Acceptance Criteria
(define acceptance criteria)
