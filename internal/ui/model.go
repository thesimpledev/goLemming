package ui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thesimpledev/golemming/internal/action"
	"github.com/thesimpledev/golemming/internal/agent"
	"github.com/thesimpledev/golemming/internal/config"
	"github.com/thesimpledev/golemming/pkg/protocol"
)

// View represents the current view/mode of the application.
type View int

const (
	ViewSetup View = iota
	ViewInput
	ViewRunning
	ViewComplete
	ViewHelp
)

// HistoryItem represents an executed action in the history.
type HistoryItem struct {
	Timestamp time.Time
	Action    *protocol.Action
	Result    *action.Result
}

// Model represents the application state.
type Model struct {
	view         View
	width        int
	height       int
	config       *config.Config
	configLoaded bool

	// Setup view
	apiKeyInput textinput.Model
	setupError  string

	// Input view
	goalInput   textinput.Model
	goalHistory []string

	// Running view
	currentGoal   string
	agent         *agent.Agent
	agentCancel   context.CancelFunc
	actionHistory []HistoryItem
	spinner       spinner.Model
	iterationNum  int
	updateCh      chan tea.Msg

	// Complete view
	finalStatus  string
	finalMessage string

	// Help view
	prevView View
}

// New creates a new Model.
func New() Model {
	// API key input
	apiInput := textinput.New()
	apiInput.Placeholder = "sk-ant-..."
	apiInput.CharLimit = 200
	apiInput.Width = 60
	apiInput.EchoMode = textinput.EchoPassword
	apiInput.EchoCharacter = '•'

	// Goal input
	goalInput := textinput.New()
	goalInput.Placeholder = "Enter your goal..."
	goalInput.CharLimit = 500
	goalInput.Width = 80

	// Spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(Primary)

	m := Model{
		view:        ViewSetup,
		apiKeyInput: apiInput,
		goalInput:   goalInput,
		spinner:     s,
		goalHistory: make([]string, 0),
		updateCh:    make(chan tea.Msg, 100),
	}

	// Try to load existing config
	cfg, err := config.Load()
	if err == nil {
		m.config = cfg
		m.configLoaded = true
		m.view = ViewInput
		m.goalInput.Focus()
	} else {
		m.apiKeyInput.Focus()
	}

	return m
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
	)
}

