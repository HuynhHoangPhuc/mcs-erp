package delivery

import (
	"net/http"
	"strings"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/application/services"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
)

// AuthMiddleware validates JWT from Authorization header and sets user+tenant in context.
func AuthMiddleware(authSvc *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid authorization header"})
				return
			}

			token := strings.TrimPrefix(header, "Bearer ")
			claims, err := authSvc.ValidateToken(token)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
				return
			}

			// Set both user claims and tenant in context
			ctx := auth.WithUser(r.Context(), claims)
			ctx = tenant.WithTenant(ctx, claims.TenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
