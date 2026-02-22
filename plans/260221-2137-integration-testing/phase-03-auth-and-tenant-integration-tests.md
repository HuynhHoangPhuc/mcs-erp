---
phase: 3
title: "Auth & Tenant Integration Tests"
status: completed
priority: P1
effort: 5h
depends_on: [1]
---

# Phase 3: Auth & Tenant Integration Tests

## Context Links
- [core/application/services/auth_service.go](/internal/core/application/services/auth_service.go) -- Login, Refresh, ValidateToken
- [core/infrastructure/jwt_service.go](/internal/core/infrastructure/jwt_service.go) -- JWT generation/validation
- [core/delivery/auth_middleware.go](/internal/core/delivery/auth_middleware.go) -- Bearer token extraction
- [platform/auth/require_permission.go](/internal/platform/auth/require_permission.go) -- RBAC middleware
- [platform/tenant/resolver.go](/internal/platform/tenant/resolver.go) -- subdomain + X-Tenant-ID
- [platform/tenant/middleware.go](/internal/platform/tenant/middleware.go) -- tenant context injection
- [platform/database/postgres.go](/internal/platform/database/postgres.go) -- SetTenantSchema, WithTenantTx

## Overview
Test the complete auth lifecycle (login -> JWT -> API calls -> refresh -> logout) and tenant isolation (schema switching, cross-tenant prevention).

## Key Insights
- Login resolves tenant via `public.users_lookup(email -> schema)`, then authenticates in tenant schema
- JWT claims carry `UserID`, `TenantID`, `Email`, `Permissions`
- AuthMiddleware extracts Bearer token, validates, sets both user + tenant context
- RequirePermission checks `claims.Permissions` against required permission string
- Tenant middleware skips public paths: `/healthz`, `/api/v1/auth/login`, `/api/v1/auth/register`
- `database.WithTenantTx` sets `SET LOCAL search_path = $schema, public` per transaction

## Requirements

### Functional
- Login with valid credentials returns access + refresh tokens
- Login with wrong password returns 401
- Login with non-existent email returns 401
- Login with deactivated user returns error
- Authenticated request with valid token succeeds
- Request without token returns 401
- Request with expired token returns 401
- Request with tampered token returns 401
- Token refresh returns new token pair
- RBAC: request without required permission returns 403
- Tenant A cannot read Tenant B's data via API
- `X-Tenant-ID` header resolves tenant correctly
- Schema name validation rejects injection attempts

### Non-functional
- Auth test suite < 20s
- All tests safe for `t.Parallel()`

## Architecture
Two test files in `internal/core/`:
- `auth_integration_test.go` -- auth flow tests
- `tenant_isolation_test.go` -- multi-tenant isolation tests

## Related Code Files

### Files to Create
- `internal/core/auth_integration_test.go` (~180 lines)
- `internal/core/tenant_isolation_test.go` (~150 lines)

### Files to Reference
- `internal/core/domain/user.go` -- User struct
- `internal/core/domain/role.go` -- Role struct, permissions
- `internal/core/domain/permission.go` -- permission constants
- `internal/core/delivery/auth_handler.go` -- login/refresh/logout endpoints
- `internal/core/delivery/user_handler.go` -- user CRUD endpoints

## Implementation Steps

### Step 1: Create `internal/core/auth_integration_test.go`

```go
package core_test

var testDB *testutil.TestDB

func TestMain(m *testing.M) {
    // Start Postgres container, run migrations, seed, cleanup
}

// --- Login Flow ---

func TestLogin_ValidCredentials_ReturnsTokenPair(t *testing.T) {
    // 1. SeedAdmin in schema "tenant_a"
    // 2. POST /api/v1/auth/login {"email":"admin@test.com","password":"..."}
    // 3. Assert 200, body has access_token + refresh_token + expires_in
}

func TestLogin_WrongPassword_Returns401(t *testing.T) {
    // POST /api/v1/auth/login with wrong password
    // Assert 401
}

func TestLogin_NonExistentEmail_Returns401(t *testing.T)

func TestLogin_DeactivatedUser_ReturnsError(t *testing.T) {
    // 1. Seed user with is_active=false
    // 2. POST login
    // 3. Assert error response
}

// --- Token Validation ---

func TestAuthenticatedRequest_ValidToken_Succeeds(t *testing.T) {
    // 1. Login, get token
    // 2. GET /api/v1/users with Bearer token + X-Tenant-ID
    // 3. Assert 200
}

func TestAuthenticatedRequest_NoToken_Returns401(t *testing.T) {
    // GET /api/v1/users without Authorization header
}

func TestAuthenticatedRequest_ExpiredToken_Returns401(t *testing.T) {
    // Generate token with -1h expiry using test JWT service
}

func TestAuthenticatedRequest_TamperedToken_Returns401(t *testing.T) {
    // Modify payload bytes of a valid token
}

// --- Token Refresh ---

func TestRefresh_ValidRefreshToken_ReturnsNewPair(t *testing.T) {
    // 1. Login, get refresh_token
    // 2. POST /api/v1/auth/refresh {"refresh_token":"..."}
    // 3. Assert new access_token != old
}

func TestRefresh_InvalidToken_Returns401(t *testing.T)

// --- RBAC ---

func TestRBAC_WithPermission_Returns200(t *testing.T) {
    // Token with "core:user:read" -> GET /api/v1/users -> 200
}

func TestRBAC_WithoutPermission_Returns403(t *testing.T) {
    // Token with no permissions -> GET /api/v1/users -> 403
}

func TestRBAC_AdminCanCreateUser(t *testing.T) {
    // Token with "core:user:write" -> POST /api/v1/users -> 201
}
```

