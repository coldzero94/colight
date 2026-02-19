package ai_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/stretchr/testify/assert"
)

func TestClassifyAIError_RateLimit(t *testing.T) {
	cases := []error{
		errors.New("HTTP 429: rate limit exceeded"),
		errors.New("quota exceeded for model"),
		errors.New("googleapi: Error 429: Quota exceeded"),
	}
	for _, err := range cases {
		assert.Equal(t, "rate_limit", ai.ClassifyError(err), "err: %v", err)
	}
}

func TestClassifyAIError_Timeout(t *testing.T) {
	cases := []error{
		context.DeadlineExceeded,
		fmt.Errorf("wrapped: %w", context.DeadlineExceeded),
		errors.New("request timeout after 30s"),
	}
	for _, err := range cases {
		assert.Equal(t, "timeout", ai.ClassifyError(err), "err: %v", err)
	}
}

func TestClassifyAIError_ProviderError(t *testing.T) {
	cases := []error{
		errors.New("HTTP 503: Service Unavailable"),
		errors.New("HTTP 500: Internal Server Error"),
		errors.New("HTTP 502: Bad Gateway"),
	}
	for _, err := range cases {
		assert.Equal(t, "provider_error", ai.ClassifyError(err), "err: %v", err)
	}
}

func TestClassifyAIError_ContextExceeded(t *testing.T) {
	cases := []error{
		errors.New("context length exceeded: too many tokens"),
		errors.New("too many tokens in request"),
	}
	for _, err := range cases {
		assert.Equal(t, "context_exceeded", ai.ClassifyError(err), "err: %v", err)
	}
}

func TestClassifyAIError_InvalidRequest(t *testing.T) {
	cases := []error{
		errors.New("HTTP 400: invalid request"),
		errors.New("HTTP 400: Bad Request"),
		errors.New("invalid argument: model not supported"),
	}
	for _, err := range cases {
		assert.Equal(t, "invalid_request", ai.ClassifyError(err), "err: %v", err)
	}
}

func TestClassifyAIError_Unknown(t *testing.T) {
	cases := []error{
		errors.New("unknown error occurred"),
		errors.New("something completely unexpected"),
	}
	for _, err := range cases {
		assert.Equal(t, "unknown", ai.ClassifyError(err), "err: %v", err)
	}
}

func TestClassifyAIError_Nil(t *testing.T) {
	assert.Equal(t, "", ai.ClassifyError(nil))
}
