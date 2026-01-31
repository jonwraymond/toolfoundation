// Package toolmodel provides a single, canonical data model for tools in the system,
// based on the MCP Tool specification (2025-11-25).
// Target protocol version: MCPVersion.
//
// This package is the only place that defines what a "tool" is (fields, constraints, IDs),
// and all downstream modules depend on it for tool definitions. It provides:
//
//   - Tool and ToolIcon types matching the MCP Tool specification
//   - Namespace and Version extensions for stable tool identification
//   - Backend binding types (MCP, Provider, Local) for execution metadata
//   - Optional tool Tags for search/discovery layers
//   - JSON Schema validation helpers for inputs and outputs
//   - JSON serialization compatible with the MCP Tool spec
//   - Tag normalization helpers for discovery layers
//
// The package enforces MCP schema rules by default:
//
//   - inputSchema MUST be a valid JSON Schema object
//   - Schemas may be provided as map[string]any, json.RawMessage, or []byte
//   - JSON Schema 2020-12 is assumed when no $schema is present
//   - External $ref resolution is disabled to prevent network access
//
// This package has no networking or JSON-RPC dependencies and is safe to embed
// in public APIs.
//
// Tool name validation mirrors MCP rules: 1-128 chars, [A-Za-z0-9_.-] only.
package model
