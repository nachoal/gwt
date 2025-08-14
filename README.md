# gwt - Git Worktree Manager

A beautiful and simple CLI tool for managing git worktrees with automatic setup and configuration.

## Features

- 🎯 Simple commands: `new`, `list`, `switch`, `remove`, `clean`
- 🎨 Beautiful TUI with progress indicators
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
# Or, manual:
# echo 'eval "$(gwt shell)"' >> ~/.zshrc
# Then reload your shell or: source ~/.zshrc
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
- `gwt new <branch>` - Create new worktree
- `gwt list` - Show all worktrees
- `gwt switch <branch>` - Change to worktree directory
- `gwt remove <branch>` - Delete a worktree
- `gwt clean` - Remove merged worktrees

With shell integration enabled, quality-of-life helpers are available:
- `gwt new feature/foo -c` → after creation, cd to the new worktree and run your `claude` alias
- `gwt new feature/foo -c "plan the changes"` → runs `claude "plan the changes"`
- `gwt new feature/foo -c issue https://link` → runs `claude "/issue-analysis https://link"`
- `gwt done <branch> [base]` → cd to base, `git pull --ff-only`, then remove the `<branch>` worktree
  - Tip: When run inside a worktree, `gwt done` can be used with no args; it infers the current branch and default base.

Tip: If you previously had an alias named `gwt`, the installed function safely overrides it. If you prefer manual install, ensure the `eval "$(gwt shell)"` line appears after your aliases in your rc.
