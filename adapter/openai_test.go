package adapter

import (
	"errors"
	"testing"
)

func TestNewOpenAIAdapter(t *testing.T) {
	adapter := NewOpenAIAdapter()
	if adapter == nil {
		t.Fatal("NewOpenAIAdapter() returned nil")
	}
}

func TestOpenAIAdapter_Name(t *testing.T) {
	adapter := NewOpenAIAdapter()
	if adapter.Name() != "openai" {
		t.Errorf("Name() = %q, want %q", adapter.Name(), "openai")
	}
}

func TestOpenAIAdapter_SupportsFeature(t *testing.T) {
	adapter := NewOpenAIAdapter()

	// Supported features
	supported := []SchemaFeature{
		FeatureEnum,
		FeatureDefault,
		FeatureAdditionalProperties,
		FeatureMinimum,
		FeatureMaximum,
		FeatureMinLength,
		FeatureMaxLength,
		FeatureConst,
	}
	for _, feature := range supported {
		if !adapter.SupportsFeature(feature) {
			t.Errorf("SupportsFeature(%v) = false, want true", feature)
		}
	}

	// Unsupported features
	unsupported := []SchemaFeature{
		FeatureRef,
		FeatureDefs,
		FeatureAnyOf,
		FeatureOneOf,
		FeatureAllOf,
		FeatureNot,
		FeaturePattern,
		FeatureFormat,
	}
	for _, feature := range unsupported {
		if adapter.SupportsFeature(feature) {
			t.Errorf("SupportsFeature(%v) = true, want false", feature)
		}
	}
}

func TestOpenAIAdapter_ToCanonical(t *testing.T) {
	adapter := NewOpenAIAdapter()

	tests := []struct {
		name     string
		input    any
		wantName string
		wantErr  bool
	}{
		{
			name: "OpenAITool pointer",
			input: &OpenAITool{
				Type: "function",
				Function: OpenAIFunction{
					Name:        "test-function",
					Description: "A test function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"input": map[string]any{"type": "string"},
						},
					},
				},
			},
			wantName: "test-function",
			wantErr:  false,
		},
		{
			name: "OpenAITool value",
			input: OpenAITool{
				Type: "function",
				Function: OpenAIFunction{
					Name:       "value-function",
					Parameters: map[string]any{"type": "object"},
				},
			},
			wantName: "value-function",
			wantErr:  false,
		},
		{
			name: "OpenAIFunction pointer",
			input: &OpenAIFunction{
				Name:       "fn-ptr",
				Parameters: map[string]any{"type": "object"},
			},
			wantName: "fn-ptr",
			wantErr:  false,
		},
		{
			name: "OpenAIFunction value",
			input: OpenAIFunction{
				Name:       "fn-value",
				Parameters: map[string]any{"type": "object"},
			},
			wantName: "fn-value",
			wantErr:  false,
		},
		{
			name: "with strict mode",
			input: &OpenAITool{
				Type: "function",
				Function: OpenAIFunction{
					Name:       "strict-fn",
					Parameters: map[string]any{"type": "object"},
					Strict:     boolPtr(true),
				},
			},
			wantName: "strict-fn",
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
			input: &OpenAITool{
				Type: "function",
				Function: OpenAIFunction{
					Parameters: map[string]any{"type": "object"},
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
			if ct.SourceFormat != "openai" {
				t.Errorf("SourceFormat = %q, want %q", ct.SourceFormat, "openai")
			}
		})
	}
}

func TestOpenAIAdapter_ToCanonical_StrictMode(t *testing.T) {
	adapter := NewOpenAIAdapter()

	input := &OpenAITool{
		Type: "function",
		Function: OpenAIFunction{
			Name:       "strict-test",
			Parameters: map[string]any{"type": "object"},
			Strict:     boolPtr(true),
		},
	}

	ct, err := adapter.ToCanonical(input)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	strict, ok := ct.SourceMeta["strict"].(bool)
	if !ok || !strict {
		t.Errorf("SourceMeta[strict] = %v, want true", ct.SourceMeta["strict"])
	}
}

