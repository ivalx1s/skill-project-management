# EPIC-03: planner

## Description
Планировщик — CLI команды для построения графа зависимостей, фаз и визуализации

## Scope
R1-R5 из SPEC.md Part 1: иерархический граф, топосорт/фазы, вывод плана, рендер Graphviz, детекция проблем

## Acceptance Criteria
- task-board plan выводит фазы на любом уровне иерархии
- link/unlink эскалирует зависимости автоматом
- plan --render генерирует SVG через Graphviz
- Циклы детектятся, critical path показывается
- plan --save пишет plan.md
