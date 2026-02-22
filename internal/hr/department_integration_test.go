//go:build integration

package hr_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"

	coredomain "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/testutil"
)

func TestDepartmentCRUD(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	token := loginAndGetToken(t, srv.URL, admin.Email, admin.Password)

	createResp := getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/departments", token, schema, jsonBody(t, map[string]any{
		"name":        "Department A",
		"description": "department for integration test",
	})), http.StatusCreated)
	deptID := fmt.Sprintf("%v", createResp["id"])
	if deptID == "" {
		t.Fatalf("expected department id")
	}

	listResp := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/departments", token, schema, nil), http.StatusOK)
	if listResp["total"].(float64) < 1 {
		t.Fatalf("expected total >= 1, got %v", listResp["total"])
	}

	getResp := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/departments/"+deptID, token, schema, nil), http.StatusOK)
	if fmt.Sprintf("%v", getResp["name"]) != "Department A" {
		t.Fatalf("expected department name Department A, got %v", getResp["name"])
	}

	_ = getJSON(t, mustAuthReq(t, http.MethodPut, srv.URL+"/api/v1/departments/"+deptID, token, schema, jsonBody(t, map[string]any{
		"name":        "Department B",
		"description": "updated",
	})), http.StatusOK)

	updated := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/departments/"+deptID, token, schema, nil), http.StatusOK)
	if fmt.Sprintf("%v", updated["name"]) != "Department B" {
		t.Fatalf("expected updated department name Department B, got %v", updated["name"])
	}

	_ = getJSON(t, mustAuthReq(t, http.MethodDelete, srv.URL+"/api/v1/departments/"+deptID, token, schema, nil), http.StatusOK)
	_ = getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/departments/"+deptID, token, schema, nil), http.StatusNotFound)
}

func TestDepartmentValidationAndPermissions(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	readOnly := testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermDeptRead})
	writeOnly := testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermDeptWrite})
	noPerm := testutil.GenerateTestToken(t, uuid.New(), schema, nil)

	_, status := testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/departments", writeOnly, schema, jsonBody(t, map[string]any{
		"name": "Only Writer",
	})))
	if status != http.StatusCreated {
		t.Fatalf("expected writer to create department (201), got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/departments", readOnly, schema, nil))
	if status != http.StatusOK {
		t.Fatalf("expected reader to list departments (200), got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/departments", noPerm, schema, nil))
	if status != http.StatusForbidden {
		t.Fatalf("expected no-permission token to be forbidden on read, got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/departments", noPerm, schema, jsonBody(t, map[string]any{"name": "NoPerm"})))
	if status != http.StatusForbidden {
		t.Fatalf("expected no-permission token to be forbidden on write, got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/departments", writeOnly, schema, jsonBody(t, map[string]any{})))
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 when department name is missing, got %d", status)
	}
}
