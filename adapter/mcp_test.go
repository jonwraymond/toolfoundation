package adapter

import (
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jonwraymond/toolfoundation/model"
)

func TestNewMCPAdapter(t *testing.T) {
	adapter := NewMCPAdapter()
	if adapter == nil {
		t.Fatal("NewMCPAdapter() returned nil")
	}
}

func TestMCPAdapter_Name(t *testing.T) {
	adapter := NewMCPAdapter()
	if adapter.Name() != "mcp" {
		t.Errorf("Name() = %q, want %q", adapter.Name(), "mcp")
	}
}

func TestMCPAdapter_SupportsFeature(t *testing.T) {
	adapter := NewMCPAdapter()

	// MCP should support all features
	for _, feature := range AllFeatures() {
		if !adapter.SupportsFeature(feature) {
			t.Errorf("SupportsFeature(%v) = false, want true", feature)
		}
	}
}

func TestMCPAdapter_ToCanonical(t *testing.T) {
	adapter := NewMCPAdapter()

	tests := []struct {
		name        string
		input       any
		wantName    string
		wantNS      string
		wantVersion string
		wantErr     bool
	}{
		{
			name: "model.Tool pointer",
			input: &model.Tool{
				Tool: mcp.Tool{
					Name:        "test-tool",
					Description: "A test tool",
					InputSchema: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"input": map[string]any{"type": "string"},
						},
					},
				},
				Namespace: "test-ns",
				Version:   "1.0.0",
				Tags:      []string{"alpha", "beta"},
			},
			wantName:    "test-tool",
			wantNS:      "test-ns",
			wantVersion: "1.0.0",
			wantErr:     false,
		},
		{
			name: "model.Tool value",
			input: model.Tool{
				Tool: mcp.Tool{
					Name:        "value-tool",
					InputSchema: map[string]any{"type": "object"},
				},
			},
			wantName: "value-tool",
			wantErr:  false,
		},
		{
			name: "mcp.Tool pointer",
			input: &mcp.Tool{
				Name:        "mcp-ptr-tool",
				InputSchema: map[string]any{"type": "object"},
			},
			wantName: "mcp-ptr-tool",
			wantErr:  false,
		},
		{
			name: "mcp.Tool value",
			input: mcp.Tool{
				Name:        "mcp-value-tool",
				InputSchema: map[string]any{"type": "object"},
			},
			wantName: "mcp-value-tool",
			wantErr:  false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
		{
			name:    "unsupported type",
			input:   "not a tool",
			wantErr: true,
		},
		{
			name: "missing name",
			input: &model.Tool{
				Tool: mcp.Tool{
					InputSchema: map[string]any{"type": "object"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, err := adapter.ToCanonical(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ToCanonical() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				var convErr *ConversionError
				if !errors.As(err, &convErr) {
					t.Errorf("expected ConversionError, got %T", err)
				}
				return
			}

			if ct.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", ct.Name, tt.wantName)
			}
			if ct.Namespace != tt.wantNS {
				t.Errorf("Namespace = %q, want %q", ct.Namespace, tt.wantNS)
			}
			if ct.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", ct.Version, tt.wantVersion)
			}
			if ct.SourceFormat != "mcp" {
				t.Errorf("SourceFormat = %q, want %q", ct.SourceFormat, "mcp")
			}
		})
	}
}

