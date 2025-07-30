package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ia/gwt/internal/worktree"
)

type listModel struct {
	table      table.Model
	worktrees  []worktree.Worktree
	err        error
	quitting   bool
}

var (
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)
	
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true)
)

func NewListModel() listModel {
	columns := []table.Column{
		{Title: "Branch", Width: 30},
		{Title: "Path", Width: 50},
		{Title: "Status", Width: 15},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = headerStyle
	s.Selected = selectedStyle
	t.SetStyles(s)

	return listModel{
		table: t,
	}
}

func (m listModel) Init() tea.Cmd {
	return m.loadWorktrees
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			// TODO: Add switch functionality
			return m, nil
		case "d":
			// TODO: Add delete functionality
			return m, nil
		}

	case worktreesLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.worktrees = msg.worktrees
		
		rows := []table.Row{}
		for _, wt := range m.worktrees {
			status := "Clean"
			// TODO: Check git status
			
			// Make paths relative for display
			path := wt.Path
			if strings.Contains(path, "git-worktrees") {
				parts := strings.Split(path, "git-worktrees/")
				if len(parts) > 1 {
					path = "~/" + parts[1]
				}
			}
			
			rows = append(rows, table.Row{wt.Branch, path, status})
		}
		m.table.SetRows(rows)
		return m, nil
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m listModel) View() string {
	if m.err != nil {
		return errorStyle.Render("Error: " + m.err.Error())
	}

	s := titleStyle.Render("Git Worktrees") + "\n\n"
	
	if len(m.worktrees) == 0 {
		s += infoStyle.Render("No worktrees found for this project") + "\n"
		s += infoStyle.Render("Create one with: gwt new <branch-name>") + "\n"
	} else {
		s += m.table.View() + "\n\n"
		s += infoStyle.Render("↑/↓: Navigate • Enter: Switch • d: Delete • q: Quit")
	}

	return s
}

type worktreesLoadedMsg struct {
	worktrees []worktree.Worktree
	err       error
}

func (m listModel) loadWorktrees() tea.Msg {
	worktrees, err := worktree.List()
	return worktreesLoadedMsg{
		worktrees: worktrees,
		err:       err,
	}
}