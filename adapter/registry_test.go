package adapter

import (
	"errors"
	"sort"
	"sync"
	"testing"
)

func TestRegistry_Register_Success(t *testing.T) {
	r := NewRegistry()
	adapter := &mockAdapter{name: "test"}

	err := r.Register(adapter)

	if err != nil {
		t.Errorf("Register() = %v, want nil", err)
	}
}

func TestRegistry_Register_Duplicate(t *testing.T) {
	r := NewRegistry()
	adapter1 := &mockAdapter{name: "test"}
	adapter2 := &mockAdapter{name: "test"}

	_ = r.Register(adapter1)
	err := r.Register(adapter2)

	if err == nil {
		t.Error("Register() duplicate = nil, want error")
	}
}

func TestRegistry_Get_Found(t *testing.T) {
	r := NewRegistry()
	adapter := &mockAdapter{name: "test"}
	_ = r.Register(adapter)

	got, err := r.Get("test")

	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}
	if got != adapter {
		t.Errorf("Get() = %v, want %v", got, adapter)
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r := NewRegistry()

	_, err := r.Get("nonexistent")

	if err == nil {
		t.Error("Get() error = nil, want error for not found")
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(&mockAdapter{name: "alpha"})
	_ = r.Register(&mockAdapter{name: "beta"})
	_ = r.Register(&mockAdapter{name: "gamma"})

	got := r.List()

	if len(got) != 3 {
		t.Fatalf("List() length = %d, want 3", len(got))
	}

	// Sort for stable comparison
	sort.Strings(got)
	want := []string{"alpha", "beta", "gamma"}
	for i, name := range want {
		if got[i] != name {
			t.Errorf("List()[%d] = %q, want %q", i, got[i], name)
		}
	}
}

func TestRegistry_List_Empty(t *testing.T) {
	r := NewRegistry()

	got := r.List()

	if len(got) != 0 {
		t.Errorf("List() = %v, want empty slice", got)
	}
}

func TestRegistry_Unregister_Success(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(&mockAdapter{name: "test"})

	err := r.Unregister("test")

	if err != nil {
		t.Errorf("Unregister() = %v, want nil", err)
	}

	// Verify it's gone
	_, err = r.Get("test")
	if err == nil {
		t.Error("Get() after Unregister() should return error")
	}
}

func TestRegistry_Unregister_NotFound(t *testing.T) {
	r := NewRegistry()

	err := r.Unregister("nonexistent")

	if err == nil {
		t.Error("Unregister() nonexistent = nil, want error")
	}
}

func TestRegistry_Convert_Success(t *testing.T) {
	r := NewRegistry()

	// Source adapter that converts input to canonical
	source := &mockAdapter{
		name: "source",
		toCanonicalFunc: func(raw any) (*CanonicalTool, error) {
			return &CanonicalTool{
				Name:        "test-tool",
				Description: "A test tool",
				InputSchema: &JSONSchema{Type: "object"},
			}, nil
		},
		supportsFunc: func(f SchemaFeature) bool { return true },
	}

	// Target adapter that converts canonical to output
	target := &mockAdapter{
		name: "target",
		fromCanonicalFunc: func(tool *CanonicalTool) (any, error) {
			return map[string]string{
				"name":        tool.Name,
				"description": tool.Description,
			}, nil
		},
		supportsFunc: func(f SchemaFeature) bool { return true },
	}

	_ = r.Register(source)
	_ = r.Register(target)

	result, err := r.Convert("input", "source", "target")

	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	output, ok := result.Tool.(map[string]string)
	if !ok {
		t.Fatalf("Convert().Tool type = %T, want map[string]string", result.Tool)
	}
	if output["name"] != "test-tool" {
		t.Errorf("output name = %q, want %q", output["name"], "test-tool")
	}
}

func TestRegistry_Convert_MissingSourceAdapter(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(&mockAdapter{name: "target"})

	_, err := r.Convert("input", "source", "target")

	if err == nil {
		t.Error("Convert() with missing source = nil, want error")
	}
}

func TestRegistry_Convert_MissingTargetAdapter(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(&mockAdapter{name: "source"})

	_, err := r.Convert("input", "source", "target")

	if err == nil {
		t.Error("Convert() with missing target = nil, want error")
	}
}

func TestRegistry_Convert_ToCanonicalError(t *testing.T) {
	r := NewRegistry()

	source := &mockAdapter{
		name: "source",
		toCanonicalFunc: func(raw any) (*CanonicalTool, error) {
			return nil, errors.New("conversion failed")
		},
	}
	target := &mockAdapter{name: "target"}

	_ = r.Register(source)
	_ = r.Register(target)

	_, err := r.Convert("input", "source", "target")

	if err == nil {
		t.Error("Convert() with ToCanonical error = nil, want error")
	}
}

func TestRegistry_Convert_FromCanonicalError(t *testing.T) {
	r := NewRegistry()

	source := &mockAdapter{
		name: "source",
		toCanonicalFunc: func(raw any) (*CanonicalTool, error) {
			return &CanonicalTool{
				Name:        "test",
				InputSchema: &JSONSchema{Type: "object"},
			}, nil
		},
	}
	target := &mockAdapter{
		name: "target",
		fromCanonicalFunc: func(tool *CanonicalTool) (any, error) {
			return nil, errors.New("output failed")
		},
	}

	_ = r.Register(source)
	_ = r.Register(target)

	_, err := r.Convert("input", "source", "target")

	if err == nil {
		t.Error("Convert() with FromCanonical error = nil, want error")
	}
}

func TestRegistry_Convert_FeatureWarnings(t *testing.T) {
	r := NewRegistry()

	// Source that produces a schema with $ref
	source := &mockAdapter{
		name: "source",
		toCanonicalFunc: func(raw any) (*CanonicalTool, error) {
			return &CanonicalTool{
				Name: "test",
				InputSchema: &JSONSchema{
					Type: "object",
					Ref:  "#/$defs/Something",
				},
			}, nil
		},
		supportsFunc: func(f SchemaFeature) bool { return true },
	}

	// Target that doesn't support $ref
	target := &mockAdapter{
		name: "target",
		fromCanonicalFunc: func(tool *CanonicalTool) (any, error) {
			return tool.Name, nil
		},
		supportsFunc: func(f SchemaFeature) bool {
			return f != FeatureRef
		},
	}

	_ = r.Register(source)
	_ = r.Register(target)

	result, err := r.Convert("input", "source", "target")

	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if len(result.Warnings) == 0 {
		t.Error("Convert() warnings = empty, want warning for $ref")
	}

	// Check that the warning is about $ref
	found := false
	for _, w := range result.Warnings {
		if w.Feature == FeatureRef {
			found = true
			if w.FromAdapter != "source" {
				t.Errorf("warning FromAdapter = %q, want %q", w.FromAdapter, "source")
			}
			if w.ToAdapter != "target" {
				t.Errorf("warning ToAdapter = %q, want %q", w.ToAdapter, "target")
			}
		}
	}
	if !found {
		t.Error("Convert() missing $ref warning")
	}
}

func TestRegistry_Concurrent(t *testing.T) {
	r := NewRegistry()

	// Pre-register an adapter
	_ = r.Register(&mockAdapter{name: "base"})

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Spawn multiple goroutines doing various operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Register
			adapter := &mockAdapter{name: "adapter-" + string(rune('a'+id))}
			_ = r.Register(adapter)

			// Get
			_, err := r.Get("base")
			if err != nil {
				errors <- err
			}

			// List
			_ = r.List()
		}(i)
	}

	// Additional goroutines doing reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = r.List()
				_, _ = r.Get("base")
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
	}
}

func TestRegistry_Convert_SameFormat(t *testing.T) {
	r := NewRegistry()

	adapter := &mockAdapter{
		name: "same",
		toCanonicalFunc: func(raw any) (*CanonicalTool, error) {
			return &CanonicalTool{
				Name:        "test",
				InputSchema: &JSONSchema{Type: "object"},
			}, nil
		},
		fromCanonicalFunc: func(tool *CanonicalTool) (any, error) {
			return tool.Name, nil
		},
		supportsFunc: func(f SchemaFeature) bool { return true },
	}

	_ = r.Register(adapter)

	result, err := r.Convert("input", "same", "same")

	if err != nil {
		t.Fatalf("Convert() same format error = %v", err)
	}
	if result.Tool != "test" {
		t.Errorf("Convert() same format result = %v, want %q", result.Tool, "test")
	}
}