func TestMCPAdapter_ToCanonical_SchemaConversion(t *testing.T) {
	adapter := NewMCPAdapter()

	input := &model.Tool{
		Tool: mcp.Tool{
			Name:        "schema-test",
			Description: "Test schema conversion",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "The name",
						"minLength":   float64(1),
						"maxLength":   float64(100),
					},
					"count": map[string]any{
						"type":    "integer",
						"minimum": float64(0),
						"maximum": float64(1000),
						"default": float64(10),
					},
					"tags": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
				"required":             []any{"name"},
				"additionalProperties": false,
			},
		},
	}

	ct, err := adapter.ToCanonical(input)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	if ct.InputSchema == nil {
		t.Fatal("InputSchema is nil")
	}

	schema := ct.InputSchema

	if schema.Type != "object" {
		t.Errorf("Type = %q, want %q", schema.Type, "object")
	}

	if len(schema.Properties) != 3 {
		t.Errorf("Properties count = %d, want %d", len(schema.Properties), 3)
	}

	nameProp := schema.Properties["name"]
	if nameProp == nil {
		t.Fatal("name property is nil")
	}
	if nameProp.Type != "string" {
		t.Errorf("name.Type = %q, want %q", nameProp.Type, "string")
	}
	if nameProp.MinLength == nil || *nameProp.MinLength != 1 {
		t.Errorf("name.MinLength = %v, want 1", nameProp.MinLength)
	}
	if nameProp.MaxLength == nil || *nameProp.MaxLength != 100 {
		t.Errorf("name.MaxLength = %v, want 100", nameProp.MaxLength)
	}

	countProp := schema.Properties["count"]
	if countProp == nil {
		t.Fatal("count property is nil")
	}
	if countProp.Minimum == nil || *countProp.Minimum != 0 {
		t.Errorf("count.Minimum = %v, want 0", countProp.Minimum)
	}
	if countProp.Maximum == nil || *countProp.Maximum != 1000 {
		t.Errorf("count.Maximum = %v, want 1000", countProp.Maximum)
	}
	if countProp.Default != float64(10) {
		t.Errorf("count.Default = %v, want 10", countProp.Default)
	}

	tagsProp := schema.Properties["tags"]
	if tagsProp == nil {
		t.Fatal("tags property is nil")
	}
	if tagsProp.Items == nil {
		t.Fatal("tags.Items is nil")
	}
	if tagsProp.Items.Type != "string" {
		t.Errorf("tags.Items.Type = %q, want %q", tagsProp.Items.Type, "string")
	}

	if len(schema.Required) != 1 || schema.Required[0] != "name" {
		t.Errorf("Required = %v, want [name]", schema.Required)
	}

	if schema.AdditionalProperties == nil || *schema.AdditionalProperties != false {
		t.Errorf("AdditionalProperties = %v, want false", schema.AdditionalProperties)
	}
}

func TestMCPAdapter_ToCanonical_AdvancedSchema(t *testing.T) {
	adapter := NewMCPAdapter()

	input := &model.Tool{
		Tool: mcp.Tool{
			Name: "advanced-schema",
			InputSchema: map[string]any{
				"type": "object",
				"$defs": map[string]any{
					"address": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"street": map[string]any{"type": "string"},
						},
					},
				},
				"properties": map[string]any{
					"choice": map[string]any{
						"anyOf": []any{
							map[string]any{"type": "string"},
							map[string]any{"type": "number"},
						},
					},
					"exclusive": map[string]any{
						"oneOf": []any{
							map[string]any{"const": "a"},
							map[string]any{"const": "b"},
						},
					},
					"combined": map[string]any{
						"allOf": []any{
							map[string]any{"type": "object"},
							map[string]any{"required": []any{"id"}},
						},
					},
					"notEmpty": map[string]any{
						"not": map[string]any{"const": ""},
					},
					"ref": map[string]any{
						"$ref": "#/$defs/address",
					},
					"formatted": map[string]any{
						"type":    "string",
						"format":  "email",
						"pattern": "^[a-z]+@",
					},
					"fixed": map[string]any{
						"const": "fixed-value",
					},
					"limited": map[string]any{
						"enum": []any{"a", "b", "c"},
					},
				},
			},
		},
	}

	ct, err := adapter.ToCanonical(input)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	schema := ct.InputSchema

	// Check $defs
	if len(schema.Defs) != 1 {
		t.Errorf("Defs count = %d, want 1", len(schema.Defs))
	}
	if schema.Defs["address"] == nil {
		t.Error("address def is nil")
	}

	// Check anyOf
	choiceProp := schema.Properties["choice"]
	if len(choiceProp.AnyOf) != 2 {
		t.Errorf("choice.AnyOf count = %d, want 2", len(choiceProp.AnyOf))
	}

	// Check oneOf
	exclusiveProp := schema.Properties["exclusive"]
	if len(exclusiveProp.OneOf) != 2 {
		t.Errorf("exclusive.OneOf count = %d, want 2", len(exclusiveProp.OneOf))
	}

	// Check allOf
	combinedProp := schema.Properties["combined"]
	if len(combinedProp.AllOf) != 2 {
		t.Errorf("combined.AllOf count = %d, want 2", len(combinedProp.AllOf))
	}

	// Check not
	notEmptyProp := schema.Properties["notEmpty"]
	if notEmptyProp.Not == nil {
		t.Error("notEmpty.Not is nil")
	}

	// Check $ref
	refProp := schema.Properties["ref"]
	if refProp.Ref != "#/$defs/address" {
		t.Errorf("ref.Ref = %q, want %q", refProp.Ref, "#/$defs/address")
	}

	// Check format and pattern
	formattedProp := schema.Properties["formatted"]
	if formattedProp.Format != "email" {
		t.Errorf("formatted.Format = %q, want %q", formattedProp.Format, "email")
	}
	if formattedProp.Pattern != "^[a-z]+@" {
		t.Errorf("formatted.Pattern = %q, want %q", formattedProp.Pattern, "^[a-z]+@")
	}

	// Check const
	fixedProp := schema.Properties["fixed"]
	if fixedProp.Const != "fixed-value" {
		t.Errorf("fixed.Const = %v, want %q", fixedProp.Const, "fixed-value")
	}

	// Check enum
	limitedProp := schema.Properties["limited"]
	if len(limitedProp.Enum) != 3 {
		t.Errorf("limited.Enum count = %d, want 3", len(limitedProp.Enum))
	}
}

