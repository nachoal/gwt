package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nachoal/gwt/internal/config"
	"github.com/nachoal/gwt/internal/worktree"
	"github.com/spf13/cobra"
)

var (
	checkMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓")
	xMark     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗")
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove worktrees for merged branches",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		if !cfg.Settings.AutoCleanMerged {
			return fmt.Errorf("auto clean is disabled in config")
		}

		// Get list of merged branches
		mergedCmd := exec.Command("git", "branch", "--merged", "main")
		output, err := mergedCmd.Output()
		if err != nil {
			return err
		}

		mergedBranches := make(map[string]bool)
		for _, line := range strings.Split(string(output), "\n") {
			branch := strings.TrimSpace(line)
			branch = strings.TrimPrefix(branch, "* ")
			if branch != "" && branch != "main" && branch != "master" {
				mergedBranches[branch] = true
			}
		}

		// Get worktrees
		worktrees, err := worktree.List()
		if err != nil {
			return err
		}

		removedCount := 0
		for _, wt := range worktrees {
			if mergedBranches[wt.Branch] {
				fmt.Printf("Removing merged worktree: %s\n", fileStyle.Render(wt.Branch))
				if err := worktree.Remove(wt.Path, false); err != nil {
					fmt.Printf("  %s Failed: %v\n", xMark, err)
				} else {
					fmt.Printf("  %s Done\n", checkMark)
					removedCount++
				}
			}
		}

		if removedCount == 0 {
			fmt.Println(infoStyle.Render("No merged worktrees to clean"))
		} else {
			fmt.Printf("\n%s Cleaned %d worktree(s)\n", 
				successStyle.Render("✓"), removedCount)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}