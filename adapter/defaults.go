package adapter

// DefaultRegistry returns a registry pre-configured with all built-in adapters.
// The registry includes MCP, OpenAI, and Anthropic adapters.
func DefaultRegistry() *AdapterRegistry {
	registry := NewRegistry()

	// Register all built-in adapters
	_ = registry.Register(NewMCPAdapter())
	_ = registry.Register(NewOpenAIAdapter())
	_ = registry.Register(NewAnthropicAdapter())

	return registry
}
