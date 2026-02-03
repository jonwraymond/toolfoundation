package adapter

import (
	"errors"
	"fmt"
)

// GeminiFunctionDeclaration represents a Gemini function declaration.
type GeminiFunctionDeclaration struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// GeminiTool wraps function declarations in the Gemini tools format.
type GeminiTool struct {
	FunctionDeclarations []GeminiFunctionDeclaration `json:"functionDeclarations,omitempty"`
}

// GeminiAdapter converts between Gemini function declarations and CanonicalTool.
type GeminiAdapter struct{}

// NewGeminiAdapter creates a new Gemini adapter.
func NewGeminiAdapter() *GeminiAdapter {
	return &GeminiAdapter{}
}

// Name returns the adapter's identifier.
func (a *GeminiAdapter) Name() string {
	return "gemini"
}

// geminiFeatures defines which JSON Schema features Gemini supports (OpenAPI subset).
var geminiFeatures = map[SchemaFeature]bool{
	FeatureRef:                  true,
	FeatureDefs:                 true,
	FeatureAnyOf:                true,
	FeaturePattern:              true,
	FeatureFormat:               true,
	FeatureAdditionalProperties: true,
	FeatureMinimum:              true,
	FeatureMaximum:              true,
	FeatureMinLength:            true,
	FeatureMaxLength:            true,
	FeatureMinItems:             true,
	FeatureMaxItems:             true,
	FeatureMinProperties:        true,
	FeatureMaxProperties:        true,
	FeatureEnum:                 true,
	FeatureDefault:              true,
	FeatureTitle:                true,
	FeatureNullable:             true,

	FeatureConst:       false,
	FeatureMultipleOf:  false,
	FeatureOneOf:       false,
	FeatureAllOf:       false,
	FeatureNot:         false,
	FeatureExamples:    false,
	FeatureUniqueItems: false,
	FeatureDeprecated:  false,
	FeatureReadOnly:    false,
	FeatureWriteOnly:   false,
}

// ToCanonical converts a Gemini function declaration to canonical format.
// Accepts *GeminiFunctionDeclaration, GeminiFunctionDeclaration, *GeminiTool, or GeminiTool.
func (a *GeminiAdapter) ToCanonical(raw any) (*CanonicalTool, error) {
	if raw == nil {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical",
			Cause:     errors.New("input is nil"),
		}
	}

	var fn *GeminiFunctionDeclaration

	switch v := raw.(type) {
	case *GeminiFunctionDeclaration:
		fn = v
	case GeminiFunctionDeclaration:
		fn = &v
	case *GeminiTool:
		if len(v.FunctionDeclarations) != 1 {
			return nil, &ConversionError{
				Adapter:   a.Name(),
				Direction: "to_canonical",
				Cause:     errors.New("gemini tool must contain exactly one function declaration"),
			}
		}
		fn = &v.FunctionDeclarations[0]
	case GeminiTool:
		if len(v.FunctionDeclarations) != 1 {
			return nil, &ConversionError{
				Adapter:   a.Name(),
				Direction: "to_canonical",
				Cause:     errors.New("gemini tool must contain exactly one function declaration"),
			}
		}
		fn = &v.FunctionDeclarations[0]
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

	inputSchema := schemaFromMap(fn.Parameters)
	if inputSchema == nil {
		inputSchema = &JSONSchema{Type: "object"}
	}

	ct := &CanonicalTool{
		Name:         fn.Name,
		Description:  fn.Description,
		InputSchema:  inputSchema,
		SourceFormat: "gemini",
		SourceMeta:   make(map[string]any),
	}

	return ct, nil
}

// FromCanonical converts a canonical tool to Gemini format.
// Returns *GeminiTool.
func (a *GeminiAdapter) FromCanonical(ct *CanonicalTool) (any, error) {
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

	fn := GeminiFunctionDeclaration{
		Name:        ct.Name,
		Description: ct.Description,
	}

	if ct.InputSchema != nil {
		fn.Parameters = filterGeminiSchema(ct.InputSchema).ToMap()
	} else {
		fn.Parameters = map[string]any{"type": "object"}
	}

	return &GeminiTool{
		FunctionDeclarations: []GeminiFunctionDeclaration{fn},
	}, nil
}

// SupportsFeature returns whether this adapter supports a schema feature.
func (a *GeminiAdapter) SupportsFeature(feature SchemaFeature) bool {
	supported, ok := geminiFeatures[feature]
	return ok && supported
}

// filterGeminiSchema removes unsupported features from a schema for Gemini.
func filterGeminiSchema(schema *JSONSchema) *JSONSchema {
	if schema == nil {
		return nil
	}

	filtered := &JSONSchema{
		Type:        schema.Type,
		Title:       schema.Title,
		Description: schema.Description,
		Default:     schema.Default,
		Pattern:     schema.Pattern,
		Format:      schema.Format,
		Ref:         schema.Ref,
	}

	if schema.Nullable != nil {
		v := *schema.Nullable
		filtered.Nullable = &v
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
	if schema.AdditionalProperties != nil {
		v := *schema.AdditionalProperties
		filtered.AdditionalProperties = &v
	}

	if schema.Required != nil {
		filtered.Required = make([]string, len(schema.Required))
		copy(filtered.Required, schema.Required)
	}
	if schema.Enum != nil {
		filtered.Enum = make([]any, len(schema.Enum))
		copy(filtered.Enum, schema.Enum)
	}

	if schema.Properties != nil {
		filtered.Properties = make(map[string]*JSONSchema, len(schema.Properties))
		for k, v := range schema.Properties {
			filtered.Properties[k] = filterGeminiSchema(v)
		}
	}

	if schema.Items != nil {
		filtered.Items = filterGeminiSchema(schema.Items)
	}

	if schema.AnyOf != nil {
		filtered.AnyOf = make([]*JSONSchema, len(schema.AnyOf))
		for i, v := range schema.AnyOf {
			filtered.AnyOf[i] = filterGeminiSchema(v)
		}
	}

	if schema.Defs != nil {
		filtered.Defs = make(map[string]*JSONSchema, len(schema.Defs))
		for k, v := range schema.Defs {
			filtered.Defs[k] = filterGeminiSchema(v)
		}
	}

	return filtered
}
