---
title: "MCS-ERP System Implementation"
description: "Multi-tenant agentic ERP with academic MVP: HR, Subject, Room, Timetable, AI Agent"
status: in-progress
priority: P1
effort: 14w
branch: main
tags: [erp, go, react, ddd, cqrs, ai-agent, multi-tenant]
created: 2026-02-19
updated: 2026-02-19
---

# MCS-ERP System Implementation Plan

## Architecture Summary
Go 1.22+ modular monolith (DDD, CQRS, stdlib net/http, sqlc, Watermill). React 19 + TanStack frontend (Turborepo). PostgreSQL 16 schema-per-tenant + Redis 7. Docker Compose deploy.

## Phases

| # | Phase | Effort | Status | Progress |
|---|-------|--------|--------|----------|
| 1 | [Project Scaffolding](./phase-01-project-scaffolding.md) | 3d | complete | 100% |
| 2 | [Core Platform](./phase-02-core-platform.md) | 5d | complete | 100% |
| 3 | [Auth & RBAC](./phase-03-auth-rbac.md) | 5d | complete | 100% |
| 4 | [HR Module](./phase-04-hr-module.md) | 5d | complete | 100% |
| 5 | [Subject Module](./phase-05-subject-module.md) | 5d | complete | 100% |
| 6 | [Room Module](./phase-06-room-module.md) | 3d | complete | 100% |
| 7 | [Timetable Module](./phase-07-timetable-module.md) | 8d | complete | 100% |
| 8 | [AI Agent Module](./phase-08-ai-agent-module.md) | 6d | complete | 100% |

**Total estimated effort:** ~8 weeks (40 working days)

## Dependencies
```
Phase 1 (Scaffolding)
  -> Phase 2 (Core Platform)
    -> Phase 3 (Auth & RBAC)
      -> Phase 4 (HR) \
      -> Phase 5 (Subject) > all independent, can parallelize
      -> Phase 6 (Room)   /
        -> Phase 7 (Timetable) -- depends on HR, Subject, Room
          -> Phase 8 (AI Agent) -- depends on all domain modules
```

## Key Research References
- [Brainstorm Report](../reports/brainstorm-260219-1151-mcs-erp-architecture.md)
- [Odoo Architecture](../reports/researcher-260219-1151-odoo-architecture.md)
- [Go ERP Patterns](../reports/researcher-260219-1151-go-erp-architecture-research.md)
- [Go Backend Patterns](./research/researcher-01-go-backend-patterns.md)
- [Frontend/Scheduling/AI](./research/researcher-02-frontend-scheduling-ai.md)

## Tech Stack
| Layer | Choice |
|-------|--------|
| Backend | Go 1.22+, net/http (REST external), gRPC (internal), sqlc, Watermill, manual DI |
| Frontend | React 19, TanStack (Router/Query/Table/Form), shadcn/ui, Tailwind |
| Database | PostgreSQL 16 (schema-per-tenant), Redis 7 |
| AI | langchaingo, multi-provider (Claude/OpenAI/Ollama) |
| Deploy | Docker Compose |
| CI | GitHub Actions |

## Modules
1. **Core** - auth, tenant, RBAC, user, module registry, event bus
2. **HR** - teacher, department, availability, qualifications
3. **Subject** - subject, prerequisite DAG, categories, credits
4. **Room** - room, capacity, equipment, availability
5. **Timetable** - semester, scheduling (greedy+SA), assignments, review workflow
6. **Agent** - multi-LLM, tool registry, chat (SSE), inline suggestions

## Validation Log

### Session 1 — 2026-02-19
**Trigger:** Initial plan creation validation
**Questions asked:** 7

#### Questions & Answers

1. **[Architecture]** The plan uses `X-Tenant-ID` header for tenant resolution. In production, how should tenants be identified when users access the system?
   - Options: Subdomain | X-Tenant-ID header only | JWT claim | URL path prefix
   - **Answer:** Subdomain (faculty-a.mcs-erp.com → tenant 'faculty-a')
   - **Rationale:** More natural for multi-org SaaS. Frontend doesn't need to manage header. DNS-based routing.

2. **[Architecture]** The plan assumes schedule generation uses HTTP 202 + polling. For a 30-second operation, what async pattern?
   - Options: SSE progress stream | HTTP 202 + polling | WebSocket
   - **Answer:** SSE progress stream
   - **Rationale:** Real-time feedback during scheduling. Reuses same SSE pattern as AI chat. No polling overhead.

3. **[Assumptions]** How should user login flow work with schema-per-tenant?
   - Options: User belongs to 1 tenant | Tenant selection screen first | Multi-tenant user
   - **Answer:** User belongs to 1 tenant (MVP)
   - **Rationale:** Simplifies auth flow. System looks up user by email across `public.users` lookup table → resolves tenant. No tenant picker needed.

