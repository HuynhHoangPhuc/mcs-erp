# Phase 03: Auth & RBAC

## Context Links
- [Parent Plan](./plan.md)
- Depends on: [Phase 02 - Core Platform](./phase-02-core-platform.md)
- [Brainstorm Report](../reports/brainstorm-260219-1151-mcs-erp-architecture.md)

## Overview
- **Date:** 2026-02-19
- **Priority:** P1
- **Status:** Complete
- **Description:** JWT authentication (access + refresh tokens), user CRUD, role-based access control with permissions, login/register endpoints, auth middleware, frontend login page + auth context + protected routes.

## Key Insights
- JWT access token (15min) + refresh token (7d) stored in httpOnly cookie
- RBAC: roles have permissions; permissions are strings like `hr:teacher:read`
- Auth middleware extracts tenant + user from JWT claims; sets both in context
- Core module owns User, Role, Permission entities — other modules reference via ID
- Frontend uses TanStack Router's `beforeLoad` for route guards

## Requirements

### Functional
- User registration (admin-only; no self-signup for MVP)
- Login endpoint returns JWT access + refresh tokens
- Refresh token rotation (issue new pair, invalidate old)
- User CRUD: create, read, update, deactivate (soft delete)
- Role CRUD: create, read, update, delete
- Assign roles to users; roles contain permission sets
- Auth middleware validates JWT, extracts claims, enforces permissions per route
- Frontend: login page, auth context provider, protected route wrapper

### Non-Functional
- Passwords hashed with bcrypt (cost 12)
- JWT signed with HS256 (symmetric; swap to RS256 later if needed)
- Token refresh is atomic (no race conditions)
- Rate limiting on login endpoint (10 req/min per IP)

## Architecture

```
internal/core/
├── domain/
│   ├── user.go              # User entity
│   ├── role.go              # Role entity
│   ├── permission.go        # Permission value object
│   └── repository.go        # UserRepository, RoleRepository interfaces
├── application/
│   ├── commands/
│   │   ├── create_user.go
│   │   ├── update_user.go
│   │   └── assign_role.go
│   ├── queries/
│   │   ├── get_user.go
│   │   ├── list_users.go
│   │   └── get_user_permissions.go
│   └── services/
│       └── auth_service.go  # Login, refresh, validate
├── infrastructure/
│   ├── postgres_user_repo.go
│   ├── postgres_role_repo.go
│   └── jwt_service.go       # JWT sign/verify
└── delivery/
    ├── auth_handler.go       # POST /api/auth/login, /api/auth/refresh
    ├── user_handler.go       # CRUD /api/users
    ├── role_handler.go       # CRUD /api/roles
    └── auth_middleware.go    # JWT validation + permission check

internal/platform/auth/      # Shared auth utilities
├── claims.go                # JWT claims struct
├── context.go               # UserFromContext, PermissionsFromContext
└── require_permission.go    # Middleware: RequirePermission("hr:teacher:read")
```

### Frontend
```
web/apps/shell/src/
├── routes/
│   ├── __root.tsx           # Root layout (modify: add auth check)
│   ├── _authenticated.tsx   # Authenticated layout wrapper
│   ├── _authenticated/
│   │   └── index.tsx        # Dashboard (placeholder)
│   └── login.tsx            # Login page
├── providers/
│   └── auth-provider.tsx    # Auth context: user, login, logout, isAuthenticated
├── hooks/
│   └── use-auth.ts          # useAuth() hook
└── lib/
    └── api-client.ts        # Axios/fetch with JWT interceptor
```

## Related Code Files

