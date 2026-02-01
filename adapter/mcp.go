package adapter

import (
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jonwraymond/toolfoundation/model"
)

// MCPAdapter converts between model.Tool and CanonicalTool.
// MCP supports all JSON Schema 2020-12 features.
type MCPAdapter struct{}

// NewMCPAdapter creates a new MCP adapter.
func NewMCPAdapter() *MCPAdapter {
	return &MCPAdapter{}
}

// Name returns the adapter's identifier.
func (a *MCPAdapter) Name() string {
	return "mcp"
}

// ToCanonical converts an MCP tool to the canonical format.
// Accepts *model.Tool, model.Tool, *mcp.Tool, or mcp.Tool.
func (a *MCPAdapter) ToCanonical(raw any) (*CanonicalTool, error) {
	if raw == nil {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical",
			Cause:     errors.New("input is nil"),
		}
	}

	var tool *model.Tool

	switch v := raw.(type) {
	case *model.Tool:
		tool = v
	case model.Tool:
		tool = &v
	case *mcp.Tool:
		tool = &model.Tool{Tool: *v}
	case mcp.Tool:
		tool = &model.Tool{Tool: v}
	default:
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical",
			Cause:     fmt.Errorf("unsupported type: %T", raw),
		}
	}

	if tool.Name == "" {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical",
			Cause:     errors.New("tool name is required"),
		}
	}

	// Convert InputSchema to JSONSchema
	inputSchema, err := schemaFromAny(tool.InputSchema)
	if err != nil {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical",
			Cause:     fmt.Errorf("invalid input schema: %w", err),
		}
	}

	// Convert OutputSchema if present
	var outputSchema *JSONSchema
	if tool.OutputSchema != nil {
		outputSchema, err = schemaFromAny(tool.OutputSchema)
		if err != nil {
			return nil, &ConversionError{
				Adapter:   a.Name(),
				Direction: "to_canonical",
				Cause:     fmt.Errorf("invalid output schema: %w", err),
			}
		}
	}

	ct := &CanonicalTool{
		Namespace:    tool.Namespace,
		Name:         tool.Name,
		Version:      tool.Version,
		Description:  tool.Description,
		Tags:         tool.Tags,
		InputSchema:  inputSchema,
		OutputSchema: outputSchema,
		SourceFormat: "mcp",
		SourceMeta:   make(map[string]any),
	}

	// Preserve MCP-specific fields in SourceMeta for round-trip
	if tool.Title != "" {
		ct.SourceMeta["title"] = tool.Title
	}
	if tool.Meta != nil {
		ct.SourceMeta["meta"] = tool.Meta
	}
	if tool.Annotations != nil {
		ct.SourceMeta["annotations"] = tool.Annotations
	}
	if len(tool.Icons) > 0 {
		ct.SourceMeta["icons"] = tool.Icons
	}

	return ct, nil
}

// FromCanonical converts a canonical tool to model.Tool.
func (a *MCPAdapter) FromCanonical(ct *CanonicalTool) (any, error) {
	if ct == nil {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "from_canonical",
			Cause:     errors.New("canonical tool is nil"),
		}
	}

	if ct.Name == "" {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "from_canonical",
			Cause:     errors.New("tool name is required"),
		}
	}

	tool := &model.Tool{
		Tool: mcp.Tool{
			Name:        ct.Name,
			Description: ct.Description,
		},
		Namespace: ct.Namespace,
		Version:   ct.Version,
		Tags:      ct.Tags,
	}

	// Convert InputSchema
	if ct.InputSchema != nil {
		tool.InputSchema = ct.InputSchema.ToMap()
	}

	// Convert OutputSchema
	if ct.OutputSchema != nil {
		tool.OutputSchema = ct.OutputSchema.ToMap()
	}

	// Restore MCP-specific fields from SourceMeta
	if ct.SourceMeta != nil {
		if title, ok := ct.SourceMeta["title"].(string); ok {
			tool.Title = title
		}
		if meta, ok := ct.SourceMeta["meta"].(mcp.Meta); ok {
			tool.Meta = meta
		} else if metaMap, ok := ct.SourceMeta["meta"].(map[string]any); ok {
			tool.Meta = mcp.Meta(metaMap)
		}
		if annotations, ok := ct.SourceMeta["annotations"].(*mcp.ToolAnnotations); ok {
			tool.Annotations = annotations
		}
		if icons, ok := ct.SourceMeta["icons"].([]mcp.Icon); ok {
			tool.Icons = icons
		}
	}

	return tool, nil
}

