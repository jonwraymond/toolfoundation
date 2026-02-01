package model

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jonwraymond/toolfoundation/version"
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
			name: "invalid Provider backend nil Provider",
			backend: ToolBackend{
				Kind:     BackendKindProvider,
				Provider: nil, // Provider is nil but Kind is provider
			},
			wantErr: true,
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

func TestNewMCPBackend(t *testing.T) {
	backend := NewMCPBackend("my-server")

	if backend.Kind != BackendKindMCP {
		t.Errorf("Kind = %q, want %q", backend.Kind, BackendKindMCP)
	}
	if backend.MCP == nil {
		t.Fatal("MCP should not be nil")
	}
	if backend.MCP.ServerName != "my-server" {
		t.Errorf("ServerName = %q, want %q", backend.MCP.ServerName, "my-server")
	}
	if err := backend.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestNewLocalBackend(t *testing.T) {
	backend := NewLocalBackend("my-handler")

	if backend.Kind != BackendKindLocal {
		t.Errorf("Kind = %q, want %q", backend.Kind, BackendKindLocal)
	}
	if backend.Local == nil {
		t.Fatal("Local should not be nil")
	}
	if backend.Local.Name != "my-handler" {
		t.Errorf("Name = %q, want %q", backend.Local.Name, "my-handler")
	}
	if err := backend.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestNewProviderBackend(t *testing.T) {
	backend := NewProviderBackend("openai", "gpt-4-tool")

	if backend.Kind != BackendKindProvider {
		t.Errorf("Kind = %q, want %q", backend.Kind, BackendKindProvider)
	}
	if backend.Provider == nil {
		t.Fatal("Provider should not be nil")
	}
	if backend.Provider.ProviderID != "openai" {
		t.Errorf("ProviderID = %q, want %q", backend.Provider.ProviderID, "openai")
	}
	if backend.Provider.ToolID != "gpt-4-tool" {
		t.Errorf("ToolID = %q, want %q", backend.Provider.ToolID, "gpt-4-tool")
	}
	if err := backend.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestTool_Clone(t *testing.T) {
	original := &Tool{
		Tool: mcp.Tool{
			Meta: mcp.Meta{
				"traceId": "abc123",
			},
			Annotations: &mcp.ToolAnnotations{
				Title:           "Annotated Title",
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: boolPtr(false),
				OpenWorldHint:   boolPtr(true),
			},
			Name:        "clone-test",
			Title:       "Clone Test Tool",
			Description: "A tool for testing Clone()",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"input": map[string]any{"type": "string"},
				},
				"required": []any{"input"},
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"result": map[string]any{"type": "boolean"},
				},
			},
			Icons: []mcp.Icon{
				{Source: "https://example.com/icon.png", MIMEType: "image/png"},
			},
		},
		Namespace: "test-ns",
		Version:   "1.2.3",
		Tags:      []string{"alpha", "beta"},
	}

	clone := original.Clone()

	// Verify clone is not the same pointer
	if clone == original {
		t.Error("Clone() should return a new pointer")
	}

	// Verify basic fields
	if clone.Name != original.Name {
		t.Errorf("Name = %q, want %q", clone.Name, original.Name)
	}
	if clone.Title != original.Title {
		t.Errorf("Title = %q, want %q", clone.Title, original.Title)
	}
	if clone.Description != original.Description {
		t.Errorf("Description = %q, want %q", clone.Description, original.Description)
	}
	if clone.Namespace != original.Namespace {
		t.Errorf("Namespace = %q, want %q", clone.Namespace, original.Namespace)
	}
	if clone.Version != original.Version {
		t.Errorf("Version = %q, want %q", clone.Version, original.Version)
	}

	// Verify Meta is a separate copy
	if clone.Meta == nil {
		t.Fatal("Meta should not be nil")
	}
	if clone.GetMeta()["traceId"] != "abc123" {
		t.Errorf("Meta traceId = %v, want %q", clone.GetMeta()["traceId"], "abc123")
	}
	// Modify clone's Meta and verify original is unchanged
	clone.Meta["newKey"] = "newValue"
	if _, ok := original.Meta["newKey"]; ok {
		t.Error("Modifying clone's Meta should not affect original")
	}

	// Verify Annotations is a separate copy
	if clone.Annotations == nil {
		t.Fatal("Annotations should not be nil")
	}
	if clone.Annotations.Title != original.Annotations.Title {
		t.Errorf("Annotations.Title = %q, want %q", clone.Annotations.Title, original.Annotations.Title)
	}
	if clone.Annotations.ReadOnlyHint != original.Annotations.ReadOnlyHint {
		t.Errorf("Annotations.ReadOnlyHint = %v, want %v", clone.Annotations.ReadOnlyHint, original.Annotations.ReadOnlyHint)
	}
	if *clone.Annotations.DestructiveHint != *original.Annotations.DestructiveHint {
		t.Errorf("Annotations.DestructiveHint = %v, want %v", *clone.Annotations.DestructiveHint, *original.Annotations.DestructiveHint)
	}
	// Modify clone's Annotations and verify original is unchanged
	clone.Annotations.Title = "Modified"
	if original.Annotations.Title == "Modified" {
		t.Error("Modifying clone's Annotations should not affect original")
	}

	// Verify Tags is a separate copy
	if len(clone.Tags) != len(original.Tags) {
		t.Errorf("Tags length = %d, want %d", len(clone.Tags), len(original.Tags))
	}
	clone.Tags[0] = "modified"
	if original.Tags[0] == "modified" {
		t.Error("Modifying clone's Tags should not affect original")
	}

	// Verify InputSchema is a separate copy
	if clone.InputSchema == nil {
		t.Fatal("InputSchema should not be nil")
	}
	cloneSchema := clone.InputSchema.(map[string]any)
	cloneSchema["type"] = "array"
	originalSchema := original.InputSchema.(map[string]any)
	if originalSchema["type"] == "array" {
		t.Error("Modifying clone's InputSchema should not affect original")
	}

	// Verify Icons is a separate copy
	if len(clone.Icons) != len(original.Icons) {
		t.Errorf("Icons length = %d, want %d", len(clone.Icons), len(original.Icons))
	}
}

