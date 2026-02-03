package adapter

import (
	"encoding/json"
	"errors"
	"fmt"
)

// AnthropicTool represents the Anthropic tool format.
// Based on Anthropic API spec - defined locally to avoid SDK coupling.
type AnthropicTool struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	InputSchema   map[string]any         `json:"input_schema"`
	InputExamples []any                  `json:"input_examples,omitempty"`
	CacheControl  *AnthropicCacheControl `json:"cache_control,omitempty"`
}

// AnthropicCacheControl for prompt caching.
type AnthropicCacheControl struct {
	Type string `json:"type"` // "ephemeral"
}

// AnthropicAdapter converts between Anthropic tool format and CanonicalTool.
type AnthropicAdapter struct{}

// NewAnthropicAdapter creates a new Anthropic adapter.
func NewAnthropicAdapter() *AnthropicAdapter {
	return &AnthropicAdapter{}
}

// Name returns the adapter's identifier.
func (a *AnthropicAdapter) Name() string {
	return "anthropic"
}

// anthropicFeatures defines which JSON Schema features Anthropic supports.
var anthropicFeatures = map[SchemaFeature]bool{
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
	FeatureAnyOf:                true, // Anthropic supports anyOf

	// NOT supported features
	FeatureRef:        false,
	FeatureDefs:       false,
	FeatureOneOf:      false, // Limited support
	FeatureAllOf:      false, // Limited support
	FeatureNot:        false,
	FeaturePattern:    false,
	FeatureFormat:     false,
	FeatureTitle:      false,
	FeatureExamples:   false,
	FeatureNullable:   false,
	FeatureDeprecated: false,
	FeatureReadOnly:   false,
	FeatureWriteOnly:  false,
}

// ToCanonical converts an Anthropic tool to the canonical format.
// Accepts *AnthropicTool or AnthropicTool.
func (a *AnthropicAdapter) ToCanonical(raw any) (*CanonicalTool, error) {
	if raw == nil {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical",
			Cause:     errors.New("input is nil"),
		}
	}

	var tool *AnthropicTool

	switch v := raw.(type) {
	case *AnthropicTool:
		tool = v
	case AnthropicTool:
		tool = &v
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
	inputSchema := schemaFromMap(tool.InputSchema)

	ct := &CanonicalTool{
		Name:         tool.Name,
		Description:  tool.Description,
		InputSchema:  inputSchema,
		SourceFormat: "anthropic",
		SourceMeta:   make(map[string]any),
	}

	// Preserve Anthropic-specific fields in SourceMeta for round-trip
	if tool.CacheControl != nil {
		ct.SourceMeta["cache_control"] = tool.CacheControl
	}
	if len(tool.InputExamples) > 0 {
		ct.SourceMeta["input_examples"] = tool.InputExamples
		// Best-effort: convert examples to strings for canonical Examples.
		if len(ct.Examples) == 0 {
			for _, example := range tool.InputExamples {
				if data, err := json.Marshal(example); err == nil {
					ct.Examples = append(ct.Examples, string(data))
				}
			}
		}
	}

	return ct, nil
}

// FromCanonical converts a canonical tool to Anthropic format.
// Returns *AnthropicTool.
func (a *AnthropicAdapter) FromCanonical(ct *CanonicalTool) (any, error) {
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

	tool := &AnthropicTool{
		Name:        ct.Name,
		Description: canonicalDescription(ct),
	}

	// Convert InputSchema to input_schema map, filtering unsupported features
	if ct.InputSchema != nil {
		tool.InputSchema = filterAnthropicSchema(ct.InputSchema).ToMap()
	} else {
		tool.InputSchema = map[string]any{"type": "object"}
	}

	// Restore cache_control from SourceMeta
	if ct.SourceMeta != nil {
		if cc, ok := ct.SourceMeta["cache_control"].(*AnthropicCacheControl); ok {
			tool.CacheControl = cc
		}
		if rawExamples, ok := ct.SourceMeta["input_examples"]; ok {
			switch v := rawExamples.(type) {
			case []any:
				tool.InputExamples = v
			case []map[string]any:
				tool.InputExamples = make([]any, 0, len(v))
				for _, ex := range v {
					tool.InputExamples = append(tool.InputExamples, ex)
				}
			}
		}
	}

	return tool, nil
}

// SupportsFeature returns whether this adapter supports a schema feature.
func (a *AnthropicAdapter) SupportsFeature(feature SchemaFeature) bool {
	supported, ok := anthropicFeatures[feature]
	return ok && supported
}

// filterAnthropicSchema removes unsupported features from a schema for Anthropic.
func filterAnthropicSchema(schema *JSONSchema) *JSONSchema {
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
			filtered.Properties[k] = filterAnthropicSchema(v)
		}
	}

	// Recursively filter items
	if schema.Items != nil {
		filtered.Items = filterAnthropicSchema(schema.Items)
	}

	// Anthropic supports anyOf (unlike OpenAI)
	if schema.AnyOf != nil {
		filtered.AnyOf = make([]*JSONSchema, len(schema.AnyOf))
		for i, v := range schema.AnyOf {
			filtered.AnyOf[i] = filterAnthropicSchema(v)
		}
	}

	// Note: Explicitly NOT copying unsupported fields:
	// - Ref, Defs ($ref, $defs)
	// - OneOf, AllOf, Not (limited support)
	// - Pattern, Format (string validation)

	return filtered
}
