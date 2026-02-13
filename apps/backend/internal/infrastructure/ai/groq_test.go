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
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	// Override the URL by replacing the client
	origCall := provider.Call
	_ = origCall
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
		json.NewEncoder(w).Encode(resp)
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
		w.Write([]byte(`{"error":"rate limited"}`))
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
		json.NewEncoder(w).Encode(resp)
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
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	provider.client = &http.Client{
		Transport: &rewriteTransport{base: server.Client().Transport, target: server.URL},
	}

	_, err := provider.Call(context.Background(), LLMRequest{UserPrompt: "test", JSONMode: true})
	require.NoError(t, err)
}

func TestNewGroqProvider(t *testing.T) {
	p := NewGroqProvider("my-key")
	assert.Equal(t, "my-key", p.apiKey)
	assert.Equal(t, "llama-3.3-70b-versatile", p.model)
	assert.NotNil(t, p.client)
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