### Files to Create
- `internal/core/domain/user.go` — User entity: ID, Email, PasswordHash, Name, IsActive, TenantID
- `internal/core/domain/role.go` — Role entity: ID, Name, Permissions []string, TenantID
- `internal/core/domain/permission.go` — Permission constants + `HasPermission(perms, required)` helper
- `internal/core/domain/repository.go` — UserRepository, RoleRepository interfaces
- `internal/core/application/commands/create_user.go` — CreateUserCmd + handler
- `internal/core/application/commands/update_user.go` — UpdateUserCmd + handler
- `internal/core/application/commands/assign_role.go` — AssignRoleCmd + handler
- `internal/core/application/queries/get_user.go` — GetUserQuery + handler
- `internal/core/application/queries/list_users.go` — ListUsersQuery + handler
- `internal/core/application/queries/get_user_permissions.go` — returns merged permissions from all roles
- `internal/core/application/services/auth_service.go` — Login(email, pass), Refresh(token), ValidateToken(token)
- `internal/core/infrastructure/postgres_user_repo.go` — sqlc-backed UserRepository
- `internal/core/infrastructure/postgres_role_repo.go` — sqlc-backed RoleRepository
- `internal/core/infrastructure/jwt_service.go` — Sign, Verify, token expiry config
- `internal/core/delivery/auth_handler.go` — POST /api/auth/login, POST /api/auth/refresh, POST /api/auth/logout
- `internal/core/delivery/user_handler.go` — GET/POST/PUT/DELETE /api/users
- `internal/core/delivery/role_handler.go` — GET/POST/PUT/DELETE /api/roles
- `internal/core/delivery/auth_middleware.go` — extracts JWT, sets user+tenant in context
- `internal/platform/auth/claims.go` — `Claims{UserID, TenantID, Email, Permissions}`
- `internal/platform/auth/context.go` — `UserFromContext(ctx)`, `WithUser(ctx, claims)`
- `internal/platform/auth/require_permission.go` — middleware factory
- `internal/core/module.go` — Core module implementing `pkg/module.Module` interface
- `sqlc/queries/core/users.sql` — SQL queries for user CRUD
- `sqlc/queries/core/roles.sql` — SQL queries for role CRUD
- `migrations/core/000001_create_users_table.up.sql`
- `migrations/core/000002_create_roles_table.up.sql`
- `migrations/core/000003_create_user_roles_table.up.sql`
- `migrations/public/000002_create_users_lookup_table.up.sql` — `public.users_lookup(email, tenant_schema)` for cross-tenant login resolution
<!-- Updated: Validation Session 1 - Add users_lookup table for single-tenant login -->
- `web/apps/shell/src/routes/login.tsx`
- `web/apps/shell/src/routes/_authenticated.tsx`
- `web/apps/shell/src/providers/auth-provider.tsx`
- `web/apps/shell/src/hooks/use-auth.ts`
- `web/apps/shell/src/lib/api-client.ts`

## Implementation Steps

1. **Domain entities** — Define User, Role, Permission in `core/domain/`
   - User: `ID uuid`, `Email string`, `PasswordHash string`, `Name string`, `IsActive bool`, `CreatedAt`, `UpdatedAt`
   - Role: `ID uuid`, `Name string`, `Permissions []string`, `Description string`
   - Permission: string constants like `core:user:read`, `core:user:write`, `hr:teacher:read`

2. **Repository interfaces** — `UserRepository` (FindByID, FindByEmail, Save, Update, List), `RoleRepository` (FindByID, Save, Delete, List, FindByUserID)

3. **SQL migrations** — users, roles, user_roles junction table
   ```sql
   -- users
   CREATE TABLE users (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       email VARCHAR(255) NOT NULL UNIQUE,
       password_hash VARCHAR(255) NOT NULL,
       name VARCHAR(255) NOT NULL,
       is_active BOOLEAN NOT NULL DEFAULT true,
       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
       updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
   );
   -- roles
   CREATE TABLE roles (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       name VARCHAR(100) NOT NULL UNIQUE,
       permissions TEXT[] NOT NULL DEFAULT '{}',
       description TEXT,
       created_at TIMESTAMPTZ NOT NULL DEFAULT now()
   );
   -- user_roles
   CREATE TABLE user_roles (
       user_id UUID REFERENCES users(id) ON DELETE CASCADE,
       role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
       PRIMARY KEY (user_id, role_id)
   );
   ```

