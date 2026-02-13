package ai

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// GeminiProvider implements LLMProvider for Google Gemini (official SDK)
type GeminiProvider struct {
	client *genai.Client
	model  string
}

// NewGeminiProvider creates a new Gemini provider using official google.golang.org/genai SDK
func NewGeminiProvider(ctx context.Context, apiKey string) (*GeminiProvider, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
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
	// Build content array
	var contents []*genai.Content

	// Combine system and user prompts (new SDK uses Text function)
	promptText := req.UserPrompt
	if req.SystemPrompt != "" {
		promptText = req.SystemPrompt + "\n\n" + req.UserPrompt
	}

	contents = genai.Text(promptText)

	// Build generate options
	opts := &genai.GenerateContentConfig{}
	if req.Temperature > 0 {
		temp := float32(req.Temperature)
		opts.Temperature = &temp
	}
	if req.MaxTokens > 0 {
		opts.MaxOutputTokens = int32(req.MaxTokens)
	}
	if req.JSONMode {
		opts.ResponseMIMEType = "application/json"
	}

	// Generate content using new SDK
	resp, err := g.client.Models.GenerateContent(ctx, g.model, contents, opts)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("Gemini API call failed: %w", err)
	}

	// Extract response text
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return LLMResponse{}, fmt.Errorf("empty response from Gemini")
	}

	content := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			content += part.Text
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
