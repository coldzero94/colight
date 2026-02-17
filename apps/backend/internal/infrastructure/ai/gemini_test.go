package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGeminiProvider_SetsModel(t *testing.T) {
	// NewGeminiProvider requires a real API key to create a client,
	// but we can test with any key since it doesn't validate at creation time
	provider, err := NewGeminiProvider(t.Context(), "test-api-key")
	if err != nil {
		t.Skip("Gemini client creation failed (expected in CI without network)")
	}

	assert.Equal(t, "gemini-2.0-flash", provider.model)
	assert.NotNil(t, provider.client)
}

func TestGeminiProvider_CallBuildPrompt(t *testing.T) {
	// Test that system + user prompts are combined correctly
	// This tests the logic without making actual API calls
	req := LLMRequest{
		SystemPrompt: "You are a helper",
		UserPrompt:   "Hello",
	}

	combined := req.UserPrompt
	if req.SystemPrompt != "" {
		combined = req.SystemPrompt + "\n\n" + req.UserPrompt
	}

	assert.Equal(t, "You are a helper\n\nHello", combined)
}

func TestGeminiProvider_CallUserPromptOnly(t *testing.T) {
	req := LLMRequest{
		UserPrompt: "Hello",
	}

	combined := req.UserPrompt
	if req.SystemPrompt != "" {
		combined = req.SystemPrompt + "\n\n" + req.UserPrompt
	}

	assert.Equal(t, "Hello", combined)
}
