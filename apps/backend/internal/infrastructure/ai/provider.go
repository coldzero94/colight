package ai

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/coby/colight/apps/backend/internal/infrastructure/config"
)

// AIProvider routes AI calls to the correct provider based on model name.
// Each provider is initialized independently based on API key availability.
type AIProvider struct {
	gemini    LLMProvider      // Gemini (optional)
	groq      *GroqProvider    // Groq multi-model (optional)
	groqLLM   LLMProvider      // test override: routes "groq" calls to this instead of groq field
	groqSem   chan struct{}    // semaphore to serialize Groq calls (avoid rate limit)
	throttler *ModelThrottler  // RPM throttler for all models
}

// NewAIProvider creates providers for all available API keys.
func NewAIProvider(ctx context.Context, cfg *config.Config) (*AIProvider, error) {
	p := &AIProvider{}

	// Gemini (optional)
	if cfg.GeminiAPIKey != "" {
		g, err := NewGeminiProvider(ctx, cfg.GeminiAPIKey, "")
		if err != nil {
			slog.Warn("Gemini provider init failed, skipping", "error", err)
		} else {
			p.gemini = g
		}
	}

	// Groq (optional)
	if cfg.GroqAPIKey != "" {
		p.groq = NewGroqProvider(cfg.GroqAPIKey, "")
	}

	// At least one provider must be available
	if p.gemini == nil && p.groq == nil {
		return nil, fmt.Errorf("no AI provider available: set at least one of GEMINI_API_KEY or GROQ_API_KEY")
	}

	// Serialize Groq calls to avoid rate limit on free tier (1 concurrent call)
	p.groqSem = make(chan struct{}, 1)

	// Initialize RPM throttler for all models
	p.throttler = NewModelThrottler()

	return p, nil
}

// CallByModelName routes to the correct provider based on model name from prompt templates.
func (p *AIProvider) CallByModelName(ctx context.Context, modelName string, req LLMRequest) (LLMResponse, error) {
	// Enforce RPM throttling before calling (skip if throttler not initialized, e.g., in tests)
	if p.throttler != nil {
		if err := p.throttler.Wait(ctx, modelName); err != nil {
			return LLMResponse{}, fmt.Errorf("throttle wait failed: %w", err)
		}
	}
	switch modelName {
	case "gemini-3-pro", "gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.5-flash-lite",
		"gemini-2.0-flash", "gemini-2.0-flash-lite", "gemini-flash", "gemini":
		if p.gemini == nil {
			return LLMResponse{}, fmt.Errorf("Gemini not available (missing GEMINI_API_KEY)")
		}
		if gp, ok := p.gemini.(*GeminiProvider); ok {
			return gp.CallWithModel(ctx, modelName, req)
		}
		return p.gemini.Call(ctx, req)

	case "groq", "llama":
		if p.groqLLM != nil {
			return p.groqLLM.Call(ctx, req)
		}
		if p.groq == nil {
			return LLMResponse{}, fmt.Errorf("Groq not available (missing GROQ_API_KEY)")
		}
		return p.callGroqSerialized(ctx, func() (LLMResponse, error) {
			return p.groq.Call(ctx, req)
		})

	default:
		// Check if it's a known Groq model alias (e.g. "groq/compound", "llama-3.3-70b-versatile")
		if _, ok := GroqModelAliases[modelName]; ok {
			if p.groqLLM != nil {
				return p.groqLLM.Call(ctx, req)
			}
			if p.groq == nil {
				return LLMResponse{}, fmt.Errorf("Groq not available for model %s (missing GROQ_API_KEY)", modelName)
			}
			return p.callGroqSerialized(ctx, func() (LLMResponse, error) {
				return p.groq.CallWithModel(ctx, modelName, req)
			})
		}

		// Check if it's a known Gemini model alias (e.g. "gemini-lite")
		if _, ok := GeminiModelAliases[modelName]; ok {
			if p.gemini == nil {
				return LLMResponse{}, fmt.Errorf("Gemini not available for model %s (missing GEMINI_API_KEY)", modelName)
			}
			if gp, ok := p.gemini.(*GeminiProvider); ok {
				return gp.CallWithModel(ctx, modelName, req)
			}
			return p.gemini.Call(ctx, req)
		}

		return LLMResponse{}, fmt.Errorf("unknown model: %s", modelName)
	}
}

// callGroqSerialized acquires the Groq semaphore before calling, ensuring
// only one Groq API call runs at a time to avoid rate limit on free tier.
func (p *AIProvider) callGroqSerialized(ctx context.Context, fn func() (LLMResponse, error)) (LLMResponse, error) {
	if p.groqSem == nil {
		return fn()
	}
	select {
	case p.groqSem <- struct{}{}:
		defer func() { <-p.groqSem }()
		return fn()
	case <-ctx.Done():
		return LLMResponse{}, ctx.Err()
	}
}

// Groq returns the Groq provider. Returns nil if not configured.
func (p *AIProvider) Groq() *GroqProvider {
	return p.groq
}

// Throttler returns the model throttler for monitoring quota hits.
func (p *AIProvider) Throttler() *ModelThrottler {
	return p.throttler
}

// Close closes all AI clients.
func (p *AIProvider) Close() error {
	return nil
}
