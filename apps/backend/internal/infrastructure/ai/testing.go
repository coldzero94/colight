package ai

// NewAIProviderForTest creates an AIProvider with injected LLMProviders for testing.
func NewAIProviderForTest(light, heavy LLMProvider) *AIProvider {
	return &AIProvider{light: light, heavy: heavy}
}
