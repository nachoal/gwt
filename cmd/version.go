package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	buildVersion = "dev"
	buildCommit  = ""
	buildDate    = ""
)

type versionInfo struct {
	Version    string `json:"version"`
	Commit     string `json:"commit,omitempty"`
	Date       string `json:"date,omitempty"`
	GoVersion  string `json:"go_version"`
	Module     string `json:"module,omitempty"`
	Executable string `json:"executable,omitempty"`
}

func shortVersion() string {
	return fmt.Sprintf("gwt %s", getVersionInfo().Version)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show gwt version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOut, _ := cmd.Flags().GetBool("json")
		info := getVersionInfo()

		if jsonOut {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(info)
		}

		fmt.Printf("gwt %s\n", info.Version)
		if info.Commit != "" {
			fmt.Printf("commit: %s\n", info.Commit)
		}
		if info.Date != "" {
			fmt.Printf("date: %s\n", info.Date)
		}
		if info.Module != "" {
			fmt.Printf("module: %s\n", info.Module)
		}
		fmt.Printf("go: %s\n", info.GoVersion)
		if info.Executable != "" {
			fmt.Printf("executable: %s\n", info.Executable)
		}
		return nil
	},
}

func getVersionInfo() versionInfo {
	info := versionInfo{
		Version:   buildVersion,
		Commit:    buildCommit,
		Date:      buildDate,
		GoVersion: runtime.Version(),
	}

	if exe, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			info.Executable = resolved
		} else {
			info.Executable = exe
		}
	}

	if bi, ok := debug.ReadBuildInfo(); ok {
		if info.Module == "" && bi.Main.Path != "" {
			info.Module = bi.Main.Path
		}
		if info.Version == "dev" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			info.Version = bi.Main.Version
		}
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				if info.Commit == "" {
					info.Commit = s.Value
				}
			case "vcs.time":
				if info.Date == "" {
					info.Date = s.Value
				}
			}
		}
	}

	return info
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().Bool("json", false, "Machine-readable JSON output")
}
