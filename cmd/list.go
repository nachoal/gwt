package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nachoal/gwt/internal/ui"
	"github.com/nachoal/gwt/internal/worktree"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all worktrees for the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Root-mode listing: enumerate configured root for all projects' worktrees
		rootMode, _ := cmd.Flags().GetBool("root")
		overridePath, _ := cmd.Flags().GetString("path")
		if rootMode {
			return listFromRoot(overridePath)
		}

		// Default: Show interactive list (current repo only).
		// Render UI to stderr so stdout can be used for the selected path (shell integration).
		p := tea.NewProgram(ui.NewListModel(), tea.WithInputTTY(), tea.WithOutput(os.Stderr))
		m, err := p.Run()
		if err != nil {
			return err
		}

		// If the user selected a worktree (Enter), print its path to stdout.
		type selectedPathModel interface{ SelectedPath() string }
		if sp, ok := m.(selectedPathModel); ok {
			if path := sp.SelectedPath(); path != "" {
				fmt.Println(path)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().Bool("root", false, "List all gwt worktrees under the configured root")
	listCmd.Flags().String("path", "", "Override root path to scan (defaults to settings.root)")
}

// listFromRoot handles the POC root enumeration flow and prints a concise table.
func listFromRoot(override string) error {
	results, rootPath, err := worktree.ListFromRoot(override)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println(infoStyle.Render("No worktrees found under:"), fileStyle.Render(rootPath))
		return nil
	}

	fmt.Println(titleStyle.Render("All Worktrees") + " " + infoStyle.Render("(root: ") + fileStyle.Render(rootPath) + infoStyle.Render(")"))

	// Compute simple column widths
	maxProj, maxBranch := 7, 6 // headers length
	for _, r := range results {
		if l := len(r.Project); l > maxProj {
			maxProj = l
		}
		if l := len(r.Branch); l > maxBranch {
			maxBranch = l
		}
	}

	hdr := fmt.Sprintf("%-*s  %-*s  %-7s  %s", maxProj, "Project", maxBranch, "Branch", "HEAD", "Path")
	fmt.Println(infoStyle.Render(hdr))
	for _, r := range results {
		line := fmt.Sprintf("%-*s  %-*s  %-7s  %s", maxProj, r.Project, maxBranch, r.Branch, r.Head, r.Path)
		fmt.Println(line)
	}
	return nil
}
