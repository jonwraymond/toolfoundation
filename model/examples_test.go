package model_test

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jonwraymond/toolfoundation/model"
)

// Example_createTool demonstrates creating a Tool with the toolmodel library.
func Example_createTool() {
	// Create a tool that embeds the MCP Tool type
	tool := model.Tool{
		Tool: mcp.Tool{
			Name:        "search",
			Description: "Search for documents by query",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of results",
						"default":     10,
					},
				},
				"required": []any{"query"},
			},
		},
		Namespace: "docs",
		Version:   "1.0.0",
	}

	fmt.Printf("Tool ID: %s\n", tool.ToolID())
	fmt.Printf("Name: %s\n", tool.Name)
	fmt.Printf("Description: %s\n", tool.Description)
	// Output:
	// Tool ID: docs:search
	// Name: search
	// Description: Search for documents by query
}

// Example_parseToolID demonstrates parsing a tool ID into its components.
func Example_parseToolID() {
	// Parse a namespaced tool ID
	namespace, name, err := model.ParseToolID("filesystem:read")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Namespace: %q, Name: %q\n", namespace, name)

	// Parse a simple tool ID (no namespace)
	namespace, name, err = model.ParseToolID("echo")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Namespace: %q, Name: %q\n", namespace, name)
	// Output:
	// Namespace: "filesystem", Name: "read"
	// Namespace: "", Name: "echo"
}

// Example_toolJSON demonstrates JSON serialization of tools.
func Example_toolJSON() {
	tool := model.Tool{
		Tool: mcp.Tool{
			Name:        "greet",
			Description: "Greet a user",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
			},
		},
		Namespace: "example",
		Version:   "2.0.0",
	}

	// ToMCPJSON strips toolmodel extensions (namespace, version)
	mcpJSON, _ := tool.ToMCPJSON()
	fmt.Println("MCP JSON (no namespace/version):")
	var mcpResult map[string]any
	if err := json.Unmarshal(mcpJSON, &mcpResult); err != nil {
		log.Fatal(err)
	}
	_, hasNamespace := mcpResult["namespace"]
	fmt.Printf("  Has namespace: %v\n", hasNamespace)

	// ToJSON includes all fields
	fullJSON, _ := tool.ToJSON()
	fmt.Println("Full JSON (with namespace/version):")
	var fullResult map[string]any
	if err := json.Unmarshal(fullJSON, &fullResult); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  namespace: %v\n", fullResult["namespace"])
	fmt.Printf("  version: %v\n", fullResult["version"])
	// Output:
	// MCP JSON (no namespace/version):
	//   Has namespace: false
	// Full JSON (with namespace/version):
	//   namespace: example
	//   version: 2.0.0
}

// Example_fromMCPJSON demonstrates deserializing an MCP Tool from JSON.
func Example_fromMCPJSON() {
	mcpJSON := `{
		"name": "calculate",
		"description": "Perform calculations",
		"inputSchema": {
			"type": "object",
			"properties": {
				"expression": {"type": "string"}
			},
			"required": ["expression"]
		}
	}`

	tool, err := model.FromMCPJSON([]byte(mcpJSON))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %s\n", tool.Name)
	fmt.Printf("Namespace: %q (empty from MCP JSON)\n", tool.Namespace)
	// Output:
	// Name: calculate
	// Namespace: "" (empty from MCP JSON)
}

// Example_validateInput demonstrates validating tool input with SchemaValidator.
func Example_validateInput() {
	tool := model.Tool{
		Tool: mcp.Tool{
			Name:        "send_email",
			Description: "Send an email",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"to":      map[string]any{"type": "string"},
					"subject": map[string]any{"type": "string"},
					"body":    map[string]any{"type": "string"},
				},
				"required": []any{"to", "subject"},
			},
		},
	}

	validator := model.NewDefaultValidator()

	// Valid input
	validArgs := map[string]any{
		"to":      "user@example.com",
		"subject": "Hello",
		"body":    "Hi there!",
	}
	err := validator.ValidateInput(&tool, validArgs)
	fmt.Printf("Valid input: error=%v\n", err)

	// Invalid input (missing required field)
	invalidArgs := map[string]any{
		"body": "Hi there!",
	}
	err = validator.ValidateInput(&tool, invalidArgs)
	fmt.Printf("Invalid input: error=%v\n", err != nil)
	// Output:
	// Valid input: error=<nil>
	// Invalid input: error=true
}

// Example_validateSchema demonstrates direct schema validation.
func Example_validateSchema() {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"age":  map[string]any{"type": "integer", "minimum": 0},
		},
		"required": []any{"name"},
	}

	validator := model.NewDefaultValidator()

	// Valid instance
	valid := map[string]any{"name": "Alice", "age": 30}
	err := validator.Validate(schema, valid)
	fmt.Printf("Valid: %v\n", err == nil)

	// Invalid instance (wrong type)
	invalid := map[string]any{"name": 123}
	err = validator.Validate(schema, invalid)
	fmt.Printf("Invalid type: %v\n", err != nil)

	// Invalid instance (missing required)
	missing := map[string]any{"age": 25}
	err = validator.Validate(schema, missing)
	fmt.Printf("Missing required: %v\n", err != nil)
	// Output:
	// Valid: true
	// Invalid type: true
	// Missing required: true
}

// Example_backendBinding demonstrates creating backend binding information.
func Example_backendBinding() {
	// MCP backend (tool from an MCP server)
	mcpBackend := model.ToolBackend{
		Kind: model.BackendKindMCP,
		MCP: &model.MCPBackend{
			ServerName: "filesystem-server",
		},
	}
	fmt.Printf("MCP Backend: kind=%s, server=%s\n", mcpBackend.Kind, mcpBackend.MCP.ServerName)

	// Provider backend (external tool provider)
	providerBackend := model.ToolBackend{
		Kind: model.BackendKindProvider,
		Provider: &model.ProviderBackend{
			ProviderID: "openai",
			ToolID:     "gpt-4-vision",
		},
	}
	fmt.Printf("Provider Backend: kind=%s, provider=%s, tool=%s\n",
		providerBackend.Kind, providerBackend.Provider.ProviderID, providerBackend.Provider.ToolID)

	// Local backend (locally implemented tool)
	localBackend := model.ToolBackend{
		Kind: model.BackendKindLocal,
		Local: &model.LocalBackend{
			Name: "custom-handler",
		},
	}
	fmt.Printf("Local Backend: kind=%s, name=%s\n", localBackend.Kind, localBackend.Local.Name)
	// Output:
	// MCP Backend: kind=mcp, server=filesystem-server
	// Provider Backend: kind=provider, provider=openai, tool=gpt-4-vision
	// Local Backend: kind=local, name=custom-handler
}

// Example_noParametersTool demonstrates creating a tool with no parameters.
func Example_noParametersTool() {
	// MCP recommended schema for tools with no parameters
	tool := model.Tool{
		Tool: mcp.Tool{
			Name:        "get_time",
			Description: "Get the current time",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
			},
		},
	}

	validator := model.NewDefaultValidator()

	// Empty object is valid
	err := validator.ValidateInput(&tool, map[string]any{})
	fmt.Printf("Empty object: valid=%v\n", err == nil)

	// Extra properties are rejected
	err = validator.ValidateInput(&tool, map[string]any{"unexpected": "value"})
	fmt.Printf("Extra properties: rejected=%v\n", err != nil)
	// Output:
	// Empty object: valid=true
	// Extra properties: rejected=true
}
