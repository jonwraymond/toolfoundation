# toolfoundation

[![CI](https://github.com/jonwraymond/toolfoundation/actions/workflows/ci.yml/badge.svg)](https://github.com/jonwraymond/toolfoundation/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/jonwraymond/toolfoundation.svg)](https://pkg.go.dev/github.com/jonwraymond/toolfoundation)
[![Go Report Card](https://goreportcard.com/badge/github.com/jonwraymond/toolfoundation)](https://goreportcard.com/report/github.com/jonwraymond/toolfoundation)

Foundation layer for the [ApertureStack](https://github.com/jonwraymond) AI tool ecosystem.

## Installation

```bash
go get github.com/jonwraymond/toolfoundation
```

## Packages

| Package | Description |
|---------|-------------|
| [`model`](https://pkg.go.dev/github.com/jonwraymond/toolfoundation/model) | Canonical MCP tool schema definitions and validation |
| [`adapter`](https://pkg.go.dev/github.com/jonwraymond/toolfoundation/adapter) | Protocol-agnostic tool format conversion (MCP, OpenAI, Anthropic) |
| [`version`](https://pkg.go.dev/github.com/jonwraymond/toolfoundation/version) | Semantic versioning, constraints, compatibility matrices |

## Quick Start

### Define a Tool

```go
import (
    "github.com/jonwraymond/toolfoundation/model"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

tool := model.Tool{
    Namespace: "github",
    Tool: mcp.Tool{
        Name:        "list_repos",
        Description: "List repositories for a user",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "username": map[string]any{"type": "string"},
            },
            "required": []string{"username"},
        },
    },
    Version: "1.0.0",
}

if err := tool.Validate(); err != nil {
    log.Fatal(err)
}
```

### Convert Between Formats

```go
import "github.com/jonwraymond/toolfoundation/adapter"

registry := adapter.DefaultRegistry()

// Convert MCP tool to OpenAI format
result, err := registry.Convert(tool, "mcp", "openai")
if err != nil {
    log.Fatal(err)
}

// Check for feature loss warnings
for _, w := range result.Warnings {
    log.Printf("Warning: %s", w)
}

openaiTool := result.Tool.(*adapter.OpenAITool)
```

### Version Constraints

```go
import "github.com/jonwraymond/toolfoundation/version"

v := version.MustParse("1.5.0")
constraint, _ := version.ParseConstraint("^1.0.0")

if constraint.Check(v) {
    fmt.Println("Version is compatible")
}
```

## Documentation

- [API Reference (pkg.go.dev)](https://pkg.go.dev/github.com/jonwraymond/toolfoundation)
- [Tool Schemas](./docs/schemas.md)
- [Design Notes](./docs/design-notes.md)
- [User Journey](./docs/user-journey.md)
- [Contributing](./CONTRIBUTING.md)

## Features

- **MCP-aligned**: Tool type embeds official MCP SDK types
- **Protocol-agnostic**: Convert between MCP, OpenAI, and Anthropic formats
- **Feature loss visibility**: Warnings when target format lacks source features
- **JSON Schema validation**: Built-in schema validation (2020-12 and draft-07)
- **Minimal dependencies**: Foundation layer with minimal external dependencies

## License

MIT License - see [LICENSE](./LICENSE)
