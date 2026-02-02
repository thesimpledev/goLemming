// Package agent provides the autonomous agent implementation.
package agent

// State represents the current state of an agent.
type State int

const (
	StateIdle State = iota
	StateRunning
	StateCompleted
	StateFailed
	StateStopped
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateRunning:
		return "running"
	case StateCompleted:
		return "completed"
	case StateFailed:
		return "failed"
	case StateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// IsTerminal returns true if the state is a terminal state.
func (s State) IsTerminal() bool {
	return s == StateCompleted || s == StateFailed || s == StateStopped
}
