package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nachoal/gwt/internal/config"
	"github.com/nachoal/gwt/internal/worktree"
)

type step struct {
	name   string
	status string // "pending", "running", "done", "error"
	err    error
}

type createModel struct {
	branchName     string
	fromBranch     string
	steps          []step
	currentStep    int
	spinner        spinner.Model
	done           bool
	err            error
	loadedConfig   *config.Config
	worktreePath   string
	currentCommand string
	setupCommands  []string
	setupHistory   []setupCommandResult
	setupIndex     int
	setupRunning   bool
	setupStartedAt time.Time
	setupElapsed   time.Duration
}

type setupCommandResult struct {
	command  string
	duration time.Duration
	err      error
}

var (
	checkMark = uiRenderer.NewStyle().Foreground(lipgloss.Color("42")).Render("✓")
	xMark     = uiRenderer.NewStyle().Foreground(lipgloss.Color("196")).Render("✗")
	bullet    = uiRenderer.NewStyle().Foreground(lipgloss.Color("241")).Render("•")

	stepStyle = uiRenderer.NewStyle().PaddingLeft(2)
)

func NewCreateModel(branchName, fromBranch string) createModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = uiRenderer.NewStyle().Foreground(lipgloss.Color("205"))

	return createModel{
		branchName: branchName,
		fromBranch: fromBranch,
		steps: []step{
			{name: "Loading configuration", status: "running"},
			{name: "Creating worktree", status: "pending"},
			{name: "Copying files", status: "pending"},
			{name: "Running setup commands", status: "pending"},
		},
		spinner: s,
	}
}

func (m createModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.runNextStep(),
	)
}

func (m createModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
		if m.done && msg.String() == "enter" {
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case setupStartMsg:
		m.setupCommands = append([]string(nil), msg.commands...)
		m.setupHistory = make([]setupCommandResult, 0, len(m.setupCommands))
		m.setupIndex = 0
		m.setupElapsed = 0

		if len(m.setupCommands) == 0 {
			m.steps[m.currentStep].status = "done"
			m.currentStep++
			if m.currentStep < len(m.steps) {
				m.steps[m.currentStep].status = "running"
				return m, m.runNextStep()
			}

			m.done = true
			// Automatically quit after a short delay to show the success message
			return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return tea.Quit()
			})
		}

		m.currentCommand = m.setupCommands[m.setupIndex]
		m.setupStartedAt = time.Now()
		m.setupRunning = true
		return m, tea.Batch(
			m.runSetupCommand(m.currentCommand),
			setupTickCmd(),
		)

	case setupTickMsg:
		if !m.setupRunning {
			return m, nil
		}
		m.setupElapsed = time.Since(m.setupStartedAt)
		return m, setupTickCmd()

	case setupCommandCompleteMsg:
		m.setupRunning = false
		m.setupElapsed = msg.duration
		m.setupHistory = append(m.setupHistory, setupCommandResult{
			command:  msg.command,
			duration: msg.duration,
			err:      msg.err,
		})

		if msg.err != nil {
			m.steps[m.currentStep].status = "error"
			m.steps[m.currentStep].err = msg.err
			m.err = msg.err
			m.done = true
			// Automatically quit after showing error
			return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
				return tea.Quit()
			})
		}

		m.setupIndex++
		if m.setupIndex < len(m.setupCommands) {
			m.currentCommand = m.setupCommands[m.setupIndex]
			m.setupStartedAt = time.Now()
			m.setupElapsed = 0
			m.setupRunning = true
			return m, tea.Batch(
				m.runSetupCommand(m.currentCommand),
				setupTickCmd(),
			)
		}

		m.currentCommand = ""
		m.setupElapsed = 0
		m.steps[m.currentStep].status = "done"
		m.currentStep++
		if m.currentStep < len(m.steps) {
			m.steps[m.currentStep].status = "running"
			return m, m.runNextStep()
		}

		m.done = true
		// Automatically quit after a short delay to show the success message
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tea.Quit()
		})

	case stepCompleteMsg:
		if msg.err != nil {
			m.steps[m.currentStep].status = "error"
			m.steps[m.currentStep].err = msg.err
			m.err = msg.err
			m.done = true
			// Automatically quit after showing error
			return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
				return tea.Quit()
			})
		}

		// Store data from completed steps
		if msg.worktreePath != "" {
			m.worktreePath = msg.worktreePath
		}
		if msg.config != nil {
			m.loadedConfig = msg.config
		}

		m.steps[m.currentStep].status = "done"
		m.currentStep++
		if m.currentStep < len(m.steps) {
			m.steps[m.currentStep].status = "running"
			return m, m.runNextStep()
		}

		m.done = true
		// Automatically quit after a short delay to show the success message
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tea.Quit()
		})
	}

	return m, nil
}

