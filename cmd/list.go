package cmd

import (
	"encoding/json"
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
		noTUI, _ := cmd.Flags().GetBool("no-tui")
		plain, _ := cmd.Flags().GetBool("plain")
		jsonOut, _ := cmd.Flags().GetBool("json")

		format, err := resolveOutputFormat(plain, jsonOut)
		if err != nil {
			return err
		}

		if rootMode {
			return listFromRoot(overridePath, format)
		}

		useTUI := !noTUI && format == outputFormatPretty && hasInteractiveTTY()
		if useTUI {
			// Default interactive mode for humans.
			// Render UI to stderr so stdout can carry the selected path (shell integration).
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
		}

		return listCurrentRepo(format)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().Bool("root", false, "List all gwt worktrees under the configured root")
	listCmd.Flags().String("path", "", "Override root path to scan (defaults to settings.root)")
	listCmd.Flags().Bool("no-tui", false, "Run without the interactive list UI (auto-enabled when no interactive TTY is available)")
	listCmd.Flags().Bool("plain", false, "Plain text output without styling")
	listCmd.Flags().Bool("json", false, "Machine-readable JSON output")
}

func listCurrentRepo(format outputFormat) error {
	results, err := worktree.List()
	if err != nil {
		return err
	}

	if len(results) == 0 {
		switch format {
		case outputFormatPretty:
			fmt.Println(infoStyle.Render("No worktrees found for this repository"))
		case outputFormatPlain:
			fmt.Println("count=0")
		case outputFormatJSON:
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(results)
		}
		return nil
	}

	if format == outputFormatJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	maxBranch := 6 // "Branch"
	for _, r := range results {
		if l := len(r.Branch); l > maxBranch {
			maxBranch = l
		}
	}

	header := fmt.Sprintf("%-*s  %-7s  %s", maxBranch, "Branch", "HEAD", "Path")
	if format == outputFormatPretty {
		fmt.Println(titleStyle.Render("Worktrees"))
		fmt.Println(infoStyle.Render(header))
	}
	if format == outputFormatPlain {
		fmt.Println("branch\thead\tpath")
	}

	for _, r := range results {
		if format == outputFormatPretty {
			line := fmt.Sprintf("%-*s  %-7s  %s", maxBranch, r.Branch, r.Head, r.Path)
			fmt.Println(line)
		}
		if format == outputFormatPlain {
			fmt.Printf("%s\t%s\t%s\n", r.Branch, r.Head, r.Path)
		}
	}

	return nil
}

type rootListResult struct {
	Root  string              `json:"root"`
	Items []worktree.RootItem `json:"items"`
}

// listFromRoot handles root enumeration and prints it in the requested format.
func listFromRoot(override string, format outputFormat) error {
	results, rootPath, err := worktree.ListFromRoot(override)
	if err != nil {
		return err
	}

	if format == outputFormatJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rootListResult{Root: rootPath, Items: results})
	}

	if len(results) == 0 {
		if format == outputFormatPretty {
			fmt.Println(infoStyle.Render("No worktrees found under:"), fileStyle.Render(rootPath))
		}
		if format == outputFormatPlain {
			fmt.Printf("root=%s\n", rootPath)
			fmt.Println("count=0")
		}
		return nil
	}

	if format == outputFormatPretty {
		fmt.Println(titleStyle.Render("All Worktrees") + " " + infoStyle.Render("(root: ") + fileStyle.Render(rootPath) + infoStyle.Render(")"))
	}
	if format == outputFormatPlain {
		fmt.Printf("root=%s\n", rootPath)
		fmt.Println("project\tbranch\thead\tpath")
	}

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

	if format == outputFormatPretty {
		hdr := fmt.Sprintf("%-*s  %-*s  %-7s  %s", maxProj, "Project", maxBranch, "Branch", "HEAD", "Path")
		fmt.Println(infoStyle.Render(hdr))
	}

	for _, r := range results {
		if format == outputFormatPretty {
			line := fmt.Sprintf("%-*s  %-*s  %-7s  %s", maxProj, r.Project, maxBranch, r.Branch, r.Head, r.Path)
			fmt.Println(line)
		}
		if format == outputFormatPlain {
			fmt.Printf("%s\t%s\t%s\t%s\n", r.Project, r.Branch, r.Head, r.Path)
		}
	}
	return nil
}
