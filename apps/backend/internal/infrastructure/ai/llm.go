package ai

import "context"

// LLMProvider is the common interface for lightweight AI models (Gemini, Groq)
type LLMProvider interface {
	Call(ctx context.Context, req LLMRequest) (LLMResponse, error)
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