func TestTool_Clone_Nil(t *testing.T) {
	var tool *Tool
	clone := tool.Clone()
	if clone != nil {
		t.Error("Clone() of nil should return nil")
	}
}

func TestTool_Clone_MinimalFields(t *testing.T) {
	original := &Tool{
		Tool: mcp.Tool{
			Name:        "minimal",
			InputSchema: map[string]any{"type": "object"},
		},
	}

	clone := original.Clone()

	if clone.Name != original.Name {
		t.Errorf("Name = %q, want %q", clone.Name, original.Name)
	}
	if clone.Annotations != nil {
		t.Error("Annotations should be nil for minimal tool")
	}
	if clone.Meta != nil {
		t.Error("Meta should be nil for minimal tool")
	}
	if clone.Tags != nil {
		t.Error("Tags should be nil for minimal tool")
	}
}

func TestTool_ParsedVersion(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		wantMajor  int
		wantMinor  int
		wantPatch  int
		wantPrerel string
		wantErr    bool
	}{
		{
			name:      "simple version",
			version:   "1.2.3",
			wantMajor: 1,
			wantMinor: 2,
			wantPatch: 3,
			wantErr:   false,
		},
		{
			name:      "version with v prefix",
			version:   "v2.0.0",
			wantMajor: 2,
			wantMinor: 0,
			wantPatch: 0,
			wantErr:   false,
		},
		{
			name:       "version with prerelease",
			version:    "1.0.0-alpha.1",
			wantMajor:  1,
			wantMinor:  0,
			wantPatch:  0,
			wantPrerel: "alpha.1",
			wantErr:    false,
		},
		{
			name:    "empty version",
			version: "",
			wantErr: true,
		},
		{
			name:    "invalid version",
			version: "not-a-version",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &Tool{
				Tool: mcp.Tool{
					Name:        "test",
					InputSchema: map[string]any{"type": "object"},
				},
				Version: tt.version,
			}

			got, err := tool.ParsedVersion()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsedVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.Major != tt.wantMajor {
				t.Errorf("Major = %d, want %d", got.Major, tt.wantMajor)
			}
			if got.Minor != tt.wantMinor {
				t.Errorf("Minor = %d, want %d", got.Minor, tt.wantMinor)
			}
			if got.Patch != tt.wantPatch {
				t.Errorf("Patch = %d, want %d", got.Patch, tt.wantPatch)
			}
			if got.Prerelease != tt.wantPrerel {
				t.Errorf("Prerelease = %q, want %q", got.Prerelease, tt.wantPrerel)
			}
		})
	}
}

