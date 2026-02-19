# System Architecture

## High-Level Overview

MCS-ERP is a **modular monolith** using Domain-Driven Design (DDD) principles. The system is organized into 6 bounded contexts (modules), each responsible for a specific domain:

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP/REST API Layer                   │
│              (stdlib net/http with ServeMux)             │
└──────────┬──────────────────────────────────────────────┘
           │
     ┌─────┴──────────┬─────────────┬──────────┬──────────┬──────────┐
     │                │             │          │          │          │
┌────▼────┐    ┌─────▼────┐  ┌─────▼────┐ ┌──▼──────┐ ┌─▼────────┐ ┌▼─────────┐
│   Core   │    │    HR    │  │ Subject  │ │  Room  │ │Timetable │ │  Agent   │
│ (auth)   │    │(teachers)│  │(subjects)│ │(rooms) │ │(schedule)│ │(AI tools)│
└────┬────┘    └─────┬────┘  └─────┬────┘ └──┬─────┘ └─┬────────┘ └┬────────┘
     │                │             │          │        │          │
     │                └────────┬─────┴──────────┴────────┴──────────┘
     │                         │
     │                    ┌────▼──────────────┐
     │                    │  Platform Layer   │
     │                    │ (tenant, auth,    │
     │                    │  database, grpc)  │
     │                    └────┬──────────────┘
     │                         │
     └─────────────────────────┼──────────────────┐
                               │                  │
                        ┌──────▼──────┐  ┌────────▼──────┐
                        │  PostgreSQL  │  │    Redis      │
                        │   (schema-   │  │  (optional    │
                        │ per-tenant)  │  │   caching)    │
                        └──────────────┘  └───────────────┘
```

## Module Dependency Graph

```
Timetable
├── depends on: Core, HR, Subject, Room
└── public API: SemesterRepo, ScheduleRepo

Agent
├── depends on: Core
└── public API: ToolRegistry, AgentService

Room
├── depends on: Core
└── public API: RoomRepo, RoomAvailabilityRepo

Subject
├── depends on: Core
└── public API: SubjectRepo

HR
├── depends on: Core
└── public API: TeacherRepo, AvailabilityRepo

Core (Foundation)
├── depends on: (none)
└── public API: AuthService, UserRepo, RoleRepo
```

## Module Boundaries

### 1. Core Module
**Location:** `/internal/core`

Responsible for:
- User authentication (JWT-based login/logout)
- Role-based access control (RBAC)
- Permission management (granular resource-action permissions)
- User lifecycle (create, read, list)

**Key Entities:**
- User (email, password_hash, roles[], tenant_schema)
- Role (name, permissions[])
- Permission (module:resource:action format)

**Public API (for other modules):**
- `AuthService()` — JWT validation, claims extraction
- `UserRepo` — User CRUD interface

**Routes:** 12 endpoints for auth, users, roles

**Database Tables:**
- users
- roles
- user_roles (M2M join)
- permissions
- role_permissions (M2M join)

---

### 2. HR Module
**Location:** `/internal/hr`

Responsible for:
- Teacher management (CRUD)
- Department management (CRUD)
- Teacher availability tracking (7 days × 10 time periods)

**Key Entities:**
- Teacher (id, name, email, department_id, is_active)
- Department (id, name, description)
- Availability (teacher_id, day 0–6, period 1–10, is_available)

**Cross-Module Dependencies:**
- Depends on: Core
- Exports to: Timetable (TeacherRepo, AvailabilityRepo)

**Routes:** 8 endpoints for teachers, departments, availability

**Database Tables:**
- teachers
- departments
- teacher_availability (7 × 10 = 70 possible slots per teacher)

**Key Pattern: Repository Export**
```go
// In hr/module.go
func (m *Module) TeacherRepo() domain.TeacherRepository { return m.teacherRepo }
func (m *Module) AvailabilityRepo() domain.AvailabilityRepository { return m.availRepo }

