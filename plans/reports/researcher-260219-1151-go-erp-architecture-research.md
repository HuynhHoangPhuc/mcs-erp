# Research Report: Go ERP/Modular Business System Architecture

**Date:** 2026-02-19
**Slug:** go-erp-architecture-research

---

## 1. DDD + Clean Architecture in Go

### Core Pattern

```
/internal
  /<module>           # bounded context
    /domain           # entities, value objects, aggregates, repository interfaces
    /application      # use cases / command handlers / query handlers
    /infrastructure   # DB adapters, external APIs, repo implementations
    /ports            # HTTP handlers, gRPC, CLI (delivery layer)
```

Dependency rule: `ports → application → domain`. Infrastructure implements domain interfaces.

### Key Patterns

**Aggregate Root**
```go
// domain/order/order.go
type Order struct {
    id       OrderID
    lines     []OrderLine
    status    Status
    events    []DomainEvent  // uncommitted events
}

func (o *Order) AddLine(product ProductID, qty int) error {
    // business invariant enforcement here
    o.lines = append(o.lines, ...)
    o.events = append(o.events, OrderLineAdded{...})
    return nil
}
```

**Repository Interface (domain layer)**
```go
type OrderRepository interface {
    FindByID(ctx context.Context, id OrderID) (*Order, error)
    Save(ctx context.Context, order *Order) error
}
```

**Use Case (application layer)**
```go
type AddOrderLineHandler struct {
    repo   domain.OrderRepository
    events EventPublisher
}
func (h *AddOrderLineHandler) Handle(ctx context.Context, cmd AddOrderLineCmd) error { ... }
```

### Real-World Go Examples
- **wild-workouts-go-ddd-example** (ThreeDotsLabs) — canonical reference, blog series on `threedots.tech`
- **go-clean-arch** (bxcodec) — simpler REST example
- **eventsourcing** (hallgren) — event sourcing focused

### Pros
- Clear invariant ownership; domain logic not scattered across services
- Bounded contexts map cleanly to ERP modules (HR, Finance, Timetable, etc.)
- Easy to swap infrastructure (Postgres → Mongo, HTTP → gRPC)
- Testable domain in isolation (no DB needed)

### Cons
- Boilerplate heavy: interfaces + structs per layer
- Over-engineering risk for simple CRUD modules
- Go lacks generics-friendly DI containers (wire/fx help but add complexity)
- Circular import risk if boundaries aren't strict; use `internal/` enforcement

### Recommended Libs
| Lib | Purpose |
|-----|---------|
| `google/wire` | Compile-time DI |
| `uber-go/fx` | Runtime DI with lifecycle hooks |
| `go-ozzo/validation` | Domain validation |

### Folder Structure for ERP Context
```
/cmd
  /api          # main entry
/internal
  /hr           # bounded context
    /domain
    /application
    /infrastructure
    /delivery
  /finance
  /timetable
  /shared       # shared kernel (value objects, events, errors)
/pkg            # truly reusable, zero domain knowledge
```

---

## 2. CQRS + Event-Driven in Go

### Pattern Overview

CQRS separates **write** (Commands → state mutation) from **read** (Queries → projections/read models). Event sourcing stores events as source of truth; current state is a projection.

```
Command → CommandBus → CommandHandler → Aggregate → DomainEvent
DomainEvent → EventBus → EventHandler → ReadModel / Projector
Query → QueryBus → QueryHandler → ReadModel → Response
```

### Watermill (ThreeDotsLabs) — Primary Recommendation

```go
// Publisher/Subscriber abstraction
pub, _ := kafka.NewPublisher(kafka.PublisherConfig{...}, logger)
sub, _ := kafka.NewSubscriber(kafka.SubscriberConfig{...}, logger)

router, _ := message.NewRouter(message.RouterConfig{}, logger)
router.AddHandler("order-handler",
    "orders.events",      // subscribes from
    sub,
    "orders.processed",   // publishes to
    pub,
    func(msg *message.Message) ([]*message.Message, error) { ... },
)
```

Watermill CQRS component:
```go
cqrsFacade, _ := cqrs.NewFacade(cqrs.FacadeConfig{
    GenerateCommandsTopic:  func(commandName string) string { return "commands." + commandName },
    CommandHandlers:        func(...) []cqrs.CommandHandler { return handlers },
    CommandPublisher:       pub,
    CommandSubscriberConstructor: ...,
    EventHandlers:          func(...) []cqrs.EventHandler { return projectors },
    Router:                 router,
    Marshaler:              cqrs.JSONMarshaler{},
})
```