func TestOpenAIAdapter_FromCanonical(t *testing.T) {
	adapter := NewOpenAIAdapter()

	tests := []struct {
		name     string
		input    *CanonicalTool
		wantName string
		wantErr  bool
	}{
		{
			name: "full canonical tool",
			input: &CanonicalTool{
				Name:        "test-tool",
				Description: "A test tool",
				InputSchema: &JSONSchema{
					Type: "object",
					Properties: map[string]*JSONSchema{
						"input": {Type: "string"},
					},
				},
			},
			wantName: "test-tool",
			wantErr:  false,
		},
		{
			name: "minimal canonical tool",
			input: &CanonicalTool{
				Name: "minimal",
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

			tool, ok := result.(*OpenAITool)
			if !ok {
				t.Fatalf("FromCanonical() returned %T, want *OpenAITool", result)
			}

			if tool.Type != "function" {
				t.Errorf("Type = %q, want %q", tool.Type, "function")
			}
			if tool.Function.Name != tt.wantName {
				t.Errorf("Function.Name = %q, want %q", tool.Function.Name, tt.wantName)
			}
		})
	}
}

func TestOpenAIAdapter_FromCanonical_SchemaConversion(t *testing.T) {
	adapter := NewOpenAIAdapter()

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

	tool := result.(*OpenAITool)
	params := tool.Function.Parameters

	if params["type"] != "object" {
		t.Errorf("type = %v, want %q", params["type"], "object")
	}

	props, ok := params["properties"].(map[string]any)
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
	if nameProp["maxLength"] != 100 {
		t.Errorf("name.maxLength = %v, want 100", nameProp["maxLength"])
	}

	countProp, ok := props["count"].(map[string]any)
	if !ok {
		t.Fatal("count property is not map[string]any")
	}
	if countProp["minimum"] != float64(0) {
		t.Errorf("count.minimum = %v, want 0", countProp["minimum"])
	}
	if countProp["maximum"] != float64(1000) {
		t.Errorf("count.maximum = %v, want 1000", countProp["maximum"])
	}
	if countProp["default"] != float64(10) {
		t.Errorf("count.default = %v, want 10", countProp["default"])
	}

	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("required is not []string")
	}
	if len(required) != 1 || required[0] != "name" {
		t.Errorf("required = %v, want [name]", required)
	}

	if params["additionalProperties"] != false {
		t.Errorf("additionalProperties = %v, want false", params["additionalProperties"])
	}
}

func TestOpenAIAdapter_FromCanonical_FiltersUnsupportedFeatures(t *testing.T) {
	adapter := NewOpenAIAdapter()

	// Create a schema with features OpenAI doesn't support
	input := &CanonicalTool{
		Name: "filtered-test",
		InputSchema: &JSONSchema{
			Type: "object",
			Properties: map[string]*JSONSchema{
				"validated": {
					Type:    "string",
					Pattern: "^[a-z]+$", // Not supported
					Format:  "email",    // Not supported
				},
				"choice": {
					AnyOf: []*JSONSchema{ // Not supported
						{Type: "string"},
						{Type: "number"},
					},
				},
				"ref": {
					Ref: "#/$defs/address", // Not supported
				},
			},
			Defs: map[string]*JSONSchema{ // Not supported
				"address": {Type: "object"},
			},
			Required: []string{"validated"},
		},
	}

	result, err := adapter.FromCanonical(input)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	tool := result.(*OpenAITool)
	params := tool.Function.Parameters

	props, _ := params["properties"].(map[string]any)

	// Pattern and format should be filtered out
	validatedProp, _ := props["validated"].(map[string]any)
	if _, ok := validatedProp["pattern"]; ok {
		t.Error("pattern should be filtered out")
	}
	if _, ok := validatedProp["format"]; ok {
		t.Error("format should be filtered out")
	}

	// AnyOf should be filtered out
	choiceProp, _ := props["choice"].(map[string]any)
	if _, ok := choiceProp["anyOf"]; ok {
		t.Error("anyOf should be filtered out")
	}

	// $ref should be filtered out
	refProp, _ := props["ref"].(map[string]any)
	if _, ok := refProp["$ref"]; ok {
		t.Error("$ref should be filtered out")
	}

	// $defs should be filtered out
	if _, ok := params["$defs"]; ok {
		t.Error("$defs should be filtered out")
	}

	// But required should still be present
	required, ok := params["required"].([]string)
	if !ok || len(required) == 0 {
		t.Error("required should be preserved")
	}
}