// waitForUpdate listens on the update channel and returns messages.
func (m Model) waitForUpdate() tea.Msg {
	return <-m.updateCh
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case ActionExecutedMsg:
		m.actionHistory = append(m.actionHistory, HistoryItem{
			Timestamp: time.Now(),
			Action:    msg.Action,
			Result:    msg.Result,
		})
		m.iterationNum++
		// Continue listening for more updates
		return m, m.waitForUpdate

	case AgentDoneMsg:
		m.view = ViewComplete
		if msg.Success {
			m.finalStatus = "completed"
		} else {
			m.finalStatus = "failed"
		}
		m.finalMessage = msg.Message
		return m, nil

	case AgentErrorMsg:
		m.view = ViewComplete
		m.finalStatus = "error"
		m.finalMessage = msg.Error.Error()
		return m, nil
	}

	// Update focused inputs
	var cmd tea.Cmd
	switch m.view {
	case ViewSetup:
		m.apiKeyInput, cmd = m.apiKeyInput.Update(msg)
	case ViewInput:
		m.goalInput, cmd = m.goalInput.Update(msg)
	}

	return m, cmd
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		if m.view == ViewRunning && m.agentCancel != nil {
			m.agentCancel()
			m.view = ViewComplete
			m.finalStatus = "stopped"
			m.finalMessage = "Stopped by user"
			return m, nil
		}
		return m, tea.Quit

	case "?":
		// Show help (except when typing in input)
		if m.view == ViewInput && m.goalInput.Value() == "" {
			m.prevView = m.view
			m.view = ViewHelp
			return m, nil
		}

	case "esc":
		if m.view == ViewHelp {
			m.view = m.prevView
			if m.view == ViewInput {
				m.goalInput.Focus()
			}
			return m, nil
		}
		if m.view == ViewRunning && m.agentCancel != nil {
			m.agentCancel()
			m.view = ViewComplete
			m.finalStatus = "stopped"
			m.finalMessage = "Stopped by user"
			return m, nil
		}
		if m.view == ViewComplete {
			// Return to input view
			m.view = ViewInput
			m.actionHistory = nil
			m.iterationNum = 0
			m.goalInput.SetValue("")
			m.goalInput.Focus()
			return m, nil
		}

	case "enter":
		if m.view == ViewHelp {
			m.view = m.prevView
			if m.view == ViewInput {
				m.goalInput.Focus()
			}
			return m, nil
		}
		return m.handleEnter()
	}

	// Update focused inputs
	var cmd tea.Cmd
	switch m.view {
	case ViewSetup:
		m.apiKeyInput, cmd = m.apiKeyInput.Update(msg)
	case ViewInput:
		m.goalInput, cmd = m.goalInput.Update(msg)
	}

	return m, cmd
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.view {
	case ViewSetup:
		apiKey := strings.TrimSpace(m.apiKeyInput.Value())
		if apiKey == "" {
			m.setupError = "API key is required"
			return m, nil
		}
		if !strings.HasPrefix(apiKey, "sk-") {
			m.setupError = "Invalid API key format (should start with sk-)"
			return m, nil
		}

		// Create config with the API key and save it
		m.config = config.DefaultConfig()
		m.config.APIKey = apiKey
		if err := m.config.Save(); err != nil {
			m.setupError = "Failed to save config: " + err.Error()
			return m, nil
		}

		m.configLoaded = true
		m.view = ViewInput
		m.apiKeyInput.Blur()
		m.goalInput.Focus()
		return m, nil

	case ViewInput:
		goal := strings.TrimSpace(m.goalInput.Value())
		if goal == "" {
			return m, nil
		}

		// Save to history
		m.goalHistory = append(m.goalHistory, goal)
		m.currentGoal = goal
		m.actionHistory = nil
		m.iterationNum = 0

		// Start agent
		m.view = ViewRunning
		m.goalInput.Blur()

		return m, m.startAgent(goal)

	case ViewComplete:
		// Return to input view
		m.view = ViewInput
		m.actionHistory = nil
		m.iterationNum = 0
		m.goalInput.SetValue("")
		m.goalInput.Focus()
		return m, nil
	}

	return m, nil
}

func (m *Model) startAgent(goal string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.agentCancel = cancel

	ag := agent.New(m.config)
	m.agent = ag

	updateCh := m.updateCh

	// Set up action callback to send to channel
	ag.OnAction(func(act *protocol.Action, result *action.Result) {
		select {
		case updateCh <- ActionExecutedMsg{Action: act, Result: result}:
		default:
		}
	})

	// Run agent in goroutine
	go func() {
		err := ag.Run(ctx, goal)

		// Send completion message
		state := ag.State()
		switch state {
		case agent.StateCompleted:
			select {
			case updateCh <- AgentDoneMsg{Success: true, Message: ag.Result()}:
			default:
			}
		case agent.StateFailed:
			select {
			case updateCh <- AgentDoneMsg{Success: false, Message: ag.Result()}:
			default:
			}
		case agent.StateStopped:
			// Already handled by cancel
		default:
			if err != nil && ctx.Err() != context.Canceled {
				select {
				case updateCh <- AgentErrorMsg{Error: err}:
				default:
				}
			}
		}
	}()

	// Return command to start listening for updates
	return m.waitForUpdate
}

// View renders the model.
func (m Model) View() string {
	switch m.view {
	case ViewSetup:
		return m.viewSetup()
	case ViewInput:
		return m.viewInput()
	case ViewRunning:
		return m.viewRunning()
	case ViewComplete:
		return m.viewComplete()
	case ViewHelp:
		return m.viewHelp()
	}
	return ""
}