// Timetable imports these in main.go via NewModuleWithRepos(...)
```

---

### 3. Subject Module
**Location:** `/internal/subject`

Responsible for:
- Subject catalog (CRUD)
- Subject categorization
- Prerequisite management (directed acyclic graph)
- Prerequisite validation (cycle detection)

**Key Entities:**
- Subject (id, code, name, credits, category_id)
- Category (id, name)
- Prerequisite (subject_id, prerequisite_subject_id)

**Cross-Module Dependencies:**
- Depends on: Core
- Exports to: Timetable (SubjectRepo)

**Routes:** 8 endpoints for subjects, categories, prerequisites

**Database Tables:**
- subjects
- subject_categories (or embedded in subjects)
- subject_prerequisites (directed edges)

**Key Pattern: DAG Validation**
```go
// NewPrerequisiteRepo validates acyclic graph on insert
// Uses DFS to detect cycles before persisting edge

type PrerequisiteRepository interface {
    AddPrerequisite(ctx, schema, subjectID, prereqID) error  // Returns ErrCyclicDependency
    ListPrerequisites(ctx, schema, subjectID) []Subject
    GetPrerequisiteChain(ctx, schema, subjectID) []Subject   // All transitive deps
}
```

---

### 4. Room Module
**Location:** `/internal/room`

Responsible for:
- Classroom/lab resource management
- Room metadata (capacity, equipment, location)
- Room availability tracking (7 days × 10 periods per room)
- Migration logic across tenant schemas

**Key Entities:**
- Room (id, name, code, building, floor, capacity, equipment[], is_active)
- RoomAvailability (room_id, day 0–6, period 1–10, is_available)

**Cross-Module Dependencies:**
- Depends on: Core
- Exports to: Timetable (RoomRepo, RoomAvailabilityRepo)

**Routes:** 6 endpoints for rooms and availability

**Database Tables:**
- rooms
- room_availability (7 × 10 = 70 possible slots per room)

**Key Pattern: Migrate(ctx)**
```go
// Room module implements Migrate to run DDL across all active tenant schemas
func (m *Module) Migrate(ctx context.Context) error {
    migrator := database.NewMigrator(m.pool)
    if err := migrator.MigrateAll(ctx, sqlCreateRoomsTable); err != nil {
        return err
    }
    return migrator.MigrateAll(ctx, sqlCreateRoomAvailabilityTable)
}
```

---

### 5. Timetable Module
**Location:** `/internal/timetable`

Responsible for:
- Semester management (CRUD)
- Subject-teacher-room assignment generation
- Schedule conflict detection and resolution
- Admin approval workflow

**Key Entities:**
- Semester (id, name, start_date, end_date, is_active)
- SemesterSubject (semester_id, subject_id, teacher_id, requested_hours)
- Schedule (id, semester_id, status: DRAFT/APPROVED)
- Assignment (id, schedule_id, subject_id, teacher_id, room_id, day, period)

**Cross-Module Dependencies:**
- Depends on: Core, HR, Subject, Room
- Imports: TeacherRepo, AvailabilityRepo, SubjectRepo, RoomRepo, RoomAvailabilityRepo

**Routes:** 8 endpoints for semesters, schedule generation, approval

**Database Tables:**
- semesters
- semester_subjects
- schedules
- assignments

**Key Pattern: Cross-Module Adapter**
```go
// timetable/infrastructure/cross_module_reader.go
type CrossModuleReader struct {
    teacherRepo    hrDomain.TeacherRepository
    availRepo      hrDomain.AvailabilityRepository
    subjectRepo    subjectDomain.SubjectRepository
    roomRepo       roomDomain.RoomRepository
    roomAvailRepo  roomDomain.RoomAvailabilityRepository
}