func TestOpenAIAdapter_FromCanonical_StrictModePreserved(t *testing.T) {
	adapter := NewOpenAIAdapter()

	input := &CanonicalTool{
		Name: "strict-test",
		InputSchema: &JSONSchema{
			Type: "object",
		},
		SourceMeta: map[string]any{
			"strict": true,
		},
	}

	result, err := adapter.FromCanonical(input)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	tool := result.(*OpenAITool)
	if tool.Function.Strict == nil || !*tool.Function.Strict {
		t.Error("strict mode should be preserved")
	}
}

func TestOpenAIAdapter_RoundTrip(t *testing.T) {
	adapter := NewOpenAIAdapter()

	original := &OpenAITool{
		Type: "function",
		Function: OpenAIFunction{
			Name:        "roundtrip-test",
			Description: "Testing round-trip conversion",
			Parameters: map[string]any{
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
						"enum":    []any{0, 5, 10},
					},
				},
				"required":             []any{"input"},
				"additionalProperties": false,
			},
			Strict: boolPtr(true),
		},
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

	restored := result.(*OpenAITool)

	// Verify basic fields
	if restored.Type != "function" {
		t.Errorf("Type = %q, want %q", restored.Type, "function")
	}
	if restored.Function.Name != original.Function.Name {
		t.Errorf("Name = %q, want %q", restored.Function.Name, original.Function.Name)
	}
	if restored.Function.Description != original.Function.Description {
		t.Errorf("Description = %q, want %q", restored.Function.Description, original.Function.Description)
	}

	// Verify strict mode preserved
	if restored.Function.Strict == nil || !*restored.Function.Strict {
		t.Error("strict mode should be preserved")
	}

	// Verify schema structure
	params := restored.Function.Parameters
	if params["type"] != "object" {
		t.Errorf("type = %v, want %q", params["type"], "object")
	}

	props, _ := params["properties"].(map[string]any)
	inputProp, _ := props["input"].(map[string]any)
	if inputProp["type"] != "string" {
		t.Errorf("input.type = %v, want %q", inputProp["type"], "string")
	}
	if inputProp["minLength"] != 1 {
		t.Errorf("input.minLength = %v, want 1", inputProp["minLength"])
	}

	countProp, _ := props["count"].(map[string]any)
	if countProp["default"] != float64(5) {
		t.Errorf("count.default = %v, want 5", countProp["default"])
	}
}

func TestOpenAIAdapter_ImplementsInterface(t *testing.T) {
	var _ Adapter = (*OpenAIAdapter)(nil)
	var _ Adapter = NewOpenAIAdapter()
}

func TestOpenAIAdapter_FilterSchema_NilInput(t *testing.T) {
	// Test that filterOpenAISchema handles nil input
	result := filterOpenAISchema(nil)
	if result != nil {
		t.Error("filterOpenAISchema(nil) should return nil")
	}
}

func TestOpenAIAdapter_FilterSchema_WithItems(t *testing.T) {
	adapter := NewOpenAIAdapter()

	input := &CanonicalTool{
		Name: "items-test",
		InputSchema: &JSONSchema{
			Type: "object",
			Properties: map[string]*JSONSchema{
				"tags": {
					Type: "array",
					Items: &JSONSchema{
						Type:      "string",
						MinLength: intPtr(1),
					},
				},
			},
		},
	}

	result, err := adapter.FromCanonical(input)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	tool := result.(*OpenAITool)
	props, _ := tool.Function.Parameters["properties"].(map[string]any)
	tagsProp, _ := props["tags"].(map[string]any)
	items, ok := tagsProp["items"].(map[string]any)
	if !ok {
		t.Fatal("items should be present")
	}
	if items["type"] != "string" {
		t.Errorf("items.type = %v, want string", items["type"])
	}
}
