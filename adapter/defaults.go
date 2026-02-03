package adapter

// DefaultRegistry returns a registry pre-configured with all built-in adapters.
// The registry includes MCP, OpenAI, Anthropic, A2A, and Gemini adapters.
func DefaultRegistry() *AdapterRegistry {
	registry := NewRegistry()

	// Register all built-in adapters
	_ = registry.Register(NewMCPAdapter())
	_ = registry.Register(NewOpenAIAdapter())
	_ = registry.Register(NewAnthropicAdapter())
	_ = registry.Register(NewA2AAdapter())
	_ = registry.Register(NewGeminiAdapter())

	return registry
}
