package domain

import "context"

// AgentTool wraps a callable tool with RBAC metadata.
// Modules implement this interface and register tools at startup via ToolRegistry.
type AgentTool interface {
	// Name returns the tool identifier used by the LLM (must be unique).
	Name() string
	// Description explains what the tool does (shown to LLM in system prompt).
	Description() string
	// RequiredPermission returns the RBAC permission string needed to use this tool.
	// Empty string means no permission required.
	RequiredPermission() string
	// Call executes the tool with the given JSON-encoded input and returns JSON-encoded output.
	Call(ctx context.Context, input string) (string, error)
}
