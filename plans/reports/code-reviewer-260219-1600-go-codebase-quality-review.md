# Code Review Report: MCS-ERP Go Backend

**Date:** 2026-02-19
**Scope:** Full Go backend -- 6 modules, ~100 Go source files
**Focus:** Security, architecture, error handling, patterns, module wiring

---

## Overall Assessment

Solid foundation for an MVP modular monolith. DDD layering is consistent, tenant isolation via schema-per-tenant is well-implemented, and the module registry with topological sort is clean. However, there are several security and robustness issues that need attention before production deployment.

---

## Critical Issues

### C1. gRPC Tenant Interceptor: No Schema Name Validation

**File:** `/Users/phuc/Developer/mcs-erp/internal/platform/grpc/tenant_interceptor.go` (line 26)

The gRPC interceptor accepts `x-tenant-id` from metadata and passes it directly to `tenant.WithTenant()` without validating the schema name. The HTTP middleware runs through `Resolve()` which validates against a regex, but gRPC bypasses this entirely. A malicious internal caller could inject arbitrary schema names (e.g., `public`, `pg_catalog`, `information_schema`) to access system tables.

```go
// CURRENT (vulnerable):
ctx = tenant.WithTenant(ctx, tenantIDs[0])

// FIX: validate before setting
schema := tenantIDs[0]
if !validSchemaRegex.MatchString(schema) {
    return nil, status.Error(codes.InvalidArgument, "invalid tenant id format")
}
ctx = tenant.WithTenant(ctx, schema)
```

### C2. Tenant Middleware Error Response: JSON Injection

**File:** `/Users/phuc/Developer/mcs-erp/internal/platform/tenant/middleware.go` (line 24)

The error message includes `err.Error()` via string concatenation inside a raw JSON string. If the error message contains quotes or backslashes, the JSON output is malformed or could be exploited for response splitting.

```go
// CURRENT (injectable):
http.Error(w, `{"error":"tenant resolution failed: `+err.Error()+`"}`, http.StatusBadRequest)

// FIX: use proper JSON marshaling
writeJSON(w, http.StatusBadRequest, map[string]string{"error": "tenant resolution failed: " + err.Error()})
```

### C3. Password Hash Exposed in List/FindByID Queries

**File:** `/Users/phuc/Developer/mcs-erp/internal/core/infrastructure/postgres_user_repo.go` (lines 39-42, 57-60, 113-114, 124)

All user queries (`FindByID`, `FindByEmail`, `List`) SELECT `password_hash` from the database. While the API handlers strip it from JSON responses, the hash is loaded into memory unnecessarily for read-only operations, and the `List` endpoint scans it into every user struct. If any serialization path changes, hashes could leak.

```go
// FIX for List: use SELECT without password_hash
"SELECT id, email, name, is_active, created_at, updated_at FROM users ORDER BY ..."
```

### C4. No Request Body Size Limit

**Files:** All handlers using `json.NewDecoder(r.Body).Decode()`

No handler limits the request body size. An attacker can send a multi-GB body to exhaust server memory. The `http.MaxBytesReader` wrapper is missing.

```go
// FIX: add at handler entry or middleware level
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
```

---

## High Priority

### H1. Schedule Generation is Synchronous and Blocking

**File:** `/Users/phuc/Developer/mcs-erp/internal/timetable/delivery/schedule_handler.go` (line 87)

`GenerateSchedule` runs `ParallelAnneal` synchronously in the HTTP handler. With `DefaultSAConfig` (500,000 iterations * NumCPU workers), this could take minutes. The `WriteTimeout` is 15 seconds, so the connection will be killed for any non-trivial problem size. The handler also holds a DB transaction via `semesterRepo.Update()` during this entire computation.

**Fix:** Run scheduling asynchronously; return 202 Accepted with a job ID; poll or SSE for status.

### H2. Conversation Access Control Missing (IDOR)

**File:** `/Users/phuc/Developer/mcs-erp/internal/agent/delivery/conversation_handler.go` (lines 53-65, 104-128, 132-145)

`GetConversation`, `UpdateConversation`, and `DeleteConversation` look up by conversation ID without verifying the conversation belongs to the authenticated user. Any user with `agent:chat:use` permission can read/modify/delete any other user's conversations by guessing UUIDs.

```go
// FIX: verify ownership after lookup
conv, err := h.repo.FindConversationByID(r.Context(), id)
if conv.UserID != claims.UserID {
    writeJSON(w, http.StatusForbidden, ...)
    return
}
```

