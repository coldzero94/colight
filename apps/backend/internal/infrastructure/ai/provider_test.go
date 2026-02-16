package ai

import (
	"context"
	"testing"

	"github.com/coby/colight/apps/backend/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAIProvider_GroqOnly(t *testing.T) {
	cfg := &config.Config{
		GroqAPIKey: "test-groq-key",
	}

	provider, err := NewAIProvider(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.Groq())
}

func TestNewAIProvider_NoKeysReturnsError(t *testing.T) {
	cfg := &config.Config{}

	_, err := NewAIProvider(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no AI provider available")
}

func TestAIProvider_Close(t *testing.T) {
	cfg := &config.Config{
		GroqAPIKey: "test-key",
	}

	provider, err := NewAIProvider(context.Background(), cfg)
	require.NoError(t, err)

	err = provider.Close()
	assert.NoError(t, err)
}

func TestAIProvider_CallByModelName_GroqRoute(t *testing.T) {
	mock := &MockStreamingProvider{
		Response: LLMResponse{Content: "test"},
	}
	provider := NewAIProviderForTest(mock, nil)

	resp, err := provider.CallByModelName(context.Background(), "groq", LLMRequest{
		UserPrompt: "hello",
	})
	require.NoError(t, err)
	assert.Equal(t, "test", resp.Content)
}

func TestAIProvider_CallByModelName_ClaudeMissing(t *testing.T) {
	mock := &MockStreamingProvider{}
	provider := NewAIProviderForTest(mock, nil)

	_, err := provider.CallByModelName(context.Background(), "claude-sonnet-4-5", LLMRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Claude not available")
}

func TestAIProvider_CallByModelName_GeminiRoute(t *testing.T) {
	mock := &MockStreamingProvider{
		Response: LLMResponse{Content: "gemini-response"},
	}
	provider := NewAIProviderForTest(mock, nil)

	resp, err := provider.CallByModelName(context.Background(), "gemini-2.0-flash", LLMRequest{
		UserPrompt: "hello",
	})
	require.NoError(t, err)
	assert.Equal(t, "gemini-response", resp.Content)
}
