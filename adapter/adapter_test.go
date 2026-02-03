package adapter

import (
	"errors"
	"testing"
)

func TestSchemaFeature_String(t *testing.T) {
	tests := []struct {
		feature SchemaFeature
		want    string
	}{
		{FeatureRef, "$ref"},
		{FeatureDefs, "$defs"},
		{FeatureAnyOf, "anyOf"},
		{FeatureOneOf, "oneOf"},
		{FeatureAllOf, "allOf"},
		{FeatureNot, "not"},
		{FeaturePattern, "pattern"},
		{FeatureFormat, "format"},
		{FeatureAdditionalProperties, "additionalProperties"},
		{FeatureMinimum, "minimum"},
		{FeatureMaximum, "maximum"},
		{FeatureMinLength, "minLength"},
		{FeatureMaxLength, "maxLength"},
		{FeatureEnum, "enum"},
		{FeatureConst, "const"},
		{FeatureDefault, "default"},
		{FeatureTitle, "title"},
		{FeatureExamples, "examples"},
		{FeatureMultipleOf, "multipleOf"},
		{FeatureMinItems, "minItems"},
		{FeatureMaxItems, "maxItems"},
		{FeatureMinProperties, "minProperties"},
		{FeatureMaxProperties, "maxProperties"},
		{FeatureUniqueItems, "uniqueItems"},
		{FeatureNullable, "nullable"},
		{FeatureDeprecated, "deprecated"},
		{FeatureReadOnly, "readOnly"},
		{FeatureWriteOnly, "writeOnly"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.feature.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSchemaFeature_String_Unknown(t *testing.T) {
	unknown := SchemaFeature(999)
	got := unknown.String()
	if got == "" {
		t.Error("String() for unknown feature should not be empty")
	}
}

func TestAllFeatures(t *testing.T) {
	features := AllFeatures()

	// Check that it returns a non-empty slice
	if len(features) == 0 {
		t.Fatal("AllFeatures() returned empty slice")
	}

	// Check that all known features are included
	knownFeatures := []SchemaFeature{
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

	for _, known := range knownFeatures {
		found := false
		for _, f := range features {
			if f == known {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("AllFeatures() missing %s", known)
		}
	}

	// Verify stable ordering (calling twice returns same order)
	features2 := AllFeatures()
	if len(features) != len(features2) {
		t.Fatal("AllFeatures() returned different lengths on second call")
	}
	for i := range features {
		if features[i] != features2[i] {
			t.Errorf("AllFeatures() ordering not stable at index %d", i)
		}
	}
}

func TestConversionError_Error(t *testing.T) {
	cause := errors.New("underlying error")
	err := &ConversionError{
		Adapter:   "mcp",
		Direction: "to_canonical",
		Cause:     cause,
	}

	got := err.Error()

	// Should contain adapter name
	if !containsString(got, "mcp") {
		t.Errorf("Error() = %q, should contain adapter name", got)
	}

	// Should contain direction
	if !containsString(got, "to_canonical") {
		t.Errorf("Error() = %q, should contain direction", got)
	}

	// Should contain cause message
	if !containsString(got, "underlying error") {
		t.Errorf("Error() = %q, should contain cause message", got)
	}
}

func TestConversionError_Error_FromCanonical(t *testing.T) {
	cause := errors.New("output error")
	err := &ConversionError{
		Adapter:   "openai",
		Direction: "from_canonical",
		Cause:     cause,
	}

	got := err.Error()

	if !containsString(got, "openai") {
		t.Errorf("Error() = %q, should contain adapter name", got)
	}
	if !containsString(got, "from_canonical") {
		t.Errorf("Error() = %q, should contain direction", got)
	}
}

func TestConversionError_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := &ConversionError{
		Adapter:   "anthropic",
		Direction: "to_canonical",
		Cause:     cause,
	}

	unwrapped := err.Unwrap()

	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}

	// Verify errors.Is works
	if !errors.Is(err, cause) {
		t.Error("errors.Is(err, cause) = false, want true")
	}
}

func TestConversionError_Unwrap_NilCause(t *testing.T) {
	err := &ConversionError{
		Adapter:   "mcp",
		Direction: "to_canonical",
		Cause:     nil,
	}

	unwrapped := err.Unwrap()

	if unwrapped != nil {
		t.Errorf("Unwrap() = %v, want nil", unwrapped)
	}
}

func TestFeatureLossWarning_String(t *testing.T) {
	warning := FeatureLossWarning{
		Feature:     FeatureRef,
		FromAdapter: "mcp",
		ToAdapter:   "openai",
	}

	got := warning.String()

	// Should include feature name
	if !containsString(got, "$ref") {
		t.Errorf("String() = %q, should contain feature name", got)
	}

	// Should include adapter names
	if !containsString(got, "mcp") {
		t.Errorf("String() = %q, should contain from adapter", got)
	}
	if !containsString(got, "openai") {
		t.Errorf("String() = %q, should contain to adapter", got)
	}
}

func TestFeatureLossWarning_String_AllFeatures(t *testing.T) {
	// Test that String() works for all features
	for _, feature := range AllFeatures() {
		warning := FeatureLossWarning{
			Feature:     feature,
			FromAdapter: "source",
			ToAdapter:   "target",
		}

		got := warning.String()
		if got == "" {
			t.Errorf("String() for feature %v returned empty string", feature)
		}
	}
}

func TestAdapter_Interface(t *testing.T) {
	// This test verifies the Adapter interface has the expected methods
	// by creating a mock implementation
	var _ Adapter = &mockAdapter{}
}

// mockAdapter is a test implementation of Adapter
type mockAdapter struct {
	name              string
	toCanonicalFunc   func(any) (*CanonicalTool, error)
	fromCanonicalFunc func(*CanonicalTool) (any, error)
	supportsFunc      func(SchemaFeature) bool
}

func (m *mockAdapter) Name() string {
	return m.name
}

func (m *mockAdapter) ToCanonical(raw any) (*CanonicalTool, error) {
	if m.toCanonicalFunc != nil {
		return m.toCanonicalFunc(raw)
	}
	return nil, nil
}

func (m *mockAdapter) FromCanonical(tool *CanonicalTool) (any, error) {
	if m.fromCanonicalFunc != nil {
		return m.fromCanonicalFunc(tool)
	}
	return nil, nil
}

func (m *mockAdapter) SupportsFeature(feature SchemaFeature) bool {
	if m.supportsFunc != nil {
		return m.supportsFunc(feature)
	}
	return false
}

// containsString checks if s contains substr
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
