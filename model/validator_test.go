package model

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestDefaultValidator_Validate_BasicTypes(t *testing.T) {
	v := NewDefaultValidator()

	tests := []struct {
		name     string
		schema   map[string]any
		instance any
		wantErr  bool
	}{
		{
			name:     "string type valid",
			schema:   map[string]any{"type": "string"},
			instance: "hello",
			wantErr:  false,
		},
		{
			name:     "string type invalid",
			schema:   map[string]any{"type": "string"},
			instance: 123,
			wantErr:  true,
		},
		{
			name:     "integer type valid",
			schema:   map[string]any{"type": "integer"},
			instance: 42,
			wantErr:  false,
		},
		{
			name:     "number type valid",
			schema:   map[string]any{"type": "number"},
			instance: 3.14,
			wantErr:  false,
		},
		{
			name:     "boolean type valid",
			schema:   map[string]any{"type": "boolean"},
			instance: true,
			wantErr:  false,
		},
		{
			name:     "array type valid",
			schema:   map[string]any{"type": "array"},
			instance: []any{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "object type valid",
			schema:   map[string]any{"type": "object"},
			instance: map[string]any{"key": "value"},
			wantErr:  false,
		},
		{
			name:     "null type valid",
			schema:   map[string]any{"type": "null"},
			instance: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.schema, tt.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultValidator_Validate_ObjectProperties(t *testing.T) {
	v := NewDefaultValidator()

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"age":  map[string]any{"type": "integer"},
		},
		"required": []any{"name"},
	}

	tests := []struct {
		name     string
		instance any
		wantErr  bool
	}{
		{
			name:     "valid with all properties",
			instance: map[string]any{"name": "Alice", "age": 30},
			wantErr:  false,
		},
		{
			name:     "valid with only required",
			instance: map[string]any{"name": "Bob"},
			wantErr:  false,
		},
		{
			name:     "invalid missing required",
			instance: map[string]any{"age": 25},
			wantErr:  true,
		},
		{
			name:     "invalid wrong type for property",
			instance: map[string]any{"name": "Charlie", "age": "not a number"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(schema, tt.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultValidator_Validate_NoParameters(t *testing.T) {
	v := NewDefaultValidator()

	// MCP recommended "no parameters" schema
	recommendedSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
	}

	// MCP allowed "no parameters" schema
	allowedSchema := map[string]any{
		"type": "object",
	}

	tests := []struct {
		name     string
		schema   map[string]any
		instance any
		wantErr  bool
	}{
		{
			name:     "recommended schema empty object",
			schema:   recommendedSchema,
			instance: map[string]any{},
			wantErr:  false,
		},
		{
			name:     "recommended schema rejects extra props",
			schema:   recommendedSchema,
			instance: map[string]any{"extra": "value"},
			wantErr:  true,
		},
		{
			name:     "allowed schema empty object",
			schema:   allowedSchema,
			instance: map[string]any{},
			wantErr:  false,
		},
		{
			name:     "allowed schema accepts any props",
			schema:   allowedSchema,
			instance: map[string]any{"any": "value"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.schema, tt.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultValidator_Validate_SchemaDialect(t *testing.T) {
	v := NewDefaultValidator()

	tests := []struct {
		name    string
		schema  map[string]any
		wantErr bool
		errType error
	}{
		{
			name:    "no $schema (defaults to 2020-12)",
			schema:  map[string]any{"type": "string"},
			wantErr: false,
		},
		{
			name: "explicit 2020-12",
			schema: map[string]any{
				"$schema": SchemaDialect202012,
				"type":    "string",
			},
			wantErr: false,
		},
		{
			name: "draft-07",
			schema: map[string]any{
				"$schema": SchemaDialectDraft07,
				"type":    "string",
			},
			wantErr: false,
		},
		{
			name: "draft-07 without hash",
			schema: map[string]any{
				"$schema": SchemaDialectDraft07Alt,
				"type":    "string",
			},
			wantErr: false,
		},
		{
			name: "unsupported draft-04",
			schema: map[string]any{
				"$schema": "http://json-schema.org/draft-04/schema#",
				"type":    "string",
			},
			wantErr: true,
			errType: ErrUnsupportedSchema,
		},
		{
			name: "unsupported draft-06",
			schema: map[string]any{
				"$schema": "http://json-schema.org/draft-06/schema#",
				"type":    "string",
			},
			wantErr: true,
			errType: ErrUnsupportedSchema,
		},
		{
			name: "unsupported arbitrary schema",
			schema: map[string]any{
				"$schema": "https://example.com/my-schema",
				"type":    "string",
			},
			wantErr: true,
			errType: ErrUnsupportedSchema,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.schema, "test")
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.errType != nil && err != nil {
				if !strings.Contains(err.Error(), tt.errType.Error()) {
					t.Errorf("Validate() error = %v, want error containing %v", err, tt.errType)
				}
			}
		})
	}
}

func TestDefaultValidator_Validate_JsonschemaSchema(t *testing.T) {
	v := NewDefaultValidator()

	// Test with *jsonschema.Schema directly
	schema := &jsonschema.Schema{
		Type: "string",
	}

	err := v.Validate(schema, "hello")
	if err != nil {
		t.Errorf("Validate() with *jsonschema.Schema error = %v", err)
	}

	err = v.Validate(schema, 123)
	if err == nil {
		t.Error("Validate() with invalid instance should return error")
	}
}

func TestDefaultValidator_Validate_DoesNotMutateSchema(t *testing.T) {
	v := NewDefaultValidator()

	schema := &jsonschema.Schema{
		Schema: SchemaDialectDraft07,
		Type:   "string",
	}

	if err := v.Validate(schema, "ok"); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if schema.Schema != SchemaDialectDraft07 {
		t.Errorf("schema.Schema mutated: got %q, want %q", schema.Schema, SchemaDialectDraft07)
	}
}

func TestDefaultValidator_Validate_RawMessageSchema(t *testing.T) {
	v := NewDefaultValidator()

	schema := json.RawMessage(`{"type":"string"}`)
	if err := v.Validate(schema, "ok"); err != nil {
		t.Errorf("Validate() with json.RawMessage error = %v", err)
	}
	if err := v.Validate(schema, 123); err == nil {
		t.Error("Validate() with json.RawMessage should fail for invalid instance")
	}

	rawBytes := []byte(`{"type":"integer"}`)
	if err := v.Validate(rawBytes, 42); err != nil {
		t.Errorf("Validate() with []byte error = %v", err)
	}
	if err := v.Validate(rawBytes, "nope"); err == nil {
		t.Error("Validate() with []byte should fail for invalid instance")
	}
}

func TestDefaultValidator_Validate_InvalidSchemaType(t *testing.T) {
	v := NewDefaultValidator()

	// Test with invalid schema type
	err := v.Validate("not a schema", "test")
	if err == nil {
		t.Error("Validate() with string schema should return error")
	}
	if !strings.Contains(err.Error(), ErrInvalidSchema.Error()) {
		t.Errorf("Validate() error = %v, want error containing %v", err, ErrInvalidSchema)
	}
}

func TestDefaultValidator_ValidateInput(t *testing.T) {
	v := NewDefaultValidator()

	tool := Tool{
		Tool: mcp.Tool{
			Name:        "test-tool",
			Description: "A test tool",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{"type": "string"},
				},
				"required": []any{"query"},
			},
		},
	}

	tests := []struct {
		name    string
		args    any
		wantErr bool
	}{
		{
			name:    "valid input",
			args:    map[string]any{"query": "search term"},
			wantErr: false,
		},
		{
			name:    "missing required field",
			args:    map[string]any{},
			wantErr: true,
		},
		{
			name:    "wrong type",
			args:    map[string]any{"query": 123},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateInput(&tool, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultValidator_ValidateInput_RawMessageSchema(t *testing.T) {
	v := NewDefaultValidator()

	tool := Tool{
		Tool: mcp.Tool{
			Name:        "raw-schema-tool",
			Description: "Tool with RawMessage schema",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"q":{"type":"string"}},"required":["q"]}`),
		},
	}

	if err := v.ValidateInput(&tool, map[string]any{"q": "ok"}); err != nil {
		t.Errorf("ValidateInput() with RawMessage schema error = %v", err)
	}
	if err := v.ValidateInput(&tool, map[string]any{"q": 123}); err == nil {
		t.Error("ValidateInput() with RawMessage schema should fail for invalid input")
	}
}

func TestDefaultValidator_ValidateInput_NilSchema(t *testing.T) {
	v := NewDefaultValidator()

	tool := Tool{
		Tool: mcp.Tool{
			Name:        "nil-schema-tool",
			Description: "A tool with nil InputSchema",
			InputSchema: nil,
		},
	}

	err := v.ValidateInput(&tool, map[string]any{"anything": "goes"})
	if err == nil {
		t.Error("ValidateInput() with nil InputSchema should return error")
	}
	if !strings.Contains(err.Error(), ErrInvalidSchema.Error()) {
		t.Errorf("ValidateInput() error = %v, want error containing %v", err, ErrInvalidSchema)
	}
}

func TestDefaultValidator_ValidateOutput(t *testing.T) {
	v := NewDefaultValidator()

	t.Run("with OutputSchema", func(t *testing.T) {
		tool := Tool{
			Tool: mcp.Tool{
				Name:        "output-tool",
				Description: "A tool with OutputSchema",
				InputSchema: map[string]any{"type": "object"},
				OutputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"result": map[string]any{"type": "string"},
					},
				},
			},
		}

		// Valid output
		err := v.ValidateOutput(&tool, map[string]any{"result": "success"})
		if err != nil {
			t.Errorf("ValidateOutput() with valid output error = %v", err)
		}

		// Invalid output
		err = v.ValidateOutput(&tool, map[string]any{"result": 123})
		if err == nil {
			t.Error("ValidateOutput() with invalid output should return error")
		}
	})

	t.Run("without OutputSchema", func(t *testing.T) {
		tool := Tool{
			Tool: mcp.Tool{
				Name:         "no-output-tool",
				Description:  "A tool without OutputSchema",
				InputSchema:  map[string]any{"type": "object"},
				OutputSchema: nil,
			},
		}

		// Should return nil when no OutputSchema
		err := v.ValidateOutput(&tool, "anything")
		if err != nil {
			t.Errorf("ValidateOutput() without OutputSchema should return nil, got %v", err)
		}
	})
}

func TestDefaultValidator_ExternalRefBlocked(t *testing.T) {
	v := NewDefaultValidator()

	// Schema with external $ref
	schema := map[string]any{
		"$ref": "https://example.com/external-schema.json",
	}

	err := v.Validate(schema, map[string]any{})
	if err == nil {
		t.Error("Validate() with external $ref should return error")
	}
	if !strings.Contains(err.Error(), ErrExternalRef.Error()) {
		t.Errorf("Validate() error = %v, want error containing %v", err, ErrExternalRef)
	}
}

func TestDefaultValidator_LocalRef(t *testing.T) {
	v := NewDefaultValidator()

	// Schema with local $ref (should work)
	schema := map[string]any{
		"$defs": map[string]any{
			"name": map[string]any{"type": "string"},
		},
		"type": "object",
		"properties": map[string]any{
			"firstName": map[string]any{"$ref": "#/$defs/name"},
			"lastName":  map[string]any{"$ref": "#/$defs/name"},
		},
	}

	// Valid instance
	err := v.Validate(schema, map[string]any{
		"firstName": "John",
		"lastName":  "Doe",
	})
	if err != nil {
		t.Errorf("Validate() with local $ref error = %v", err)
	}

	// Invalid instance
	err = v.Validate(schema, map[string]any{
		"firstName": 123,
	})
	if err == nil {
		t.Error("Validate() should fail when local $ref type mismatches")
	}
}

func TestSchemaValidator_Interface(t *testing.T) {
	t.Helper()
	// Ensure DefaultValidator implements SchemaValidator interface
	var _ SchemaValidator = (*DefaultValidator)(nil)
	var _ SchemaValidator = NewDefaultValidator()
}
