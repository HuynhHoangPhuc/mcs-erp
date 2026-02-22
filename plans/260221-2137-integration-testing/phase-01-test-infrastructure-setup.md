---
phase: 1
title: "Test Infrastructure Setup"
status: pending
priority: P1
effort: 6h
---

# Phase 1: Test Infrastructure Setup

## Context Links
- [database/postgres.go](/internal/platform/database/postgres.go) -- pool + tenant schema helpers
- [database/migrator.go](/internal/platform/database/migrator.go) -- migration runner
- [config/config.go](/internal/platform/config/config.go) -- env config loader
- [docker-compose.yml](/docker-compose.yml) -- Postgres 16 + Redis 7

## Overview
Set up shared test infrastructure: testcontainers-go for PostgreSQL/Redis, helper functions for DB setup/teardown, fixture seeding, HTTP test server creation, and JWT token generation.

## Key Insights
<!-- Updated: Validation Session 1 - Docker Compose Postgres instead of testcontainers -->
- Every repo query requires `SET LOCAL search_path = $schema` inside a transaction
- `database.WithTenantTx()` is the gateway for all tenant-scoped DB ops
- Modules take `*pgxpool.Pool` + `*services.AuthService` as constructor args
- Tenant resolution in tests: use `X-Tenant-ID` header (no subdomain needed)
- Migrations live in `migrations/` directory, run via goose
- **Test DB:** Connect to Docker Compose Postgres (no testcontainers-go). Requires `docker compose up -d` before tests.
- **Build tags:** All DB-dependent test files use `//go:build integration` header
- **Migrations:** Use goose programmatically (`goose.Up()`) in test setup
- **Mock LLM:** Create mock LLM provider for agent chat tests

## Requirements

### Functional
- Postgres testcontainer starts before test suite, stops after
- Redis testcontainer for agent module tests
- Each test gets isolated tenant schema (uuid-based name)
- Migration runner applies all `migrations/*.up.sql` to test schemas
- Helper to create authenticated `*http.Request` with JWT + tenant header
- Fixture builders for: User, Role, Teacher, Subject, Room, Semester

### Non-functional
- Container startup < 10s
- Test schema creation + migration < 2s per schema
- All helpers must be safe for `t.Parallel()`

## Architecture

```
internal/testutil/
  testutil.go           -- TestDB struct, container lifecycle, schema helpers
  fixtures.go           -- seed data builders (User, Teacher, Subject, Room)
  http_helpers.go       -- httptest server creation, authenticated requests
  jwt_helpers.go        -- token generation for test users
```

## Related Code Files

### Files to Create
- `internal/testutil/testutil.go` -- Postgres pool + schema management (Docker Compose)
- `internal/testutil/fixtures.go` -- fixture builder functions
- `internal/testutil/http_helpers.go` -- HTTP test utilities
- `internal/testutil/jwt_helpers.go` -- JWT token helpers
- `internal/testutil/mock_llm_provider.go` -- Mock LLM provider for agent tests

### Files to Modify
- `Makefile` -- add `test-integration` target with `-tags integration` flag

## Implementation Steps

<!-- Updated: Validation Session 1 - Use Docker Compose Postgres + goose -->
### Step 1: No testcontainers dependency needed
Tests connect to Docker Compose Postgres. Ensure `docker compose up -d` runs before tests.
Add test DATABASE_URL env var: `postgres://mcs:mcs_dev_pass@localhost:5432/mcs_erp_test?sslmode=disable`

### Step 2: Create `internal/testutil/testutil.go` (~120 lines)
```go
package testutil

// TestDB wraps a pgxpool.Pool connected to Docker Compose Postgres.
type TestDB struct {
    Pool *pgxpool.Pool
}

// NewTestDB connects to Docker Compose Postgres and returns a pool.
// Reads DATABASE_URL from env, defaults to local dev Postgres.
func NewTestDB(t *testing.T) *TestDB

// CreateTenantSchema creates a unique schema for the test and runs goose migrations.
func (db *TestDB) CreateTenantSchema(t *testing.T) string

// Close closes the pool (no container to terminate).
func (db *TestDB) Close(t *testing.T)
```

