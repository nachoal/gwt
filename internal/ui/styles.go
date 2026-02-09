package ui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Use stderr for style rendering so TUI styling still works when stdout is
	// captured (e.g. shell integration using command substitution).
	uiRenderer = lipgloss.NewRenderer(os.Stderr)

	titleStyle = uiRenderer.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	infoStyle = uiRenderer.NewStyle().
			Foreground(lipgloss.Color("241"))

	errorStyle = uiRenderer.NewStyle().
			Foreground(lipgloss.Color("196"))
)
