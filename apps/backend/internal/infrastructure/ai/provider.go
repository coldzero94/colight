package ai

import (
	"context"
	"fmt"

	"github.com/coby/colight/apps/backend/internal/infrastructure/config"
)

// AIProvider provides access to all AI models (light + heavy)
type AIProvider struct {
	light LLMProvider
	// heavy (Claude) will be added in Phase 3.2
	// embedding will be added in Phase 2.1
}

// NewAIProvider creates a new AI provider based on configuration
func NewAIProvider(ctx context.Context, cfg *config.Config) (*AIProvider, error) {
	// Select light provider based on LLM_LIGHT_PROVIDER env var
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

	return &AIProvider{
		light: light,
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

// Close closes all AI clients
func (p *AIProvider) Close() error {
	// Close Gemini client if it's a GeminiProvider
	if gp, ok := p.light.(*GeminiProvider); ok {
		return gp.Close()
	}
	return nil
}
