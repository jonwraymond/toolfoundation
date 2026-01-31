package adapter

import (
	"reflect"
	"testing"
	"time"
)

func TestCanonicalTool_ID_WithNamespace(t *testing.T) {
	tool := &CanonicalTool{
		Namespace: "myns",
		Name:      "mytool",
	}

	got := tool.ID()
	want := "myns:mytool"

	if got != want {
		t.Errorf("ID() = %q, want %q", got, want)
	}
}

func TestCanonicalTool_ID_WithoutNamespace(t *testing.T) {
	tool := &CanonicalTool{
		Name: "mytool",
	}

	got := tool.ID()
	want := "mytool"

	if got != want {
		t.Errorf("ID() = %q, want %q", got, want)
	}
}

func TestCanonicalTool_Validate_Valid(t *testing.T) {
	tool := &CanonicalTool{
		Name: "mytool",
		InputSchema: &JSONSchema{
			Type: "object",
		},
	}

	err := tool.Validate()
	if err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestCanonicalTool_Validate_MissingName(t *testing.T) {
	tool := &CanonicalTool{
		InputSchema: &JSONSchema{
			Type: "object",
		},
	}

	err := tool.Validate()
	if err == nil {
		t.Error("Validate() = nil, want error for missing name")
	}
}

func TestCanonicalTool_Validate_MissingInputSchema(t *testing.T) {
	tool := &CanonicalTool{
		Name: "mytool",
	}

	err := tool.Validate()
	if err == nil {
		t.Error("Validate() = nil, want error for missing input schema")
	}
}

func TestJSONSchema_DeepCopy_Nil(t *testing.T) {
	var s *JSONSchema
	got := s.DeepCopy()
	if got != nil {
		t.Errorf("DeepCopy() on nil = %v, want nil", got)
	}
}

func TestJSONSchema_DeepCopy_Simple(t *testing.T) {
	min := 1.0
	max := 100.0
	minLen := 1
	maxLen := 50
	additionalProps := false

	original := &JSONSchema{
		Type:                 "object",
		Description:          "A test schema",
		Pattern:              "^[a-z]+$",
		Format:               "email",
		Minimum:              &min,
		Maximum:              &max,
		MinLength:            &minLen,
		MaxLength:            &maxLen,
		AdditionalProperties: &additionalProps,
	}

	copied := original.DeepCopy()

	// Verify values are equal
	if copied.Type != original.Type {
		t.Errorf("Type = %q, want %q", copied.Type, original.Type)
	}
	if copied.Description != original.Description {
		t.Errorf("Description = %q, want %q", copied.Description, original.Description)
	}
	if *copied.Minimum != *original.Minimum {
		t.Errorf("Minimum = %v, want %v", *copied.Minimum, *original.Minimum)
	}

	// Verify no aliasing of pointers
	if copied.Minimum == original.Minimum {
		t.Error("Minimum pointer is aliased, want deep copy")
	}
	if copied.AdditionalProperties == original.AdditionalProperties {
		t.Error("AdditionalProperties pointer is aliased, want deep copy")
	}
}

func TestJSONSchema_DeepCopy_NestedProperties(t *testing.T) {
	original := &JSONSchema{
		Type: "object",
		Properties: map[string]*JSONSchema{
			"name": {
				Type:        "string",
				Description: "The name",
			},
			"age": {
				Type: "integer",
			},
		},
		Required: []string{"name"},
	}

	copied := original.DeepCopy()

	// Verify properties are equal
	if len(copied.Properties) != len(original.Properties) {
		t.Errorf("Properties length = %d, want %d", len(copied.Properties), len(original.Properties))
	}

	// Verify no aliasing of Properties map
	if &copied.Properties == &original.Properties {
		t.Error("Properties map is aliased, want deep copy")
	}

	// Verify nested schema is not aliased
	if copied.Properties["name"] == original.Properties["name"] {
		t.Error("Nested schema is aliased, want deep copy")
	}

	// Verify Required slice is not aliased
	original.Required[0] = "modified"
	if copied.Required[0] == "modified" {
		t.Error("Required slice is aliased, want deep copy")
	}
}

func TestJSONSchema_DeepCopy_Items(t *testing.T) {
	original := &JSONSchema{
		Type: "array",
		Items: &JSONSchema{
			Type: "string",
		},
	}

	copied := original.DeepCopy()

	if copied.Items == original.Items {
		t.Error("Items is aliased, want deep copy")
	}
	if copied.Items.Type != original.Items.Type {
		t.Errorf("Items.Type = %q, want %q", copied.Items.Type, original.Items.Type)
	}
}

func TestJSONSchema_DeepCopy_Defs(t *testing.T) {
	original := &JSONSchema{
		Type: "object",
		Ref:  "#/$defs/Person",
		Defs: map[string]*JSONSchema{
			"Person": {
				Type: "object",
				Properties: map[string]*JSONSchema{
					"name": {Type: "string"},
				},
			},
		},
	}

	copied := original.DeepCopy()

	if copied.Defs == nil {
		t.Fatal("Defs is nil after copy")
	}
	if copied.Defs["Person"] == original.Defs["Person"] {
		t.Error("Defs entry is aliased, want deep copy")
	}
}

func TestJSONSchema_DeepCopy_Combinators(t *testing.T) {
	original := &JSONSchema{
		AnyOf: []*JSONSchema{
			{Type: "string"},
			{Type: "integer"},
		},
		OneOf: []*JSONSchema{
			{Type: "boolean"},
		},
		AllOf: []*JSONSchema{
			{Type: "object"},
		},
		Not: &JSONSchema{Type: "null"},
	}

	copied := original.DeepCopy()

	if len(copied.AnyOf) != len(original.AnyOf) {
		t.Errorf("AnyOf length = %d, want %d", len(copied.AnyOf), len(original.AnyOf))
	}
	if copied.AnyOf[0] == original.AnyOf[0] {
		t.Error("AnyOf entry is aliased, want deep copy")
	}
	if copied.Not == original.Not {
		t.Error("Not is aliased, want deep copy")
	}
}

func TestJSONSchema_DeepCopy_Enum(t *testing.T) {
	original := &JSONSchema{
		Type: "string",
		Enum: []any{"red", "green", "blue"},
	}

	copied := original.DeepCopy()

	if len(copied.Enum) != len(original.Enum) {
		t.Errorf("Enum length = %d, want %d", len(copied.Enum), len(original.Enum))
	}

	// Modify original to verify no aliasing
	original.Enum[0] = "modified"
	if copied.Enum[0] == "modified" {
		t.Error("Enum slice is aliased, want deep copy")
	}
}

func TestJSONSchema_ToMap_Empty(t *testing.T) {
	s := &JSONSchema{}
	got := s.ToMap()

	if len(got) != 0 {
		t.Errorf("ToMap() = %v, want empty map", got)
	}
}

func TestJSONSchema_ToMap_Simple(t *testing.T) {
	s := &JSONSchema{
		Type:        "string",
		Description: "A string field",
	}

	got := s.ToMap()

	if got["type"] != "string" {
		t.Errorf("type = %v, want %q", got["type"], "string")
	}
	if got["description"] != "A string field" {
		t.Errorf("description = %v, want %q", got["description"], "A string field")
	}
}

func TestJSONSchema_ToMap_OmitsZeroFields(t *testing.T) {
	s := &JSONSchema{
		Type: "integer",
		// All other fields are zero values
	}

	got := s.ToMap()

	// Should only have "type"
	if len(got) != 1 {
		t.Errorf("ToMap() has %d fields, want 1 (only type)", len(got))
	}
	if _, exists := got["description"]; exists {
		t.Error("description should be omitted when empty")
	}
	if _, exists := got["minimum"]; exists {
		t.Error("minimum should be omitted when nil")
	}
}

func TestJSONSchema_ToMap_WithConstraints(t *testing.T) {
	min := 0.0
	max := 100.0
	minLen := 1
	maxLen := 50

	s := &JSONSchema{
		Type:      "number",
		Minimum:   &min,
		Maximum:   &max,
		MinLength: &minLen,
		MaxLength: &maxLen,
		Pattern:   "^\\d+$",
	}

	got := s.ToMap()

	if got["minimum"] != min {
		t.Errorf("minimum = %v, want %v", got["minimum"], min)
	}
	if got["maximum"] != max {
		t.Errorf("maximum = %v, want %v", got["maximum"], max)
	}
	if got["minLength"] != minLen {
		t.Errorf("minLength = %v, want %v", got["minLength"], minLen)
	}
	if got["maxLength"] != maxLen {
		t.Errorf("maxLength = %v, want %v", got["maxLength"], maxLen)
	}
	if got["pattern"] != "^\\d+$" {
		t.Errorf("pattern = %v, want %q", got["pattern"], "^\\d+$")
	}
}

func TestJSONSchema_ToMap_NestedProperties(t *testing.T) {
	s := &JSONSchema{
		Type: "object",
		Properties: map[string]*JSONSchema{
			"name": {
				Type:        "string",
				Description: "The name",
			},
		},
		Required: []string{"name"},
	}

	got := s.ToMap()

	props, ok := got["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties is not map[string]any, got %T", got["properties"])
	}

	nameSchema, ok := props["name"].(map[string]any)
	if !ok {
		t.Fatalf("properties.name is not map[string]any, got %T", props["name"])
	}

	if nameSchema["type"] != "string" {
		t.Errorf("properties.name.type = %v, want %q", nameSchema["type"], "string")
	}

	required, ok := got["required"].([]string)
	if !ok {
		t.Fatalf("required is not []string, got %T", got["required"])
	}
	if len(required) != 1 || required[0] != "name" {
		t.Errorf("required = %v, want [\"name\"]", required)
	}
}

