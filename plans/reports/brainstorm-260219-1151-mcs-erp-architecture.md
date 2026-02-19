# Brainstorm Report: MCS-ERP System Architecture

**Date:** 2026-02-19
**Status:** Agreed

---

## Problem Statement

Build a **multi-tenant, agentic-first ERP system** inspired by Odoo's modular architecture, starting with an academic MVP (HR for teachers, Subject management with prerequisites, AI-powered Timetable scheduling). System must be general-purpose and extensible for multiple business domains via pluggable modules.

## Requirements Summary

| Requirement | Decision |
|-------------|----------|
| Team | Solo developer |
| Backend | Go + DDD + Clean Architecture + CQRS |
| Frontend | TanStack (Router, Query, Table, Form) + React |
| Database | PostgreSQL + Redis |
| Multi-tenancy | Schema-per-tenant |
| Auth | Built-in JWT + RBAC |
| AI Agent | Multi-provider (Claude, OpenAI, etc.) — chat + inline actions |
| Event Bus | In-process (Watermill) — evolve to NATS/Kafka later |
| Module System | Compile-time registry with SDK for technical partners |
| Deployment | Docker Compose (MVP) → Kubernetes (scale) |
| Scheduling | Greedy + Simulated Annealing in Go |
| Timeline | 3-4 months for MVP |

---

## Evaluated Approaches

### Approach 1: Full Distributed Microservices (Rejected)
- Kafka + K8s + micro-frontends + separate services per module
- **Pros:** Maximum scalability, independent deployments
- **Cons:** 5-10x ops overhead, solo dev bottleneck, months spent on infra not features
- **Verdict:** Over-engineered for MVP. Premature optimization.

### Approach 2: Modular Monolith with Clean Boundaries (Selected)
- Single Go binary with DDD bounded contexts as internal modules
- In-process event bus (Watermill) with interface contracts between modules
- Monorepo frontend with lazy-loaded module routes
- **Pros:** Ship fast, clean architecture, easy to split later, one deploy target
- **Cons:** Must discipline module boundaries manually (no process isolation)
- **Verdict:** Best balance of architecture quality and shipping speed.

### Approach 3: Serverless/FaaS (Not Considered)
- Would fragment the codebase and make DDD difficult
- Cold starts hurt ERP responsiveness
- **Verdict:** Wrong fit for ERP workloads.

---

## Final Recommended Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────┐
│                    FRONTEND (React + TanStack)           │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────┐  │
│  │ Core UI  │ │ HR Module│ │Timetable │ │ Subject   │  │
│  │ (Shell)  │ │  Routes  │ │  Routes  │ │  Routes   │  │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └─────┬─────┘  │
│       └─────────────┴────────────┴─────────────┘        │
│                         │ REST/SSE                       │
└─────────────────────────┼───────────────────────────────┘
                          │
┌─────────────────────────┼───────────────────────────────┐
│              GO MODULAR MONOLITH                         │
│                         │                                │
│  ┌──────────────────────┴──────────────────────────┐    │
│  │              API Gateway Layer                    │    │
│  │   (HTTP Router + Middleware + Auth + Tenant)      │    │
│  └──────┬────────────┬────────────┬────────────┬────┘    │
│         │            │            │            │         │
│  ┌──────┴──┐  ┌──────┴──┐  ┌─────┴───┐  ┌────┴─────┐  │
│  │  Core   │  │   HR    │  │Timetable│  │ Subject  │  │
│  │ Module  │  │ Module  │  │ Module  │  │ Module   │  │
│  │─────────│  │─────────│  │─────────│  │──────────│  │
│  │• Auth   │  │• Teacher│  │• Schedule│  │• Subject │  │
│  │• Tenant │  │• Dept   │  │• Slots  │  │• Prereq  │  │
│  │• User   │  │• Avail  │  │• Solver │  │• Category│  │
│  │• RBAC   │  │• Skills │  │• AI Asst│  │• Credits │  │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬─────┘  │
│       │            │            │            │         │
│  ┌────┴────────────┴────────────┴────────────┴────┐    │
│  │           Event Bus (Watermill in-process)      │    │
│  └────────────────────┬───────────────────────────┘    │
│                       │                                 │
│  ┌────────────────────┴───────────────────────────┐    │
│  │              AI Agent Layer                      │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │    │
│  │  │ Chat     │  │ Tool     │  │ Multi-LLM    │  │    │
│  │  │ Handler  │  │ Registry │  │ Provider     │  │    │
│  │  └──────────┘  └──────────┘  └──────────────┘  │    │
│  └─────────────────────────────────────────────────┘    │
│                                                         │
│  ┌─────────────────────────────────────────────────┐    │
│  │           Infrastructure Layer                   │    │
│  │  PostgreSQL │ Redis │ Module Registry │ Migrations│   │
│  └─────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

