package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ErrInvalidToolID is returned when a tool ID string is malformed.
var ErrInvalidToolID = errors.New("invalid tool ID format")
var ErrInvalidTool = errors.New("invalid tool")
var ErrInvalidBackend = errors.New("invalid backend")

const (
	maxToolNameLen = 128
)

// MCPVersion is the MCP protocol version this package targets.
// Keep in sync with the latest MCP spec.
const MCPVersion = "2025-11-25"

// Decision Log:
// We evaluate the official MCP Go SDK (github.com/modelcontextprotocol/go-sdk/mcp)
// and choose to embed mcp.Tool in our Tool struct.
// - Usage: Embedding allows us to inherit all standard fields and JSON tags from the SDK,
//   ensuring 1:1 compatibility with the spec as interpreted by the official SDK.
// - Gaps: mcp.Tool uses `any` for InputSchema and OutputSchema, which is correct for
//   flexibility but requires us to handle validation carefully (which is a separate requirement).
//   mcp.Tool does not support Namespace or Version, so we add them.
// - Type Aliasing: We use type aliasing for ToolIcon (mcp.Icon) as it matches our needs.

// Tool mirrors the MCP Tool definition and adds Namespace and Version.
// It embeds mcp.Tool to ensure compatibility with the official SDK.
type Tool struct {
	mcp.Tool
	// Namespace provides a way to namespace tools, e.g. for stable IDs.
	Namespace string `json:"namespace,omitempty"`
	// Version is an optional version string for the tool.
	Version string `json:"version,omitempty"`
	// Tags is an optional set of search keywords for discovery layers (e.g. toolindex).
	Tags []string `json:"tags,omitempty"`
}

// ToolIcon is an alias for mcp.Icon from the official SDK.
type ToolIcon = mcp.Icon

// NormalizeTags normalizes a list of tags for indexing/search.
// Rules:
// - lowercase
// - trim whitespace
// - replace internal whitespace with '-'
// - allow only [a-z0-9-_.]
// - dedupe while preserving order
// - drop empty/invalid tags
// - max tag length: 64 chars
// - max tag count: 20
func NormalizeTags(tags []string) []string {
	const (
		maxTagLen   = 64
		maxTagCount = 20
	)
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))

	for _, raw := range tags {
		if len(out) >= maxTagCount {
			break
		}
		t := strings.TrimSpace(strings.ToLower(raw))
		if t == "" {
			continue
		}

		// Replace any whitespace run with '-'
		t = strings.Join(strings.Fields(t), "-")
		if t == "" {
			continue
		}

		// Filter to allowed characters.
		b := make([]byte, 0, len(t))
		for i := 0; i < len(t); i++ {
			c := t[i]
			if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' {
				b = append(b, c)
			}
		}
		if len(b) == 0 {
			continue
		}
		if len(b) > maxTagLen {
			b = b[:maxTagLen]
		}
		normalized := string(b)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

// BackendKind defines the type of backend backing a tool.
type BackendKind string

const (
	BackendKindMCP      BackendKind = "mcp"
	BackendKindProvider BackendKind = "provider"
	BackendKindLocal    BackendKind = "local"
)

// ToolBackend defines the binding information for a tool's execution.
// A tool can have multiple backends, but typically one active one.
type ToolBackend struct {
	Kind     BackendKind      `json:"kind"`
	MCP      *MCPBackend      `json:"mcp,omitempty"`
	Provider *ProviderBackend `json:"provider,omitempty"`
	Local    *LocalBackend    `json:"local,omitempty"`
}

// MCPBackend defines metadata for an MCP server backend.
type MCPBackend struct {
	// ServerName identifies the MCP server (e.g. in a registry or config).
	ServerName string `json:"serverName,omitempty"`
}

// ProviderBackend defines metadata for an external/manual tool provider.
type ProviderBackend struct {
	ProviderID string `json:"providerId"`
	ToolID     string `json:"toolId"`
}

// LocalBackend defines metadata for a locally executed tool.
type LocalBackend struct {
	// Name identifies the local function or handler.
	Name string `json:"name"`
}

