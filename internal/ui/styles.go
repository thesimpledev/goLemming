// Package ui provides the terminal user interface using Bubble Tea.
package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	Primary   = lipgloss.Color("#7C3AED") // Purple
	Secondary = lipgloss.Color("#10B981") // Green
	Accent    = lipgloss.Color("#F59E0B") // Amber
	Error     = lipgloss.Color("#EF4444") // Red
	Muted     = lipgloss.Color("#6B7280") // Gray
	Text      = lipgloss.Color("#F9FAFB") // Light
	Dim       = lipgloss.Color("#9CA3AF") // Dimmed text

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Italic(true)

	PromptStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	InputStyle = lipgloss.NewStyle().
			Foreground(Text)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Secondary)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Accent)

	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	DimStyle = lipgloss.NewStyle().
			Foreground(Dim)

	// Action styles
	ActionTypeStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			Width(12)

	ActionDetailStyle = lipgloss.NewStyle().
				Foreground(Text)

	TimestampStyle = lipgloss.NewStyle().
			Foreground(Muted)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Muted).
			Padding(0, 1)

	FocusedBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(0, 1)

	// Status styles
	StatusRunning = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	StatusComplete = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)

	StatusFailed = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	// Help style
	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginTop(1)

	// Header
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			Background(lipgloss.Color("#1F2937")).
			Padding(0, 2).
			MarginBottom(1)

	// Logo
	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary)
)

const Logo = `
   ▄████  ▒█████   ██▓    ▓█████  ███▄ ▄███▓ ███▄ ▄███▓ ██▓ ███▄    █   ▄████
  ██▒ ▀█▒▒██▒  ██▒▓██▒    ▓█   ▀ ▓██▒▀█▀ ██▒▓██▒▀█▀ ██▒▓██▒ ██ ▀█   █  ██▒ ▀█▒
 ▒██░▄▄▄░▒██░  ██▒▒██░    ▒███   ▓██    ▓██░▓██    ▓██░▒██▒▓██  ▀█ ██▒▒██░▄▄▄░
 ░▓█  ██▓▒██   ██░▒██░    ▒▓█  ▄ ▒██    ▒██ ▒██    ▒██ ░██░▓██▒  ▐▌██▒░▓█  ██▓
 ░▒▓███▀▒░ ████▓▒░░██████▒░▒████▒▒██▒   ░██▒▒██▒   ░██▒░██░▒██░   ▓██░░▒▓███▀▒
  ░▒   ▒ ░ ▒░▒░▒░ ░ ▒░▓  ░░░ ▒░ ░░ ▒░   ░  ░░ ▒░   ░  ░░▓  ░ ▒░   ▒ ▒  ░▒   ▒
   ░   ░   ░ ▒ ▒░ ░ ░ ▒  ░ ░ ░  ░░  ░      ░░  ░      ░ ▒ ░░ ░░   ░ ▒░  ░   ░
 ░ ░   ░ ░ ░ ░ ▒    ░ ░      ░   ░      ░   ░      ░    ▒ ░   ░   ░ ░ ░ ░   ░
       ░     ░ ░      ░  ░   ░  ░       ░          ░    ░           ░       ░`

const LogoSmall = `╭─────────────────────────╮
│  🤖 GoLemming v0.1.0   │
╰─────────────────────────╯`
