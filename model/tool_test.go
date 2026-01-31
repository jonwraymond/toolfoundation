package model

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestTool_ToolID(t *testing.T) {
	tests := []struct {
		name   string
		tool   Tool
		wantID string
	}{
		{
			name: "with namespace",
			tool: Tool{
				Tool:      mcp.Tool{Name: "read"},
				Namespace: "filesystem",
			},
			wantID: "filesystem:read",
		},
		{
			name: "without namespace",
			tool: Tool{
				Tool: mcp.Tool{Name: "read"},
			},
			wantID: "read",
		},
		{
			name: "empty namespace explicitly",
			tool: Tool{
				Tool:      mcp.Tool{Name: "write"},
				Namespace: "",
			},
			wantID: "write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tool.ToolID()
			if got != tt.wantID {
				t.Errorf("ToolID() = %q, want %q", got, tt.wantID)
			}
		})
	}
}

func TestParseToolID(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		wantNamespace string
		wantName      string
		wantErr       bool
	}{
		{
			name:          "with namespace",
			id:            "filesystem:read",
			wantNamespace: "filesystem",
			wantName:      "read",
			wantErr:       false,
		},
		{
			name:          "without namespace",
			id:            "read",
			wantNamespace: "",
			wantName:      "read",
			wantErr:       false,
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: true,
		},
		{
			name:    "multiple colons",
			id:      "a:b:c",
			wantErr: true,
		},
		{
			name:    "leading colon",
			id:      ":name",
			wantErr: true,
		},
		{
			name:    "trailing colon",
			id:      "namespace:",
			wantErr: true,
		},
		{
			name:    "just a colon",
			id:      ":",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNamespace, gotName, err := ParseToolID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToolID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
				return
			}
			if err != nil {
				if err != ErrInvalidToolID {
					t.Errorf("ParseToolID(%q) error = %v, want ErrInvalidToolID", tt.id, err)
				}
				return
			}
			if gotNamespace != tt.wantNamespace {
				t.Errorf("ParseToolID(%q) namespace = %q, want %q", tt.id, gotNamespace, tt.wantNamespace)
			}
			if gotName != tt.wantName {
				t.Errorf("ParseToolID(%q) name = %q, want %q", tt.id, gotName, tt.wantName)
			}
		})
	}
}

func TestParseToolID_RoundTrip(t *testing.T) {
	// Test that ToolID() output can be parsed back correctly
	tests := []struct {
		namespace string
		name      string
	}{
		{"filesystem", "read"},
		{"", "read"},
		{"my-namespace", "my-tool"},
	}

	for _, tt := range tests {
		tool := Tool{
			Tool:      mcp.Tool{Name: tt.name},
			Namespace: tt.namespace,
		}
		id := tool.ToolID()
		gotNamespace, gotName, err := ParseToolID(id)
		if err != nil {
			t.Errorf("ParseToolID(ToolID()) failed for namespace=%q, name=%q: %v", tt.namespace, tt.name, err)
			continue
		}
		if gotNamespace != tt.namespace || gotName != tt.name {
			t.Errorf("Round-trip failed: got (%q, %q), want (%q, %q)", gotNamespace, gotName, tt.namespace, tt.name)
		}
	}
}

func TestBackendKind_Constants(t *testing.T) {
	// Verify the constants have expected string values
	if BackendKindMCP != "mcp" {
		t.Errorf("BackendKindMCP = %q, want %q", BackendKindMCP, "mcp")
	}
	if BackendKindProvider != "provider" {
		t.Errorf("BackendKindProvider = %q, want %q", BackendKindProvider, "provider")
	}
	if BackendKindLocal != "local" {
		t.Errorf("BackendKindLocal = %q, want %q", BackendKindLocal, "local")
	}
}

