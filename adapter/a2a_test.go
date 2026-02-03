package adapter

import "testing"

func TestNewA2AAdapter(t *testing.T) {
	adapter := NewA2AAdapter()
	if adapter == nil {
		t.Fatal("NewA2AAdapter() returned nil")
	}
}

func TestA2AAdapter_Name(t *testing.T) {
	adapter := NewA2AAdapter()
	if adapter.Name() != "a2a" {
		t.Errorf("Name() = %q, want %q", adapter.Name(), "a2a")
	}
}

func TestA2AAdapter_ToCanonical(t *testing.T) {
	adapter := NewA2AAdapter()

	skill := &A2AAgentSkill{
		ID:          "tools:search:1.2.3",
		Name:        "Search",
		Description: "Search for documents",
		Tags:        []string{"search"},
		Examples:    []string{"find invoices"},
		InputModes:  []string{"application/json"},
		OutputModes: []string{"application/json"},
	}

	ct, err := adapter.ToCanonical(skill)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	if ct.Namespace != "tools" {
		t.Errorf("Namespace = %q, want %q", ct.Namespace, "tools")
	}
	if ct.Name != "search" {
		t.Errorf("Name = %q, want %q", ct.Name, "search")
	}
	if ct.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", ct.Version, "1.2.3")
	}
	if ct.DisplayName != "Search" {
		t.Errorf("DisplayName = %q, want %q", ct.DisplayName, "Search")
	}
	if ct.InputSchema == nil || ct.InputSchema.Type != "object" {
		t.Errorf("InputSchema.Type = %v, want %q", ct.InputSchema.Type, "object")
	}
}

func TestA2AAdapter_FromCanonical(t *testing.T) {
	adapter := NewA2AAdapter()

	ct := &CanonicalTool{
		Namespace:   "tools",
		Name:        "search",
		Version:     "1.2.3",
		DisplayName: "Search",
		Description: "Search for documents",
		Tags:        []string{"search"},
		Examples:    []string{"find invoices"},
		InputModes:  []string{"application/json"},
		OutputModes: []string{"application/json"},
		InputSchema: &JSONSchema{Type: "object"},
	}

	raw, err := adapter.FromCanonical(ct)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	skill := raw.(*A2AAgentSkill)
	if skill.ID != "tools:search:1.2.3" {
		t.Errorf("ID = %q, want %q", skill.ID, "tools:search:1.2.3")
	}
	if skill.Name != "Search" {
		t.Errorf("Name = %q, want %q", skill.Name, "Search")
	}
}

func TestA2AAdapter_FromCanonical_UsesSourceMetaSkillID(t *testing.T) {
	adapter := NewA2AAdapter()

	ct := &CanonicalTool{
		Name:        "search",
		Description: "Search for documents",
		InputSchema: &JSONSchema{Type: "object"},
		SourceMeta:  map[string]any{"skillId": "custom-id"},
	}

	raw, err := adapter.FromCanonical(ct)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	skill := raw.(*A2AAgentSkill)
	if skill.ID != "custom-id" {
		t.Errorf("ID = %q, want %q", skill.ID, "custom-id")
	}
}

func TestA2AAdapter_ProviderRoundTrip(t *testing.T) {
	adapter := NewA2AAdapter()

	card := &A2AAgentCard{
		Name:        "Test Agent",
		Description: "Handles test workflows",
		Version:     "1.0.0",
		SupportedInterfaces: []A2AAgentInterface{
			{
				URL:             "https://example.com/a2a",
				ProtocolBinding: "JSONRPC",
				ProtocolVersion: "1.0",
			},
		},
		Capabilities: A2AAgentCapabilities{
			Streaming: boolPtr(true),
		},
		DefaultInputModes:  []string{"application/json"},
		DefaultOutputModes: []string{"application/json"},
		Skills: []A2AAgentSkill{
			{
				ID:          "tools:search:1.0.0",
				Name:        "Search",
				Description: "Search skill",
				Tags:        []string{"search"},
			},
		},
	}

	provider, err := adapter.ToCanonicalProvider(card)
	if err != nil {
		t.Fatalf("ToCanonicalProvider() error = %v", err)
	}

	if provider.Name != "Test Agent" {
		t.Errorf("Name = %q, want %q", provider.Name, "Test Agent")
	}
	if provider.Capabilities["streaming"] != true {
		t.Errorf("Capabilities[streaming] = %v, want true", provider.Capabilities["streaming"])
	}

	roundTripped, err := adapter.FromCanonicalProvider(provider)
	if err != nil {
		t.Fatalf("FromCanonicalProvider() error = %v", err)
	}

	if roundTripped.Name != card.Name {
		t.Errorf("Name = %q, want %q", roundTripped.Name, card.Name)
	}
	if len(roundTripped.SupportedInterfaces) != 1 {
		t.Errorf("SupportedInterfaces length = %d, want 1", len(roundTripped.SupportedInterfaces))
	}
	if len(roundTripped.Skills) != 1 {
		t.Errorf("Skills length = %d, want 1", len(roundTripped.Skills))
	}
}

func TestA2AAdapter_FromCanonicalProvider_MissingInterfaces(t *testing.T) {
	adapter := NewA2AAdapter()

	_, err := adapter.FromCanonicalProvider(&CanonicalProvider{
		Name:        "Agent",
		Description: "Desc",
		Version:     "1.0.0",
	})
	if err == nil {
		t.Error("expected error for missing supportedInterfaces")
	}
}
