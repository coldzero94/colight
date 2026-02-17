package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// LLMProvider is the common interface for lightweight AI models (Gemini, Groq)
type LLMProvider interface {
	Call(ctx context.Context, req LLMRequest) (LLMResponse, error)
}

// ExtractJSON extracts a JSON object from an AI response that may contain
// surrounding text, markdown code fences, or double-brace wrapping.
// Tries multiple strategies in order: direct parse, code fence extraction,
// last code fence block, first-last brace extraction.
func ExtractJSON(content string, target any) error {
	// 1. Try direct parse
	if err := json.Unmarshal([]byte(content), target); err == nil {
		return nil
	}

	// 2. Find the LAST ```json ... ``` block (models often put the real JSON last)
	if jsonStr := extractLastCodeFence(content); jsonStr != "" {
		if err := json.Unmarshal([]byte(jsonStr), target); err == nil {
			return nil
		}
	}

	// 3. Find first '{' and last '}' and try to parse
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start != -1 && end > start {
		jsonStr := content[start : end+1]
		if err := json.Unmarshal([]byte(jsonStr), target); err == nil {
			return nil
		}
	}

	return fmt.Errorf("no valid JSON found in response (len=%d): %.200s", len(content), content)
}

// extractLastCodeFence finds the last ```...``` block and returns its content.
func extractLastCodeFence(content string) string {
	// Find the last opening fence
	lastOpen := strings.LastIndex(content, "```json")
	if lastOpen == -1 {
		lastOpen = strings.LastIndex(content, "```")
	}
	if lastOpen == -1 {
		return ""
	}

	// Move past the opening fence line
	afterFence := content[lastOpen:]
	newline := strings.Index(afterFence, "\n")
	if newline == -1 {
		return ""
	}
	inner := afterFence[newline+1:]

	// Find closing fence
	closeFence := strings.Index(inner, "```")
	if closeFence == -1 {
		return ""
	}

	return strings.TrimSpace(inner[:closeFence])
}

// LLMRequest represents a request to the lightweight LLM
type LLMRequest struct {
	SystemPrompt string
	UserPrompt   string
	Temperature  float64
	MaxTokens    int
	JSONMode     bool // structured output (JSON schema)
}

// LLMResponse represents a response from the lightweight LLM
type LLMResponse struct {
	Content      string
	InputTokens  int
	OutputTokens int
	Model        string
}

// StreamCallback is called for each text chunk during streaming.
type StreamCallback func(chunk string)

// StreamingLLMProvider extends LLMProvider with streaming capability.
// Only implemented by providers that support streaming (e.g. Claude).
type StreamingLLMProvider interface {
	LLMProvider
	// Stream sends a request and calls onChunk for each text delta.
	// Returns the final accumulated LLMResponse (including token usage) after completion.
	Stream(ctx context.Context, req LLMRequest, onChunk StreamCallback) (LLMResponse, error)
}