### Backend Architecture (Go)

#### Directory Structure (DDD + Clean Architecture)
```
mcs-erp/
├── cmd/
│   └── server/main.go              # Entry point
├── internal/
│   ├── core/                        # Core bounded context
│   │   ├── domain/                  # Entities, value objects, interfaces
│   │   ├── application/             # Use cases (commands + queries)
│   │   ├── infrastructure/          # Repos, external services
│   │   └── delivery/                # HTTP handlers
│   ├── hr/                          # HR bounded context
│   │   ├── domain/
│   │   ├── application/
│   │   ├── infrastructure/
│   │   └── delivery/
│   ├── timetable/                   # Timetable bounded context
│   │   ├── domain/
│   │   ├── application/
│   │   ├── infrastructure/
│   │   └── delivery/
│   ├── subject/                     # Subject bounded context
│   │   ├── domain/
│   │   ├── application/
│   │   ├── infrastructure/
│   │   └── delivery/
│   ├── agent/                       # AI Agent bounded context
│   │   ├── domain/
│   │   ├── application/
│   │   ├── infrastructure/
│   │   │   ├── llm/                 # Multi-provider LLM clients
│   │   │   └── tools/               # Tool definitions for LLM
│   │   └── delivery/
│   └── platform/                    # Shared infrastructure
│       ├── auth/                    # JWT + RBAC
│       ├── tenant/                  # Schema-per-tenant resolver
│       ├── eventbus/                # Watermill wrapper
│       ├── database/                # Connection pool, migrations
│       ├── module/                  # Module registry & lifecycle
│       └── config/                  # Configuration management
├── pkg/                             # Public SDK for module developers
│   ├── module/                      # Module interface definitions
│   ├── events/                      # Event type definitions
│   └── erptypes/                    # Shared domain types
├── migrations/                      # SQL migrations (per-module)
├── web/                             # Frontend monorepo
└── docker-compose.yml
```

#### CQRS Pattern (Per Module)
```
application/
├── commands/
│   ├── create-teacher-command.go
│   ├── update-teacher-command.go
│   └── command-bus.go
├── queries/
│   ├── get-teacher-query.go
│   ├── list-teachers-query.go
│   └── query-bus.go
└── events/
    ├── teacher-created-event.go
    └── teacher-updated-event.go
```

