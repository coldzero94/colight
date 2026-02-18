package ai

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"strings"
	"time"
)

// RetryConfig configures retry behavior for AI API calls
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  time.Second,
	}
}

// CallWithRetry calls an AI provider with exponential backoff retry logic
func CallWithRetry(ctx context.Context, provider LLMProvider, req LLMRequest, config RetryConfig) (LLMResponse, error) {
	var lastErr error
	start := time.Now()

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * config.BaseDelay
			jitter := time.Duration(rand.Int63n(int64(delay / 2)))
			sleepDuration := delay + jitter

			slog.Warn("ai_retry", "attempt", attempt, "max_retries", config.MaxRetries, "delay", sleepDuration.String(), "error", lastErr.Error())

			select {
			case <-time.After(sleepDuration):
			case <-ctx.Done():
				return LLMResponse{}, ctx.Err()
			}
		}

		resp, err := provider.Call(ctx, req)
		if err == nil {
			slog.Info("ai_call", "model", resp.Model, "duration", time.Since(start).String(), "input_tokens", resp.InputTokens, "output_tokens", resp.OutputTokens)
			return resp, nil
		}

		lastErr = err

		// Don't retry on certain errors
		if !isRetryableError(err) {
			break
		}
	}

	slog.Error("ai_call_failed", "attempts", config.MaxRetries+1, "duration", time.Since(start).String(), "error", lastErr.Error())
	return LLMResponse{}, fmt.Errorf("AI call failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}

// CallByModelNameWithRetry calls AIProvider.CallByModelName with exponential backoff retry logic
func CallByModelNameWithRetry(ctx context.Context, provider *AIProvider, modelName string, req LLMRequest, config RetryConfig) (LLMResponse, error) {
	var lastErr error
	start := time.Now()

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * config.BaseDelay
			jitter := time.Duration(rand.Int63n(int64(delay / 2)))
			sleepDuration := delay + jitter

			slog.Warn("ai_retry", "model", modelName, "attempt", attempt, "max_retries", config.MaxRetries, "delay", sleepDuration.String(), "error", lastErr.Error())

			select {
			case <-time.After(sleepDuration):
			case <-ctx.Done():
				return LLMResponse{}, ctx.Err()
			}
		}

		resp, err := provider.CallByModelName(ctx, modelName, req)
		if err == nil {
			slog.Info("ai_call", "model", resp.Model, "duration", time.Since(start).String(), "input_tokens", resp.InputTokens, "output_tokens", resp.OutputTokens)
			return resp, nil
		}

		lastErr = err

		// Don't retry on certain errors
		if !isRetryableError(err) {
			break
		}
	}

	slog.Error("ai_call_failed", "model", modelName, "attempts", config.MaxRetries+1, "duration", time.Since(start).String(), "error", lastErr.Error())
	return LLMResponse{}, fmt.Errorf("AI call failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	errStr := strings.ToLower(err.Error())

	// Retry on rate limits, timeouts, and temporary errors
	retryableKeywords := []string{
		"rate limit",
		"429",
		"timeout",
		"temporary",
		"unavailable",
		"503",
		"502",
	}

	for _, keyword := range retryableKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}

	return false
}

// isQuotaError checks if error is quota/rate limit related
func isQuotaError(err error) bool {
	errStr := strings.ToLower(err.Error())
	quotaKeywords := []string{
		"429",
		"rate limit",
		"quota exceeded",
		"quota",
		"resource exhausted",
		"too many requests",
	}
	for _, keyword := range quotaKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}
	return false
}

// FallbackConfig defines model fallback chain with priority order
type FallbackConfig struct {
	Models     []string      // Priority: ["gemini-2.5-pro", ...]
	MaxRetries int           // Retries per model before fallback (default: 1)
	BaseDelay  time.Duration // Delay between retries (default: 1s)
}

// DefaultFallbackConfig returns the default model fallback chain.
// Tier 1: Gemini models (high quality, 1,500 RPD each)
// Tier 2: Groq models (very high limit, cheap)
func DefaultFallbackConfig() FallbackConfig {
	return FallbackConfig{
		Models: []string{
			"gemini-2.5-pro",        // Tier 1: Best quality (1,500 RPD)
			"gemini-2.0-flash",      // Tier 1 Spare 1 (1,500 RPD, deprecated Mar 31)
			"gemini-2.0-flash-lite", // Tier 1 Spare 2 (1,500 RPD, deprecated Mar 31)
			"llama-4-scout",         // Tier 2: Groq primary (128K context)
			"llama-3.3-70b",         // Tier 2 Spare 1 (8K context, stable)
			"qwen3-32b",             // Tier 2 Spare 2 (32K context, alternative)
		},
		MaxRetries: 1,
		BaseDelay:  time.Second,
	}
}

// CallWithModelFallback attempts AI call with fallback chain.
// On quota errors (429, "quota exceeded"), tries next model.
// Returns: (response, modelUsed, error)
func CallWithModelFallback(
	ctx context.Context,
	provider *AIProvider,
	req LLMRequest,
	config FallbackConfig,
) (LLMResponse, string, error) {
	if len(config.Models) == 0 {
		return LLMResponse{}, "", fmt.Errorf("no models in fallback chain")
	}

	start := time.Now()
	var lastErr error
	attemptedModels := []string{}

	for modelIdx, modelName := range config.Models {
		attemptedModels = append(attemptedModels, modelName)
		slog.Info("ai_fallback_attempt", "model", modelName, "position", modelIdx+1, "total", len(config.Models))

		// Try this model with retries
		for attempt := 0; attempt <= config.MaxRetries; attempt++ {
			if attempt > 0 {
				delay := time.Duration(math.Pow(2, float64(attempt-1))) * config.BaseDelay
				jitter := time.Duration(rand.Int63n(int64(delay / 2)))
				sleepDuration := delay + jitter
				slog.Warn("ai_fallback_retry", "model", modelName, "attempt", attempt, "delay", sleepDuration.String())

				select {
				case <-time.After(sleepDuration):
				case <-ctx.Done():
					return LLMResponse{}, "", ctx.Err()
				}
			}

			resp, err := provider.CallByModelName(ctx, modelName, req)
			if err == nil {
				slog.Info("ai_fallback_success", "model", resp.Model, "position", modelIdx+1, "attempted", len(attemptedModels), "duration", time.Since(start).String(), "input_tokens", resp.InputTokens, "output_tokens", resp.OutputTokens)
				return resp, resp.Model, nil
			}

			lastErr = err

			// Check if quota/rate limit error
			if !isQuotaError(err) {
				// Non-retryable error — fail immediately
				slog.Error("ai_fallback_nonretryable", "model", modelName, "error", err.Error())
				return LLMResponse{}, "", fmt.Errorf("non-retryable error on %s: %w", modelName, err)
			}

			slog.Warn("ai_fallback_quota", "model", modelName, "attempt", attempt+1, "error", err.Error())
		}

		// Exhausted retries for this model, try next
		nextModel := "none"
		if modelIdx+1 < len(config.Models) {
			nextModel = config.Models[modelIdx+1]
		}
		slog.Warn("ai_fallback_model_exhausted", "model", modelName, "next", nextModel)
	}

	// All models failed
	slog.Error("ai_fallback_all_failed", "attempted", attemptedModels, "duration", time.Since(start).String(), "error", lastErr.Error())
	return LLMResponse{}, "", fmt.Errorf("all models failed (tried: %v): %w", attemptedModels, lastErr)
}
