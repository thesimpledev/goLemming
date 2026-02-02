package llm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/thesimpledev/golemming/pkg/protocol"
)

// ParseAction parses an action from LLM response text.
func ParseAction(response string) (*protocol.Action, error) {
	// Clean up the response - remove markdown code blocks if present
	cleaned := cleanResponse(response)

	// Parse JSON
	var action protocol.Action
	if err := json.Unmarshal([]byte(cleaned), &action); err != nil {
		return nil, fmt.Errorf("failed to parse action JSON: %w\nResponse was: %s", err, truncate(response, 200))
	}

	// Validate the action
	if err := action.Validate(); err != nil {
		return nil, fmt.Errorf("invalid action: %w", err)
	}

	return &action, nil
}

// cleanResponse removes markdown code blocks and extra whitespace.
func cleanResponse(response string) string {
	response = strings.TrimSpace(response)

	// Remove markdown code blocks
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
	}

	if strings.HasSuffix(response, "```") {
		response = strings.TrimSuffix(response, "```")
	}

	response = strings.TrimSpace(response)

	// Find the JSON object boundaries
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start >= 0 && end > start {
		response = response[start : end+1]
	}

	return response
}

// truncate truncates a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
