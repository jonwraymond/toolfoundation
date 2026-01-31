package model

import (
    "encoding/json"
    "strings"
    "testing"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// contractValidator is a minimal validator used to exercise SchemaValidator contracts.
type contractValidator struct{
    lastSchema any
    lastInstance any
    err error
}

func (v *contractValidator) Validate(schema any, instance any) error {
    v.lastSchema = schema
    v.lastInstance = instance
    return v.err
}

func (v *contractValidator) ValidateInput(tool *Tool, args any) error {
    if tool == nil || tool.InputSchema == nil {
        return ErrInvalidSchema
    }
    return v.Validate(tool.InputSchema, args)
}

func (v *contractValidator) ValidateOutput(tool *Tool, result any) error {
    if tool == nil || tool.OutputSchema == nil {
        return nil
    }
    return v.Validate(tool.OutputSchema, result)
}

func TestSchemaValidator_Contract(t *testing.T) {
    v := &contractValidator{}

    t.Run("ValidateInput uses InputSchema", func(t *testing.T) {
        schema := json.RawMessage(`{"type":"object"}`)
        tool := &Tool{Tool: mcp.Tool{InputSchema: schema}}
        args := map[string]any{"x": "y"}
        if err := v.ValidateInput(tool, args); err != nil {
            t.Fatalf("expected nil, got %v", err)
        }
        if string(v.lastSchema.(json.RawMessage)) != string(schema) {
            t.Fatalf("expected InputSchema passed to Validate")
        }
    })

    t.Run("ValidateOutput is no-op when OutputSchema nil", func(t *testing.T) {
        tool := &Tool{}
        if err := v.ValidateOutput(tool, map[string]any{"ok": true}); err != nil {
            t.Fatalf("expected nil, got %v", err)
        }
    })

    t.Run("ValidateOutput uses OutputSchema when present", func(t *testing.T) {
        schema := json.RawMessage(`{"type":"object"}`)
        tool := &Tool{Tool: mcp.Tool{OutputSchema: schema}}
        if err := v.ValidateOutput(tool, map[string]any{"ok": true}); err != nil {
            t.Fatalf("expected nil, got %v", err)
        }
        if string(v.lastSchema.(json.RawMessage)) != string(schema) {
            t.Fatalf("expected OutputSchema passed to Validate")
        }
    })

    t.Run("ValidateInput errors on nil tool or nil schema", func(t *testing.T) {
        if err := v.ValidateInput(nil, nil); err == nil {
            t.Fatalf("expected error")
        }
        if err := v.ValidateInput(&Tool{}, nil); err == nil {
            t.Fatalf("expected error")
        }
    })
}

func TestDefaultValidator_Contract(t *testing.T) {
    v := NewDefaultValidator()

	t.Run("ValidateInput returns ErrInvalidSchema on nil InputSchema", func(t *testing.T) {
		err := v.ValidateInput(&Tool{}, map[string]any{})
		if err == nil || !strings.Contains(err.Error(), ErrInvalidSchema.Error()) {
			t.Fatalf("expected ErrInvalidSchema, got %v", err)
		}
	})

	t.Run("ValidateInput returns ErrInvalidSchema on nil tool", func(t *testing.T) {
		err := v.ValidateInput(nil, map[string]any{})
		if err == nil || !strings.Contains(err.Error(), ErrInvalidSchema.Error()) {
			t.Fatalf("expected ErrInvalidSchema, got %v", err)
		}
	})

	t.Run("ValidateOutput returns nil on nil OutputSchema", func(t *testing.T) {
		if err := v.ValidateOutput(&Tool{}, map[string]any{}); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("ValidateOutput returns ErrInvalidSchema on nil tool", func(t *testing.T) {
		err := v.ValidateOutput(nil, map[string]any{})
		if err == nil || !strings.Contains(err.Error(), ErrInvalidSchema.Error()) {
			t.Fatalf("expected ErrInvalidSchema, got %v", err)
		}
	})
}
