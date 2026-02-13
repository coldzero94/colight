package ai

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiProvider implements LLMProvider for Google Gemini
type GeminiProvider struct {
	client *genai.Client
	model  string
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(ctx context.Context, apiKey string) (*GeminiProvider, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiProvider{
		client: client,
		model:  "gemini-2.0-flash-exp",
	}, nil
}

// Call sends a request to Gemini and returns the response
func (g *GeminiProvider) Call(ctx context.Context, req LLMRequest) (LLMResponse, error) {
	// Build messages
	parts := []genai.Part{}
	if req.SystemPrompt != "" {
		parts = append(parts, genai.Text(req.SystemPrompt+"\n\n"))
	}
	parts = append(parts, genai.Text(req.UserPrompt))

	// Configure model
	model := g.client.GenerativeModel(g.model)
	if req.Temperature > 0 {
		temp := float32(req.Temperature)
		model.Temperature = &temp
	}
	if req.MaxTokens > 0 {
		maxTokens := int32(req.MaxTokens)
		model.MaxOutputTokens = &maxTokens
	}
	if req.JSONMode {
		model.ResponseMIMEType = "application/json"
	}

	// Generate content
	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("Gemini API call failed: %w", err)
	}

	// Extract response text
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return LLMResponse{}, fmt.Errorf("empty response from Gemini")
	}

	content := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			content += string(txt)
		}
	}

	// Extract token usage
	inputTokens := 0
	outputTokens := 0
	if resp.UsageMetadata != nil {
		inputTokens = int(resp.UsageMetadata.PromptTokenCount)
		outputTokens = int(resp.UsageMetadata.CandidatesTokenCount)
	}

	return LLMResponse{
		Content:      content,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		Model:        g.model,
	}, nil
}

// Close closes the Gemini client
func (g *GeminiProvider) Close() error {
	return g.client.Close()
}