func TestToolBackend_Structures(t *testing.T) {
	// Test that backend structures can be instantiated correctly
	t.Run("MCP backend", func(t *testing.T) {
		backend := ToolBackend{
			Kind: BackendKindMCP,
			MCP: &MCPBackend{
				ServerName: "my-server",
			},
		}
		if backend.Kind != BackendKindMCP {
			t.Errorf("Kind = %q, want %q", backend.Kind, BackendKindMCP)
		}
		if backend.MCP.ServerName != "my-server" {
			t.Errorf("MCP.ServerName = %q, want %q", backend.MCP.ServerName, "my-server")
		}
	})

	t.Run("Provider backend", func(t *testing.T) {
		backend := ToolBackend{
			Kind: BackendKindProvider,
			Provider: &ProviderBackend{
				ProviderID: "openai",
				ToolID:     "gpt-4-tool",
			},
		}
		if backend.Kind != BackendKindProvider {
			t.Errorf("Kind = %q, want %q", backend.Kind, BackendKindProvider)
		}
		if backend.Provider.ProviderID != "openai" {
			t.Errorf("Provider.ProviderID = %q, want %q", backend.Provider.ProviderID, "openai")
		}
		if backend.Provider.ToolID != "gpt-4-tool" {
			t.Errorf("Provider.ToolID = %q, want %q", backend.Provider.ToolID, "gpt-4-tool")
		}
	})

	t.Run("Local backend", func(t *testing.T) {
		backend := ToolBackend{
			Kind: BackendKindLocal,
			Local: &LocalBackend{
				Name: "my-handler",
			},
		}
		if backend.Kind != BackendKindLocal {
			t.Errorf("Kind = %q, want %q", backend.Kind, BackendKindLocal)
		}
		if backend.Local.Name != "my-handler" {
			t.Errorf("Local.Name = %q, want %q", backend.Local.Name, "my-handler")
		}
	})
}

func TestToolBackend_Validate(t *testing.T) {
	tests := []struct {
		name    string
		backend ToolBackend
		wantErr bool
	}{
		{
			name: "valid MCP backend",
			backend: ToolBackend{
				Kind: BackendKindMCP,
				MCP:  &MCPBackend{ServerName: "server"},
			},
			wantErr: false,
		},
		{
			name: "invalid MCP backend missing server name",
			backend: ToolBackend{
				Kind: BackendKindMCP,
				MCP:  &MCPBackend{},
			},
			wantErr: true,
		},
		{
			name: "valid Provider backend",
			backend: ToolBackend{
				Kind: BackendKindProvider,
				Provider: &ProviderBackend{
					ProviderID: "provider",
					ToolID:     "tool",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid Provider backend missing ProviderID",
			backend: ToolBackend{
				Kind: BackendKindProvider,
				Provider: &ProviderBackend{
					ToolID: "tool",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid Provider backend missing ToolID",
			backend: ToolBackend{
				Kind: BackendKindProvider,
				Provider: &ProviderBackend{
					ProviderID: "provider",
				},
			},
			wantErr: true,
		},
		{
			name: "valid Local backend",
			backend: ToolBackend{
				Kind:  BackendKindLocal,
				Local: &LocalBackend{Name: "handler"},
			},
			wantErr: false,
		},
		{
			name: "invalid Local backend missing Name",
			backend: ToolBackend{
				Kind:  BackendKindLocal,
				Local: &LocalBackend{},
			},
			wantErr: true,
		},
		{
			name: "unknown backend kind",
			backend: ToolBackend{
				Kind: "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.backend.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !errors.Is(err, ErrInvalidBackend) {
				t.Fatalf("expected ErrInvalidBackend, got %v", err)
			}
		})
	}
}

func TestTool_EmbedsMCPTool(t *testing.T) {
	// Verify Tool correctly embeds mcp.Tool and can access its fields
	tool := Tool{
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
		Namespace: "test",
		Version:   "1.0.0",
		Tags:      []string{"alpha", "beta"},
	}

	if tool.Name != "test-tool" {
		t.Errorf("Name = %q, want %q", tool.Name, "test-tool")
	}
	if tool.Description != "A test tool" {
		t.Errorf("Description = %q, want %q", tool.Description, "A test tool")
	}
	if tool.Namespace != "test" {
		t.Errorf("Namespace = %q, want %q", tool.Namespace, "test")
	}
	if tool.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", tool.Version, "1.0.0")
	}
	if len(tool.Tags) != 2 || tool.Tags[0] != "alpha" || tool.Tags[1] != "beta" {
		t.Errorf("Tags = %#v, want %v", tool.Tags, []string{"alpha", "beta"})
	}
}

func TestToolIcon_Alias(t *testing.T) {
	// Verify ToolIcon is a proper alias for mcp.Icon
	icon := ToolIcon{
		Source:   "https://example.com/icon.png",
		MIMEType: "image/png",
	}

	// ToolIcon should be usable as mcp.Icon
	acceptsIcon := func(_ mcp.Icon) {}
	acceptsIcon(icon)

	mcpIcon := icon
	if mcpIcon.Source != "https://example.com/icon.png" {
		t.Errorf("Icon Source = %q, want %q", mcpIcon.Source, "https://example.com/icon.png")
	}
}

func TestTool_ToMCPJSON(t *testing.T) {
	tool := Tool{
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
		Namespace: "test-ns",
		Version:   "1.0.0",
		Tags:      []string{"search", "discovery"},
	}

	data, err := tool.ToMCPJSON()
	if err != nil {
		t.Fatalf("ToMCPJSON() error = %v", err)
	}

	// Parse the JSON and verify namespace/version are NOT present
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal ToMCPJSON result: %v", err)
	}

	if _, ok := result["namespace"]; ok {
		t.Error("ToMCPJSON() should not include namespace field")
	}
	if _, ok := result["version"]; ok {
		t.Error("ToMCPJSON() should not include version field")
	}
	if _, ok := result["tags"]; ok {
		t.Error("ToMCPJSON() should not include tags field")
	}

	// Verify MCP fields are present
	if result["name"] != "test-tool" {
		t.Errorf("ToMCPJSON() name = %v, want %q", result["name"], "test-tool")
	}
	if result["description"] != "A test tool" {
		t.Errorf("ToMCPJSON() description = %v, want %q", result["description"], "A test tool")
	}
}

func TestTool_ToJSON(t *testing.T) {
	tool := Tool{
		Tool: mcp.Tool{
			Name:        "test-tool",
			Description: "A test tool",
			InputSchema: map[string]any{
				"type": "object",
			},
		},
		Namespace: "test-ns",
		Version:   "1.0.0",
		Tags:      []string{"a", "b"},
	}

	data, err := tool.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Parse the JSON and verify all fields are present
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal ToJSON result: %v", err)
	}

	if result["namespace"] != "test-ns" {
		t.Errorf("ToJSON() namespace = %v, want %q", result["namespace"], "test-ns")
	}
	if result["version"] != "1.0.0" {
		t.Errorf("ToJSON() version = %v, want %q", result["version"], "1.0.0")
	}
	if tags, ok := result["tags"].([]any); !ok || len(tags) != 2 {
		t.Errorf("ToJSON() tags = %v, want 2 tags", result["tags"])
	}
	if result["name"] != "test-tool" {
		t.Errorf("ToJSON() name = %v, want %q", result["name"], "test-tool")
	}
}