#### Module Registry Pattern
```go
// pkg/module/module.go
type Module interface {
    Name() string
    Dependencies() []string
    RegisterRoutes(router Router)
    RegisterEvents(bus EventBus)
    Migrate(db Database) error
    RegisterAITools(registry ToolRegistry)
}
```
- Modules register at compile-time via `init()` or explicit registration in `main.go`
- Topological sort resolves dependency order (inspired by Odoo's `modules.graph`)
- Each module declares its routes, events, migrations, and AI tools

#### Multi-Tenancy (Schema-per-Tenant)
```
PostgreSQL
├── public/                  # Shared: tenants table, global config
├── tenant_abc/              # Tenant ABC: all module tables
├── tenant_xyz/              # Tenant XYZ: all module tables
└── ...
```
- Middleware extracts tenant from subdomain/header → sets `search_path`
- Migration runner iterates all tenant schemas on module install/upgrade
- Redis stores tenant config cache + session data

#### AI Agent Architecture
```
User Message → Chat Handler → Intent Classifier
                                    │
                    ┌───────────────┼───────────────┐
                    ▼               ▼               ▼
              Tool Calling    Conversation      Inline Suggest
              (schedule,      (Q&A, explain)    (auto-fill,
               assign, etc.)                     recommend)
                    │
                    ▼
              Tool Registry → Module Tools
              (each module registers its AI-callable functions)
                    │
                    ▼
              Multi-LLM Provider
              ├── Claude (primary)
              ├── OpenAI (fallback)
              └── Ollama (local/dev)
```

**Key design:** Each module registers AI tools (like MCP/function-calling tools). The agent layer doesn't know module internals — it just dispatches tool calls. This keeps AI completely decoupled from business logic.

### Frontend Architecture

#### Monorepo Structure
```
web/
├── packages/
│   ├── shell/                    # App shell, layout, navigation, auth
│   │   ├── src/
│   │   │   ├── routes/           # TanStack Router root
│   │   │   ├── components/       # Shared layout components
│   │   │   └── providers/        # Auth, theme, tenant context
│   │   └── package.json
│   ├── ui/                       # Shared UI component library
│   │   ├── src/components/       # Buttons, tables, forms, modals
│   │   └── package.json
│   ├── module-hr/                # HR module frontend
│   │   ├── src/
│   │   │   ├── routes/           # Lazy-loaded HR routes
│   │   │   ├── components/       # HR-specific components
│   │   │   └── api/              # TanStack Query hooks for HR
│   │   └── package.json
│   ├── module-timetable/         # Timetable module frontend
│   ├── module-subject/           # Subject module frontend
│   └── module-agent/             # AI chat + inline action components
├── pnpm-workspace.yaml
└── turbo.json
```

#### Key Libraries
| Purpose | Library |
|---------|---------|
| Routing | TanStack Router (file-based, type-safe) |
| Data fetching | TanStack Query (cache, mutations, optimistic updates) |
| Tables | TanStack Table (sorting, filtering, pagination) |
| Forms | TanStack Form + Zod validation |
| State | Zustand (lightweight, per-module stores) |
| UI Kit | shadcn/ui + Tailwind CSS |
| Build | Vite + Turborepo |
| AI Chat | Custom chat component with SSE streaming |

### Timetable Scheduling Algorithm

#### Strategy: Greedy Initialization + Simulated Annealing

**Hard constraints (must satisfy):**
- No teacher teaches 2 subjects at same time
- No room double-booked
- Teacher must be available at assigned slot
- Subject prerequisites ordering (semester-level, not time-slot level)

**Soft constraints (optimize):**
- Minimize gaps in teacher schedules
- Prefer teacher's preferred time slots
- Distribute subjects evenly across the week
- Group related subjects near each other

**Algorithm flow:**
1. Build constraint graph from subjects + prerequisites + teacher availability
2. Greedy assign: place subjects with most constraints first
3. Simulated annealing: iteratively swap/move slots, accept worse solutions with decreasing probability
4. AI review: LLM analyzes the schedule for human-readable issues
5. Admin reviews + modifies (manually or via AI chat)

### Subject Prerequisite Graph

Model prerequisites as a **DAG (Directed Acyclic Graph)**:
```
domain/
├── subject.go           # Subject entity
├── prerequisite.go      # PrerequisiteEdge value object
└── prerequisite-graph.go # DAG operations (topological sort, cycle detection)
```
- On subject creation/update, validate no cycles via DFS
- Topological sort provides valid teaching order per semester
- API endpoint returns the full dependency graph for frontend visualization

---

## Key Decisions Summary

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Architecture | Modular monolith | Ship fast, split later. Clean DDD boundaries. |
| Event bus | Watermill (in-process) | Zero ops overhead. Swap to NATS/Kafka via config change. |
| Module system | Compile-time registry | Type-safe, no runtime plugin fragility. SDK for partners. |
| Multi-tenancy | Schema-per-tenant | Strong isolation for business clients. |
| Scheduling | Greedy + SA in Go | No external solver dependency. Handles 500+ entities. |
| AI integration | Tool-calling pattern | Modules register tools. AI layer is LLM-agnostic. |
| Frontend | Monorepo + lazy routes | Module isolation without micro-frontend complexity. |
| Deployment | Docker Compose → K8s | Simple start, scale when needed. |

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Scope creep (too many features) | HIGH | HIGH | Strict MVP: HR + Subject + Timetable only. No email, no reporting. |
| Schema-per-tenant migration complexity | MEDIUM | MEDIUM | Automated migration runner. Test with 3+ tenants from start. |
| AI agent unpredictable behavior | MEDIUM | HIGH | Strict tool definitions, validation layer, human-in-the-loop for destructive actions. |
| Solo dev burnout | HIGH | HIGH | Ship incrementally. Core → HR → Subject → Timetable → AI. |
| Scheduling algorithm edge cases | MEDIUM | MEDIUM | Fallback: if SA fails, present partial schedule + let admin fix. |

---

## Implementation Order (3-4 Month MVP)

### Month 1: Foundation
- [ ] Project scaffolding (Go module, monorepo, Docker Compose)
- [ ] Core module: auth (JWT), tenant management, RBAC, user management
- [ ] Database: Postgres + Redis setup, schema-per-tenant migrations
- [ ] Module registry + event bus (Watermill)
- [ ] Frontend shell: TanStack Router, auth flow, layout

### Month 2: Domain Modules
- [ ] HR module: teacher CRUD, department, availability management
- [ ] Subject module: CRUD, prerequisite DAG, cycle detection, categories
- [ ] Frontend: HR pages, Subject pages with dependency graph visualization

### Month 3: Timetable + AI
- [ ] Room module: room CRUD, capacity, equipment tags
- [ ] Timetable module: scheduling algorithm (greedy + SA), room assignment
- [ ] Semester management: open subjects, auto-assign teachers + rooms
- [ ] AI agent: multi-LLM provider, tool registry, chat API
- [ ] Frontend: timetable UI (grid/calendar view), AI chat sidebar

### Month 4: Polish + Deploy
- [ ] AI inline actions (suggestions, auto-fill)
- [ ] Admin review workflow (approve/reject/modify timetable)
- [ ] Docker Compose production config
- [ ] Documentation + module SDK docs for technical partners

---

## Tech Stack Summary

| Layer | Technology |
|-------|-----------|
| Language | Go 1.22+ |
| HTTP Framework | net/http (stdlib, Go 1.22+ enhanced routing) |
| SQL | sqlc (type-safe SQL codegen from raw SQL) |
| Event Bus | Watermill |
| DI | Manual (constructor injection, no framework) |
| AI/LLM | langchaingo + custom multi-provider |
| Frontend | React 19 + TypeScript |
| Routing | TanStack Router |
| Data Layer | TanStack Query |
| UI Kit | shadcn/ui + Tailwind CSS |
| Build | Vite + Turborepo + pnpm |
| Database | PostgreSQL 16 + Redis 7 |
| Deploy | Docker Compose (MVP) |
| CI/CD | GitHub Actions |

---

## Resolved Decisions
1. **HTTP framework**: net/http stdlib (Go 1.22+ enhanced routing)
2. **SQL layer**: sqlc — write SQL, generate type-safe Go code
3. **DI**: Manual constructor injection — no framework
4. **Room/Facility**: In scope for MVP — basic rooms with capacity, scheduler considers room availability

## Unresolved Questions
1. **Student data**: Does timetable need to consider student enrollment numbers or just teacher capacity?

---

## References
- [Research: Odoo Architecture](/plans/reports/researcher-260219-1151-odoo-architecture.md)
- [Research: Go ERP Architecture](/plans/reports/researcher-260219-1151-go-erp-architecture-research.md)
