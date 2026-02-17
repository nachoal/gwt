# gwt - Git Worktree Manager

A beautiful and simple CLI tool for managing git worktrees with automatic setup and configuration.

## Features

- 🎯 Simple commands: `new`, `list`, `switch`, `remove`, `clean`
- 🎨 Beautiful TUI with progress indicators
- 🤖 Agent-friendly non-interactive fallback when no TTY is available
- 🧾 Machine-friendly output flags: `--no-tui`, `--plain`, `--json`
- ⚡ Auto-copy files (.env, credentials, etc.) to new worktrees
- 🔧 Run setup commands automatically (npm install, etc.)
- 📁 Organized structure: `~/git-worktrees/<project>/<branch>`
- 🧹 Clean up merged branches easily

## Installation

```bash
go install github.com/nachoal/gwt@latest
```

## Quick Start

1. Initialize config in your project:
```bash
gwt init
```

2. Create a new worktree:
```bash
gwt new feature/awesome
```

3. List worktrees:
```bash
gwt list
```

4. Enable shell integration (add to ~/.zshrc):
```bash
# One-time install (zsh/bash):
gwt shell --install
# Writes a tiny source block to your rc and helper file under ~/.config/gwt/
# (for example ~/.config/gwt/shell.zsh)
```

5. Switch between worktrees:
```bash
gwt sw feature/awesome
```

## Configuration

Edit `.worktree.yaml` in your project:

```yaml
version: 1

# Files to copy from main worktree
copy:
  - .env
  - .env.local
  - credentials/

# Commands to run after creation
setup:
  - npm install
  - npm run prepare

settings:
  root: ~/git-worktrees
  auto_clean_merged: true
  confirm_delete: true
```

## Commands

- `gwt init` - Initialize config file
- `gwt new <branch>` - Create a new worktree (`--no-tui`, `--plain`, `--json`)
- `gwt list` - Show worktrees (`--no-tui`, `--plain`, `--json`)
- `gwt switch <branch>` - Change to worktree directory
- `gwt remove <branch>` - Delete a worktree
- `gwt done [branch] [base]` - Update base and remove the branch worktree
- `gwt clean` - Remove merged worktrees
- `gwt version` - Show version/build metadata and executable path
- `gwt -v` / `gwt --version` - Short version output

Binary sanity check:
- `which -a gwt` to find duplicate installations in `PATH`
- `gwt version --json` to confirm the executable path that actually ran

By default, `gwt new` and `gwt list` use TUI only when interactive TTY is available; otherwise they automatically fall back to non-TUI output.

With shell integration enabled, extra quality-of-life helpers are available:
- `gwt new feature/foo -c` → after creation, cd to the new worktree and run your `claude` alias
- `gwt new feature/foo -c "plan the changes"` → runs `claude "plan the changes"`
- `gwt new feature/foo -c issue https://link` → runs `claude "/issue-analysis https://link"`
- `gwt done [branch] [base]` → runs the real CLI command and then cd's to the base worktree
  - Tip: When run inside a worktree, `gwt done` can be used with no args; it infers the current branch and default base.

Tip: If you previously had an alias named `gwt`, the installed function safely overrides it.
