package adapter

import (
	"errors"
	"fmt"
)

// OpenAIFunction represents the OpenAI function calling format.
// Based on OpenAI API spec - defined locally to avoid SDK coupling.
type OpenAIFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters"`
	Strict      *bool          `json:"strict,omitempty"`
}

// OpenAITool wraps a function for the tools array format.
type OpenAITool struct {
	Type     string         `json:"type"` // always "function"
	Function OpenAIFunction `json:"function"`
}

// OpenAIAdapter converts between OpenAI function format and CanonicalTool.
type OpenAIAdapter struct{}

// NewOpenAIAdapter creates a new OpenAI adapter.
func NewOpenAIAdapter() *OpenAIAdapter {
	return &OpenAIAdapter{}
}

// Name returns the adapter's identifier.
func (a *OpenAIAdapter) Name() string {
	return "openai"
}

// openAIFeatures defines which JSON Schema features OpenAI supports.
var openAIFeatures = map[SchemaFeature]bool{
	// Supported features
	FeatureEnum:                 true,
	FeatureDefault:              true,
	FeatureAdditionalProperties: true,
	FeatureMinimum:              true,
	FeatureMaximum:              true,
	FeatureMinLength:            true,
	FeatureMaxLength:            true,
	FeatureMultipleOf:           true,
	FeatureMinItems:             true,
	FeatureMaxItems:             true,
	FeatureMinProperties:        true,
	FeatureMaxProperties:        true,
	FeatureUniqueItems:          true,
	FeatureConst:                true,

	// NOT supported features
	FeatureRef:     false,
	FeatureDefs:    false,
	FeatureAnyOf:   false,
	FeatureOneOf:   false,
	FeatureAllOf:   false,
	FeatureNot:     false,
	FeaturePattern: false,
	FeatureFormat:  false,
	FeatureTitle:   false,
	FeatureExamples: false,
	FeatureNullable: false,
	FeatureDeprecated: false,
	FeatureReadOnly: false,
	FeatureWriteOnly: false,
}

// ToCanonical converts an OpenAI tool to the canonical format.
// Accepts *OpenAITool, OpenAITool, *OpenAIFunction, or OpenAIFunction.
func (a *OpenAIAdapter) ToCanonical(raw any) (*CanonicalTool, error) {
	if raw == nil {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical",
			Cause:     errors.New("input is nil"),
		}
	}

	var fn *OpenAIFunction

	switch v := raw.(type) {
	case *OpenAITool:
		fn = &v.Function
	case OpenAITool:
		fn = &v.Function
	case *OpenAIFunction:
		fn = v
	case OpenAIFunction:
		fn = &v
	default:
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical",
			Cause:     fmt.Errorf("unsupported type: %T", raw),
		}
	}

	if fn.Name == "" {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical",
			Cause:     errors.New("function name is required"),
		}
	}

	// Convert Parameters to JSONSchema
	inputSchema := schemaFromMap(fn.Parameters)

	ct := &CanonicalTool{
		Name:         fn.Name,
		Description:  fn.Description,
		InputSchema:  inputSchema,
		SourceFormat: "openai",
		SourceMeta:   make(map[string]any),
	}

	// Preserve OpenAI-specific fields in SourceMeta for round-trip
	if fn.Strict != nil {
		ct.SourceMeta["strict"] = *fn.Strict
	}

	return ct, nil
}

// FromCanonical converts a canonical tool to OpenAI format.
// Returns *OpenAITool.
func (a *OpenAIAdapter) FromCanonical(ct *CanonicalTool) (any, error) {
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

	fn := OpenAIFunction{
		Name:        ct.Name,
		Description: canonicalDescription(ct),
	}

	// Convert InputSchema to parameters map, filtering unsupported features
	if ct.InputSchema != nil {
		fn.Parameters = filterOpenAISchema(ct.InputSchema).ToMap()
	} else {
		fn.Parameters = map[string]any{"type": "object"}
	}

	// Restore strict from SourceMeta
	if ct.SourceMeta != nil {
		if strict, ok := ct.SourceMeta["strict"].(bool); ok {
			fn.Strict = &strict
		}
	}

	return &OpenAITool{
		Type:     "function",
		Function: fn,
	}, nil
}

// SupportsFeature returns whether this adapter supports a schema feature.
func (a *OpenAIAdapter) SupportsFeature(feature SchemaFeature) bool {
	supported, ok := openAIFeatures[feature]
	return ok && supported
}

// filterOpenAISchema removes unsupported features from a schema for OpenAI.
func filterOpenAISchema(schema *JSONSchema) *JSONSchema {
	if schema == nil {
		return nil
	}

	filtered := &JSONSchema{
		Type:        schema.Type,
		Description: schema.Description,
		Default:     schema.Default,
		Const:       schema.Const,
	}

	// Copy supported pointer fields
	if schema.MultipleOf != nil {
		v := *schema.MultipleOf
		filtered.MultipleOf = &v
	}
	if schema.Minimum != nil {
		v := *schema.Minimum
		filtered.Minimum = &v
	}
	if schema.Maximum != nil {
		v := *schema.Maximum
		filtered.Maximum = &v
	}
	if schema.MinLength != nil {
		v := *schema.MinLength
		filtered.MinLength = &v
	}
	if schema.MaxLength != nil {
		v := *schema.MaxLength
		filtered.MaxLength = &v
	}
	if schema.MinItems != nil {
		v := *schema.MinItems
		filtered.MinItems = &v
	}
	if schema.MaxItems != nil {
		v := *schema.MaxItems
		filtered.MaxItems = &v
	}
	if schema.MinProperties != nil {
		v := *schema.MinProperties
		filtered.MinProperties = &v
	}
	if schema.MaxProperties != nil {
		v := *schema.MaxProperties
		filtered.MaxProperties = &v
	}
	if schema.UniqueItems != nil {
		v := *schema.UniqueItems
		filtered.UniqueItems = &v
	}
	if schema.AdditionalProperties != nil {
		v := *schema.AdditionalProperties
		filtered.AdditionalProperties = &v
	}

	// Copy supported slices
	if schema.Required != nil {
		filtered.Required = make([]string, len(schema.Required))
		copy(filtered.Required, schema.Required)
	}
	if schema.Enum != nil {
		filtered.Enum = make([]any, len(schema.Enum))
		copy(filtered.Enum, schema.Enum)
	}

	// Recursively filter properties
	if schema.Properties != nil {
		filtered.Properties = make(map[string]*JSONSchema, len(schema.Properties))
		for k, v := range schema.Properties {
			filtered.Properties[k] = filterOpenAISchema(v)
		}
	}

	// Recursively filter items
	if schema.Items != nil {
		filtered.Items = filterOpenAISchema(schema.Items)
	}

	// Note: Explicitly NOT copying unsupported fields:
	// - Ref, Defs ($ref, $defs)
	// - AnyOf, OneOf, AllOf, Not (combinators)
	// - Pattern, Format (string validation)

	return filtered
}
