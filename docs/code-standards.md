# Code Standards & Patterns

## Go File & Directory Naming

### File Naming Convention
Use **snake_case** for Go files, matching the package name:
```
internal/core/
├── auth_handler.go       # HTTP handler for auth endpoints
├── user_handler.go       # HTTP handler for user endpoints
├── role_handler.go       # HTTP handler for role endpoints
├── auth_service.go       # Business logic service
├── permission.go         # Permission constants & helpers
├── user.go               # User entity definition
└── errors.go             # Error type definitions

internal/core/infrastructure/
├── jwt_service.go        # JWT token generation/validation
├── postgres_user_repo.go # PostgreSQL user repository
├── postgres_role_repo.go # PostgreSQL role repository
```

### Directory Naming Convention
Use **kebab-case** for directories:
```
internal/
├── core/
├── platform/
│   ├── auth/
│   ├── config/
│   ├── database/
│   ├── eventbus/
│   ├── grpc/
│   ├── module/
│   └── tenant/
├── hr/
├── subject/
├── room/
├── timetable/
└── agent/
```

**Rationale:** Go idiom uses snake_case for files. Directory kebab-case improves discoverability for LLM tools (Grep, Glob).

---

## Domain-Driven Design (DDD) Layers

Every module follows a 6-layer structure:

### 1. Domain Layer (`domain/`)
**Purpose:** Business logic, rules, and interfaces (no framework dependencies)

**Files:**
- **{entity}.go** — Value objects, aggregates (e.g., `user.go`, `teacher.go`)
- **repository.go** — Interface definitions (no implementation)
- **events.go** — Domain events (e.g., UserCreated, TeacherAssigned)
- **errors.go** — Custom error types

**Constraints:**
- No imports from `delivery`, `infrastructure`, or `application`
- No HTTP, database, or external service dependencies
- Pure business rules: validation, invariants, state transitions

**Example: domain/user.go**
```go
package core

import "context"

// User represents an authenticated user in the system.
type User struct {
    ID       string   // UUID
    Email    string
    Roles    []string // Role names
    Permissions []string
}

// UserRepository defines operations on users (no implementation here).
type UserRepository interface {
    Create(ctx context.Context, schema string, user *User) error
    GetByEmail(ctx context.Context, schema, email string) (*User, error)
    ListByTenant(ctx context.Context, schema string) ([]*User, error)
}

// ErrUserNotFound indicates the user does not exist.
type ErrUserNotFound struct {
    Email string
}

func (e ErrUserNotFound) Error() string {
    return "user not found: " + e.Email
}
```

---

### 2. Application Layer (`application/`)
**Purpose:** Use cases, orchestration, business logic services

