package cmd

import (
	"fmt"
	"github.com/nachoal/gwt/internal/worktree"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:     "switch <branch-name>",
	Aliases: []string{"sw"},
	Short:   "Switch to a worktree (requires shell integration)",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branchName := args[0]

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

		// Output the path for shell function to cd to
		fmt.Println(targetPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
