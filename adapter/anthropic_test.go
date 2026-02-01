package adapter

import (
	"errors"
	"testing"
)

func TestNewAnthropicAdapter(t *testing.T) {
	adapter := NewAnthropicAdapter()
	if adapter == nil {
		t.Fatal("NewAnthropicAdapter() returned nil")
	}
}

func TestAnthropicAdapter_Name(t *testing.T) {
	adapter := NewAnthropicAdapter()
	if adapter.Name() != "anthropic" {
		t.Errorf("Name() = %q, want %q", adapter.Name(), "anthropic")
	}
}

func TestAnthropicAdapter_SupportsFeature(t *testing.T) {
	adapter := NewAnthropicAdapter()

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
		FeatureAnyOf, // Anthropic supports anyOf
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

func TestAnthropicAdapter_ToCanonical(t *testing.T) {
	adapter := NewAnthropicAdapter()

	tests := []struct {
		name     string
		input    any
		wantName string
		wantErr  bool
	}{
		{
			name: "AnthropicTool pointer",
			input: &AnthropicTool{
				Name:        "test-tool",
				Description: "A test tool",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"input": map[string]any{"type": "string"},
					},
				},
			},
			wantName: "test-tool",
			wantErr:  false,
		},
		{
			name: "AnthropicTool value",
			input: AnthropicTool{
				Name:        "value-tool",
				InputSchema: map[string]any{"type": "object"},
			},
			wantName: "value-tool",
			wantErr:  false,
		},
		{
			name: "with cache control",
			input: &AnthropicTool{
				Name:        "cached-tool",
				InputSchema: map[string]any{"type": "object"},
				CacheControl: &AnthropicCacheControl{
					Type: "ephemeral",
				},
			},
			wantName: "cached-tool",
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
			input: &AnthropicTool{
				InputSchema: map[string]any{"type": "object"},
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
			if ct.SourceFormat != "anthropic" {
				t.Errorf("SourceFormat = %q, want %q", ct.SourceFormat, "anthropic")
			}
		})
	}
}

func TestAnthropicAdapter_ToCanonical_CacheControl(t *testing.T) {
	adapter := NewAnthropicAdapter()

	input := &AnthropicTool{
		Name:        "cache-test",
		InputSchema: map[string]any{"type": "object"},
		CacheControl: &AnthropicCacheControl{
			Type: "ephemeral",
		},
	}

	ct, err := adapter.ToCanonical(input)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	cc, ok := ct.SourceMeta["cache_control"].(*AnthropicCacheControl)
	if !ok || cc.Type != "ephemeral" {
		t.Errorf("SourceMeta[cache_control] = %v, want ephemeral", ct.SourceMeta["cache_control"])
	}
}

func TestAnthropicAdapter_FromCanonical(t *testing.T) {
	adapter := NewAnthropicAdapter()

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

			tool, ok := result.(*AnthropicTool)
			if !ok {
				t.Fatalf("FromCanonical() returned %T, want *AnthropicTool", result)
			}

			if tool.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", tool.Name, tt.wantName)
			}
		})
	}
}

func TestAnthropicAdapter_FromCanonical_SchemaConversion(t *testing.T) {
	adapter := NewAnthropicAdapter()

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

	tool := result.(*AnthropicTool)
	schema := tool.InputSchema

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

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("required is not []string")
	}
	if len(required) != 1 || required[0] != "name" {
		t.Errorf("required = %v, want [name]", required)
	}
}

func TestAnthropicAdapter_FromCanonical_PreservesAnyOf(t *testing.T) {
	adapter := NewAnthropicAdapter()

	input := &CanonicalTool{
		Name: "anyof-test",
		InputSchema: &JSONSchema{
			Type: "object",
			Properties: map[string]*JSONSchema{
				"value": {
					AnyOf: []*JSONSchema{
						{Type: "string"},
						{Type: "number"},
					},
				},
			},
		},
	}

	result, err := adapter.FromCanonical(input)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	tool := result.(*AnthropicTool)
	props, _ := tool.InputSchema["properties"].(map[string]any)
	valueProp, _ := props["value"].(map[string]any)

	anyOf, ok := valueProp["anyOf"].([]any)
	if !ok {
		t.Fatal("anyOf should be preserved for Anthropic")
	}
	if len(anyOf) != 2 {
		t.Errorf("anyOf length = %d, want 2", len(anyOf))
	}
}

