package ai

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// Supported Gemini model IDs
const (
	GeminiModel25Flash     = "gemini-2.5-flash"
	GeminiModel25FlashLite = "gemini-2.5-flash-lite"
	GeminiModel25Pro       = "gemini-2.5-pro"
	GeminiModel20Flash     = "gemini-2.0-flash" // Deprecated: shutting down March 31, 2026
)

// GeminiModelAliases maps short names to full Gemini model IDs.
var GeminiModelAliases = map[string]string{
	// Full IDs (identity)
	GeminiModel25Flash:     GeminiModel25Flash,
	GeminiModel25FlashLite: GeminiModel25FlashLite,
	GeminiModel25Pro:       GeminiModel25Pro,
	GeminiModel20Flash:     GeminiModel20Flash,
	// Short aliases
	"gemini-flash": GeminiModel25Flash,
	"gemini-2.5":   GeminiModel25Flash,
	"gemini":       GeminiModel25Flash,
	"gemini-lite":  GeminiModel25FlashLite,
	"gemini-pro":   GeminiModel25Pro,
	"gemini-2.0":   GeminiModel20Flash,
}

// ResolveGeminiModel returns the full model ID for a given alias.
// Returns the input unchanged if no alias is found.
func ResolveGeminiModel(name string) string {
	if resolved, ok := GeminiModelAliases[name]; ok {
		return resolved
	}
	return name
}

// GeminiProvider implements LLMProvider for Google Gemini (official SDK)
type GeminiProvider struct {
	client *genai.Client
	model  string
}

// NewGeminiProvider creates a new Gemini provider using official google.golang.org/genai SDK.
// If model is empty, defaults to gemini-2.5-flash.
func NewGeminiProvider(ctx context.Context, apiKey, model string) (*GeminiProvider, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	if model == "" {
		model = GeminiModel25Flash
	}

	return &GeminiProvider{
		client: client,
		model:  ResolveGeminiModel(model),
	}, nil
}

// Call sends a request to Gemini using the default model and returns the response
func (g *GeminiProvider) Call(ctx context.Context, req LLMRequest) (LLMResponse, error) {
	return g.callInternal(ctx, g.model, req)
}

// CallWithModel sends a request to Gemini with a specific model override.
// The model name can be a full ID or a short alias.
func (g *GeminiProvider) CallWithModel(ctx context.Context, model string, req LLMRequest) (LLMResponse, error) {
	return g.callInternal(ctx, ResolveGeminiModel(model), req)
}

func (g *GeminiProvider) callInternal(ctx context.Context, model string, req LLMRequest) (LLMResponse, error) {
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
	resp, err := g.client.Models.GenerateContent(ctx, model, contents, opts)
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
		Model:        model,
	}, nil
}
