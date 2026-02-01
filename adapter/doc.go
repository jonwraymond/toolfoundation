// Package adapter provides protocol-agnostic tool format conversion.
// It enables bidirectional transformation between MCP, OpenAI, and Anthropic
// tool definitions through a canonical intermediate representation.
//
// This is a pure data-transform library with no I/O, network, or runtime execution.
//
// # Overview
//
// The adapter package uses a hub-and-spoke architecture where all conversions
// pass through a canonical intermediate format (CanonicalTool). This allows
// N adapters to support N² conversions with only N implementations.
//
// # Basic Usage
//
// Use the default registry with all built-in adapters:
//
//	registry := adapter.DefaultRegistry()
//
//	// Convert an MCP tool to OpenAI format
//	result, err := registry.Convert(mcpTool, "mcp", "openai")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Check for feature loss warnings
//	for _, w := range result.Warnings {
//	    log.Printf("Warning: %s", w)
//	}
//
//	openaiTool := result.Tool.(*adapter.OpenAITool)
//
// # Supported Formats
//
// The package includes adapters for three tool formats:
//
//   - MCP (Model Context Protocol) - Full JSON Schema 2020-12 support
//   - OpenAI - Function calling format with strict mode support
//   - Anthropic - Tool use format with anyOf support
//
// # Feature Loss Warnings
//
// Different formats support different JSON Schema features. When converting
// from a format with more features to one with fewer, the adapter emits
// warnings indicating which features were lost:
//
//	result, _ := registry.Convert(tool, "mcp", "openai")
//	if len(result.Warnings) > 0 {
//	    fmt.Println("The following features are not supported by OpenAI:")
//	    for _, w := range result.Warnings {
//	        fmt.Printf("  - %s at %s\n", w.Feature, w.Path)
//	    }
//	}
//
// # Feature Support Matrix
//
//	Feature          MCP    OpenAI  Anthropic
//	─────────────────────────────────────────
//	$ref/$defs       Yes    No      No
//	anyOf            Yes    No      Yes
//	oneOf            Yes    No      No
//	allOf            Yes    No      No
//	not              Yes    No      No
//	pattern          Yes    Yes     Yes
//	format           Yes    No      Yes
//	enum/const       Yes    Yes     Yes
//	min/max          Yes    Yes     Yes
//
// # Custom Adapters
//
// Implement the Adapter interface to add support for new formats:
//
//	type Adapter interface {
//	    Name() string
//	    ToCanonical(raw any) (*CanonicalTool, error)
//	    FromCanonical(ct *CanonicalTool) (any, error)
//	    SupportsFeature(feature SchemaFeature) bool
//	}
//
// Register custom adapters with the registry:
//
//	registry := adapter.NewRegistry()
//	registry.Register(adapter.NewMCPAdapter())
//	registry.Register(myCustomAdapter)
//
// # Type Definitions
//
// The package defines local types for OpenAI and Anthropic formats to avoid
// SDK coupling:
//
//   - OpenAITool / OpenAIFunction - OpenAI function calling format
//   - AnthropicTool - Anthropic tool use format
//
// These types can be serialized directly to JSON for API requests.
//
// # Thread Safety
//
// The AdapterRegistry is thread-safe for concurrent reads after initial
// registration. Register all adapters during initialization before
// concurrent use.
package adapter
