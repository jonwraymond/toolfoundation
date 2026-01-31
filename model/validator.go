package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
)

// Common validation errors.
var (
	// ErrUnsupportedSchema is returned when the $schema dialect is not supported.
	ErrUnsupportedSchema = errors.New("unsupported JSON Schema dialect")

	// ErrExternalRef is returned when an external $ref is encountered.
	ErrExternalRef = errors.New("external $ref resolution is disabled")

	// ErrInvalidSchema is returned when a schema is not a valid JSON Schema object.
	ErrInvalidSchema = errors.New("invalid JSON Schema")
)

// Supported JSON Schema dialects.
const (
	SchemaDialect202012     = "https://json-schema.org/draft/2020-12/schema"
	SchemaDialectDraft07    = "http://json-schema.org/draft-07/schema#"
	SchemaDialectDraft07Alt = "http://json-schema.org/draft-07/schema"
)

// SchemaValidator validates JSON instances against JSON Schemas.
//
// Contract:
//   - Thread-safety: implementations must be safe for concurrent use unless documented otherwise.
//   - Ownership: implementations must not mutate caller-owned schema objects or instances.
//   - Errors: validation failures should wrap/return ErrInvalidSchema or ErrUnsupportedSchema where appropriate.
//   - ValidateInput: must validate against tool.InputSchema; return ErrInvalidSchema if tool is nil or InputSchema is nil.
//   - ValidateOutput: must validate against tool.OutputSchema when present; return nil when OutputSchema is nil.
//   - Determinism: repeated calls with same inputs should be deterministic.
// Implementations can be swapped to use different validation libraries.
type SchemaValidator interface {
	// Validate validates an instance against a JSON Schema.
	// The schema must be a valid JSON Schema object (map[string]any or *jsonschema.Schema).
	// The instance is the data to validate.
	// Returns nil if validation passes, otherwise returns a validation error.
	Validate(schema any, instance any) error

	// ValidateInput validates tool input arguments against a tool's InputSchema.
	// Convenience method that extracts InputSchema from Tool.
	ValidateInput(tool *Tool, args any) error

	// ValidateOutput validates tool output against a tool's OutputSchema if present.
	// Returns nil if OutputSchema is not defined.
	ValidateOutput(tool *Tool, result any) error
}

// DefaultValidator is the default SchemaValidator implementation using jsonschema-go.
// It supports JSON Schema 2020-12 (default) and draft-07.
// External $ref resolution is disabled to prevent network access.
//
// Limitations (from jsonschema-go):
//   - The "format" keyword is not validated by default (treated as annotation)
//   - Content-related keywords (contentEncoding, contentMediaType) are not validated
type DefaultValidator struct{}

// NewDefaultValidator creates a new DefaultValidator.
func NewDefaultValidator() *DefaultValidator {
	return &DefaultValidator{}
}

