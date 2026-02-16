package ai

import "context"

// NewAIProviderForTest creates an AIProvider with injected mocks for testing.
// light handles gemini/groq routes, heavy handles claude routes.
func NewAIProviderForTest(light, heavy LLMProvider) *AIProvider {
	return &AIProvider{
		gemini:  light,
		claude:  heavy,
		groqLLM: light,
	}
}

// MockStreamingProvider is a test double for StreamingLLMProvider.
type MockStreamingProvider struct {
	Response LLMResponse
	Chunks   []string
	Err      error
	Calls    int
}

func (m *MockStreamingProvider) Call(ctx context.Context, req LLMRequest) (LLMResponse, error) {
	m.Calls++
	if m.Err != nil {
		return LLMResponse{}, m.Err
	}
	return m.Response, nil
}

func (m *MockStreamingProvider) Stream(ctx context.Context, req LLMRequest, onChunk StreamCallback) (LLMResponse, error) {
	m.Calls++
	if m.Err != nil {
		return LLMResponse{}, m.Err
	}
	for _, chunk := range m.Chunks {
		if ctx.Err() != nil {
			return LLMResponse{}, ctx.Err()
		}
		onChunk(chunk)
	}
	return m.Response, nil
}
