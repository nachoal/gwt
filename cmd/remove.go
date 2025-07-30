package cmd

import (
	"fmt"
	"github.com/ia/gwt/internal/worktree"
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
		
		// Remove the worktree
		if err := worktree.Remove(targetPath, force); err != nil {
			return err
		}
		
		fmt.Println(successStyle.Render("✓") + " Removed worktree: " + fileStyle.Render(branchName))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolP("force", "f", false, "Force removal even if there are uncommitted changes")
}