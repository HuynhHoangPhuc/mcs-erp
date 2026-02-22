---
phase: 6
title: "Security & Performance Tests"
status: completed
priority: P2
effort: 3h
depends_on: [1, 5]
---

# Phase 6: Security & Performance Tests

## Context Links
- [platform/tenant/resolver.go](/internal/platform/tenant/resolver.go) -- tenant resolution (injection surface)
- [platform/database/postgres.go](/internal/platform/database/postgres.go) -- SetTenantSchema validation
- [core/delivery/auth_middleware.go](/internal/core/delivery/auth_middleware.go) -- auth bypass surface
- [agent/delivery/conversation_handler.go](/internal/agent/delivery/conversation_handler.go) -- IDOR target
- [timetable/scheduler/](/internal/timetable/scheduler/) -- performance-critical code

## Overview
Security tests for IDOR, SQL injection, RBAC bypass, cross-tenant leakage. Performance benchmarks for scheduling engine and API response times.

## Key Insights
- Schema name validation: `^[a-zA-Z_][a-zA-Z0-9_]*$` regex in `database.SetTenantSchema`
- Tenant resolver: dashes to underscores conversion, regex validation
- IDOR risk: conversation endpoints use path params (`/conversations/{id}`)
- Scheduling perf target: 200 subjects, 80 teachers, 50 rooms < 30s
- API perf target: CRUD operations < 200ms
- Concurrent tenant operations must not leak data

## Requirements

### Security Tests
- IDOR: user A accessing user B's conversation returns 403/404
- SQL injection on search/filter query params
- Schema name injection via X-Tenant-ID header
- RBAC bypass: accessing endpoints without required permissions
- Cross-tenant data leakage under concurrent access
- JWT with wrong signing method rejected
- JWT with claims from tenant A used against tenant B's data

### Performance Tests
- Benchmark: schedule generation 200 subjects, 80 teachers, 50 rooms
- Benchmark: CRUD API response times (p50, p95, p99)
- Benchmark: concurrent multi-tenant CRUD operations
- Go benchmark tests using `testing.B`

## Architecture
```
internal/security_test/
  idor_test.go            -- IDOR vulnerability tests
  injection_test.go       -- SQL injection + schema injection
  rbac_bypass_test.go     -- permission bypass attempts

internal/timetable/scheduler/
  benchmark_test.go       -- scheduling performance benchmarks

internal/platform/
  concurrent_tenant_test.go -- concurrent multi-tenant operations
```

## Related Code Files

### Files to Create
- `internal/security_test/idor_test.go` (~100 lines)
- `internal/security_test/injection_test.go` (~100 lines)
- `internal/security_test/rbac_bypass_test.go` (~80 lines)
- `internal/timetable/scheduler/benchmark_test.go` (~120 lines)
- `internal/platform/concurrent_tenant_test.go` (~100 lines)

## Implementation Steps

### Step 1: IDOR Tests — `internal/security_test/idor_test.go`

```go
package security_test

func TestIDOR_ConversationAccess(t *testing.T) {
    // 1. User A creates conversation (get conv ID)
    // 2. User B (different user, same tenant) tries GET /conversations/{id}
    // 3. Assert 403 or 404 (not 200)
}

func TestIDOR_UserProfile(t *testing.T) {
    // User A tries to update User B's profile
    // Assert rejected
}

func TestIDOR_CrossTenant_ConversationAccess(t *testing.T) {
    // 1. Tenant A user creates conversation
    // 2. Tenant B user (with valid token for tenant B) tries accessing it
    // 3. Assert not found (different schema)
}
```

### Step 2: Injection Tests — `internal/security_test/injection_test.go`

```go
func TestSQLInjection_SearchParams(t *testing.T) {
    // Table-driven with payloads:
    payloads := []string{
        "'; DROP TABLE users; --",
        "1 OR 1=1",
        "\" OR \"\"=\"",
        "1; SELECT * FROM information_schema.tables",
        "admin'--",
    }
    // For each: GET /api/v1/teachers?search={payload}
    // Assert: no 500 (query parameterized), returns 200/400
}

func TestSchemaInjection_XTenantID(t *testing.T) {
    // Table-driven with malicious schema names:
    payloads := []string{
        "'; DROP SCHEMA public CASCADE; --",
        "test; DROP TABLE users",
        "../../../etc/passwd",
        "",
        "a" + strings.Repeat("a", 1000), // very long name
    }
    // For each: send X-Tenant-ID header with payload
    // Assert: 400 (rejected by regex), NOT 500
}

func TestSchemaInjection_SetTenantSchema(t *testing.T) {
    // Direct unit test of database.SetTenantSchema
    // Verify regex rejects special characters
}
```

### Step 3: RBAC Bypass Tests — `internal/security_test/rbac_bypass_test.go`

