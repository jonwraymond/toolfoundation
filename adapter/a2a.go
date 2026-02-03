package adapter

import (
	"errors"
	"fmt"
	"strings"
)

// A2AAdapter converts between A2A AgentSkill and CanonicalTool.
// Provider-level conversions are exposed via ToCanonicalProvider/FromCanonicalProvider.
type A2AAdapter struct{}

// NewA2AAdapter creates a new A2A adapter.
func NewA2AAdapter() *A2AAdapter {
	return &A2AAdapter{}
}

// Name returns the adapter's identifier.
func (a *A2AAdapter) Name() string {
	return "a2a"
}

// A2AAgentCard represents an A2A Agent Card (simplified for adapter use).
type A2AAgentCard struct {
	Name               string                    `json:"name"`
	Description        string                    `json:"description"`
	SupportedInterfaces []A2AAgentInterface      `json:"supportedInterfaces"`
	Provider           *A2AAgentProvider         `json:"provider,omitempty"`
	Version            string                    `json:"version"`
	DocumentationURL   string                    `json:"documentationUrl,omitempty"`
	Capabilities       A2AAgentCapabilities      `json:"capabilities"`
	SecuritySchemes    map[string]SecurityScheme `json:"securitySchemes,omitempty"`
	SecurityRequirements []SecurityRequirement   `json:"securityRequirements,omitempty"`
	DefaultInputModes  []string                  `json:"defaultInputModes"`
	DefaultOutputModes []string                  `json:"defaultOutputModes"`
	Skills             []A2AAgentSkill           `json:"skills"`
	Signatures         []map[string]any          `json:"signatures,omitempty"`
	IconURL            string                    `json:"iconUrl,omitempty"`
}

// A2AAgentProvider describes the provider of an agent.
type A2AAgentProvider struct {
	URL          string `json:"url"`
	Organization string `json:"organization"`
}

// A2AAgentCapabilities describes agent capability flags.
type A2AAgentCapabilities struct {
	Streaming        *bool              `json:"streaming,omitempty"`
	PushNotifications *bool             `json:"pushNotifications,omitempty"`
	Extensions       []A2AAgentExtension `json:"extensions,omitempty"`
	ExtendedAgentCard *bool             `json:"extendedAgentCard,omitempty"`
}

// A2AAgentExtension describes a supported extension.
type A2AAgentExtension struct {
	URI         string         `json:"uri,omitempty"`
	Description string         `json:"description,omitempty"`
	Required    *bool          `json:"required,omitempty"`
	Params      map[string]any `json:"params,omitempty"`
}

// A2AAgentSkill describes a distinct skill.
type A2AAgentSkill struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	Description          string                 `json:"description"`
	Tags                 []string               `json:"tags"`
	Examples             []string               `json:"examples,omitempty"`
	InputModes           []string               `json:"inputModes,omitempty"`
	OutputModes          []string               `json:"outputModes,omitempty"`
	SecurityRequirements []SecurityRequirement  `json:"securityRequirements,omitempty"`
}

// A2AAgentInterface describes a supported protocol binding.
type A2AAgentInterface struct {
	URL            string `json:"url"`
	ProtocolBinding string `json:"protocolBinding"`
	Tenant         string `json:"tenant,omitempty"`
	ProtocolVersion string `json:"protocolVersion"`
}

// ToCanonical converts an A2A AgentSkill to the canonical format.
// Accepts *A2AAgentSkill or A2AAgentSkill.
func (a *A2AAdapter) ToCanonical(raw any) (*CanonicalTool, error) {
	if raw == nil {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical",
			Cause:     errors.New("input is nil"),
		}
	}

	var skill *A2AAgentSkill

	switch v := raw.(type) {
	case *A2AAgentSkill:
		skill = v
	case A2AAgentSkill:
		skill = &v
	case *A2AAgentCard:
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical",
			Cause:     errors.New("agent card contains multiple skills; use ToCanonicalProvider"),
		}
	default:
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical",
			Cause:     fmt.Errorf("unsupported type: %T", raw),
		}
	}

	return canonicalFromA2ASkill(skill)
}

// FromCanonical converts a canonical tool to an A2A AgentSkill.
// Returns *A2AAgentSkill.
func (a *A2AAdapter) FromCanonical(ct *CanonicalTool) (any, error) {
	if ct == nil {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "from_canonical",
			Cause:     errors.New("canonical tool is nil"),
		}
	}

	skill, err := a2aSkillFromCanonical(ct)
	if err != nil {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "from_canonical",
			Cause:     err,
		}
	}

	return skill, nil
}

