//go:build integration

package hr_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"

	coredomain "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/testutil"
)

func TestTeacherCRUDAndAvailability(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	token := loginAndGetToken(t, srv.URL, admin.Email, admin.Password)
	departmentID := createDepartment(t, srv.URL, token, schema, "Engineering")

	teacherID := createTeacher(t, srv.URL, token, schema, map[string]any{
		"name":           "Alice Nguyen",
		"email":          fmt.Sprintf("alice_%s@example.com", uuid.NewString()),
		"department_id":  departmentID,
		"qualifications": []string{"MSc", "PhD"},
	})

	listResp := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/teachers?offset=0&limit=10", token, schema, nil), http.StatusOK)
	if listResp["total"].(float64) < 1 {
		t.Fatalf("expected at least 1 teacher, got %v", listResp["total"])
	}

	getResp := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/teachers/"+teacherID, token, schema, nil), http.StatusOK)
	if fmt.Sprintf("%v", getResp["id"]) != teacherID {
		t.Fatalf("expected teacher id %s, got %v", teacherID, getResp["id"])
	}

	updatePayload := map[string]any{
		"name":           "Alice Updated",
		"email":          fmt.Sprintf("alice_updated_%s@example.com", uuid.NewString()),
		"department_id":  departmentID,
		"qualifications": []string{"PhD"},
		"is_active":      true,
	}
	updateReq := mustAuthReq(t, http.MethodPut, srv.URL+"/api/v1/teachers/"+teacherID, token, schema, jsonBody(t, updatePayload))
	updated := getJSON(t, updateReq, http.StatusOK)
	if fmt.Sprintf("%v", updated["name"]) != "Alice Updated" {
		t.Fatalf("expected updated teacher name, got %v", updated["name"])
	}

	slots := []map[string]any{
		{"day": 1, "period": 2, "is_available": true},
		{"day": 1, "period": 3, "is_available": false},
	}
	setAvailReq := mustAuthReq(t, http.MethodPut, srv.URL+"/api/v1/teachers/"+teacherID+"/availability", token, schema, jsonBody(t, map[string]any{"slots": slots}))
	_ = getJSON(t, setAvailReq, http.StatusOK)

	availResp := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/teachers/"+teacherID+"/availability", token, schema, nil), http.StatusOK)
	gotSlots, ok := availResp["slots"].([]any)
	if !ok {
		t.Fatalf("expected slots array, got %T", availResp["slots"])
	}
	if len(gotSlots) != len(slots) {
		t.Fatalf("expected %d availability slots, got %d", len(slots), len(gotSlots))
	}

	invalidSlotReq := mustAuthReq(t, http.MethodPut, srv.URL+"/api/v1/teachers/"+teacherID+"/availability", token, schema, jsonBody(t, map[string]any{
		"slots": []map[string]any{{"day": 7, "period": 1, "is_available": true}},
	}))
	_ = getJSON(t, invalidSlotReq, http.StatusBadRequest)

	missingNameReq := mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/teachers", token, schema, jsonBody(t, map[string]any{"email": "x@example.com"}))
	_ = getJSON(t, missingNameReq, http.StatusBadRequest)
}

func TestTeacherEndpoints_RequirePermissions(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	readOnlyToken := testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermTeacherRead})
	writeOnlyToken := testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermTeacherWrite})
	noPermToken := testutil.GenerateTestToken(t, uuid.New(), schema, nil)

	readReq := mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/teachers", readOnlyToken, schema, nil)
	_, readStatus := testutil.DoJSON[map[string]any](t, http.DefaultClient, readReq)
	if readStatus != http.StatusOK {
		t.Fatalf("expected read-only token to read teachers (200), got %d", readStatus)
	}

	writeReq := mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/teachers", writeOnlyToken, schema, jsonBody(t, map[string]any{
		"name":           "Writer",
		"email":          fmt.Sprintf("writer_%s@example.com", uuid.NewString()),
		"qualifications": []string{"MSc"},
	}))
	_, writeStatus := testutil.DoJSON[map[string]any](t, http.DefaultClient, writeReq)
	if writeStatus != http.StatusCreated {
		t.Fatalf("expected write-only token to create teacher (201), got %d", writeStatus)
	}

	forbiddenReadReq := mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/teachers", noPermToken, schema, nil)
	_, forbiddenReadStatus := testutil.DoJSON[map[string]any](t, http.DefaultClient, forbiddenReadReq)
	if forbiddenReadStatus != http.StatusForbidden {
		t.Fatalf("expected no-permission token to get 403 on read, got %d", forbiddenReadStatus)
	}

	forbiddenWriteReq := mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/teachers", noPermToken, schema, jsonBody(t, map[string]any{
		"name":           "NoPerm",
		"email":          fmt.Sprintf("noperm_%s@example.com", uuid.NewString()),
		"qualifications": []string{"MSc"},
	}))
	_, forbiddenWriteStatus := testutil.DoJSON[map[string]any](t, http.DefaultClient, forbiddenWriteReq)
	if forbiddenWriteStatus != http.StatusForbidden {
		t.Fatalf("expected no-permission token to get 403 on write, got %d", forbiddenWriteStatus)
	}
}

func loginAndGetToken(t *testing.T, baseURL, email, password string) string {
	t.Helper()
	resp := postJSON(t, baseURL+"/api/v1/auth/login", map[string]string{"email": email, "password": password})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login expected 200, got %d", resp.StatusCode)
	}
	var pair map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&pair); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	token := fmt.Sprintf("%v", pair["access_token"])
	if token == "" {
		t.Fatalf("empty access token")
	}
	return token
}

func createDepartment(t *testing.T, baseURL, token, schema, name string) string {
	t.Helper()
	resp := getJSON(t, mustAuthReq(t, http.MethodPost, baseURL+"/api/v1/departments", token, schema, jsonBody(t, map[string]any{
		"name":        name,
		"description": "integration department",
	})), http.StatusCreated)
	id := fmt.Sprintf("%v", resp["id"])
	if id == "" {
		t.Fatalf("department id empty")
	}
	return id
}

func createTeacher(t *testing.T, baseURL, token, schema string, payload map[string]any) string {
	t.Helper()
	resp := getJSON(t, mustAuthReq(t, http.MethodPost, baseURL+"/api/v1/teachers", token, schema, jsonBody(t, payload)), http.StatusCreated)
	id := fmt.Sprintf("%v", resp["id"])
	if id == "" {
		t.Fatalf("teacher id empty")
	}
	return id
}

func postJSON(t *testing.T, url string, payload any) *http.Response {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post %s: %v", url, err)
	}
	return resp
}

func mustAuthReq(t *testing.T, method, url, token, schema string, body []byte) *http.Request {
	t.Helper()
	reader := bytes.NewReader(body)
	req, err := testutil.AuthenticatedRequest(method, url, token, schema, reader)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	return req
}

func jsonBody(t *testing.T, payload any) []byte {
	t.Helper()
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return b
}

func getJSON(t *testing.T, req *http.Request, expected int) map[string]any {
	t.Helper()
	payload, status := testutil.DoJSON[map[string]any](t, http.DefaultClient, req)
	if status != expected {
		t.Fatalf("expected status %d, got %d (payload=%v)", expected, status, payload)
	}
	return payload
}