```go
func TestRBACBypass_AllProtectedEndpoints(t *testing.T) {
    // Table-driven: for each protected endpoint, send request with
    // token that has NO permissions
    endpoints := []struct{
        method, path string
    }{
        {"POST", "/api/v1/users"},
        {"POST", "/api/v1/teachers"},
        {"POST", "/api/v1/subjects"},
        {"POST", "/api/v1/rooms"},
        {"POST", "/api/v1/timetable/semesters"},
        {"POST", "/api/v1/timetable/semesters/xxx/generate"},
    }
    // Assert all return 403
}

func TestRBACBypass_ReadOnlyUser(t *testing.T) {
    // User with only read permissions cannot write
    // Token with "hr:teacher:read" -> POST /api/v1/teachers -> 403
}
```

### Step 4: Scheduling Benchmarks — `internal/timetable/scheduler/benchmark_test.go`

```go
func BenchmarkScheduleGeneration_Small(b *testing.B) {
    // 10 subjects, 5 teachers, 5 rooms
    problem := buildTestProblem(10, 5, 5)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        Solve(problem, DefaultSAConfig())
    }
}

func BenchmarkScheduleGeneration_Medium(b *testing.B) {
    // 50 subjects, 20 teachers, 15 rooms
    problem := buildTestProblem(50, 20, 15)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        Solve(problem, DefaultSAConfig())
    }
}

func TestScheduleGeneration_LargeScale_Under30s(t *testing.T) {
    // 200 subjects, 80 teachers, 50 rooms
    problem := buildTestProblem(200, 80, 50)
    start := time.Now()
    _, err := Solve(problem, DefaultSAConfig())
    elapsed := time.Since(start)
    require.NoError(t, err)
    assert.Less(t, elapsed, 30*time.Second, "large-scale scheduling must complete under 30s")
}

// buildTestProblem generates a realistic Problem with random but valid data.
func buildTestProblem(nSubjects, nTeachers, nRooms int) Problem {
    // Generate subjects with 2-4 hours/week
    // Generate teachers with 60-80% availability
    // Generate rooms with 90% availability
    // Random teacher pre-assignments for 30% of subjects
}
```

### Step 5: Concurrent Tenant Tests — `internal/platform/concurrent_tenant_test.go`

```go
func TestConcurrentTenantOperations_NoDataLeakage(t *testing.T) {
    // 1. Create 5 tenant schemas
    // 2. Seed different data in each
    // 3. Launch 5 goroutines, each:
    //    a. Login as tenant user
    //    b. List teachers (should only see own tenant's data)
    //    c. Create a teacher
    //    d. List again, verify count
    // 4. Use sync.WaitGroup, assert no cross-tenant leakage
    // 5. Verify each tenant's final count matches expected
}

func TestConcurrentTenantOperations_APIResponseTime(t *testing.T) {
    // 1. Seed data
    // 2. Fire 50 concurrent GET requests across 5 tenants
    // 3. Measure p50, p95, p99 response times
    // 4. Assert p95 < 200ms
}
```

## Todo List
- [ ] Create idor_test.go (conversation, user, cross-tenant)
- [ ] Create injection_test.go (SQL injection params, schema injection)
- [ ] Create rbac_bypass_test.go (all protected endpoints)
- [ ] Create scheduler benchmark_test.go (small, medium, large scale)
- [ ] Create concurrent_tenant_test.go (data leakage, response times)
- [ ] Verify security tests: `go test ./internal/security_test/... -v`
- [ ] Run benchmarks: `go test ./internal/timetable/scheduler/ -bench=. -benchmem`
- [ ] Verify concurrent tests: `go test ./internal/platform/ -run Concurrent -race`

## Success Criteria
- IDOR tests confirm user A cannot access user B's conversations
- SQL injection payloads do not cause 500 errors or data exposure
- Schema injection via X-Tenant-ID rejected with 400
- All protected endpoints return 403 for unprivileged tokens
- 200-subject scheduling completes under 30 seconds
- API p95 response time under 200ms with 50 concurrent requests
- No cross-tenant data leakage under concurrent access

## Risk Assessment
| Risk | Impact | Mitigation |
|------|--------|------------|
| SA performance varies by machine | Medium | Use generous 30s target, test on CI baseline |
| False positive injection tests | Low | Verify parameterized queries, not just status codes |
| Concurrent test flakiness | Medium | Use deterministic seed data, adequate timeouts |

## Security Considerations
- Tests themselves validate security properties
- Ensure test payloads cover OWASP top 10 injection patterns
- Schema regex must reject ALL special characters (not just common ones)
- Verify JWT signing method enforcement (HS256 only)

## Next Steps
- After all phases pass, add `make test-integration` target to Makefile
- Consider CI pipeline with testcontainers (GitHub Actions with Docker)
- Future: add load testing with `k6` or `vegeta` for production-like scenarios
