package ui

import (
	"github.com/thesimpledev/golemming/internal/action"
	"github.com/thesimpledev/golemming/pkg/protocol"
)

// ActionExecutedMsg is sent when an action is executed.
type ActionExecutedMsg struct {
	Action *protocol.Action
	Result *action.Result
}

// AgentDoneMsg is sent when the agent completes.
type AgentDoneMsg struct {
	Success bool
	Message string
}

// AgentErrorMsg is sent when the agent encounters an error.
type AgentErrorMsg struct {
	Error error
}

// ScreenshotMsg is sent when a screenshot is captured.
type ScreenshotMsg struct{}

// ConfigSavedMsg is sent when configuration is saved.
type ConfigSavedMsg struct{}

// TickMsg is sent for periodic updates.
type TickMsg struct{}
