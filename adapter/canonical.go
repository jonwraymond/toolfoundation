package adapter

import (
	"errors"
	"time"
)

// CanonicalTool is the protocol-agnostic representation of a tool definition.
// It serves as the intermediate format for converting between MCP, OpenAI,
// and Anthropic tool formats.
type CanonicalTool struct {
	// Namespace groups related tools (e.g., "github", "slack")
	Namespace string

	// Name is the tool's identifier (required)
	Name string

	// DisplayName is a human-friendly name for UI presentation.
	DisplayName string

	// Version is the semantic version of the tool
	Version string

	// Description explains what the tool does
	Description string

	// Summary is a short description for discovery results.
	Summary string

	// Category classifies the tool's purpose
	Category string

	// Tags are keywords for discovery
	Tags []string

	// InputModes lists supported input media types (e.g., "application/json").
	InputModes []string

	// OutputModes lists supported output media types.
	OutputModes []string

	// Examples provides example prompts or usage scenarios.
	Examples []string

	// Deterministic indicates whether the tool returns deterministic results.
	Deterministic *bool

	// Idempotent indicates whether the tool is idempotent.
	Idempotent *bool

	// Streaming indicates whether the tool supports streaming output.
	Streaming *bool

	// SecuritySchemes defines auth schemes required by this tool.
	SecuritySchemes map[string]SecurityScheme

	// SecurityRequirements defines required schemes/scopes for this tool.
	SecurityRequirements []SecurityRequirement

	// Annotations contains protocol-agnostic annotations for UI or policy.
	Annotations map[string]any

	// UIHints provides UI-specific hints for rendering tool inputs.
	UIHints map[string]any

	// InputSchema defines the tool's input parameters (required)
	InputSchema *JSONSchema

	// OutputSchema defines the tool's output format
	OutputSchema *JSONSchema

	// Timeout is the maximum execution time
	Timeout time.Duration

	// SourceFormat is the original format (e.g., "mcp", "openai", "anthropic")
	SourceFormat string

	// SourceMeta contains format-specific metadata for round-trip conversion
	SourceMeta map[string]any

	// RequiredScopes are authorization scopes needed to use the tool
	RequiredScopes []string
}

// SecurityScheme describes a security scheme definition.
// It uses a generic map to avoid coupling to any single spec.
type SecurityScheme map[string]any

// SecurityRequirement maps scheme names to required scopes.
type SecurityRequirement map[string][]string

// CanonicalProvider describes a tool provider (e.g., an A2A AgentCard).
type CanonicalProvider struct {
	// Name is the provider name (required).
	Name string

	// Description explains what the provider does.
	Description string

	// Version is the provider version.
	Version string

	// Capabilities lists provider capabilities (e.g., streaming, push notifications).
	Capabilities map[string]any

	// SecuritySchemes defines auth schemes supported by the provider.
	SecuritySchemes map[string]SecurityScheme

	// SecurityRequirements defines required schemes/scopes to access the provider.
	SecurityRequirements []SecurityRequirement

	// DefaultInputModes lists default input media types for all tools.
	DefaultInputModes []string

	// DefaultOutputModes lists default output media types for all tools.
	DefaultOutputModes []string

	// Skills are the tools offered by the provider.
	Skills []CanonicalTool

	// SourceFormat is the original format (e.g., "a2a").
	SourceFormat string

	// SourceMeta contains format-specific metadata for round-trip conversion.
	SourceMeta map[string]any
}

// ID returns the tool's fully qualified identifier.
// If Namespace is set, returns "namespace:name", otherwise just "name".
func (t *CanonicalTool) ID() string {
	if t.Namespace == "" {
		return t.Name
	}
	return t.Namespace + ":" + t.Name
}

// Validate checks that the tool has all required fields.
// Returns an error if Name or InputSchema is missing.
func (t *CanonicalTool) Validate() error {
	if t.Name == "" {
		return errors.New("tool name is required")
	}
	if t.InputSchema == nil {
		return errors.New("tool input schema is required")
	}
	return nil
}

