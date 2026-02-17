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
    ls|list)
      shift
      # Root listing prints a table; don't treat its output as a path.
      for arg in "$@"; do
        if [ "$arg" = "--root" ]; then
          command gwt list "$@"
          return $?
        fi
      done

      local wt_path
      wt_path=$(GWT_FORCE_TUI=1 command gwt list "$@")
      if [ $? -eq 0 ] && [ -n "$wt_path" ]; then
        cd "$wt_path"
        # Emit OSC 7 to inform WezTerm of directory change
        printf "\033]7;file://%s%s\033\\" "${HOST:-$HOSTNAME}" "$PWD"
      fi
      ;;

    sw|switch)
      shift
      local wt_path
      wt_path=$(command gwt switch "$@")
      if [ $? -eq 0 ] && [ -n "$wt_path" ]; then
        cd "$wt_path"
        # Emit OSC 7 to inform WezTerm of directory change
        printf "\033]7;file://%s%s\033\\" "${HOST:-$HOSTNAME}" "$PWD"
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
        # Emit OSC 7 to inform WezTerm of directory change
        printf "\033]7;file://%s%s\033\\" "${HOST:-$HOSTNAME}" "$PWD"
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
      local wt_path
      wt_path=$(command gwt done --print-path "$@")
      local _gwt_ec=$?
      if [ $_gwt_ec -ne 0 ]; then
        return $_gwt_ec
      fi
      if [ -n "$wt_path" ]; then
        cd "$wt_path" || return $?
        # Emit OSC 7 to inform WezTerm of directory change
        printf "\033]7;file://%s%s\033\\" "${HOST:-$HOSTNAME}" "$PWD"
      fi
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

func shellScriptPathForRC(rcPath string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	scriptName := "shell.sh"
	base := strings.ToLower(filepath.Base(rcPath))
	switch {
	case strings.Contains(base, "zsh"):
		scriptName = "shell.zsh"
	case strings.Contains(base, "bash"):
		scriptName = "shell.bash"
	}
	return filepath.Join(home, ".config", "gwt", scriptName), nil
}

func sourceLineForShellScript(scriptPath string) string {
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(scriptPath, home+string(os.PathSeparator)) {
		return "source ~" + scriptPath[len(home):]
	}
	return "source " + scriptPath
}

func removeManagedBlock(content string) string {
	if i := strings.Index(content, startMarker); i >= 0 {
		if j := strings.Index(content[i+len(startMarker):], endMarker); j >= 0 {
			return content[:i] + content[i+len(startMarker)+j+len(endMarker):]
		}
	}
	return content
}

func installShell(rcPath string) error {
	scriptPath, err := shellScriptPathForRC(rcPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(scriptPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(scriptPath, []byte(shellFunction+"\n"), 0o644); err != nil {
		return err
	}

	data, _ := os.ReadFile(rcPath)
	content := removeManagedBlock(string(data))
	// ensure trailing newline
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	block := startMarker + sourceLineForShellScript(scriptPath) + "\n" + endMarker
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
	if err := os.WriteFile(rcPath, []byte(content), 0o644); err != nil {
		return err
	}

	if scriptPath, err := shellScriptPathForRC(rcPath); err == nil {
		_ = os.Remove(scriptPath)
	}
	return nil
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
	Long:  "Output shell integration to enable 'gwt switch' to cd, plus helpers for 'new -c' and 'done'. Use --install to write a small source line into your shell rc (zsh/bash).",
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
			scriptPath, _ := shellScriptPathForRC(rc)
			fmt.Printf("Installed gwt shell integration into %s\n", rc)
			if scriptPath != "" {
				fmt.Printf("Wrote shell helper to %s\n", scriptPath)
			}
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
		fmt.Printf("%s", shellFunction)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.Flags().Bool("install", false, "Install a managed source line into your rc file")
	shellCmd.Flags().Bool("remove", false, "Remove managed shell integration from your rc file")
	shellCmd.Flags().String("rc", "", "Path to rc file (defaults based on $SHELL)")
}
