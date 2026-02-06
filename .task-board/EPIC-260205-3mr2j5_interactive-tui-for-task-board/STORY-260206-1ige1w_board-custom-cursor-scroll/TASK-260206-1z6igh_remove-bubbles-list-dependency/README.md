# TASK-260206-1z6igh: remove-bubbles-list-dependency

## Description
Remove bubbles/list from main.go. Delete: list.Model field, list.New() in main(), BoardItem struct with Title/Description/FilterValue, refreshList using list.SetItems, selectNodeByID using list.Items, all m.list references. Remove bubbles/list import. Clean go.mod if no other usage.

## Scope
(define task scope)

## Acceptance Criteria
(define acceptance criteria)
