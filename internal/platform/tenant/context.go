package tenant

import (
	"context"
	"fmt"
)

type ctxKey struct{}

// WithTenant stores the tenant schema name in the context.
func WithTenant(ctx context.Context, schema string) context.Context {
	return context.WithValue(ctx, ctxKey{}, schema)
}

// FromContext extracts the tenant schema name from the context.
// Returns error if no tenant is set.
func FromContext(ctx context.Context) (string, error) {
	v, ok := ctx.Value(ctxKey{}).(string)
	if !ok || v == "" {
		return "", fmt.Errorf("tenant not found in context")
	}
	return v, nil
}
