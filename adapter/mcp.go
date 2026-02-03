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
		DisplayName:  tool.Title,
		Version:      tool.Version,
		Description:  tool.Description,
		Tags:         tool.Tags,
		InputSchema:  inputSchema,
		OutputSchema: outputSchema,
		SourceFormat: "mcp",
		SourceMeta:   make(map[string]any),
	}

	if ct.DisplayName == "" && tool.Annotations != nil && tool.Annotations.Title != "" {
		ct.DisplayName = tool.Annotations.Title
	}

	if tool.Annotations != nil {
		ct.Annotations = annotationsToMap(tool.Annotations)
		idempotent := tool.Annotations.IdempotentHint
		ct.Idempotent = &idempotent
	}

	if tool.Meta != nil {
		if summary, ok := tool.Meta["summary"].(string); ok {
			ct.Summary = summary
		}
		if category, ok := tool.Meta["category"].(string); ok {
			ct.Category = category
		}
		if modes := stringSliceFromAny(tool.Meta["inputModes"]); len(modes) > 0 {
			ct.InputModes = modes
		}
		if modes := stringSliceFromAny(tool.Meta["outputModes"]); len(modes) > 0 {
			ct.OutputModes = modes
		}
		if examples := stringSliceFromAny(tool.Meta["examples"]); len(examples) > 0 {
			ct.Examples = examples
		}
		if deterministic, ok := tool.Meta["deterministic"].(bool); ok {
			ct.Deterministic = &deterministic
		}
		if streaming, ok := tool.Meta["streaming"].(bool); ok {
			ct.Streaming = &streaming
		}
		if schemes := securitySchemesFromAny(tool.Meta["securitySchemes"]); len(schemes) > 0 {
			ct.SecuritySchemes = schemes
		}
		if requirements := securityRequirementsFromAny(tool.Meta["securityRequirements"]); len(requirements) > 0 {
			ct.SecurityRequirements = requirements
		}
		if hints, ok := tool.Meta["uiHints"].(map[string]any); ok && len(hints) > 0 {
			ct.UIHints = hints
		}
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
			tool.Annotations = cloneAnnotations(annotations)
		}
		if icons, ok := ct.SourceMeta["icons"].([]mcp.Icon); ok {
			tool.Icons = icons
		}
	}

	if ct.DisplayName != "" {
		tool.Title = ct.DisplayName
	}

	annotations, annotationsSet := annotationsFromCanonical(ct, tool.Annotations)
	if annotationsSet {
		tool.Annotations = annotations
	}

	metaSet := false
	if tool.Meta == nil {
		tool.Meta = mcp.Meta{}
	}
	if ct.Summary != "" {
		tool.Meta["summary"] = ct.Summary
		metaSet = true
	}
	if ct.Category != "" {
		tool.Meta["category"] = ct.Category
		metaSet = true
	}
	if len(ct.InputModes) > 0 {
		tool.Meta["inputModes"] = ct.InputModes
		metaSet = true
	}
	if len(ct.OutputModes) > 0 {
		tool.Meta["outputModes"] = ct.OutputModes
		metaSet = true
	}
	if len(ct.Examples) > 0 {
		tool.Meta["examples"] = ct.Examples
		metaSet = true
	}
	if ct.Deterministic != nil {
		tool.Meta["deterministic"] = *ct.Deterministic
		metaSet = true
	}
	if ct.Streaming != nil {
		tool.Meta["streaming"] = *ct.Streaming
		metaSet = true
	}
	if len(ct.SecuritySchemes) > 0 {
		tool.Meta["securitySchemes"] = ct.SecuritySchemes
		metaSet = true
	}
	if len(ct.SecurityRequirements) > 0 {
		tool.Meta["securityRequirements"] = ct.SecurityRequirements
		metaSet = true
	}
	if len(ct.UIHints) > 0 {
		tool.Meta["uiHints"] = ct.UIHints
		metaSet = true
	}
	if !metaSet && len(tool.Meta) == 0 {
		tool.Meta = nil
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
	if v, ok := m["title"].(string); ok {
		s.Title = v
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
	if v, ok := m["examples"].([]any); ok {
		s.Examples = v
	}
	if v, ok := m["examples"].([]string); ok {
		s.Examples = make([]any, len(v))
		for i, item := range v {
			s.Examples[i] = item
		}
	}

	// Numeric fields
	if v, ok := asFloat(m["multipleOf"]); ok {
		s.MultipleOf = &v
	}
	if v, ok := asFloat(m["minimum"]); ok {
		s.Minimum = &v
	}
	if v, ok := asFloat(m["maximum"]); ok {
		s.Maximum = &v
	}
	if v, ok := m["minLength"]; ok {
		if i, ok := asInt(v); ok {
			s.MinLength = &i
		}
	}
	if v, ok := m["maxLength"]; ok {
		if i, ok := asInt(v); ok {
			s.MaxLength = &i
		}
	}
	if v, ok := m["minItems"]; ok {
		if i, ok := asInt(v); ok {
			s.MinItems = &i
		}
	}
	if v, ok := m["maxItems"]; ok {
		if i, ok := asInt(v); ok {
			s.MaxItems = &i
		}
	}
	if v, ok := m["minProperties"]; ok {
		if i, ok := asInt(v); ok {
			s.MinProperties = &i
		}
	}
	if v, ok := m["maxProperties"]; ok {
		if i, ok := asInt(v); ok {
			s.MaxProperties = &i
		}
	}

	// Boolean fields
	if v, ok := m["additionalProperties"].(bool); ok {
		s.AdditionalProperties = &v
	}
	if v, ok := m["uniqueItems"].(bool); ok {
		s.UniqueItems = &v
	}
	if v, ok := m["nullable"].(bool); ok {
		s.Nullable = &v
	}
	if v, ok := m["deprecated"].(bool); ok {
		s.Deprecated = &v
	}
	if v, ok := m["readOnly"].(bool); ok {
		s.ReadOnly = &v
	}
	if v, ok := m["writeOnly"].(bool); ok {
		s.WriteOnly = &v
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
	if v, ok := m["enum"].([]string); ok {
		s.Enum = make([]any, len(v))
		for i, item := range v {
			s.Enum[i] = item
		}
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

func asFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint:
		return float64(t), true
	case uint32:
		return float64(t), true
	case uint64:
		return float64(t), true
	default:
		return 0, false
	}
}

func asInt(v any) (int, bool) {
	switch t := v.(type) {
	case int:
		return t, true
	case int32:
		return int(t), true
	case int64:
		return int(t), true
	case uint:
		return int(t), true
	case uint32:
		return int(t), true
	case uint64:
		return int(t), true
	case float64:
		return int(t), true
	case float32:
		return int(t), true
	default:
		return 0, false
	}
}

func annotationsToMap(ann *mcp.ToolAnnotations) map[string]any {
	if ann == nil {
		return nil
	}

	out := map[string]any{}
	if ann.DestructiveHint != nil {
		out["destructiveHint"] = *ann.DestructiveHint
	}
	if ann.OpenWorldHint != nil {
		out["openWorldHint"] = *ann.OpenWorldHint
	}
	out["idempotentHint"] = ann.IdempotentHint
	out["readOnlyHint"] = ann.ReadOnlyHint
	if ann.Title != "" {
		out["title"] = ann.Title
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func stringSliceFromAny(v any) []string {
	switch t := v.(type) {
	case []string:
		out := make([]string, len(t))
		copy(out, t)
		return out
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func securitySchemesFromAny(v any) map[string]SecurityScheme {
	switch t := v.(type) {
	case map[string]SecurityScheme:
		out := make(map[string]SecurityScheme, len(t))
		for k, scheme := range t {
			out[k] = scheme
		}
		return out
	case map[string]any:
		out := make(map[string]SecurityScheme, len(t))
		for k, raw := range t {
			switch scheme := raw.(type) {
			case map[string]any:
				out[k] = SecurityScheme(scheme)
			case SecurityScheme:
				out[k] = scheme
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		return nil
	}
}

func securityRequirementsFromAny(v any) []SecurityRequirement {
	switch t := v.(type) {
	case []SecurityRequirement:
		out := make([]SecurityRequirement, len(t))
		copy(out, t)
		return out
	case []any:
		out := make([]SecurityRequirement, 0, len(t))
		for _, item := range t {
			reqMap, ok := item.(map[string]any)
			if !ok {
				continue
			}
			req := SecurityRequirement{}
			for key, rawScopes := range reqMap {
				switch rawScopes.(type) {
				case nil:
					req[key] = nil
				case []any, []string:
					req[key] = stringSliceFromAny(rawScopes)
				default:
					// Unsupported shape; skip.
				}
			}
			if len(req) > 0 {
				out = append(out, req)
			}
		}
		return out
	default:
		return nil
	}
}

func annotationsFromCanonical(ct *CanonicalTool, base *mcp.ToolAnnotations) (*mcp.ToolAnnotations, bool) {
	if ct == nil && base == nil {
		return nil, false
	}

	ann := cloneAnnotations(base)
	has := hasMCPAnnotations(ann)
	if ann == nil {
		ann = &mcp.ToolAnnotations{}
	}

	if ct != nil && ct.Annotations != nil {
		if val, ok := ct.Annotations["destructiveHint"]; ok {
			if b, ok := val.(bool); ok {
				ann.DestructiveHint = &b
				has = true
			}
		}
		if val, ok := ct.Annotations["openWorldHint"]; ok {
			if b, ok := val.(bool); ok {
				ann.OpenWorldHint = &b
				has = true
			}
		}
		if val, ok := ct.Annotations["idempotentHint"]; ok {
			if b, ok := val.(bool); ok {
				ann.IdempotentHint = b
				has = true
			}
		}
		if val, ok := ct.Annotations["readOnlyHint"]; ok {
			if b, ok := val.(bool); ok {
				ann.ReadOnlyHint = b
				has = true
			}
		}
		if val, ok := ct.Annotations["title"]; ok {
			if s, ok := val.(string); ok {
				ann.Title = s
				has = true
			}
		}
	}

	if ct != nil && ct.Idempotent != nil {
		ann.IdempotentHint = *ct.Idempotent
		has = true
	}

	if !has {
		return nil, false
	}
	return ann, true
}

func cloneAnnotations(src *mcp.ToolAnnotations) *mcp.ToolAnnotations {
	if src == nil {
		return nil
	}
	out := &mcp.ToolAnnotations{
		IdempotentHint: src.IdempotentHint,
		ReadOnlyHint:   src.ReadOnlyHint,
		Title:          src.Title,
	}
	if src.DestructiveHint != nil {
		v := *src.DestructiveHint
		out.DestructiveHint = &v
	}
	if src.OpenWorldHint != nil {
		v := *src.OpenWorldHint
		out.OpenWorldHint = &v
	}
	return out
}

func hasMCPAnnotations(ann *mcp.ToolAnnotations) bool {
	if ann == nil {
		return false
	}
	if ann.DestructiveHint != nil {
		return true
	}
	if ann.OpenWorldHint != nil {
		return true
	}
	if ann.Title != "" {
		return true
	}
	if ann.IdempotentHint {
		return true
	}
	if ann.ReadOnlyHint {
		return true
	}
	return false
}