// Validate validates an instance against a JSON Schema.
func (v *DefaultValidator) Validate(schema any, instance any) error {
	// Convert schema to jsonschema.Schema
	jsSchema, err := v.toJSONSchema(schema)
	if err != nil {
		return err
	}

	// Check $schema dialect
	if err := v.checkDialect(jsSchema); err != nil {
		return err
	}

	// Resolve the schema with a loader that blocks external refs
	resolved, err := jsSchema.Resolve(&jsonschema.ResolveOptions{
		Loader: v.blockExternalRefs,
	})
	if err != nil {
		return fmt.Errorf("schema resolution failed: %w", err)
	}

	// Validate the instance
	if err := resolved.Validate(instance); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// ValidateInput validates tool input arguments against the tool's InputSchema.
func (v *DefaultValidator) ValidateInput(tool *Tool, args any) error {
	if tool == nil {
		return fmt.Errorf("%w: tool is nil", ErrInvalidSchema)
	}
	if tool.InputSchema == nil {
		return fmt.Errorf("%w: InputSchema is nil", ErrInvalidSchema)
	}
	return v.Validate(tool.InputSchema, args)
}

// ValidateOutput validates tool output against the tool's OutputSchema if present.
// Returns nil if OutputSchema is not defined.
func (v *DefaultValidator) ValidateOutput(tool *Tool, result any) error {
	if tool == nil {
		return fmt.Errorf("%w: tool is nil", ErrInvalidSchema)
	}
	if tool.OutputSchema == nil {
		return nil // OutputSchema is optional
	}
	return v.Validate(tool.OutputSchema, result)
}

// toJSONSchema converts various schema representations to jsonschema.Schema.
func (v *DefaultValidator) toJSONSchema(schema any) (*jsonschema.Schema, error) {
	switch s := schema.(type) {
	case *jsonschema.Schema:
		if s == nil {
			return nil, fmt.Errorf("%w: nil schema", ErrInvalidSchema)
		}
		// Return a shallow copy to avoid mutating caller-owned schema objects.
		copySchema := *s
		return &copySchema, nil
	case jsonschema.Schema:
		return &s, nil
	case json.RawMessage:
		if len(s) == 0 {
			return nil, fmt.Errorf("%w: empty schema", ErrInvalidSchema)
		}
		var jsSchema jsonschema.Schema
		if err := json.Unmarshal(s, &jsSchema); err != nil {
			return nil, fmt.Errorf("%w: failed to parse schema: %v", ErrInvalidSchema, err)
		}
		return &jsSchema, nil
	case []byte:
		if len(s) == 0 {
			return nil, fmt.Errorf("%w: empty schema", ErrInvalidSchema)
		}
		var jsSchema jsonschema.Schema
		if err := json.Unmarshal(s, &jsSchema); err != nil {
			return nil, fmt.Errorf("%w: failed to parse schema: %v", ErrInvalidSchema, err)
		}
		return &jsSchema, nil
	case map[string]any:
		// Marshal to JSON and unmarshal to jsonschema.Schema
		data, err := json.Marshal(s)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to marshal schema: %v", ErrInvalidSchema, err)
		}
		var jsSchema jsonschema.Schema
		if err := json.Unmarshal(data, &jsSchema); err != nil {
			return nil, fmt.Errorf("%w: failed to parse schema: %v", ErrInvalidSchema, err)
		}
		return &jsSchema, nil
	default:
		return nil, fmt.Errorf("%w: expected map[string]any or *jsonschema.Schema, got %T", ErrInvalidSchema, schema)
	}
}

// checkDialect validates that the $schema dialect is supported.
// If no $schema is specified, JSON Schema 2020-12 is assumed (per MCP rules).
// For draft-07 schemas, the $schema is cleared and validation proceeds using 2020-12 rules
// (most keywords are compatible between these versions).
func (v *DefaultValidator) checkDialect(schema *jsonschema.Schema) error {
	if schema.Schema == "" {
		// No $schema specified, default to 2020-12 (allowed per MCP spec)
		return nil
	}

	dialect := schema.Schema
	switch {
	case dialect == SchemaDialect202012:
		return nil
	case dialect == SchemaDialectDraft07 || dialect == SchemaDialectDraft07Alt:
		// Clear $schema for draft-07 to allow validation with 2020-12 rules.
		// jsonschema-go only supports 2020-12, but draft-07 schemas are largely compatible.
		schema.Schema = ""
		return nil
	case strings.HasPrefix(dialect, "https://json-schema.org/draft/2020-12/"):
		// Allow 2020-12 variants
		return nil
	case strings.HasPrefix(dialect, "http://json-schema.org/draft-07/"):
		// Clear $schema for draft-07 variants
		schema.Schema = ""
		return nil
	default:
		return fmt.Errorf("%w: %s (only 2020-12 and draft-07 are supported)", ErrUnsupportedSchema, dialect)
	}
}

// blockExternalRefs is a loader that blocks all external $ref resolution.
// This prevents network access during schema validation.
func (v *DefaultValidator) blockExternalRefs(uri *url.URL) (*jsonschema.Schema, error) {
	return nil, fmt.Errorf("%w: %s", ErrExternalRef, uri.String())
}

// Ensure DefaultValidator implements SchemaValidator.
var _ SchemaValidator = (*DefaultValidator)(nil)
