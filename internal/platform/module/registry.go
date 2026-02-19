package module

import (
	"fmt"

	pkgmod "github.com/HuynhHoangPhuc/mcs-erp/pkg/module"
)

// Registry stores modules and resolves their startup order via topological sort.
type Registry struct {
	modules map[string]pkgmod.Module
}

// NewRegistry creates an empty module registry.
func NewRegistry() *Registry {
	return &Registry{modules: make(map[string]pkgmod.Module)}
}

// Register adds a module. Returns error on duplicate name.
func (r *Registry) Register(m pkgmod.Module) error {
	name := m.Name()
	if _, exists := r.modules[name]; exists {
		return fmt.Errorf("module %q already registered", name)
	}
	r.modules[name] = m
	return nil
}

// ResolveOrder returns modules in dependency order using Kahn's algorithm.
// Returns error on circular or missing dependencies.
func (r *Registry) ResolveOrder() ([]pkgmod.Module, error) {
	// Build in-degree map
	inDegree := make(map[string]int)
	dependents := make(map[string][]string) // dep -> modules that depend on it

	for name := range r.modules {
		inDegree[name] = 0
	}

	for name, m := range r.modules {
		for _, dep := range m.Dependencies() {
			if _, ok := r.modules[dep]; !ok {
				return nil, fmt.Errorf("module %q depends on unregistered module %q", name, dep)
			}
			inDegree[name]++
			dependents[dep] = append(dependents[dep], name)
		}
	}

	// Kahn's algorithm
	var queue []string
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	var order []pkgmod.Module
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		order = append(order, r.modules[curr])

		for _, dep := range dependents[curr] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if len(order) != len(r.modules) {
		return nil, fmt.Errorf("circular dependency detected among modules")
	}

	return order, nil
}

// Get returns a module by name, or nil if not found.
func (r *Registry) Get(name string) pkgmod.Module {
	return r.modules[name]
}
