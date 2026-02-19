# Phase 01: Project Scaffolding

## Context Links
- [Parent Plan](./plan.md)
- Dependencies: None (first phase)
- [Brainstorm Report](../reports/brainstorm-260219-1151-mcs-erp-architecture.md)

## Overview
- **Date:** 2026-02-19
- **Priority:** P1
- **Status:** Complete
- **Description:** Initialize Go module, Docker Compose (Postgres+Redis), frontend monorepo (Turborepo+pnpm), CI skeleton, full directory structure.

## Key Insights
- Go 1.22+ enhanced routing eliminates need for chi/gorilla
- Turborepo + pnpm workspace for frontend module isolation
- sqlc codegen against a `_template` schema; runtime `SET LOCAL search_path`
- Docker Compose for local dev; same containers for prod MVP

## Requirements

### Functional
- Go module compiles and runs empty HTTP server on :8080
- Docker Compose starts Postgres 16 + Redis 7
- Frontend dev server starts on :3000 with empty shell
- Makefile provides `dev`, `build`, `test`, `lint`, `migrate`, `sqlc` targets

### Non-Functional
- Hot-reload for Go (air) and frontend (Vite HMR)
- All config via environment variables (.env file)
- CI runs lint + test on PR

## Architecture

```
mcs-erp/
├── cmd/
│   └── server/
│       └── main.go                    # Entry point
├── internal/
│   ├── platform/                      # Shared infra (created empty)
│   ├── core/                          # Core module (created empty)
│   ├── hr/                            # HR module (created empty)
│   ├── subject/                       # Subject module (created empty)
│   ├── room/                          # Room module (created empty)
│   ├── timetable/                     # Timetable module (created empty)
│   └── agent/                         # AI Agent module (created empty)
├── pkg/
│   ├── module/                        # Module interface definitions
│   └── erptypes/                      # Shared types (IDs, errors)
├── migrations/                        # SQL migrations (per-module dirs)
├── proto/                             # Protobuf definitions (per-module)
│   └── buf.yaml                       # buf config
├── sqlc/                              # sqlc config + queries
│   └── sqlc.yaml
├── web/                               # Frontend monorepo
│   ├── apps/
│   │   └── shell/                     # Main host app (Vite + React)
│   ├── packages/
│   │   ├── ui/                        # Shared UI (shadcn/ui + Tailwind)
│   │   ├── api-client/                # API types + TanStack Query hooks
│   │   ├── module-hr/                 # HR frontend (empty)
│   │   ├── module-subject/            # Subject frontend (empty)
│   │   ├── module-room/               # Room frontend (empty)
│   │   ├── module-timetable/          # Timetable frontend (empty)
│   │   └── module-agent/              # AI chat frontend (empty)
│   ├── pnpm-workspace.yaml
│   └── turbo.json
├── docker-compose.yml
├── Makefile
├── .env.example
├── .gitignore
├── .air.toml                          # Go hot-reload config
└── .github/
    └── workflows/
        └── ci.yml
```

## Related Code Files

### Files to Create
- `cmd/server/main.go` — minimal HTTP server, healthcheck endpoint
- `go.mod`, `go.sum` — Go module init
- `Makefile` — dev, build, test, lint, migrate, sqlc targets
- `docker-compose.yml` — Postgres 16, Redis 7, app service
- `.env.example` — DATABASE_URL, REDIS_URL, PORT, JWT_SECRET
- `.air.toml` — Go hot-reload config
- `.gitignore` — Go + Node + IDE + .env
- `.github/workflows/ci.yml` — lint + test
- `sqlc/sqlc.yaml` — sqlc config pointing to template schema
- `pkg/module/module.go` — Module interface definition
- `pkg/erptypes/id.go` — UUID-based ID types
- `pkg/erptypes/errors.go` — Common error types
- `web/pnpm-workspace.yaml` — pnpm workspace config
- `web/turbo.json` — Turborepo pipeline config
- `web/apps/shell/package.json` — Shell app with React 19, TanStack Router, Vite
- `web/apps/shell/vite.config.ts` — Vite config
- `web/apps/shell/tsconfig.json`
- `web/apps/shell/src/main.tsx` — React entry
- `web/apps/shell/src/router.ts` — TanStack Router setup (empty route tree)
- `web/apps/shell/src/routes/__root.tsx` — Root layout component
- `web/apps/shell/index.html`
- `web/packages/ui/package.json` — Shared UI lib
- `web/packages/ui/tailwind.config.ts` — Tailwind config
- `web/packages/api-client/package.json` — API client package

