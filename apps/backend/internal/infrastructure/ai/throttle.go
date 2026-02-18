package ai

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// QuotaHitEvent records a rate limit hit for a model.
type QuotaHitEvent struct {
	Model     string    `json:"model"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// ModelThrottler enforces RPM (Requests Per Minute) limits for each model.
// Uses token bucket algorithm from golang.org/x/time/rate.
type ModelThrottler struct {
	limiters  map[string]*rate.Limiter
	mu        sync.RWMutex
	quotaHits []QuotaHitEvent // Recent quota hit events (ring buffer)
	hitMu     sync.Mutex
}

// NewModelThrottler creates a throttler with per-model RPM limits.
func NewModelThrottler() *ModelThrottler {
	return &ModelThrottler{
		limiters: make(map[string]*rate.Limiter),
	}
}

// Wait blocks until the model's rate limit allows the request.
// Returns immediately if model has no configured limit.
func (t *ModelThrottler) Wait(ctx context.Context, modelName string) error {
	limiter := t.getLimiter(modelName)
	if limiter == nil {
		return nil // No limit for this model
	}
	return limiter.Wait(ctx)
}

// getLimiter returns the rate limiter for a model, creating it if needed.
func (t *ModelThrottler) getLimiter(modelName string) *rate.Limiter {
	t.mu.RLock()
	limiter, exists := t.limiters[modelName]
	t.mu.RUnlock()

	if exists {
		return limiter
	}

	// Create limiter on first use
	t.mu.Lock()
	defer t.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := t.limiters[modelName]; exists {
		return limiter
	}

	// Get RPM limit for this model
	rpm := getModelRPM(modelName)
	if rpm == 0 {
		return nil // No limit
	}

	// Create token bucket: RPM requests per minute
	// rate.Every converts to interval between tokens
	limiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(rpm)), 1)
	t.limiters[modelName] = limiter

	return limiter
}

// ModelLimits contains rate limit information for a model.
type ModelLimits struct {
	RPM      int    // Requests per minute
	TPM      int    // Tokens per minute (in thousands, 0 = unlimited)
	RPD      int    // Requests per day (0 = unlimited)
	Context  int    // Context window size (in K tokens)
	Provider string // Provider name: "gemini", "groq"
	Note     string // Additional notes (e.g., "deprecated")
}

// allModelLimits is the single source of truth for all supported model limits.
// Keys are canonical model IDs (full IDs for Groq, short IDs for Gemini).
// Source: Google AI Studio dashboard & https://console.groq.com/docs/rate-limits
var allModelLimits = map[string]ModelLimits{
	// === Gemini models (Google AI Studio free tier) ===
	// High quota (1,500 RPD, Unlimited TPM)
	"gemini-3-pro":          {RPM: 15, TPM: 0, RPD: 1500, Context: 1000, Provider: "gemini", Note: "Latest, Unlimited TPM"},
	"gemini-2.5-pro":        {RPM: 15, TPM: 0, RPD: 1500, Context: 1000, Provider: "gemini", Note: "Unlimited TPM"},
	"gemini-2.0-flash":      {RPM: 15, TPM: 0, RPD: 1500, Context: 1000, Provider: "gemini", Note: "Deprecated Mar 31, 2026"},
	"gemini-2.0-flash-lite": {RPM: 15, TPM: 0, RPD: 1500, Context: 1000, Provider: "gemini", Note: "Deprecated Mar 31, 2026"},
	// Low quota (20 RPD) — avoid for production traffic
	"gemini-2.5-flash":      {RPM: 5, TPM: 250, RPD: 20, Context: 1000, Provider: "gemini", Note: "Low quota (20 RPD)"},
	"gemini-2.5-flash-lite": {RPM: 10, TPM: 250, RPD: 20, Context: 1000, Provider: "gemini", Note: "Low quota (20 RPD)"},

	// === Groq models (Free tier, source: console.groq.com/docs/rate-limits) ===
	// Production models
	"llama-3.3-70b-versatile":                   {RPM: 30, TPM: 12, RPD: 1000, Context: 131, Provider: "groq", Note: "Production, stable"},
	"llama-3.1-8b-instant":                      {RPM: 30, TPM: 6, RPD: 14400, Context: 131, Provider: "groq", Note: "Production, highest RPD (14,400)"},
	"openai/gpt-oss-120b":                       {RPM: 30, TPM: 8, RPD: 1000, Context: 131, Provider: "groq", Note: "Production, reasoning+search"},
	"openai/gpt-oss-20b":                        {RPM: 30, TPM: 8, RPD: 1000, Context: 131, Provider: "groq", Note: "Production, fastest GPT"},
	"groq/compound":                             {RPM: 30, TPM: 70, RPD: 250, Context: 131, Provider: "groq", Note: "Production, auto-routing, low RPD (250)"},
	"groq/compound-mini":                        {RPM: 30, TPM: 70, RPD: 250, Context: 131, Provider: "groq", Note: "Production, agentic AI, low RPD (250)"},
	// Preview models
	"meta-llama/llama-4-scout-17b-16e-instruct":    {RPM: 30, TPM: 30, RPD: 1000, Context: 131, Provider: "groq", Note: "Preview, highest TPM (30K)"},
	"meta-llama/llama-4-maverick-17b-128e-instruct": {RPM: 30, TPM: 6, RPD: 1000, Context: 131, Provider: "groq", Note: "Preview, specialized reasoning"},
	"moonshotai/kimi-k2-instruct":                  {RPM: 60, TPM: 10, RPD: 1000, Context: 131, Provider: "groq", Note: "Preview, highest RPM (60)"},
	"qwen/qwen3-32b":                               {RPM: 60, TPM: 6, RPD: 1000, Context: 131, Provider: "groq", Note: "Preview, highest RPM (60)"},

}

// shortAliasToCanonical maps short aliases to canonical model IDs for limit lookup.
var shortAliasToCanonical = map[string]string{
	// Groq short aliases
	"llama-4-scout":    "meta-llama/llama-4-scout-17b-16e-instruct",
	"llama-4-maverick": "meta-llama/llama-4-maverick-17b-128e-instruct",
	"llama-3.3-70b":    "llama-3.3-70b-versatile",
	"llama-3.3":        "llama-3.3-70b-versatile",
	"llama-3.1":        "llama-3.1-8b-instant",
	"qwen3-32b":        "qwen/qwen3-32b",
	"qwen3":            "qwen/qwen3-32b",
	"gpt-oss-120b":     "openai/gpt-oss-120b",
	"gpt-oss-20b":      "openai/gpt-oss-20b",
	"kimi-k2":          "moonshotai/kimi-k2-instruct",
	"compound":         "groq/compound",
	"compound-mini":    "groq/compound-mini",
}

// GetModelLimits returns comprehensive rate limits for a model.
// Handles both canonical IDs and short aliases.
func GetModelLimits(modelName string) ModelLimits {
	if limit, ok := allModelLimits[modelName]; ok {
		return limit
	}
	if canonical, ok := shortAliasToCanonical[modelName]; ok {
		if limit, ok := allModelLimits[canonical]; ok {
			return limit
		}
	}
	return ModelLimits{RPM: 0, TPM: 0, RPD: 0, Context: 0, Provider: "unknown", Note: "Unknown model"}
}

// SupportedModelInfo contains model info for admin API responses.
type SupportedModelInfo struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	RPM      int    `json:"rpm"`
	TPM      int    `json:"tpm"`
	RPD      int    `json:"rpd"`
	Context  int    `json:"context_size"`
	Note     string `json:"note"`
}

// GetSelectableModels returns all models that admins can assign to prompt templates.
func GetSelectableModels() []SupportedModelInfo {
	var models []SupportedModelInfo
	for id, limit := range allModelLimits {
		models = append(models, SupportedModelInfo{
			ID:       id,
			Provider: limit.Provider,
			RPM:      limit.RPM,
			TPM:      limit.TPM,
			RPD:      limit.RPD,
			Context:  limit.Context,
			Note:     limit.Note,
		})
	}
	return models
}

// IsSelectableModel checks if a model can be assigned to prompt templates.
func IsSelectableModel(modelName string) bool {
	if _, ok := allModelLimits[modelName]; ok {
		return true
	}
	if canonical, ok := shortAliasToCanonical[modelName]; ok {
		if _, ok := allModelLimits[canonical]; ok {
			return true
		}
	}
	return false
}

// RecordQuotaHit records a rate limit hit event for monitoring.
func (t *ModelThrottler) RecordQuotaHit(model string, err error) {
	t.hitMu.Lock()
	defer t.hitMu.Unlock()

	event := QuotaHitEvent{
		Model:     model,
		Error:     err.Error(),
		Timestamp: time.Now(),
	}
	t.quotaHits = append(t.quotaHits, event)

	// Keep only last 200 events
	if len(t.quotaHits) > 200 {
		t.quotaHits = t.quotaHits[len(t.quotaHits)-200:]
	}
}

// QuotaHitSummary contains aggregated quota hit info for a model.
type QuotaHitSummary struct {
	Model     string    `json:"model"`
	Provider  string    `json:"provider"`
	HitCount  int       `json:"hit_count"`
	LastHitAt time.Time `json:"last_hit_at"`
	LastError string    `json:"last_error"`
}

// GetQuotaHitSummary returns aggregated quota hit stats per model within the given duration.
func (t *ModelThrottler) GetQuotaHitSummary(since time.Duration) []QuotaHitSummary {
	t.hitMu.Lock()
	defer t.hitMu.Unlock()

	cutoff := time.Now().Add(-since)
	modelHits := make(map[string]*QuotaHitSummary)

	for _, e := range t.quotaHits {
		if e.Timestamp.Before(cutoff) {
			continue
		}
		if _, ok := modelHits[e.Model]; !ok {
			limits := GetModelLimits(e.Model)
			modelHits[e.Model] = &QuotaHitSummary{
				Model:    e.Model,
				Provider: limits.Provider,
			}
		}
		s := modelHits[e.Model]
		s.HitCount++
		if e.Timestamp.After(s.LastHitAt) {
			s.LastHitAt = e.Timestamp
			s.LastError = e.Error
		}
	}

	result := make([]QuotaHitSummary, 0, len(modelHits))
	for _, s := range modelHits {
		result = append(result, *s)
	}
	return result
}

// getModelRPM returns the RPM limit for a model.
func getModelRPM(modelName string) int {
	return GetModelLimits(modelName).RPM
}
