package adapter

import "fmt"

// SchemaFeature represents a JSON Schema feature that may or may not be
// supported by a particular protocol adapter.
type SchemaFeature int

const (
	// FeatureRef is the $ref keyword for schema references
	FeatureRef SchemaFeature = iota
	// FeatureDefs is the $defs keyword for schema definitions
	FeatureDefs
	// FeatureAnyOf allows any of the listed schemas
	FeatureAnyOf
	// FeatureOneOf requires exactly one of the listed schemas
	FeatureOneOf
	// FeatureAllOf requires all of the listed schemas
	FeatureAllOf
	// FeatureNot disallows the specified schema
	FeatureNot
	// FeaturePattern is regex pattern validation
	FeaturePattern
	// FeatureFormat is semantic format validation
	FeatureFormat
	// FeatureAdditionalProperties controls extra property handling
	FeatureAdditionalProperties
	// FeatureMinimum is minimum numeric value
	FeatureMinimum
	// FeatureMaximum is maximum numeric value
	FeatureMaximum
	// FeatureMinLength is minimum string length
	FeatureMinLength
	// FeatureMaxLength is maximum string length
	FeatureMaxLength
	// FeatureEnum restricts to a set of values
	FeatureEnum
	// FeatureConst restricts to a single value
	FeatureConst
	// FeatureDefault provides a default value
	FeatureDefault
	// FeatureTitle is a schema title
	FeatureTitle
	// FeatureExamples provides example values
	FeatureExamples
	// FeatureMultipleOf is numeric multiple constraint
	FeatureMultipleOf
	// FeatureMinItems is minimum array length
	FeatureMinItems
	// FeatureMaxItems is maximum array length
	FeatureMaxItems
	// FeatureMinProperties is minimum object property count
	FeatureMinProperties
	// FeatureMaxProperties is maximum object property count
	FeatureMaxProperties
	// FeatureUniqueItems requires array items to be unique
	FeatureUniqueItems
	// FeatureNullable indicates nullable values (OpenAPI)
	FeatureNullable
	// FeatureDeprecated marks schema as deprecated
	FeatureDeprecated
	// FeatureReadOnly indicates read-only properties
	FeatureReadOnly
	// FeatureWriteOnly indicates write-only properties
	FeatureWriteOnly
)

// featureNames maps features to their string representations
var featureNames = map[SchemaFeature]string{
	FeatureRef:                  "$ref",
	FeatureDefs:                 "$defs",
	FeatureAnyOf:                "anyOf",
	FeatureOneOf:                "oneOf",
	FeatureAllOf:                "allOf",
	FeatureNot:                  "not",
	FeaturePattern:              "pattern",
	FeatureFormat:               "format",
	FeatureAdditionalProperties: "additionalProperties",
	FeatureMinimum:              "minimum",
	FeatureMaximum:              "maximum",
	FeatureMinLength:            "minLength",
	FeatureMaxLength:            "maxLength",
	FeatureEnum:                 "enum",
	FeatureConst:                "const",
	FeatureDefault:              "default",
	FeatureTitle:                "title",
	FeatureExamples:             "examples",
	FeatureMultipleOf:           "multipleOf",
	FeatureMinItems:             "minItems",
	FeatureMaxItems:             "maxItems",
	FeatureMinProperties:        "minProperties",
	FeatureMaxProperties:        "maxProperties",
	FeatureUniqueItems:          "uniqueItems",
	FeatureNullable:             "nullable",
	FeatureDeprecated:           "deprecated",
	FeatureReadOnly:             "readOnly",
	FeatureWriteOnly:            "writeOnly",
}

// String returns the JSON Schema keyword name for this feature.
func (f SchemaFeature) String() string {
	if name, ok := featureNames[f]; ok {
		return name
	}
	return fmt.Sprintf("SchemaFeature(%d)", f)
}

// AllFeatures returns all known schema features in a stable order.
func AllFeatures() []SchemaFeature {
	return []SchemaFeature{
		FeatureRef,
		FeatureDefs,
		FeatureAnyOf,
		FeatureOneOf,
		FeatureAllOf,
		FeatureNot,
		FeaturePattern,
		FeatureFormat,
		FeatureAdditionalProperties,
		FeatureMinimum,
		FeatureMaximum,
		FeatureMinLength,
		FeatureMaxLength,
		FeatureEnum,
		FeatureConst,
		FeatureDefault,
		FeatureTitle,
		FeatureExamples,
		FeatureMultipleOf,
		FeatureMinItems,
		FeatureMaxItems,
		FeatureMinProperties,
		FeatureMaxProperties,
		FeatureUniqueItems,
		FeatureNullable,
		FeatureDeprecated,
		FeatureReadOnly,
		FeatureWriteOnly,
	}
}

// Adapter defines the contract for protocol-specific tool adapters.
// Each adapter handles bidirectional conversion between a specific
// tool format (MCP, OpenAI, Anthropic) and the canonical representation.
//
// Contract:
//   - Thread-safety: implementations must be safe for concurrent use unless documented otherwise.
//   - Ownership: must not mutate caller-owned inputs; return new objects on conversion.
//   - Errors: ToCanonical/FromCanonical must return typed errors (e.g., *ConversionError) for invalid input.
//   - Determinism: same input yields structurally equivalent canonical output.
type Adapter interface {
	// Name returns the adapter's identifier (e.g., "mcp", "openai", "anthropic")
	Name() string

	// ToCanonical converts a protocol-specific tool to canonical format.
	// The raw parameter type depends on the adapter (e.g., mcp.Tool, OpenAIFunction).
	ToCanonical(raw any) (*CanonicalTool, error)

	// FromCanonical converts a canonical tool to the protocol-specific format.
	// The returned type depends on the adapter.
	FromCanonical(tool *CanonicalTool) (any, error)

	// SupportsFeature returns whether this adapter supports a schema feature.
	// Features not supported will generate warnings during conversion.
	SupportsFeature(feature SchemaFeature) bool
}

// ConversionError represents an error during tool format conversion.
type ConversionError struct {
	// Adapter is the name of the adapter that encountered the error
	Adapter string

	// Direction is "to_canonical" or "from_canonical"
	Direction string

	// Cause is the underlying error
	Cause error
}

// Error returns a formatted error message including adapter, direction, and cause.
func (e *ConversionError) Error() string {
	return fmt.Sprintf("%s adapter %s: %v", e.Adapter, e.Direction, e.Cause)
}

// Unwrap returns the underlying cause for use with errors.Is and errors.As.
func (e *ConversionError) Unwrap() error {
	return e.Cause
}

// FeatureLossWarning indicates that a schema feature will be lost during conversion.
// This is a warning, not an error - the conversion proceeds but with reduced fidelity.
type FeatureLossWarning struct {
	// Feature is the schema feature that will be lost
	Feature SchemaFeature

	// Path is the JSON pointer path to the schema location using the feature.
	// Empty string indicates the root schema.
	Path string

	// FromAdapter is the source adapter name
	FromAdapter string

	// ToAdapter is the target adapter name
	ToAdapter string
}

// String returns a human-readable warning message.
func (w FeatureLossWarning) String() string {
	path := w.Path
	if path == "" {
		path = "/"
	}
	return fmt.Sprintf("feature %s lost converting from %s to %s at %s",
		w.Feature, w.FromAdapter, w.ToAdapter, path)
}
