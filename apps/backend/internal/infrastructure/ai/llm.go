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
