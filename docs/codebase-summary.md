# MCS-ERP Codebase Summary

## Overview
Multi-tenant, agentic-first ERP system built with Go (backend), React (frontend), and PostgreSQL (database). Designed as a modular monolith using Domain-Driven Design (DDD) principles with schema-per-tenant isolation.

**Module path:** github.com/HuynhHoangPhuc/mcs-erp
**Go version:** 1.22+

## Project Structure

```
.
├── cmd/server/              # Go entry point (main.go, module bootstrap)
├── internal/                # Core business logic (DDD modules)
│   ├── core/                # Auth, users, roles, permissions (no dependencies)
│   ├── hr/                  # Teachers, departments, availability
│   ├── subject/             # Subjects, categories, prerequisites (DAG validation)
│   ├── room/                # Rooms, availability, capacity tracking
│   ├── timetable/           # Semester, scheduling (greedy + annealing)
│   ├── agent/               # AI chatbot, multi-LLM provider, tool registry
│   └── platform/            # Shared infrastructure (tenant, auth, database, gRPC, config)
├── pkg/                     # Public packages (module interface, erptypes)
├── migrations/              # SQL schema migrations (per module)
├── sqlc/                    # sqlc config + generated query types
├── proto/                   # Protobuf definitions (internal gRPC)
├── web/                     # Frontend monorepo (React 19, TanStack)
├── docker-compose.yml       # PostgreSQL 16, Redis 7
├── go.mod/go.sum            # Go dependencies
└── Makefile                 # Build targets
```

## Backend Packages

### /pkg (Public API)
- **pkg/module/module.go** — Module interface (Name, Dependencies, RegisterRoutes, RegisterEvents, Migrate)
- **pkg/erptypes/id.go** — UUID helpers
- **pkg/erptypes/errors.go** — Domain error types

### /internal/platform (Infrastructure Layer)
Shared services used by all modules:

