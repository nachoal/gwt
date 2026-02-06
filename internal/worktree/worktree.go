package worktree

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Worktree struct {
	Path   string
	Branch string
	Head   string
}

func GetProjectName() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	url := strings.TrimSpace(string(output))
	// Extract project name from URL
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid remote URL")
	}

	projectName := parts[len(parts)-1]
	projectName = strings.TrimSuffix(projectName, ".git")
	return projectName, nil
}

func GetDefaultBranch() (string, error) {
	// Try to get the default branch from git config
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.Output()
	if err == nil {
		// Extract branch name from refs/remotes/origin/main format
		branch := strings.TrimSpace(string(output))
		parts := strings.Split(branch, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1], nil
		}
	}

	// Fallback: check if common default branches exist
	commonDefaults := []string{"main", "master", "develop"}
	for _, branch := range commonDefaults {
		cmd := exec.Command("git", "rev-parse", "--verify", fmt.Sprintf("origin/%s", branch))
		if err := cmd.Run(); err == nil {
			return branch, nil
		}
	}

	// Last resort: use main
	return "main", nil
}

func GetWorktreePath(root, projectName, branchName string) string {
	return filepath.Join(root, projectName, branchName)
}

func Create(branchName, fromBranch, targetPath string) error {
	// Create parent directory if it doesn't exist
	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Check if branch exists
	checkCmd := exec.Command("git", "rev-parse", "--verify", branchName)
	if err := checkCmd.Run(); err == nil {
		// Branch exists, create worktree from it
		cmd := exec.Command("git", "worktree", "add", targetPath, branchName)
		return cmd.Run()
	}

	// Create new branch from base
	cmd := exec.Command("git", "worktree", "add", "-b", branchName, targetPath, fromBranch)
	return cmd.Run()
}

// FindMainWorktree returns the path to the main worktree (the original clone).
// It works from any worktree or the main repo itself.
func FindMainWorktree() (string, error) {
	// git rev-parse --git-common-dir gives the shared .git dir.
	// From that we can derive the main worktree root.
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	commonDir := strings.TrimSpace(string(out))
	if !filepath.IsAbs(commonDir) {
		cwd, _ := os.Getwd()
		commonDir = filepath.Join(cwd, commonDir)
	}
	// The common dir is typically <main-worktree>/.git
	// filepath.Dir gives us the main worktree root.
	mainWorktree := filepath.Dir(commonDir)
	return mainWorktree, nil
}

func List() ([]Worktree, error) {
	// Discover the main worktree so the command works from any worktree
	// or even when a worktree's state is broken.
	mainWT, err := FindMainWorktree()
	if err != nil {
		return nil, fmt.Errorf("not in a git repository: %w", err)
	}
	cmd := exec.Command("git", "-C", mainWT, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var worktrees []Worktree
	lines := strings.Split(string(output), "\n")

	var current Worktree
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			current = Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		} else if strings.HasPrefix(line, "branch ") {
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		} else if strings.HasPrefix(line, "HEAD ") {
			current.Head = strings.TrimPrefix(line, "HEAD ")
		}
	}
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}

func Remove(path string, force bool) error {
	// Run from the main worktree so that removing the current worktree
	// (i.e. the cwd) does not fail because git can't remove its own cwd.
	mainWT, _ := FindMainWorktree()

	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	cmd := exec.Command("git", args...)
	if mainWT != "" {
		cmd.Dir = mainWT
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
		}
		return err
	}
	return nil
}

// GetCommonGitDir returns the common git directory for the given worktree path.
func GetCommonGitDir(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-common-dir")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	dir := strings.TrimSpace(string(out))
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(path, dir)
	}
	return dir, nil
}

// DeleteBranchFromWorktreePath deletes the branch from the parent repository using the
// worktree's common git directory. If force is true, uses -D; else uses -d.
// If the branch does not exist, it returns nil.
func DeleteBranchFromWorktreePath(worktreePath, branch string, force bool) error {
	if branch == "" {
		return nil
	}
	common, err := GetCommonGitDir(worktreePath)
	if err != nil {
		return err
	}
	return DeleteBranchWithGitDir(common, branch, force)
}

// DeleteBranchWithGitDir deletes a branch using a known common git dir.
func DeleteBranchWithGitDir(commonGitDir, branch string, force bool) error {
	if branch == "" || commonGitDir == "" {
		return nil
	}
	// Check if branch exists
	check := exec.Command("git", "--git-dir", commonGitDir, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	if err := check.Run(); err != nil {
		// Not found
		return nil
	}
	args := []string{"--git-dir", commonGitDir, "branch"}
	if force {
		args = append(args, "-D", branch)
	} else {
		args = append(args, "-d", branch)
	}
	cmd := exec.Command("git", args...)
	return cmd.Run()
}

func CopyFiles(srcRoot, destRoot string, files []string) error {
	for _, file := range files {
		src := filepath.Join(srcRoot, file)
		dest := filepath.Join(destRoot, file)

		// Create destination directory
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(dest), err)
		}

		// Check if source is a directory
		info, err := os.Stat(src)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Skip if file doesn't exist
			}
			return err
		}

		if info.IsDir() {
			// Copy directory contents recursively.
			// Use src/. so that if dest already exists (e.g. from git checkout),
			// contents are merged into it instead of creating a nested subdirectory.
			cmd := exec.Command("cp", "-r", src+"/.", dest)
			cmd.Stderr = os.Stderr // Show error output
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to copy directory %s to %s: %w", src, dest, err)
			}
		} else {
			// Copy single file
			cmd := exec.Command("cp", src, dest)
			cmd.Stderr = os.Stderr // Show error output
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to copy file %s to %s: %w", src, dest, err)
			}
		}
	}
	return nil
}

func RunSetupCommands(workdir string, commands []string) error {
	for _, command := range commands {
		cmd := exec.Command("sh", "-c", command)
		cmd.Dir = workdir
		// Capture output to avoid interfering with TUI
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Include output in error for debugging
			return fmt.Errorf("failed to run '%s': %w\nOutput: %s", command, err, string(output))
		}
	}
	return nil
}

// RunSetupCommandsOpts runs setup commands with optional verbose output and timing.
// If verbose is true, command stdout/stderr are streamed to out.
// If timed is true, a short summary line is written showing the duration per command.
func RunSetupCommandsOpts(workdir string, commands []string, verbose bool, timed bool, out io.Writer) error {
	if out == nil {
		out = os.Stdout
	}
	for _, command := range commands {
		var start time.Time
		if timed || verbose {
			fmt.Fprintf(out, "→ %s\n", command)
			start = time.Now()
		}
		cmd := exec.Command("sh", "-c", command)
		cmd.Dir = workdir
		if verbose {
			cmd.Stdout = out
			cmd.Stderr = out
			if err := cmd.Run(); err != nil {
				if timed {
					fmt.Fprintf(out, "✗ failed in %s\n", time.Since(start).Round(time.Millisecond))
				}
				return err
			}
		} else {
			// Non-verbose: still execute but suppress output
			if err := cmd.Run(); err != nil {
				if timed {
					fmt.Fprintf(out, "✗ failed in %s\n", time.Since(start).Round(time.Millisecond))
				}
				return err
			}
		}
		if timed || verbose {
			fmt.Fprintf(out, "✓ done in %s\n", time.Since(start).Round(time.Millisecond))
		}
	}
	return nil
}
