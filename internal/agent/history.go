package agent

import (
	"time"

	"github.com/thesimpledev/golemming/internal/llm"
)

// History tracks the action history for an agent.
type History struct {
	entries []Entry
}

// Entry represents a single history entry with timestamp.
type Entry struct {
	Timestamp time.Time
	LLMEntry  llm.HistoryEntry
}

// NewHistory creates a new history tracker.
func NewHistory() *History {
	return &History{
		entries: make([]Entry, 0),
	}
}

// Add adds an entry to the history.
func (h *History) Add(entry llm.HistoryEntry) {
	h.entries = append(h.entries, Entry{
		Timestamp: time.Now(),
		LLMEntry:  entry,
	})
}

// GetLLMHistory returns the history in LLM format.
func (h *History) GetLLMHistory() []llm.HistoryEntry {
	result := make([]llm.HistoryEntry, len(h.entries))
	for i, entry := range h.entries {
		result[i] = entry.LLMEntry
	}
	return result
}

// Len returns the number of entries.
func (h *History) Len() int {
	return len(h.entries)
}

// Last returns the last entry, or nil if empty.
func (h *History) Last() *Entry {
	if len(h.entries) == 0 {
		return nil
	}
	return &h.entries[len(h.entries)-1]
}

// All returns all entries.
func (h *History) All() []Entry {
	return h.entries
}