4. **sqlc queries** — write SQL in `sqlc/queries/core/`, generate Go code

5. **JWT service** — sign access token (15min, claims: user_id, tenant_id, email), sign refresh token (7d, claims: user_id, tenant_id), verify and parse

6. **Auth service** — `Login(email, pass)`: lookup user in `public.users_lookup` table (email → tenant_schema), switch to tenant schema, verify bcrypt, generate tokens with tenant_id claim. `Refresh(token)`: verify refresh, issue new pair. `ValidateToken(token)`: parse claims
<!-- Updated: Validation Session 1 - Single-tenant user login via public.users_lookup -->

7. **Command handlers** — CreateUser (hash password, save), UpdateUser, AssignRole

8. **Query handlers** — GetUser, ListUsers (with pagination), GetUserPermissions (merge from all roles)

9. **HTTP handlers** — auth_handler (login, refresh, logout), user_handler (CRUD), role_handler (CRUD)

10. **Auth middleware** — extract `Authorization: Bearer <token>`, validate, set claims in context. Return 401 on invalid/expired token.

11. **Permission middleware** — `RequirePermission("hr:teacher:read")` checks `PermissionsFromContext(ctx)`. Return 403 if missing.

12. **Core module registration** — implement `Module` interface, register routes under `/api/auth/*`, `/api/users/*`, `/api/roles/*`

13. **Seed admin user** — migration or bootstrap step: create default admin role with all permissions, create admin user

14. **Frontend: api-client** — fetch wrapper with JWT in `Authorization` header, auto-refresh on 401

15. **Frontend: auth-provider** — React context: `user`, `login()`, `logout()`, `isAuthenticated`. Store tokens in memory (access) + httpOnly cookie (refresh).

16. **Frontend: login page** — email + password form, call login API, redirect to dashboard

17. **Frontend: authenticated layout** — TanStack Router `beforeLoad` checks auth; redirects to `/login` if not authenticated

## Todo List
- [x] User, Role, Permission domain entities
- [x] Repository interfaces
- [x] SQL migrations (users, roles, user_roles)
- [x] sqlc queries for user + role CRUD
- [x] JWT service (sign, verify, refresh)
- [x] Auth service (login, refresh, validate)
- [x] Command handlers (create user, update user, assign role)
- [x] Query handlers (get user, list users, get permissions)
- [x] Auth HTTP handler (login, refresh, logout)
- [x] User + Role CRUD HTTP handlers
- [x] Auth middleware (JWT extraction + validation)
- [x] Permission middleware (RequirePermission)
- [x] Core module registration
- [x] Seed admin user + role
- [x] Frontend api-client with JWT interceptor
- [x] Frontend auth provider + useAuth hook
- [x] Frontend login page
- [x] Frontend authenticated route guard
- [x] Unit tests: JWT service, auth service, permission check

## Success Criteria
- POST `/api/auth/login` with valid credentials returns JWT tokens
- Protected endpoints return 401 without token, 403 without permission
- User CRUD works for admin role
- Frontend login flow works end-to-end
- Token refresh works transparently

## Risk Assessment
- **JWT secret rotation**: Not in MVP; document as future improvement
- **Refresh token theft**: httpOnly cookie mitigates; add token family tracking later
- **bcrypt timing attacks**: bcrypt is constant-time by design; safe

## Security Considerations
- Passwords: bcrypt cost 12, never stored or logged in plaintext
- JWT: short-lived access (15min), refresh rotation invalidates old tokens
- CORS: restrict to frontend origin only
- Rate limit login to prevent brute force
- Permission strings are server-authoritative; frontend hides UI but server enforces

## Next Steps
- Phase 4: HR Module (teacher CRUD, departments, availability)
- Phase 5: Subject Module (can start in parallel with Phase 4)
- Phase 6: Room Module (can start in parallel with Phase 4)
