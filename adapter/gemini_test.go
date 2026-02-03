package adapter

import "testing"

func TestNewGeminiAdapter(t *testing.T) {
	adapter := NewGeminiAdapter()
	if adapter == nil {
		t.Fatal("NewGeminiAdapter() returned nil")
	}
}

func TestGeminiAdapter_Name(t *testing.T) {
	adapter := NewGeminiAdapter()
	if adapter.Name() != "gemini" {
		t.Errorf("Name() = %q, want %q", adapter.Name(), "gemini")
	}
}

func TestGeminiAdapter_SupportsFeature(t *testing.T) {
	adapter := NewGeminiAdapter()

	supported := []SchemaFeature{
		FeatureRef,
		FeatureDefs,
		FeatureAnyOf,
		FeaturePattern,
		FeatureFormat,
		FeatureAdditionalProperties,
		FeatureMinimum,
		FeatureMaximum,
		FeatureMinLength,
		FeatureMaxLength,
		FeatureMinItems,
		FeatureMaxItems,
		FeatureMinProperties,
		FeatureMaxProperties,
		FeatureEnum,
		FeatureDefault,
		FeatureTitle,
		FeatureNullable,
	}
	for _, feature := range supported {
		if !adapter.SupportsFeature(feature) {
			t.Errorf("SupportsFeature(%v) = false, want true", feature)
		}
	}

	unsupported := []SchemaFeature{
		FeatureConst,
		FeatureMultipleOf,
		FeatureOneOf,
		FeatureAllOf,
		FeatureNot,
		FeatureExamples,
		FeatureUniqueItems,
		FeatureDeprecated,
		FeatureReadOnly,
		FeatureWriteOnly,
	}
	for _, feature := range unsupported {
		if adapter.SupportsFeature(feature) {
			t.Errorf("SupportsFeature(%v) = true, want false", feature)
		}
	}
}

func TestGeminiAdapter_ToCanonical(t *testing.T) {
	adapter := NewGeminiAdapter()

	input := &GeminiFunctionDeclaration{
		Name:        "lookup",
		Description: "Lookup records",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "string"},
			},
		},
	}

	ct, err := adapter.ToCanonical(input)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}
	if ct.Name != "lookup" {
		t.Errorf("Name = %q, want %q", ct.Name, "lookup")
	}
	if ct.InputSchema == nil || ct.InputSchema.Type != "object" {
		t.Errorf("InputSchema.Type = %v, want %q", ct.InputSchema.Type, "object")
	}
}

func TestGeminiAdapter_ToCanonical_ToolWrapper(t *testing.T) {
	adapter := NewGeminiAdapter()

	input := &GeminiTool{
		FunctionDeclarations: []GeminiFunctionDeclaration{
			{Name: "wrapped", Parameters: map[string]any{"type": "object"}},
		},
	}

	ct, err := adapter.ToCanonical(input)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}
	if ct.Name != "wrapped" {
		t.Errorf("Name = %q, want %q", ct.Name, "wrapped")
	}
}

func TestGeminiAdapter_ToCanonical_MultipleFunctionsError(t *testing.T) {
	adapter := NewGeminiAdapter()

	input := &GeminiTool{
		FunctionDeclarations: []GeminiFunctionDeclaration{
			{Name: "one"},
			{Name: "two"},
		},
	}

	if _, err := adapter.ToCanonical(input); err == nil {
		t.Error("expected error for multiple function declarations")
	}
}

func TestGeminiAdapter_FromCanonical(t *testing.T) {
	adapter := NewGeminiAdapter()

	ct := &CanonicalTool{
		Name:        "lookup",
		Description: "Lookup records",
		InputSchema: &JSONSchema{Type: "object"},
	}

	raw, err := adapter.FromCanonical(ct)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	tool := raw.(*GeminiTool)
	if len(tool.FunctionDeclarations) != 1 {
		t.Fatalf("FunctionDeclarations length = %d, want 1", len(tool.FunctionDeclarations))
	}
	if tool.FunctionDeclarations[0].Name != "lookup" {
		t.Errorf("Name = %q, want %q", tool.FunctionDeclarations[0].Name, "lookup")
	}
}
