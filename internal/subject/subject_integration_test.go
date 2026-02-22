//go:build integration

package subject_test

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

func TestSubjectAndCategoryCRUD(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	token := loginAndGetToken(t, srv.URL, admin.Email, admin.Password)

	catResp := getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/categories", token, schema, jsonBody(t, map[string]any{
		"name":        "Core Sciences",
		"description": "core category",
	})), http.StatusCreated)
	categoryID := fmt.Sprintf("%v", catResp["id"])
	if categoryID == "" {
		t.Fatalf("expected created category id")
	}

	categories := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/categories", token, schema, nil), http.StatusOK)
	categoryItems, ok := categories["items"].([]any)
	if !ok || len(categoryItems) < 1 {
		t.Fatalf("expected non-empty category list, got %v", categories["items"])
	}

	subjPayload := map[string]any{
		"name":           "Algorithms",
		"code":           "CS101",
		"description":    "intro algorithms",
		"category_id":    categoryID,
		"credits":        3,
		"hours_per_week": 3,
	}
	subjResp := getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/subjects", token, schema, jsonBody(t, subjPayload)), http.StatusCreated)
	subjectID := fmt.Sprintf("%v", subjResp["id"])
	if subjectID == "" {
		t.Fatalf("expected created subject id")
	}

	dupResp := getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/subjects", token, schema, jsonBody(t, subjPayload)), http.StatusConflict)
	if fmt.Sprintf("%v", dupResp["error"]) == "" {
		t.Fatalf("expected duplicate subject code conflict error message")
	}

	subjList := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/subjects?offset=0&limit=10", token, schema, nil), http.StatusOK)
	if subjList["total"].(float64) < 1 {
		t.Fatalf("expected total >= 1, got %v", subjList["total"])
	}

	filtered := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/subjects?category_id="+categoryID, token, schema, nil), http.StatusOK)
	if filtered["total"].(float64) < 1 {
		t.Fatalf("expected filtered subjects total >= 1, got %v", filtered["total"])
	}

	getResp := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/subjects/"+subjectID, token, schema, nil), http.StatusOK)
	if fmt.Sprintf("%v", getResp["code"]) != "CS101" {
		t.Fatalf("expected subject code CS101, got %v", getResp["code"])
	}

	updateResp := getJSON(t, mustAuthReq(t, http.MethodPut, srv.URL+"/api/v1/subjects/"+subjectID, token, schema, jsonBody(t, map[string]any{
		"name":           "Algorithms Advanced",
		"code":           "CS101A",
		"description":    "advanced",
		"category_id":    categoryID,
		"credits":        4,
		"hours_per_week": 4,
		"is_active":      true,
	})), http.StatusOK)
	if fmt.Sprintf("%v", updateResp["name"]) != "Algorithms Advanced" {
		t.Fatalf("expected updated subject name, got %v", updateResp["name"])
	}

	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/subjects", token, schema, jsonBody(t, map[string]any{"code": "NO_NAME"})), http.StatusBadRequest)
	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/categories", token, schema, jsonBody(t, map[string]any{})), http.StatusBadRequest)
}

func TestSubjectAndCategoryPermissions(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	readOnly := testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermSubjectRead})
	writeOnly := testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermSubjectWrite})
	noPerm := testutil.GenerateTestToken(t, uuid.New(), schema, nil)

	_, status := testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/subjects", readOnly, schema, nil))
	if status != http.StatusOK {
		t.Fatalf("expected read token to list subjects (200), got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/categories", writeOnly, schema, jsonBody(t, map[string]any{"name": "Cat Write"})))
	if status != http.StatusCreated {
		t.Fatalf("expected write token to create category (201), got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/categories", noPerm, schema, nil))
	if status != http.StatusForbidden {
		t.Fatalf("expected no-permission token forbidden on category read, got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/subjects", noPerm, schema, jsonBody(t, map[string]any{"name": "X", "code": "X1"})))
	if status != http.StatusForbidden {
		t.Fatalf("expected no-permission token forbidden on subject write, got %d", status)
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
