---
phase: 2
title: "Smoke Tests & Build Validation"
status: completed
priority: P1
effort: 3h
depends_on: [1]
---

# Phase 2: Smoke Tests & Build Validation

## Context Links
- [cmd/server/main.go](/cmd/server/main.go) -- server entry point
- [Makefile](/Makefile) -- build/test targets
- [docker-compose.yml](/docker-compose.yml) -- infrastructure
- [Phase 1](phase-01-test-infrastructure-setup.md) -- test helpers

## Overview
Validate the project compiles, Docker infrastructure starts, migrations run, health check responds, and the full module registry bootstraps without error.

## Key Insights
- Health check is at `GET /healthz` (public, no tenant needed)
- Module bootstrap runs `Migrate()` + `RegisterRoutes()` + `RegisterEvents()` in topo order
- `go build ./...` must pass (catches import cycles, type errors)
- Frontend build is separate: `cd web && pnpm build`

## Requirements

### Functional
- `go build ./...` compiles without errors
- `make lint` passes
- Postgres testcontainer + migration runner succeeds
- Health check returns `{"status":"ok"}` with 200
- All 6 modules register and bootstrap without panic
- Module dependency resolution (Kahn's algorithm) produces valid order

### Non-functional
- Smoke test suite completes < 30s
- Zero flaky tests

## Architecture
Single test file with `TestMain` for container lifecycle.

## Related Code Files

### Files to Create
- `internal/platform/smoke_test.go` -- build validation + health check + module bootstrap

### Files to Reference
- `internal/platform/module/registry.go` -- ResolveOrder
- `internal/platform/module/bootstrap.go` -- Bootstrap function
- `internal/platform/database/migrator.go` -- schema migration

## Implementation Steps

### Step 1: Create `internal/platform/smoke_test.go` (~120 lines)

```go
package platform_test

// TestMain sets up Postgres testcontainer shared across all tests in this package.
func TestMain(m *testing.M)

// TestBuild_Compiles validates the project compiles.
// (This test existing and running IS the validation.)
func TestBuild_Compiles(t *testing.T)

// TestHealthCheck_Returns200 starts TestServer and hits /healthz.
func TestHealthCheck_Returns200(t *testing.T) {
    // 1. Create TestDB
    // 2. Create TestServer
    // 3. GET /healthz
    // 4. Assert status 200, body contains "ok"
}

// TestModuleRegistry_ResolvesOrder validates Kahn's algorithm produces correct order.
func TestModuleRegistry_ResolvesOrder(t *testing.T) {
    // 1. Register all 6 modules
    // 2. Call ResolveOrder()
    // 3. Assert "core" comes before hr, subject, room
    // 4. Assert "hr", "subject", "room" come before "timetable"
}

// TestModuleBootstrap_AllModulesStart validates Bootstrap succeeds.
func TestModuleBootstrap_AllModulesStart(t *testing.T) {
    // 1. Create TestDB with schema
    // 2. Register all modules
    // 3. Call Bootstrap(ctx, registry, mux)
    // 4. Assert no error
}

// TestMigrations_ApplyCleanly validates all migration files apply without error.
func TestMigrations_ApplyCleanly(t *testing.T) {
    // 1. Create fresh schema
    // 2. Run all migrations
    // 3. Query pg_tables to verify expected tables exist
}
```

### Step 2: Create `internal/platform/smoke_frontend_test.go` (~40 lines, build-tag gated)

```go
//go:build frontend

package platform_test

// TestFrontend_Builds validates pnpm build succeeds.
func TestFrontend_Builds(t *testing.T) {
    // Skip if pnpm not available
    // exec.Command("pnpm", "--dir", "web", "build")
}
```

## Todo List
- [ ] Create smoke_test.go with TestMain
- [ ] Implement TestHealthCheck_Returns200
- [ ] Implement TestModuleRegistry_ResolvesOrder
- [ ] Implement TestModuleBootstrap_AllModulesStart
- [ ] Implement TestMigrations_ApplyCleanly
- [ ] Create smoke_frontend_test.go (build-tag gated)
- [ ] Verify `go test ./internal/platform/ -v -run Smoke` passes

## Success Criteria
- All smoke tests pass in < 30s
- Module registry correctly orders all 6 modules
- Health check returns 200 with `{"status":"ok"}`
- Migrations create expected tables (users, roles, teachers, subjects, rooms, etc.)

## Risk Assessment
| Risk | Impact | Mitigation |
|------|--------|------------|
| Container slow first pull | Low | CI caches Docker images |
| Frontend pnpm not in CI | Low | Build-tag gate `//go:build frontend` |

## Security Considerations
- None specific to smoke tests

## Next Steps
- Phase 3: Auth & tenant tests build on TestServer pattern established here
