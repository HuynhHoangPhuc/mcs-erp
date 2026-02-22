//go:build integration

package core_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/testutil"
)

func TestTenantIsolation_SeparateSchemas(t *testing.T) {
	db := testutil.NewTestDB(t)
	schemaA := db.CreateTenantSchema(t)
	schemaB := db.CreateTenantSchema(t)

	adminA := testutil.SeedAdmin(t, db.Pool, schemaA)
	adminB := testutil.SeedAdmin(t, db.Pool, schemaB)

	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	tokenA := loginAndGetToken(t, srv.URL, adminA.Email, adminA.Password)
	tokenB := loginAndGetToken(t, srv.URL, adminB.Email, adminB.Password)

	itemsA := listUsers(t, srv.URL, tokenA, schemaA)
	itemsB := listUsers(t, srv.URL, tokenB, schemaB)

	if len(itemsA) == 0 || len(itemsB) == 0 {
		t.Fatalf("expected users in both tenant lists")
	}

	if containsEmail(itemsA, adminB.Email) {
		t.Fatalf("tenant A unexpectedly sees tenant B user")
	}
	if containsEmail(itemsB, adminA.Email) {
		t.Fatalf("tenant B unexpectedly sees tenant A user")
	}
}

func TestTenantIsolation_CrossTenantRead_Fails(t *testing.T) {
	db := testutil.NewTestDB(t)
	schemaA := db.CreateTenantSchema(t)
	schemaB := db.CreateTenantSchema(t)

	adminA := testutil.SeedAdmin(t, db.Pool, schemaA)
	_ = testutil.SeedAdmin(t, db.Pool, schemaB)

	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	tokenA := loginAndGetToken(t, srv.URL, adminA.Email, adminA.Password)
	items := listUsers(t, srv.URL, tokenA, schemaB)

	if !containsEmail(items, adminA.Email) {
		t.Fatalf("expected tenant A user visible with tenant A token")
	}

	for _, it := range items {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		email := fmt.Sprintf("%v", m["email"])
		if email != "" && email != adminA.Email {
			t.Fatalf("unexpected cross-tenant data leakage: %s", email)
		}
	}
}

func TestTenantResolver_XTenantIDHeader(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost/api/v1/users", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("X-Tenant-ID", "test_tenant")

	schema, err := tenant.Resolve(req)
	if err != nil {
		t.Fatalf("resolve tenant: %v", err)
	}
	if schema != "test_tenant" {
		t.Fatalf("expected test_tenant, got %s", schema)
	}
}

func TestTenantResolver_SubdomainExtraction(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://faculty-a.mcs-erp.com/api/v1/users", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	schema, err := tenant.Resolve(req)
	if err != nil {
		t.Fatalf("resolve tenant: %v", err)
	}
	if schema != "faculty_a" {
		t.Fatalf("expected faculty_a, got %s", schema)
	}
}

func TestTenantResolver_MissingTenant_Returns400(t *testing.T) {
	db := testutil.NewTestDB(t)
	_ = db.CreateTenantSchema(t)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/users", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing tenant context, got %d", resp.StatusCode)
	}
}

func TestSchemaNameValidation_RejectsInjection(t *testing.T) {
	db := testutil.NewTestDB(t)
	cases := []string{"'; DROP TABLE users; --", "../../etc/passwd", "", "schema-name"}

	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			err := database.WithTenantTx(context.Background(), db.Pool, tc, func(tx pgx.Tx) error {
				return nil
			})
			if err == nil {
				t.Fatalf("expected error for invalid schema: %q", tc)
			}
		})
	}
}

func TestWithTenantTx_SetsSearchPath(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)

	err := database.WithTenantTx(context.Background(), db.Pool, schema, func(tx pgx.Tx) error {
		var searchPath string
		if err := tx.QueryRow(context.Background(), "SHOW search_path").Scan(&searchPath); err != nil {
			return err
		}
		if searchPath == "" {
			return fmt.Errorf("empty search_path")
		}
		if !strings.Contains(searchPath, schema) {
			return fmt.Errorf("search_path %q does not contain schema %q", searchPath, schema)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithTenantTx failed: %v", err)
	}
}

func listUsers(t *testing.T, baseURL, token, schema string) []any {
	t.Helper()

	req, err := testutil.AuthenticatedRequest(http.MethodGet, baseURL+"/api/v1/users", token, schema, nil)
	if err != nil {
		t.Fatalf("create users request: %v", err)
	}

	payload, status := testutil.DoJSON[map[string]any](t, http.DefaultClient, req)
	if status != http.StatusOK {
		t.Fatalf("expected users list status 200, got %d", status)
	}

	items, ok := payload["items"].([]any)
	if !ok {
		t.Fatalf("response items has unexpected type: %T", payload["items"])
	}
	return items
}

func containsEmail(items []any, email string) bool {
	for _, it := range items {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		if fmt.Sprintf("%v", m["email"]) == email {
			return true
		}
	}
	return false
}
