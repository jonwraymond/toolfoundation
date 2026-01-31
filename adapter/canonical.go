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

	// Version is the semantic version of the tool
	Version string

	// Description explains what the tool does
	Description string

	// Category classifies the tool's purpose
	Category string

	// Tags are keywords for discovery
	Tags []string

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

	// Minimum is the minimum numeric value
	Minimum *float64

	// Maximum is the maximum numeric value
	Maximum *float64

	// MinLength is the minimum string length
	MinLength *int

	// MaxLength is the maximum string length
	MaxLength *int

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
		Description: s.Description,
		Const:       s.Const,
		Default:     s.Default,
		Pattern:     s.Pattern,
		Format:      s.Format,
		Ref:         s.Ref,
	}

	// Deep copy pointer fields
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
	if s.AdditionalProperties != nil {
		v := *s.AdditionalProperties
		copied.AdditionalProperties = &v
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

	// Pointer fields
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
	if s.AdditionalProperties != nil {
		m["additionalProperties"] = *s.AdditionalProperties
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
