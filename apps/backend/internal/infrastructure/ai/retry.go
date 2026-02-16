package ai

import (
	"context"
	"fmt"
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

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * config.BaseDelay
			jitter := time.Duration(rand.Int63n(int64(delay / 2)))
			sleepDuration := delay + jitter

			select {
			case <-time.After(sleepDuration):
			case <-ctx.Done():
				return LLMResponse{}, ctx.Err()
			}
		}

		resp, err := provider.Call(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on certain errors
		if !isRetryableError(err) {
			break
		}
	}

	return LLMResponse{}, fmt.Errorf("AI call failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}

// CallByModelNameWithRetry calls AIProvider.CallByModelName with exponential backoff retry logic
func CallByModelNameWithRetry(ctx context.Context, provider *AIProvider, modelName string, req LLMRequest, config RetryConfig) (LLMResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * config.BaseDelay
			jitter := time.Duration(rand.Int63n(int64(delay / 2)))
			sleepDuration := delay + jitter

			select {
			case <-time.After(sleepDuration):
			case <-ctx.Done():
				return LLMResponse{}, ctx.Err()
			}
		}

		resp, err := provider.CallByModelName(ctx, modelName, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on certain errors
		if !isRetryableError(err) {
			break
		}
	}

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
