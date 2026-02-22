---
title: "MCS-ERP Integration Testing"
description: "Comprehensive integration test suite for all 6 modules with tenant isolation, auth, cross-module, security, and performance tests"
status: completed
priority: P1
effort: 28h
branch: main
tags: [testing, integration, go, docker-compose, goose]
created: 2026-02-21
---

# MCS-ERP Integration Testing Plan

## Objective
Build a comprehensive integration test suite validating all 6 bounded contexts, tenant isolation, auth flows, cross-module interactions, security, and performance.

## Architecture
- **Test DB**: Docker Compose Postgres (reuse existing `docker-compose.yml`)
- **Test helpers**: `internal/testutil/` shared package
- **Test files**: alongside source in `internal/{module}/*_test.go`
- **Build tags**: `//go:build integration` for DB-dependent tests
- **Runner**: `make test` (unit only), `make test-integration` (all with Docker)
- **Migrations**: goose programmatic (`goose.Up()`) in test setup
- **Assertions**: testify/assert + testify/require

## Phases

| # | Phase | Status | Effort | File |
|---|-------|--------|--------|------|
| 1 | Test Infrastructure Setup | completed | 6h | [phase-01](phase-01-test-infrastructure-setup.md) |
| 2 | Smoke Tests & Build Validation | completed | 3h | [phase-02](phase-02-smoke-tests-and-build-validation.md) |
| 3 | Auth & Tenant Integration Tests | completed | 5h | [phase-03](phase-03-auth-and-tenant-integration-tests.md) |
| 4 | Module Integration Tests | completed | 6h | [phase-04](phase-04-module-integration-tests.md) |
| 5 | Cross-Module & Timetable Tests | completed | 5h | [phase-05](phase-05-cross-module-and-timetable-tests.md) |
| 6 | Security & Performance Tests | completed | 3h | [phase-06](phase-06-security-and-performance-tests.md) |

## Dependencies
- Phase 1 must complete first (all other phases depend on test helpers)
- Phases 2-4 can run in parallel after Phase 1
- Phase 5 depends on Phases 3-4 (needs auth + module helpers)
- Phase 6 depends on Phase 5 (reuses fixtures)

## Key Decisions
- Docker Compose Postgres (reuse existing infra, no testcontainers dep)
- Schema-per-test-tenant for full isolation within single DB
- No mocks for DB layer; real Postgres queries
- `//go:build integration` tag separates DB tests from pure unit tests
- goose programmatic migration in test setup
- Mock LLM provider for agent chat tests (no real API calls)
- Constraint-only validation for scheduler output (no exact assignment checks)
- Security tests in separate `internal/security_test/` package
- `/healthz` endpoint already exists

## Validation Log

### Session 1 — 2026-02-21
**Trigger:** Initial plan creation validation
**Questions asked:** 7

#### Questions & Answers

1. **[Architecture]** Phase 1 plans testcontainers-go for PostgreSQL (spins up real containers in tests). This adds a Docker dependency to `go test`. Alternative is to use the already-running docker-compose Postgres for tests. Which approach?
   - Options: testcontainers-go (Recommended) | Docker Compose Postgres | SQLite for unit tests only
   - **Answer:** Docker Compose Postgres
   - **Rationale:** Reuses existing docker-compose.yml. Faster, no extra dependency. Requires `docker compose up` before tests.

2. **[Scope]** Phase 4 plans to test the Agent SSE chat endpoint, but LLM providers (Claude/OpenAI) won't be available in test environments. How should agent/chat tests handle LLM calls?
   - Options: Mock LLM provider (Recommended) | Skip chat generation tests | Use Ollama in CI
   - **Answer:** Mock LLM provider
   - **Rationale:** Create test LLM provider returning canned responses. Validates SSE format, tool dispatch, message persistence without real API calls.

3. **[Architecture]** Phase 6 places security tests in a separate `internal/security_test/` package. Alternative: colocate security tests with each module. Where should security tests live?
   - Options: Separate security_test/ package (Recommended) | Colocated with modules
   - **Answer:** Separate security_test/ package
   - **Rationale:** Centralized security test suite. Easy to run all security tests at once. Cross-cutting concerns in one place.

