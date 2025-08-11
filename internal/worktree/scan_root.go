package worktree

import (
    "os"
    "os/exec"
    "path/filepath"
    "sort"
    "strings"

    "github.com/nachoal/gwt/internal/config"
)

// RootItem represents a worktree discovered under the configured root.
type RootItem struct {
    Project string
    Branch  string
    Path    string
    Head    string
}

// ListFromRoot scans the gwt root (or override) for worktrees and returns a flat list.
// It treats a directory as a worktree iff the entry ".git" exists and is a file
// (not a directory), which is how git worktrees are represented.
// It also attempts to read the current branch and HEAD short SHA via git.
func ListFromRoot(override string) ([]RootItem, string, error) {
    root := override
    if root == "" {
        cfg, err := config.LoadConfig()
        if err != nil {
            return nil, "", err
        }
        root = cfg.Settings.Root
    } else {
        // Expand leading ~/
        if strings.HasPrefix(root, "~/") {
            if home, err := os.UserHomeDir(); err == nil {
                root = filepath.Join(home, root[2:])
            }
        }
    }

    absRoot, err := filepath.Abs(root)
    if err == nil {
        root = absRoot
    }

    if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
        return []RootItem{}, root, nil
    }

    projects, _ := os.ReadDir(root)
    var items []RootItem

    for _, p := range projects {
        if !p.IsDir() {
            continue
        }
        projPath := filepath.Join(root, p.Name())
        // BFS/stack to find leaf dirs that are worktrees (detect .git file at top level)
        stack := []string{projPath}
        for len(stack) > 0 {
            // pop
            d := stack[len(stack)-1]
            stack = stack[:len(stack)-1]

            // Check for worktree marker: .git file (not directory)
            gitPath := filepath.Join(d, ".git")
            if info, err := os.Lstat(gitPath); err == nil {
                if !info.IsDir() { // file => worktree
                    // Collect details
                    branch := readBranch(d)
                    if branch == "HEAD" || branch == "" {
                        // Fallback from path (relative to project)
                        rel := strings.TrimPrefix(d, projPath+string(os.PathSeparator))
                        branch = filepath.ToSlash(rel)
                    }
                    head := readHead(d)
                    items = append(items, RootItem{
                        Project: p.Name(),
                        Branch:  branch,
                        Path:    d,
                        Head:    head,
                    })
                    continue // do not descend into a worktree
                }
                // If .git is a directory, it's a full repo clone; skip descending
                continue
            }

            // .git not present here; descend into subdirectories
            entries, err := os.ReadDir(d)
            if err != nil {
                continue
            }
            for _, e := range entries {
                if !e.IsDir() {
                    continue
                }
                name := e.Name()
                if strings.HasPrefix(name, ".") { // skip hidden dirs
                    continue
                }
                // Skip common heavy dirs if ever encountered before we hit a worktree (defensive)
                if name == "node_modules" || name == "vendor" || name == "dist" || name == "build" {
                    continue
                }
                // Do not follow symlinks
                if info, err := e.Info(); err == nil && (info.Mode()&os.ModeSymlink) != 0 {
                    continue
                }
                stack = append(stack, filepath.Join(d, name))
            }
        }
    }

    // Sort by project then branch
    sort.Slice(items, func(i, j int) bool {
        if items[i].Project == items[j].Project {
            return items[i].Branch < items[j].Branch
        }
        return items[i].Project < items[j].Project
    })

    return items, root, nil
}

func readBranch(dir string) string {
    // git rev-parse --abbrev-ref HEAD
    cmd := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD")
    out, err := cmd.Output()
    if err != nil {
        return ""
    }
    return strings.TrimSpace(string(out))
}

func readHead(dir string) string {
    cmd := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD")
    out, err := cmd.Output()
    if err != nil {
        return ""
    }
    return strings.TrimSpace(string(out))
}