// ToolID returns the canonical identifier for a tool.
// Format: "namespace:name" when namespace is present, otherwise just "name".
func (t *Tool) ToolID() string {
	if t.Namespace == "" {
		return t.Name
	}
	return t.Namespace + ":" + t.Name
}

// ParseToolID parses a tool ID string into namespace and name components.
// The format is "namespace:name" or just "name" (empty namespace).
// Returns an error if the ID is empty or contains multiple colons.
func ParseToolID(id string) (namespace, name string, err error) {
	if id == "" {
		return "", "", ErrInvalidToolID
	}

	// Count colons - we only allow at most one
	colonCount := strings.Count(id, ":")
	if colonCount > 1 {
		return "", "", ErrInvalidToolID
	}

	if colonCount == 0 {
		// No namespace, just the name
		return "", id, nil
	}

	// Split on the single colon
	parts := strings.SplitN(id, ":", 2)
	namespace = parts[0]
	name = parts[1]

	// Both namespace and name must be non-empty when colon is present
	if namespace == "" || name == "" {
		return "", "", ErrInvalidToolID
	}

	return namespace, name, nil
}

// Validate checks basic invariants of Tool required by toolmodel consumers.
// It does not validate JSON schemas; use SchemaValidator for that.
func (t *Tool) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidTool)
	}
	if len(t.Name) > maxToolNameLen {
		return fmt.Errorf("%w: name exceeds %d characters", ErrInvalidTool, maxToolNameLen)
	}
	var invalidChars []string
	seen := make(map[rune]bool)
	for _, r := range t.Name {
		if !validToolNameRune(r) {
			if !seen[r] {
				invalidChars = append(invalidChars, string(r))
				seen[r] = true
			}
		}
	}
	if len(invalidChars) > 0 {
		return fmt.Errorf("%w: name contains invalid characters: %s", ErrInvalidTool, strings.Join(invalidChars, ", "))
	}
	if t.InputSchema == nil {
		return fmt.Errorf("%w: inputSchema is required", ErrInvalidTool)
	}
	return nil
}

// Validate checks basic invariants of ToolBackend.
func (b ToolBackend) Validate() error {
	switch b.Kind {
	case BackendKindMCP:
		if b.MCP == nil || b.MCP.ServerName == "" {
			return fmt.Errorf("%w: MCP backend requires ServerName", ErrInvalidBackend)
		}
	case BackendKindProvider:
		if b.Provider == nil {
			return fmt.Errorf("%w: Provider backend requires Provider details", ErrInvalidBackend)
		}
		if b.Provider.ProviderID == "" {
			return fmt.Errorf("%w: Provider backend requires ProviderID", ErrInvalidBackend)
		}
		if b.Provider.ToolID == "" {
			return fmt.Errorf("%w: Provider backend requires ToolID", ErrInvalidBackend)
		}
	case BackendKindLocal:
		if b.Local == nil || b.Local.Name == "" {
			return fmt.Errorf("%w: Local backend requires Name", ErrInvalidBackend)
		}
	default:
		return fmt.Errorf("%w: unknown backend kind %q", ErrInvalidBackend, b.Kind)
	}
	return nil
}

func validToolNameRune(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_' || r == '-' || r == '.'
}

// ToMCPJSON serializes the Tool to JSON that is compatible with the MCP Tool spec.
// This strips toolmodel-specific fields (Namespace, Version) and returns only
// the standard MCP Tool fields.
func (t *Tool) ToMCPJSON() ([]byte, error) {
	return json.Marshal(t.Tool)
}

// ToJSON serializes the full Tool (including toolmodel extensions) to JSON.
func (t *Tool) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

// FromMCPJSON deserializes an MCP Tool JSON into a Tool struct.
// The Namespace and Version fields will be empty after this call.
func FromMCPJSON(data []byte) (*Tool, error) {
	var mcpTool mcp.Tool
	if err := json.Unmarshal(data, &mcpTool); err != nil {
		return nil, err
	}
	return &Tool{Tool: mcpTool}, nil
}

// FromJSON deserializes a full Tool JSON (including toolmodel extensions) into a Tool struct.
func FromJSON(data []byte) (*Tool, error) {
	var tool Tool
	if err := json.Unmarshal(data, &tool); err != nil {
		return nil, err
	}
	return &tool, nil
}
