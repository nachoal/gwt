package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const shellFunction = `
# gwt shell integration
# Ensure we override any prior alias/function
unalias gwt 2>/dev/null || true
unset -f gwt 2>/dev/null || true
function gwt {
  local sub="$1"
  if [ -z "$sub" ]; then
    command gwt
    return $?
  fi

  case "$sub" in
    sw|switch)
      shift
      local wt_path
      wt_path=$(command gwt switch "$@")
      if [ $? -eq 0 ] && [ -n "$wt_path" ]; then
        cd "$wt_path"
      fi
      ;;

    new)
      shift
      # Parse args; support: -f/--from, -v, -t and shell-only -c/--claude
      local branch=""
      local claude=0
      local claude_prompt=""
      local pass=()

      while [ $# -gt 0 ]; do
        case "$1" in
          -c|--claude)
            claude=1
            shift
            if [ $# -gt 0 ]; then
              if [ "$1" = "issue" ]; then
                shift
                if [ $# -gt 0 ]; then
                  claude_prompt="/issue-analysis $1"
                  shift
                else
                  echo "gwt: -c issue requires a <link>" >&2
                  return 1
                fi
              else
                claude_prompt="$1"
                shift
              fi
            fi
            ;;
          -f|--from)
            pass+=("$1")
            shift
            if [ $# -gt 0 ]; then
              pass+=("$1")
              shift
            else
              echo "gwt: -f/--from requires a value" >&2
              return 1
            fi
            ;;
          -*)
            pass+=("$1")
            shift
            ;;
          *)
            if [ -z "$branch" ]; then
              branch="$1"
            fi
            pass+=("$1")
            shift
            ;;
        esac
      done

      if [ -z "$branch" ]; then
        echo "gwt: new requires <branch-name>" >&2
        return 1
      fi

      command gwt new "${pass[@]}"
      local _gwt_ec=$?
      if [ $_gwt_ec -ne 0 ]; then
        return $_gwt_ec
      fi

      local wt_path
      wt_path=$(command gwt switch "$branch")
      if [ $? -eq 0 ] && [ -n "$wt_path" ]; then
        cd "$wt_path" || return $?
      else
        return 1
      fi

      if [ $claude -eq 1 ]; then
        if [ -n "$claude_prompt" ]; then
          claude "$claude_prompt"
        else
          claude
        fi
      fi
      ;;

    done)
      shift
      local branch="$1"
      local base_branch="$2"
      # If no branch provided, try to infer from current worktree
      if [ -z "$branch" ]; then
        branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)
        if [ -z "$branch" ] || [ "$branch" = "HEAD" ]; then
          echo "Usage: gwt done <branch> [base] (or run inside a worktree)" >&2
          return 1
        fi
      fi
      # If no base provided, try to detect the repo's default
      if [ -z "$base_branch" ]; then
        local _defref
        _defref=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null)
        if [ -n "$_defref" ]; then
          base_branch="${_defref##*/}"
        fi
        if [ -z "$base_branch" ]; then
          for b in main master develop; do
            local _p
            _p=$(command gwt switch "$b" 2>/dev/null)
            if [ $? -eq 0 ] && [ -n "$_p" ]; then
              base_branch="$b"
              break
            fi
          done
        fi
      fi
      if [ -z "$base_branch" ]; then
        echo "gwt: could not determine base branch; specify explicitly" >&2
        return 1
      fi
      local wt_path
      wt_path=$(command gwt switch "$base_branch")
      if [ $? -ne 0 ] || [ -z "$wt_path" ]; then
        echo "gwt: base worktree '$base_branch' not found" >&2
        return 1
      fi
      cd "$wt_path" || return $?
      git pull --ff-only || return $?
      command gwt remove "$branch"
      ;;

    *)
      shift
      command gwt "$sub" "$@"
      ;;
  esac
}
`

const startMarker = "# >>> gwt shell integration >>>\n"
const endMarker = "# <<< gwt shell integration <<<\n"

func installShell(rcPath string) error {
	data, _ := os.ReadFile(rcPath)
	content := string(data)
	// remove existing block
	if i := strings.Index(content, startMarker); i >= 0 {
		if j := strings.Index(content[i+len(startMarker):], endMarker); j >= 0 {
			content = content[:i] + content[i+len(startMarker)+j+len(endMarker):]
		}
	}
	// ensure trailing newline
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	block := startMarker + shellFunction + "\n" + endMarker
	content += block

	if err := os.MkdirAll(filepath.Dir(rcPath), 0o755); err != nil {
		return err
	}
	// backup
	if _, err := os.Stat(rcPath); err == nil {
		_ = os.WriteFile(rcPath+".bak.gwt", data, 0o644)
	}
	return os.WriteFile(rcPath, []byte(content), 0o644)
}

func removeShell(rcPath string) error {
	data, err := os.ReadFile(rcPath)
	if err != nil {
		return err
	}
	content := string(data)
	i := strings.Index(content, startMarker)
	if i < 0 {
		return errors.New("no gwt shell block found")
	}
	j := strings.Index(content[i+len(startMarker):], endMarker)
	if j < 0 {
		return errors.New("unterminated gwt shell block")
	}
	content = content[:i] + content[i+len(startMarker)+j+len(endMarker):]
	return os.WriteFile(rcPath, []byte(content), 0o644)
}

func detectDefaultRC() string {
	shell := os.Getenv("SHELL")
	home, _ := os.UserHomeDir()
	if strings.Contains(shell, "zsh") {
		return filepath.Join(home, ".zshrc")
	}
	if strings.Contains(shell, "bash") {
		// Prefer .bashrc; macOS often uses .bash_profile
		rc := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(rc); err == nil {
			return rc
		}
		return filepath.Join(home, ".bash_profile")
	}
	return filepath.Join(home, ".profile")
}

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Output or install shell integration",
	Long:  "Output shell integration to enable 'gwt switch' to cd, plus helpers for 'new -c' and 'done'. Use --install to write into your shell rc (zsh/bash).",
	RunE: func(cmd *cobra.Command, args []string) error {
		install, _ := cmd.Flags().GetBool("install")
		remove, _ := cmd.Flags().GetBool("remove")
		rc, _ := cmd.Flags().GetString("rc")
		if install && remove {
			return fmt.Errorf("--install and --remove are mutually exclusive")
		}
		if rc == "" {
			rc = detectDefaultRC()
		}
		if install {
			if err := installShell(rc); err != nil {
				return err
			}
			fmt.Printf("Installed gwt shell integration into %s\n", rc)
			fmt.Println("Open a new shell or 'source' your rc to activate.")
			return nil
		}
		if remove {
			if err := removeShell(rc); err != nil {
				return err
			}
			fmt.Printf("Removed gwt shell integration from %s\n", rc)
			return nil
		}
		fmt.Print(shellFunction)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.Flags().Bool("install", false, "Install the shell function into your rc file")
	shellCmd.Flags().Bool("remove", false, "Remove the shell function from your rc file")
	shellCmd.Flags().String("rc", "", "Path to rc file (defaults based on $SHELL)")
}
