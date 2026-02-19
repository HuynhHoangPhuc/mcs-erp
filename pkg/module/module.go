package module

import (
	"context"
	"net/http"
)

// Module defines the interface every ERP module must implement.
// Modules are registered at compile time and started in dependency order.
type Module interface {
	// Name returns the unique module identifier (e.g. "core", "hr", "subject").
	Name() string

	// Dependencies returns names of modules this one depends on.
	Dependencies() []string

	// RegisterRoutes mounts the module's REST handlers on the given mux.
	RegisterRoutes(mux *http.ServeMux)

	// RegisterEvents sets up event subscriptions (Watermill handlers).
	RegisterEvents(ctx context.Context) error

	// Migrate runs database migrations for the module.
	Migrate(ctx context.Context) error
}