## Implementation Steps

1. **Init Go module**
   ```bash
   go mod init github.com/phuc/mcs-erp
   ```
2. **Create directory structure** — all `internal/` module dirs with empty `.gitkeep`
3. **Write `cmd/server/main.go`** — stdlib net/http server, `/healthz` endpoint, graceful shutdown via `signal.NotifyContext`
4. **Write `pkg/module/module.go`** — `Module` interface with `Name()`, `Dependencies()`, `RegisterRoutes()`, `RegisterEvents()`, `Migrate()`, `RegisterAITools()`
5. **Write `pkg/erptypes/id.go`** — `type ID = uuid.UUID`, `NewID()`, `ParseID()`
6. **Write `pkg/erptypes/errors.go`** — `ErrNotFound`, `ErrConflict`, `ErrForbidden`, `ErrValidation`
7. **Write `docker-compose.yml`** — Postgres 16 (port 5432, volume), Redis 7 (port 6379), healthchecks
8. **Write `.env.example`** — all config vars with defaults
9. **Write `Makefile`** — targets: `dev` (air), `build`, `test` (go test ./...), `lint` (golangci-lint), `migrate`, `sqlc` (sqlc generate), `swagger` (swaggo/swag generate)
<!-- Updated: Validation Session 1 - Add OpenAPI/Swagger tooling -->
10. **Write `.air.toml`** — watch `cmd/`, `internal/`, `pkg/`; build `cmd/server`
11. **Write `sqlc/sqlc.yaml`** — version 2, engine postgres, queries + schema paths
11b. **Setup protobuf tooling** — install `buf`, create `proto/buf.yaml` + `proto/buf.gen.yaml` (Go + gRPC codegen), add `proto` and `swagger` Makefile targets
<!-- Updated: Validation Session 1 - gRPC for internal, REST for external -->
12. **Init frontend monorepo**
    ```bash
    cd web && pnpm init && pnpm add -D turbo
    ```
13. **Write `web/pnpm-workspace.yaml`** — packages: `apps/*`, `packages/*`
14. **Write `web/turbo.json`** — build/dev/lint/test pipeline
15. **Scaffold shell app** — `web/apps/shell/` with Vite + React 19 + TanStack Router
16. **Scaffold UI package** — `web/packages/ui/` with Tailwind + shadcn/ui init
17. **Scaffold api-client package** — `web/packages/api-client/` with TanStack Query
18. **Create empty module frontend packages** — `module-hr`, `module-subject`, `module-room`, `module-timetable`, `module-agent` with `package.json` + `src/index.ts`
19. **Write `.github/workflows/ci.yml`** — Go lint+test, pnpm build
20. **Write `.gitignore`** — Go binaries, node_modules, .env, dist/, tmp/
21. **Verify**: `docker compose up -d`, `make dev`, `cd web && pnpm dev` all start without errors

## Todo List
- [x] Go module init + directory structure
- [x] `cmd/server/main.go` with healthcheck
- [x] `pkg/module/module.go` interface
- [x] `pkg/erptypes/` ID and error types
- [x] `docker-compose.yml` (Postgres + Redis)
- [x] `.env.example` + `.air.toml`
- [x] `Makefile` with all targets
- [x] `sqlc/sqlc.yaml` config
- [x] Frontend monorepo init (pnpm + turbo)
- [x] Shell app scaffolding (Vite + React + TanStack Router)
- [x] UI package with Tailwind + shadcn/ui
- [x] api-client package
- [x] Empty module frontend packages
- [x] GitHub Actions CI
- [x] `.gitignore`
- [x] Smoke test: all services start

## Success Criteria
- `docker compose up -d` starts Postgres + Redis healthy
- `make dev` starts Go server, `curl localhost:8080/healthz` returns 200
- `cd web && pnpm dev` starts Vite dev server on :3000
- `make lint` and `make test` pass (no errors on empty project)
- CI pipeline runs successfully on push
- `make swagger` generates OpenAPI spec

## Risk Assessment
- **pnpm/Turborepo version conflicts**: Pin exact versions in package.json
- **sqlc schema mismatch**: Use `_template` schema for codegen, test early

## Security Considerations
- `.env` excluded from git via `.gitignore`
- Docker Compose uses non-root Postgres user
- No secrets in CI config; use GitHub secrets for future deployments

## Next Steps
- Phase 2: Core Platform (config, database pool, tenant resolver, migrations, module registry, event bus)
