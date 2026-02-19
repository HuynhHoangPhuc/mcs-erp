# Go Backend Patterns for ERP Modular Monolith
**Date:** 2026-02-19 | **Author:** researcher-01

---

## Topic 1: sqlc with Schema-per-Tenant PostgreSQL

### Core Problem
sqlc generates code at build time against a static schema. Multi-tenant schema-per-tenant means the schema name (`tenant_abc.orders`) is only known at runtime.

### Solution: sqlc.yaml with `search_path` + runtime SET

**Codegen approach:** Write SQL queries without schema prefix; rely on `search_path`:
```sql
-- queries/orders.sql
-- name: ListOrders :many
SELECT * FROM orders WHERE tenant_id = $1;
```
sqlc generates against a fixed "template" schema (e.g., `public` or `_template`). At runtime, `SET search_path = tenant_xyz` routes to correct schema.

**pgxpool pattern for schema switching:**
```go
// Acquire conn from pool, set search_path, use, release
conn, _ := pool.Acquire(ctx)
defer conn.Release()
conn.Exec(ctx, "SET search_path = " + tenantSchema)
q := db.New(conn)  // sqlc Querier wraps the conn
```

**WARNING:** Do NOT use `SET search_path` on a pooled connection without resetting—pollutes the pool. Use `SET LOCAL` inside a transaction instead:
```go
tx, _ := pool.Begin(ctx)
tx.Exec(ctx, "SET LOCAL search_path = " + tenantSchema)
q := db.New(tx)
// ... queries
tx.Commit(ctx)
```

**Safer: pgxpool.Pool with `BeforeAcquire` / `AfterRelease`:**
```go
config.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
    schema := tenantFromCtx(ctx)
    conn.Exec(ctx, "SET search_path = " + schema + ", public")
    return true
}
config.AfterRelease = func(conn *pgx.Conn) bool {
    conn.Exec(context.Background(), "RESET search_path")
    return true
}
```

### Migrations Across Tenant Schemas
- Tool: `golang-migrate` or `goose` — run per-schema in a loop
- Pattern: maintain a `public.tenants` table; on migration, iterate all tenants:
```go
tenants, _ := getTenants(adminConn)
for _, t := range tenants {
    adminConn.Exec(ctx, "SET search_path = " + t.Schema)
    m.Up() // golang-migrate applies to current search_path
}
```
- Always apply to a `_template` schema first; new tenant schemas clone from it via `CREATE SCHEMA new_tenant; -- then run migrations`

### sqlc codegen with dynamic schema: verdict
sqlc **cannot** generate per-tenant code (no dynamic codegen). The pattern is: **one codegen against a canonical schema + runtime `search_path`**. This is idiomatic and widely used.

---

## Topic 2: Watermill GoChannel for CQRS in Modular Monolith

### Setup
```go
// One shared GoChannel pub/sub for entire process
pubSub := gochannel.NewGoChannel(
    gochannel.Config{BufferSize: 100, Persistent: false},
    watermill.NewStdLogger(false, false),
)
```

### CQRS Pattern (Commands, Queries, Events)
Watermill's `cqrs` package provides `CommandBus`, `EventBus`, `CommandProcessor`, `EventProcessor`:
```go
eventBus, _ := cqrs.NewEventBusWithConfig(pubSub, cqrs.EventBusConfig{
    GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
        return params.EventName, nil  // topic = event type name
    },
    Marshaler: cqrs.JSONMarshaler{},
})
```

### Module-to-Module Communication
- Each bounded context owns its topic namespace: `inventory.StockReserved`, `billing.InvoiceCreated`
- Publishers: live inside the emitting module's service layer
- Subscribers: live inside the consuming module's event handlers
- Wire up at app startup (not inside modules themselves):
```go
// main.go or app.go
inventoryHandler := inventory.NewEventHandler(inventoryRepo)
billingHandler   := billing.NewEventHandler(billingRepo)

router.AddNoPublisherHandler("billing.on_stock_reserved",
    "inventory.StockReserved", pubSub, billingHandler.Handle)
router.Run(ctx)
```

### Migration Path: GoChannel → NATS/Kafka
Watermill's abstraction is the killer feature. Only the `Publisher`/`Subscriber` implementation changes:
```go
// Dev/test:
pubSub := gochannel.NewGoChannel(...)
// Prod (swap one line):
pubSub, _ := nats.NewStreamingSubscriber(nats.StreamingSubscriberConfig{...})
```
Command/event handlers and the router are **unchanged**. This is the primary reason to choose Watermill.

### Bounded Context Structure
```
modules/
  inventory/
    commands/    # command handlers
    events/      # event handlers (consumers)
    publisher.go # wraps EventBus, publishes domain events
  billing/
    events/      # subscribes to inventory events
```

---

## Topic 3: Go Module Registry Pattern

### Core Pattern: `init()` + global registry map
Standard Go idiom (used by database/sql drivers, Caddy, etc.):
```go
// registry/registry.go
var modules = map[string]Module{}

func Register(m Module) {
    modules[m.Name()] = m
}

// Each module registers itself:
// inventory/module.go
func init() {
    registry.Register(&InventoryModule{})
}
```
Import with blank identifier in `main.go`:
```go
import _ "myapp/modules/inventory"
import _ "myapp/modules/billing"
```

### Module Interface Design
```go
type Module interface {
    Name() string
    Dependencies() []string      // for topo sort
    Init(app *App) error          // DI wiring
    RegisterRoutes(r *chi.Mux)
    RegisterEvents(bus *EventBus)
    Migrations() []Migration
    // Optional: AITools() []aitool.Tool
}
```

### Topological Sort for Dependency Resolution
```go
func sortModules(modules []Module) ([]Module, error) {
    // Kahn's algorithm: build adj list from Dependencies()
    // Return error on cycle detection
}
```
Standard: use `golang.org/x/tools/graph` or implement Kahn's inline (~40 lines).

### Caddy's Pattern (reference)
Caddy uses `caddy.RegisterModule(m)` in `init()`, stores in a sync.Map keyed by module ID (`namespace.name`). Each module implements `caddy.Module` interface returning `caddy.ModuleInfo`. Caddy does **not** do topo sort—it relies on JSON config ordering. For an ERP with real deps, implement topo sort explicitly.

### Compile-time Safety
Use interface assertions in `init()`:
```go
var _ Module = (*InventoryModule)(nil) // compile fails if interface not satisfied
```

---

## Key Decisions for MCS-ERP

| Concern | Recommendation |
|---|---|
| sqlc multi-tenant | One codegen + `SET LOCAL search_path` in tx |
| Connection pool | pgxpool `BeforeAcquire`/`AfterRelease` for schema reset |
| Migrations | goose per-tenant loop from `public.tenants` |
| Event bus | Watermill GoChannel now, NATS later (swap 1 line) |
| Module wiring | `init()` registry + topo sort + `_ "module/path"` imports |
| CQRS topics | Namespaced: `{module}.{EventName}` |

---

## Unresolved Questions

1. Should `search_path` be set per-connection (pool hook) or per-transaction (`SET LOCAL`)? Per-tx is safer but adds overhead per query batch.
2. Watermill GoChannel is not persistent—if a handler panics, events are lost. Is at-least-once delivery needed for in-process phase?
3. Topo sort: should circular module dependencies be a hard error at startup or a warning?
4. Template schema name convention: `_template`, `public`, or tenant-specific seed schema?
