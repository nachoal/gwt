# gwt - Git Worktree Manager

## Project Overview
A beautiful CLI tool for managing git worktrees with automatic setup and configuration. Built with Go using Bubble Tea for delightful terminal UIs.

## Architecture
- **Language**: Go 1.22+
- **CLI Framework**: Cobra
- **TUI Framework**: Bubble Tea + Lipgloss
- **Config Format**: YAML

## Key Features
- Creates worktrees in organized structure: `~/git-worktrees/<project>/<branch>`
- Automatically copies specified files (credentials, .env, etc.) to new worktrees
- Runs setup commands after creation (npm install, etc.)
- Beautiful TUI with progress indicators and interactive lists
- Shell integration for easy directory switching
- Cleanup of merged branches

## Project Structure
```
gwt/
├── cmd/                  # CLI commands
│   ├── root.go          # Root command setup
│   ├── new.go           # Create worktree
│   ├── list.go          # List worktrees
│   ├── remove.go        # Remove worktree
│   ├── switch.go        # Switch worktree (shell)
│   ├── clean.go         # Clean merged worktrees
│   ├── init.go          # Initialize config
│   └── shell.go         # Shell integration
├── internal/
│   ├── config/          # Configuration management
│   ├── worktree/        # Git worktree operations
│   └── ui/              # Bubble Tea UI components
│       ├── create.go    # Creation progress UI
│       ├── list.go      # Interactive list UI
│       └── styles.go    # Shared Lipgloss styles
└── .worktree.yaml       # Example config file
```

## Commands
- `gwt init` - Create .worktree.yaml config
- `gwt new <branch>` - Create new worktree
- `gwt list` - Show interactive worktree list
- `gwt switch <branch>` - Change directory to worktree
- `gwt remove <branch>` - Delete worktree
- `gwt clean` - Remove merged worktrees

## Configuration (.worktree.yaml)
```yaml
version: 1
copy:
  - .env
  - credentials/
setup:
  - npm install
settings:
  root: ~/git-worktrees
  auto_clean_merged: true
  confirm_delete: true
```

## Development

### Building
```bash
go build -o gwt .
```

### Testing
```bash
go test ./...
```

### Linting
```bash
golangci-lint run
```

## Design Decisions
1. **Bubble Tea for TUI**: Provides beautiful, interactive terminal UIs with minimal code
2. **YAML Configuration**: Human-readable and widely supported
3. **Organized Directory Structure**: Prevents conflicts between projects
4. **Automatic Setup**: Reduces manual steps when creating worktrees
5. **Shell Integration**: Optional but enables seamless directory switching

## Future Enhancements
- Profile support for different branch patterns (feature/*, hotfix/*)
- Git status integration in list view
- Worktree templates
- Integration with git hooks
- Parallel setup command execution