func TestMCPAdapter_FromCanonical(t *testing.T) {
	adapter := NewMCPAdapter()

	tests := []struct {
		name        string
		input       *CanonicalTool
		wantName    string
		wantNS      string
		wantVersion string
		wantErr     bool
	}{
		{
			name: "full canonical tool",
			input: &CanonicalTool{
				Namespace:   "test-ns",
				Name:        "test-tool",
				Version:     "1.0.0",
				Description: "A test tool",
				Tags:        []string{"alpha", "beta"},
				InputSchema: &JSONSchema{
					Type: "object",
					Properties: map[string]*JSONSchema{
						"input": {Type: "string"},
					},
				},
			},
			wantName:    "test-tool",
			wantNS:      "test-ns",
			wantVersion: "1.0.0",
			wantErr:     false,
		},
		{
			name: "minimal canonical tool",
			input: &CanonicalTool{
				Name: "minimal",
				InputSchema: &JSONSchema{
					Type: "object",
				},
			},
			wantName: "minimal",
			wantErr:  false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
		{
			name: "missing name",
			input: &CanonicalTool{
				InputSchema: &JSONSchema{Type: "object"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.FromCanonical(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("FromCanonical() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				var convErr *ConversionError
				if !errors.As(err, &convErr) {
					t.Errorf("expected ConversionError, got %T", err)
				}
				return
			}

			tool, ok := result.(*model.Tool)
			if !ok {
				t.Fatalf("FromCanonical() returned %T, want *model.Tool", result)
			}

			if tool.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", tool.Name, tt.wantName)
			}
			if tool.Namespace != tt.wantNS {
				t.Errorf("Namespace = %q, want %q", tool.Namespace, tt.wantNS)
			}
			if tool.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", tool.Version, tt.wantVersion)
			}
		})
	}
}

func TestMCPAdapter_FromCanonical_SchemaConversion(t *testing.T) {
	adapter := NewMCPAdapter()

	input := &CanonicalTool{
		Name:        "schema-test",
		Description: "Test schema conversion",
		InputSchema: &JSONSchema{
			Type: "object",
			Properties: map[string]*JSONSchema{
				"name": {
					Type:        "string",
					Description: "The name",
					MinLength:   intPtr(1),
					MaxLength:   intPtr(100),
				},
				"count": {
					Type:    "integer",
					Minimum: floatPtr(0),
					Maximum: floatPtr(1000),
					Default: float64(10),
				},
			},
			Required:             []string{"name"},
			AdditionalProperties: boolPtr(false),
		},
	}

	result, err := adapter.FromCanonical(input)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	tool := result.(*model.Tool)

	if tool.InputSchema == nil {
		t.Fatal("InputSchema is nil")
	}

	schema, ok := tool.InputSchema.(map[string]any)
	if !ok {
		t.Fatalf("InputSchema is %T, want map[string]any", tool.InputSchema)
	}

	if schema["type"] != "object" {
		t.Errorf("type = %v, want %q", schema["type"], "object")
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties is not map[string]any")
	}

	nameProp, ok := props["name"].(map[string]any)
	if !ok {
		t.Fatal("name property is not map[string]any")
	}
	if nameProp["type"] != "string" {
		t.Errorf("name.type = %v, want %q", nameProp["type"], "string")
	}
	if nameProp["minLength"] != 1 {
		t.Errorf("name.minLength = %v, want 1", nameProp["minLength"])
	}

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("required is not []string")
	}
	if len(required) != 1 || required[0] != "name" {
		t.Errorf("required = %v, want [name]", required)
	}

	if schema["additionalProperties"] != false {
		t.Errorf("additionalProperties = %v, want false", schema["additionalProperties"])
	}
}