4. **[Architecture]** How should cross-module data access work in the modular monolith?
   - Options: Read-only service interfaces | Event-driven data sync | Shared query layer | API calls (HTTP)
   - **Answer:** Read-only service interfaces
   - **Rationale:** Clean dependency via interfaces. Each module exposes a Reader service. Timetable imports HR/Subject/Room readers. Testable via mocks.

5. **[Scope]** What's the faculty's actual time slot structure?
   - Options: Fixed periods (1-10/day) | Flexible time blocks | Morning/Afternoon sessions
   - **Answer:** Fixed periods (1-10 per day)
   - **Rationale:** Standard academic schedule. Simplifies constraint model — each period is a discrete slot.

6. **[Scope]** Should AI tool calls be audited differently than normal API calls?
   - Options: Same audit as API | Enhanced AI audit log | Human approval for destructive actions
   - **Answer:** Same audit as API
   - **Rationale:** KISS. No special AI audit trail for MVP. Existing request logging covers tool execution.

7. **[Scope]** How should the REST API be documented?
   - Options: OpenAPI/Swagger | Manual markdown | No docs for MVP
   - **Answer:** OpenAPI/Swagger
   - **Rationale:** Auto-generated from Go handlers. Industry standard. Enables frontend devs and partners to self-serve.

#### Confirmed Decisions
- **Tenant resolution:** Subdomain-based — update Phase 02 tenant middleware
- **Async scheduling:** SSE progress stream — update Phase 07 generation endpoint
- **Login flow:** Single-tenant user — update Phase 03 auth, add public.users_lookup table
- **Cross-module access:** Read-only service interfaces — confirmed Phase 07 approach
- **Time slots:** Fixed periods 1-10 — confirmed Phase 07 domain model
- **AI audit:** Same as API — no changes to Phase 08
- **API docs:** OpenAPI/Swagger — add to Phase 01 scaffolding

#### Action Items
- [ ] Phase 02: Update tenant resolver to extract from subdomain (not just header)
- [ ] Phase 03: Add public.users_lookup table for cross-tenant email lookup
- [ ] Phase 07: Change schedule generation from HTTP 202+polling to SSE progress stream
- [ ] Phase 01: Add OpenAPI/Swagger tooling (swaggo/swag or oapi-codegen) to scaffolding
- [ ] All phases: gRPC for internal module-to-module communication, REST for external API

#### Impact on Phases
- Phase 01: Add swagger/openapi dependency, protobuf tooling (buf/protoc), Makefile targets
- Phase 02: Tenant resolver primary = subdomain, fallback = X-Tenant-ID header (for dev/testing). Add gRPC interceptors for tenant context propagation.
- Phase 03: Login lookup via public.users_lookup(email → tenant_schema), not per-schema search
- Phase 04-06: Each module exposes gRPC service for internal reads (reader service), REST handlers for external CRUD
- Phase 07: Generate endpoint returns SSE stream with progress events + final schedule. Cross-module reads via gRPC (HR, Subject, Room reader services)
- Phase 08: AI tool execution calls module gRPC services internally

### Session 2 — 2026-02-19
**Trigger:** Re-validation before implementation — surface remaining assumptions and tradeoffs
**Questions asked:** 7

#### Questions & Answers

1. **[Architecture]** Phase 02 plans a real gRPC server (:9090) with interceptors for internal module communication. In an MVP modular monolith, all modules run in the same process. Should we use actual gRPC or direct Go interfaces for cross-module reads?
   - Options: Direct Go interfaces (Recommended) | In-process gRPC | Full gRPC server on :9090
   - **Answer:** Full gRPC server on :9090
   - **Rationale:** Ready for microservice split from day one. Establishes proto contracts early. User accepts heavier setup for future-proofing.

2. **[Architecture]** The frontend uses separate pnpm packages per module (module-hr, module-subject, etc.) with Turborepo. For an MVP with one developer, should we simplify this structure?
   - Options: Separate packages (as planned) | Directories in shell app (Recommended) | Fewer packages
   - **Answer:** Separate packages (as planned)
   - **Rationale:** Clean module boundaries, parallel builds, independent versioning. Worth the boilerplate for long-term maintainability.

3. **[Scope]** Redis 7 is in Docker Compose but only explicitly used for AI conversation caching (Phase 08). Should Redis be included in the MVP stack or defer until needed?
   - Options: Include Redis from start (Recommended) | Defer Redis | Redis for specific features only
   - **Answer:** Include Redis from start
   - **Rationale:** Already in Docker Compose. Useful for rate limiting, session cache, future pub/sub. Minimal overhead.

4. **[Architecture]** Both Teacher (Phase 04) and Room (Phase 06) availability use identical weekly slot grids (day + period). Should the availability pattern be shared code from the start?
   - Options: Shared in pkg/erptypes (Recommended) | Duplicate per module | Share after Room module
   - **Answer:** Duplicate per module
   - **Rationale:** Avoids coupling between HR and Room domains. Each module owns its availability types. May diverge later with module-specific needs.

