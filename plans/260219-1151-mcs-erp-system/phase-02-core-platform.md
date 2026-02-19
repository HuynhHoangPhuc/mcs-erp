# Phase 02: Core Platform

## Context Links
- [Parent Plan](./plan.md)
- Depends on: [Phase 01 - Project Scaffolding](./phase-01-project-scaffolding.md)
- [Go Backend Patterns Research](./research/researcher-01-go-backend-patterns.md)

## Overview
- **Date:** 2026-02-19
- **Priority:** P1
- **Status:** Complete
- **Description:** Build shared platform layer: config loader, pgxpool database with schema-per-tenant switching, tenant resolver middleware (REST + gRPC), migration runner, module registry with topological sort, Watermill event bus wrapper, gRPC server + interceptors for internal communication.

## Key Insights
- sqlc codegen against `_template` schema; runtime `SET LOCAL search_path` in transactions
- pgxpool `BeforeAcquire`/`AfterRelease` hooks for schema switching (or per-tx `SET LOCAL`)
- Watermill GoChannel for in-process pub/sub; swap to NATS later by changing one line
- Kahn's algorithm for module dependency topo sort (~40 lines)
- `public` schema stores tenants table + global config; tenant schemas store module data

## Requirements

### Functional
- Load config from env vars with sensible defaults
- pgxpool connects to Postgres; schema switching works per-tenant
- Tenant resolver middleware extracts tenant from `X-Tenant-ID` header (or subdomain later)
- Migration runner applies SQL files to all tenant schemas + `_template`
- Module registry accepts modules, resolves dependencies via topo sort, bootstraps in order
- Watermill event bus publishes/subscribes domain events in-process

### Non-Functional
- Connection pool: min 5, max 25 connections
- Config validation at startup (fail fast on missing required vars)
- Migration runner is idempotent (re-run safe)

## Architecture

```
internal/platform/
├── config/
│   └── config.go              # Env-based config struct + loader
├── database/
│   ├── postgres.go            # pgxpool setup, schema switching helpers
│   └── migrator.go            # Per-tenant migration runner
├── tenant/
│   ├── middleware.go           # HTTP middleware: extract tenant, set ctx
│   ├── resolver.go            # Resolve tenant from header/subdomain
│   └── context.go             # TenantFromContext / WithTenant helpers
├── module/
│   ├── registry.go            # Module registry + topo sort
│   └── bootstrap.go           # Bootstrap sequence: migrate, register routes/events
├── eventbus/
│   ├── bus.go                 # Watermill GoChannel wrapper
│   └── publisher.go           # Typed event publisher helper
└── grpc/
    ├── server.go              # gRPC server setup
    ├── tenant_interceptor.go  # Propagate tenant context via gRPC metadata
    └── auth_interceptor.go    # Propagate auth claims via gRPC metadata
```

## Related Code Files

### Files to Create
- `internal/platform/config/config.go` — Config struct: DB URL, Redis URL, Port, JWT secret, log level
- `internal/platform/database/postgres.go` — `NewPool(cfg)`, `WithTenantSchema(ctx, pool, schema)` helper
- `internal/platform/database/migrator.go` — `RunMigrations(pool, migrationsDir)`, iterates tenant schemas
- `internal/platform/tenant/context.go` — `TenantFromContext(ctx)`, `WithTenant(ctx, tenantID)`
- `internal/platform/tenant/resolver.go` — `Resolve(r *http.Request) (string, error)`
- `internal/platform/tenant/middleware.go` — `Middleware(next http.Handler) http.Handler`
- `internal/platform/module/registry.go` — `Registry`, `Register(m Module)`, `ResolveOrder() ([]Module, error)`
- `internal/platform/module/bootstrap.go` — `Bootstrap(registry, pool, router, eventbus)`
- `internal/platform/eventbus/bus.go` — `NewEventBus()`, wraps `gochannel.NewGoChannel`
- `internal/platform/eventbus/publisher.go` — `Publish(ctx, event)` with JSON marshaling
- `migrations/000001_create_tenants_table.up.sql` — `public.tenants` table
- `migrations/000001_create_tenants_table.down.sql`

### Files to Modify
- `cmd/server/main.go` — wire config, pool, registry, event bus, tenant middleware
- `go.mod` — add pgx, watermill, uuid dependencies

## Implementation Steps

1. **Config loader** (`config/config.go`)
   - Define `Config` struct with fields: `DatabaseURL`, `RedisURL`, `Port`, `JWTSecret`, `LogLevel`
   - `Load()` reads from `os.Getenv` with defaults; returns error if `DatabaseURL` empty
   - No external lib needed; stdlib `os.Getenv` + simple validation

2. **Database pool** (`database/postgres.go`)
   - `NewPool(ctx, dbURL) (*pgxpool.Pool, error)` — parse config, set pool min/max
   - `SetTenantSchema(ctx, conn, schema)` — executes `SET LOCAL search_path = $schema, public` (used inside tx)
   - `CreateTenantSchema(ctx, pool, schemaName)` — `CREATE SCHEMA IF NOT EXISTS`
   - Add pgx dependency: `github.com/jackc/pgx/v5`

3. **Migration runner** (`database/migrator.go`)
   - Use `golang-migrate/migrate` with pgx driver
   - `RunMigrations(pool, dir)`:
     a. Apply to `_template` schema first
     b. Query `public.tenants` for all active schemas
     c. For each tenant schema, set search_path and run migrations
   - Migrations dir structure: `migrations/{module}/` — each module has own migration files