// NewModuleWithRepos convenience constructor (in timetable/module.go)
func NewModuleWithRepos(
    pool        *pgxpool.Pool,
    authSvc     *services.AuthService,
    teacherRepo hrDomain.TeacherRepository,
    availRepo   hrDomain.AvailabilityRepository,
    subjectRepo subjectDomain.SubjectRepository,
    roomRepo    roomDomain.RoomRepository,
    roomAvail   roomDomain.RoomAvailabilityRepository,
) *Module {
    reader := infrastructure.NewCrossModuleReaderFromRepos(...)
    return NewModule(pool, authSvc, reader)
}
```

**Scheduling Algorithm:**
1. Build constraint problem (teachers, subjects, rooms, time periods)
2. Greedy algorithm: Assign subjects to available slots
3. Simulated annealing: Optimize for minimal conflicts
4. Return DRAFT schedule or error if infeasible
5. Admin reviews and approves to APPROVED status

---

### 6. Agent Module
**Location:** `/internal/agent`

Responsible for:
- Multi-LLM provider integration (Claude, OpenAI, Ollama)
- Tool registry for modules to extend agent capabilities
- Conversation management (create, list, read, delete)
- Real-time message streaming (SSE)
- Inline suggestions (rule-based, no LLM call)

**Key Entities:**
- Conversation (id, user_id, title, created_at, updated_at)
- Message (id, conversation_id, role: USER/ASSISTANT, content)
- Tool (id, name, description, schema, handler)

**Cross-Module Dependencies:**
- Depends on: Core
- No direct exports; provides ToolRegistry for others to register tools

**Routes:** 8 endpoints for chat (SSE), conversations CRUD, suggestions

**Database Tables:**
- conversations
- messages

**Key Pattern: Tool Registry**
```go
// agent/infrastructure/tool_registry.go
type ToolRegistry struct {
    tools map[string]Tool
}

// Other modules register tools at bootstrap
registry.RegisterTool("create_semester", func(ctx, args) {...})
registry.RegisterTool("list_teachers", func(ctx, args) {...})

// Agent calls registered tools during chat processing
```

**Features:**
- SSE streaming: Chunks sent as agent processes tool calls
- Redis cache (optional): Store conversations for quick retrieval
- Multi-provider: Switch between Claude, OpenAI, Ollama via config
- Tool-calling pattern: Agent decides which tools to invoke

---

## Platform Layer

**Location:** `/internal/platform`

Shared infrastructure used by all modules:

### Tenant Management (`platform/tenant`)
- **WithTenant(ctx, schema)** — Stores tenant schema in context
- **FromContext(ctx)** — Retrieves tenant schema (or errors)
- **Resolver** — Extracts tenant from subdomain or X-Tenant-ID header

### Database (`platform/database`)
- **NewPool(ctx, dsn)** — Create pgx connection pool
- **NewMigrator(pool)** — Run DDL across all tenant schemas
- **Schema isolation:** `SET LOCAL search_path = $1` per transaction

### Authentication (`platform/auth`)
- **UserFromContext(ctx)** — Extract JWT claims
- **RequirePermission(perm)** — Middleware for permission checks (403 if denied)

### Module Registry (`platform/module`)
- **Registry** — Store and manage modules
- **ResolveOrder()** — Topological sort (Kahn's algorithm) to resolve startup order
- **Detects circular dependencies** at startup

### Event Bus (`platform/eventbus`)
- **Watermill in-process pub/sub** for event-driven features
- Future: Swap in Redis/Kafka transport layer without changing module code

### gRPC Server (`platform/grpc`)
- **Optional internal communication** between modules
- Tenant interceptor for schema isolation on gRPC calls

### Config (`platform/config`)
- **Load()** — Parse environment variables
- DATABASE_URL, JWT_SECRET, REDIS_URL, AI_PROVIDER, API keys

---

## Data Flow: Example Timetable Generation

```
1. User (POST /api/v1/timetable/semesters/{id}/generate)
   ↓
2. HTTP Handler (ScheduleHandler.GenerateSchedule)
   ├─ Extract tenant from context
   ├─ Validate authorization (PermTimetableWrite)
   └─ Call ScheduleService
   ↓
3. ScheduleService (orchestration layer)
   ├─ Load semester, subjects, teachers from respective repos
   ├─ Query teacher availability (from HR module)
   ├─ Query subject prerequisites (from Subject module)
   ├─ Query rooms & availability (from Room module)
   └─ Call ProblemBuilder.BuildSchedule()
   ↓
4. ProblemBuilder (greedy + annealing algorithm)
   ├─ Iterate subjects to assign
   ├─ Check teacher availability
   ├─ Check room availability
   ├─ Check prerequisite constraints
   ├─ Run annealing for optimization
   └─ Return draft assignments
   ↓
