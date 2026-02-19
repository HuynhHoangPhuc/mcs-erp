package tenant

import (
	"encoding/json"
	"net/http"
	"strings"
)

// publicPaths are routes that skip tenant resolution.
var publicPaths = []string{"/healthz", "/api/v1/auth/login", "/api/v1/auth/register"}

// Middleware resolves the tenant from the request and injects it into context.
// Public paths (healthz, login) skip tenant resolution.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, p := range publicPaths {
			if strings.HasPrefix(r.URL.Path, p) {
				next.ServeHTTP(w, r)
				return
			}
		}

		schema, err := Resolve(r)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "tenant resolution failed: " + err.Error()})
			return
		}

		ctx := WithTenant(r.Context(), schema)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
