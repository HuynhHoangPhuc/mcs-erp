# MCS-ERP Documentation

Complete documentation for the multi-tenant, agentic-first ERP system focused on academic course scheduling.

## Quick Navigation

### For New Developers
Start here to understand the codebase:
1. **[Codebase Summary](./codebase-summary.md)** — Overview of Go packages, frontend structure, and tech stack
2. **[System Architecture](./system-architecture.md)** — 6 modules, their boundaries, and how they interact
3. **[Code Standards](./code-standards.md)** — DDD patterns, naming conventions, and implementation examples

### For Project Managers
Start here for product context:
1. **[Project Overview & PDR](./project-overview-pdr.md)** — Product vision, functional requirements, and success metrics
2. **[Development Roadmap](./development-roadmap.md)** — Timeline, phases, resource planning, and next steps

### For Architects
Start here for system design:
1. **[System Architecture](./system-architecture.md)** — Module boundaries, data flow, and security architecture
2. **[Code Standards](./code-standards.md)** — DDD layers, repository pattern, and cross-module communication

---

## Documentation Files

| File | Purpose | Size | Audience |
|------|---------|------|----------|
| **codebase-summary.md** | Go packages, frontend structure, tech stack, development setup | 475 LOC | Developers |
| **system-architecture.md** | 6 modules, boundaries, data flow, multi-tenancy, deployment | 614 LOC | Architects, Developers |
| **code-standards.md** | DDD patterns, HTTP handlers, repository pattern, code review checklist | 1,017 LOC | Developers |
| **project-overview-pdr.md** | Product vision, FRs, NFRs, data model, API endpoints, success metrics | 645 LOC | PMs, Stakeholders |
| **development-roadmap.md** | MVP status, immediate next steps, phased timeline, resource planning | 554 LOC | PMs, Developers |

**Total:** 3,305 LOC across 5 files

---

## Key Information At a Glance

### Tech Stack
- **Backend:** Go 1.22+, stdlib net/http, PostgreSQL 16, Redis 7
- **Frontend:** React 19, TypeScript, TanStack (Router/Query/Form), shadcn/ui, Turborepo
- **Architecture:** Modular monolith with DDD layers (domain → application → infrastructure → delivery)
- **Multi-tenancy:** Schema-per-tenant PostgreSQL isolation
- **AI:** Multi-LLM support (Claude, OpenAI, Ollama) with tool registry

### 6 Modules
1. **Core** — Authentication, RBAC, permissions (no dependencies)
2. **HR** — Teachers, departments, availability (depends on core)
3. **Subject** — Subjects, categories, prerequisites with DAG validation (depends on core)
4. **Room** — Rooms, capacity, availability (depends on core)
5. **Timetable** — Scheduling engine with greedy + simulated annealing (depends on core, hr, subject, room)
6. **Agent** — AI chatbot with tool registry (depends on core)

### Current Status
- ✅ All 8 MVP phases complete (Feb 2026)
- ✅ 6 modules fully implemented
- ✅ 60+ REST API endpoints
- ⚠️ Frontend module UIs are scaffolds (need implementation)
- ⚠️ Comprehensive tests still needed (target 80% coverage); integration + security suites now guard key workflows

### Immediate Next Steps (Week 1-2)
1. Backend testing & hardening (2 days)
2. Frontend module implementation (5 days)
3. API documentation & OpenAPI spec (1 day)
4. Development environment setup docs (1 day)

---

## Common Tasks

### Setting Up Development Environment
1. Read [codebase-summary.md](./codebase-summary.md) → "Development" section
2. Follow quick start: `docker compose up -d && make dev`
3. Run unit tests: `make test`
4. Run integration tests (requires Docker Compose): `make test-integration`
5. Run linter: `make lint`

### Adding a New Feature
1. Review module structure in [system-architecture.md](./system-architecture.md)
2. Follow DDD layers from [code-standards.md](./code-standards.md)
3. Add permission constant to core/domain/permission.go
4. Add API endpoint with auth middleware
5. Add unit tests (target 80% coverage)

### Understanding Module Boundaries
1. Check module dependency graph in [system-architecture.md](./system-architecture.md)
2. Review cross-module adapter pattern (timetable → hr/subject/room)
3. See module.go for public API exports

### Reviewing Code
1. Use [code-standards.md](./code-standards.md) → "Code Review Checklist"
2. Verify DDD layers are used correctly
3. Check HTTP handler pattern is followed
4. Ensure tenant context propagated through all layers

### Planning Features
1. Check [development-roadmap.md](./development-roadmap.md) for timeline
2. Review [project-overview-pdr.md](./project-overview-pdr.md) for scope
3. Cross-reference with [system-architecture.md](./system-architecture.md) for feasibility

---

## Code Examples by Topic

### DDD Layer Pattern
See [code-standards.md](./code-standards.md) → "Domain-Driven Design (DDD) Layers"
- Domain layer (domain/user.go)
- Application layer (commands/queries)
- Infrastructure layer (postgres_user_repo.go)
- Delivery layer (user_handler.go)
- Module layer (module.go)

### HTTP Handler Pattern
See [code-standards.md](./code-standards.md) → "HTTP Handler Patterns"
- Standard handler with tenant extraction, validation, business logic, response
- Middleware chaining (auth → permission → handler)

### Repository Pattern with Tenant Isolation
See [code-standards.md](./code-standards.md) → "Repository Pattern with Tenant Isolation"
- Interface definition in domain/
- Implementation with `SET LOCAL search_path = $1` for tenant isolation
- ListByTenant, GetByID, Create, Update, Delete patterns

