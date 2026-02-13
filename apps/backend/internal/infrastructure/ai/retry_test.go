package ai

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLLMProvider is a test double for LLMProvider
type mockLLMProvider struct {
	calls     int
	responses []LLMResponse
	errors    []error
}

func (m *mockLLMProvider) Call(_ context.Context, _ LLMRequest) (LLMResponse, error) {
	idx := m.calls
	m.calls++
	if idx < len(m.errors) && m.errors[idx] != nil {
		return LLMResponse{}, m.errors[idx]
	}
	if idx < len(m.responses) {
		return m.responses[idx], nil
	}
	return LLMResponse{}, errors.New("unexpected call")
}

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, time.Second, cfg.BaseDelay)
}

func TestCallWithRetry_SuccessOnFirstAttempt(t *testing.T) {
	provider := &mockLLMProvider{
		responses: []LLMResponse{{Content: "hello", Model: "test"}},
	}

	resp, err := CallWithRetry(context.Background(), provider, LLMRequest{UserPrompt: "hi"}, RetryConfig{MaxRetries: 3, BaseDelay: time.Millisecond})
	require.NoError(t, err)
	assert.Equal(t, "hello", resp.Content)
	assert.Equal(t, 1, provider.calls)
}

func TestCallWithRetry_SuccessAfterRetries(t *testing.T) {
	provider := &mockLLMProvider{
		errors:    []error{errors.New("rate limit exceeded"), errors.New("429 too many requests"), nil},
		responses: []LLMResponse{{}, {}, {Content: "ok", Model: "test"}},
	}

	resp, err := CallWithRetry(context.Background(), provider, LLMRequest{}, RetryConfig{MaxRetries: 3, BaseDelay: time.Millisecond})
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Content)
	assert.Equal(t, 3, provider.calls)
}

func TestCallWithRetry_NonRetryableError(t *testing.T) {
	provider := &mockLLMProvider{
		errors: []error{errors.New("invalid api key")},
	}

	_, err := CallWithRetry(context.Background(), provider, LLMRequest{}, RetryConfig{MaxRetries: 3, BaseDelay: time.Millisecond})
	require.Error(t, err)
	assert.Equal(t, 1, provider.calls) // Should not retry
}

func TestCallWithRetry_ExhaustsRetries(t *testing.T) {
	provider := &mockLLMProvider{
		errors: []error{
			errors.New("rate limit"),
			errors.New("rate limit"),
			errors.New("rate limit"),
			errors.New("rate limit"),
		},
	}

	_, err := CallWithRetry(context.Background(), provider, LLMRequest{}, RetryConfig{MaxRetries: 3, BaseDelay: time.Millisecond})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "after 4 attempts")
	assert.Equal(t, 4, provider.calls) // 1 initial + 3 retries
}

func TestCallWithRetry_ContextCancelled(t *testing.T) {
	provider := &mockLLMProvider{
		errors: []error{errors.New("timeout error"), errors.New("timeout error")},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := CallWithRetry(ctx, provider, LLMRequest{}, RetryConfig{MaxRetries: 3, BaseDelay: time.Second})
	require.Error(t, err)
	// Should fail due to context cancellation during backoff
	assert.True(t, provider.calls <= 2)
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"rate limit", errors.New("rate limit exceeded"), true},
		{"429 status", errors.New("HTTP 429"), true},
		{"timeout", errors.New("request timeout"), true},
		{"temporary", errors.New("temporary failure"), true},
		{"unavailable", errors.New("service unavailable"), true},
		{"503 status", errors.New("HTTP 503"), true},
		{"502 status", errors.New("HTTP 502"), true},
		{"invalid key", errors.New("invalid api key"), false},
		{"bad request", errors.New("bad request"), false},
		{"not found", errors.New("model not found"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.retryable, isRetryableError(tt.err))
		})
	}
}
