package adapter

import (
	"sort"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jonwraymond/toolfoundation/model"
)

func TestDefaultRegistry(t *testing.T) {
	registry := DefaultRegistry()
	if registry == nil {
		t.Fatal("DefaultRegistry() returned nil")
	}

	// Verify all adapters are registered
	adapters := registry.List()
	sort.Strings(adapters)

	expected := []string{"anthropic", "mcp", "openai"}
	if len(adapters) != len(expected) {
		t.Errorf("List() = %v, want %v", adapters, expected)
	}

	for i, name := range expected {
		if adapters[i] != name {
			t.Errorf("List()[%d] = %q, want %q", i, adapters[i], name)
		}
	}

	// Verify each adapter can be retrieved
	for _, name := range expected {
		adapter, err := registry.Get(name)
		if err != nil {
			t.Errorf("Get(%q) error = %v", name, err)
		}
		if adapter.Name() != name {
			t.Errorf("Get(%q).Name() = %q, want %q", name, adapter.Name(), name)
		}
	}
}

func TestDefaultRegistry_MCPToOpenAI(t *testing.T) {
	registry := DefaultRegistry()

	tool := &model.Tool{
		Tool: mcp.Tool{
			Name:        "test-tool",
			Description: "A test tool",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"input": map[string]any{"type": "string"},
				},
				"required": []any{"input"},
			},
		},
		Namespace: "test",
		Version:   "1.0.0",
	}

	result, err := registry.Convert(tool, "mcp", "openai")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	openaiTool, ok := result.Tool.(*OpenAITool)
	if !ok {
		t.Fatalf("Convert() returned %T, want *OpenAITool", result.Tool)
	}

	if openaiTool.Type != "function" {
		t.Errorf("Type = %q, want %q", openaiTool.Type, "function")
	}
	if openaiTool.Function.Name != "test-tool" {
		t.Errorf("Function.Name = %q, want %q", openaiTool.Function.Name, "test-tool")
	}
}

func TestDefaultRegistry_MCPToAnthropic(t *testing.T) {
	registry := DefaultRegistry()

	tool := &model.Tool{
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
	}

	result, err := registry.Convert(tool, "mcp", "anthropic")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	anthropicTool, ok := result.Tool.(*AnthropicTool)
	if !ok {
		t.Fatalf("Convert() returned %T, want *AnthropicTool", result.Tool)
	}

	if anthropicTool.Name != "test-tool" {
		t.Errorf("Name = %q, want %q", anthropicTool.Name, "test-tool")
	}
}

func TestDefaultRegistry_OpenAIToAnthropic(t *testing.T) {
	registry := DefaultRegistry()

	openaiTool := &OpenAITool{
		Type: "function",
		Function: OpenAIFunction{
			Name:        "cross-format",
			Description: "Testing cross-format conversion",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"input": map[string]any{"type": "string"},
				},
			},
		},
	}

	result, err := registry.Convert(openaiTool, "openai", "anthropic")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	anthropicTool, ok := result.Tool.(*AnthropicTool)
	if !ok {
		t.Fatalf("Convert() returned %T, want *AnthropicTool", result.Tool)
	}

	if anthropicTool.Name != "cross-format" {
		t.Errorf("Name = %q, want %q", anthropicTool.Name, "cross-format")
	}
}

func TestDefaultRegistry_FeatureLossWarnings(t *testing.T) {
	registry := DefaultRegistry()

	// Create a tool with features OpenAI doesn't support
	tool := &model.Tool{
		Tool: mcp.Tool{
			Name: "advanced-tool",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"choice": map[string]any{
						"anyOf": []any{
							map[string]any{"type": "string"},
							map[string]any{"type": "number"},
						},
					},
				},
			},
		},
	}

	result, err := registry.Convert(tool, "mcp", "openai")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// Should have warnings about anyOf
	hasAnyOfWarning := false
	for _, w := range result.Warnings {
		if w.Feature == FeatureAnyOf {
			hasAnyOfWarning = true
			break
		}
	}

	if !hasAnyOfWarning {
		t.Error("Expected warning about anyOf feature loss")
	}
}

func TestDefaultRegistry_NoWarningsForSupportedFeatures(t *testing.T) {
	registry := DefaultRegistry()

	// Create a tool with only basic features
	tool := &model.Tool{
		Tool: mcp.Tool{
			Name: "basic-tool",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":    "string",
						"default": "default-value",
					},
					"count": map[string]any{
						"type":    "integer",
						"minimum": 0,
						"maximum": 100,
					},
				},
				"required": []any{"name"},
			},
		},
	}

	result, err := registry.Convert(tool, "mcp", "openai")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if len(result.Warnings) > 0 {
		t.Errorf("Expected no warnings for basic features, got: %v", result.Warnings)
	}
}

