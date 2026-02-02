package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/thesimpledev/golemming/internal/action"
	"github.com/thesimpledev/golemming/internal/capture"
	"github.com/thesimpledev/golemming/internal/config"
	"github.com/thesimpledev/golemming/internal/llm"
	"github.com/thesimpledev/golemming/pkg/protocol"
)

// Agent represents an autonomous desktop automation agent.
type Agent struct {
	config   *config.Config
	client   *llm.Client
	executor *action.Executor
	history  *History

	goal    string
	state   State
	result  string
	mu      sync.RWMutex

	onAction func(action *protocol.Action, result *action.Result)
}

// New creates a new agent.
func New(cfg *config.Config) *Agent {
	return &Agent{
		config:   cfg,
		client:   llm.NewClient(cfg.APIKey, cfg.Model),
		executor: action.NewExecutor(cfg.RequireAbsolutePaths),
		history:  NewHistory(),
		state:    StateIdle,
	}
}

// OnAction sets a callback for when an action is executed.
func (a *Agent) OnAction(fn func(action *protocol.Action, result *action.Result)) {
	a.onAction = fn
}

// Run starts the agent with the given goal.
func (a *Agent) Run(ctx context.Context, goal string) error {
	a.mu.Lock()
	if a.state != StateIdle {
		a.mu.Unlock()
		return fmt.Errorf("agent is not idle (current state: %s)", a.state)
	}
	a.goal = goal
	a.state = StateRunning
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		if a.state == StateRunning {
			a.state = StateStopped
		}
		a.mu.Unlock()
	}()

	for i := 0; i < a.config.MaxIterations; i++ {
		select {
		case <-ctx.Done():
			a.mu.Lock()
			a.state = StateStopped
			a.mu.Unlock()
			return ctx.Err()
		default:
		}

		done, err := a.step(ctx)
		if err != nil {
			return err
		}
		if done {
			return nil
		}

		// Wait for UI stabilization
		time.Sleep(a.config.StabilizationDelay())
	}

	a.mu.Lock()
	a.state = StateFailed
	a.result = "max iterations reached"
	a.mu.Unlock()
	return fmt.Errorf("max iterations (%d) reached", a.config.MaxIterations)
}

// step executes a single agent step. Returns true if the agent is done.
func (a *Agent) step(ctx context.Context) (bool, error) {
	// Capture screenshot
	screenshot, err := capture.CaptureAndEncode()
	if err != nil {
		return false, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	// Get action from LLM
	nextAction, err := a.client.GetAction(ctx, a.goal, screenshot, a.history.GetLLMHistory())
	if err != nil {
		return false, fmt.Errorf("failed to get action from LLM: %w", err)
	}

	// Check for terminal actions
	if nextAction.Type == protocol.ActionDone {
		a.mu.Lock()
		a.state = StateCompleted
		a.result = nextAction.Summary
		a.mu.Unlock()
		return true, nil
	}

	if nextAction.Type == protocol.ActionFailed {
		a.mu.Lock()
		a.state = StateFailed
		a.result = nextAction.Reason
		a.mu.Unlock()
		return true, nil
	}

	// Execute the action
	result := a.executor.Execute(nextAction)

	// Record in history
	historyEntry := action.ToHistoryEntry(nextAction, result)
	a.history.Add(historyEntry)

	// Notify callback
	if a.onAction != nil {
		a.onAction(nextAction, result)
	}

	return false, nil
}

// State returns the current agent state.
func (a *Agent) State() State {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// Result returns the agent result (summary or failure reason).
func (a *Agent) Result() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.result
}

// History returns the action history.
func (a *Agent) History() *History {
	return a.history
}

// Stop stops the agent.
func (a *Agent) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.state == StateRunning {
		a.state = StateStopped
	}
}