func TestMCPAdapter_RoundTrip(t *testing.T) {
	adapter := NewMCPAdapter()

	original := &model.Tool{
		Tool: mcp.Tool{
			Meta: mcp.Meta{
				"traceId": "abc123",
			},
			Annotations: &mcp.ToolAnnotations{
				Title:          "Test Title",
				ReadOnlyHint:   true,
				IdempotentHint: true,
			},
			Name:        "roundtrip-tool",
			Title:       "Roundtrip Test",
			Description: "Testing round-trip conversion",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"input": map[string]any{
						"type":        "string",
						"description": "The input value",
						"minLength":   float64(1),
					},
					"count": map[string]any{
						"type":    "integer",
						"minimum": float64(0),
						"default": float64(5),
					},
				},
				"required": []any{"input"},
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"result": map[string]any{"type": "boolean"},
				},
			},
			Icons: []mcp.Icon{
				{Source: "https://example.com/icon.png", MIMEType: "image/png"},
			},
		},
		Namespace: "test-ns",
		Version:   "2.0.0",
		Tags:      []string{"alpha", "beta"},
	}

	// Convert to canonical
	canonical, err := adapter.ToCanonical(original)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	// Convert back
	result, err := adapter.FromCanonical(canonical)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	restored := result.(*model.Tool)

	// Verify basic fields
	if restored.Name != original.Name {
		t.Errorf("Name = %q, want %q", restored.Name, original.Name)
	}
	if restored.Description != original.Description {
		t.Errorf("Description = %q, want %q", restored.Description, original.Description)
	}
	if restored.Namespace != original.Namespace {
		t.Errorf("Namespace = %q, want %q", restored.Namespace, original.Namespace)
	}
	if restored.Version != original.Version {
		t.Errorf("Version = %q, want %q", restored.Version, original.Version)
	}

	// Verify tags
	if len(restored.Tags) != len(original.Tags) {
		t.Errorf("Tags length = %d, want %d", len(restored.Tags), len(original.Tags))
	}

	// Verify MCP-specific fields restored from SourceMeta
	if restored.Title != original.Title {
		t.Errorf("Title = %q, want %q", restored.Title, original.Title)
	}

	// Verify InputSchema structure is preserved
	if restored.InputSchema == nil {
		t.Fatal("InputSchema is nil")
	}
	restoredSchema := restored.InputSchema.(map[string]any)
	if restoredSchema["type"] != "object" {
		t.Errorf("InputSchema.type = %v, want %q", restoredSchema["type"], "object")
	}

	// Verify OutputSchema is preserved
	if restored.OutputSchema == nil {
		t.Fatal("OutputSchema is nil")
	}
}