// JSONSchema represents a JSON Schema definition.
// It is a superset supporting features from MCP, OpenAI, and Anthropic formats.
type JSONSchema struct {
	// Type is the JSON type (object, array, string, number, integer, boolean, null)
	Type string

	// Title is a short schema name.
	Title string

	// Properties maps property names to their schemas (for object types)
	Properties map[string]*JSONSchema

	// Required lists property names that must be present
	Required []string

	// Items is the schema for array elements
	Items *JSONSchema

	// Description explains the schema
	Description string

	// Enum restricts values to a fixed set
	Enum []any

	// Const restricts to a single value
	Const any

	// Default is the default value
	Default any

	// Examples provide sample values.
	Examples []any

	// MultipleOf is a numeric multiple constraint.
	MultipleOf *float64

	// Minimum is the minimum numeric value
	Minimum *float64

	// Maximum is the maximum numeric value
	Maximum *float64

	// MinLength is the minimum string length
	MinLength *int

	// MaxLength is the maximum string length
	MaxLength *int

	// MinItems is the minimum array length
	MinItems *int

	// MaxItems is the maximum array length
	MaxItems *int

	// MinProperties is the minimum number of properties
	MinProperties *int

	// MaxProperties is the maximum number of properties
	MaxProperties *int

	// UniqueItems requires array items to be unique
	UniqueItems *bool

	// Pattern is a regex pattern for string validation
	Pattern string

	// Format is a semantic format (e.g., "email", "uri", "date-time")
	Format string

	// Ref is a JSON Pointer reference to another schema ($ref)
	Ref string

	// Defs contains schema definitions ($defs)
	Defs map[string]*JSONSchema

	// AdditionalProperties controls whether extra properties are allowed
	AdditionalProperties *bool

	// Nullable indicates nullable values (OpenAPI-compatible).
	Nullable *bool

	// Deprecated marks the schema as deprecated.
	Deprecated *bool

	// ReadOnly indicates read-only properties.
	ReadOnly *bool

	// WriteOnly indicates write-only properties.
	WriteOnly *bool

	// AnyOf allows any of the listed schemas
	AnyOf []*JSONSchema

	// OneOf requires exactly one of the listed schemas
	OneOf []*JSONSchema

	// AllOf requires all of the listed schemas
	AllOf []*JSONSchema

	// Not disallows the specified schema
	Not *JSONSchema
}

// DeepCopy creates a deep copy of the JSONSchema.
// Returns nil if the receiver is nil.
func (s *JSONSchema) DeepCopy() *JSONSchema {
	if s == nil {
		return nil
	}

	copied := &JSONSchema{
		Type:        s.Type,
		Title:       s.Title,
		Description: s.Description,
		Const:       s.Const,
		Default:     s.Default,
		Pattern:     s.Pattern,
		Format:      s.Format,
		Ref:         s.Ref,
	}

	// Deep copy pointer fields
	if s.MultipleOf != nil {
		v := *s.MultipleOf
		copied.MultipleOf = &v
	}
	if s.Minimum != nil {
		v := *s.Minimum
		copied.Minimum = &v
	}
	if s.Maximum != nil {
		v := *s.Maximum
		copied.Maximum = &v
	}
	if s.MinLength != nil {
		v := *s.MinLength
		copied.MinLength = &v
	}
	if s.MaxLength != nil {
		v := *s.MaxLength
		copied.MaxLength = &v
	}
	if s.MinItems != nil {
		v := *s.MinItems
		copied.MinItems = &v
	}
	if s.MaxItems != nil {
		v := *s.MaxItems
		copied.MaxItems = &v
	}
	if s.MinProperties != nil {
		v := *s.MinProperties
		copied.MinProperties = &v
	}
	if s.MaxProperties != nil {
		v := *s.MaxProperties
		copied.MaxProperties = &v
	}
	if s.UniqueItems != nil {
		v := *s.UniqueItems
		copied.UniqueItems = &v
	}
	if s.AdditionalProperties != nil {
		v := *s.AdditionalProperties
		copied.AdditionalProperties = &v
	}
	if s.Nullable != nil {
		v := *s.Nullable
		copied.Nullable = &v
	}
	if s.Deprecated != nil {
		v := *s.Deprecated
		copied.Deprecated = &v
	}
	if s.ReadOnly != nil {
		v := *s.ReadOnly
		copied.ReadOnly = &v
	}
	if s.WriteOnly != nil {
		v := *s.WriteOnly
		copied.WriteOnly = &v
	}

	// Deep copy slices
	if s.Required != nil {
		copied.Required = make([]string, len(s.Required))
		copy(copied.Required, s.Required)
	}
	if s.Enum != nil {
		copied.Enum = make([]any, len(s.Enum))
		copy(copied.Enum, s.Enum)
	}
	if s.Examples != nil {
		copied.Examples = make([]any, len(s.Examples))
		copy(copied.Examples, s.Examples)
	}

	// Deep copy Properties map
	if s.Properties != nil {
		copied.Properties = make(map[string]*JSONSchema, len(s.Properties))
		for k, v := range s.Properties {
			copied.Properties[k] = v.DeepCopy()
		}
	}

	// Deep copy Defs map
	if s.Defs != nil {
		copied.Defs = make(map[string]*JSONSchema, len(s.Defs))
		for k, v := range s.Defs {
			copied.Defs[k] = v.DeepCopy()
		}
	}

	// Deep copy Items
	copied.Items = s.Items.DeepCopy()

	// Deep copy combinators
	if s.AnyOf != nil {
		copied.AnyOf = make([]*JSONSchema, len(s.AnyOf))
		for i, v := range s.AnyOf {
			copied.AnyOf[i] = v.DeepCopy()
		}
	}
	if s.OneOf != nil {
		copied.OneOf = make([]*JSONSchema, len(s.OneOf))
		for i, v := range s.OneOf {
			copied.OneOf[i] = v.DeepCopy()
		}
	}
	if s.AllOf != nil {
		copied.AllOf = make([]*JSONSchema, len(s.AllOf))
		for i, v := range s.AllOf {
			copied.AllOf[i] = v.DeepCopy()
		}
	}

	// Deep copy Not
	copied.Not = s.Not.DeepCopy()

	return copied
}

