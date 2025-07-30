package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nachoal/gwt/internal/ui"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <branch-name>",
	Short: "Create a new worktree",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branchName := args[0]
		fromBranch, _ := cmd.Flags().GetString("from")
		
		// Create a beautiful TUI for the worktree creation process
		p := tea.NewProgram(ui.NewCreateModel(branchName, fromBranch))
		_, err := p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringP("from", "f", "main", "Base branch to create worktree from")
}