// SupportsFeature returns whether this adapter supports a schema feature.
// A2A skill metadata does not carry JSON Schema, so features are not supported.
func (a *A2AAdapter) SupportsFeature(feature SchemaFeature) bool {
	return false
}

// ToCanonicalProvider converts an A2A AgentCard to a CanonicalProvider.
func (a *A2AAdapter) ToCanonicalProvider(raw any) (*CanonicalProvider, error) {
	if raw == nil {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical_provider",
			Cause:     errors.New("input is nil"),
		}
	}

	var card *A2AAgentCard
	switch v := raw.(type) {
	case *A2AAgentCard:
		card = v
	case A2AAgentCard:
		card = &v
	default:
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical_provider",
			Cause:     fmt.Errorf("unsupported type: %T", raw),
		}
	}

	if card.Name == "" || card.Description == "" || card.Version == "" {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical_provider",
			Cause:     errors.New("agent card name, description, and version are required"),
		}
	}

	if len(card.SupportedInterfaces) == 0 {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "to_canonical_provider",
			Cause:     errors.New("agent card supportedInterfaces is required"),
		}
	}

	provider := &CanonicalProvider{
		Name:                card.Name,
		Description:         card.Description,
		Version:             card.Version,
		Capabilities:        capabilitiesToMap(card.Capabilities),
		SecuritySchemes:     card.SecuritySchemes,
		SecurityRequirements: card.SecurityRequirements,
		DefaultInputModes:   card.DefaultInputModes,
		DefaultOutputModes:  card.DefaultOutputModes,
		SourceFormat:        "a2a",
		SourceMeta:          map[string]any{},
	}

	provider.Skills = make([]CanonicalTool, 0, len(card.Skills))
	for _, skill := range card.Skills {
		ct, err := canonicalFromA2ASkill(&skill)
		if err != nil {
			return nil, err
		}
		provider.Skills = append(provider.Skills, *ct)
	}

	provider.SourceMeta["supportedInterfaces"] = card.SupportedInterfaces
	if card.Provider != nil {
		provider.SourceMeta["provider"] = *card.Provider
	}
	if card.DocumentationURL != "" {
		provider.SourceMeta["documentationUrl"] = card.DocumentationURL
	}
	if card.IconURL != "" {
		provider.SourceMeta["iconUrl"] = card.IconURL
	}
	if len(card.Signatures) > 0 {
		provider.SourceMeta["signatures"] = card.Signatures
	}

	return provider, nil
}

// FromCanonicalProvider converts a CanonicalProvider to an A2A AgentCard.
func (a *A2AAdapter) FromCanonicalProvider(provider *CanonicalProvider) (*A2AAgentCard, error) {
	if provider == nil {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "from_canonical_provider",
			Cause:     errors.New("canonical provider is nil"),
		}
	}

	if provider.Name == "" || provider.Description == "" || provider.Version == "" {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "from_canonical_provider",
			Cause:     errors.New("provider name, description, and version are required"),
		}
	}

	supportedInterfaces, ok := provider.SourceMeta["supportedInterfaces"].([]A2AAgentInterface)
	if !ok || len(supportedInterfaces) == 0 {
		return nil, &ConversionError{
			Adapter:   a.Name(),
			Direction: "from_canonical_provider",
			Cause:     errors.New("supportedInterfaces required in SourceMeta"),
		}
	}

	card := &A2AAgentCard{
		Name:                 provider.Name,
		Description:          provider.Description,
		SupportedInterfaces:  supportedInterfaces,
		Version:              provider.Version,
		DocumentationURL:     stringFromMeta(provider.SourceMeta, "documentationUrl"),
		Capabilities:         capabilitiesFromMap(provider.Capabilities),
		SecuritySchemes:      provider.SecuritySchemes,
		SecurityRequirements: provider.SecurityRequirements,
		DefaultInputModes:    provider.DefaultInputModes,
		DefaultOutputModes:   provider.DefaultOutputModes,
		IconURL:              stringFromMeta(provider.SourceMeta, "iconUrl"),
	}

	if rawProvider, ok := provider.SourceMeta["provider"].(A2AAgentProvider); ok {
		card.Provider = &rawProvider
	}
	if rawProvider, ok := provider.SourceMeta["provider"].(*A2AAgentProvider); ok {
		card.Provider = rawProvider
	}
	if rawSignatures, ok := provider.SourceMeta["signatures"].([]map[string]any); ok {
		card.Signatures = rawSignatures
	}

	card.Skills = make([]A2AAgentSkill, 0, len(provider.Skills))
	for i := range provider.Skills {
		skill, err := a2aSkillFromCanonical(&provider.Skills[i])
		if err != nil {
			return nil, err
		}
		card.Skills = append(card.Skills, *skill)
	}

	return card, nil
}