func TestMCPAdapter_ToCanonical_MetaFields(t *testing.T) {
	adapter := NewMCPAdapter()

	tool := &model.Tool{
		Tool: mcp.Tool{
			Name:        "meta-tool",
			Description: "meta tool",
			InputSchema: map[string]any{"type": "object"},
			Meta: mcp.Meta{
				"summary":             "short summary",
				"category":            "utility",
				"inputModes":          []any{"application/json"},
				"outputModes":         []any{"application/json"},
				"examples":            []any{"example"},
				"deterministic":       true,
				"streaming":           true,
				"securitySchemes":     map[string]any{"apiKey": map[string]any{"type": "apiKey"}},
				"securityRequirements": []any{map[string]any{"apiKey": []any{}}},
				"uiHints":             map[string]any{"layout": "compact"},
			},
		},
	}

	ct, err := adapter.ToCanonical(tool)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}
	if ct.Summary != "short summary" {
		t.Errorf("Summary = %q, want %q", ct.Summary, "short summary")
	}
	if ct.Category != "utility" {
		t.Errorf("Category = %q, want %q", ct.Category, "utility")
	}
	if len(ct.InputModes) != 1 || ct.InputModes[0] != "application/json" {
		t.Errorf("InputModes = %v, want [application/json]", ct.InputModes)
	}
	if len(ct.OutputModes) != 1 || ct.OutputModes[0] != "application/json" {
		t.Errorf("OutputModes = %v, want [application/json]", ct.OutputModes)
	}
	if len(ct.Examples) != 1 || ct.Examples[0] != "example" {
		t.Errorf("Examples = %v, want [example]", ct.Examples)
	}
	if ct.Deterministic == nil || !*ct.Deterministic {
		t.Errorf("Deterministic = %v, want true", ct.Deterministic)
	}
	if ct.Streaming == nil || !*ct.Streaming {
		t.Errorf("Streaming = %v, want true", ct.Streaming)
	}
	if ct.SecuritySchemes == nil || ct.SecuritySchemes["apiKey"] == nil {
		t.Error("SecuritySchemes missing apiKey")
	}
	if len(ct.SecurityRequirements) != 1 {
		t.Errorf("SecurityRequirements length = %d, want 1", len(ct.SecurityRequirements))
	}
	if ct.UIHints["layout"] != "compact" {
		t.Errorf("UIHints layout = %v, want compact", ct.UIHints["layout"])
	}
}

func TestMCPAdapter_FromCanonical_MetaFields(t *testing.T) {
	adapter := NewMCPAdapter()

	deterministic := true
	streaming := true

	ct := &CanonicalTool{
		Name:          "meta-tool",
		Description:   "meta tool",
		InputSchema:   &JSONSchema{Type: "object"},
		Summary:       "short summary",
		Category:      "utility",
		InputModes:    []string{"application/json"},
		OutputModes:   []string{"application/json"},
		Examples:      []string{"example"},
		Deterministic: &deterministic,
		Streaming:     &streaming,
		SecuritySchemes: map[string]SecurityScheme{
			"apiKey": {"type": "apiKey"},
		},
		SecurityRequirements: []SecurityRequirement{
			{"apiKey": {}},
		},
		UIHints: map[string]any{"layout": "compact"},
	}

	raw, err := adapter.FromCanonical(ct)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	tool := raw.(*model.Tool)
	if tool.Meta == nil {
		t.Fatal("Meta is nil")
	}
	if tool.Meta["summary"] != "short summary" {
		t.Errorf("meta.summary = %v, want short summary", tool.Meta["summary"])
	}
	if tool.Meta["category"] != "utility" {
		t.Errorf("meta.category = %v, want utility", tool.Meta["category"])
	}
}

func TestMCPAdapter_ImplementsInterface(t *testing.T) {
	var _ Adapter = (*MCPAdapter)(nil)
	var _ Adapter = NewMCPAdapter()
}

func TestMCPAdapter_ToCanonical_WithJSONSchemaInput(t *testing.T) {
	adapter := NewMCPAdapter()

	// Test with *JSONSchema as InputSchema
	jsonSchema := &JSONSchema{
		Type: "object",
		Properties: map[string]*JSONSchema{
			"name": {Type: "string"},
		},
		Required: []string{"name"},
	}

	// We need to test that schemaFromAny handles *JSONSchema
	// This is done indirectly through the adapter
	input := &model.Tool{
		Tool: mcp.Tool{
			Name:        "jsonschema-test",
			InputSchema: jsonSchema.ToMap(), // Convert to map first
		},
	}

	ct, err := adapter.ToCanonical(input)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	if ct.InputSchema == nil {
		t.Fatal("InputSchema should not be nil")
	}
	if ct.InputSchema.Type != "object" {
		t.Errorf("InputSchema.Type = %q, want %q", ct.InputSchema.Type, "object")
	}
}

func TestMCPAdapter_ToCanonical_WithOutputSchema(t *testing.T) {
	adapter := NewMCPAdapter()

	input := &model.Tool{
		Tool: mcp.Tool{
			Name: "output-schema-test",
			InputSchema: map[string]any{
				"type": "object",
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"result": map[string]any{"type": "boolean"},
				},
			},
		},
	}

	ct, err := adapter.ToCanonical(input)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	if ct.OutputSchema == nil {
		t.Fatal("OutputSchema should not be nil")
	}
	if ct.OutputSchema.Type != "object" {
		t.Errorf("OutputSchema.Type = %q, want %q", ct.OutputSchema.Type, "object")
	}
}

