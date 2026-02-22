//go:build integration

package agent_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent"
	agentsvc "github.com/HuynhHoangPhuc/mcs-erp/internal/agent/application/services"
	agentdomain "github.com/HuynhHoangPhuc/mcs-erp/internal/agent/domain"
	agentinfra "github.com/HuynhHoangPhuc/mcs-erp/internal/agent/infrastructure"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/core"
	coredelivery "github.com/HuynhHoangPhuc/mcs-erp/internal/core/delivery"
	coredomain "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/hr"
	platformmod "github.com/HuynhHoangPhuc/mcs-erp/internal/platform/module"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/room"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/subject"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/testutil"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable"
	pkgmod "github.com/HuynhHoangPhuc/mcs-erp/pkg/module"
)

func TestConversationCRUDAndOwnership(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	srv := testServerWithAgentMock(t, db.Pool)
	defer srv.Close()

	userA := uuid.New()
	userB := uuid.New()
	tokenA := testutil.GenerateTestToken(t, userA, schema, []string{coredomain.PermAgentChat})
	tokenB := testutil.GenerateTestToken(t, userB, schema, []string{coredomain.PermAgentChat})

	createResp := getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/agent/conversations", tokenA, schema, jsonBody(t, map[string]any{"title": "Test Chat"})), http.StatusCreated)
	convID := fmt.Sprintf("%v", createResp["id"])
	if convID == "" {
		t.Fatalf("expected conversation id")
	}

	listA := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/agent/conversations", tokenA, schema, nil), http.StatusOK)
	if listA["total"].(float64) < 1 {
		t.Fatalf("expected user A to see at least one conversation, got %v", listA["total"])
	}

	listB := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/agent/conversations", tokenB, schema, nil), http.StatusOK)
	if listB["total"].(float64) != 0 {
		t.Fatalf("expected user B to see 0 conversations from user A, got %v", listB["total"])
	}

	_ = getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/agent/conversations/"+convID, tokenA, schema, nil), http.StatusOK)
	_ = getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/agent/conversations/"+convID, tokenB, schema, nil), http.StatusForbidden)

	_ = getJSON(t, mustAuthReq(t, http.MethodPatch, srv.URL+"/api/v1/agent/conversations/"+convID, tokenA, schema, jsonBody(t, map[string]any{"title": "Updated Chat"})), http.StatusOK)
	_ = getJSON(t, mustAuthReq(t, http.MethodPatch, srv.URL+"/api/v1/agent/conversations/"+convID, tokenA, schema, jsonBody(t, map[string]any{"title": ""})), http.StatusBadRequest)

	_ = getJSON(t, mustAuthReq(t, http.MethodDelete, srv.URL+"/api/v1/agent/conversations/"+convID, tokenA, schema, nil), http.StatusOK)
	_ = getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/agent/conversations/"+convID, tokenA, schema, nil), http.StatusNotFound)
}

func TestChatSSEAndSuggestions(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	srv := testServerWithAgentMock(t, db.Pool)
	defer srv.Close()

	userID := uuid.New()
	token := testutil.GenerateTestToken(t, userID, schema, []string{coredomain.PermAgentChat})

	chatReq := mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/agent/chat", token, schema, jsonBody(t, map[string]any{"message": "hello agent"}))
	resp, err := http.DefaultClient.Do(chatReq)
	if err != nil {
		t.Fatalf("chat request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected chat status 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("expected text/event-stream content-type, got %q", ct)
	}
	if resp.Header.Get("X-Conversation-ID") == "" {
		t.Fatalf("expected X-Conversation-ID header")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read chat stream: %v", err)
	}
	stream := string(bodyBytes)
	if !strings.Contains(stream, "data: \"mock-response\"") {
		t.Fatalf("expected streamed mock token in SSE output, got %q", stream)
	}
	if !strings.Contains(stream, "data: [DONE]") {
		t.Fatalf("expected SSE done marker, got %q", stream)
	}

	sugg := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/agent/suggestions?entity_type=teacher&entity_id=abc", token, schema, nil), http.StatusOK)
	suggestions, ok := sugg["suggestions"].([]any)
	if !ok || len(suggestions) == 0 {
		t.Fatalf("expected non-empty suggestions, got %v", sugg["suggestions"])
	}

	emptySugg := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/agent/suggestions?entity_type=unknown", token, schema, nil), http.StatusOK)
	emptyList, ok := emptySugg["suggestions"].([]any)
	if !ok {
		t.Fatalf("expected suggestions array, got %T", emptySugg["suggestions"])
	}
	if len(emptyList) != 0 {
		t.Fatalf("expected empty suggestions for unknown entity, got %d", len(emptyList))
	}
}

func TestAgentPermissionRequired(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	srv := testServerWithAgentMock(t, db.Pool)
	defer srv.Close()

	noPerm := testutil.GenerateTestToken(t, uuid.New(), schema, nil)

	_, status := testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/agent/conversations", noPerm, schema, nil))
	if status != http.StatusForbidden {
		t.Fatalf("expected forbidden when missing agent permission, got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/agent/suggestions?entity_type=teacher", noPerm, schema, nil))
	if status != http.StatusForbidden {
		t.Fatalf("expected forbidden on suggestions when missing agent permission, got %d", status)
	}
}

func testServerWithAgentMock(t *testing.T, pool *pgxpool.Pool) *httptest.Server {
	t.Helper()

	registry := platformmod.NewRegistry()
	coreMod := core.NewModuleWithDeps(pool, testutil.TestJWTService())
	mustRegister(t, registry, coreMod)

	hrMod := hr.NewModule(pool, coreMod.AuthService())
	mustRegister(t, registry, hrMod)

	subjectMod := subject.NewModule(pool, coreMod.AuthService())
	mustRegister(t, registry, subjectMod)

	roomMod := room.NewModule(pool, coreMod.AuthService())
	mustRegister(t, registry, roomMod)

	timetableMod := timetable.NewModuleWithRepos(
		pool,
		coreMod.AuthService(),
		hrMod.TeacherRepo(),
		hrMod.AvailabilityRepo(),
		subjectMod.SubjectRepo(),
		roomMod.RoomRepo(),
		roomMod.RoomAvailabilityRepo(),
	)
	mustRegister(t, registry, timetableMod)

	toolRegistry := agentinfra.NewToolRegistry()
	providerSvc := agentsvc.NewProviderServiceWithLLM(agentdomain.LLMConfig{}, &testutil.MockLLMProvider{Response: "mock-response"})
	agentMod := agent.NewModuleWithProvider(pool, coreMod.AuthService(), toolRegistry, providerSvc, testutil.TestRedis(t))
	mustRegister(t, registry, agentMod)

	mux := http.NewServeMux()
	ctx := context.Background()
	if err := platformmod.Bootstrap(ctx, registry, mux); err != nil {
		t.Fatalf("bootstrap test server: %v", err)
	}

	handler := coredelivery.MaxBodySize(1 << 20)(tenant.Middleware(mux))
	return httptest.NewServer(handler)
}

func mustRegister(t *testing.T, registry *platformmod.Registry, m pkgmod.Module) {
	t.Helper()
	if err := registry.Register(m); err != nil {
		t.Fatalf("register module %s: %v", m.Name(), err)
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