func (m Model) viewSetup() string {
	var b strings.Builder

	b.WriteString(LogoStyle.Render(LogoSmall))
	b.WriteString("\n\n")

	b.WriteString(TitleStyle.Render("Welcome to GoLemming"))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render("Autonomous desktop automation agent"))
	b.WriteString("\n\n")

	b.WriteString(MutedStyle.Render("To get started, please enter your Anthropic API key:"))
	b.WriteString("\n")
	b.WriteString(DimStyle.Render("(Your key will be saved to ~/.golemming/config.json)"))
	b.WriteString("\n\n")

	b.WriteString(PromptStyle.Render("API Key: "))
	b.WriteString(m.apiKeyInput.View())
	b.WriteString("\n")

	if m.setupError != "" {
		b.WriteString("\n")
		b.WriteString(ErrorStyle.Render("Error: " + m.setupError))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(DimStyle.Render("Tip: You can also set the ANTHROPIC_API_KEY environment variable"))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Press Enter to continue • Ctrl+C to quit"))

	return b.String()
}

func (m Model) viewInput() string {
	var b strings.Builder

	b.WriteString(LogoStyle.Render(LogoSmall))
	b.WriteString("\n\n")

	b.WriteString(MutedStyle.Render("Enter a goal for the agent to accomplish:"))
	b.WriteString("\n\n")

	b.WriteString(PromptStyle.Render("❯ "))
	b.WriteString(m.goalInput.View())
	b.WriteString("\n")

	if len(m.goalHistory) > 0 {
		b.WriteString("\n")
		b.WriteString(DimStyle.Render("Recent goals:"))
		b.WriteString("\n")
		start := len(m.goalHistory) - 3
		if start < 0 {
			start = 0
		}
		for i := start; i < len(m.goalHistory); i++ {
			goal := m.goalHistory[i]
			if len(goal) > 60 {
				goal = goal[:60] + "..."
			}
			b.WriteString(DimStyle.Render(fmt.Sprintf("  %d. %s", i+1, goal)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Enter to run • ? for help • Ctrl+C to quit"))

	return b.String()
}

func (m Model) viewRunning() string {
	var b strings.Builder

	// Header
	b.WriteString(m.spinner.View())
	b.WriteString(" ")
	b.WriteString(StatusRunning.Render("Running"))
	b.WriteString(MutedStyle.Render(fmt.Sprintf(" • Iteration %d", m.iterationNum)))
	b.WriteString("\n\n")

	// Goal
	b.WriteString(MutedStyle.Render("Goal: "))
	b.WriteString(m.currentGoal)
	b.WriteString("\n\n")

	// Action history
	if len(m.actionHistory) > 0 {
		b.WriteString(MutedStyle.Render("Actions:"))
		b.WriteString("\n")

		// Show last 10 actions
		start := len(m.actionHistory) - 10
		if start < 0 {
			start = 0
		}

		for i := start; i < len(m.actionHistory); i++ {
			item := m.actionHistory[i]
			b.WriteString(m.formatActionLine(item))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Press Esc or Ctrl+C to stop"))

	return b.String()
}

func (m Model) viewComplete() string {
	var b strings.Builder

	// Status
	switch m.finalStatus {
	case "completed":
		b.WriteString(StatusComplete.Render("✓ Completed"))
	case "failed":
		b.WriteString(StatusFailed.Render("✗ Failed"))
	case "stopped":
		b.WriteString(WarningStyle.Render("⏹ Stopped"))
	case "error":
		b.WriteString(ErrorStyle.Render("⚠ Error"))
	}
	b.WriteString("\n\n")

	// Goal
	b.WriteString(MutedStyle.Render("Goal: "))
	b.WriteString(m.currentGoal)
	b.WriteString("\n\n")

	// Result message
	if m.finalMessage != "" {
		b.WriteString(MutedStyle.Render("Result: "))
		b.WriteString(m.finalMessage)
		b.WriteString("\n\n")
	}

	// Summary
	b.WriteString(MutedStyle.Render(fmt.Sprintf("Total actions: %d", len(m.actionHistory))))
	b.WriteString("\n")

	// Action history summary
	if len(m.actionHistory) > 0 {
		b.WriteString("\n")
		b.WriteString(MutedStyle.Render("Action log:"))
		b.WriteString("\n")

		for _, item := range m.actionHistory {
			b.WriteString(m.formatActionLine(item))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Press Enter or Esc for new goal • Ctrl+C to quit"))

	return b.String()
}

func (m Model) formatActionLine(item HistoryItem) string {
	timestamp := TimestampStyle.Render(item.Timestamp.Format("15:04:05"))
	actionType := ActionTypeStyle.Render(string(item.Action.Type))
	detail := m.formatActionDetail(item.Action)

	status := SuccessStyle.Render("✓")
	if !item.Result.Success {
		status = ErrorStyle.Render("✗ " + item.Result.Error)
	}

	return fmt.Sprintf("  %s %s %s %s", timestamp, actionType, detail, status)
}

func (m Model) formatActionDetail(act *protocol.Action) string {
	switch act.Type {
	case protocol.ActionClick:
		btn := act.Button
		if btn == "" {
			btn = "left"
		}
		if act.Double {
			return ActionDetailStyle.Render(fmt.Sprintf("double %s at (%d, %d)", btn, act.X, act.Y))
		}
		return ActionDetailStyle.Render(fmt.Sprintf("%s at (%d, %d)", btn, act.X, act.Y))
	case protocol.ActionType_:
		text := act.Text
		if len(text) > 30 {
			text = text[:30] + "..."
		}
		return ActionDetailStyle.Render(fmt.Sprintf("%q", text))
	case protocol.ActionKey:
		return ActionDetailStyle.Render(act.Key)
	case protocol.ActionScroll:
		return ActionDetailStyle.Render(fmt.Sprintf("%s %d", act.Direction, act.Amount))
	case protocol.ActionFileRead:
		return ActionDetailStyle.Render(filepath.Base(act.Path))
	case protocol.ActionFileWrite:
		return ActionDetailStyle.Render(filepath.Base(act.Path))
	case protocol.ActionWait:
		return ActionDetailStyle.Render(fmt.Sprintf("%dms", act.Ms))
	default:
		return ""
	}
}

func (m Model) viewHelp() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("GoLemming Help"))
	b.WriteString("\n\n")

	b.WriteString(SubtitleStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"Enter", "Execute goal / Continue"},
		{"Esc", "Stop agent / Go back / New goal"},
		{"Ctrl+C", "Stop agent / Quit application"},
		{"?", "Show this help screen"},
	}

	for _, s := range shortcuts {
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			PromptStyle.Render(fmt.Sprintf("%-10s", s.key)),
			MutedStyle.Render(s.desc)))
	}

	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render("Available Actions"))
	b.WriteString("\n\n")

	actions := []struct {
		name string
		desc string
	}{
		{"click", "Click at screen coordinates"},
		{"type", "Type text characters"},
		{"key", "Press keyboard key/combo"},
		{"scroll", "Scroll mouse wheel"},
		{"file_read", "Read file contents"},
		{"file_write", "Write content to file"},
		{"wait", "Wait for UI stabilization"},
		{"done", "Task completed successfully"},
		{"failed", "Task cannot be completed"},
	}

	for _, a := range actions {
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			ActionTypeStyle.Render(fmt.Sprintf("%-12s", a.name)),
			MutedStyle.Render(a.desc)))
	}

	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render("Command Line Usage"))
	b.WriteString("\n\n")

	b.WriteString(DimStyle.Render("  Interactive mode (TUI):"))
	b.WriteString("\n")
	b.WriteString("    golemming\n\n")

	b.WriteString(DimStyle.Render("  Headless mode (for scripts):"))
	b.WriteString("\n")
	b.WriteString("    golemming -goal \"Open Calculator\"\n")
	b.WriteString("    golemming -goal \"...\" -max-iterations 50\n")

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Press Enter or Esc to go back"))

	return b.String()
}