Key implementation details:
- Connect to `localhost:5432/mcs_erp_test` (or `DATABASE_URL` env)
- Create test database if not exists
- Create `public.tenants` and `public.users_lookup` tables in setup
- `CreateTenantSchema` generates `test_<uuid>` schema name
- Apply migrations using `goose.Up()` programmatically per schema
- Use `t.Cleanup()` for schema drop + pool close

### Step 3: Create `internal/testutil/fixtures.go` (~120 lines)
```go
// SeedAdmin creates a user with admin role + all permissions in the given schema.
func SeedAdmin(t *testing.T, pool *pgxpool.Pool, schema string) *SeedResult

// SeedTeacher creates a teacher record with availability.
func SeedTeacher(t *testing.T, pool *pgxpool.Pool, schema string, opts ...TeacherOption) *TeacherFixture

// SeedSubject creates a subject record.
func SeedSubject(t *testing.T, pool *pgxpool.Pool, schema string, opts ...SubjectOption) *SubjectFixture

// SeedRoom creates a room record with availability.
func SeedRoom(t *testing.T, pool *pgxpool.Pool, schema string, opts ...RoomOption) *RoomFixture

// SeedResult holds IDs and credentials for a seeded admin.
type SeedResult struct {
    UserID   uuid.UUID
    Email    string
    Password string
    Schema   string
}
```

### Step 4: Create `internal/testutil/jwt_helpers.go` (~60 lines)
```go
// TestJWTService creates a JWTService with a known test secret.
func TestJWTService() *infrastructure.JWTService

// GenerateTestToken creates a valid JWT for test requests.
func GenerateTestToken(t *testing.T, userID uuid.UUID, tenantID string, perms []string) string
```

### Step 5: Create `internal/testutil/http_helpers.go` (~80 lines)
```go
// TestServer creates an httptest.Server wired with all modules against the test DB.
func TestServer(t *testing.T, pool *pgxpool.Pool) *httptest.Server

// AuthenticatedRequest creates an HTTP request with Bearer token + X-Tenant-ID.
func AuthenticatedRequest(method, url, token, tenantID string, body io.Reader) *http.Request

// DoJSON sends a request and decodes the JSON response.
func DoJSON[T any](t *testing.T, client *http.Client, req *http.Request) (T, int)
```

### Step 6: Create `internal/testutil/redis_helpers.go` (~40 lines)
```go
// TestRedis returns the Docker Compose Redis URL for agent tests.
func TestRedis(t *testing.T) string
```

### Step 7: Create `internal/testutil/mock_llm_provider.go` (~60 lines)
```go
// MockLLMProvider implements agent domain.Provider interface.
// Returns canned responses for testing SSE streaming and tool dispatch.
type MockLLMProvider struct {
    Response string
    Tools    []domain.ToolCall
}

func (m *MockLLMProvider) Complete(ctx context.Context, messages []domain.Message) (*domain.Response, error)
func (m *MockLLMProvider) StreamComplete(ctx context.Context, messages []domain.Message, ch chan<- domain.Chunk) error
```

## Todo List
- [ ] Add testcontainers-go dependencies
- [ ] Implement `testutil.go` with Postgres container lifecycle
- [ ] Implement migration runner that reads `migrations/` directory
- [ ] Implement `fixtures.go` with seed functions
- [ ] Implement `jwt_helpers.go`
- [ ] Implement `http_helpers.go` with TestServer
- [ ] Implement `redis_helpers.go`
- [ ] Verify `go test ./internal/testutil/... -v` passes

## Success Criteria
- `TestDB` starts Postgres container, creates schema, runs migrations
- `SeedAdmin` inserts user + role + lookup entry in correct schema
- `TestServer` serves HTTP with all modules registered
- `GenerateTestToken` produces tokens accepted by `AuthMiddleware`
- All helpers work with `-race` flag

## Risk Assessment
| Risk | Impact | Mitigation |
|------|--------|------------|
| testcontainers slow on CI | Medium | Reuse container across test suite via `TestMain` |
| Migration file ordering | Low | Sort by filename prefix (000001, 000002...) |
| Port conflicts on CI | Low | testcontainers uses random ports |

## Security Considerations
- Test JWT secret is hardcoded (`test-secret-do-not-use-in-production`)
- Test DB credentials are ephemeral (container-only)

## Next Steps
- Phases 2-4 depend on this infrastructure
- TestServer creation pattern used by all subsequent phases
