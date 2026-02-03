# TODO — Ideas Backlog

Ideas to digest, discuss, and maybe implement. Or maybe not.

---

## 1. Git worktree per task

**Idea:** Присрать worktree под каждую таску. Когда агент берёт таску — автоматом создаётся git worktree, агент работает в изолированном бранче. Научить правильно форкать/мержить обратно.

**Why:** Параллельная работа нескольких агентов без конфликтов. Каждый в своём worktree, каждый в своём бранче. Мерж — контролируемый.

**Open questions:**
- Naming convention для бранчей (e.g. `task/TASK-42_do-stuff`)?
- Автоматический merge обратно или ручной?
- Cleanup worktree после завершения таски?
- Base branch — всегда main или от parent story?

---

## 2. Independent ID generation (no system.md)

**Idea:** Идентификатор элемента не должен жить в `system.md` борды. Сейчас глобальный auto-increment counter — single point of contention при параллельной разработке. Два агента одновременно создают таски → race condition на system.md.

**Why:** При параллельной работе (особенно с worktrees из идеи #1) system.md будет постоянно конфликтовать. Нужен ID, который можно сгенерить независимо.

**Options to consider:**
- UUID-based (ugly but unique)
- Hash-based (e.g. short hash от timestamp + random)
- Distributed counter (each worktree has own range?)
- Content-addressable (hash от title + parent?)
- Hybrid: human-readable prefix + unique suffix (e.g. `TASK-a3f2`)

**Open questions:**
- Сохранять ли человекочитаемость (`TASK-12` vs `TASK-a3f2bc`)?
- Backward compatibility с текущими ID?
- Как резолвить если всё-таки коллизия?

---
