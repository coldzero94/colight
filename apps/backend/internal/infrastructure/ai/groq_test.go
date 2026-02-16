package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestGroqServer(t *testing.T, handler http.HandlerFunc) (*GroqProvider, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	provider := &GroqProvider{
		apiKey: "test-key",
		model:  "llama-3.3-70b-versatile",
		client: server.Client(),
	}
	return provider, server
}

func TestGroqProvider_Call_Success(t *testing.T) {
	provider, server := newTestGroqServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Verify request format
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		var req groqRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "llama-3.3-70b-versatile", req.Model)
		assert.Len(t, req.Messages, 2) // system + user
		assert.Equal(t, "system", req.Messages[0].Role)
		assert.Equal(t, "user", req.Messages[1].Role)

		resp := groqResponse{
			ID: "test-id",
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: `{"result":"success"}`}},
			},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
			}{PromptTokens: 10, CompletionTokens: 20},
			Model: "llama-3.3-70b-versatile",
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	})
	defer server.Close()

	// Use a custom transport to redirect requests to test server
	provider.client = &http.Client{
		Transport: &rewriteTransport{base: server.Client().Transport, target: server.URL},
	}

	resp, err := provider.Call(context.Background(), LLMRequest{
		SystemPrompt: "You are a helper",
		UserPrompt:   "Hello",
		Temperature:  0.5,
		MaxTokens:    100,
		JSONMode:     true,
	})
	require.NoError(t, err)
	assert.Equal(t, `{"result":"success"}`, resp.Content)
	assert.Equal(t, 10, resp.InputTokens)
	assert.Equal(t, 20, resp.OutputTokens)
}

func TestGroqProvider_Call_NoSystemPrompt(t *testing.T) {
	provider, server := newTestGroqServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req groqRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Len(t, req.Messages, 1) // only user
		assert.Equal(t, "user", req.Messages[0].Role)

		resp := groqResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "response"}},
			},
			Model: "llama-3.3-70b-versatile",
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	})
	defer server.Close()

	provider.client = &http.Client{
		Transport: &rewriteTransport{base: server.Client().Transport, target: server.URL},
	}

	resp, err := provider.Call(context.Background(), LLMRequest{
		UserPrompt: "Hello",
	})
	require.NoError(t, err)
	assert.Equal(t, "response", resp.Content)
}

func TestGroqProvider_Call_HTTPError(t *testing.T) {
	provider, server := newTestGroqServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate limited"}`))
	})
	defer server.Close()

	provider.client = &http.Client{
		Transport: &rewriteTransport{base: server.Client().Transport, target: server.URL},
	}

	_, err := provider.Call(context.Background(), LLMRequest{UserPrompt: "test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 429")
}

func TestGroqProvider_Call_EmptyResponse(t *testing.T) {
	provider, server := newTestGroqServer(t, func(w http.ResponseWriter, _ *http.Request) {
		resp := groqResponse{Choices: nil}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	})
	defer server.Close()

	provider.client = &http.Client{
		Transport: &rewriteTransport{base: server.Client().Transport, target: server.URL},
	}

	_, err := provider.Call(context.Background(), LLMRequest{UserPrompt: "test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty response")
}

func TestGroqProvider_Call_JSONMode(t *testing.T) {
	provider, server := newTestGroqServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req groqRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.NotNil(t, req.ResponseFormat)
		assert.Equal(t, "json_object", req.ResponseFormat.Type)

		resp := groqResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "{}"}},
			},
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	})
	defer server.Close()

	provider.client = &http.Client{
		Transport: &rewriteTransport{base: server.Client().Transport, target: server.URL},
	}

	_, err := provider.Call(context.Background(), LLMRequest{UserPrompt: "test", JSONMode: true})
	require.NoError(t, err)
}

func TestNewGroqProvider_DefaultModel(t *testing.T) {
	p := NewGroqProvider("my-key", "")
	assert.Equal(t, "my-key", p.apiKey)
	assert.Equal(t, "llama-3.3-70b-versatile", p.model)
	assert.NotNil(t, p.client)
}

func TestNewGroqProvider_CustomModel(t *testing.T) {
	p := NewGroqProvider("my-key", "qwen/qwen3-32b")
	assert.Equal(t, "qwen/qwen3-32b", p.model)
}

func TestNewGroqProvider_AliasResolution(t *testing.T) {
	p := NewGroqProvider("my-key", "qwen3")
	assert.Equal(t, "qwen/qwen3-32b", p.model)

	p = NewGroqProvider("my-key", "llama-4-scout")
	assert.Equal(t, "meta-llama/llama-4-scout-17b-16e-instruct", p.model)

	p = NewGroqProvider("my-key", "compound")
	assert.Equal(t, "groq/compound", p.model)
}

func TestGroqProvider_CallWithModel(t *testing.T) {
	var capturedModel string
	provider, server := newTestGroqServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req groqRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		capturedModel = req.Model

		resp := groqResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "ok"}},
			},
			Model: req.Model,
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	})
	defer server.Close()

	provider.client = &http.Client{
		Transport: &rewriteTransport{base: server.Client().Transport, target: server.URL},
	}

	// Call with alias should resolve to full model ID
	_, err := provider.CallWithModel(context.Background(), "qwen3", LLMRequest{UserPrompt: "test"})
	require.NoError(t, err)
	assert.Equal(t, "qwen/qwen3-32b", capturedModel)

	// Call with full ID should pass through
	_, err = provider.CallWithModel(context.Background(), "openai/gpt-oss-120b", LLMRequest{UserPrompt: "test"})
	require.NoError(t, err)
	assert.Equal(t, "openai/gpt-oss-120b", capturedModel)
}

func TestResolveGroqModel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"qwen3", "qwen/qwen3-32b"},
		{"gpt-oss-120b", "openai/gpt-oss-120b"},
		{"llama-4-scout", "meta-llama/llama-4-scout-17b-16e-instruct"},
		{"kimi-k2", "moonshotai/kimi-k2-instruct"},
		{"compound", "groq/compound"},
		{"llama-3.3", "llama-3.3-70b-versatile"},
		// Full IDs pass through
		{"qwen/qwen3-32b", "qwen/qwen3-32b"},
		{"openai/gpt-oss-120b", "openai/gpt-oss-120b"},
		// Unknown returns as-is
		{"unknown-model", "unknown-model"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ResolveGroqModel(tt.input))
		})
	}
}

// rewriteTransport redirects all requests to the test server
type rewriteTransport struct {
	base   http.RoundTripper
	target string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = t.target[len("http://"):]
	if t.base != nil {
		return t.base.RoundTrip(req)
	}
	return http.DefaultTransport.RoundTrip(req)
}
