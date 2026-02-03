# Architecture

This document describes the architecture of the toolfoundation package.

## Package Structure

```
toolfoundation/
├── model/      # Canonical tool definitions
├── adapter/    # Format conversion
├── version/    # Semantic versioning
└── examples/   # Usage examples
```

## Dependency Graph

```
┌─────────────────────────────────────────────────────────┐
│                     External                            │
│  ┌─────────────────────────────────────────────────┐   │
│  │  github.com/modelcontextprotocol/go-sdk/mcp     │   │
│  └─────────────────────────────────────────────────┘   │
│                          ▲                              │
└──────────────────────────│──────────────────────────────┘
                           │
┌──────────────────────────│──────────────────────────────┐
│                   toolfoundation                        │
│                          │                              │
│  ┌───────────┐           │                              │
│  │  version  │ (standalone - no internal deps)          │
│  └───────────┘           │                              │
│       ▲                  │                              │
│       │                  │                              │
│  ┌────┴──────────────────┴─────┐                       │
│  │           model             │                        │
│  │  - Tool type (embeds mcp)   │                        │
│  │  - Schema validation        │                        │
│  │  - Backend bindings         │                        │
│  └─────────────────────────────┘                        │
│                ▲                                        │
│                │                                        │
│  ┌─────────────┴───────────────┐                       │
│  │          adapter            │                        │
│  │  - CanonicalTool            │                        │
│  │  - CanonicalProvider        │                        │
│  │  - MCPAdapter               │                        │
│  │  - OpenAIAdapter            │                        │
│  │  - AnthropicAdapter         │                        │
│  │  - A2AAdapter               │                        │
│  │  - GeminiAdapter            │                        │
│  │  - AdapterRegistry          │                        │
│  └─────────────────────────────┘                        │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## Package Responsibilities

### version

**Standalone package with no internal dependencies.**

- Semantic version parsing (`Parse`, `MustParse`)
- Version comparison (`Compare`, `LessThan`, `GreaterThan`, `Equal`)
- Version constraints (`ParseConstraint`, `Check`)
- Compatibility matrices (`Matrix`, `Negotiate`)

### model

**Depends on: version, mcp-go SDK**

- Canonical `Tool` type (embeds `mcp.Tool`)
- Tool extensions: `Namespace`, `Version`, `Tags`, `Backend`
- Schema validation (`SchemaValidator`, `DefaultValidator`)
- Tag normalization (`NormalizeTags`)
- Backend factory functions (`NewMCPBackend`, `NewLocalBackend`, `NewProviderBackend`)

### adapter

**Depends on: model**

- `CanonicalTool` - intermediate representation
- `CanonicalProvider` - provider/agent metadata representation
- `Adapter` interface for format converters
- Built-in adapters: MCP, OpenAI, Anthropic, A2A, Gemini
- `AdapterRegistry` for managing converters
- Feature loss detection and warnings

## Data Flow

### Tool Conversion

```
Source Format    Canonical        Target Format
─────────────    ─────────        ─────────────

┌─────────┐      ┌─────────┐      ┌─────────┐
│   MCP   │─────▶│Canonical│─────▶│ OpenAI  │
│  Tool   │      │  Tool   │      │  Tool   │
└─────────┘      └─────────┘      └─────────┘
      ▲                                │
      │          ToCanonical()         │
      │          FromCanonical()       │
      │                                ▼
┌─────────┐                      ┌─────────┐
│Anthropic│◀─────────────────────│ OpenAI  │
│  Tool   │                      │  Tool   │
└─────────┘                      └─────────┘
```

All conversions pass through `CanonicalTool`, enabling N adapters to support
N² conversions with only 2N implementations.

### Feature Loss Detection

```
┌─────────────────┐
│  Source Schema  │
│  (full features)│
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│ Target Adapter  │────▶│FeatureLossWarning│
│SupportsFeature()│     │ - Feature       │
└─────────────────┘     │ - Path          │
                        │ - Message       │
                        └─────────────────┘
```

## Design Principles

1. **Pure Transformations**: Conversions have no I/O or side effects
2. **Minimal Dependencies**: Only essential external packages
3. **MCP Alignment**: Tool type embeds official MCP SDK types
4. **Explicit Loss**: Feature degradation is warnings, not errors
5. **Thread Safety**: Registry is safe for concurrent reads after setup

## Extension Points

### Custom Adapters

Implement the `Adapter` interface:

```go
type Adapter interface {
    Name() string
    ToCanonical(raw any) (*CanonicalTool, error)
    FromCanonical(ct *CanonicalTool) (any, error)
    SupportsFeature(feature SchemaFeature) bool
}
```

### Custom Validators

Implement the `SchemaValidator` interface:

```go
type SchemaValidator interface {
    Validate(schema any, data any) error
    ValidateInput(tool *Tool, input any) error
    ValidateOutput(tool *Tool, output any) error
}
```