func TestDefaultRegistry_FeatureLossWithOutputSchema(t *testing.T) {
	registry := DefaultRegistry()

	// Create a tool with unsupported features in OutputSchema
	tool := &model.Tool{
		Tool: mcp.Tool{
			Name: "output-schema-tool",
			InputSchema: map[string]any{
				"type": "object",
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"result": map[string]any{
						"anyOf": []any{
							map[string]any{"type": "string"},
							map[string]any{"type": "null"},
						},
					},
				},
			},
		},
	}

	result, err := registry.Convert(tool, "mcp", "openai")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// Should have warnings about anyOf in OutputSchema
	hasAnyOfWarning := false
	for _, w := range result.Warnings {
		if w.Feature == FeatureAnyOf {
			hasAnyOfWarning = true
			break
		}
	}

	if !hasAnyOfWarning {
		t.Error("Expected warning about anyOf feature loss in OutputSchema")
	}
}

func TestDefaultRegistry_FeatureLossWithNestedDefs(t *testing.T) {
	registry := DefaultRegistry()

	// Create a tool with $defs containing unsupported features
	tool := &model.Tool{
		Tool: mcp.Tool{
			Name: "defs-tool",
			InputSchema: map[string]any{
				"type": "object",
				"$defs": map[string]any{
					"address": map[string]any{
						"type":    "object",
						"pattern": "^[A-Z]", // pattern inside $defs
					},
				},
				"properties": map[string]any{
					"addr": map[string]any{
						"$ref": "#/$defs/address",
					},
				},
			},
		},
	}

	result, err := registry.Convert(tool, "mcp", "openai")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// Should have warnings about $defs, $ref, and pattern
	hasDefsWarning := false
	hasRefWarning := false
	hasPatternWarning := false
	for _, w := range result.Warnings {
		switch w.Feature {
		case FeatureDefs:
			hasDefsWarning = true
		case FeatureRef:
			hasRefWarning = true
		case FeaturePattern:
			hasPatternWarning = true
		}
	}

	if !hasDefsWarning {
		t.Error("Expected warning about $defs feature loss")
	}
	if !hasRefWarning {
		t.Error("Expected warning about $ref feature loss")
	}
	if !hasPatternWarning {
		t.Error("Expected warning about pattern feature loss inside $defs")
	}
}

func TestDefaultRegistry_FeatureLossWithCombinators(t *testing.T) {
	registry := DefaultRegistry()

	// Create a tool with oneOf, allOf, and not (all unsupported by OpenAI)
	tool := &model.Tool{
		Tool: mcp.Tool{
			Name: "combinator-tool",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
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
						"not": map[string]any{
							"const": "",
						},
					},
				},
			},
		},
	}

	result, err := registry.Convert(tool, "mcp", "openai")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// Should have warnings about oneOf, allOf, not, and const inside them
	hasOneOfWarning := false
	hasAllOfWarning := false
	hasNotWarning := false
	for _, w := range result.Warnings {
		switch w.Feature {
		case FeatureOneOf:
			hasOneOfWarning = true
		case FeatureAllOf:
			hasAllOfWarning = true
		case FeatureNot:
			hasNotWarning = true
		}
	}

	if !hasOneOfWarning {
		t.Error("Expected warning about oneOf feature loss")
	}
	if !hasAllOfWarning {
		t.Error("Expected warning about allOf feature loss")
	}
	if !hasNotWarning {
		t.Error("Expected warning about not feature loss")
	}
}

func TestDefaultRegistry_FeatureLossWithItems(t *testing.T) {
	registry := DefaultRegistry()

	// Create a tool with unsupported features in array items
	tool := &model.Tool{
		Tool: mcp.Tool{
			Name: "items-tool",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"tags": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type":    "string",
							"pattern": "^[a-z]+$", // pattern inside items
						},
					},
				},
			},
		},
	}

	result, err := registry.Convert(tool, "mcp", "openai")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// Should have warnings about pattern in items
	hasPatternWarning := false
	for _, w := range result.Warnings {
		if w.Feature == FeaturePattern {
			hasPatternWarning = true
			break
		}
	}

	if !hasPatternWarning {
		t.Error("Expected warning about pattern feature loss in array items")
	}
}