| Package | Purpose |
|---------|---------|
| **platform/config** | Environment loading (DATABASE_URL, JWT_SECRET, REDIS_URL) |
| **platform/database** | PostgreSQL connection pool, schema-per-tenant setup, migrator |
| **platform/tenant** | Tenant context (WithTenant, FromContext), tenant resolver from subdomain/header |
| **platform/auth** | JWT claims extraction, permission middleware (RequirePermission) |
| **platform/module** | Module registry, topological sort (Kahn's algorithm) for startup order |
| **platform/eventbus** | Watermill in-process pub/sub (extensible for event-driven features) |
| **platform/grpc** | gRPC server setup, tenant interceptor for internal services |

### /internal/core (Auth & RBAC)
**Dependencies:** None (foundation module)

Implements:
- User authentication (login/refresh/logout with JWT)
- Role-based access control (RBAC) with granular permissions
- User-role assignment
- Permission checking middleware

**Structure (DDD):**
- **domain/** — User, Role, Permission value objects; UserRepository interface
- **application/services/** — AuthService (login, refresh validation)
- **infrastructure/** — PostgresUserRepo, PostgresRoleRepo, JWTService
- **delivery/** — REST handlers, auth middleware

**Key Components:**
- `NewModuleWithDeps(pool, jwtSvc)` — Wires core module
- `AuthMiddleware(authSvc)` — Validates JWT, extracts claims, stores in context
- `RequirePermission(perm)` — Checks if user has permission (403 if missing)

**Routes:**
```
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
POST   /api/v1/users
GET    /api/v1/users
GET    /api/v1/users/{id}
POST   /api/v1/users/{id}/roles
POST   /api/v1/roles
GET    /api/v1/roles
GET    /api/v1/roles/{id}
DELETE /api/v1/roles/{id}
```

### /internal/hr (Human Resources)
**Dependencies:** core

Manages teachers and departments with availability tracking.

**Entities:**
- Teacher (id, name, email, department_id, is_active)
- Department (id, name, description)
- Availability (teacher_id, day 0–6, period 1–10, is_available)

**Key Patterns:**
- **Cross-module access:** `TeacherRepo()`, `AvailabilityRepo()` exported for timetable module
- **WithTenantTx:** All repos wrap queries in `SET LOCAL search_path = schema`

**Routes:**
```
POST   /api/v1/teachers
GET    /api/v1/teachers
GET    /api/v1/teachers/{id}
PUT    /api/v1/teachers/{id}
GET    /api/v1/teachers/{id}/availability
PUT    /api/v1/teachers/{id}/availability
POST   /api/v1/departments
GET    /api/v1/departments
GET    /api/v1/departments/{id}
PUT    /api/v1/departments/{id}
DELETE /api/v1/departments/{id}
```

### /internal/subject (Subject Catalog)
**Dependencies:** core

Manages subjects with prerequisite DAG and cycle detection.

**Entities:**
- Subject (id, code, name, credits, category_id)
- Category (id, name)
- Prerequisite (subject_id, prerequisite_subject_id)

**Key Patterns:**
- **DAG validation:** NewPrerequisiteRepo ensures acyclic graphs on insert
- **Prerequisite chain:** ListPrerequisites, GetPrerequisiteChain for forward/backward lookup
- **Cross-module access:** `SubjectRepo()` exported for timetable

**Routes:**
```
POST   /api/v1/subjects
GET    /api/v1/subjects
GET    /api/v1/subjects/{id}
PUT    /api/v1/subjects/{id}
POST   /api/v1/categories
GET    /api/v1/categories
GET    /api/v1/categories/{id}
POST   /api/v1/subjects/{id}/prerequisites
DELETE /api/v1/subjects/{id}/prerequisites/{prereqId}
GET    /api/v1/subjects/{id}/prerequisites
GET    /api/v1/subjects/{id}/prerequisite-chain
```

### /internal/room (Room Management)
**Dependencies:** core

Manages classroom/lab resources with availability by time slot.

**Entities:**
- Room (id, name, code, building, floor, capacity, equipment[], is_active)
- RoomAvailability (room_id, day 0–6, period 1–10, is_available)

**Key Patterns:**
- **Migrations:** Migrate(ctx) runs DDL across all active tenant schemas
- **Cross-module access:** `RoomRepo()`, `RoomAvailabilityRepo()` for timetable
- **Capacity & equipment:** Used by scheduler for constraint checking

**Routes:**
```
POST   /api/v1/rooms
GET    /api/v1/rooms
GET    /api/v1/rooms/{id}
PUT    /api/v1/rooms/{id}
GET    /api/v1/rooms/{id}/availability
PUT    /api/v1/rooms/{id}/availability
```

### /internal/timetable (Scheduling Engine)
**Dependencies:** core, hr, subject, room

Generates conflict-free schedules using greedy algorithm + simulated annealing.

**Entities:**
- Semester (id, name, start_date, end_date, is_active)
- SemesterSubject (semester_id, subject_id, teacher_id, requested_hours)
- Schedule (id, semester_id, status: DRAFT/APPROVED, created_at)
- Assignment (id, schedule_id, subject_id, teacher_id, room_id, day, period)

**Key Patterns:**
- **Cross-module adapters:** Infrastructure layer imports repos from hr/subject/room
- `NewModuleWithRepos(...)` — Convenience constructor for main.go (adapter wiring inside)
- **ProblemBuilder interface:** Abstraction for scheduler algorithm (greedy, annealing, etc.)
- **Stream progress:** SSE endpoint for long-running schedule generation

**Routes:**
```
POST   /api/v1/timetable/semesters
GET    /api/v1/timetable/semesters
GET    /api/v1/timetable/semesters/{id}
POST   /api/v1/timetable/semesters/{id}/subjects
POST   /api/v1/timetable/semesters/{id}/subjects/{subjectId}/teacher
POST   /api/v1/timetable/semesters/{id}/generate
GET    /api/v1/timetable/semesters/{id}/schedule
POST   /api/v1/timetable/semesters/{id}/approve
PUT    /api/v1/timetable/assignments/{id}
```

### /internal/agent (AI Chatbot)
**Dependencies:** core

Multi-provider LLM agent with tool registry for other modules to extend.

**Entities:**
- Conversation (id, user_id, title, created_at, updated_at)
- Message (id, conversation_id, role: USER/ASSISTANT, content)
- Tool (id, name, description, schema, handler_func)

**Key Patterns:**
- **Tool registry:** Shared with other modules; they register tools at bootstrap
- **Provider service:** Multi-LLM support (Claude/OpenAI/Ollama)
- **Message cache:** Optional Redis cache for retrieval
- **SSE streaming:** Real-time message chunks during agent execution
- **Inline suggestions:** Rule-based suggestions without LLM call

**Routes:**
```
POST   /api/v1/agent/chat (SSE)
GET    /api/v1/agent/conversations
POST   /api/v1/agent/conversations
GET    /api/v1/agent/conversations/{id}
PATCH  /api/v1/agent/conversations/{id}
DELETE /api/v1/agent/conversations/{id}
GET    /api/v1/agent/suggestions
```

## Database Schema

### Schema-per-Tenant
- **Public schema:** Shared tenants table, users_lookup (email → tenant mapping)
- **Tenant schemas:** One schema per tenant (e.g., `tenant_abc123`) containing:
  - users, roles, permissions, teachers, departments, subjects, rooms, timetables, etc.

### Isolation Mechanism
All queries wrap database access in:
```go
db.Exec("SET LOCAL search_path = $1", tenantSchema)
```
Within the same transaction, all unqualified tables resolve to tenant schema.

## Go Conventions

### File Naming
- **snake_case** for Go files (e.g., `auth_service.go`, `postgres_user_repo.go`)
- **kebab-case** for directories (e.g., `/internal/core/application/services/`)

### DDD Layers (per module)
1. **domain/** — Value objects, interfaces, business rules (no framework deps)
2. **application/commands/** — Command handlers (mutations)
3. **application/queries/** — Query handlers (reads)
4. **application/services/** — Business logic orchestration
5. **infrastructure/** — Database, external service implementations
6. **delivery/** — REST handlers, middleware, JSON marshaling

### HTTP Handler Pattern
```go
type MyHandler struct {
    repo domain.MyRepository
}

func (h *MyHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request body
    var req CreateMyRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    // 2. Extract tenant context
    tenantSchema, err := tenant.FromContext(r.Context())
    if err != nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    // 3. Call service/repo
    result, err := h.repo.Create(r.Context(), tenantSchema, &req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 4. Return JSON response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

### Repository Pattern with Tenant
```go
type PostgresUserRepo struct {
    pool *pgxpool.Pool
}

func (r *PostgresUserRepo) Create(ctx context.Context, schema string, req *CreateRequest) (*User, error) {
    const query = `SET LOCAL search_path = $1;
        INSERT INTO users (id, email, ...) VALUES (...)
        RETURNING id, email, ...`

    user := &User{}
    err := r.pool.QueryRow(ctx, query, schema, ...).Scan(&user.ID, &user.Email, ...)
    return user, err
}
```

### Module Registration (main.go)
```go
// Create registry (dependency resolution at startup)
registry := platformmod.NewRegistry()

// Register in dependency order (registry validates)
registry.Register(coreMod)        // No deps
registry.Register(hrMod)           // Depends on core
registry.Register(subjectMod)      // Depends on core
registry.Register(roomMod)         // Depends on core
registry.Register(timetableMod)    // Depends on core, hr, subject, room

// Resolve and bootstrap
modules, err := registry.ResolveOrder()
for _, mod := range modules {
    mod.RegisterRoutes(mux)
    mod.RegisterEvents(ctx)
    mod.Migrate(ctx)
}
```

## Frontend Structure (/web)

### Monorepo Layout (Turborepo + pnpm)
```
web/
├── apps/shell/              # Main React 19 app (routes, providers)
│   └── src/
│       ├── routes/          # Page components
│       ├── providers/       # Context providers (auth, theme)
│       ├── hooks/           # Custom React hooks
│       └── lib/             # Utilities
├── packages/ui/             # Shared shadcn/ui components
├── packages/api-client/     # OpenAPI-generated types + TanStack Query hooks
├── packages/module-hr/      # HR feature module
├── packages/module-subject/ # Subject feature module
├── packages/module-room/    # Room feature module
├── packages/module-timetable/ # Timetable feature module
└── packages/module-agent/   # Agent feature module
```

### Tech Stack
- **React 19** with TypeScript
- **TanStack Router** — File-based routing
- **TanStack Query (React Query)** — Server state management + API integration
- **TanStack Table** — Advanced data tables
- **TanStack Form** — Form state management
- **shadcn/ui** — Headless UI component library
- **Tailwind CSS** — Utility-first styling
- **Turborepo** — Monorepo task orchestration
- **pnpm** — Package manager

## Key Implementation Patterns

### 1. Multi-Tenancy
- Subdomain routing: `tenant-abc.example.com` → tenant resolution
- Fallback: `X-Tenant-ID` header for API clients
- Automatic schema isolation via `SET LOCAL search_path`

### 2. RBAC (Role-Based Access Control)
- Permission constants: `module:resource:action` (e.g., `hr:teacher:write`)
- Middleware chain: AuthMiddleware → RequirePermission → Handler
- Admin role: Has all permissions

### 3. Cross-Module Communication
- **Importer pattern:** Timetable imports and calls repos from hr/subject/room
- **Adapter layer:** CrossModuleReader translates between module interfaces
- **No circular deps:** Registry enforces DAG via topological sort

### 4. Error Handling
- Domain errors in domain/errors.go (e.g., ErrNotFound, ErrConflict)
- HTTP mapping: HTTP status codes in delivery layer (no leaking domain to HTTP)

### 5. Testing
- Unit tests per layer (domain, application, infrastructure)
- Test doubles for repos and services
- Table-driven tests for multiple scenarios

## Configuration

### Environment Variables (.env)
```
DATABASE_URL=postgres://user:pass@localhost:5432/mcs_erp
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h
REDIS_URL=redis://localhost:6379
AI_PROVIDER=claude|openai|ollama
OPENAI_API_KEY=...
CLAUDE_API_KEY=...
```

### Module Initialization Order
1. Load config
2. Create database pool (with _template schema)
3. Create JWT service
4. Create module registry
5. Register modules (core → hr → subject → room → timetable → agent)
6. Start HTTP server with mux

## Development

### Quick Start
```bash
# Install dependencies
go mod download
pnpm install --filter=web

# Start PostgreSQL & Redis
docker compose up -d

# Run migrations (manual via migrations/ directory)
# Run backend (hot-reload with air)
make dev

# Run frontend
cd web && pnpm dev
```

### Makefile Targets
- `make dev` — Start Go server with hot-reload
- `make build` — Build Go binary
- `make test` — Run Go tests
- `make lint` — Run golangci-lint
- `make sqlc` — Generate sqlc query types
- `make proto` — Generate Protobuf/gRPC code
- `make swagger` — Generate OpenAPI docs

### Code Generation
- **sqlc:** Database query types generated from SQL
- **Protobuf:** gRPC service definitions (internal communication)
- **buf:** Protocol buffer code generator

## Security Considerations

### Input Validation
- Request body limits (http.MaxBytesReader)
- JSON schema validation in handlers
- SQL parameter binding (no string concatenation)

### Authentication
- JWT tokens (expiring, cryptographically signed)
- Refresh token flow for long-lived sessions

### Authorization
- Permission checks on every protected endpoint
- Tenant isolation prevents cross-tenant data access
- IDOR prevention: Always validate tenant ownership in queries

### Data Protection
- Passwords hashed with bcrypt
- Tenant schema isolation at database level
- Redis messages optional (can be disabled)

## Performance Considerations

- **Connection pooling:** pgx pool with configurable size
- **Query optimization:** Prepared statements via sqlc
- **Caching:** Optional Redis for agent message cache
- **Scheduling:** Annealing algorithm for large timetables
- **Pagination:** Limit/offset on all list endpoints

## Next Steps (Beyond MVP)

1. **Testing coverage** — Add comprehensive unit + integration tests
2. **Frontend polish** — Complete module UIs, improve UX
3. **Deployment** — Docker images, Kubernetes manifests
4. **Monitoring** — Logging aggregation, metrics, alerting
5. **API docs** — Auto-generated OpenAPI/Swagger UI
6. **Event-driven features** — Watermill event handlers for cross-module workflows