5. **[Scope]** Phase 07 mentions teacher-subject assignment as 'manual or auto' but doesn't detail auto-assignment logic. For MVP, how should teachers be assigned to subjects before scheduling?
   - Options: Manual only (Recommended) | Auto-assign by qualification match | Hybrid: manual with suggestions
   - **Answer:** Hybrid: manual with suggestions
   - **Rationale:** Manual assignment with a "suggest" button that recommends teachers based on availability + qualifications. Adds value without full auto-assign complexity.

6. **[Scope]** Phase 08 includes 'inline suggestions' (contextual action buttons on entity pages). Is this MVP-critical or can it be deferred?
   - Options: Defer to post-MVP (Recommended) | Include in MVP | Include but simplified
   - **Answer:** Include in MVP
   - **Rationale:** Rule-based suggestions are simple (no LLM call). Adds discovery of AI features. Worth the effort.

7. **[Risk]** Phase 05 flags concurrent prerequisite edits as a risk (two users adding edges simultaneously could bypass cycle detection). Should we address this in MVP?
   - Options: Advisory lock on graph mutations (Recommended) | Skip for MVP | Optimistic locking with version
   - **Answer:** Optimistic locking with version
   - **Rationale:** Version column on prerequisites graph. Retry on conflict. More complex than advisory locks but non-blocking.

#### Confirmed Decisions
- **Internal communication:** Full gRPC on :9090 — keep Phase 02 gRPC plan as-is
- **Frontend structure:** Separate packages per module — keep plan as-is
- **Redis:** Include from Phase 01 — no changes needed (already planned)
- **Availability types:** Duplicate per module — Phase 04 and 06 each own their availability, no shared extraction
- **Teacher assignment:** Hybrid manual+suggestions — update Phase 07 semester setup
- **AI inline suggestions:** Include in MVP — keep Phase 08 scope as-is
- **DAG concurrency:** Optimistic locking with version — update Phase 05 prerequisite graph

#### Action Items
- [ ] Phase 07: Add teacher-subject suggestion logic (qualification match + availability check) to semester setup
- [ ] Phase 05: Replace advisory lock approach with optimistic locking (version column on subject_prerequisites)
- [ ] Phase 04/06: Confirm each module defines its own availability types independently (no shared extraction)

#### Impact on Phases
- Phase 05: Add version column to subject_prerequisites table. Implement optimistic locking in add_prerequisite command (check version, retry on conflict).
- Phase 07: Add teacher suggestion endpoint/logic to semester setup. "Suggest teachers" button queries HR for teachers with matching qualifications + available slots. Admin confirms assignments.

## Implementation Completion Summary — 2026-02-19

**All 8 phases COMPLETE.** Full implementation of multi-tenant agentic ERP system ready for comprehensive testing.

### Delivered Components

#### Backend (Go)
- Phase 01: Project scaffolding (Go module, Docker Compose, Makefile, CI)
- Phase 02: Core platform (tenant resolution, migration runner, module registry, event bus, gRPC)
- Phase 03: Authentication & RBAC (JWT, user/role/permission CRUD, auth middleware)
- Phase 04: HR module (teacher/department CRUD, availability grid, domain events, AI tools)
- Phase 05: Subject module (prerequisite DAG, cycle detection, topological sort, optimistic locking)
- Phase 06: Room module (room CRUD, availability, equipment tagging, AI tools)
- Phase 07: Timetable module (semester management, greedy + simulated annealing scheduling, conflict detection, SSE progress stream, cross-module integration)
- Phase 08: AI Agent module (multi-LLM support, tool registry, SSE chat, inline suggestions, conversation history)

#### Frontend (React + TypeScript)
- Turborepo monorepo with per-module packages
- Shell app with TanStack Router, authentication flow
- Shared UI package (shadcn/ui + Tailwind)
- API client with TanStack Query
- Module-specific pages and components:
  - HR: teacher list, form, availability grid
  - Subject: subject list, form, DAG visualization
  - Room: room list, form, availability grid
  - Timetable: semester wizard, timetable grid, conflict panel
  - Agent: chat sidebar, suggestion bar, conversation history

#### Database (PostgreSQL)
- Schema-per-tenant architecture with `_template` for codegen
- Full migration suite across all modules
- Optimistic locking for DAG mutations
- Conversation history + Redis caching

#### Security & Quality
- Code review completed: Fixed gRPC tenant validation, JSON injection, IDOR on conversations, request body limits, os.Exit in goroutines
- `go build ./...` and `go vet ./...` pass cleanly
- All domain layer tests for critical paths (cycle detection, topo sort, constraint satisfaction)
- RBAC permission enforcement on all endpoints and AI tools

### Status Transition
- **Previous:** All phases pending (0% progress)
- **Current:** All phases complete (100% progress)
- **Next:** Comprehensive integration testing and user acceptance validation
