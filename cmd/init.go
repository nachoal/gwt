package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/nachoal/gwt/internal/config"
	"github.com/spf13/cobra"
)

var (
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)
	
	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a .worktree.yaml config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(".worktree.yaml"); err == nil {
			return fmt.Errorf("config file already exists")
		}

		cfg := config.DefaultConfig()
		if err := config.SaveConfig(cfg); err != nil {
			return err
		}

		fmt.Println(successStyle.Render("✓") + " Created " + fileStyle.Render(".worktree.yaml"))
		fmt.Println(infoStyle.Render("Edit this file to customize your worktree setup"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}