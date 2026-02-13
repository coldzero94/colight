package ai

import (
	"context"
	"fmt"

	"github.com/coby/colight/apps/backend/internal/infrastructure/config"
)

// AIProvider provides access to all AI models (light + heavy + embedding)
type AIProvider struct {
	light LLMProvider // Gemini Flash or Groq (경량 작업)
	heavy LLMProvider // Claude Sonnet 4.5 (심층 분석)
	// embedding will be added later
}

// NewAIProvider creates a new AI provider based on configuration
func NewAIProvider(ctx context.Context, cfg *config.Config) (*AIProvider, error) {
	// 1. Select light provider based on LLM_LIGHT_PROVIDER env var
	var light LLMProvider
	var err error

	switch cfg.LLMLightProvider {
	case "groq":
		if cfg.GroqAPIKey == "" {
			return nil, fmt.Errorf("GROQ_API_KEY is required when LLM_LIGHT_PROVIDER=groq")
		}
		light = NewGroqProvider(cfg.GroqAPIKey)
	default: // "gemini"
		if cfg.GeminiAPIKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY is required when LLM_LIGHT_PROVIDER=gemini")
		}
		light, err = NewGeminiProvider(ctx, cfg.GeminiAPIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini provider: %w", err)
		}
	}

	// 2. Initialize heavy provider (Claude) if API key is available
	var heavy LLMProvider
	if cfg.AnthropicAPIKey != "" {
		heavy = NewClaudeProvider(cfg.AnthropicAPIKey)
	}

	return &AIProvider{
		light: light,
		heavy: heavy,
	}, nil
}

// CallLight calls the lightweight LLM (Gemini or Groq)
func (p *AIProvider) CallLight(ctx context.Context, req LLMRequest) (LLMResponse, error) {
	return p.light.Call(ctx, req)
}

// Light returns the lightweight LLM provider for direct use by services
func (p *AIProvider) Light() LLMProvider {
	return p.light
}

// Heavy returns the heavy LLM provider (Claude) for direct use
func (p *AIProvider) Heavy() LLMProvider {
	return p.heavy
}

// CallByModelName calls the appropriate provider based on model name
// This allows prompt templates to specify which model to use
func (p *AIProvider) CallByModelName(ctx context.Context, modelName string, req LLMRequest) (LLMResponse, error) {
	switch {
	case modelName == "claude-sonnet-4-5" || modelName == "claude":
		if p.heavy == nil {
			return LLMResponse{}, fmt.Errorf("Claude provider not initialized (missing ANTHROPIC_API_KEY)")
		}
		return p.heavy.Call(ctx, req)

	case modelName == "gemini-2.0-flash" || modelName == "gemini-flash" || modelName == "gemini":
		return p.light.Call(ctx, req)

	case modelName == "groq" || modelName == "llama":
		return p.light.Call(ctx, req)

	default:
		// Default to light model for unknown models
		return p.light.Call(ctx, req)
	}
}

// CallHeavy calls the heavy model (Claude Sonnet 4.5)
func (p *AIProvider) CallHeavy(ctx context.Context, req LLMRequest) (LLMResponse, error) {
	if p.heavy == nil {
		return LLMResponse{}, fmt.Errorf("heavy model not initialized")
	}
	return p.heavy.Call(ctx, req)
}

// Close closes all AI clients
func (p *AIProvider) Close() error {
	// New Gemini SDK (google.golang.org/genai) doesn't require explicit Close()
	return nil
}
