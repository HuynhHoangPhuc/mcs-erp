//go:build integration

package room_test

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

func TestRoomCRUDAndAvailability(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	token := loginAndGetToken(t, srv.URL, admin.Email, admin.Password)

	roomResp := getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/rooms", token, schema, jsonBody(t, map[string]any{
		"name":      "Auditorium A",
		"code":      "RA-101",
		"building":  "Main",
		"floor":     1,
		"capacity":  60,
		"equipment": []string{"projector", "whiteboard"},
	})), http.StatusCreated)
	roomID := fmt.Sprintf("%v", roomResp["id"])
	if roomID == "" {
		t.Fatalf("expected room id")
	}

	dupResp := getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/rooms", token, schema, jsonBody(t, map[string]any{
		"name":      "Auditorium A Duplicate",
		"code":      "RA-101",
		"building":  "Main",
		"floor":     1,
		"capacity":  50,
		"equipment": []string{"projector"},
	})), http.StatusInternalServerError)
	if fmt.Sprintf("%v", dupResp["error"]) == "" {
		t.Fatalf("expected duplicate room create error payload")
	}

	listResp := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/rooms", token, schema, nil), http.StatusOK)
	if listResp["total"].(float64) < 1 {
		t.Fatalf("expected rooms total >= 1, got %v", listResp["total"])
	}

	byID := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/rooms/"+roomID, token, schema, nil), http.StatusOK)
	if fmt.Sprintf("%v", byID["code"]) != "RA-101" {
		t.Fatalf("expected room code RA-101, got %v", byID["code"])
	}

	updated := getJSON(t, mustAuthReq(t, http.MethodPut, srv.URL+"/api/v1/rooms/"+roomID, token, schema, jsonBody(t, map[string]any{
		"name":      "Auditorium A Updated",
		"code":      "RA-101U",
		"building":  "Main",
		"floor":     2,
		"capacity":  80,
		"equipment": []string{"projector", "audio"},
		"is_active": true,
	})), http.StatusOK)
	if fmt.Sprintf("%v", updated["name"]) != "Auditorium A Updated" {
		t.Fatalf("expected updated room name, got %v", updated["name"])
	}

	_ = getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/rooms?building=Main&min_capacity=70", token, schema, nil), http.StatusOK)
	_ = getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/rooms?equipment=audio", token, schema, nil), http.StatusOK)

	slots := []map[string]any{{"day": 1, "period": 2, "is_available": true}, {"day": 3, "period": 4, "is_available": false}}
	_ = getJSON(t, mustAuthReq(t, http.MethodPut, srv.URL+"/api/v1/rooms/"+roomID+"/availability", token, schema, jsonBody(t, map[string]any{"slots": slots})), http.StatusOK)

	avail := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/rooms/"+roomID+"/availability", token, schema, nil), http.StatusOK)
	availSlots, ok := avail["slots"].([]any)
	if !ok || len(availSlots) != len(slots) {
		t.Fatalf("expected %d room availability slots, got %v", len(slots), avail["slots"])
	}

	_ = getJSON(t, mustAuthReq(t, http.MethodPut, srv.URL+"/api/v1/rooms/"+roomID+"/availability", token, schema, jsonBody(t, map[string]any{
		"slots": []map[string]any{{"day": 9, "period": 1, "is_available": true}},
	})), http.StatusBadRequest)

	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/rooms", token, schema, jsonBody(t, map[string]any{"name": "X", "code": "Y", "capacity": 0})), http.StatusBadRequest)
}

func TestRoomPermissions(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	readOnly := testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermRoomRead})
	writeOnly := testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermRoomWrite})
	noPerm := testutil.GenerateTestToken(t, uuid.New(), schema, nil)

	_, status := testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/rooms", readOnly, schema, nil))
	if status != http.StatusOK {
		t.Fatalf("expected read token to list rooms (200), got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/rooms", writeOnly, schema, jsonBody(t, map[string]any{
		"name":      "Writer Room",
		"code":      "WR-1",
		"building":  "B1",
		"floor":     1,
		"capacity":  40,
		"equipment": []string{"projector"},
	})))
	if status != http.StatusCreated {
		t.Fatalf("expected write token to create room (201), got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/rooms", noPerm, schema, nil))
	if status != http.StatusForbidden {
		t.Fatalf("expected no-permission token forbidden on read, got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/rooms", noPerm, schema, jsonBody(t, map[string]any{
		"name":      "NoPerm Room",
		"code":      "NP-1",
		"building":  "B1",
		"floor":     1,
		"capacity":  40,
		"equipment": []string{"projector"},
	})))
	if status != http.StatusForbidden {
		t.Fatalf("expected no-permission token forbidden on write, got %d", status)
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