**Subdirectories:**
- **services/** — Business logic orchestrators
- **commands/** — Write operations (mutations)
- **queries/** — Read operations (no side effects)

**Pattern: Commands**
```go
// application/commands/create_user_command.go
package commands

import (
    "context"
    "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
)

type CreateUserCommand struct {
    Email    string
    Password string
}

type CreateUserHandler struct {
    userRepo domain.UserRepository
}

func NewCreateUserHandler(repo domain.UserRepository) *CreateUserHandler {
    return &CreateUserHandler{userRepo: repo}
}

func (h *CreateUserHandler) Handle(ctx context.Context, schema string, cmd *CreateUserCommand) (*domain.User, error) {
    // Validate input
    if cmd.Email == "" {
        return nil, errors.New("email required")
    }

    // Check preconditions
    existing, _ := h.userRepo.GetByEmail(ctx, schema, cmd.Email)
    if existing != nil {
        return nil, errors.New("email already exists")
    }

    // Create & persist
    user := &domain.User{Email: cmd.Email}
    if err := h.userRepo.Create(ctx, schema, user); err != nil {
        return nil, err
    }

    return user, nil
}
```

**Pattern: Queries**
```go
// application/queries/list_users_query.go
package queries

type ListUsersQuery struct {
    Limit  int
    Offset int
}

type ListUsersHandler struct {
    userRepo domain.UserRepository
}

func (h *ListUsersHandler) Handle(ctx context.Context, schema string, q *ListUsersQuery) ([]*domain.User, error) {
    return h.userRepo.ListByTenant(ctx, schema)  // Query returns data, no mutations
}
```

---

### 3. Infrastructure Layer (`infrastructure/`)
**Purpose:** Database, external services, framework-specific implementations

**Files:**
- **postgres_{entity}_repo.go** — Repository implementations
- **{service}.go** — Service implementations (JWT, cache, etc.)
- **adapters/** — Cross-module adapters (timetable imports hr/subject/room repos)

**Constraints:**
- Imports domain layer
- Contains database queries, external API calls
- Handles schema isolation (`SET LOCAL search_path`)

**Example: infrastructure/postgres_user_repo.go**
```go
package infrastructure

import (
    "context"
    "database/sql"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
)

type PostgresUserRepo struct {
    pool *pgxpool.Pool
}

func NewPostgresUserRepo(pool *pgxpool.Pool) *PostgresUserRepo {
    return &PostgresUserRepo{pool: pool}
}

func (r *PostgresUserRepo) Create(ctx context.Context, schema string, user *domain.User) error {
    const query = `
        SET LOCAL search_path = $1;
        INSERT INTO users (id, email, password_hash, created_at)
        VALUES ($2, $3, $4, now())
        RETURNING id, email
    `

    err := r.pool.QueryRow(ctx, query, schema, user.ID, user.Email).Scan(&user.ID, &user.Email)
    if err != nil {
        return err
    }
    return nil
}

func (r *PostgresUserRepo) GetByEmail(ctx context.Context, schema, email string) (*domain.User, error) {
    const query = `
        SET LOCAL search_path = $1;
        SELECT id, email FROM users WHERE email = $2
    `

    user := &domain.User{}
    err := r.pool.QueryRow(ctx, query, schema, email).Scan(&user.ID, &user.Email)
    if err == sql.ErrNoRows {
        return nil, domain.ErrUserNotFound{Email: email}
    }
    if err != nil {
        return nil, err
    }
    return user, nil
}
```

**Pattern: Cross-Module Adapter (timetable)**
```go
// timetable/infrastructure/cross_module_reader.go
package infrastructure

import (
    hrDomain "github.com/HuynhHoangPhuc/mcs-erp/internal/hr/domain"
    subjectDomain "github.com/HuynhHoangPhuc/mcs-erp/internal/subject/domain"
    roomDomain "github.com/HuynhHoangPhuc/mcs-erp/internal/room/domain"
)

// CrossModuleReader adapts repos from other modules to timetable's needs.
type CrossModuleReader struct {
    teacherRepo   hrDomain.TeacherRepository
    availRepo     hrDomain.AvailabilityRepository
    subjectRepo   subjectDomain.SubjectRepository
    roomRepo      roomDomain.RoomRepository
    roomAvailRepo roomDomain.RoomAvailabilityRepository
}

func NewCrossModuleReaderFromRepos(
    teacherRepo   hrDomain.TeacherRepository,
    availRepo     hrDomain.AvailabilityRepository,
    subjectRepo   subjectDomain.SubjectRepository,
    roomRepo      roomDomain.RoomRepository,
    roomAvailRepo roomDomain.RoomAvailabilityRepository,
) *CrossModuleReader {
    return &CrossModuleReader{
        teacherRepo:   teacherRepo,
        availRepo:     availRepo,
        subjectRepo:   subjectRepo,
        roomRepo:      roomRepo,
        roomAvailRepo: roomAvailRepo,
    }
}

// GetTeacher delegates to hr.TeacherRepository
func (r *CrossModuleReader) GetTeacher(ctx context.Context, schema, teacherID string) (*hrDomain.Teacher, error) {
    return r.teacherRepo.GetByID(ctx, schema, teacherID)
}

// GetTeacherAvailability delegates to hr.AvailabilityRepository
func (r *CrossModuleReader) GetTeacherAvailability(ctx context.Context, schema, teacherID string) (*hrDomain.Availability, error) {
    return r.availRepo.GetByTeacherID(ctx, schema, teacherID)
}
```

---

### 4. Delivery Layer (`delivery/`)
**Purpose:** HTTP handlers, middleware, request/response marshaling

**Files:**
- **{resource}_handler.go** — HTTP handler implementations
- **json_helpers.go** — Request/response marshaling helpers
- **middleware.go** — Handler-specific middleware

**Constraints:**
- Depends on domain, application, infrastructure
- HTTP-specific code only (no business logic)
- Parse JSON, validate input, call application layer

**Pattern: HTTP Handler**
```go
// delivery/user_handler.go
package delivery

import (
    "encoding/json"
    "net/http"
    "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
    "github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
)

type UserHandler struct {
    userRepo domain.UserRepository
}

func NewUserHandler(repo domain.UserRepository) *UserHandler {
    return &UserHandler{userRepo: repo}
}

type CreateUserRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type UserResponse struct {
    ID    string `json:"id"`
    Email string `json:"email"`
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // 1. Extract tenant
    schema, err := tenant.FromContext(r.Context())
    if err != nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    // 2. Parse request
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    // 3. Validate
    if req.Email == "" {
        http.Error(w, "email required", http.StatusBadRequest)
        return
    }

    // 4. Call repo/service
    user := &domain.User{Email: req.Email}
    if err := h.userRepo.Create(r.Context(), schema, user); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 5. Return response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(UserResponse{ID: user.ID, Email: user.Email})
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
    schema, err := tenant.FromContext(r.Context())
    if err != nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    users, err := h.userRepo.ListByTenant(r.Context(), schema)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}
```

**Pattern: Middleware**
```go
// delivery/auth_middleware.go
package delivery

import (
    "net/http"
    "github.com/HuynhHoangPhuc/mcs-erp/internal/core/application/services"
    "github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"
)

// AuthMiddleware validates JWT and extracts claims into context.
func AuthMiddleware(authSvc *services.AuthService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract Authorization header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
                return
            }

            // Validate token
            claims, err := authSvc.ValidateToken(authHeader)
            if err != nil {
                http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
                return
            }

            // Store claims in context
            ctx := auth.WithUser(r.Context(), claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

---

### 5. Module Interface (`module.go`)
**Purpose:** Glue module components together, implement pkg/module.Module interface

**Pattern:**
```go
// core/module.go
package core

import (
    "context"
    "net/http"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/HuynhHoangPhuc/mcs-erp/internal/core/application/services"
    "github.com/HuynhHoangPhuc/mcs-erp/internal/core/delivery"
    "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
    "github.com/HuynhHoangPhuc/mcs-erp/internal/core/infrastructure"
)

// Module implements pkg/module.Module for the core (auth/user/role) module.
type Module struct {
    pool       *pgxpool.Pool
    jwtService *infrastructure.JWTService
    authSvc    *services.AuthService
    userRepo   domain.UserRepository
    roleRepo   domain.RoleRepository
}

// NewModuleWithDeps creates the core module wired with concrete dependencies.
func NewModuleWithDeps(pool *pgxpool.Pool, jwtSvc *infrastructure.JWTService) *Module {
    userRepo := infrastructure.NewPostgresUserRepo(pool)
    roleRepo := infrastructure.NewPostgresRoleRepo(pool)
    authSvc := services.NewAuthService(userRepo, roleRepo, jwtSvc)

    return &Module{
        pool:       pool,
        jwtService: jwtSvc,
        authSvc:    authSvc,
        userRepo:   userRepo,
        roleRepo:   roleRepo,
    }
}

func (m *Module) Name() string            { return "core" }
func (m *Module) Dependencies() []string   { return nil }
func (m *Module) Migrate(ctx context.Context) error { return nil }
func (m *Module) RegisterEvents(ctx context.Context) error { return nil }

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
    authHandler := delivery.NewAuthHandler(m.authSvc)
    userHandler := delivery.NewUserHandler(m.userRepo)
    roleHandler := delivery.NewRoleHandler(m.roleRepo)

    // Register all routes
    mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
    mux.HandleFunc("GET /api/v1/users", userHandler.ListUsers)
    // ... more routes
}

// AuthService returns the auth service for use by other modules.
func (m *Module) AuthService() *services.AuthService { return m.authSvc }
```

---

## HTTP Handler Patterns

### Standard Handler Pattern

```go
func (h *MyHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
    // 1. Extract tenant
    tenantSchema, err := tenant.FromContext(r.Context())
    if err != nil {
        http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
        return
    }

    // 2. Parse request body
    var req CreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
        return
    }

    // 3. Validate input
    if err := req.Validate(); err != nil {
        http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
        return
    }

    // 4. Call service/repo
    result, err := h.repo.Create(r.Context(), tenantSchema, &req)
    if err != nil {
        // Handle known errors with appropriate status codes
        if errors.Is(err, domain.ErrConflict) {
            http.Error(w, `{"error":"resource already exists"}`, http.StatusConflict)
            return
        }
        // Generic 500 for unknown errors
        http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
        return
    }

    // 5. Return response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(result)
}
```

### Middleware Chaining

```go
// Register route with chained middleware
authMw := delivery.AuthMiddleware(authSvc)
permMw := auth.RequirePermission("core:user:write")

mux.Handle("POST /api/v1/users",
    authMw(permMw(http.HandlerFunc(userHandler.CreateUser))))
```

---

## Repository Pattern with Tenant Isolation

### Interface Definition
```go
// domain/repository.go
type UserRepository interface {
    Create(ctx context.Context, schema string, user *User) error
    GetByID(ctx context.Context, schema, id string) (*User, error)
    GetByEmail(ctx context.Context, schema, email string) (*User, error)
    ListByTenant(ctx context.Context, schema string) ([]*User, error)
    Update(ctx context.Context, schema string, user *User) error
    Delete(ctx context.Context, schema, id string) error
}
```

### Implementation with WithTenantTx
```go
// infrastructure/postgres_user_repo.go
func (r *PostgresUserRepo) GetByEmail(ctx context.Context, schema, email string) (*User, error) {
    // Always set search_path before querying
    query := `
        SET LOCAL search_path = $1;
        SELECT id, email, password_hash, created_at
        FROM users
        WHERE email = $2
    `

    user := &User{}
    err := r.pool.QueryRow(ctx, query, schema, email).Scan(
        &user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, domain.ErrUserNotFound{Email: email}
    }
    if err != nil {
        return nil, err
    }

    return user, nil
}

func (r *PostgresUserRepo) ListByTenant(ctx context.Context, schema string) ([]*User, error) {
    query := `
        SET LOCAL search_path = $1;
        SELECT id, email, created_at
        FROM users
        ORDER BY created_at DESC
        LIMIT 100
    `

    rows, err := r.pool.Query(ctx, query, schema)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*User
    for rows.Next() {
        user := &User{}
        if err := rows.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
            return nil, err
        }
        users = append(users, user)
    }

    return users, rows.Err()
}
```

---

## Error Handling

### Domain Errors
```go
// domain/errors.go
package core

import "fmt"

// ErrNotFound indicates the resource was not found.
type ErrNotFound struct {
    Resource string
    ID       string
}

func (e ErrNotFound) Error() string {
    return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// ErrConflict indicates a resource already exists.
type ErrConflict struct {
    Resource string
    Field    string
    Value    string
}

func (e ErrConflict) Error() string {
    return fmt.Sprintf("%s %s already exists: %s", e.Resource, e.Field, e.Value)
}

// ErrValidation indicates invalid input.
type ErrValidation struct {
    Field   string
    Message string
}

func (e ErrValidation) Error() string {
    return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}
```

### HTTP Error Responses
```go
// delivery handlers map domain errors to HTTP status codes
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // ... validation ...

    user, err := h.userRepo.Create(r.Context(), schema, &req)
    if err != nil {
        var conflict domain.ErrConflict
        var notFound domain.ErrNotFound
        var validation domain.ErrValidation

        if errors.As(err, &conflict) {
            http.Error(w, fmt.Sprintf(`{"error":"%s"}`, conflict.Error()), http.StatusConflict)
            return
        }
        if errors.As(err, &validation) {
            http.Error(w, fmt.Sprintf(`{"error":"%s"}`, validation.Error()), http.StatusBadRequest)
            return
        }

        http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
```

---

## Permission Constants

Permission strings follow the pattern: `module:resource:action`

**Definition: core/domain/permission.go**
```go
const (
    // Core permissions
    PermUserRead  = "core:user:read"
    PermUserWrite = "core:user:write"
    PermRoleRead  = "core:role:read"
    PermRoleWrite = "core:role:write"

    // HR permissions
    PermTeacherRead  = "hr:teacher:read"
    PermTeacherWrite = "hr:teacher:write"
    PermDeptRead     = "hr:department:read"
    PermDeptWrite    = "hr:department:write"

    // Subject permissions
    PermSubjectRead  = "subject:subject:read"
    PermSubjectWrite = "subject:subject:write"

    // Room permissions
    PermRoomRead  = "room:room:read"
    PermRoomWrite = "room:room:write"

    // Timetable permissions
    PermTimetableRead  = "timetable:timetable:read"
    PermTimetableWrite = "timetable:timetable:write"

    // Agent permissions
    PermAgentChat      = "agent:chat:use"
    PermAgentChatRead  = "agent:chat:read"
    PermAgentChatWrite = "agent:chat:write"
)
```

**Usage:**
```go
import "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"

mux.Handle("POST /api/v1/teachers",
    authMw(auth.RequirePermission(domain.PermTeacherWrite)(http.HandlerFunc(teacherHandler.Create))))
```

---

## Context Usage

### Tenant Context
```go
import "github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"

// Store tenant in context
ctx = tenant.WithTenant(ctx, "tenant_abc123")

// Extract tenant from context
schema, err := tenant.FromContext(ctx)
if err != nil {
    // Tenant not found in context
}
```

### User/Auth Context
```go
import "github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"

// Store user in context (done by middleware)
ctx = auth.WithUser(ctx, claims)

// Extract user from context
claims, err := auth.UserFromContext(ctx)
if err != nil {
    // User not found in context
}
```

---

## Code Quality Guidelines

### 1. Dependency Injection
Inject dependencies via constructors; no globals:
```go
// Good
func NewUserHandler(repo domain.UserRepository) *UserHandler {
    return &UserHandler{repo: repo}
}

// Bad: Using global
var globalUserRepo domain.UserRepository

func CreateUser(w http.ResponseWriter, r *http.Request) {
    // Uses global repo
}
```

### 2. Interface Segregation
Define small, focused interfaces in the domain layer:
```go
// Good: Small, single-purpose interface
type UserRepository interface {
    GetByEmail(ctx context.Context, schema, email string) (*User, error)
}

// Bad: Large interface mixing concerns
type UserService interface {
    Create(...) error
    Update(...) error
    Delete(...) error
    GetByEmail(...) (*User, error)
    SendEmail(...) error
    LogActivity(...) error
}
```

### 3. Error Wrapping
Use `fmt.Errorf("%w", err)` to preserve error chain:
```go
result, err := r.pool.QueryRow(ctx, query).Scan(&id)
if err != nil {
    return fmt.Errorf("failed to scan user: %w", err)
}
```

### 4. SQL Query Safety
Always use parameterized queries (sqlc handles this):
```go
// Good: Parameters via placeholders
query := "SELECT id, email FROM users WHERE email = $1"
r.pool.QueryRow(ctx, query, email)

// Bad: String concatenation (SQL injection risk)
query := "SELECT id, email FROM users WHERE email = '" + email + "'"
```

### 5. Logging
Use structured logging with slog:
```go
import "log/slog"

slog.Info("user created", "user_id", user.ID, "email", user.Email)
slog.Error("failed to create user", "error", err, "email", email)
```

### 6. Testing
Organize tests alongside code with table-driven patterns:
```go
// user_test.go (same directory as user.go)
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        email   string
        wantErr bool
    }{
        {"user@example.com", false},
        {"invalid", true},
        {"", true},
    }

    for _, tt := range tests {
        t.Run(tt.email, func(t *testing.T) {
            err := domain.ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## Module Bootstrap Pattern (main.go)

```go
func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
    slog.SetDefault(logger)

    // Load config
    cfg, err := config.Load()
    if err != nil {
        slog.Error("config load failed", "error", err)
        os.Exit(1)
    }

    // Setup graceful shutdown
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    // Database pool
    pool, err := database.NewPool(ctx, cfg.DatabaseURL)
    if err != nil {
        slog.Error("database connection failed", "error", err)
        os.Exit(1)
    }
    defer pool.Close()

    // Shared services
    jwtSvc := infrastructure.NewJWTService(cfg.JWTSecret, cfg.JWTExpiry)
    registry := platformmod.NewRegistry()

    // Module registration (order matters for dependencies)
    coreMod := core.NewModuleWithDeps(pool, jwtSvc)
    if err := registry.Register(coreMod); err != nil {
        slog.Error("failed to register core module", "error", err)
        os.Exit(1)
    }

    hrMod := hr.NewModule(pool, coreMod.AuthService())
    if err := registry.Register(hrMod); err != nil {
        slog.Error("failed to register hr module", "error", err)
        os.Exit(1)
    }

    subjectMod := subject.NewModule(pool, coreMod.AuthService())
    if err := registry.Register(subjectMod); err != nil {
        slog.Error("failed to register subject module", "error", err)
        os.Exit(1)
    }

    roomMod := room.NewModule(pool, coreMod.AuthService())
    if err := registry.Register(roomMod); err != nil {
        slog.Error("failed to register room module", "error", err)
        os.Exit(1)
    }

    timetableMod := timetable.NewModuleWithRepos(
        pool, coreMod.AuthService(),
        hrMod.TeacherRepo(), hrMod.AvailabilityRepo(),
        subjectMod.SubjectRepo(),
        roomMod.RoomRepo(), roomMod.RoomAvailabilityRepo(),
    )
    if err := registry.Register(timetableMod); err != nil {
        slog.Error("failed to register timetable module", "error", err)
        os.Exit(1)
    }

    agentRegistry := agentinfra.NewToolRegistry()
    agentMod := agent.NewModule(pool, coreMod.AuthService(), agentRegistry, cfg.LLMConfig, cfg.RedisURL)
    if err := registry.Register(agentMod); err != nil {
        slog.Error("failed to register agent module", "error", err)
        os.Exit(1)
    }

    // Resolve startup order
    modules, err := registry.ResolveOrder()
    if err != nil {
        slog.Error("failed to resolve module order", "error", err)
        os.Exit(1)
    }

    // Initialize modules
    mux := http.NewServeMux()
    for _, m := range modules {
        if err := m.Migrate(ctx); err != nil {
            slog.Error("migration failed", "module", m.Name(), "error", err)
            os.Exit(1)
        }
        if err := m.RegisterEvents(ctx); err != nil {
            slog.Error("event registration failed", "module", m.Name(), "error", err)
            os.Exit(1)
        }
        m.RegisterRoutes(mux)
    }

    // Middleware wrapping mux
    wrappedMux := tenantMiddleware(mux)

    // Start HTTP server
    server := &http.Server{
        Addr:    ":8080",
        Handler: wrappedMux,
    }

    go func() {
        slog.Info("starting server", "addr", server.Addr)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            slog.Error("server error", "error", err)
        }
    }()

    // Graceful shutdown
    <-ctx.Done()
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := server.Shutdown(shutdownCtx); err != nil {
        slog.Error("shutdown error", "error", err)
    }
}

func tenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract tenant from subdomain or header
        schema, err := tenant.Resolver(r)
        if err != nil {
            http.Error(w, `{"error":"tenant not found"}`, http.StatusUnauthorized)
            return
        }

        // Store in context
        ctx := tenant.WithTenant(r.Context(), schema)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

## Code Review Checklist

Before committing, ensure:

- [ ] File naming follows snake_case (Go files) and kebab-case (directories)
- [ ] Code organized into DDD layers (domain → application → infrastructure → delivery)
- [ ] Interfaces defined in domain layer only
- [ ] All HTTP handlers follow standard pattern (extract tenant → parse → validate → call → respond)
- [ ] Tenant context propagated through all layers
- [ ] No business logic in delivery layer
- [ ] All database queries use parameters (no string concatenation)
- [ ] Error handling with custom error types
- [ ] Permission checks on protected endpoints
- [ ] Tests provided for business logic
- [ ] No global state or magic globals
- [ ] Dependency injection via constructors
- [ ] Logging with slog using structured fields
- [ ] No hardcoded secrets or credentials
