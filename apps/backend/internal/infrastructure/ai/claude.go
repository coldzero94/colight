package ai

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClaudeProvider implements heavy model interface for Claude (Anthropic)
type ClaudeProvider struct {
	client anthropic.Client
	model  anthropic.Model
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(apiKey string) *ClaudeProvider {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &ClaudeProvider{
		client: client,
		model:  anthropic.ModelClaudeSonnet4_5_20250929,
	}
}

// Call sends a request to Claude and returns the response
func (c *ClaudeProvider) Call(ctx context.Context, req LLMRequest) (LLMResponse, error) {
	// Build messages
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(req.UserPrompt)),
	}

	// Build request parameters
	params := anthropic.MessageNewParams{
		Model:     c.model,
		Messages:  messages,
		MaxTokens: int64(req.MaxTokens),
	}

	// Add system prompt if provided
	if req.SystemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: req.SystemPrompt, Type: "text"},
		}
	}

	// Add temperature if provided (skip for now, use default)
	// Temperature is optional and has complex type

	// Call API
	response, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("Claude API call failed: %w", err)
	}

	// Extract text from response
	content := ""
	for _, block := range response.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	// Extract token usage
	inputTokens := 0
	outputTokens := 0
	if response.Usage.InputTokens > 0 {
		inputTokens = int(response.Usage.InputTokens)
	}
	if response.Usage.OutputTokens > 0 {
		outputTokens = int(response.Usage.OutputTokens)
	}

	return LLMResponse{
		Content:      content,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		Model:        string(c.model),
	}, nil
}
