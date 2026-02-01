package adapter_test

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jonwraymond/toolfoundation/adapter"
	"github.com/jonwraymond/toolfoundation/model"
)

func ExampleDefaultRegistry() {
	registry := adapter.DefaultRegistry()

	// List available adapters
	adapters := registry.List()
	fmt.Printf("Adapter count: %d\n", len(adapters))
	// Output:
	// Adapter count: 3
}

func ExampleAdapterRegistry_Convert() {
	// Create a tool
	tool := &model.Tool{
		Tool: mcp.Tool{
			Name:        "get_weather",
			Description: "Get current weather for a location",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type":        "string",
						"description": "City name",
					},
				},
				"required": []string{"location"},
			},
		},
	}

	// Convert MCP to OpenAI format
	registry := adapter.DefaultRegistry()
	result, err := registry.Convert(tool, "mcp", "openai")
	if err != nil {
		log.Fatal(err)
	}

	openaiTool := result.Tool.(*adapter.OpenAITool)
	fmt.Println("Type:", openaiTool.Type)
	fmt.Println("Function name:", openaiTool.Function.Name)
	// Output:
	// Type: function
	// Function name: get_weather
}

func ExampleAdapterRegistry_Convert_withWarnings() {
	// Create a tool with features not supported by all adapters
	tool := &model.Tool{
		Tool: mcp.Tool{
			Name:        "search",
			Description: "Search with flexible input",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						// anyOf is not supported by OpenAI
						"anyOf": []any{
							map[string]any{"type": "string"},
							map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						},
					},
				},
			},
		},
	}

	registry := adapter.DefaultRegistry()
	result, _ := registry.Convert(tool, "mcp", "openai")

	if len(result.Warnings) > 0 {
		fmt.Println("Feature loss detected")
		for _, w := range result.Warnings {
			fmt.Printf("  - %s\n", w.Feature)
		}
	}
	// Output:
	// Feature loss detected
	//   - anyOf
}

func ExampleOpenAITool() {
	openaiTool := &adapter.OpenAITool{
		Type: "function",
		Function: adapter.OpenAIFunction{
			Name:        "calculate",
			Description: "Perform a calculation",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"expression": map[string]any{"type": "string"},
				},
				"required": []string{"expression"},
			},
		},
	}

	// Serialize to JSON for API request
	data, _ := json.MarshalIndent(openaiTool, "", "  ")
	fmt.Println(string(data))
	// Output:
	// {
	//   "type": "function",
	//   "function": {
	//     "name": "calculate",
	//     "description": "Perform a calculation",
	//     "parameters": {
	//       "properties": {
	//         "expression": {
	//           "type": "string"
	//         }
	//       },
	//       "required": [
	//         "expression"
	//       ],
	//       "type": "object"
	//     }
	//   }
	// }
}

func ExampleAnthropicTool() {
	anthropicTool := &adapter.AnthropicTool{
		Name:        "search_docs",
		Description: "Search documentation",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{"type": "string"},
			},
			"required": []string{"query"},
		},
	}

	// Serialize to JSON for API request
	data, _ := json.MarshalIndent(anthropicTool, "", "  ")
	fmt.Println(string(data))
	// Output:
	// {
	//   "name": "search_docs",
	//   "description": "Search documentation",
	//   "input_schema": {
	//     "properties": {
	//       "query": {
	//         "type": "string"
	//       }
	//     },
	//     "required": [
	//       "query"
	//     ],
	//     "type": "object"
	//   }
	// }
}

func ExampleNewRegistry() {
	// Create a custom registry with only specific adapters
	registry := adapter.NewRegistry()

	// Register only the adapters you need
	_ = registry.Register(adapter.NewMCPAdapter())
	_ = registry.Register(adapter.NewOpenAIAdapter())

	fmt.Printf("Custom registry has %d adapters\n", len(registry.List()))
	// Output:
	// Custom registry has 2 adapters
}

func ExampleMCPAdapter_SupportsFeature() {
	mcp := adapter.NewMCPAdapter()
	openai := adapter.NewOpenAIAdapter()

	// MCP supports all JSON Schema features
	fmt.Println("MCP supports $ref:", mcp.SupportsFeature(adapter.FeatureRef))
	fmt.Println("MCP supports anyOf:", mcp.SupportsFeature(adapter.FeatureAnyOf))

	// OpenAI has limited support
	fmt.Println("OpenAI supports $ref:", openai.SupportsFeature(adapter.FeatureRef))
	fmt.Println("OpenAI supports anyOf:", openai.SupportsFeature(adapter.FeatureAnyOf))
	// Output:
	// MCP supports $ref: true
	// MCP supports anyOf: true
	// OpenAI supports $ref: false
	// OpenAI supports anyOf: false
}
