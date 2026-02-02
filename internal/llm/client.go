package llm

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/thesimpledev/golemming/pkg/protocol"
)

// Client wraps the Anthropic API client.
type Client struct {
	client *anthropic.Client
	model  string
}

// NewClient creates a new LLM client.
func NewClient(apiKey, model string) *Client {
	client := anthropic.NewClient()
	return &Client{
		client: &client,
		model:  model,
	}
}

// GetAction sends a screenshot and context to the LLM and returns the next action.
func (c *Client) GetAction(ctx context.Context, goal string, screenshotBase64 string, history []HistoryEntry) (*protocol.Action, error) {
	userPrompt := BuildUserPrompt(goal, history)

	// Build the message with image
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: SystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock(userPrompt),
				anthropic.NewImageBlockBase64("image/jpeg", screenshotBase64),
			),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call Anthropic API: %w", err)
	}

	// Extract text response
	if len(message.Content) == 0 {
		return nil, fmt.Errorf("empty response from LLM")
	}

	var responseText string
	for _, block := range message.Content {
		if block.Type == "text" {
			responseText = block.Text
			break
		}
	}

	if responseText == "" {
		return nil, fmt.Errorf("no text response from LLM")
	}

	// Parse the action
	return ParseAction(responseText)
}
