package model_test

import (
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jonwraymond/toolfoundation/model"
)

func ExampleTool() {
	tool := model.Tool{
		Namespace: "files",
		Tool: mcp.Tool{
			Name:        "read_file",
			Description: "Read contents of a file",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Path to the file",
					},
				},
				"required": []string{"path"},
			},
		},
		Version: "1.0.0",
		Tags:    []string{"files", "io"},
	}

	fmt.Println("Tool ID:", tool.ToolID())
	fmt.Println("Name:", tool.Name)
	// Output:
	// Tool ID: files:read_file
	// Name: read_file
}

func ExampleTool_Validate() {
	tool := model.Tool{
		Tool: mcp.Tool{
			Name:        "my_tool",
			Description: "A valid tool",
			InputSchema: map[string]any{
				"type": "object",
			},
		},
	}

	if err := tool.Validate(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Tool is valid")
	// Output:
	// Tool is valid
}

func ExampleTool_Clone() {
	original := &model.Tool{
		Namespace: "demo",
		Tool: mcp.Tool{
			Name:        "original_tool",
			Description: "Original description",
			InputSchema: map[string]any{"type": "object"},
		},
		Version: "1.0.0",
	}

	clone := original.Clone()
	clone.Name = "cloned_tool"
	clone.Version = "2.0.0"

	fmt.Println("Original:", original.Name, original.Version)
	fmt.Println("Clone:", clone.Name, clone.Version)
	// Output:
	// Original: original_tool 1.0.0
	// Clone: cloned_tool 2.0.0
}

func ExampleNormalizeTags() {
	tags := []string{
		"  Machine Learning  ",
		"AI",
		"machine-learning", // duplicate after normalization
		"Data Science!",
	}

	normalized := model.NormalizeTags(tags)
	fmt.Println(normalized)
	// Output:
	// [machine-learning ai data-science]
}

func ExampleNewMCPBackend() {
	backend := model.NewMCPBackend("my-mcp-server")
	fmt.Println("Kind:", backend.Kind)
	fmt.Println("Server:", backend.MCP.ServerName)
	// Output:
	// Kind: mcp
	// Server: my-mcp-server
}

func ExampleNewLocalBackend() {
	backend := model.NewLocalBackend("search-handler")
	fmt.Println("Kind:", backend.Kind)
	fmt.Println("Handler:", backend.Local.Name)
	// Output:
	// Kind: local
	// Handler: search-handler
}

func ExampleNewProviderBackend() {
	backend := model.NewProviderBackend("openai", "gpt-4-turbo")
	fmt.Println("Kind:", backend.Kind)
	fmt.Println("Provider:", backend.Provider.ProviderID)
	fmt.Println("Tool:", backend.Provider.ToolID)
	// Output:
	// Kind: provider
	// Provider: openai
	// Tool: gpt-4-turbo
}

func ExampleParseToolID() {
	// With namespace
	ns, name, _ := model.ParseToolID("github:list_repos")
	fmt.Printf("Namespace: %q, Name: %q\n", ns, name)

	// Without namespace
	ns, name, _ = model.ParseToolID("simple_tool")
	fmt.Printf("Namespace: %q, Name: %q\n", ns, name)
	// Output:
	// Namespace: "github", Name: "list_repos"
	// Namespace: "", Name: "simple_tool"
}

func ExampleNewDefaultValidator() {
	validator := model.NewDefaultValidator()

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"age":  map[string]any{"type": "integer", "minimum": 0},
		},
		"required": []string{"name"},
	}

	// Valid input
	err := validator.Validate(schema, map[string]any{
		"name": "Alice",
		"age":  30,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Valid input accepted")

	// Invalid input (missing required field)
	err = validator.Validate(schema, map[string]any{
		"age": 25,
	})
	if err != nil {
		fmt.Println("Invalid input rejected")
	}
	// Output:
	// Valid input accepted
	// Invalid input rejected
}
