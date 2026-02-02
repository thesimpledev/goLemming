// Package protocol defines the action types and structures for agent-LLM communication.
package protocol

// ActionType represents the type of action an agent can perform.
type ActionType string

const (
	ActionClick     ActionType = "click"
	ActionType_     ActionType = "type" // Named ActionType_ to avoid conflict with Go's type keyword
	ActionKey       ActionType = "key"
	ActionScroll    ActionType = "scroll"
	ActionFileRead  ActionType = "file_read"
	ActionFileWrite ActionType = "file_write"
	ActionWait      ActionType = "wait"
	ActionDone      ActionType = "done"
	ActionFailed    ActionType = "failed"
)

// Action represents an action to be performed by the agent.
type Action struct {
	Type      ActionType `json:"type"`
	X         int        `json:"x,omitempty"`
	Y         int        `json:"y,omitempty"`
	Button    string     `json:"button,omitempty"`
	Double    bool       `json:"double,omitempty"`
	Text      string     `json:"text,omitempty"`
	Key       string     `json:"key,omitempty"`
	Direction string     `json:"direction,omitempty"`
	Amount    int        `json:"amount,omitempty"`
	Path      string     `json:"path,omitempty"`
	Content   string     `json:"content,omitempty"`
	Ms        int        `json:"ms,omitempty"`
	Summary   string     `json:"summary,omitempty"`
	Reason    string     `json:"reason,omitempty"`
}

// Validate checks if the action has valid fields for its type.
func (a *Action) Validate() error {
	switch a.Type {
	case ActionClick:
		if a.Button == "" {
			a.Button = "left"
		}
	case ActionType_:
		if a.Text == "" {
			return &ValidationError{Field: "text", Message: "text is required for type action"}
		}
	case ActionKey:
		if a.Key == "" {
			return &ValidationError{Field: "key", Message: "key is required for key action"}
		}
	case ActionScroll:
		if a.Direction == "" {
			return &ValidationError{Field: "direction", Message: "direction is required for scroll action"}
		}
		if a.Amount == 0 {
			a.Amount = 3
		}
	case ActionFileRead:
		if a.Path == "" {
			return &ValidationError{Field: "path", Message: "path is required for file_read action"}
		}
	case ActionFileWrite:
		if a.Path == "" {
			return &ValidationError{Field: "path", Message: "path is required for file_write action"}
		}
	case ActionWait:
		if a.Ms == 0 {
			a.Ms = 500
		}
	case ActionDone:
		// Summary is optional
	case ActionFailed:
		if a.Reason == "" {
			return &ValidationError{Field: "reason", Message: "reason is required for failed action"}
		}
	default:
		return &ValidationError{Field: "type", Message: "unknown action type: " + string(a.Type)}
	}
	return nil
}

// ValidationError represents a validation error for an action field.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
