package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClaudeProvider(t *testing.T) {
	p := NewClaudeProvider("test-api-key")
	require.NotNil(t, p)
	assert.Equal(t, "claude-sonnet-4-5-20250929", string(p.model))
}

func TestClaudeProvider_ImplementsLLMProvider(t *testing.T) {
	var _ LLMProvider = (*ClaudeProvider)(nil)
}

func TestClaudeProvider_ImplementsStreamingLLMProvider(t *testing.T) {
	var _ StreamingLLMProvider = (*ClaudeProvider)(nil)
}

func TestClaudeProvider_CallWithInvalidKey(t *testing.T) {
	p := NewClaudeProvider("invalid-key")

	_, err := p.Call(context.Background(), LLMRequest{
		SystemPrompt: "You are a test assistant",
		UserPrompt:   "Hello",
		MaxTokens:    10,
	})

	// Should fail with API error (invalid key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Claude API call failed")
}

func TestClaudeProvider_StreamWithInvalidKey(t *testing.T) {
	p := NewClaudeProvider("invalid-key")

	chunks := []string{}
	_, err := p.Stream(context.Background(), LLMRequest{
		SystemPrompt: "You are a test assistant",
		UserPrompt:   "Hello",
		MaxTokens:    10,
	}, func(chunk string) {
		chunks = append(chunks, chunk)
	})

	// Should fail with API error
	assert.Error(t, err)
	assert.Empty(t, chunks) // No chunks should have been received
}
