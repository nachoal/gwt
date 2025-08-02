package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nachoal/gwt/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all worktrees for the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Show interactive list with Bubble Tea
		p := tea.NewProgram(ui.NewListModel())
		_, err := p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
