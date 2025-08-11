# Repository Guidelines

## Project Structure & Module Organization
- Go 1.22 module; entrypoint `main.go`.
- CLI commands live in `cmd/` (e.g., `new.go`, `list.go`, `switch.go`, `remove.go`, `clean.go`, `init.go`, `shell.go`).
- Core logic in `internal/`: `worktree/` (git worktree ops), `config/` (YAML config), `ui/` (Bubble Tea TUI).
- Example config: `.worktree.yaml`. Local builds produce `./gwt`.

## Build, Test, and Development Commands
- Build: `go build -o gwt .` (outputs `./gwt`).
- Run locally: `go run . list` or `go run . new feature/foo`. Install for global use: `go install github.com/nachoal/gwt@latest`.
- Tests: `go test ./...`; coverage: `go test -cover ./...`.
- Static checks: `go vet ./...`; optional lint: `golangci-lint run` (if installed).

## Coding Style & Naming Conventions
- Format: `gofmt -s -w .` (and `goimports -w .` if available). Code must be gofmt-clean.
- Packages: lower-case, no underscores (e.g., `internal/worktree`).
- Files: concise, action-oriented names (e.g., `clean.go`, `switch.go`).
- Exports: CamelCase; errors via `fmt.Errorf` with `%w` for wrapping; avoid global state.

## Testing Guidelines
- Use the standard `testing` package; place tests next to code as `*_test.go` (e.g., `internal/worktree/worktree_test.go`).
- Prefer table-driven tests; cover pure logic. For code that shells out, isolate helpers and unit test them; keep end-to-end tests minimal.
- Run `go test ./...` (and optionally `-cover`) before opening a PR.

## Commit & Pull Request Guidelines
- Conventional Commits (as in history): `feat: …`, `fix: …`, `docs: …`.
- PRs include: clear description (What/Why), linked issues (`#123`), and screenshots or terminal snippets for TUI changes.
- Require: passing `go test ./...` and `go vet ./...`; formatted code; docs updated (`README.md`/`CLAUDE.md`) when behavior changes.

## Security & Configuration Tips
- Do not commit secrets; `.env*` are ignored by `.gitignore`. Keep `.worktree.yaml` free of sensitive values or share example-only snippets.
- Shell integration: preview with `gwt shell`; add via `eval "$(gwt shell)"` in your shell rc without committing shell-specific changes.

