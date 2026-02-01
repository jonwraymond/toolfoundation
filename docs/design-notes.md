# toolfoundation Design Notes

## Overview

toolfoundation provides the canonical data types that all other ApertureStack
components depend on. It contains two packages: `model` for tool definitions
and `adapter` for format conversion.

## model Package

### Design Decisions

1. **MCP SDK Embedding**: The `Tool` type embeds `mcp.Tool` from the official
   MCP Go SDK rather than reimplementing the fields. This ensures 1:1 spec
   compatibility while allowing extension.

2. **Namespace + Name = ID**: Tool IDs are `namespace:name` format, providing
   stable identifiers across registry operations.

3. **Backend Abstraction**: `ToolBackend` supports three kinds:
   - `local`: In-process handler function
   - `provider`: External tool provider
   - `mcp`: Remote MCP server

4. **Tag Normalization**: Tags are normalized (lowercase, trimmed, deduped)
   to ensure consistent search behavior.

### Error Handling

- `Validate()` returns descriptive errors for invalid tools
- Schema validation uses JSON Schema draft 2020-12
- Empty names, invalid characters, and missing schemas are rejected

### Schema Validation Policy

The default validator supports the following dialects:

- JSON Schema 2020-12 (default)
- JSON Schema draft-07

External `$ref` resolution is **disabled** to prevent network access during
validation. Validation behavior is deterministic and does not perform I/O.

Limitations (from the underlying jsonschema-go implementation):

- `format` is treated as annotation (not validated)
- `contentEncoding` and `contentMediaType` are not validated

## adapter Package

### Design Decisions

1. **Canonical Intermediate**: All conversions go through `CanonicalTool`,
   a protocol-agnostic intermediate representation.

2. **Pure Transforms**: Conversions have no I/O or side effects. Same input
   always produces same output.

3. **Feature Loss Warnings**: When the target format doesn't support a feature
   (e.g., `$ref` in OpenAI), warnings are returned instead of errors.

4. **Minimal Dependencies**: The MCP adapter depends on the MCP SDK. OpenAI
   and Anthropic adapters use self-contained types.

### Supported Formats

| Format | Adapter | Notes |
|--------|---------|-------|
| MCP | MCPAdapter | Full spec support |
| OpenAI | OpenAIAdapter | Strict mode support |
| Anthropic | AnthropicAdapter | Full spec support |

### Feature Compatibility

| Feature | MCP | OpenAI | Anthropic |
|---------|:---:|:------:|:---------:|
| `$ref/$defs` | Yes | No | No |
| `anyOf/oneOf` | Yes | No | Yes |
| `pattern` | Yes | Yes* | Yes |
| `enum/const` | Yes | Yes | Yes |

*OpenAI supports pattern only in strict mode.

### Feature Loss Warnings

Adapters emit `FeatureLossWarning` entries when the target format does not
support a schema feature used by the source. Conversions still succeed, but
consumers should review warnings before exposing the converted tool to users.

## Dependencies

- `github.com/modelcontextprotocol/go-sdk/mcp` - MCP type definitions
- `github.com/santhosh-tekuri/jsonschema` - Schema validation (optional)

## Links

- [index](index.md)
- [tool schemas](schemas.md)
- [user journey](user-journey.md)
