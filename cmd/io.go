package cmd

import (
	"fmt"
	"os"
	"strings"
)

type outputFormat string

const (
	outputFormatPretty outputFormat = "pretty"
	outputFormatPlain  outputFormat = "plain"
	outputFormatJSON   outputFormat = "json"
)

func resolveOutputFormat(plain bool, jsonOut bool) (outputFormat, error) {
	if plain && jsonOut {
		return "", fmt.Errorf("--plain and --json are mutually exclusive")
	}
	if jsonOut {
		return outputFormatJSON, nil
	}
	if plain {
		return outputFormatPlain, nil
	}
	return outputFormatPretty, nil
}

func hasInteractiveTTY() bool {
	term := strings.TrimSpace(os.Getenv("TERM"))
	if term == "" || term == "dumb" {
		return false
	}
	if !isTTY(os.Stdin) || !isTTY(os.Stderr) {
		return false
	}
	// Allow captured stdout only when explicitly requested by shell integration.
	if !isTTY(os.Stdout) && os.Getenv("GWT_FORCE_TUI") != "1" {
		return false
	}
	return true
}

func isTTY(f *os.File) bool {
	if f == nil {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