func TestFromMCPJSON(t *testing.T) {
	mcpJSON := `{
		"name": "mcp-tool",
		"description": "A tool from MCP",
		"inputSchema": {"type": "object"}
	}`

	tool, err := FromMCPJSON([]byte(mcpJSON))
	if err != nil {
		t.Fatalf("FromMCPJSON() error = %v", err)
	}

	if tool.Name != "mcp-tool" {
		t.Errorf("FromMCPJSON() name = %q, want %q", tool.Name, "mcp-tool")
	}
	if tool.Description != "A tool from MCP" {
		t.Errorf("FromMCPJSON() description = %q, want %q", tool.Description, "A tool from MCP")
	}
	// Namespace and Version should be empty
	if tool.Namespace != "" {
		t.Errorf("FromMCPJSON() namespace = %q, want empty", tool.Namespace)
	}
	if tool.Version != "" {
		t.Errorf("FromMCPJSON() version = %q, want empty", tool.Version)
	}
}

func TestFromMCPJSON_InvalidJSON(t *testing.T) {
	_, err := FromMCPJSON([]byte("not valid json"))
	if err == nil {
		t.Error("FromMCPJSON() with invalid JSON should return error")
	}
}

func TestFromJSON(t *testing.T) {
	toolJSON := `{
		"name": "full-tool",
		"description": "A full tool",
		"inputSchema": {"type": "object"},
		"namespace": "my-ns",
		"version": "2.0.0",
		"tags": ["t1", "t2"]
	}`

	tool, err := FromJSON([]byte(toolJSON))
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	if tool.Name != "full-tool" {
		t.Errorf("FromJSON() name = %q, want %q", tool.Name, "full-tool")
	}
	if tool.Namespace != "my-ns" {
		t.Errorf("FromJSON() namespace = %q, want %q", tool.Namespace, "my-ns")
	}
	if tool.Version != "2.0.0" {
		t.Errorf("FromJSON() version = %q, want %q", tool.Version, "2.0.0")
	}
	if len(tool.Tags) != 2 || tool.Tags[0] != "t1" || tool.Tags[1] != "t2" {
		t.Errorf("FromJSON() tags = %#v, want %v", tool.Tags, []string{"t1", "t2"})
	}
}

