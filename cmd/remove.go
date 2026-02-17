package cmd

import (
	"fmt"
	"os"

	"github.com/nachoal/gwt/internal/worktree"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <branch-name>",
	Aliases: []string{"rm"},
	Short:   "Remove a worktree",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branchName := args[0]
		force, _ := cmd.Flags().GetBool("force")
		if err := removeWorktreeByBranch(branchName, force); err != nil {
			return err
		}

		fmt.Println(successStyle.Render("✓") + " Removed worktree and branch: " + fileStyle.Render(branchName))
		return nil
	},
}

func removeWorktreeByBranch(branchName string, force bool) error {
	// Find the worktree path
	worktrees, err := worktree.List()
	if err != nil {
		return err
	}

	var targetPath string
	for _, wt := range worktrees {
		if wt.Branch == branchName {
			targetPath = wt.Path
			break
		}
	}

	if targetPath == "" {
		return fmt.Errorf("worktree for branch '%s' not found", branchName)
	}

	// Determine common git dir before removal (branch is checked out here)
	commonGitDir, _ := worktree.GetCommonGitDir(targetPath)

	// Remove the worktree first to unlock the branch
	if err := worktree.Remove(targetPath, force); err != nil {
		return err
	}

	// Also delete the branch (safe delete unless --force)
	if err := worktree.DeleteBranchWithGitDir(commonGitDir, branchName, force); err != nil {
		// Warn but don't fail the command if branch deletion fails (e.g., unmerged)
		fmt.Fprintln(os.Stderr, infoStyle.Render("Note: could not delete branch ")+fileStyle.Render(branchName))
	}

	return nil
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolP("force", "f", false, "Force removal even if there are uncommitted changes")
}
