# MCS-ERP

Multi-tenant, agentic-first ERP system. Academic MVP targeting course scheduling with AI agent assistance.

## Tech Stack

- **Backend:** Go 1.22+, stdlib net/http (REST), gRPC (internal), sqlc, Watermill
- **Frontend:** React 19, TanStack (Router/Query/Table/Form), shadcn/ui, Tailwind, Turborepo+pnpm
- **Database:** PostgreSQL 16 (schema-per-tenant), Redis 7
- **AI:** langchaingo, multi-provider (Claude/OpenAI/Ollama)

## Quick Start

```bash
# Start infrastructure
docker compose up -d

# Backend (hot-reload)
cp .env.example .env
make dev

# Frontend
cd web && pnpm install && pnpm dev
```

## Project Structure

```
cmd/server/          # Go entry point
internal/            # DDD modules (core, hr, subject, room, timetable, agent)
pkg/                 # Shared packages (module interface, erptypes)
migrations/          # SQL migrations
proto/               # Protobuf definitions
sqlc/                # sqlc config + queries
web/                 # Frontend monorepo (Turborepo)
  apps/shell/        # Main React app
  packages/ui/       # Shared UI components
  packages/api-client/ # API types + TanStack Query hooks
  packages/module-*/ # Module frontends
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make dev` | Start Go server with hot-reload (air) |
| `make build` | Build Go binary |
| `make test` | Run Go tests |
| `make lint` | Run golangci-lint |
| `make sqlc` | Generate sqlc code |
| `make proto` | Generate protobuf/gRPC code |
| `make swagger` | Generate OpenAPI docs |
