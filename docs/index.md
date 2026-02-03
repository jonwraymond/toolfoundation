# toolfoundation

Foundation layer providing canonical schema definitions and protocol-agnostic
format conversion for the ApertureStack tool framework.

## Packages

| Package | Purpose |
|---------|---------|
| `model` | Canonical MCP tool schema definitions, validation, backend bindings |
| `adapter` | Protocol-agnostic tool format conversion (MCP, OpenAI, Anthropic, A2A, Gemini) |
| `version` | Semantic version parsing, constraints, compatibility matrices |

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

fmt.Println(tool.ToolID()) // "github:get_repo" (or "github:get_repo:1.0.0" when version is set)
```

### Converting Between Formats (adapter package)

```go
import "github.com/jonwraymond/toolfoundation/adapter"

// Use the default registry with all built-in adapters
registry := adapter.DefaultRegistry()

result, err := registry.Convert(mcpTool, "mcp", "openai")
if err != nil {
  log.Fatal(err)
}

// Check for feature loss warnings
for _, w := range result.Warnings {
  log.Printf("Feature loss: %s", w)
}
```

### Versioning Utilities (version package)

```go
import "github.com/jonwraymond/toolfoundation/version"

v1 := version.MustParse("v1.2.0")
v2 := version.MustParse("v1.3.1")

if v2.GreaterThan(v1) {
  fmt.Println("upgrade available")
}

matrix := version.NewMatrix()
matrix.Add(version.Compatibility{
  Component:  "toolexec",
  MinVersion: version.MustParse("v1.0.0"),
})

ok, msg := matrix.Check("toolexec", v1)
if !ok {
  log.Fatal(msg)
}
```

## Key Features

- **MCP-aligned**: Tool type embeds official MCP SDK types
- **Protocol-agnostic**: Convert between MCP, OpenAI, and Anthropic formats
- **Loss visibility**: Feature loss during conversion is tracked as warnings
- **Minimal dependencies**: Foundation has minimal external dependencies

## Schema contracts

See the dedicated schema reference for field constraints, JSON Schema rules,
and recommended patterns:

- [tool schemas](schemas.md)

## Schema Validation Policy

Schema validation follows JSON Schema 2020-12 by default with draft-07 support.
External `$ref` resolution is disabled to prevent network access. See
[design notes](design-notes.md) for details and limitations.

## Links

- [model design notes](design-notes.md)
- [user journey](user-journey.md)
- [tool schemas](schemas.md)
- [ai-tools-stack documentation](https://jonwraymond.github.io/ai-tools-stack/)
