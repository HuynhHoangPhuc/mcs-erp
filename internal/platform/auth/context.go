package auth

import (
	"context"
	"fmt"
)

type userCtxKey struct{}

// WithUser stores auth claims in the context.
func WithUser(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, userCtxKey{}, claims)
}

// UserFromContext extracts auth claims from the context.
func UserFromContext(ctx context.Context) (*Claims, error) {
	c, ok := ctx.Value(userCtxKey{}).(*Claims)
	if !ok || c == nil {
		return nil, fmt.Errorf("user not found in context")
	}
	return c, nil
}