func canonicalFromA2ASkill(skill *A2AAgentSkill) (*CanonicalTool, error) {
	if skill == nil {
		return nil, errors.New("agent skill is nil")
	}
	if skill.ID == "" {
		return nil, errors.New("agent skill id is required")
	}

	namespace, name, version := parseA2ASkillID(skill.ID)
	if name == "" {
		name = skill.ID
	}

	displayName := skill.Name
	if displayName == "" {
		displayName = name
	}

	ct := &CanonicalTool{
		Namespace:           namespace,
		Name:                name,
		Version:             version,
		DisplayName:         displayName,
		Description:         skill.Description,
		Tags:                skill.Tags,
		InputModes:          skill.InputModes,
		OutputModes:         skill.OutputModes,
		Examples:            skill.Examples,
		SecurityRequirements: skill.SecurityRequirements,
		InputSchema:          &JSONSchema{Type: "object"},
		SourceFormat:         "a2a",
		SourceMeta:           map[string]any{"skillId": skill.ID},
	}

	return ct, nil
}

func a2aSkillFromCanonical(ct *CanonicalTool) (*A2AAgentSkill, error) {
	if ct == nil {
		return nil, errors.New("canonical tool is nil")
	}
	if ct.Name == "" {
		return nil, errors.New("tool name is required")
	}

	skillID := skillIDFromCanonical(ct)
	if skillID == "" {
		return nil, errors.New("skill id is required")
	}

	name := ct.DisplayName
	if name == "" {
		name = ct.Name
	}

	return &A2AAgentSkill{
		ID:                   skillID,
		Name:                 name,
		Description:          ct.Description,
		Tags:                 ct.Tags,
		Examples:             ct.Examples,
		InputModes:           ct.InputModes,
		OutputModes:          ct.OutputModes,
		SecurityRequirements: ct.SecurityRequirements,
	}, nil
}

func parseA2ASkillID(id string) (namespace, name, version string) {
	parts := strings.Split(id, ":")
	switch len(parts) {
	case 1:
		return "", id, ""
	case 2:
		if parts[0] == "" || parts[1] == "" {
			return "", id, ""
		}
		return parts[0], parts[1], ""
	case 3:
		if parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return "", id, ""
		}
		return parts[0], parts[1], parts[2]
	default:
		return "", id, ""
	}
}

func skillIDFromCanonical(ct *CanonicalTool) string {
	if ct == nil {
		return ""
	}
	if ct.SourceMeta != nil {
		if id, ok := ct.SourceMeta["skillId"].(string); ok && id != "" {
			return id
		}
	}
	if ct.Namespace != "" && ct.Version != "" {
		return strings.Join([]string{ct.Namespace, ct.Name, ct.Version}, ":")
	}
	if ct.Namespace != "" {
		return strings.Join([]string{ct.Namespace, ct.Name}, ":")
	}
	return ct.Name
}

func capabilitiesToMap(cap A2AAgentCapabilities) map[string]any {
	out := map[string]any{}
	if cap.Streaming != nil {
		out["streaming"] = *cap.Streaming
	}
	if cap.PushNotifications != nil {
		out["pushNotifications"] = *cap.PushNotifications
	}
	if cap.ExtendedAgentCard != nil {
		out["extendedAgentCard"] = *cap.ExtendedAgentCard
	}
	if len(cap.Extensions) > 0 {
		out["extensions"] = cap.Extensions
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func capabilitiesFromMap(m map[string]any) A2AAgentCapabilities {
	if len(m) == 0 {
		return A2AAgentCapabilities{}
	}
	cap := A2AAgentCapabilities{}
	if v, ok := m["streaming"].(bool); ok {
		cap.Streaming = &v
	}
	if v, ok := m["pushNotifications"].(bool); ok {
		cap.PushNotifications = &v
	}
	if v, ok := m["extendedAgentCard"].(bool); ok {
		cap.ExtendedAgentCard = &v
	}
	if raw, ok := m["extensions"].([]A2AAgentExtension); ok {
		cap.Extensions = raw
	}
	return cap
}

func stringFromMeta(meta map[string]any, key string) string {
	if meta == nil {
		return ""
	}
	if v, ok := meta[key].(string); ok {
		return v
	}
	return ""
}
