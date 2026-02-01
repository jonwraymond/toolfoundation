// Package main demonstrates basic usage of the toolfoundation package.
//
// This example shows how to:
// - Create a tool using the model package
// - Use the default adapter registry
// - Convert tools between MCP, OpenAI, and Anthropic formats
// - Handle feature loss warnings during conversion
package main

import (
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jonwraymond/toolfoundation/adapter"
	"github.com/jonwraymond/toolfoundation/model"
)

func main() {
	// Create a tool using the model package
	tool := &model.Tool{
		Tool: mcp.Tool{
			Name:        "search_documents",
			Description: "Search for documents matching a query",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query",
						"minLength":   1,
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of results",
						"minimum":     1,
						"maximum":     100,
						"default":     10,
					},
					"format": map[string]any{
						"type":        "string",
						"description": "Output format",
						"enum":        []any{"json", "xml", "text"},
					},
					// This feature uses anyOf which has limited support
					"filters": map[string]any{
						"anyOf": []any{
							map[string]any{
								"type": "string",
							},
							map[string]any{
								"type":  "array",
								"items": map[string]any{"type": "string"},
							},
						},
						"description": "Filter criteria (string or array)",
					},
				},
				"required": []any{"query"},
			},
		},
		Namespace: "documents",
		Version:   "1.0.0",
		Tags:      []string{"search", "documents", "query"},
	}

	fmt.Println("=== toolfoundation Example ===")
	fmt.Printf("\nCreated tool: %s (namespace: %s, version: %s)\n",
		tool.Name, tool.Namespace, tool.Version)

	// Get the default registry with all built-in adapters
	registry := adapter.DefaultRegistry()
	fmt.Printf("\nRegistered adapters: %v\n", registry.List())

	// Convert MCP tool to OpenAI format
	fmt.Println("\n--- Converting MCP → OpenAI ---")
	openaiResult, err := registry.Convert(tool, "mcp", "openai")
	if err != nil {
		log.Fatalf("Failed to convert to OpenAI: %v", err)
	}

	// Check for feature loss warnings
	if len(openaiResult.Warnings) > 0 {
		fmt.Println("Feature loss warnings:")
		for _, w := range openaiResult.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	openaiTool := openaiResult.Tool.(*adapter.OpenAITool)
	fmt.Printf("OpenAI tool type: %s\n", openaiTool.Type)
	fmt.Printf("OpenAI function name: %s\n", openaiTool.Function.Name)

	// Convert MCP tool to Anthropic format
	fmt.Println("\n--- Converting MCP → Anthropic ---")
	anthropicResult, err := registry.Convert(tool, "mcp", "anthropic")
	if err != nil {
		log.Fatalf("Failed to convert to Anthropic: %v", err)
	}

	// Anthropic supports anyOf, so fewer warnings expected
	if len(anthropicResult.Warnings) > 0 {
		fmt.Println("Feature loss warnings:")
		for _, w := range anthropicResult.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	} else {
		fmt.Println("No feature loss warnings (Anthropic supports anyOf)")
	}

	anthropicTool := anthropicResult.Tool.(*adapter.AnthropicTool)
	fmt.Printf("Anthropic tool name: %s\n", anthropicTool.Name)

	// Convert back from OpenAI to MCP
	fmt.Println("\n--- Converting OpenAI → MCP (round-trip) ---")
	mcpResult, err := registry.Convert(openaiTool, "openai", "mcp")
	if err != nil {
		log.Fatalf("Failed to convert back to MCP: %v", err)
	}

	restoredTool := mcpResult.Tool.(*model.Tool)
	fmt.Printf("Restored tool name: %s\n", restoredTool.Name)
	fmt.Printf("Restored description: %s\n", restoredTool.Description)

	// Demonstrate cross-format conversion: OpenAI → Anthropic
	fmt.Println("\n--- Converting OpenAI → Anthropic (cross-format) ---")
	crossResult, err := registry.Convert(openaiTool, "openai", "anthropic")
	if err != nil {
		log.Fatalf("Failed cross-format conversion: %v", err)
	}

	crossTool := crossResult.Tool.(*adapter.AnthropicTool)
	fmt.Printf("Cross-converted tool name: %s\n", crossTool.Name)

	// Use backend factory helpers
	fmt.Println("\n--- Using Backend Factory Helpers ---")
	mcpBackend := model.NewMCPBackend("document-server")
	localBackend := model.NewLocalBackend("search-handler")
	providerBackend := model.NewProviderBackend("openai", "gpt-4-search")

	fmt.Printf("MCP Backend: %s (server: %s)\n", mcpBackend.Kind, mcpBackend.MCP.ServerName)
	fmt.Printf("Local Backend: %s (name: %s)\n", localBackend.Kind, localBackend.Local.Name)
	fmt.Printf("Provider Backend: %s (provider: %s, tool: %s)\n",
		providerBackend.Kind, providerBackend.Provider.ProviderID, providerBackend.Provider.ToolID)

	// Demonstrate Clone
	fmt.Println("\n--- Demonstrating Tool.Clone() ---")
	clone := tool.Clone()
	clone.Name = "search_documents_v2"
	clone.Version = "2.0.0"
	fmt.Printf("Original: %s v%s\n", tool.Name, tool.Version)
	fmt.Printf("Clone: %s v%s\n", clone.Name, clone.Version)

	// Demonstrate ParsedVersion
	fmt.Println("\n--- Demonstrating Tool.ParsedVersion() ---")
	v, err := tool.ParsedVersion()
	if err != nil {
		log.Fatalf("Failed to parse version: %v", err)
	}
	fmt.Printf("Parsed version: Major=%d, Minor=%d, Patch=%d\n", v.Major, v.Minor, v.Patch)

	fmt.Println("\n=== Example Complete ===")
}