func (m createModel) View() string {
	s := titleStyle.Render("Creating worktree") + "\n\n"

	for _, step := range m.steps {
		icon := bullet
		if step.status == "done" {
			icon = checkMark
		} else if step.status == "error" {
			icon = xMark
		} else if step.status == "running" {
			icon = m.spinner.View()
		}

		line := fmt.Sprintf("%s %s", icon, step.name)
		if step.err != nil {
			line += "\n" + stepStyle.Render(errorStyle.Render("  → "+step.err.Error()))
		}
		s += stepStyle.Render(line) + "\n"
	}

	if len(m.setupCommands) > 0 || len(m.setupHistory) > 0 {
		s += "\n"
		s += infoStyle.Render("Setup command trace") + "\n"
		total := len(m.setupCommands)
		for idx, result := range m.setupHistory {
			icon := checkMark
			if result.err != nil {
				icon = xMark
			}
			s += stepStyle.Render(fmt.Sprintf(
				"%s [%d/%d] %s (%s)",
				icon,
				idx+1,
				total,
				result.command,
				formatSetupDuration(result.duration),
			)) + "\n"
		}

		if m.setupRunning && m.currentCommand != "" {
			s += stepStyle.Render(fmt.Sprintf(
				"%s [%d/%d] %s (%s)",
				m.spinner.View(),
				m.setupIndex+1,
				total,
				m.currentCommand,
				formatSetupDuration(m.setupElapsed),
			)) + "\n"
		}
	}

	if m.done {
		s += "\n"
		if m.err != nil {
			s += errorStyle.Render("Failed to create worktree") + "\n"
		} else {
			s += uiRenderer.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).
				Render("✓ Worktree created successfully!") + "\n\n"
			s += uiRenderer.NewStyle().Foreground(lipgloss.Color("86")).
				Render("📁 "+m.worktreePath) + "\n"
		}
	}

	return s
}

type stepCompleteMsg struct {
	err          error
	worktreePath string
	config       *config.Config
}

type setupStartMsg struct {
	commands []string
}

type setupCommandCompleteMsg struct {
	command  string
	duration time.Duration
	err      error
}

type setupTickMsg time.Time

func setupTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return setupTickMsg(t)
	})
}

func (m createModel) runSetupCommand(command string) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		err := worktree.RunSetupCommands(m.worktreePath, []string{command})
		return setupCommandCompleteMsg{
			command:  command,
			duration: time.Since(start),
			err:      err,
		}
	}
}

func (m createModel) runNextStep() tea.Cmd {
	return func() tea.Msg {
		switch m.currentStep {
		case 0: // Load configuration
			time.Sleep(100 * time.Millisecond) // Brief pause for visual effect
			cfg, err := config.LoadConfig()
			if err != nil {
				return stepCompleteMsg{err: err}
			}
			return stepCompleteMsg{config: cfg}

		case 1: // Create worktree
			projectName, err := worktree.GetProjectName()
			if err != nil {
				return stepCompleteMsg{err: err}
			}

			cfg, err := m.getConfig()
			if err != nil {
				return stepCompleteMsg{err: err}
			}
			targetPath := worktree.GetWorktreePath(cfg.Settings.Root, projectName, m.branchName)

			if err := worktree.Create(m.branchName, m.fromBranch, targetPath); err != nil {
				return stepCompleteMsg{err: err}
			}
			return stepCompleteMsg{worktreePath: targetPath}

		case 2: // Copy files
			if m.worktreePath == "" {
				return stepCompleteMsg{err: fmt.Errorf("worktree path not set")}
			}

			cfg, err := m.getConfig()
			if err != nil {
				return stepCompleteMsg{err: err}
			}

			// Get the current working directory (main worktree)
			mainPath, err := os.Getwd()
			if err != nil {
				return stepCompleteMsg{err: err}
			}

			if err := worktree.CopyFiles(mainPath, m.worktreePath, cfg.Copy); err != nil {
				return stepCompleteMsg{err: err}
			}
			return stepCompleteMsg{}

		case 3: // Run setup commands
			if m.worktreePath == "" {
				return stepCompleteMsg{err: fmt.Errorf("worktree path not set")}
			}

			cfg, err := m.getConfig()
			if err != nil {
				return stepCompleteMsg{err: err}
			}

			return setupStartMsg{commands: cfg.Setup}
		}

		return stepCompleteMsg{err: fmt.Errorf("unknown step: %d", m.currentStep)}
	}
}

func (m createModel) getConfig() (*config.Config, error) {
	if m.loadedConfig != nil {
		return m.loadedConfig, nil
	}
	return config.LoadConfig()
}

func formatSetupDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < time.Second {
		return d.Round(10 * time.Millisecond).String()
	}
	return d.Round(100 * time.Millisecond).String()
}