5. Persist Schedule & Assignments
   ├─ BEGIN TRANSACTION
   ├─ SET LOCAL search_path = tenant_schema
   ├─ INSERT INTO schedules VALUES (...)
   ├─ INSERT INTO assignments VALUES (...) [multiple rows]
   ├─ COMMIT
   └─ Return schedule ID
   ↓
6. Return HTTP 201 with schedule ID & assignments (JSON array)
```

## DDD Layers (per module)

Each module follows this structure:

```
internal/{module}/
├── domain/                  # Business logic, interfaces, value objects
│   ├── {entity}.go         # User, Teacher, Subject, etc.
│   ├── repository.go       # Interface definitions
│   ├── events.go           # Domain events
│   └── errors.go           # Custom error types
├── application/            # Use cases, orchestration
│   ├── services/           # Business logic orchestration
│   │   └── {service}.go    # e.g., AuthService, TeacherService
│   ├── commands/           # Write operations
│   │   └── {command}.go    # e.g., CreateTeacher command + handler
│   └── queries/            # Read operations
│       └── {query}.go      # e.g., ListTeachers query + handler
├── infrastructure/         # Database, external services
│   ├── postgres_{repo}.go  # Database repository implementations
│   ├── {service}.go        # External service clients (e.g., JWTService)
│   └── adapters/           # Cross-module adapters (if timetable)
├── delivery/               # HTTP handlers, middleware
│   ├── {handler}.go        # HTTP request handlers
│   ├── json_helpers.go     # Request/response marshaling
│   └── middleware.go       # Handler-specific middleware
└── module.go               # Module interface implementation
```

### Example: HR Module Structure
```
internal/hr/
├── domain/
│   ├── teacher.go
│   ├── department.go
│   ├── availability.go
│   ├── repository.go      # TeacherRepository, DepartmentRepository interfaces
│   └── errors.go
├── application/
│   └── services/
│       ├── teacher_service.go
│       └── department_service.go
├── infrastructure/
│   ├── postgres_teacher_repo.go
│   ├── postgres_department_repo.go
│   └── postgres_availability_repo.go
├── delivery/
│   ├── teacher_handler.go
│   ├── department_handler.go
│   ├── availability_handler.go
│   └── json_helpers.go
└── module.go               # Module interface + RegisterRoutes
```

## Request Lifecycle

```
HTTP Request (POST /api/v1/teachers)
    ↓
Router (net/http.ServeMux)
    ├─ Matches "POST /api/v1/teachers"
    └─ Calls registered handler
    ↓
Tenant Middleware (platform/tenant)
    ├─ Extracts tenant from subdomain/header
    ├─ Stores in context (WithTenant)
    └─ Calls next handler
    ↓
Auth Middleware (core/delivery)
    ├─ Validates JWT from Authorization header
    ├─ Extracts claims (user_id, roles, permissions)
    ├─ Stores in context (auth.WithUser)
    └─ Calls next handler
    ↓
Permission Middleware (platform/auth)
    ├─ Checks if user has required permission
    ├─ Returns 403 Forbidden if missing
    └─ Calls next handler
    ↓
Handler (e.g., TeacherHandler.CreateTeacher)
    ├─ Parses request body → CreateTeacherRequest
    ├─ Extracts tenant from context
    ├─ Validates input
    ├─ Calls repo.Create(ctx, tenantSchema, data)
    ↓
Repository (PostgresTeacherRepo)
    ├─ Executes query with SET LOCAL search_path = tenantSchema
    ├─ Scans result into Teacher struct
    └─ Returns to handler
    ↓
Handler (continued)
    ├─ Marshals Teacher to JSON
    ├─ Sets response headers (Content-Type: application/json)
    ├─ Writes HTTP 201 + body
    └─ Returns
    ↓
HTTP Response (201 Created with JSON body)
```

## Multi-Tenancy Architecture

### Isolation Strategy: Schema-Per-Tenant

**Setup:**
- PostgreSQL `public` schema: Tenants lookup table, shared auth
- Per-tenant schemas: `tenant_abc`, `tenant_xyz`, etc.
- Template schema: `_template` (used to clone new tenant schemas)

**Isolation Mechanism:**
```sql
-- In every transaction before querying tables:
SET LOCAL search_path = 'tenant_abc123';