func TestAnthropicAdapter_FromCanonical_FiltersUnsupportedFeatures(t *testing.T) {
	adapter := NewAnthropicAdapter()

	// Create a schema with features Anthropic doesn't support
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
				"exclusive": {
					OneOf: []*JSONSchema{ // Not supported
						{Const: "a"},
						{Const: "b"},
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

	tool := result.(*AnthropicTool)
	schema := tool.InputSchema

	props, _ := schema["properties"].(map[string]any)

	// Pattern and format should be filtered out
	validatedProp, _ := props["validated"].(map[string]any)
	if _, ok := validatedProp["pattern"]; ok {
		t.Error("pattern should be filtered out")
	}
	if _, ok := validatedProp["format"]; ok {
		t.Error("format should be filtered out")
	}

	// OneOf should be filtered out
	exclusiveProp, _ := props["exclusive"].(map[string]any)
	if _, ok := exclusiveProp["oneOf"]; ok {
		t.Error("oneOf should be filtered out")
	}

	// $ref should be filtered out
	refProp, _ := props["ref"].(map[string]any)
	if _, ok := refProp["$ref"]; ok {
		t.Error("$ref should be filtered out")
	}

	// $defs should be filtered out
	if _, ok := schema["$defs"]; ok {
		t.Error("$defs should be filtered out")
	}

	// But required should still be present
	required, ok := schema["required"].([]string)
	if !ok || len(required) == 0 {
		t.Error("required should be preserved")
	}
}

func TestAnthropicAdapter_FromCanonical_CacheControlPreserved(t *testing.T) {
	adapter := NewAnthropicAdapter()

	input := &CanonicalTool{
		Name: "cache-test",
		InputSchema: &JSONSchema{
			Type: "object",
		},
		SourceMeta: map[string]any{
			"cache_control": &AnthropicCacheControl{
				Type: "ephemeral",
			},
		},
	}

	result, err := adapter.FromCanonical(input)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	tool := result.(*AnthropicTool)
	if tool.CacheControl == nil || tool.CacheControl.Type != "ephemeral" {
		t.Error("cache_control should be preserved")
	}
}

func TestAnthropicAdapter_RoundTrip(t *testing.T) {
	adapter := NewAnthropicAdapter()

	original := &AnthropicTool{
		Name:        "roundtrip-test",
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
					"enum":    []any{0, 5, 10},
				},
				"choice": map[string]any{
					"anyOf": []any{
						map[string]any{"type": "string"},
						map[string]any{"type": "number"},
					},
				},
			},
			"required":             []any{"input"},
			"additionalProperties": false,
		},
		CacheControl: &AnthropicCacheControl{
			Type: "ephemeral",
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

	restored := result.(*AnthropicTool)

	// Verify basic fields
	if restored.Name != original.Name {
		t.Errorf("Name = %q, want %q", restored.Name, original.Name)
	}
	if restored.Description != original.Description {
		t.Errorf("Description = %q, want %q", restored.Description, original.Description)
	}

	// Verify cache control preserved
	if restored.CacheControl == nil || restored.CacheControl.Type != "ephemeral" {
		t.Error("cache_control should be preserved")
	}

	// Verify schema structure
	schema := restored.InputSchema
	if schema["type"] != "object" {
		t.Errorf("type = %v, want %q", schema["type"], "object")
	}

	props, _ := schema["properties"].(map[string]any)
	inputProp, _ := props["input"].(map[string]any)
	if inputProp["type"] != "string" {
		t.Errorf("input.type = %v, want %q", inputProp["type"], "string")
	}
	if inputProp["minLength"] != 1 {
		t.Errorf("input.minLength = %v, want 1", inputProp["minLength"])
	}

	// Verify anyOf is preserved (Anthropic supports it)
	choiceProp, _ := props["choice"].(map[string]any)
	anyOf, ok := choiceProp["anyOf"].([]any)
	if !ok || len(anyOf) != 2 {
		t.Error("anyOf should be preserved for Anthropic")
	}
}

func TestAnthropicAdapter_ImplementsInterface(t *testing.T) {
	var _ Adapter = (*AnthropicAdapter)(nil)
	var _ Adapter = NewAnthropicAdapter()
}

func TestAnthropicAdapter_DiffersFromOpenAI_AnyOf(t *testing.T) {
	// Verify that Anthropic supports anyOf while OpenAI doesn't
	anthropic := NewAnthropicAdapter()
	openai := NewOpenAIAdapter()

	if !anthropic.SupportsFeature(FeatureAnyOf) {
		t.Error("Anthropic should support anyOf")
	}
	if openai.SupportsFeature(FeatureAnyOf) {
		t.Error("OpenAI should not support anyOf")
	}

	// Create a schema with anyOf
	input := &CanonicalTool{
		Name: "anyof-diff-test",
		InputSchema: &JSONSchema{
			Type: "object",
			Properties: map[string]*JSONSchema{
				"value": {
					AnyOf: []*JSONSchema{
						{Type: "string"},
						{Type: "number"},
					},
				},
			},
		},
	}

	// Anthropic should preserve anyOf
	anthropicResult, _ := anthropic.FromCanonical(input)
	anthropicTool := anthropicResult.(*AnthropicTool)
	anthropicProps, _ := anthropicTool.InputSchema["properties"].(map[string]any)
	anthropicValue, _ := anthropicProps["value"].(map[string]any)
	if _, ok := anthropicValue["anyOf"]; !ok {
		t.Error("Anthropic should preserve anyOf")
	}

	// OpenAI should filter out anyOf
	openaiResult, _ := openai.FromCanonical(input)
	openaiTool := openaiResult.(*OpenAITool)
	openaiProps, _ := openaiTool.Function.Parameters["properties"].(map[string]any)
	openaiValue, _ := openaiProps["value"].(map[string]any)
	if _, ok := openaiValue["anyOf"]; ok {
		t.Error("OpenAI should filter out anyOf")
	}
}

func TestAnthropicAdapter_FilterSchema_NilInput(t *testing.T) {
	// Test that filterAnthropicSchema handles nil input
	result := filterAnthropicSchema(nil)
	if result != nil {
		t.Error("filterAnthropicSchema(nil) should return nil")
	}
}

func TestAnthropicAdapter_FilterSchema_WithItems(t *testing.T) {
	adapter := NewAnthropicAdapter()

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

	tool := result.(*AnthropicTool)
	props, _ := tool.InputSchema["properties"].(map[string]any)
	tagsProp, _ := props["tags"].(map[string]any)
	items, ok := tagsProp["items"].(map[string]any)
	if !ok {
		t.Fatal("items should be present")
	}
	if items["type"] != "string" {
		t.Errorf("items.type = %v, want string", items["type"])
	}
}