### Cross-Module Communication
See [system-architecture.md](./system-architecture.md) → "Module Boundaries"
- Adapter pattern (CrossModuleReader in timetable)
- Repository imports from other modules
- No circular dependencies (topological sort at startup)

### Module Bootstrap
See [code-standards.md](./code-standards.md) → "Module Bootstrap Pattern (main.go)"
- Module registry initialization
- Dependency resolution with Kahn's algorithm
- Route registration and migration

---

## Policies & Conventions

### Go Naming
- **Files:** snake_case (e.g., `auth_service.go`, `postgres_user_repo.go`)
- **Directories:** kebab-case (e.g., `/internal/core/application/services/`)
- **Packages:** lowercase no underscore (e.g., `package core`)

### Permissions
Format: `module:resource:action`
- Example: `hr:teacher:read`, `core:user:write`
- List in [code-standards.md](./code-standards.md) → "Permission Constants"

### API Endpoints
- Pattern: `/api/v1/{module}/{resource}`
- Auth: JWT in Authorization header
- Tenant: From subdomain or X-Tenant-ID header
- All endpoints documented in [project-overview-pdr.md](./project-overview-pdr.md) → "API Endpoints"

### Error Handling
- Domain errors: Custom types in domain/errors.go
- HTTP mapping: Delivery layer maps domain errors to status codes
- Errors must implement error interface with descriptive messages

---

## Frequently Asked Questions

**Q: Where do I add a new permission?**
A: core/domain/permission.go → Add permission constant (module:resource:action format)

**Q: How do I add a cross-module feature?**
A: See [system-architecture.md](./system-architecture.md) → "Cross-Module Communication". Use adapter pattern like timetable/infrastructure/cross_module_reader.go

**Q: What's the tenant isolation strategy?**
A: Schema-per-tenant. See [system-architecture.md](./system-architecture.md) → "Multi-Tenancy Architecture". All queries use `SET LOCAL search_path = tenant_schema`

**Q: How do I run tests?**
A: `make test` in project root. See [codebase-summary.md](./codebase-summary.md) → "Development" for details.

**Q: Where's the OpenAPI spec?**
A: Not yet generated. Roadmap: Phase 9 (see [development-roadmap.md](./development-roadmap.md))

**Q: Can I use gRPC between modules?**
A: Protobuf defined but not used yet. See platform/grpc/. Planned for future phases.

**Q: How do I add a new module?**
A: Implement pkg/module.Module interface. See [code-standards.md](./code-standards.md) → "Module Bootstrap Pattern"

**Q: What's the scheduling algorithm?**
A: Greedy assignment + simulated annealing optimization. See [system-architecture.md](./system-architecture.md) → "Timetable Module"

**Q: Where's the deployment guide?**
A: Development only (Docker Compose). Production deployment in roadmap Phase 12. See [development-roadmap.md](./development-roadmap.md)

---

## Documentation Maintenance

### Weekly
- Update [development-roadmap.md](./development-roadmap.md) with sprint progress
- Add blockers/risks to roadmap

### Monthly
- Sync [system-architecture.md](./system-architecture.md) with implementation changes
- Review and update [code-standards.md](./code-standards.md) based on PRs

### Quarterly
- Update [project-overview-pdr.md](./project-overview-pdr.md) with success metrics
- Full review of all docs for accuracy

### On Release
- Update [development-roadmap.md](./development-roadmap.md) version history
- Update phase status (✅/⏳/❌)
- Add release notes

---

## Getting Help

### I need to understand...

**...how modules interact?**
→ [system-architecture.md](./system-architecture.md) → "Module Dependency Graph" & "Request Lifecycle"

**...the code structure?**
→ [codebase-summary.md](./codebase-summary.md) → "Project Structure" & "Go Conventions"

**...how to write a handler?**
→ [code-standards.md](./code-standards.md) → "HTTP Handler Patterns"

**...product requirements?**
→ [project-overview-pdr.md](./project-overview-pdr.md) → "Functional Requirements"

**...what we're building next?**
→ [development-roadmap.md](./development-roadmap.md) → "Immediate Next Steps"

**...multi-tenancy?**
→ [system-architecture.md](./system-architecture.md) → "Multi-Tenancy Architecture"

**...RBAC?**
→ [code-standards.md](./code-standards.md) → "Permission Constants" & [system-architecture.md](./system-architecture.md) → "Security Architecture"

---

## Document Versions

| File | Version | Last Updated | Status |
|------|---------|--------------|--------|
| codebase-summary.md | 1.0 | 2026-02-19 | Current |
| system-architecture.md | 1.0 | 2026-02-19 | Current |
| code-standards.md | 1.0 | 2026-02-19 | Current |
| project-overview-pdr.md | 1.0 | 2026-02-19 | Current |
| development-roadmap.md | 1.0 | 2026-02-19 | Current |

---

## Related Resources

- **GitHub:** [HuynhHoangPhuc/mcs-erp](https://github.com/HuynhHoangPhuc/mcs-erp)
- **Backend README:** `/README.md`
- **Frontend README:** `/web/README.md`
- **Makefile Targets:** `make help` or see [codebase-summary.md](./codebase-summary.md)
- **API Endpoints:** [project-overview-pdr.md](./project-overview-pdr.md) → "API Endpoints"
- **Environment Setup:** `.env.example` in project root

---

**Last Updated:** 2026-02-19
**Maintained By:** Development Team
**Next Review:** 2026-02-26
