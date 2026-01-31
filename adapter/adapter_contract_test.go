package adapter

import (
    "errors"
    "testing"
)

var errContract = errors.New("contract error")

type contractAdapter struct {
    name string
    supported map[SchemaFeature]bool
}

func (a *contractAdapter) Name() string { return a.name }
func (a *contractAdapter) ToCanonical(raw any) (*CanonicalTool, error) {
    if raw == nil {
        return nil, &ConversionError{Adapter: a.name, Direction: "to_canonical", Cause: errContract}
    }
    return &CanonicalTool{Name: "ok"}, nil
}
func (a *contractAdapter) FromCanonical(tool *CanonicalTool) (any, error) {
    if tool == nil {
        return nil, &ConversionError{Adapter: a.name, Direction: "from_canonical", Cause: errContract}
    }
    return map[string]any{"name": tool.Name}, nil
}
func (a *contractAdapter) SupportsFeature(feature SchemaFeature) bool {
    return a.supported[feature]
}

func TestAdapter_Contract(t *testing.T) {
    a := &contractAdapter{name: "contract", supported: map[SchemaFeature]bool{}}

    t.Run("Name is stable", func(t *testing.T) {
        if a.Name() != "contract" {
            t.Fatalf("unexpected name: %s", a.Name())
        }
    })

    t.Run("ToCanonical returns ConversionError for invalid input", func(t *testing.T) {
        _, err := a.ToCanonical(nil)
        if err == nil {
            t.Fatalf("expected error")
        }
        if _, ok := err.(*ConversionError); !ok {
            t.Fatalf("expected ConversionError, got %T", err)
        }
    })

    t.Run("FromCanonical returns ConversionError for nil tool", func(t *testing.T) {
        _, err := a.FromCanonical(nil)
        if err == nil {
            t.Fatalf("expected error")
        }
        if _, ok := err.(*ConversionError); !ok {
            t.Fatalf("expected ConversionError, got %T", err)
        }
    })
}
