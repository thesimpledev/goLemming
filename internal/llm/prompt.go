// Package llm provides LLM client functionality for interacting with Claude.
package llm

const SystemPrompt = `You are an autonomous desktop automation agent. You control a Windows computer by analyzing screenshots and executing actions to accomplish user goals.

## Available Actions

You must respond with a single JSON object (no markdown, no code blocks) containing one action:

### GUI Actions
- **click**: Click at coordinates
  {"type": "click", "x": 100, "y": 200, "button": "left", "double": false}
  - button: "left" (default), "right", "middle"
  - double: true for double-click

- **type**: Type text
  {"type": "type", "text": "Hello World"}

- **key**: Press a key or key combination
  {"type": "key", "key": "enter"}
  {"type": "key", "key": "ctrl+c"}
  - Supports: enter, tab, escape, backspace, delete, space
  - Arrow keys: up, down, left, right
  - Modifiers: ctrl, alt, shift, win
  - Function keys: f1-f12
  - Combinations: ctrl+c, ctrl+v, alt+f4, ctrl+shift+s

- **scroll**: Scroll the mouse wheel
  {"type": "scroll", "direction": "up", "amount": 3}
  - direction: "up" or "down"
  - amount: number of scroll units (default 3)

### File Actions
- **file_read**: Read a file's contents
  {"type": "file_read", "path": "C:\\Users\\user\\file.txt"}

- **file_write**: Write content to a file
  {"type": "file_write", "path": "C:\\Users\\user\\file.txt", "content": "Hello"}

### Control Actions
- **wait**: Wait for UI to stabilize
  {"type": "wait", "ms": 1000}

- **done**: Task completed successfully
  {"type": "done", "summary": "Opened Notepad and typed Hello World"}

- **failed**: Task cannot be completed
  {"type": "failed", "reason": "Could not find the application"}

## Guidelines

1. **Be precise with coordinates**: Click exactly where needed. The screenshot shows the current state.

2. **One action at a time**: Return exactly one action per response. Wait for the result before continuing.

3. **Verify your actions**: After each action, check the next screenshot to confirm it worked.

4. **Use keyboard shortcuts**: They're often faster than clicking through menus (e.g., Ctrl+S to save).

5. **Handle errors gracefully**: If something doesn't work, try an alternative approach before failing.

6. **Wait when needed**: Use wait action if you expect a dialog or window to appear.

7. **File paths must be absolute**: Always use full paths like C:\Users\... for file operations.

8. **Report completion**: When the goal is achieved, use the done action with a summary.

9. **Report failures**: If you cannot complete the task after reasonable attempts, use failed with a reason.

## Response Format

Respond with ONLY a JSON object. No markdown code blocks, no explanation, no extra text.

Correct: {"type": "click", "x": 100, "y": 200}
Wrong: ` + "```json\n{\"type\": \"click\"}\n```" + `
Wrong: I'll click here: {"type": "click"}
`

// BuildUserPrompt constructs the user prompt with goal and history.
func BuildUserPrompt(goal string, history []HistoryEntry) string {
	prompt := "## Goal\n" + goal + "\n\n"

	if len(history) > 0 {
		prompt += "## Action History\n"
		for i, entry := range history {
			prompt += formatHistoryEntry(i+1, entry)
		}
		prompt += "\n"
	}

	prompt += "## Current Screenshot\nAnalyze the screenshot below and decide the next action.\n"
	return prompt
}

func formatHistoryEntry(num int, entry HistoryEntry) string {
	result := ""
	switch entry.Action.Type {
	case "click":
		result = formatClick(entry)
	case "type":
		result = formatType(entry)
	case "key":
		result = formatKey(entry)
	case "scroll":
		result = formatScroll(entry)
	case "file_read":
		result = formatFileRead(entry)
	case "file_write":
		result = formatFileWrite(entry)
	case "wait":
		result = formatWait(entry)
	default:
		result = entry.Action.Type
	}

	status := "OK"
	if entry.Error != "" {
		status = "ERROR: " + entry.Error
	}
	return formatEntry(num, result, status)
}

func formatClick(entry HistoryEntry) string {
	btn := entry.Action.Button
	if btn == "" {
		btn = "left"
	}
	dbl := ""
	if entry.Action.Double {
		dbl = " double"
	}
	return formatAction("click", "%s%s at (%d, %d)", btn, dbl, entry.Action.X, entry.Action.Y)
}

func formatType(entry HistoryEntry) string {
	text := entry.Action.Text
	if len(text) > 50 {
		text = text[:50] + "..."
	}
	return formatAction("type", "%q", text)
}

func formatKey(entry HistoryEntry) string {
	return formatAction("key", "%s", entry.Action.Key)
}

func formatScroll(entry HistoryEntry) string {
	return formatAction("scroll", "%s %d", entry.Action.Direction, entry.Action.Amount)
}

func formatFileRead(entry HistoryEntry) string {
	return formatAction("file_read", "%s", entry.Action.Path)
}

func formatFileWrite(entry HistoryEntry) string {
	return formatAction("file_write", "%s", entry.Action.Path)
}

func formatWait(entry HistoryEntry) string {
	return formatAction("wait", "%dms", entry.Action.Ms)
}

func formatAction(action, format string, args ...interface{}) string {
	return action + ": " + sprintf(format, args...)
}

func formatEntry(num int, action, status string) string {
	return sprintf("%d. %s [%s]\n", num, action, status)
}

func sprintf(format string, args ...interface{}) string {
	// Simple sprintf implementation to avoid importing fmt in hot path
	result := format
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			result = replaceFirst(result, "%s", v)
			result = replaceFirst(result, "%q", "\""+v+"\"")
		case int:
			result = replaceFirst(result, "%d", itoa(v))
		}
	}
	return result
}

func replaceFirst(s, old, new string) string {
	for i := 0; i <= len(s)-len(old); i++ {
		if s[i:i+len(old)] == old {
			return s[:i] + new + s[i+len(old):]
		}
	}
	return s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

// HistoryEntry represents a single action and its result in the history.
type HistoryEntry struct {
	Action ActionRecord
	Error  string
}

// ActionRecord holds the action details for history.
type ActionRecord struct {
	Type      string
	X, Y      int
	Button    string
	Double    bool
	Text      string
	Key       string
	Direction string
	Amount    int
	Path      string
	Content   string
	Ms        int
	Summary   string
	Reason    string
}
