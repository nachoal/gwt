package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

const shellFunction = `
# gwt shell integration
gwt() {
    if [ "$1" = "sw" ] || [ "$1" = "switch" ]; then
        local path=$(command gwt "$@")
        if [ $? -eq 0 ] && [ -n "$path" ]; then
            cd "$path"
        fi
    else
        command gwt "$@"
    fi
}
`

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Output shell integration script",
	Long:  "Output shell integration script to enable 'gwt switch' to change directories",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(shellFunction)
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
}
