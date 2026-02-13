package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GroqProvider implements LLMProvider for Groq (OpenAI-compatible API)
type GroqProvider struct {
	apiKey string
	model  string
	client *http.Client
}

// NewGroqProvider creates a new Groq provider
func NewGroqProvider(apiKey string) *GroqProvider {
	return &GroqProvider{
		apiKey: apiKey,
		model:  "llama-3.3-70b-versatile",
		client: &http.Client{},
	}
}

// groqRequest is the OpenAI-compatible request format
type groqRequest struct {
	Model       string          `json:"model"`
	Messages    []groqMessage   `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"` // "json_object"
}

// groqResponse is the OpenAI-compatible response format
type groqResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

// Call sends a request to Groq and returns the response
func (g *GroqProvider) Call(ctx context.Context, req LLMRequest) (LLMResponse, error) {
	// Build messages
	messages := []groqMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, groqMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}
	messages = append(messages, groqMessage{
		Role:    "user",
		Content: req.UserPrompt,
	})

	// Build request
	groqReq := groqRequest{
		Model:    g.model,
		Messages: messages,
	}
	if req.Temperature > 0 {
		groqReq.Temperature = req.Temperature
	}
	if req.MaxTokens > 0 {
		groqReq.MaxTokens = req.MaxTokens
	}
	if req.JSONMode {
		groqReq.ResponseFormat = &responseFormat{Type: "json_object"}
	}

	// Encode request
	body, err := json.Marshal(groqReq)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("failed to marshal Groq request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return LLMResponse{}, fmt.Errorf("failed to create Groq request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)

	// Send request
	httpResp, err := g.client.Do(httpReq)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("Groq API call failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Check status
	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return LLMResponse{}, fmt.Errorf("Groq API error (status %d): %s", httpResp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var groqResp groqResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&groqResp); err != nil {
		return LLMResponse{}, fmt.Errorf("failed to decode Groq response: %w", err)
	}

	// Extract response text
	if len(groqResp.Choices) == 0 {
		return LLMResponse{}, fmt.Errorf("empty response from Groq")
	}

	return LLMResponse{
		Content:      groqResp.Choices[0].Message.Content,
		InputTokens:  groqResp.Usage.PromptTokens,
		OutputTokens: groqResp.Usage.CompletionTokens,
		Model:        groqResp.Model,
	}, nil
}