// ToMap converts the JSONSchema to a map[string]any representation.
// Zero-valued fields are omitted from the output.
func (s *JSONSchema) ToMap() map[string]any {
	if s == nil {
		return nil
	}

	m := make(map[string]any)

	// Simple string fields
	if s.Type != "" {
		m["type"] = s.Type
	}
	if s.Title != "" {
		m["title"] = s.Title
	}
	if s.Description != "" {
		m["description"] = s.Description
	}
	if s.Pattern != "" {
		m["pattern"] = s.Pattern
	}
	if s.Format != "" {
		m["format"] = s.Format
	}
	if s.Ref != "" {
		m["$ref"] = s.Ref
	}

	// Any fields
	if s.Const != nil {
		m["const"] = s.Const
	}
	if s.Default != nil {
		m["default"] = s.Default
	}
	if len(s.Examples) > 0 {
		m["examples"] = s.Examples
	}

	// Pointer fields
	if s.MultipleOf != nil {
		m["multipleOf"] = *s.MultipleOf
	}
	if s.Minimum != nil {
		m["minimum"] = *s.Minimum
	}
	if s.Maximum != nil {
		m["maximum"] = *s.Maximum
	}
	if s.MinLength != nil {
		m["minLength"] = *s.MinLength
	}
	if s.MaxLength != nil {
		m["maxLength"] = *s.MaxLength
	}
	if s.MinItems != nil {
		m["minItems"] = *s.MinItems
	}
	if s.MaxItems != nil {
		m["maxItems"] = *s.MaxItems
	}
	if s.MinProperties != nil {
		m["minProperties"] = *s.MinProperties
	}
	if s.MaxProperties != nil {
		m["maxProperties"] = *s.MaxProperties
	}
	if s.UniqueItems != nil {
		m["uniqueItems"] = *s.UniqueItems
	}
	if s.AdditionalProperties != nil {
		m["additionalProperties"] = *s.AdditionalProperties
	}
	if s.Nullable != nil {
		m["nullable"] = *s.Nullable
	}
	if s.Deprecated != nil {
		m["deprecated"] = *s.Deprecated
	}
	if s.ReadOnly != nil {
		m["readOnly"] = *s.ReadOnly
	}
	if s.WriteOnly != nil {
		m["writeOnly"] = *s.WriteOnly
	}

	// Slices
	if len(s.Required) > 0 {
		m["required"] = s.Required
	}
	if len(s.Enum) > 0 {
		m["enum"] = s.Enum
	}

	// Properties map
	if len(s.Properties) > 0 {
		props := make(map[string]any, len(s.Properties))
		for k, v := range s.Properties {
			props[k] = v.ToMap()
		}
		m["properties"] = props
	}

	// Defs map
	if len(s.Defs) > 0 {
		defs := make(map[string]any, len(s.Defs))
		for k, v := range s.Defs {
			defs[k] = v.ToMap()
		}
		m["$defs"] = defs
	}

	// Items
	if s.Items != nil {
		m["items"] = s.Items.ToMap()
	}

	// Combinators
	if len(s.AnyOf) > 0 {
		anyOf := make([]any, len(s.AnyOf))
		for i, v := range s.AnyOf {
			anyOf[i] = v.ToMap()
		}
		m["anyOf"] = anyOf
	}
	if len(s.OneOf) > 0 {
		oneOf := make([]any, len(s.OneOf))
		for i, v := range s.OneOf {
			oneOf[i] = v.ToMap()
		}
		m["oneOf"] = oneOf
	}
	if len(s.AllOf) > 0 {
		allOf := make([]any, len(s.AllOf))
		for i, v := range s.AllOf {
			allOf[i] = v.ToMap()
		}
		m["allOf"] = allOf
	}
	if s.Not != nil {
		m["not"] = s.Not.ToMap()
	}

	return m
}
