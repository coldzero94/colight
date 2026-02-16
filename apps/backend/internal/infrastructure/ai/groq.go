package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Supported Groq model IDs
const (
	GroqModelLlama33_70B = "llama-3.3-70b-versatile"
	GroqModelLlama4Scout = "meta-llama/llama-4-scout-17b-16e-instruct"
	GroqModelQwen3_32B   = "qwen/qwen3-32b"
	GroqModelGPTOSS120B  = "openai/gpt-oss-120b"
	GroqModelKimiK2      = "moonshotai/kimi-k2-instruct"
	GroqModelCompound    = "groq/compound"
)

// GroqModelAliases maps short names to full Groq model IDs.
var GroqModelAliases = map[string]string{
	// Full IDs (identity)
	GroqModelLlama33_70B: GroqModelLlama33_70B,
	GroqModelLlama4Scout: GroqModelLlama4Scout,
	GroqModelQwen3_32B:   GroqModelQwen3_32B,
	GroqModelGPTOSS120B:  GroqModelGPTOSS120B,
	GroqModelKimiK2:      GroqModelKimiK2,
	GroqModelCompound:    GroqModelCompound,
	// Short aliases
	"llama-3.3":    GroqModelLlama33_70B,
	"llama-4-scout": GroqModelLlama4Scout,
	"qwen3":        GroqModelQwen3_32B,
	"gpt-oss-120b": GroqModelGPTOSS120B,
	"kimi-k2":      GroqModelKimiK2,
	"compound":     GroqModelCompound,
}

// ResolveGroqModel returns the full model ID for a given alias.
// Returns the input unchanged if no alias is found.
func ResolveGroqModel(name string) string {
	if resolved, ok := GroqModelAliases[name]; ok {
		return resolved
	}
	return name
}

// GroqProvider implements LLMProvider for Groq (OpenAI-compatible API)
type GroqProvider struct {
	apiKey string
	model  string
	client *http.Client
}

// NewGroqProvider creates a new Groq provider with the specified model.
// If model is empty, defaults to llama-3.3-70b-versatile.
func NewGroqProvider(apiKey, model string) *GroqProvider {
	if model == "" {
		model = GroqModelLlama33_70B
	}
	return &GroqProvider{
		apiKey: apiKey,
		model:  ResolveGroqModel(model),
		client: &http.Client{},
	}
}

// groqRequest is the OpenAI-compatible request format
type groqRequest struct {
	Model          string          `json:"model"`
	Messages       []groqMessage   `json:"messages"`
	Temperature    float64         `json:"temperature,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
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

// Call sends a request to Groq using the default model and returns the response
func (g *GroqProvider) Call(ctx context.Context, req LLMRequest) (LLMResponse, error) {
	return g.callInternal(ctx, g.model, req)
}

// CallWithModel sends a request to Groq with a specific model override.
// The model name can be a full ID or a short alias.
func (g *GroqProvider) CallWithModel(ctx context.Context, model string, req LLMRequest) (LLMResponse, error) {
	return g.callInternal(ctx, ResolveGroqModel(model), req)
}

func (g *GroqProvider) callInternal(ctx context.Context, model string, req LLMRequest) (LLMResponse, error) {
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
		Model:    model,
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
