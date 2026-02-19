package ai

import (
	"context"
	"errors"
	"strings"
)

// ClassifyError maps an AI provider error to a canonical error_type string.
// Returns "" for nil errors.
func ClassifyError(err error) string {
	if err == nil {
		return ""
	}

	// Check for context deadline (timeout)
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}

	msg := strings.ToLower(err.Error())

	// Rate limit (HTTP 429 or quota keywords)
	if strings.Contains(msg, "429") ||
		strings.Contains(msg, "quota") ||
		strings.Contains(msg, "rate limit") {
		return "rate_limit"
	}

	// Timeout keywords
	if strings.Contains(msg, "timeout") {
		return "timeout"
	}

	// Context/token limit exceeded
	if strings.Contains(msg, "context length") ||
		strings.Contains(msg, "too many tokens") {
		return "context_exceeded"
	}

	// Provider 5xx errors
	if strings.Contains(msg, "500") ||
		strings.Contains(msg, "502") ||
		strings.Contains(msg, "503") {
		return "provider_error"
	}

	// Default: bad request / unknown
	return "invalid_request"
}
