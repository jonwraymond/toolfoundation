package model

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestToolBuilder_Basic(t *testing.T) {
	tool, err := NewTool("read-file").
		Description("Reads a file from disk").
		Namespace("filesystem").
		Version("1.0.0").
		InputSchema(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string"},
			},
			"required": []string{"path"},
		}).
		Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if tool.Name != "read-file" {
		t.Errorf("Name = %q, want %q", tool.Name, "read-file")
	}
	if tool.Description != "Reads a file from disk" {
		t.Errorf("Description = %q, want %q", tool.Description, "Reads a file from disk")
	}
	if tool.Namespace != "filesystem" {
		t.Errorf("Namespace = %q, want %q", tool.Namespace, "filesystem")
	}
	if tool.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", tool.Version, "1.0.0")
	}
	if tool.InputSchema == nil {
		t.Error("InputSchema should not be nil")
	}
}

func TestToolBuilder_AllFields(t *testing.T) {
	tool, err := NewTool("full-tool").
		Description("A fully configured tool").
		Namespace("test").
		Version("2.0.0").
		Tags("search", "discovery", "TEST").
		Title("Full Tool").
		InputSchema(map[string]any{"type": "object"}).
		OutputSchema(map[string]any{"type": "object"}).
		Icons(mcp.Icon{Source: "https://example.com/icon.png", MIMEType: "image/png"}).
		Meta(mcp.Meta{"traceId": "abc123"}).
		Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if tool.Title != "Full Tool" {
		t.Errorf("Title = %q, want %q", tool.Title, "Full Tool")
	}
	if len(tool.Tags) != 3 {
		t.Errorf("Tags len = %d, want 3", len(tool.Tags))
	}
	if tool.Tags[0] != "search" || tool.Tags[1] != "discovery" || tool.Tags[2] != "test" {
		t.Errorf("Tags = %v, want [search discovery test] (normalized)", tool.Tags)
	}
	if tool.OutputSchema == nil {
		t.Error("OutputSchema should not be nil")
	}
	if len(tool.Icons) != 1 {
		t.Errorf("Icons len = %d, want 1", len(tool.Icons))
	}
	if tool.GetMeta()["traceId"] != "abc123" {
		t.Errorf("Meta traceId = %v, want %q", tool.GetMeta()["traceId"], "abc123")
	}
}

func TestToolBuilder_Annotations(t *testing.T) {
	t.Run("ReadOnly", func(t *testing.T) {
		tool := NewTool("ro-tool").
			InputSchema(map[string]any{"type": "object"}).
			ReadOnly().
			MustBuild()

		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
			t.Error("ReadOnly() should set ReadOnlyHint to true")
		}
	})

	t.Run("Idempotent", func(t *testing.T) {
		tool := NewTool("idem-tool").
			InputSchema(map[string]any{"type": "object"}).
			Idempotent().
			MustBuild()

		if tool.Annotations == nil || !tool.Annotations.IdempotentHint {
			t.Error("Idempotent() should set IdempotentHint to true")
		}
	})

	t.Run("Destructive", func(t *testing.T) {
		tool := NewTool("dest-tool").
			InputSchema(map[string]any{"type": "object"}).
			Destructive().
			MustBuild()

		if tool.Annotations == nil || tool.Annotations.DestructiveHint == nil || !*tool.Annotations.DestructiveHint {
			t.Error("Destructive() should set DestructiveHint to true")
		}
	})

	t.Run("NonDestructive", func(t *testing.T) {
		tool := NewTool("safe-tool").
			InputSchema(map[string]any{"type": "object"}).
			NonDestructive().
			MustBuild()

		if tool.Annotations == nil || tool.Annotations.DestructiveHint == nil || *tool.Annotations.DestructiveHint {
			t.Error("NonDestructive() should set DestructiveHint to false")
		}
	})

	t.Run("OpenWorld", func(t *testing.T) {
		tool := NewTool("external-tool").
			InputSchema(map[string]any{"type": "object"}).
			OpenWorld().
			MustBuild()

		if tool.Annotations == nil || tool.Annotations.OpenWorldHint == nil || !*tool.Annotations.OpenWorldHint {
			t.Error("OpenWorld() should set OpenWorldHint to true")
		}
	})

	t.Run("Combined", func(t *testing.T) {
		tool := NewTool("combo-tool").
			InputSchema(map[string]any{"type": "object"}).
			ReadOnly().
			Idempotent().
			NonDestructive().
			MustBuild()

		if tool.Annotations == nil {
			t.Fatal("Annotations should not be nil")
		}
		if !tool.Annotations.ReadOnlyHint {
			t.Error("ReadOnlyHint should be true")
		}
		if !tool.Annotations.IdempotentHint {
			t.Error("IdempotentHint should be true")
		}
		if tool.Annotations.DestructiveHint == nil || *tool.Annotations.DestructiveHint {
			t.Error("DestructiveHint should be false")
		}
	})
}

