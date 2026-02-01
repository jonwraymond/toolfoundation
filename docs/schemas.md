# Tool Schemas

toolfoundation defines the canonical tool schema (the `Tool` record) and the
JSON Schemas used for tool input/output validation. This page documents the
fields, constraints, and JSON Schema rules that all downstream components rely on.

## Tool schema fields and constraints

The canonical tool record is `model.Tool`, which embeds the MCP SDK `mcp.Tool`
fields and adds stack-specific extensions.

### Core MCP tool fields

| Field | Required | Constraints / Notes |
|-------|----------|---------------------|
| `name` | Yes | 1-128 chars, allowed: `[A-Za-z0-9_.-]` only. | 
| `description` | No | Human-readable description. | 
| `inputSchema` | Yes | JSON Schema object defining tool parameters. | 
| `outputSchema` | No | JSON Schema object for structured output (optional). | 
| `title` | No | Display name (preferred over `name`). | 
| `annotations` | No | Hints for clients (readOnly, idempotent, destructive, openWorld). | 
| `_meta` | No | Arbitrary metadata. | 
| `icons` | No | List of icon assets for UI. | 

### toolfoundation extensions

| Field | Required | Constraints / Notes |
|-------|----------|---------------------|
| `namespace` | No | Optional namespace for stable IDs. If present, tool ID is `namespace:name`. Namespace and name must both be non-empty when used. | 
| `version` | No | Optional semantic version string (accepts `v1.2.3` or `1.2.3`). | 
| `tags` | No | Normalized tags for discovery; see rules below. | 

### Tool ID rules

- Tool IDs are `namespace:name` when `namespace` is set, otherwise just `name`.
- Only one `:` is permitted in an ID.
- `namespace` and `name` must both be non-empty when a `:` is used.

### Tag normalization rules

Tags are normalized for stable search behavior:

- Lowercased and trimmed.
- Whitespace collapsed to `-`.
- Allowed characters: `[a-z0-9-_.]` (others are removed).
- Max tag length: 64 chars.
- Max tag count: 20.
- Duplicates removed while preserving order.

## InputSchema / OutputSchema requirements

- **InputSchema is required.** A tool without `inputSchema` is invalid.
- **OutputSchema is optional.** If omitted, output validation is skipped.
- Schemas must be valid JSON Schema objects. Accepted representations:
  - `map[string]any`
  - `json.RawMessage`
  - `[]byte`
  - `*jsonschema.Schema`
- Validation is performed by `model.SchemaValidator`:
  - `ValidateInput` must use `tool.InputSchema` and returns `ErrInvalidSchema`
    if `tool` or `InputSchema` is nil.
  - `ValidateOutput` returns nil when `OutputSchema` is nil.

## Supported dialects and limitations

- Default dialect: **JSON Schema 2020-12** (assumed when `$schema` is absent).
- Supported: **2020-12** and **draft-07**.
- External `$ref` resolution is **disabled** (no network access).
- Known limitations from the underlying validator:
  - `format` is treated as annotation (not enforced).
  - `contentEncoding` and `contentMediaType` are not validated.

## Recommended "no parameters" schema

The MCP-recommended schema for tools that take no parameters:

```json
{
  "type": "object",
  "additionalProperties": false
}
```

The MCP-allowed (but less strict) variant:

```json
{
  "type": "object"
}
```

## Example schema patterns

### Required string property

```json
{
  "type": "object",
  "properties": {
    "path": {"type": "string", "description": "File path"}
  },
  "required": ["path"],
  "additionalProperties": false
}
```

### Optional enum with default

```json
{
  "type": "object",
  "properties": {
    "encoding": {"type": "string", "enum": ["utf8", "ascii"], "default": "utf8"}
  },
  "additionalProperties": false
}
```

### Array of objects

```json
{
  "type": "object",
  "properties": {
    "items": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": {"type": "string"},
          "value": {"type": "number"}
        },
        "required": ["id"],
        "additionalProperties": false
      }
    }
  },
  "additionalProperties": false
}
```

### One-of variants

```json
{
  "type": "object",
  "properties": {
    "mode": {
      "oneOf": [
        {"type": "string", "enum": ["fast", "safe"]},
        {"type": "number", "minimum": 1, "maximum": 10}
      ]
    }
  }
}
```

> Note: some adapters (e.g., OpenAI) do not support all schema features. See
> the adapter feature matrix in the toolfoundation component docs for details.

## Authoring approaches (recommended)

1. **Go structs + schema generation** (recommended default)
   - Define input/output types in Go and generate JSON Schema.
   - Example generator: `github.com/invopop/jsonschema`.
2. **Schema builder helpers**
   - Useful for advanced constructs not easily expressed in tags.
3. **Raw map/JSON schema**
   - Fully supported, but most error-prone.

## Links

- [Design notes](design-notes.md)
- [User journey](user-journey.md)