func TestJSONSchema_ToMap_Ref(t *testing.T) {
	s := &JSONSchema{
		Ref: "#/$defs/Person",
	}

	got := s.ToMap()

	if got["$ref"] != "#/$defs/Person" {
		t.Errorf("$ref = %v, want %q", got["$ref"], "#/$defs/Person")
	}
}

func TestJSONSchema_ToMap_Defs(t *testing.T) {
	s := &JSONSchema{
		Type: "object",
		Defs: map[string]*JSONSchema{
			"Person": {
				Type: "object",
			},
		},
	}

	got := s.ToMap()

	defs, ok := got["$defs"].(map[string]any)
	if !ok {
		t.Fatalf("$defs is not map[string]any, got %T", got["$defs"])
	}

	person, ok := defs["Person"].(map[string]any)
	if !ok {
		t.Fatalf("$defs.Person is not map[string]any, got %T", defs["Person"])
	}

	if person["type"] != "object" {
		t.Errorf("$defs.Person.type = %v, want %q", person["type"], "object")
	}
}

func TestJSONSchema_ToMap_AdditionalProperties(t *testing.T) {
	f := false
	s := &JSONSchema{
		Type:                 "object",
		AdditionalProperties: &f,
	}

	got := s.ToMap()

	if got["additionalProperties"] != false {
		t.Errorf("additionalProperties = %v, want false", got["additionalProperties"])
	}
}