func TestTool_ParsedVersion_Comparison(t *testing.T) {
	tool1 := &Tool{
		Tool:    mcp.Tool{Name: "t1", InputSchema: map[string]any{"type": "object"}},
		Version: "1.0.0",
	}
	tool2 := &Tool{
		Tool:    mcp.Tool{Name: "t2", InputSchema: map[string]any{"type": "object"}},
		Version: "2.0.0",
	}

	v1, err := tool1.ParsedVersion()
	if err != nil {
		t.Fatalf("ParsedVersion() error = %v", err)
	}
	v2, err := tool2.ParsedVersion()
	if err != nil {
		t.Fatalf("ParsedVersion() error = %v", err)
	}

	if !v1.LessThan(v2) {
		t.Errorf("v1 (%v) should be less than v2 (%v)", v1, v2)
	}
	if !v2.GreaterThan(v1) {
		t.Errorf("v2 (%v) should be greater than v1 (%v)", v2, v1)
	}

	// Test compatibility
	tool3 := &Tool{
		Tool:    mcp.Tool{Name: "t3", InputSchema: map[string]any{"type": "object"}},
		Version: "1.5.0",
	}
	v3, _ := tool3.ParsedVersion()

	if !v3.Compatible(v1) {
		t.Errorf("v3 (%v) should be compatible with v1 (%v)", v3, v1)
	}
	if v3.Compatible(v2) {
		t.Errorf("v3 (%v) should not be compatible with v2 (%v)", v3, v2)
	}
}

// Verify _ import is used
var _ = version.Version{}

func TestNormalizeTags_MaxCount(t *testing.T) {
	// Create more than 20 tags to test the count limit
	tags := make([]string, 25)
	for i := range tags {
		tags[i] = "tag" + string(rune('a'+i%26))
	}

	result := NormalizeTags(tags)
	if len(result) > 20 {
		t.Errorf("NormalizeTags() returned %d tags, want max 20", len(result))
	}
}

func TestTool_Validate_InvalidCharsMultiple(t *testing.T) {
	// Test with multiple different invalid characters
	tool := Tool{
		Tool: mcp.Tool{
			Name:        "bad@name#here",
			Description: "desc",
			InputSchema: map[string]any{"type": "object"},
		},
	}

	err := tool.Validate()
	if err == nil {
		t.Fatal("Validate() expected error for multiple invalid characters")
	}
	// Should report both @ and #
	if !strings.Contains(err.Error(), "@") || !strings.Contains(err.Error(), "#") {
		t.Errorf("Validate() error should mention both invalid chars: %v", err)
	}
}

func TestTool_Clone_NilInputSchema(t *testing.T) {
	// Test Clone with nil schemas - covers deepCopyAny nil path
	tool := &Tool{
		Tool: mcp.Tool{
			Name:        "test",
			Description: "test tool",
			InputSchema: nil, // nil schema
		},
	}

	clone := tool.Clone()
	if clone.InputSchema != nil {
		t.Error("Clone() should preserve nil InputSchema")
	}
}

func TestTool_Clone_WithUnmarshallableSchemaValue(t *testing.T) {
	// Test Clone with a schema containing a channel (unmarshallable by JSON)
	// This tests the deepCopyAny fallback path when json.Marshal fails
	ch := make(chan int)
	tool := &Tool{
		Tool: mcp.Tool{
			Name:        "test",
			Description: "test tool",
			InputSchema: map[string]any{
				"type":    "object",
				"channel": ch, // channels can't be marshaled to JSON
			},
		},
	}

	// Clone should not panic - it falls back to shallow copy on marshal error
	clone := tool.Clone()
	if clone == nil {
		t.Fatal("Clone() should not return nil")
	}
	if clone.Name != tool.Name {
		t.Errorf("Clone().Name = %q, want %q", clone.Name, tool.Name)
	}
}

func TestTool_Clone_WithFuncInSchema(t *testing.T) {
	// Test Clone with a schema containing a func (unmarshallable by JSON)
	fn := func() {}
	tool := &Tool{
		Tool: mcp.Tool{
			Name:        "test",
			Description: "test tool",
			InputSchema: map[string]any{
				"type": "object",
				"func": fn, // functions can't be marshaled to JSON
			},
			OutputSchema: map[string]any{
				"type": "object",
				"func": fn,
			},
		},
	}

	// Clone should not panic - it falls back to shallow copy on marshal error
	clone := tool.Clone()
	if clone == nil {
		t.Fatal("Clone() should not return nil")
	}
}

func TestNormalizeTags_AllSpecialChars(t *testing.T) {
	// Tag with only special characters that get filtered out
	tags := []string{"@#$%^&*()"}
	result := NormalizeTags(tags)
	if len(result) != 0 {
		t.Errorf("NormalizeTags() = %v, want empty slice for all-special-char tag", result)
	}
}

func TestNormalizeTags_WhitespaceOnly(t *testing.T) {
	// Tag that becomes empty after Fields processing
	tags := []string{"   ", "\t\n"}
	result := NormalizeTags(tags)
	if len(result) != 0 {
		t.Errorf("NormalizeTags() = %v, want empty slice for whitespace-only tags", result)
	}
}
