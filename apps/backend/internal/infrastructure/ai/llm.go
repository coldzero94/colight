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
// surrounding text (e.g. "I'll analyze... {json} ..."). Finds the first '{' and
// last '}' and attempts to parse the substring as JSON.
func ExtractJSON(content string, target any) error {
	// Try direct parse first
	if err := json.Unmarshal([]byte(content), target); err == nil {
		return nil
	}

	// Find first '{' and last '}'
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 || end <= start {
		return fmt.Errorf("no JSON object found in response: %.100s", content)
	}

	jsonStr := content[start : end+1]
	if err := json.Unmarshal([]byte(jsonStr), target); err != nil {
		return fmt.Errorf("failed to parse extracted JSON: %w", err)
	}
	return nil
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