func TestFromJSON_InvalidJSON(t *testing.T) {
	_, err := FromJSON([]byte("not valid json"))
	if err == nil {
		t.Error("FromJSON() with invalid JSON should return error")
	}
}

func TestJSON_RoundTrip(t *testing.T) {
	original := Tool{
		Tool: mcp.Tool{
			Name:        "roundtrip-tool",
			Description: "Testing round-trip",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"foo": map[string]any{"type": "string"},
				},
			},
		},
		Namespace: "rt-ns",
		Version:   "3.0.0",
		Tags:      []string{"x", "y"},
	}

	// Round-trip through ToJSON/FromJSON
	data, err := original.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	restored, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	if restored.Name != original.Name {
		t.Errorf("Round-trip name = %q, want %q", restored.Name, original.Name)
	}
	if restored.Description != original.Description {
		t.Errorf("Round-trip description = %q, want %q", restored.Description, original.Description)
	}
	if restored.Namespace != original.Namespace {
		t.Errorf("Round-trip namespace = %q, want %q", restored.Namespace, original.Namespace)
	}
	if restored.Version != original.Version {
		t.Errorf("Round-trip version = %q, want %q", restored.Version, original.Version)
	}
	if len(restored.Tags) != 2 || restored.Tags[0] != "x" || restored.Tags[1] != "y" {
		t.Errorf("Round-trip tags = %#v, want %v", restored.Tags, []string{"x", "y"})
	}
}

func TestMCPJSON_RoundTrip(t *testing.T) {
	original := Tool{
		Tool: mcp.Tool{
			Name:        "mcp-roundtrip",
			Description: "Testing MCP round-trip",
			InputSchema: map[string]any{
				"type": "object",
			},
		},
		Namespace: "will-be-lost",
		Version:   "also-lost",
	}

	// Round-trip through ToMCPJSON/FromMCPJSON
	data, err := original.ToMCPJSON()
	if err != nil {
		t.Fatalf("ToMCPJSON() error = %v", err)
	}

	restored, err := FromMCPJSON(data)
	if err != nil {
		t.Fatalf("FromMCPJSON() error = %v", err)
	}

	// MCP fields should be preserved
	if restored.Name != original.Name {
		t.Errorf("MCP round-trip name = %q, want %q", restored.Name, original.Name)
	}
	if restored.Description != original.Description {
		t.Errorf("MCP round-trip description = %q, want %q", restored.Description, original.Description)
	}

	// Namespace and Version should be empty (stripped by ToMCPJSON)
	if restored.Namespace != "" {
		t.Errorf("MCP round-trip namespace = %q, want empty (stripped)", restored.Namespace)
	}
	if restored.Version != "" {
		t.Errorf("MCP round-trip version = %q, want empty (stripped)", restored.Version)
	}
}