func TestToolBuilder_CustomAnnotations(t *testing.T) {
	tool := NewTool("custom-ann").
		InputSchema(map[string]any{"type": "object"}).
		Annotations(&mcp.ToolAnnotations{
			Title:        "Custom Title",
			ReadOnlyHint: true,
		}).
		MustBuild()

	if tool.Annotations.Title != "Custom Title" {
		t.Errorf("Annotations.Title = %q, want %q", tool.Annotations.Title, "Custom Title")
	}
	if !tool.Annotations.ReadOnlyHint {
		t.Error("Annotations.ReadOnlyHint should be true")
	}
}

func TestToolBuilder_ValidationErrors(t *testing.T) {
	t.Run("missing name", func(t *testing.T) {
		_, err := NewTool("").
			InputSchema(map[string]any{"type": "object"}).
			Build()
		if err == nil {
			t.Error("Build() should fail for empty name")
		}
	})

	t.Run("missing schema", func(t *testing.T) {
		_, err := NewTool("no-schema").Build()
		if err == nil {
			t.Error("Build() should fail for missing InputSchema")
		}
	})

	t.Run("invalid name characters", func(t *testing.T) {
		_, err := NewTool("bad:name").
			InputSchema(map[string]any{"type": "object"}).
			Build()
		if err == nil {
			t.Error("Build() should fail for invalid name characters")
		}
	})
}

func TestToolBuilder_MustBuild(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tool := NewTool("must-build").
			InputSchema(map[string]any{"type": "object"}).
			MustBuild()

		if tool.Name != "must-build" {
			t.Errorf("Name = %q, want %q", tool.Name, "must-build")
		}
	})

	t.Run("panic on error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustBuild() should panic on validation error")
			}
		}()

		NewTool("").
			InputSchema(map[string]any{"type": "object"}).
			MustBuild()
	})
}

func TestToolBuilder_TagsNormalization(t *testing.T) {
	tool := NewTool("tagged").
		InputSchema(map[string]any{"type": "object"}).
		Tags("  UPPER  ", "with spaces", "with!special@chars").
		MustBuild()

	expected := []string{"upper", "with-spaces", "withspecialchars"}
	if len(tool.Tags) != len(expected) {
		t.Fatalf("Tags len = %d, want %d", len(tool.Tags), len(expected))
	}
	for i, tag := range expected {
		if tool.Tags[i] != tag {
			t.Errorf("Tags[%d] = %q, want %q", i, tool.Tags[i], tag)
		}
	}
}

func TestToolBuilder_Immutability(t *testing.T) {
	builder := NewTool("immutable").
		InputSchema(map[string]any{"type": "object"}).
		Description("original")

	tool1, _ := builder.Build()

	builder.Description("modified")
	tool2, _ := builder.Build()

	if tool1.Description == tool2.Description {
		t.Error("Build() should return independent copies")
	}
}

func TestToolBuilder_Chaining(t *testing.T) {
	builder := NewTool("chained")

	result := builder.
		Description("desc").
		Namespace("ns").
		Version("1.0").
		Tags("a", "b").
		InputSchema(map[string]any{"type": "object"}).
		Title("title").
		ReadOnly().
		Idempotent()

	if result != builder {
		t.Error("All builder methods should return the same builder instance for chaining")
	}
}