Supported backends: Kafka, RabbitMQ, NATS, Redis Streams, SQL, in-memory (for tests).

### Event Sourcing Libraries
| Lib | Notes |
|-----|-------|
| `hallgren/eventsourcing` | Lightweight, stores in PG/Redis/EventStore |
| `EventStore DB` | Purpose-built, use via Go client `EventStore/EventStore-Client-Go` |
| `looplab/eventhorizon` | Full ES+CQRS framework, more opinionated |

### Simple In-Process CQRS (no broker)
For smaller ERP modules, a command/event bus can be a simple in-memory dispatcher:
```go
type CommandBus struct {
    handlers map[reflect.Type]CommandHandler
}
func (b *CommandBus) Register(cmd any, h CommandHandler) { ... }
func (b *CommandBus) Dispatch(ctx context.Context, cmd any) error { ... }
```

### Pros
- Read/write scaling independently — projections can be optimized per UI
- Audit trail via events — critical for ERP (finance, HR compliance)
- Decouples modules: timetable → emits `TimetablePublished` → notifications module reacts
- Watermill handles retries, dead-letter, middleware natively

### Cons
- Eventual consistency complicates UX (stale reads)
- Debugging distributed event flows is harder
- Event schema evolution (versioning) needs upfront planning
- Full event sourcing is overkill for most ERP modules; use selectively

### Recommendation for ERP
- Use **watermill** with **in-memory pub/sub in dev**, **NATS or Kafka in prod**
- Apply full event sourcing only for audit-critical aggregates (financial transactions, payroll)
- Simple CRUD modules use command handlers + domain events without event store

---

## 3. Plugin/Module System in Go

### Three Approaches

#### A. Go `plugin` Package (stdlib)
```go
p, _ := plugin.Open("modules/hr.so")
sym, _ := p.Lookup("Module")
mod := sym.(ModuleInterface)
mod.Register(app)
```

Pros: No external dep, native performance.
Cons: **Same Go version + compiler flags required**; no cross-platform (.so Linux only); hot reload breaks in practice; plugin can crash host process. **Not recommended for production ERP.**

#### B. HashiCorp go-plugin (gRPC-based)
Each plugin runs as a **separate OS process**, communicating via gRPC or net/rpc.
```go
// Host
client := plugin.NewClient(&plugin.ClientConfig{
    HandshakeConfig: handshakeConfig,
    Plugins:         pluginMap,
    Cmd:             exec.Command("./plugins/payroll"),
})
rpcClient, _ := client.Client()
raw, _ := rpcClient.Dispense("payroll")
payroll := raw.(PayrollPlugin)
```

Pros: Strong isolation (plugin crash = subprocess crash, not host); cross-language possible; versioned gRPC contracts.
Cons: IPC overhead per call; serialization cost; deployment complexity (distribute binaries).

#### C. Custom Module Registry (Recommended for ERP)
No dynamic loading — modules compiled in, registered at startup:
```go
// shared/module.go
type Module interface {
    Name() string
    Routes() []Route
    Migrate(db *sql.DB) error
    EventHandlers() []EventHandler
}

// registry
type Registry struct { modules []Module }
func (r *Registry) Register(m Module) { r.modules = append(r.modules, m) }
func (r *Registry) Bootstrap(app *App) {
    for _, m := range r.modules { m.Routes(); m.Migrate(app.DB) }
}

// main.go
reg := &Registry{}
reg.Register(hr.NewModule())
reg.Register(finance.NewModule())
reg.Register(timetable.NewModule())
reg.Bootstrap(app)
```

Feature flags control which modules are active per tenant.

Pros: Simple, type-safe, no deployment complexity, easy to test.
Cons: All modules compiled into one binary (monorepo trade-off), not truly dynamic.

### Recommendation
**Use custom module registry** for a Go ERP monolith/modular-monolith. Reserve go-plugin only if you need third-party extensions at runtime. Avoid stdlib `plugin`.

---

## 4. Timetable/Scheduling Algorithms

### Problem Characteristics (Academic Timetabling)
Constraints split into:
- **Hard**: No teacher double-booked, no room double-booked, room capacity met
- **Soft**: Teacher preferences, consecutive lessons minimized, balanced schedule