### Step 2: Create `internal/core/tenant_isolation_test.go`

```go
package core_test

func TestTenantIsolation_SeparateSchemas(t *testing.T) {
    // 1. Create schema_a and schema_b
    // 2. Seed user in schema_a
    // 3. Seed different user in schema_b
    // 4. Login as schema_a user -> list users -> only sees schema_a users
    // 5. Login as schema_b user -> list users -> only sees schema_b users
}

func TestTenantIsolation_CrossTenantRead_Fails(t *testing.T) {
    // 1. Create user in schema_a
    // 2. Get token for schema_a user but send X-Tenant-ID: schema_b
    // 3. AuthMiddleware should set tenant from JWT claims, NOT from header
    // 4. Verify user cannot access schema_b data
}

func TestTenantResolver_XTenantIDHeader(t *testing.T) {
    // 1. Request with X-Tenant-ID: test_tenant
    // 2. Verify tenant.FromContext returns "test_tenant"
}

func TestTenantResolver_SubdomainExtraction(t *testing.T) {
    // 1. Request with Host: faculty-a.mcs-erp.com
    // 2. Verify resolved schema is "faculty_a"
}

func TestTenantResolver_MissingTenant_Returns400(t *testing.T) {
    // Request without subdomain or X-Tenant-ID header
    // Verify 400 response
}

func TestSchemaNameValidation_RejectsInjection(t *testing.T) {
    // Table-driven test with malicious inputs:
    // "'; DROP TABLE users; --", "../../etc/passwd", ""
    // All must return error from SetTenantSchema
}

func TestWithTenantTx_SetsSearchPath(t *testing.T) {
    // 1. Create schema "test_xyz"
    // 2. WithTenantTx -> query SHOW search_path
    // 3. Assert contains "test_xyz"
}
```

## Todo List
- [ ] Create auth_integration_test.go with TestMain
- [ ] Implement login flow tests (valid, wrong password, non-existent, deactivated)
- [ ] Implement token validation tests (no token, expired, tampered)
- [ ] Implement refresh token tests
- [ ] Implement RBAC tests (with/without permission)
- [ ] Create tenant_isolation_test.go
- [ ] Implement schema isolation tests (two tenants, separate data)
- [ ] Implement cross-tenant read prevention test
- [ ] Implement tenant resolver tests (header, subdomain, missing)
- [ ] Implement schema injection prevention tests
- [ ] Verify all pass: `go test ./internal/core/ -v -race`

## Success Criteria
- Login returns valid JWT tokens with correct claims
- Expired/tampered tokens are rejected with 401
- Users without required permissions get 403
- Tenant A cannot access Tenant B's data
- Schema name injection attempts are blocked
- All tests pass with `-race`

## Risk Assessment
| Risk | Impact | Mitigation |
|------|--------|------------|
| JWT claims tenant vs header tenant conflict | High | Test that AuthMiddleware uses JWT tenant, not header |
| bcrypt hashing slow in tests | Low | Use cost=4 for test passwords |
| Schema cleanup between tests | Medium | Use unique schema names per test |

## Security Considerations
- Validate that JWT tampering is caught (modify base64 payload)
- Validate schema name regex prevents SQL injection
- Validate cross-tenant data leakage is impossible

## Next Steps
- Phase 4 uses auth helpers established here for module CRUD tests
- Phase 6 expands on security tests (IDOR, injection)
