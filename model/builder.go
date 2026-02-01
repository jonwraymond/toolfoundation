package model

import "github.com/modelcontextprotocol/go-sdk/mcp"

// ToolBuilder provides a fluent API for constructing Tool instances.
// Use NewTool to create a builder, chain methods, and call Build() or MustBuild().
type ToolBuilder struct {
	tool Tool
}

// NewTool creates a new ToolBuilder with the given tool name.
// The name is required and must follow tool naming conventions.
func NewTool(name string) *ToolBuilder {
	return &ToolBuilder{
		tool: Tool{
			Tool: mcp.Tool{
				Name: name,
			},
		},
	}
}

// Description sets the tool's description.
func (b *ToolBuilder) Description(desc string) *ToolBuilder {
	b.tool.Description = desc
	return b
}

// Namespace sets the tool's namespace for scoped identification.
func (b *ToolBuilder) Namespace(ns string) *ToolBuilder {
	b.tool.Namespace = ns
	return b
}

// Version sets the tool's version string.
func (b *ToolBuilder) Version(v string) *ToolBuilder {
	b.tool.Version = v
	return b
}

// Tags sets the tool's discovery tags.
// Tags are normalized during Build().
func (b *ToolBuilder) Tags(tags ...string) *ToolBuilder {
	b.tool.Tags = tags
	return b
}

// InputSchema sets the tool's input schema.
// The schema should be a JSON Schema compliant map.
func (b *ToolBuilder) InputSchema(schema map[string]any) *ToolBuilder {
	b.tool.InputSchema = schema
	return b
}

// OutputSchema sets the tool's output schema.
// The schema should be a JSON Schema compliant map.
func (b *ToolBuilder) OutputSchema(schema map[string]any) *ToolBuilder {
	b.tool.OutputSchema = schema
	return b
}

// Title sets the tool's display title.
func (b *ToolBuilder) Title(title string) *ToolBuilder {
	b.tool.Title = title
	return b
}

// Icons sets the tool's icons for display.
func (b *ToolBuilder) Icons(icons ...mcp.Icon) *ToolBuilder {
	b.tool.Icons = icons
	return b
}

// Annotations sets the tool's MCP annotations.
func (b *ToolBuilder) Annotations(annotations *mcp.ToolAnnotations) *ToolBuilder {
	b.tool.Annotations = annotations
	return b
}

// Meta sets the tool's metadata.
func (b *ToolBuilder) Meta(meta mcp.Meta) *ToolBuilder {
	b.tool.Meta = meta
	return b
}

// ReadOnly marks the tool as read-only via annotations.
func (b *ToolBuilder) ReadOnly() *ToolBuilder {
	b.ensureAnnotations()
	b.tool.Annotations.ReadOnlyHint = true
	return b
}

// Idempotent marks the tool as idempotent via annotations.
func (b *ToolBuilder) Idempotent() *ToolBuilder {
	b.ensureAnnotations()
	b.tool.Annotations.IdempotentHint = true
	return b
}

// Destructive marks the tool as destructive via annotations.
func (b *ToolBuilder) Destructive() *ToolBuilder {
	b.ensureAnnotations()
	v := true
	b.tool.Annotations.DestructiveHint = &v
	return b
}

// NonDestructive explicitly marks the tool as non-destructive via annotations.
func (b *ToolBuilder) NonDestructive() *ToolBuilder {
	b.ensureAnnotations()
	v := false
	b.tool.Annotations.DestructiveHint = &v
	return b
}

// OpenWorld marks the tool as potentially interacting with external world.
func (b *ToolBuilder) OpenWorld() *ToolBuilder {
	b.ensureAnnotations()
	v := true
	b.tool.Annotations.OpenWorldHint = &v
	return b
}

// ensureAnnotations initializes Annotations if nil.
func (b *ToolBuilder) ensureAnnotations() {
	if b.tool.Annotations == nil {
		b.tool.Annotations = &mcp.ToolAnnotations{}
	}
}

// Build validates and returns the constructed Tool.
// Returns an error if validation fails.
func (b *ToolBuilder) Build() (*Tool, error) {
	if len(b.tool.Tags) > 0 {
		b.tool.Tags = NormalizeTags(b.tool.Tags)
	}

	if err := b.tool.Validate(); err != nil {
		return nil, err
	}

	result := b.tool
	return &result, nil
}

// MustBuild validates and returns the constructed Tool, panicking on error.
// Use only in tests or when the builder configuration is known to be valid.
func (b *ToolBuilder) MustBuild() *Tool {
	tool, err := b.Build()
	if err != nil {
		panic("MustBuild: " + err.Error())
	}
	return tool
}