-- All subsequent unqualified table references use tenant_abc123
SELECT * FROM teachers;  -- Actually queries tenant_abc123.teachers
```

**Context Propagation:**
1. Request arrives → Subdomain/header parsed → Tenant schema resolved
2. Stored in context via `tenant.WithTenant(ctx, schema)`
3. Passed through every layer (HTTP → service → repo)
4. Repository wraps query in `SET LOCAL search_path = schema`

**Benefits:**
- Hard isolation at database level (no accidental cross-tenant reads)
- Row-level security not needed (queries are schema-scoped)
- Easy to manage per-tenant data lifecycle (backup/restore/delete)
- Scales to many tenants (one connection pool serves all)

**Limitations:**
- Max PostgreSQL namespaces: ~16M (typically 64K–1M in practice)
- DDL changes require running migrations across all schemas

---

## Security Architecture

### Authentication
- **Method:** JWT (JSON Web Token)
- **Flow:** Login → JWT issued → Stored in client → Sent in Authorization header
- **Validation:** JWT signature verified on every protected endpoint
- **Expiry:** Configurable (default 24h)
- **Refresh:** Dedicated endpoint for refreshing expired tokens

### Authorization
- **Model:** RBAC (Role-Based Access Control)
- **Granularity:** `module:resource:action` permissions
- **Enforcement:** RequirePermission middleware on protected routes
- **Admin role:** Has all permissions

### Tenant Isolation
- **Scope:** All queries scoped to authenticated user's tenant schema
- **Enforcement:** Automatic via `SET LOCAL search_path`
- **IDOR prevention:** Repositories always query within tenant context

### Input Validation
- **JSON schema:** Request bodies validated against expected structure
- **Size limits:** http.MaxBytesReader on request bodies
- **SQL safety:** Parameterized queries via sqlc (no string concatenation)

### Data Protection
- **Passwords:** Hashed with bcrypt (never stored plaintext)
- **Secrets:** JWT secret stored in environment (not hardcoded)
- **Logs:** Sensitive data redacted (no passwords, API keys in logs)

---

## Performance Considerations

### Database
- **Connection pool:** pgx with configurable pool size (default 20)
- **Query optimization:** Indexes on foreign keys, search columns
- **Prepared statements:** Generated by sqlc
- **N+1 query prevention:** Batch queries where possible

### Caching
- **Redis (optional):** Agent message cache for quick conversation retrieval
- **HTTP caching:** Cache-Control headers on read endpoints
- **Client-side:** TanStack Query on frontend

### Scheduling
- **Algorithm:** Greedy + simulated annealing
- **Complexity:** O(n × m) where n = subjects, m = time periods
- **Timeout:** 30s limit per generation (or configurable)

### Scalability
- **Horizontal:** Stateless HTTP handlers (scale with load balancer)
- **Vertical:** Increase connection pool, database resources
- **Tenant isolation:** Each tenant's queries are independent

---

## Deployment Architecture

### Development (Docker Compose)
```yaml
services:
  postgres:
    image: postgres:16
    volumes:
      - pgdata:/var/lib/postgresql/data
  redis:
    image: redis:7
```

### Production (Suggested)
- **Backend:** Docker container, deployed on Kubernetes
- **Database:** Managed PostgreSQL (AWS RDS, Azure Database, GCP Cloud SQL)
- **Redis:** Managed cache (AWS ElastiCache, Azure Cache, GCP Memorystore)
- **Frontend:** Static assets on CDN (Vercel, Netlify, CloudFront)
- **API Gateway:** Load balancer with TLS termination
- **Monitoring:** Prometheus + Grafana, CloudWatch, Datadog, etc.

---

## Future Enhancements

1. **Microservices migration:** Extract modules to separate services using gRPC
2. **Event sourcing:** Replace CRUD tables with event logs + snapshots
3. **GraphQL API:** Alternative to REST (alongside, not replacing)
4. **WebSocket support:** Real-time updates for collaborative scheduling
5. **Advanced scheduling:** ML-based constraint optimization
6. **Multi-region deployment:** Geo-replication for disaster recovery
7. **Audit logging:** Track all data changes with who/when/why