4. **[Assumptions]** The plan assumes the app has a `GET /healthz` health check endpoint. Does this endpoint exist, or should Phase 2 create it?
   - Options: Exists already | Needs to be created | Not sure
   - **Answer:** Exists already
   - **Rationale:** No implementation needed for health check. Phase 2 only tests it.

5. **[Architecture]** Phase 1 plans to read migration files from `migrations/` directory and apply them programmatically. The project uses goose for migrations. Should tests use goose directly, or a custom migration runner?
   - Options: Use goose programmatically (Recommended) | Custom SQL file reader | Pre-seeded template schema
   - **Answer:** Use goose programmatically
   - **Rationale:** Call goose.Up() in test setup. Reuses existing migration tooling. Ensures test schema matches production.

6. **[Architecture]** The plan creates 17+ test files across all modules. Should all integration tests require Docker, or should some tests be runnable without Docker?
   - Options: Build-tag separation (Recommended) | All tests require Docker | Separate test directories
   - **Answer:** Build-tag separation
   - **Rationale:** `//go:build integration` for DB-dependent tests. `make test` runs unit only. `make test-integration` runs all. CI runs both.

7. **[Tradeoffs]** Phase 5 tests schedule generation E2E, which involves the simulated annealing algorithm (non-deterministic). How should tests validate scheduler output?
   - Options: Constraint validation only (Recommended) | Seed-controlled deterministic | Statistical validation
   - **Answer:** Constraint validation only
   - **Rationale:** Verify zero hard violations (no teacher/room conflicts). Don't assert exact assignments. Accept SA randomness.

#### Confirmed Decisions
- **Test DB:** Docker Compose Postgres — drop testcontainers-go, reuse existing infra
- **LLM testing:** Mock provider — create `testutil/mock_llm_provider.go`
- **Security tests:** Centralized in `internal/security_test/` package
- **Health check:** Exists — Phase 2 only tests, no creation needed
- **Migrations:** goose programmatic — call `goose.Up()` in test setup
- **Build tags:** `//go:build integration` — separate unit from integration tests
- **Scheduler validation:** Constraint-only — verify no conflicts, not exact assignments

#### Action Items
- [ ] Phase 1: Remove testcontainers-go dependency, use Docker Compose Postgres connection
- [ ] Phase 1: Add goose programmatic migration to test setup
- [ ] Phase 1: Add `//go:build integration` tags to all DB-dependent test files
- [ ] Phase 1: Create `testutil/mock_llm_provider.go` for agent tests
- [ ] Phase 2: Remove /healthz creation, only test existing endpoint
- [ ] Phase 4: Use mock LLM provider in agent chat SSE tests
- [ ] Phase 5: Validate constraints only, not exact assignments
- [ ] Makefile: Add `test-integration` target with build tag

#### Impact on Phases
- Phase 1: Major change — remove testcontainers-go, connect to Docker Compose Postgres instead. Add goose.Up() for migrations. Add mock LLM provider. Add build tag guidance.
- Phase 2: Minor — remove /healthz creation step. Already exists.
- Phase 4: Minor — use mock LLM provider in agent chat tests.
- Phase 5: Minor — scheduler tests validate constraints only, not exact assignments.
- All phases: Add `//go:build integration` header to all test files with DB dependency.

### Session 2 — 2026-02-22
**Trigger:** Completion of timetable cross-module reader coverage plus the comprehensive security regression suite.

**Actions:**
- Added integration tests covering the timetable cross-module reader, semester CRUD/permissions, schedule generation/approval/update flow, double-booking guards, and the empty assignment guard to address timing flakiness identified during code review.
- Added security-focused integration tests for IDOR, tenant-claim/header mismatch leakage, SQL/schema injection handling, JWT `alg=none` rejection, and RBAC bypass denial.
- Executed combined validation (`make test-integration` / `go test ./... -tags integration -race`) plus `go test ./internal/security_test/...` to cover the new suites.

**Outcome:** All Cross-Module and Security integration suites passed under race conditions; timing bounds stabilized; hard-conflict assertions replaced with explicit double-booking checks as required by the review.
