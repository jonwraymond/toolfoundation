# toolfoundation User Journey

## Overview

This guide walks through the typical usage patterns for toolfoundation,
from defining your first tool to converting between LLM provider formats.

## 1. Installation

```bash
go get github.com/jonwraymond/toolfoundation@latest
```

## 2. Define Your First Tool

```go
import (
  "github.com/jonwraymond/toolfoundation/model"
  "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Create a tool definition
tool := model.Tool{
  Namespace: "calculator",
  Tool: mcp.Tool{
    Name:        "add",
    Description: "Add two numbers together",
    InputSchema: map[string]any{
      "type": "object",
      "properties": map[string]any{
        "a": map[string]any{"type": "number"},
        "b": map[string]any{"type": "number"},
      },
      "required": []string{"a", "b"},
    },
  },
  Tags: model.NormalizeTags([]string{"math", "arithmetic"}),
}

// Validate the tool
if err := tool.Validate(); err != nil {
  log.Fatalf("Invalid tool: %v", err)
}

// Get the canonical ID
fmt.Println(tool.ToolID()) // "calculator:add"
```

## 3. Assign a Backend

```go
// Local handler backend
tool.Backend = model.ToolBackend{
  Kind: model.BackendKindLocal,
  Name: "add_handler",
}

// Or MCP server backend
tool.Backend = model.ToolBackend{
  Kind:       model.BackendKindMCP,
  ServerName: "math-server",
}
```

## 4. Convert to OpenAI Format

```go
import (
  "github.com/jonwraymond/toolfoundation/adapter"
  "github.com/jonwraymond/toolfoundation/adapter/adapters"
)

// Set up the adapter registry
registry := adapter.NewRegistry()
registry.Register(adapters.NewMCPAdapter())
registry.Register(adapters.NewOpenAIAdapter())
registry.Register(adapters.NewAnthropicAdapter())

// Convert MCP tool to OpenAI format
result, err := registry.Convert(tool.Tool, "mcp", "openai")
if err != nil {
  log.Fatalf("Conversion failed: %v", err)
}

// Check for feature loss
for _, warning := range result.Warnings {
  log.Printf("Warning: %s", warning)
}

openaiTool := result.Tool.(adapters.OpenAIFunction)
fmt.Printf("OpenAI function: %s\n", openaiTool.Name)
```

## 5. Round-Trip Conversion

```go
// Convert OpenAI → MCP
result2, err := registry.Convert(openaiTool, "openai", "mcp")
if err != nil {
  log.Fatal(err)
}

mcpTool := result2.Tool.(mcp.Tool)
```

## Common Patterns

### Batch Tool Registration

```go
tools := []model.Tool{
  {Namespace: "math", Tool: mcp.Tool{Name: "add", ...}},
  {Namespace: "math", Tool: mcp.Tool{Name: "subtract", ...}},
  {Namespace: "math", Tool: mcp.Tool{Name: "multiply", ...}},
}

for _, t := range tools {
  if err := t.Validate(); err != nil {
    log.Printf("Skipping invalid tool %s: %v", t.ToolID(), err)
    continue
  }
  // Register with index...
}
```

### Schema Validation

```go
validator := model.NewDefaultValidator()

// Validate input against tool schema
input := map[string]any{"a": 5, "b": 10}
if err := validator.ValidateInput(&tool, input); err != nil {
  log.Fatalf("Invalid input: %v", err)
}
```

### Version Compatibility

```go
import "github.com/jonwraymond/toolfoundation/version"

current := version.MustParse("v1.2.0")
required := version.MustParse("v1.0.0")

if !current.Compatible(required) {
  log.Fatalf("version %s is not compatible with %s", current, required)
}

matrix := version.NewMatrix()
matrix.Add(version.Compatibility{
  Component:  "toolfoundation",
  MinVersion: required,
})

ok, msg := matrix.Check("toolfoundation", current)
if !ok {
  log.Fatal(msg)
}
```

## Next Steps

- Register tools with [tooldiscovery/index](https://github.com/jonwraymond/tooldiscovery)
- Execute tools with [toolexec/run](https://github.com/jonwraymond/toolexec)
- Expose via MCP with [metatools-mcp](https://github.com/jonwraymond/metatools-mcp)