### Approaches Ranked by Complexity

#### A. Constraint Propagation (simplest, fast for small data)
Use backtracking with constraint propagation (Arc Consistency / AC-3).
```go
type Slot struct { Day, Period int }
type Assignment map[ClassID]Slot

func backtrack(assignment Assignment, unassigned []ClassID, constraints []Constraint) Assignment {
    if len(unassigned) == 0 { return assignment }
    cls := unassigned[0]
    for _, slot := range allSlots {
        if consistent(assignment, cls, slot, constraints) {
            assignment[cls] = slot
            result := backtrack(assignment, unassigned[1:], constraints)
            if result != nil { return result }
            delete(assignment, cls)
        }
    }
    return nil
}
```

Good for: < 50 classes, hard constraints only.

#### B. Greedy + Local Search (ITC2 style, practical sweet spot)
1. Greedy initial assignment (most-constrained first)
2. Iterative improvement: Hill Climbing or Simulated Annealing on soft violations

```go
// Simulated annealing skeleton
temp := initialTemp
for iter := 0; iter < maxIter; iter++ {
    neighbor := mutate(currentSolution)  // swap two slots
    delta := evaluate(neighbor) - evaluate(currentSolution)
    if delta < 0 || rand.Float64() < math.Exp(-delta/temp) {
        currentSolution = neighbor
    }
    temp *= coolingRate  // e.g. 0.995
}
```

Good for: Most real ERP timetabling needs. Handles soft constraints naturally.

#### C. Genetic Algorithm
```
Population → Selection → Crossover → Mutation → Fitness eval → Repeat
```
Good for: Large, complex timetabling with many soft constraints. More tuning needed.

#### D. Dedicated Solver (recommended for production)
- **OptaPlanner** (Java, but REST API exposable) — used by school ERP systems
- **OR-Tools** (Google) — has Go bindings (`google/or-tools` via CGo or REST service)
- **UniTime** — open-source academic timetabling system, can be called as microservice

### Go Implementation Libraries
| Option | Notes |
|--------|-------|
| Pure Go backtracking | Best for simple constraints, < 100 entities |
| `jmcvetta/randutil` + SA | Simulated annealing helper |
| OR-Tools via subprocess | Call Python/Java solver, consume JSON result |
| Expose OptaPlanner as REST | Cleanest separation, Go calls solver service |

### Recommended Approach for Go ERP
1. **Phase 1**: Implement greedy + simulated annealing in pure Go. Handles 80% of school timetabling cases.
2. **Phase 2**: If constraints become too complex, wrap OR-Tools or OptaPlanner as a sidecar service, Go calls via HTTP/gRPC.

Key data structures:
```go
type TimetableProblem struct {
    Teachers  []Teacher
    Rooms     []Room
    Classes   []Class
    Slots     []TimeSlot  // Day × Period combinations
    Constraints []Constraint
}

type Solution struct {
    Assignments map[ClassID]Assignment
    HardViolations int
    SoftPenalty    float64
}
```

---

## 5. AI Agent Integration Patterns in Go

### Integration Patterns

#### A. Direct LLM API Call (simplest)
```go
// Using OpenAI Go SDK or Anthropic Go client
resp, _ := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
    Model:    openai.F(openai.ChatModelGPT4o),
    Messages: openai.F(messages),
})
```

Use for: Single-shot tasks, summarization, classification embedded in service handlers.

#### B. Tool Calling / Function Calling (practical agent pattern)
LLM decides which Go functions to call; Go executes them and feeds results back:
```go
tools := []openai.ChatCompletionToolParam{
    {
        Type: openai.F(openai.ChatCompletionToolTypeFunction),
        Function: openai.F(openai.FunctionDefinitionParam{
            Name:        openai.F("get_student_schedule"),
            Description: openai.F("Returns the timetable for a student"),
            Parameters:  openai.F(studentScheduleSchema),
        }),
    },
}

// In response loop:
for _, tc := range msg.ToolCalls {
    result := dispatchTool(tc.Function.Name, tc.Function.Arguments)
    messages = append(messages, toolResultMsg(tc.ID, result))
}
```

Use for: Conversational ERP assistant (query HR data, generate reports, check timetables).

