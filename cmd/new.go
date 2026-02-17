package cmd

import (
	"encoding/json"
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
	Long: "Create a new worktree for the current project.\n\n" +
		"Tip: With shell integration enabled (eval \"$(gwt shell)\"), you can use:\n" +
		"  • gwt new <branch> -c                # cd into the new worktree and run your 'claude' alias\n" +
		"  • gwt new <branch> -c \"prompt\"     # run 'claude \"prompt\"'\n" +
		"  • gwt new <branch> -c issue <url>   # run 'claude \"/issue-analysis <url>\"'\n\n" +
		"Note: -c is provided by the shell wrapper, not by the gwt binary.",
	Example: "  gwt new feature/foo\n" +
		"  gwt new feature/foo -f develop\n" +
		"  # With shell integration:\n" +
		"  gwt new fix/bug -c issue https://example/issue/123\n",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branchName := args[0]
		fromBranch, _ := cmd.Flags().GetString("from")
		verbose, _ := cmd.Flags().GetBool("verbose")
		timed, _ := cmd.Flags().GetBool("timed")
		noTUI, _ := cmd.Flags().GetBool("no-tui")
		plain, _ := cmd.Flags().GetBool("plain")
		jsonOut, _ := cmd.Flags().GetBool("json")

		format, err := resolveOutputFormat(plain, jsonOut)
		if err != nil {
			return err
		}
		if format == outputFormatJSON && (verbose || timed) {
			return fmt.Errorf("--json cannot be combined with --verbose or --timed")
		}

		// If no from branch specified, auto-detect the default branch
		if cmd.Flags().Changed("from") == false {
			defaultBranch, err := worktree.GetDefaultBranch()
			if err == nil {
				fromBranch = defaultBranch
			}
		}

		useTUI := !noTUI &&
			format == outputFormatPretty &&
			!verbose &&
			!timed &&
			hasInteractiveTTY()

		// If non-interactive (or explicitly disabled), run the non-TUI flow.
		if !useTUI {
			return createWorktreeNonTUI(branchName, fromBranch, verbose, timed, format)
		}

		// Otherwise, run the TUI flow (render to stderr to keep stdout script-friendly).
		p := tea.NewProgram(ui.NewCreateModel(branchName, fromBranch), tea.WithOutput(os.Stderr))
		_, err = p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringP("from", "f", "", "Base branch to create worktree from (auto-detected if not specified)")
	newCmd.Flags().BoolP("verbose", "v", false, "Verbose output for setup commands (stream stdout/stderr)")
	newCmd.Flags().BoolP("timed", "t", false, "Print each setup command and how long it took")
	newCmd.Flags().Bool("no-tui", false, "Run without the TUI (auto-enabled when no interactive TTY is available)")
	newCmd.Flags().Bool("plain", false, "Plain text output without styling")
	newCmd.Flags().Bool("json", false, "Machine-readable JSON output")
}

type createResult struct {
	Status string `json:"status"`
	Branch string `json:"branch"`
	From   string `json:"from"`
	Path   string `json:"path"`
}

func createWorktreeNonTUI(branchName, fromBranch string, verbose, timed bool, format outputFormat) error {
	if format == outputFormatPretty {
		fmt.Println(titleStyle.Render("Creating worktree (non-TUI)"))
	}
	if format == outputFormatPlain {
		fmt.Println("Creating worktree")
	}

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
	if format == outputFormatPretty {
		fmt.Printf("→ Project: %s\n", projectName)
		fmt.Printf("→ From: %s\n", fromBranch)
		fmt.Printf("→ Path: %s\n", targetPath)
	}
	if format == outputFormatPlain {
		fmt.Printf("project=%s\n", projectName)
		fmt.Printf("from=%s\n", fromBranch)
		fmt.Printf("path=%s\n", targetPath)
	}

	// Step 3: Create worktree
	if err := worktree.Create(branchName, fromBranch, targetPath); err != nil {
		return err
	}
	if format == outputFormatPretty {
		fmt.Println(successStyle.Render("✓ Worktree created"))
	}
	if format == outputFormatPlain {
		fmt.Println("worktree_created=true")
	}

	// Step 4: Copy files
	mainPath, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := worktree.CopyFiles(mainPath, targetPath, cfg.Copy); err != nil {
		return err
	}
	if format == outputFormatPretty {
		fmt.Println(successStyle.Render("✓ Files copied"))
	}
	if format == outputFormatPlain {
		fmt.Println("files_copied=true")
	}

	// Step 5: Run setup commands
	if len(cfg.Setup) > 0 {
		out := os.Stdout
		if format == outputFormatJSON {
			out = nil
		} else if format == outputFormatPretty {
			fmt.Println(infoStyle.Render("Running setup commands:"))
		} else if format == outputFormatPlain {
			fmt.Println("running_setup=true")
		}
		if err := worktree.RunSetupCommandsOpts(targetPath, cfg.Setup, verbose, timed, out); err != nil {
			return err
		}
	}

	switch format {
	case outputFormatPretty:
		fmt.Println(successStyle.Render("✓ Done"))
		fmt.Println(fileStyle.Render(targetPath))
	case outputFormatPlain:
		fmt.Println("status=ok")
		fmt.Printf("path=%s\n", targetPath)
	case outputFormatJSON:
		result := createResult{
			Status: "ok",
			Branch: branchName,
			From:   fromBranch,
			Path:   targetPath,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			return err
		}
	}

	return nil
}
