package tenant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var validSchema = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Resolve extracts the tenant schema name from the request.
// Primary: subdomain (faculty-a.mcs-erp.com → faculty_a).
// Fallback: X-Tenant-ID header.
func Resolve(r *http.Request) (string, error) {
	// Try subdomain first
	host := r.Host
	if idx := strings.IndexByte(host, ':'); idx != -1 {
		host = host[:idx]
	}

	parts := strings.SplitN(host, ".", 3)
	if len(parts) >= 3 {
		sub := parts[0]
		schema := strings.ReplaceAll(sub, "-", "_")
		if validSchema.MatchString(schema) {
			return schema, nil
		}
	}

	// Fallback: X-Tenant-ID header
	if tid := r.Header.Get("X-Tenant-ID"); tid != "" {
		schema := strings.ReplaceAll(tid, "-", "_")
		if validSchema.MatchString(schema) {
			return schema, nil
		}
		return "", fmt.Errorf("invalid tenant ID: %q", tid)
	}

	return "", fmt.Errorf("tenant not found: no subdomain or X-Tenant-ID header")
}
