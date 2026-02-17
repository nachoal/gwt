package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nachoal/gwt/internal/worktree"
	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done [branch] [base]",
	Short: "Update base branch and remove a completed worktree",
	Long: "Finalize work on a branch by updating the base branch and removing the branch's worktree.\n\n" +
		"If branch is omitted, gwt infers it from the current worktree.\n" +
		"If base is omitted, gwt uses the repository default branch.",
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		branchName, err := resolveDoneBranch(args)
		if err != nil {
			return err
		}

		baseBranch, err := resolveDoneBase(args)
		if err != nil {
			return err
		}
		if branchName == baseBranch {
			return fmt.Errorf("refusing to remove base branch '%s'", baseBranch)
		}

		basePath, usedBaseWorktree, err := updateBaseBranch(baseBranch)
		if err != nil {
			return err
		}

		if usedBaseWorktree {
			fmt.Fprintln(os.Stderr, infoStyle.Render("✓ Updated base branch ")+fileStyle.Render(baseBranch))
		} else {
			fmt.Fprintln(os.Stderr, infoStyle.Render("✓ Updated local base branch ref ")+fileStyle.Render(baseBranch))
		}

		if err := removeWorktreeByBranch(branchName, false); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, successStyle.Render("✓")+" Done: removed "+fileStyle.Render(branchName))

		printPath, _ := cmd.Flags().GetBool("print-path")
		if printPath {
			fmt.Println(basePath)
		}
		return nil
	},
}

func resolveDoneBranch(args []string) (string, error) {
	if len(args) >= 1 && strings.TrimSpace(args[0]) != "" {
		return args[0], nil
	}
	branch, err := currentBranch()
	if err != nil || branch == "" || branch == "HEAD" {
		return "", fmt.Errorf("usage: gwt done <branch> [base] (or run inside a worktree)")
	}
	return branch, nil
}

func resolveDoneBase(args []string) (string, error) {
	if len(args) >= 2 && strings.TrimSpace(args[1]) != "" {
		return args[1], nil
	}
	base, err := worktree.GetDefaultBranch()
	if err != nil {
		return "", err
	}
	if base == "" {
		return "", fmt.Errorf("could not determine base branch; specify one explicitly")
	}
	return base, nil
}

func currentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func updateBaseBranch(baseBranch string) (string, bool, error) {
	basePath, err := worktreePathForBranch(baseBranch)
	if err != nil {
		return "", false, err
	}
	if basePath != "" {
		fmt.Fprintln(os.Stderr, infoStyle.Render("Updating base branch in worktree: "), fileStyle.Render(basePath))
		if err := runGitInDir(basePath, "pull", "--ff-only"); err != nil {
			return "", false, err
		}
		return basePath, true, nil
	}

	mainWT, err := worktree.FindMainWorktree()
	if err != nil {
		return "", false, err
	}
	fmt.Fprintln(os.Stderr, infoStyle.Render("No base worktree checked out; fast-forwarding local ref in "), fileStyle.Render(mainWT))
	if err := runGitInDir(mainWT, "fetch", "origin", fmt.Sprintf("%s:%s", baseBranch, baseBranch)); err != nil {
		return "", false, err
	}
	return mainWT, false, nil
}

func worktreePathForBranch(branch string) (string, error) {
	worktrees, err := worktree.List()
	if err != nil {
		return "", err
	}
	for _, wt := range worktrees {
		if wt.Branch == branch {
			return wt.Path, nil
		}
	}
	return "", nil
}

func runGitInDir(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s failed: %w", strings.Join(args, " "), err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(doneCmd)
	doneCmd.Flags().Bool("print-path", false, "Print the selected base worktree path on success")
	_ = doneCmd.Flags().MarkHidden("print-path")
}
