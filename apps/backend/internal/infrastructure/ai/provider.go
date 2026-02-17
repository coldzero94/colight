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
	gemini  LLMProvider   // Gemini Flash (optional)
	claude  LLMProvider   // Claude Sonnet 4.5 (optional)
	groq    *GroqProvider // Groq multi-model (optional)
	groqLLM LLMProvider   // test override: routes "groq" calls to this instead of groq field
	groqSem chan struct{} // semaphore to serialize Groq calls (avoid rate limit)
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

	// Claude (optional)
	if cfg.AnthropicAPIKey != "" {
		p.claude = NewClaudeProvider(cfg.AnthropicAPIKey)
	}

	// At least one provider must be available
	if p.gemini == nil && p.groq == nil && p.claude == nil {
		return nil, fmt.Errorf("no AI provider available: set at least one of GEMINI_API_KEY, GROQ_API_KEY, or ANTHROPIC_API_KEY")
	}

	// Serialize Groq calls to avoid rate limit on free tier (1 concurrent call)
	p.groqSem = make(chan struct{}, 1)

	return p, nil
}

// CallByModelName routes to the correct provider based on model name from prompt templates.
func (p *AIProvider) CallByModelName(ctx context.Context, modelName string, req LLMRequest) (LLMResponse, error) {
	switch modelName {
	case "claude-sonnet-4-5", "claude-sonnet-4.5", "claude":
		if p.claude == nil {
			return LLMResponse{}, fmt.Errorf("Claude not available (missing ANTHROPIC_API_KEY)")
		}
		return p.claude.Call(ctx, req)

	case "gemini-2.5-flash", "gemini-2.0-flash", "gemini-flash", "gemini":
		if p.gemini == nil {
			return LLMResponse{}, fmt.Errorf("Gemini not available (missing GEMINI_API_KEY)")
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

		// Check if it's a known Gemini model alias (e.g. "gemini-2.5-pro", "gemini-lite")
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

// Claude returns the Claude provider. Returns nil if not configured.
func (p *AIProvider) Claude() LLMProvider {
	return p.claude
}

// Groq returns the Groq provider. Returns nil if not configured.
func (p *AIProvider) Groq() *GroqProvider {
	return p.groq
}

// ClaudeStreaming returns the Claude provider as StreamingLLMProvider.
// Returns nil if Claude is not configured or doesn't support streaming.
func (p *AIProvider) ClaudeStreaming() StreamingLLMProvider {
	if sp, ok := p.claude.(StreamingLLMProvider); ok {
		return sp
	}
	return nil
}

// Close closes all AI clients.
func (p *AIProvider) Close() error {
	return nil
}