// SupportsFeature returns whether this adapter supports a schema feature.
// MCP supports all JSON Schema 2020-12 features.
func (a *MCPAdapter) SupportsFeature(feature SchemaFeature) bool {
	// MCP supports the full JSON Schema 2020-12 spec
	return true
}

// schemaFromAny converts any schema representation to *JSONSchema.
// Accepts map[string]any, *JSONSchema, or JSONSchema.
func schemaFromAny(schema any) (*JSONSchema, error) {
	if schema == nil {
		return nil, nil
	}

	switch v := schema.(type) {
	case *JSONSchema:
		return v.DeepCopy(), nil
	case JSONSchema:
		return v.DeepCopy(), nil
	case map[string]any:
		return schemaFromMap(v), nil
	default:
		return nil, fmt.Errorf("unsupported schema type: %T", schema)
	}
}

// schemaFromMap converts a map[string]any to *JSONSchema.
func schemaFromMap(m map[string]any) *JSONSchema {
	if m == nil {
		return nil
	}

	s := &JSONSchema{}

	// String fields
	if v, ok := m["type"].(string); ok {
		s.Type = v
	}
	if v, ok := m["description"].(string); ok {
		s.Description = v
	}
	if v, ok := m["pattern"].(string); ok {
		s.Pattern = v
	}
	if v, ok := m["format"].(string); ok {
		s.Format = v
	}
	if v, ok := m["$ref"].(string); ok {
		s.Ref = v
	}

	// Any fields
	if v, ok := m["const"]; ok {
		s.Const = v
	}
	if v, ok := m["default"]; ok {
		s.Default = v
	}

	// Numeric fields
	if v, ok := m["minimum"].(float64); ok {
		s.Minimum = &v
	}
	if v, ok := m["maximum"].(float64); ok {
		s.Maximum = &v
	}
	if v, ok := m["minLength"]; ok {
		if f, ok := v.(float64); ok {
			i := int(f)
			s.MinLength = &i
		}
	}
	if v, ok := m["maxLength"]; ok {
		if f, ok := v.(float64); ok {
			i := int(f)
			s.MaxLength = &i
		}
	}

	// Boolean fields
	if v, ok := m["additionalProperties"].(bool); ok {
		s.AdditionalProperties = &v
	}

	// Array fields
	if v, ok := m["required"].([]any); ok {
		s.Required = make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				s.Required = append(s.Required, str)
			}
		}
	}
	if v, ok := m["required"].([]string); ok {
		s.Required = v
	}
	if v, ok := m["enum"].([]any); ok {
		s.Enum = v
	}

	// Object fields
	if v, ok := m["properties"].(map[string]any); ok {
		s.Properties = make(map[string]*JSONSchema, len(v))
		for k, prop := range v {
			if propMap, ok := prop.(map[string]any); ok {
				s.Properties[k] = schemaFromMap(propMap)
			}
		}
	}
	if v, ok := m["$defs"].(map[string]any); ok {
		s.Defs = make(map[string]*JSONSchema, len(v))
		for k, def := range v {
			if defMap, ok := def.(map[string]any); ok {
				s.Defs[k] = schemaFromMap(defMap)
			}
		}
	}

	// Items
	if v, ok := m["items"].(map[string]any); ok {
		s.Items = schemaFromMap(v)
	}

	// Combinators
	if v, ok := m["anyOf"].([]any); ok {
		s.AnyOf = make([]*JSONSchema, 0, len(v))
		for _, item := range v {
			if itemMap, ok := item.(map[string]any); ok {
				s.AnyOf = append(s.AnyOf, schemaFromMap(itemMap))
			}
		}
	}
	if v, ok := m["oneOf"].([]any); ok {
		s.OneOf = make([]*JSONSchema, 0, len(v))
		for _, item := range v {
			if itemMap, ok := item.(map[string]any); ok {
				s.OneOf = append(s.OneOf, schemaFromMap(itemMap))
			}
		}
	}
	if v, ok := m["allOf"].([]any); ok {
		s.AllOf = make([]*JSONSchema, 0, len(v))
		for _, item := range v {
			if itemMap, ok := item.(map[string]any); ok {
				s.AllOf = append(s.AllOf, schemaFromMap(itemMap))
			}
		}
	}
	if v, ok := m["not"].(map[string]any); ok {
		s.Not = schemaFromMap(v)
	}

	return s
}
