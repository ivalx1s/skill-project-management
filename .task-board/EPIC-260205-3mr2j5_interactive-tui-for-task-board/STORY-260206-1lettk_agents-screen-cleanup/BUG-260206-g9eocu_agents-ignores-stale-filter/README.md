# BUG-260206-g9eocu: agents-ignores-stale-filter

## Description
In agents.go loadAgentsWithStale(), --all flag is always passed regardless of staleMinutes value. This overrides the --stale filter from settings. When staleMinutes > 0, should NOT pass --all. Fix: only add --all when staleMinutes == 0.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
(define fix acceptance criteria)
