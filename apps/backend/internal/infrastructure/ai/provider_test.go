package ai

import (
	"context"
	"testing"

	"github.com/coby/colight/apps/backend/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAIProvider_GeminiDefault(t *testing.T) {
	cfg := &config.Config{
		LLMLightProvider: "gemini",
		GeminiAPIKey:     "test-key",
	}

	provider, err := NewAIProvider(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.Light())
}

func TestNewAIProvider_GeminiMissingKey(t *testing.T) {
	cfg := &config.Config{
		LLMLightProvider: "gemini",
		GeminiAPIKey:     "",
	}

	_, err := NewAIProvider(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GEMINI_API_KEY is required")
}

func TestNewAIProvider_GroqProvider(t *testing.T) {
	cfg := &config.Config{
		LLMLightProvider: "groq",
		GroqAPIKey:       "test-groq-key",
	}

	provider, err := NewAIProvider(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.Light())
}

func TestNewAIProvider_GroqMissingKey(t *testing.T) {
	cfg := &config.Config{
		LLMLightProvider: "groq",
		GroqAPIKey:       "",
	}

	_, err := NewAIProvider(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GROQ_API_KEY is required")
}

func TestNewAIProvider_EmptyProviderDefaultsToGemini(t *testing.T) {
	cfg := &config.Config{
		LLMLightProvider: "",
		GeminiAPIKey:     "test-key",
	}

	provider, err := NewAIProvider(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestAIProvider_Close(t *testing.T) {
	cfg := &config.Config{
		LLMLightProvider: "groq",
		GroqAPIKey:       "test-key",
	}

	provider, err := NewAIProvider(context.Background(), cfg)
	require.NoError(t, err)

	err = provider.Close()
	assert.NoError(t, err)
}

func TestAIProvider_CallLight(t *testing.T) {
	// Create provider with Groq (doesn't need external connection at creation)
	cfg := &config.Config{
		LLMLightProvider: "groq",
		GroqAPIKey:       "test-key",
	}

	provider, err := NewAIProvider(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, provider.Light())
}
