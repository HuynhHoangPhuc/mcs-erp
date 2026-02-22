package testutil

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent"
	agentsvc "github.com/HuynhHoangPhuc/mcs-erp/internal/agent/application/services"
	agentdomain "github.com/HuynhHoangPhuc/mcs-erp/internal/agent/domain"
	agentinfra "github.com/HuynhHoangPhuc/mcs-erp/internal/agent/infrastructure"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/core"
	coredelivery "github.com/HuynhHoangPhuc/mcs-erp/internal/core/delivery"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/hr"
	platformmod "github.com/HuynhHoangPhuc/mcs-erp/internal/platform/module"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/room"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/subject"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable"
	pkgmod "github.com/HuynhHoangPhuc/mcs-erp/pkg/module"
)

// TestServer creates an HTTP server wired with all modules against the test DB.
func TestServer(t *testing.T, pool *pgxpool.Pool) *httptest.Server {
	t.Helper()

	registry := platformmod.NewRegistry()
	coreMod := core.NewModuleWithDeps(pool, TestJWTService())
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
	providerSvc := agentsvc.NewProviderServiceWithLLM(agentdomain.LLMConfig{}, &MockLLMProvider{Response: "mock-response"})
	agentMod := agent.NewModuleWithProvider(pool, coreMod.AuthService(), toolRegistry, providerSvc, TestRedis(t))
	mustRegister(t, registry, agentMod)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := platformmod.Bootstrap(ctx, registry, mux); err != nil {
		t.Fatalf("bootstrap test modules: %v", err)
	}

	handler := coredelivery.MaxBodySize(1 << 20)(tenant.Middleware(mux))
	return httptest.NewServer(handler)
}

// AuthenticatedRequest creates a request with bearer token and tenant header.
func AuthenticatedRequest(method, url, token, tenantID string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", tenantID)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// DoJSON executes a request and decodes response body into T.
func DoJSON[T any](t *testing.T, client *http.Client, req *http.Request) (T, int) {
	t.Helper()

	var out T
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil && err != io.EOF {
		t.Fatalf("decode json response: %v", err)
	}
	return out, resp.StatusCode
}

func mustRegister(t *testing.T, registry *platformmod.Registry, m pkgmod.Module) {
	t.Helper()
	if err := registry.Register(m); err != nil {
		t.Fatalf("register module %s: %v", m.Name(), err)
	}
}
