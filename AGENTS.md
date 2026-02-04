# Repository Guidelines

## Project Structure & Module Organization

- `tools/board-cli/`: Go module for the `task-board` CLI.
  - `cmd/`: Cobra command handlers (create/plan/link/agents/etc.).
  - `internal/`: core packages (`board/`, `plan/`, `output/`).
  - `templates/`: embedded Markdown templates for `README.md`/`progress.md`.
- `scripts/`: developer setup utilities (e.g., `scripts/setup.sh`).
- `.task-board/`: example board used to develop/validate the workflow and file formats.
- Docs: `README.md` (overview), `SKILL.md` (full skill spec), `SPEC.md` (requirements), `CLAUDE.md` (agent-oriented notes).

## Build, Test, and Development Commands

Run from the repo root unless noted:

- `./scripts/setup.sh`: installs Go via Homebrew (macOS), builds `task-board`, symlinks to `~/.local/bin/task-board`.
- `cd tools/board-cli && go build -o task-board .`: build a local binary (don’t commit it; it’s gitignored).
- `cd tools/board-cli && go test ./...`: run the full test suite.
- `cd tools/board-cli && go test ./cmd -run TestCreate -v`: run a focused test.
- `cd tools/board-cli && go fmt ./... && go vet ./...`: format + basic static checks.
- Optional: `brew install graphviz` for `task-board plan --render` (tests should not depend on Graphviz).

## Coding Style & Naming Conventions

- Go code should be `gofmt`-clean; rely on `go fmt` for indentation and imports.
- Prefer small, testable functions and table-driven tests.
- Board naming: IDs look like `TYPE-YYMMDD-xxxxxx` and directories like `TYPE-YYMMDD-xxxxxx_kebab-name/`
  (e.g., `TASK-260203-a1b2c3_fix-parser/`).

## Testing Guidelines

- Tests live next to code as `*_test.go` using the standard `testing` package.
- Use `t.TempDir()` for filesystem-heavy tests and avoid requiring network/external services.

## Commit & Pull Request Guidelines

- Commit messages in this repo are short, sentence-case imperatives (e.g., `Add …`, `Update …`, `Fix …`).
- PRs should include: purpose, CLI examples (commands + expected output), and tests for behavior changes.
- Update `README.md`/`SKILL.md` when changing user-facing flags, statuses, or file formats.