### H3. LLM Provider Constructed Per Request

**File:** `/Users/phuc/Developer/mcs-erp/internal/agent/application/services/provider_service.go` (line 22)

`GetLLM()` calls `infrastructure.NewLLM()` on every chat request, creating a new HTTP client and potentially re-authenticating each time. LLM clients should be created once at startup and reused.

```go
// FIX: cache the LLM instance with sync.Once or similar
type ProviderService struct {
    llm  infrastructure.LLMModel
    once sync.Once
    err  error
}
```

### H4. `os.Exit(1)` Called Inside Goroutines

**File:** `/Users/phuc/Developer/mcs-erp/cmd/server/main.go` (lines 173, 186)

`os.Exit(1)` is called inside goroutines (gRPC listener failure, HTTP server failure). This skips all deferred cleanup functions (pool.Close(), graceful shutdown). The gRPC case is particularly bad because a listen failure during startup will skip the pool cleanup.

```go
// FIX: use a shared errCh or cancel the parent context
errCh := make(chan error, 2)
go func() { errCh <- grpcSrv.Serve(lis) }()
```

### H5. Refresh Token Not Distinguished from Access Token

**File:** `/Users/phuc/Developer/mcs-erp/internal/core/infrastructure/jwt_service.go`

Both access and refresh tokens use the same signing key and same claims structure. The `Refresh()` method in `auth_service.go` calls `ValidateToken()` which accepts any valid JWT. This means an access token can be used as a refresh token and vice versa. There should be a `token_type` claim to differentiate them.

```go
// FIX: add a TokenType field to Claims
type Claims struct {
    jwt.RegisteredClaims
    TokenType   string    `json:"token_type"` // "access" or "refresh"
    ...
}
```

---

## Medium Priority

### M1. Event Bus Created But Never Used

**File:** `/Users/phuc/Developer/mcs-erp/cmd/server/main.go` (lines 61-66)

```go
bus, err := eventbus.New()
_ = bus // will be used by modules
```

The event bus is created but not passed to any module. The Watermill router is never started (`bus.Router().Run()` is never called), so even if modules subscribe, no events will be delivered.

### M2. `CreateSchema` Uses `fmt.Sprintf` Without Identifier Quoting

**File:** `/Users/phuc/Developer/mcs-erp/internal/platform/database/postgres.go` (line 51)

While the regex validation blocks most injection, the schema name is not quoted with `pgx.Identifier.Sanitize()` or `pq.QuoteIdentifier()`. The regex allows names like `DROP` or SQL keywords that are syntactically valid identifiers.

```go
// Safer approach:
quoted := pgx.Identifier{schema}.Sanitize()
_, err := pool.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+quoted)
```

Same issue on line 42 for `SetTenantSchema`.

### M3. No Rate Limiting on Auth Endpoints

**File:** `/Users/phuc/Developer/mcs-erp/internal/core/module.go` (lines 55-57)

Login, refresh, and register are public endpoints with no rate limiting. Brute-force password attacks are trivially possible.

### M4. Pagination Offset Can Be Negative

**Files:** `user_handler.go` (line 81), `conversation_handler.go` (line 33)

```go
offset, _ := strconv.Atoi(q.Get("offset"))
```

If `offset` is empty or invalid, `Atoi` returns 0 (fine). But negative values are allowed and passed directly to SQL `OFFSET`, which PostgreSQL handles but could produce unexpected results.

### M5. `CreateRole` Accepts Arbitrary Permission Strings

**File:** `/Users/phuc/Developer/mcs-erp/internal/core/delivery/role_handler.go` (line 44)

The `permissions` field in `CreateRole` is not validated against the defined permission constants in `domain.AllPermissions()`. An admin could create roles with typo'd or nonsensical permission strings that silently fail to grant access.

### M6. `MigrateAll` Applies Raw SQL Without Idempotency Tracking

**File:** `/Users/phuc/Developer/mcs-erp/internal/platform/database/migrator.go` (line 47)

The migrator runs raw SQL on every startup without checking if the migration has already been applied. If modules embed `CREATE TABLE` in their Migrate() method, this will fail or silently skip on second run depending on `IF NOT EXISTS` usage.

### M7. SSE Chat Handler Missing Context Cancellation Cleanup

**File:** `/Users/phuc/Developer/mcs-erp/internal/agent/delivery/chat_handler.go` (lines 72-84)

