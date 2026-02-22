//go:build integration

package platform_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent"
	agentsvc "github.com/HuynhHoangPhuc/mcs-erp/internal/agent/application/services"
	agentdomain "github.com/HuynhHoangPhuc/mcs-erp/internal/agent/domain"
	agentinfra "github.com/HuynhHoangPhuc/mcs-erp/internal/agent/infrastructure"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/core"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/hr"
	platformmod "github.com/HuynhHoangPhuc/mcs-erp/internal/platform/module"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/room"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/subject"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/testutil"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable"
	pkgmod "github.com/HuynhHoangPhuc/mcs-erp/pkg/module"
)

func TestHealthCheck_Returns200(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	_ = testutil.SeedAdmin(t, db.Pool, schema)

	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("request /healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read /healthz response: %v", err)
	}
	if !strings.Contains(string(payload), `"status":"ok"`) {
		t.Fatalf("expected health payload to contain status ok, got: %s", string(payload))
	}
}

func TestModuleRegistry_ResolvesOrder(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	_ = testutil.SeedAdmin(t, db.Pool, schema)

	registry := registerAllModulesForSmoke(t, db)
	order, err := registry.ResolveOrder()
	if err != nil {
		t.Fatalf("resolve module order: %v", err)
	}

	index := map[string]int{}
	for i, m := range order {
		index[m.Name()] = i
	}

	assertBefore(t, index, "core", "hr")
	assertBefore(t, index, "core", "subject")
	assertBefore(t, index, "core", "room")
	assertBefore(t, index, "core", "agent")
	assertBefore(t, index, "hr", "timetable")
	assertBefore(t, index, "subject", "timetable")
	assertBefore(t, index, "room", "timetable")
}

func TestModuleBootstrap_AllModulesStart(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	_ = testutil.SeedAdmin(t, db.Pool, schema)

	registry := registerAllModulesForSmoke(t, db)
	mux := http.NewServeMux()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := platformmod.Bootstrap(ctx, registry, mux); err != nil {
		t.Fatalf("bootstrap all modules: %v", err)
	}
}

func TestMigrations_ApplyCleanly(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)

	expected := []string{
		"users", "roles", "user_roles",
		"departments", "teachers", "teacher_availability",
		"subjects", "subject_categories", "subject_prerequisites",
		"rooms", "room_availability",
		"semesters", "semester_subjects", "assignments",
		"conversations", "messages",
	}

	for _, table := range expected {
		var exists bool
		err := db.Pool.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = $1 AND table_name = $2
			)`,
			schema,
			table,
		).Scan(&exists)
		if err != nil {
			t.Fatalf("check table %s existence: %v", table, err)
		}
		if !exists {
			t.Fatalf("expected table %s to exist in schema %s", table, schema)
		}
	}
}

func registerAllModulesForSmoke(t *testing.T, db *testutil.TestDB) *platformmod.Registry {
	t.Helper()

	registry := platformmod.NewRegistry()

	coreMod := core.NewModuleWithDeps(db.Pool, testutil.TestJWTService())
	mustRegister(t, registry, coreMod)

	hrMod := hr.NewModule(db.Pool, coreMod.AuthService())
	mustRegister(t, registry, hrMod)

	subjectMod := subject.NewModule(db.Pool, coreMod.AuthService())
	mustRegister(t, registry, subjectMod)

	roomMod := room.NewModule(db.Pool, coreMod.AuthService())
	mustRegister(t, registry, roomMod)

	timetableMod := timetable.NewModuleWithRepos(
		db.Pool,
		coreMod.AuthService(),
		hrMod.TeacherRepo(),
		hrMod.AvailabilityRepo(),
		subjectMod.SubjectRepo(),
		roomMod.RoomRepo(),
		roomMod.RoomAvailabilityRepo(),
	)
	mustRegister(t, registry, timetableMod)

	toolRegistry := agentinfra.NewToolRegistry()
	providerSvc := agentsvc.NewProviderServiceWithLLM(agentdomain.LLMConfig{}, &testutil.MockLLMProvider{Response: "smoke"})
	agentMod := agent.NewModuleWithProvider(db.Pool, coreMod.AuthService(), toolRegistry, providerSvc, testutil.TestRedis(t))
	mustRegister(t, registry, agentMod)

	return registry
}

func mustRegister(t *testing.T, registry *platformmod.Registry, m pkgmod.Module) {
	t.Helper()
	if err := registry.Register(m); err != nil {
		t.Fatalf("register module %s: %v", m.Name(), err)
	}
}

func assertBefore(t *testing.T, index map[string]int, a, b string) {
	t.Helper()
	ia, oka := index[a]
	ib, okb := index[b]
	if !oka || !okb {
		t.Fatalf("missing module index for %s or %s", a, b)
	}
	if ia >= ib {
		t.Fatalf("expected %s before %s, got indices %d >= %d", a, b, ia, ib)
	}
}
