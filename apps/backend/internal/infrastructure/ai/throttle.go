package ai

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ModelThrottler enforces RPM (Requests Per Minute) limits for each model.
// Uses token bucket algorithm from golang.org/x/time/rate.
type ModelThrottler struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
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

// getModelRPM returns the RPM limit for a model based on Google AI Studio dashboard.
func getModelRPM(modelName string) int {
	rpmLimits := map[string]int{
		// Gemini 3 series
		"gemini-3-pro":   15,
		"gemini-3-flash": 5,

		// Gemini 2.5 series
		"gemini-2.5-pro":        15,
		"gemini-2.5-flash":      5,
		"gemini-2.5-flash-lite": 10,

		// Gemini 2 series
		"gemini-2.0-flash":      15,
		"gemini-2.0-flash-lite": 15,

		// Groq models (no strict RPM, but add conservative limits)
		"meta-llama/llama-4-scout-17b-16e-instruct": 30,
		"llama-4-scout":                             30,
		"llama-3.3-70b-versatile":                   30,
		"llama-3.3-70b":                             30,
		"qwen/qwen3-32b":                            30,
		"qwen3-32b":                                 30,
	}

	if rpm, ok := rpmLimits[modelName]; ok {
		return rpm
	}

	return 0 // No limit for unknown models
}
