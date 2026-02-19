package module

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

// Bootstrap resolves module order and initializes each module in sequence:
// migrate → register routes → register events.
func Bootstrap(ctx context.Context, reg *Registry, mux *http.ServeMux) error {
	modules, err := reg.ResolveOrder()
	if err != nil {
		return fmt.Errorf("resolve module order: %w", err)
	}

	for _, m := range modules {
		slog.Info("bootstrapping module", "module", m.Name())

		if err := m.Migrate(ctx); err != nil {
			return fmt.Errorf("migrate module %q: %w", m.Name(), err)
		}

		m.RegisterRoutes(mux)

		if err := m.RegisterEvents(ctx); err != nil {
			return fmt.Errorf("register events for module %q: %w", m.Name(), err)
		}

		slog.Info("module ready", "module", m.Name())
	}

	return nil
}