4. **Tenant context** (`tenant/context.go`)
   - Context key type (unexported): `type ctxKey struct{}`
   - `WithTenant(ctx, tenantID string) context.Context`
   - `TenantFromContext(ctx) (string, error)` — returns error if missing

5. **Tenant resolver** (`tenant/resolver.go`)
   - `Resolve(r *http.Request) (string, error)`
   - Primary: extract from subdomain (`faculty-a.mcs-erp.com` → `faculty_a` schema)
   - Fallback: check `X-Tenant-ID` header (for dev/testing/API clients)
   - Return error if neither found
<!-- Updated: Validation Session 1 - Subdomain-based tenant resolution as primary -->

6. **Tenant middleware** (`tenant/middleware.go`)
   - `Middleware(next) http.Handler` — calls resolver, sets context, calls next
   - Returns 400 if tenant cannot be resolved
   - Skips tenant resolution for public routes (healthz, login)

7. **Module registry** (`module/registry.go`)
   - `Registry` struct with `modules map[string]Module`
   - `Register(m Module)` — stores by name, checks duplicate
   - `ResolveOrder() ([]Module, error)` — Kahn's algorithm topo sort on `Dependencies()`
   - Returns error on circular dependency (missing dep = error too)

8. **Module bootstrap** (`module/bootstrap.go`)
   - `Bootstrap(ctx, reg, pool, mux, bus)` — calls `ResolveOrder()`, then for each module:
     a. `module.Migrate(pool)`
     b. `module.RegisterRoutes(mux)`
     c. `module.RegisterEvents(bus)`
   - Log each module init

9. **Event bus** (`eventbus/bus.go`)
   - `NewEventBus() *EventBus` — wraps `gochannel.NewGoChannel(Config{BufferSize: 256})`
   - `EventBus` embeds Watermill publisher + subscriber
   - `Router() *message.Router` — creates Watermill router for handler registration

10. **Event publisher** (`eventbus/publisher.go`)
    - `Publish(ctx, topic string, event any) error` — JSON marshal event, create `message.Message`, publish
    - `Subscribe(topic, handler)` — convenience wrapper

11. **Tenants migration** — create `public.tenants` table:
    ```sql
    CREATE TABLE IF NOT EXISTS public.tenants (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        name VARCHAR(255) NOT NULL,
        schema_name VARCHAR(63) NOT NULL UNIQUE,
        is_active BOOLEAN NOT NULL DEFAULT true,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );
    ```

12. **gRPC server** (`grpc/server.go`)
    - `NewGRPCServer(cfg) *grpc.Server` — create server with interceptors
    - Tenant interceptor: extract tenant from gRPC metadata `x-tenant-id`, set in context
    - Auth interceptor: extract user claims from gRPC metadata, set in context
    - Server listens on separate port (e.g., :9090) for internal traffic
    - In modular monolith, gRPC calls are in-process (direct function calls behind gRPC interface for future microservice split)
<!-- Updated: Validation Session 1 - gRPC for internal module communication -->

13. **Proto definitions** — `proto/` directory with per-module `.proto` files
    - Each module defines its reader service (e.g., `hr.v1.HRReaderService`)
    - Use `buf` for codegen: `buf generate` → Go gRPC stubs
    - Proto dir structure: `proto/{module}/v1/{module}.proto`

14. **Wire in main.go** — load config, create pool, create event bus, create gRPC server, create registry, bootstrap, start HTTP server (REST :8080) + gRPC server (:9090) with tenant middleware

## Todo List
- [x] Config loader with env vars
- [x] pgxpool setup with schema switching
- [x] Migration runner (per-tenant iteration)
- [x] Tenant context helpers (WithTenant, TenantFromContext)
- [x] Tenant resolver (header + subdomain)
- [x] Tenant middleware
- [x] Module registry with topo sort
- [x] Module bootstrap sequence
- [x] Watermill event bus wrapper
- [x] Event publisher helper
- [x] Tenants table migration (public schema)
- [x] gRPC server setup with tenant + auth interceptors
- [x] Proto directory structure + buf config
- [x] Wire everything in main.go (REST :8080 + gRPC :9090)
- [x] Unit tests for topo sort + tenant resolver

## Success Criteria
- Config loads from env; fails fast on missing DATABASE_URL
- `pgxpool` connects; `SET LOCAL search_path` switches schemas in tx
- Migration runner creates `_template` schema + applies migrations
- Tenant middleware extracts tenant from header, sets context
- Module registry resolves order correctly; detects circular deps
- Event bus publishes/subscribes in-process (unit test)

## Risk Assessment
- **Pool connection leak on schema switch**: Use `SET LOCAL` inside tx only, never bare `SET` on pooled conn
- **Migration ordering across modules**: Each module prefixes migrations with module name to avoid collisions
- **Watermill GoChannel event loss on panic**: Acceptable for MVP; add recovery middleware

## Security Considerations
- Tenant isolation enforced at middleware level — every request must have valid tenant
- Schema names sanitized (alphanumeric + underscore only) to prevent SQL injection
- Database credentials never logged

## Next Steps
- Phase 3: Auth & RBAC (JWT tokens, user management, auth middleware)
