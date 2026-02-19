package infrastructure

import (
	"sync"

	agentdomain "github.com/HuynhHoangPhuc/mcs-erp/internal/agent/domain"
	coredomain "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
)

// ToolRegistry is a thread-safe registry where modules register AgentTools at startup.
// When building the LLM call, tools are filtered by the user's RBAC permissions.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]agentdomain.AgentTool
}

// NewToolRegistry creates an empty ToolRegistry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]agentdomain.AgentTool),
	}
}

// Register adds a tool to the registry. Panics on duplicate name (programming error at startup).
func (r *ToolRegistry) Register(t agentdomain.AgentTool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tools[t.Name()]; exists {
		panic("agent: duplicate tool registration: " + t.Name())
	}
	r.tools[t.Name()] = t
}

// GetTools returns all tools the user is permitted to use based on their permissions.
// Tools with an empty RequiredPermission are always included.
func (r *ToolRegistry) GetTools(userPerms []string) []agentdomain.AgentTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var permitted []agentdomain.AgentTool
	for _, t := range r.tools {
		req := t.RequiredPermission()
		if req == "" || coredomain.HasPermission(userPerms, req) {
			permitted = append(permitted, t)
		}
	}
	return permitted
}

// GetByName retrieves a specific tool by name, regardless of permissions.
// Callers must separately verify permission before executing the tool.
func (r *ToolRegistry) GetByName(name string) (agentdomain.AgentTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}
