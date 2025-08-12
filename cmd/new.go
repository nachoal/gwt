package cmd

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/nachoal/gwt/internal/config"
    "github.com/nachoal/gwt/internal/ui"
    "github.com/nachoal/gwt/internal/worktree"
    "github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
    Use:   "new <branch-name>",
    Short: "Create a new worktree",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        branchName := args[0]
        fromBranch, _ := cmd.Flags().GetString("from")
        verbose, _ := cmd.Flags().GetBool("verbose")
        timed, _ := cmd.Flags().GetBool("timed")

        // If no from branch specified, auto-detect the default branch
        if cmd.Flags().Changed("from") == false {
            defaultBranch, err := worktree.GetDefaultBranch()
            if err == nil {
                fromBranch = defaultBranch
            }
        }

        // If verbose or timed flags are set, run a non-TUI flow with console output
        if verbose || timed {
            return createWorktreeNonTUI(branchName, fromBranch, verbose, timed)
        }

        // Otherwise, run the TUI flow
        p := tea.NewProgram(ui.NewCreateModel(branchName, fromBranch))
        _, err := p.Run()
        return err
    },
}

func init() {
    rootCmd.AddCommand(newCmd)
    newCmd.Flags().StringP("from", "f", "", "Base branch to create worktree from (auto-detected if not specified)")
    newCmd.Flags().BoolP("verbose", "v", false, "Verbose output for setup commands (stream stdout/stderr)")
    newCmd.Flags().BoolP("timed", "t", false, "Print each setup command and how long it took")
}

func createWorktreeNonTUI(branchName, fromBranch string, verbose, timed bool) error {
    fmt.Println(titleStyle.Render("Creating worktree (non-TUI)"))

    // Step 1: Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        return err
    }

    // Step 2: Determine project and target path
    projectName, err := worktree.GetProjectName()
    if err != nil {
        return err
    }
    targetPath := worktree.GetWorktreePath(cfg.Settings.Root, projectName, branchName)
    fmt.Printf("→ Project: %s\n", projectName)
    fmt.Printf("→ From: %s\n", fromBranch)
    fmt.Printf("→ Path: %s\n", targetPath)

    // Step 3: Create worktree
    if err := worktree.Create(branchName, fromBranch, targetPath); err != nil {
        return err
    }
    fmt.Println(successStyle.Render("✓ Worktree created"))

    // Step 4: Copy files
    mainPath, err := os.Getwd()
    if err != nil {
        return err
    }
    if err := worktree.CopyFiles(mainPath, targetPath, cfg.Copy); err != nil {
        return err
    }
    fmt.Println(successStyle.Render("✓ Files copied"))

    // Step 5: Run setup commands
    if len(cfg.Setup) > 0 {
        fmt.Println(infoStyle.Render("Running setup commands:"))
        if err := worktree.RunSetupCommandsOpts(targetPath, cfg.Setup, verbose, timed, os.Stdout); err != nil {
            return err
        }
    }
    fmt.Println(successStyle.Render("✓ Done"))
    fmt.Println(fileStyle.Render(targetPath))
    return nil
}
