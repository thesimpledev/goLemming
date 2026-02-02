package action

import (
	"strings"
	"time"

	"github.com/thesimpledev/golemming/internal/input"
	"github.com/thesimpledev/golemming/pkg/protocol"
)

// Executor handles action execution.
type Executor struct {
	requireAbsolutePaths bool
}

// NewExecutor creates a new action executor.
func NewExecutor(requireAbsolutePaths bool) *Executor {
	return &Executor{
		requireAbsolutePaths: requireAbsolutePaths,
	}
}

// Execute executes an action and returns the result.
func (e *Executor) Execute(action *protocol.Action) *Result {
	switch action.Type {
	case protocol.ActionClick:
		return e.executeClick(action)
	case protocol.ActionType_:
		return e.executeType(action)
	case protocol.ActionKey:
		return e.executeKey(action)
	case protocol.ActionScroll:
		return e.executeScroll(action)
	case protocol.ActionFileRead:
		return e.executeFileRead(action)
	case protocol.ActionFileWrite:
		return e.executeFileWrite(action)
	case protocol.ActionWait:
		return e.executeWait(action)
	case protocol.ActionDone, protocol.ActionFailed:
		// These are handled by the agent, not the executor
		return &Result{Success: true}
	default:
		return &Result{Success: false, Error: "unknown action type: " + string(action.Type)}
	}
}

func (e *Executor) executeClick(action *protocol.Action) *Result {
	input.Move(action.X, action.Y)
	time.Sleep(50 * time.Millisecond) // Small delay for cursor to settle

	button := action.Button
	if button == "" {
		button = "left"
	}

	input.Click(button, action.Double)
	return &Result{Success: true}
}

func (e *Executor) executeType(action *protocol.Action) *Result {
	input.TypeText(action.Text)
	return &Result{Success: true}
}

func (e *Executor) executeKey(action *protocol.Action) *Result {
	key := action.Key

	// Check if it's a combo (contains +)
	if strings.Contains(key, "+") {
		input.KeyCombo(key)
	} else {
		input.KeyPress(key)
	}

	return &Result{Success: true}
}

func (e *Executor) executeScroll(action *protocol.Action) *Result {
	amount := action.Amount
	if amount == 0 {
		amount = 3
	}

	input.ScrollDir(amount, action.Direction)
	return &Result{Success: true}
}

func (e *Executor) executeFileRead(action *protocol.Action) *Result {
	content, err := ReadFile(action.Path, e.requireAbsolutePaths)
	if err != nil {
		return &Result{Success: false, Error: err.Error()}
	}
	return &Result{Success: true, Data: content}
}

func (e *Executor) executeFileWrite(action *protocol.Action) *Result {
	err := WriteFile(action.Path, action.Content, e.requireAbsolutePaths)
	if err != nil {
		return &Result{Success: false, Error: err.Error()}
	}
	return &Result{Success: true}
}

func (e *Executor) executeWait(action *protocol.Action) *Result {
	ms := action.Ms
	if ms == 0 {
		ms = 500
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return &Result{Success: true}
}