func TestMCPAdapter_FromCanonical_WithSourceMeta(t *testing.T) {
	adapter := NewMCPAdapter()

	// Test with map[string]any meta (simulating deserialized JSON)
	input := &CanonicalTool{
		Name: "meta-test",
		InputSchema: &JSONSchema{
			Type: "object",
		},
		SourceMeta: map[string]any{
			"title": "Test Title",
			"meta":  map[string]any{"key": "value"},
		},
	}

	result, err := adapter.FromCanonical(input)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	tool := result.(*model.Tool)
	if tool.Title != "Test Title" {
		t.Errorf("Title = %q, want %q", tool.Title, "Test Title")
	}
	if tool.Meta == nil {
		t.Fatal("Meta should not be nil")
	}
	if tool.Meta["key"] != "value" {
		t.Errorf("Meta[key] = %v, want %q", tool.Meta["key"], "value")
	}
}

func TestSchemaFromAny_JSONSchemaTypes(t *testing.T) {
	// Test *JSONSchema input
	t.Run("JSONSchema pointer", func(t *testing.T) {
		input := &JSONSchema{
			Type:        "string",
			Description: "A test string",
			MinLength:   intPtr(1),
		}

		result, err := schemaFromAny(input)
		if err != nil {
			t.Fatalf("schemaFromAny() error = %v", err)
		}

		if result.Type != "string" {
			t.Errorf("Type = %q, want %q", result.Type, "string")
		}
		if result.Description != "A test string" {
			t.Errorf("Description = %q, want %q", result.Description, "A test string")
		}
		// Verify it's a deep copy
		if result == input {
			t.Error("Result should be a deep copy, not the same pointer")
		}
	})

	// Test JSONSchema value input
	t.Run("JSONSchema value", func(t *testing.T) {
		input := JSONSchema{
			Type:    "number",
			Minimum: floatPtr(0),
			Maximum: floatPtr(100),
		}

		result, err := schemaFromAny(input)
		if err != nil {
			t.Fatalf("schemaFromAny() error = %v", err)
		}

		if result.Type != "number" {
			t.Errorf("Type = %q, want %q", result.Type, "number")
		}
		if result.Minimum == nil || *result.Minimum != 0 {
			t.Errorf("Minimum = %v, want 0", result.Minimum)
		}
	})

	// Test nil input
	t.Run("nil input", func(t *testing.T) {
		result, err := schemaFromAny(nil)
		if err != nil {
			t.Fatalf("schemaFromAny() error = %v", err)
		}
		if result != nil {
			t.Error("Result should be nil for nil input")
		}
	})

	// Test unsupported type
	t.Run("unsupported type", func(t *testing.T) {
		_, err := schemaFromAny("not a schema")
		if err == nil {
			t.Error("Expected error for unsupported type")
		}
	})
}

func TestSchemaFromMap_AllFields(t *testing.T) {
	// Test nil map
	t.Run("nil map", func(t *testing.T) {
		result := schemaFromMap(nil)
		if result != nil {
			t.Error("Result should be nil for nil map")
		}
	})

	// Test with []string required (direct slice)
	t.Run("required as []string", func(t *testing.T) {
		input := map[string]any{
			"type":     "object",
			"required": []string{"field1", "field2"},
		}

		result := schemaFromMap(input)
		if len(result.Required) != 2 {
			t.Errorf("Required length = %d, want 2", len(result.Required))
		}
	})
}

func TestMCPAdapter_ToCanonical_InvalidOutputSchema(t *testing.T) {
	adapter := NewMCPAdapter()

	// Test with invalid OutputSchema type
	input := &model.Tool{
		Tool: mcp.Tool{
			Name: "invalid-output",
			InputSchema: map[string]any{
				"type": "object",
			},
			OutputSchema: "not a valid schema", // string instead of map
		},
	}

	_, err := adapter.ToCanonical(input)
	if err == nil {
		t.Error("ToCanonical() should fail with invalid OutputSchema")
	}

	var convErr *ConversionError
	if !errors.As(err, &convErr) {
		t.Errorf("expected ConversionError, got %T", err)
	}
}

// Helper functions
func intPtr(v int) *int {
	return &v
}

func floatPtr(v float64) *float64 {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}
