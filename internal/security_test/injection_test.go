//go:build integration

package security_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"

	coredomain "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/testutil"
)

func TestSQLInjection_TeacherListQueryParams_DoNotCrash(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	token := testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermTeacherRead})
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	payloads := []string{
		"'; DROP TABLE users; --",
		"1 OR 1=1",
		`\" OR \"\"=\"`,
		"1; SELECT * FROM information_schema.tables",
		"admin'--",
	}

	for _, payload := range payloads {
		t.Run(payload, func(t *testing.T) {
			requestURL := srv.URL + "/api/v1/teachers?qualification=" + url.QueryEscape(payload)
			req := mustAuthReq(t, http.MethodGet, requestURL, token, schema, nil)
			_, status := testutil.DoJSON[map[string]any](t, http.DefaultClient, req)
			if status != http.StatusOK && status != http.StatusBadRequest {
				t.Fatalf("expected 200 or 400 for payload %q, got %d", payload, status)
			}
		})
	}
}

func TestSchemaInjection_XTenantIDHeader_Rejected(t *testing.T) {
	db := testutil.NewTestDB(t)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	payloads := []string{
		"'; DROP SCHEMA public CASCADE; --",
		"test; DROP TABLE users",
		"../../../etc/passwd",
		"",
		"a" + strings.Repeat("!", 1000),
	}

	for _, payload := range payloads {
		name := payload
		if name == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/users", nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			if payload != "" {
				req.Header.Set("X-Tenant-ID", payload)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("execute request: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected 400 for tenant payload %q, got %d", payload, resp.StatusCode)
			}
		})
	}
}

func TestSchemaInjection_SetTenantSchema_RejectsInvalidNames(t *testing.T) {
	db := testutil.NewTestDB(t)

	invalid := []string{
		"tenant-name", // hyphen not allowed by regex
		"1starts_with_digit",
		"semi;colon",
		"with space",
		"quoted\"name",
		"",
	}

	for _, schema := range invalid {
		t.Run(fmt.Sprintf("schema=%q", schema), func(t *testing.T) {
			tx, err := db.Pool.Begin(context.Background())
			if err != nil {
				t.Fatalf("begin tx: %v", err)
			}
			defer tx.Rollback(context.Background())

			err = database.SetTenantSchema(context.Background(), tx, schema)
			if err == nil {
				t.Fatalf("expected invalid schema %q to be rejected", schema)
			}
		})
	}
}

func TestJWTWrongSigningMethod_Rejected(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	badToken := makeJWTNoneToken(t, schema, []string{coredomain.PermTeacherRead})
	req := mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/teachers", badToken, schema, nil)
	_, status := testutil.DoJSON[map[string]any](t, http.DefaultClient, req)
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401 for JWT with wrong signing method, got %d", status)
	}
}

func makeJWTNoneToken(t *testing.T, tenantID string, perms []string) string {
	t.Helper()
	header := map[string]any{"alg": "none", "typ": "JWT"}
	payload := map[string]any{
		"sub":         uuid.NewString(),
		"jti":         uuid.NewString(),
		"user_id":     uuid.NewString(),
		"tenant_id":   tenantID,
		"email":       "none@example.com",
		"permissions": perms,
	}
	return jwtSegment(t, header) + "." + jwtSegment(t, payload) + "."
}

func jwtSegment(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal jwt segment: %v", err)
	}
	return base64RawURLEncode(b)
}

func base64RawURLEncode(in []byte) string {
	const table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	out := make([]byte, 0, (len(in)*4+2)/3)
	for i := 0; i < len(in); i += 3 {
		var b0, b1, b2 byte
		b0 = in[i]
		have1 := i+1 < len(in)
		have2 := i+2 < len(in)
		if have1 {
			b1 = in[i+1]
		}
		if have2 {
			b2 = in[i+2]
		}
		out = append(out, table[b0>>2])
		out = append(out, table[((b0&0x03)<<4)|(b1>>4)])
		if have1 {
			out = append(out, table[((b1&0x0f)<<2)|(b2>>6)])
		}
		if have2 {
			out = append(out, table[b2&0x3f])
		}
	}
	return string(out)
}