func TestJSONSchema_ToMap_Combinators(t *testing.T) {
	s := &JSONSchema{
		AnyOf: []*JSONSchema{
			{Type: "string"},
			{Type: "integer"},
		},
	}

	got := s.ToMap()

	anyOf, ok := got["anyOf"].([]any)
	if !ok {
		t.Fatalf("anyOf is not []any, got %T", got["anyOf"])
	}

	if len(anyOf) != 2 {
		t.Fatalf("anyOf length = %d, want 2", len(anyOf))
	}

	first, ok := anyOf[0].(map[string]any)
	if !ok {
		t.Fatalf("anyOf[0] is not map[string]any, got %T", anyOf[0])
	}

	if first["type"] != "string" {
		t.Errorf("anyOf[0].type = %v, want %q", first["type"], "string")
	}
}

func TestCanonicalTool_Fields(t *testing.T) {
	// Verify all expected fields exist on CanonicalTool
	tool := CanonicalTool{
		Namespace:      "ns",
		Name:           "test",
		Version:        "1.0.0",
		Description:    "A test tool",
		Category:       "testing",
		Tags:           []string{"test", "example"},
		InputSchema:    &JSONSchema{Type: "object"},
		OutputSchema:   &JSONSchema{Type: "object"},
		Timeout:        30 * time.Second,
		SourceFormat:   "mcp",
		SourceMeta:     map[string]any{"key": "value"},
		RequiredScopes: []string{"read", "write"},
	}

	// Just verify we can access all fields
	if tool.Namespace != "ns" {
		t.Error("Namespace field not accessible")
	}
	if tool.Timeout != 30*time.Second {
		t.Error("Timeout field not accessible")
	}
	if len(tool.RequiredScopes) != 2 {
		t.Error("RequiredScopes field not accessible")
	}
}

func TestJSONSchema_Fields(t *testing.T) {
	// Verify all expected fields exist on JSONSchema
	min := 1.0
	max := 100.0
	minLen := 1
	maxLen := 50
	additionalProps := false

	schema := JSONSchema{
		Type:                 "object",
		Properties:           map[string]*JSONSchema{"field": {Type: "string"}},
		Required:             []string{"field"},
		Items:                &JSONSchema{Type: "string"},
		Description:          "A schema",
		Enum:                 []any{"a", "b"},
		Const:                "constant",
		Default:              "default",
		Minimum:              &min,
		Maximum:              &max,
		MinLength:            &minLen,
		MaxLength:            &maxLen,
		Pattern:              "^.*$",
		Format:               "email",
		Ref:                  "#/$defs/Something",
		Defs:                 map[string]*JSONSchema{"Something": {Type: "object"}},
		AdditionalProperties: &additionalProps,
		AnyOf:                []*JSONSchema{{Type: "string"}},
		OneOf:                []*JSONSchema{{Type: "integer"}},
		AllOf:                []*JSONSchema{{Type: "object"}},
		Not:                  &JSONSchema{Type: "null"},
	}

	// Verify field types are correct
	if reflect.TypeOf(schema.Properties).String() != "map[string]*adapter.JSONSchema" {
		t.Error("Properties has wrong type")
	}
	if reflect.TypeOf(schema.Enum).String() != "[]interface {}" {
		t.Error("Enum has wrong type")
	}
}
