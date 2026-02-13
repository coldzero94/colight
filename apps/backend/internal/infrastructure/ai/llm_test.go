package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLLMRequest_Fields(t *testing.T) {
	req := LLMRequest{
		SystemPrompt: "system",
		UserPrompt:   "user",
		Temperature:  0.7,
		MaxTokens:    1000,
		JSONMode:     true,
	}

	assert.Equal(t, "system", req.SystemPrompt)
	assert.Equal(t, "user", req.UserPrompt)
	assert.Equal(t, 0.7, req.Temperature)
	assert.Equal(t, 1000, req.MaxTokens)
	assert.True(t, req.JSONMode)
}

func TestLLMResponse_Fields(t *testing.T) {
	resp := LLMResponse{
		Content:      "response content",
		InputTokens:  100,
		OutputTokens: 50,
		Model:        "test-model",
	}

	assert.Equal(t, "response content", resp.Content)
	assert.Equal(t, 100, resp.InputTokens)
	assert.Equal(t, 50, resp.OutputTokens)
	assert.Equal(t, "test-model", resp.Model)
}

func TestMockLLMProvider_ImplementsInterface(t *testing.T) {
	// Verify that mockLLMProvider satisfies LLMProvider interface
	var _ LLMProvider = (*mockLLMProvider)(nil)
}