func TestMCPJSON_RoundTrip_AllMCPFields(t *testing.T) {
	original := Tool{
		Tool: mcp.Tool{
			Meta: mcp.Meta{
				"traceId": "abc123",
			},
			Annotations: &mcp.ToolAnnotations{
				Title:           "Annotated Title",
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: boolPtr(false),
			},
			Name:        "full-mcp-tool",
			Title:       "Display Title",
			Description: "Full MCP tool coverage",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"a": map[string]any{"type": "string"},
				},
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"ok": map[string]any{"type": "boolean"},
				},
			},
			Icons: []mcp.Icon{
				{Source: "https://example.com/icon.png", MIMEType: "image/png"},
			},
		},
		Namespace: "ns",
		Version:   "1.0.0",
		Tags:      []string{"alpha"},
	}

	data, err := original.ToMCPJSON()
	if err != nil {
		t.Fatalf("ToMCPJSON() error = %v", err)
	}

	restored, err := FromMCPJSON(data)
	if err != nil {
		t.Fatalf("FromMCPJSON() error = %v", err)
	}

	if restored.Name != original.Name {
		t.Errorf("name = %q, want %q", restored.Name, original.Name)
	}
	if restored.Title != original.Title {
		t.Errorf("title = %q, want %q", restored.Title, original.Title)
	}
	if restored.Description != original.Description {
		t.Errorf("description = %q, want %q", restored.Description, original.Description)
	}
	if restored.Annotations == nil || restored.Annotations.Title != original.Annotations.Title {
		t.Errorf("annotations title = %#v, want %#v", restored.Annotations, original.Annotations)
	}
	if restored.GetMeta()["traceId"] != "abc123" {
		t.Errorf("meta traceId = %v, want %q", restored.GetMeta()["traceId"], "abc123")
	}
	if len(restored.Icons) != 1 || restored.Icons[0].Source != original.Icons[0].Source {
		t.Errorf("icons = %#v, want %#v", restored.Icons, original.Icons)
	}
	if restored.Namespace != "" || restored.Version != "" || len(restored.Tags) != 0 {
		t.Errorf("non-MCP fields should be stripped: namespace=%q version=%q tags=%v",
			restored.Namespace, restored.Version, restored.Tags)
	}
}

func TestToolValidate(t *testing.T) {
	valid := Tool{
		Tool: mcp.Tool{
			Name:        "ok",
			Description: "desc",
			InputSchema: map[string]any{"type": "object"},
		},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}

	missingName := Tool{
		Tool: mcp.Tool{
			Name:        "",
			Description: "desc",
			InputSchema: map[string]any{"type": "object"},
		},
	}
	if err := missingName.Validate(); err == nil {
		t.Fatal("Validate() expected error for empty name")
	}

	missingSchema := Tool{
		Tool: mcp.Tool{
			Name:        "no-schema",
			Description: "desc",
			InputSchema: nil,
		},
	}
	if err := missingSchema.Validate(); err == nil {
		t.Fatal("Validate() expected error for nil InputSchema")
	}
}

func boolPtr(v bool) *bool {
	return &v
}

func TestToolValidate_NameRules(t *testing.T) {
	tooLong := strings.Repeat("a", 129)
	invalidChars := "bad:name"

	tests := []struct {
		name string
		tool Tool
	}{
		{
			name: "too long name",
			tool: Tool{
				Tool: mcp.Tool{
					Name:        tooLong,
					Description: "desc",
					InputSchema: map[string]any{"type": "object"},
				},
			},
		},
		{
			name: "invalid chars",
			tool: Tool{
				Tool: mcp.Tool{
					Name:        invalidChars,
					Description: "desc",
					InputSchema: map[string]any{"type": "object"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tool.Validate(); err == nil {
				t.Fatalf("Validate() expected error for %s", tt.name)
			}
		})
	}
}

func TestNormalizeTags(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "basic normalization and dedupe",
			in:   []string{"  Foo ", "foo", "Bar Baz", "bar-baz", "A_B", "A.B"},
			want: []string{"foo", "bar-baz", "a_b", "a.b"},
		},
		{
			name: "filters invalid characters and empties",
			in:   []string{"", "   ", "###", "ok!", "good_tag"},
			want: []string{"ok", "good_tag"},
		},
		{
			name: "limits count and length",
			in:   append([]string{"tag1"}, make([]string, 25)...),
			want: []string{"tag1"},
		},
	}

	t.Run("length truncation", func(t *testing.T) {
		long := strings.Repeat("a", 100)
		got := NormalizeTags([]string{long})
		if len(got) != 1 {
			t.Fatalf("expected 1 tag, got %d", len(got))
		}
		if len(got[0]) != 64 {
			t.Fatalf("expected tag length 64, got %d", len(got[0]))
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeTags(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("NormalizeTags() len = %d, want %d (%v)", len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("NormalizeTags()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
