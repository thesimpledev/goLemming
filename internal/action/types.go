// Package action provides action execution functionality.
package action

import (
	"github.com/thesimpledev/golemming/internal/llm"
	"github.com/thesimpledev/golemming/pkg/protocol"
)

// Result represents the result of executing an action.
type Result struct {
	Success bool
	Error   string
	Data    string // For file_read, contains file contents
}

// ToHistoryEntry converts an action and result to a history entry.
func ToHistoryEntry(action *protocol.Action, result *Result) llm.HistoryEntry {
	entry := llm.HistoryEntry{
		Action: llm.ActionRecord{
			Type:      string(action.Type),
			X:         action.X,
			Y:         action.Y,
			Button:    action.Button,
			Double:    action.Double,
			Text:      action.Text,
			Key:       action.Key,
			Direction: action.Direction,
			Amount:    action.Amount,
			Path:      action.Path,
			Content:   action.Content,
			Ms:        action.Ms,
			Summary:   action.Summary,
			Reason:    action.Reason,
		},
	}
	if !result.Success {
		entry.Error = result.Error
	}
	return entry
}
