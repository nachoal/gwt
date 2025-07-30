package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func GetWorktreePath(root, projectName, branchName string) string {
	return filepath.Join(root, projectName, branchName)
}

func Create(branchName, fromBranch, targetPath string) error {
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

func List() ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
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
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)
	
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
			// Copy directory recursively
			cmd := exec.Command("cp", "-r", src, dest)
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