If the client disconnects mid-stream, `r.Context()` is cancelled. The goroutine running `ProcessMessage` will detect this via `ctx.Done()` in the streaming func, but the `for token := range tokenCh` loop will block until the channel is closed. The goroutine should detect cancellation and stop producing tokens.

---

## Low Priority

### L1. Duplicate Schema Validation Regex

Two regex patterns validate schema names independently:
- `/Users/phuc/Developer/mcs-erp/internal/platform/database/postgres.go` line 12: `schemaNameRegex`
- `/Users/phuc/Developer/mcs-erp/internal/platform/tenant/resolver.go` line 10: `validSchema`

Both are `^[a-zA-Z_][a-zA-Z0-9_]*$`. Should be a single shared utility.

### L2. `sendError` in Agent Service Uses `default` Case

**File:** `/Users/phuc/Developer/mcs-erp/internal/agent/application/services/agent_service.go` (line 290-294)

```go
func sendError(ch chan<- string, msg string) {
    select {
    case ch <- fmt.Sprintf(`{"error":%q}`, msg):
    default: // silently drops error if channel is full
    }
}
```

The `default` case silently drops error messages if the channel buffer is full. This could mask failures from the client.

### L3. Module Interface Compliance Not Verified at Compile Time

The `core.Module`, `hr.Module`, etc. structs implement `pkg/module.Module` but none have `var _ pkgmod.Module = (*Module)(nil)` compile-time checks.

### L4. `UpdateAssignment` Day/Period Validation is Zero-Value Ambiguous

**File:** `/Users/phuc/Developer/mcs-erp/internal/timetable/delivery/schedule_handler.go` (lines 237-242)

```go
if req.Day >= 0 && req.Day <= 5 { existing.Day = req.Day }
if req.Period >= 1 && req.Period <= 10 { existing.Period = req.Period }
```

Since `Day: 0` is Monday and the default int zero-value, sending `{}` (no day field) will update Day to 0. Use `*int` for optional fields.

---

## Positive Observations

1. **Consistent DDD layering** -- Every module follows domain/application/infrastructure/delivery cleanly
2. **Schema-per-tenant isolation** is well-designed with `SetTenantSchema` inside transactions
3. **Module registry with topological sort** -- Kahn's algorithm for dependency resolution is correct and detects cycles
4. **SQL injection prevention** -- All repo queries use parameterized queries ($1, $2 etc.), no string interpolation for user data
5. **Scheduler is pure** -- No database dependencies, clean separation of concerns with Problem/SAConfig
6. **Tool registry RBAC** -- Agent tools are permission-gated before execution
7. **Proper use of `context.WithValue`** for tenant and auth with unexported key types
8. **Graceful shutdown** pattern in main.go is well-structured

---

## Recommended Actions (Priority Order)

1. **Validate gRPC tenant ID** against the schema regex (C1)
2. **Fix JSON injection** in tenant middleware error response (C2)
3. **Stop loading password_hash** in read-only user queries (C3)
4. **Add `http.MaxBytesReader`** to all handlers or as middleware (C4)
5. **Add conversation ownership check** in agent CRUD handlers (H2)
6. **Differentiate access vs refresh tokens** with a `token_type` claim (H5)
7. **Make schedule generation async** or increase WriteTimeout for that route (H1)
8. **Cache LLM client instance** instead of creating per request (H3)
9. **Replace `os.Exit` in goroutines** with error channel pattern (H4)
10. **Add rate limiting** on auth endpoints (M3)

---

## Metrics

- **Files Reviewed:** ~50 Go source files
- **LOC Reviewed:** ~3,500
- **Critical Issues:** 4
- **High Priority:** 5
- **Medium Priority:** 7
- **Low Priority:** 4
- **Type Coverage:** N/A (Go is statically typed, all interface compliance verified except L3)
- **Test Coverage:** 0% (tests not yet written, excluded from scope)
- **Linting Issues:** Not run (excluded from scope)

---

## Unresolved Questions

1. Are there plans to use `public` schema tables beyond `users_lookup` and `tenants`? If so, the `SetTenantSchema` function appending `, public` to search_path could expose cross-tenant data through public schema tables.
2. Is the `_template` schema intended only for sqlc codegen or also as a base for cloning new tenant schemas? The current migrator applies SQL to both `_template` and all active tenants, but the cloning mechanism is not visible.
3. Will gRPC services be internal-only (between modules) or exposed to external clients? The lack of auth interceptor on gRPC (only tenant interceptor) suggests internal-only, but should be confirmed.
