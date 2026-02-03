package adapter

import (
	"errors"
	"fmt"
	"sync"
)

// ConversionResult contains the result of a format conversion.
type ConversionResult struct {
	// Tool is the converted tool in the target format
	Tool any

	// Warnings lists features that may have been lost during conversion
	Warnings []FeatureLossWarning
}

// AdapterRegistry is a thread-safe registry of protocol adapters.
type AdapterRegistry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

// NewRegistry creates a new empty adapter registry.
func NewRegistry() *AdapterRegistry {
	return &AdapterRegistry{
		adapters: make(map[string]Adapter),
	}
}

// Register adds an adapter to the registry.
// Returns an error if an adapter with the same name is already registered.
func (r *AdapterRegistry) Register(a Adapter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := a.Name()
	if _, exists := r.adapters[name]; exists {
		return errors.New("adapter already registered: " + name)
	}
	r.adapters[name] = a
	return nil
}

// Get retrieves an adapter by name.
// Returns an error if the adapter is not found.
func (r *AdapterRegistry) Get(name string) (Adapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, exists := r.adapters[name]
	if !exists {
		return nil, errors.New("adapter not found: " + name)
	}
	return adapter, nil
}

// List returns the names of all registered adapters.
func (r *AdapterRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	return names
}

// Unregister removes an adapter from the registry.
// Returns an error if the adapter is not found.
func (r *AdapterRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.adapters[name]; !exists {
		return errors.New("adapter not found: " + name)
	}
	delete(r.adapters, name)
	return nil
}

// Convert transforms a tool from one format to another.
// It uses the source adapter's ToCanonical and the target adapter's FromCanonical.
// Returns warnings if schema features are lost during conversion.
func (r *AdapterRegistry) Convert(tool any, fromFormat, toFormat string) (*ConversionResult, error) {
	// Get source adapter
	source, err := r.Get(fromFormat)
	if err != nil {
		return nil, err
	}

	// Get target adapter
	target, err := r.Get(toFormat)
	if err != nil {
		return nil, err
	}

	// Convert to canonical
	canonical, err := source.ToCanonical(tool)
	if err != nil {
		return nil, &ConversionError{
			Adapter:   fromFormat,
			Direction: "to_canonical",
			Cause:     err,
		}
	}

	// Check for feature loss
	warnings := detectFeatureLoss(canonical, source, target)

	// Convert from canonical
	output, err := target.FromCanonical(canonical)
	if err != nil {
		return nil, &ConversionError{
			Adapter:   toFormat,
			Direction: "from_canonical",
			Cause:     err,
		}
	}

	return &ConversionResult{
		Tool:     output,
		Warnings: warnings,
	}, nil
}

// detectFeatureLoss checks which features in the canonical tool are not
// supported by the target adapter.
func detectFeatureLoss(tool *CanonicalTool, source, target Adapter) []FeatureLossWarning {
	var warnings []FeatureLossWarning

	if tool.InputSchema != nil {
		warnings = append(warnings, detectSchemaFeatureLoss(tool.InputSchema, source, target, "")...)
	}
	if tool.OutputSchema != nil {
		warnings = append(warnings, detectSchemaFeatureLoss(tool.OutputSchema, source, target, "")...)
	}

	return warnings
}

// detectSchemaFeatureLoss checks which features in a schema are not supported.
func detectSchemaFeatureLoss(schema *JSONSchema, source, target Adapter, path string) []FeatureLossWarning {
	var warnings []FeatureLossWarning

	// Check each feature that's used in the schema
	featureUsage := map[SchemaFeature]bool{
		FeatureRef:                  schema.Ref != "",
		FeatureDefs:                 len(schema.Defs) > 0,
		FeatureAnyOf:                len(schema.AnyOf) > 0,
		FeatureOneOf:                len(schema.OneOf) > 0,
		FeatureAllOf:                len(schema.AllOf) > 0,
		FeatureNot:                  schema.Not != nil,
		FeatureTitle:                schema.Title != "",
		FeatureExamples:             len(schema.Examples) > 0,
		FeatureMultipleOf:           schema.MultipleOf != nil,
		FeaturePattern:              schema.Pattern != "",
		FeatureFormat:               schema.Format != "",
		FeatureAdditionalProperties: schema.AdditionalProperties != nil,
		FeatureMinimum:              schema.Minimum != nil,
		FeatureMaximum:              schema.Maximum != nil,
		FeatureMinLength:            schema.MinLength != nil,
		FeatureMaxLength:            schema.MaxLength != nil,
		FeatureMinItems:             schema.MinItems != nil,
		FeatureMaxItems:             schema.MaxItems != nil,
		FeatureMinProperties:        schema.MinProperties != nil,
		FeatureMaxProperties:        schema.MaxProperties != nil,
		FeatureUniqueItems:          schema.UniqueItems != nil,
		FeatureNullable:             schema.Nullable != nil,
		FeatureDeprecated:           schema.Deprecated != nil,
		FeatureReadOnly:             schema.ReadOnly != nil,
		FeatureWriteOnly:            schema.WriteOnly != nil,
		FeatureEnum:                 len(schema.Enum) > 0,
		FeatureConst:                schema.Const != nil,
		FeatureDefault:              schema.Default != nil,
	}

	for feature, used := range featureUsage {
		if used && !target.SupportsFeature(feature) {
			warnings = append(warnings, FeatureLossWarning{
				Feature:     feature,
				Path:        path,
				FromAdapter: source.Name(),
				ToAdapter:   target.Name(),
			})
		}
	}

	// Recursively check nested schemas
	if schema.Properties != nil {
		for name, prop := range schema.Properties {
			propPath := joinJSONPath(path, "properties", name)
			warnings = append(warnings, detectSchemaFeatureLoss(prop, source, target, propPath)...)
		}
	}
	if schema.Items != nil {
		warnings = append(warnings, detectSchemaFeatureLoss(schema.Items, source, target, joinJSONPath(path, "items"))...)
	}
	if schema.Defs != nil {
		for name, def := range schema.Defs {
			warnings = append(warnings, detectSchemaFeatureLoss(def, source, target, joinJSONPath(path, "$defs", name))...)
		}
	}
	for i, s := range schema.AnyOf {
		warnings = append(warnings, detectSchemaFeatureLoss(s, source, target, joinJSONPath(path, "anyOf", indexPath(i)))...)
	}
	for i, s := range schema.OneOf {
		warnings = append(warnings, detectSchemaFeatureLoss(s, source, target, joinJSONPath(path, "oneOf", indexPath(i)))...)
	}
	for i, s := range schema.AllOf {
		warnings = append(warnings, detectSchemaFeatureLoss(s, source, target, joinJSONPath(path, "allOf", indexPath(i)))...)
	}
	if schema.Not != nil {
		warnings = append(warnings, detectSchemaFeatureLoss(schema.Not, source, target, joinJSONPath(path, "not"))...)
	}

	return warnings
}

func joinJSONPath(base string, segments ...string) string {
	path := base
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		path = path + "/" + seg
	}
	return path
}

func indexPath(i int) string {
	return fmt.Sprintf("%d", i)
}
