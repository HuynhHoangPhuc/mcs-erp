//go:build integration

package security_test

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

func TestIDOR_ConversationAccess_SameTenantForbidden(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	tokenA := testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermAgentChat})
	tokenB := testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermAgentChat})

	created := getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/agent/conversations", tokenA, schema, jsonBody(t, map[string]any{"title": "A convo"})), http.StatusCreated)
	convID := fmt.Sprintf("%v", created["id"])
	if convID == "" {
		t.Fatalf("expected conversation id")
	}

	_ = getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/agent/conversations/"+convID, tokenB, schema, nil), http.StatusForbidden)
}

func TestIDOR_ConversationAccess_CrossTenantNotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	schemaA := db.CreateTenantSchema(t)
	schemaB := db.CreateTenantSchema(t)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	tokenA := testutil.GenerateTestToken(t, uuid.New(), schemaA, []string{coredomain.PermAgentChat})
	tokenB := testutil.GenerateTestToken(t, uuid.New(), schemaB, []string{coredomain.PermAgentChat})

	created := getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/agent/conversations", tokenA, schemaA, jsonBody(t, map[string]any{"title": "Tenant A convo"})), http.StatusCreated)
	convID := fmt.Sprintf("%v", created["id"])
	if convID == "" {
		t.Fatalf("expected conversation id")
	}

	_ = getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/agent/conversations/"+convID, tokenB, schemaB, nil), http.StatusNotFound)
}

func TestJWTTenantClaimMismatchHeader_NoCrossTenantLeak(t *testing.T) {
	db := testutil.NewTestDB(t)
	schemaA := db.CreateTenantSchema(t)
	schemaB := db.CreateTenantSchema(t)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	teacherA := testutil.SeedTeacher(t, db.Pool, schemaA,
		testutil.WithTeacherEmail("tenant-a-only@example.com"),
		testutil.WithTeacherName("Tenant A Teacher"),
	)
	teacherB := testutil.SeedTeacher(t, db.Pool, schemaB,
		testutil.WithTeacherEmail("tenant-b-only@example.com"),
		testutil.WithTeacherName("Tenant B Teacher"),
	)

	tokenA := testutil.GenerateTestToken(t, uuid.New(), schemaA, []string{coredomain.PermTeacherRead})

	payload := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/teachers?limit=100", tokenA, schemaB, nil), http.StatusOK)
	emails := teacherEmails(payload["items"])
	if !containsString(emails, teacherA.Email) {
		t.Fatalf("expected tenant A teacher email %q in response, got %v", teacherA.Email, emails)
	}
	if containsString(emails, teacherB.Email) {
		t.Fatalf("expected no tenant B data leakage; got email %q in response %v", teacherB.Email, emails)
	}
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

func teacherEmails(items any) []string {
	arr, ok := items.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		email := fmt.Sprintf("%v", m["email"])
		if email != "" {
			out = append(out, email)
		}
	}
	return out
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
