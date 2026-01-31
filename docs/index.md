# toolfoundation

Foundation layer providing canonical schema definitions and protocol-agnostic
format conversion for the ApertureStack tool framework.

## Packages

| Package | Purpose |
|---------|---------|
| `model` | Canonical MCP tool schema definitions, validation, backend bindings |
| `adapter` | Protocol-agnostic tool format conversion (MCP, OpenAI, Anthropic) |

## Installation

```bash
go get github.com/jonwraymond/toolfoundation@latest
```

## Quick Start

### Defining a Tool (model package)

```go
import (
  "github.com/jonwraymond/toolfoundation/model"
  "github.com/modelcontextprotocol/go-sdk/mcp"
)

tool := model.Tool{
  Namespace: "github",
  Tool: mcp.Tool{
    Name:        "get_repo",
    Description: "Fetch repository metadata",
    InputSchema: map[string]any{
      "type": "object",
      "properties": map[string]any{
        "owner": map[string]any{"type": "string"},
        "repo":  map[string]any{"type": "string"},
      },
      "required": []string{"owner", "repo"},
    },
  },
  Tags: model.NormalizeTags([]string{"GitHub", "repos"}),
}

if err := tool.Validate(); err != nil {
  log.Fatal(err)
}

fmt.Println(tool.ToolID()) // "github:get_repo"
```

### Converting Between Formats (adapter package)

```go
import (
  "github.com/jonwraymond/toolfoundation/adapter"
  "github.com/jonwraymond/toolfoundation/adapter/adapters"
)

registry := adapter.NewRegistry()
registry.Register(adapters.NewMCPAdapter())
registry.Register(adapters.NewOpenAIAdapter())

result, err := registry.Convert(mcpTool, "mcp", "openai")
if err != nil {
  log.Fatal(err)
}

for _, w := range result.Warnings {
  log.Printf("Feature loss: %s", w)
}
```

## Key Features

- **MCP-aligned**: Tool type embeds official MCP SDK types
- **Protocol-agnostic**: Convert between MCP, OpenAI, and Anthropic formats
- **Loss visibility**: Feature loss during conversion is tracked as warnings
- **Minimal dependencies**: Foundation has minimal external dependencies

## Links

- [model design notes](design-notes.md)
- [user journey](user-journey.md)
- [ai-tools-stack documentation](https://jonwraymond.github.io/ai-tools-stack/)
