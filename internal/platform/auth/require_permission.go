package auth

import (
	"net/http"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
)

// RequirePermission returns middleware that checks if the authenticated user
// has the specified permission. Returns 403 if missing.
func RequirePermission(perm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := UserFromContext(r.Context())
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			if !domain.HasPermission(claims.Permissions, perm) {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
