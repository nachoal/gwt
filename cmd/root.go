package cmd

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

var rootCmd = &cobra.Command{
	Use:   "gwt",
	Short: "A beautiful git worktree manager",
	Long: titleStyle.Render("gwt") + " - Git Worktree Manager\n\n" +
		infoStyle.Render("Simplify your git worktree workflow with automatic setup and management") + "\n\n" +
		infoStyle.Render("Enable shell integration via 'gwt shell' to get helpers:") + "\n" +
		infoStyle.Render("  • gwt new <branch> -c [issue <url> | <prompt>]  (cd + run 'claude')") + "\n" +
		infoStyle.Render("  • gwt done <branch> [base]                       (pull base, remove worktree)"),
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