#### C. LangChainGo (`tmc/langchaingo`)
Full agent framework port of LangChain to Go:
```go
llm, _ := openai.New()
tools := []tools.Tool{
    googlesearch.New(), // or custom ERP tools
}
agent := agents.NewOneShotZeroShotAgent(llm, tools, agents.WithMaxIterations(3))
result, _ := agents.Run(ctx, agent, "What classes does teacher Smith have tomorrow?")
```

Supports: OpenAI, Anthropic, Ollama, Google Gemini, HuggingFace.

#### D. MCP (Model Context Protocol) Server
Expose ERP data/tools as an MCP server that any MCP-compatible client can call:
```go
// Go MCP SDK (mark3labs/mcp-go)
server := mcp.NewServer("erp-mcp", "1.0.0")
server.AddTool(mcp.NewTool("get_timetable",
    mcp.WithDescription("Get timetable for a class"),
    mcp.WithString("class_id", mcp.Required()),
), getTimetableHandler)
server.Serve()
```

Use for: Making ERP an AI-accessible service for Claude Desktop, Cursor, etc.

#### E. Background Agent Workers (async)
```
HTTP Request → Queue job → Agent Worker (goroutine pool) → LLM API → Save result → Webhook/SSE
```
Use for: Long-running AI tasks (schedule optimization, report generation) that shouldn't block HTTP.

### Go Libraries Summary
| Lib | Stars | Use Case |
|-----|-------|----------|
| `tmc/langchaingo` | ~5k | Agent chains, RAG, tool use |
| `mark3labs/mcp-go` | ~2k | MCP server/client |
| `openai-go` (official) | ~1k | Direct OpenAI API |
| `anthropic-sdk-go` (official) | growing | Direct Anthropic API |
| `ollama/ollama` | ~70k | Local LLM, Go library available |

### Pros
- Tool calling is the most robust pattern — deterministic function execution, LLM handles intent
- MCP pattern future-proofs integration with AI tooling ecosystem
- Goroutines make async agent pipelines efficient in Go

### Cons
- LLM latency (1-30s per call) — never do synchronously in request path
- Cost at scale — cache embeddings, use smaller models for classification
- LangChainGo lags behind Python LangChain feature parity
- Hallucination risk for domain-critical operations (payroll, grades) — require confirmation step

### Recommended Pattern for Go ERP
```
User Request
    ↓
HTTP Handler → validate, authorize
    ↓
Enqueue AgentTask (Redis/NATS)
    ↓
AgentWorker goroutine pool
    ↓
Tool-calling loop (OpenAI/Anthropic)
    ↓
ERP tool functions (type-safe Go handlers)
    ↓
Result stored → push via SSE/WebSocket to client
```

---

## Summary Matrix

| Area | Primary Recommendation | Avoid |
|------|----------------------|-------|
| DDD Architecture | Bounded contexts + `wire` DI | Anemic domain model |
| CQRS/Events | Watermill + NATS, events for cross-module | Full ES everywhere |
| Plugin System | Custom module registry | stdlib `plugin` package |
| Timetabling | Greedy + Simulated Annealing (pure Go) | Custom GA (high tuning cost) |
| AI Agents | Tool calling + async worker pool | Sync LLM in HTTP handler |

---

## Reference Projects
- `ThreeDotsLabs/wild-workouts-go-ddd-example` — DDD + CQRS in Go, blog at threedots.tech
- `ThreeDotsLabs/watermill` — event-driven Go
- `hallgren/eventsourcing` — event sourcing Go
- `looplab/eventhorizon` — CQRS+ES framework
- `tmc/langchaingo` — LLM agent framework Go
- `mark3labs/mcp-go` — MCP server in Go
- `google/or-tools` — constraint solver (via service)
- ITC2 benchmark problems — standard academic timetabling test cases

---

## Unresolved Questions

1. **Multi-tenancy model**: Will modules be tenant-isolated at DB level (schema-per-tenant) or row-level? Affects bounded context design significantly.
2. **Event store choice**: Is an external broker (NATS/Kafka) in scope, or must everything be in-process + Postgres?
3. **Timetabling complexity**: How many teachers/rooms/classes per school? Determines if pure Go SA suffices or solver sidecar needed.
4. **AI agent scope**: Is the agent for internal admin use or student-facing? Determines latency tolerance and model choice (cost).
5. **Module hot-reload requirement**: If new modules must be added without restart, go-plugin or microservice split is needed — changes architecture significantly.
