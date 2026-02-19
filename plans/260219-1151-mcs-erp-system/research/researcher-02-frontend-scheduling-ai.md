# Research Report: Frontend (TanStack), Scheduling (SA), AI Tool-Calling (Go)
Date: 2026-02-19 | Sources: Training knowledge (cutoff Aug 2025)

---

## Topic 1: TanStack Router + Turborepo Monorepo with Lazy Module Routes

### Turborepo + TanStack Router Setup

**Recommended monorepo structure:**
```
apps/
  shell/               # Main host app (Vite + React)
packages/
  module-hr/           # HR module (routes, components, queries)
  module-subject/      # Subject/curriculum module
  module-timetable/    # Timetable module
  shared-ui/           # Shared components
  api-client/          # Generated API types + query hooks
```

**File-based routing in shell app** (`apps/shell/src/routes/`):
- TanStack Router v1 file-based routing uses `@tanstack/router-vite-plugin`
- Route files: `routes/hr/index.tsx`, `routes/hr/$id.tsx`
- Module routes are **not** file-based from packages; instead, use **virtual routes** or **route registration**

### Lazy Route Registration Pattern

Each module exports a route tree subtree; shell merges at startup:

```ts
// packages/module-hr/src/routes.ts
import { createRoute, lazyRouteComponent } from '@tanstack/react-router'

export function createHrRoutes(parentRoute: AnyRoute) {
  const hrRoot = createRoute({
    getParentRoute: () => parentRoute,
    path: '/hr',
    component: lazyRouteComponent(() => import('./layouts/HrLayout')),
  })
  const hrList = createRoute({
    getParentRoute: () => hrRoot,
    path: '/',
    component: lazyRouteComponent(() => import('./pages/HrList')),
  })
  return hrRoot.addChildren([hrList])
}
```

```ts
// apps/shell/src/router.ts
import { createRouter, createRootRoute } from '@tanstack/react-router'
import { createHrRoutes } from 'module-hr'
import { createSubjectRoutes } from 'module-subject'

const rootRoute = createRootRoute({ component: RootLayout })
const routeTree = rootRoute.addChildren([
  createHrRoutes(rootRoute),
  createSubjectRoutes(rootRoute),
])
export const router = createRouter({ routeTree })
```

**Key:** `lazyRouteComponent` enables code-splitting per module. Turborepo build cache ensures modules rebuild only when changed.

### TanStack Query Integration per Module

Each module owns its query keys and hooks:
```ts
// packages/module-hr/src/queries/use-teachers.ts
import { useQuery } from '@tanstack/react-query'
import { apiClient } from 'api-client'

export const teacherKeys = { all: ['teachers'] as const }
export function useTeachers() {
  return useQuery({ queryKey: teacherKeys.all, queryFn: apiClient.getTeachers })
}
```

Shell provides `QueryClientProvider` once at root; modules consume it via hooks. No cross-module query sharing needed.

**Turborepo `turbo.json` pipeline:**
```json
{
  "pipeline": {
    "build": { "dependsOn": ["^build"], "outputs": ["dist/**"] },
    "dev": { "cache": false, "persistent": true }
  }
}
```

---

## Topic 2: Simulated Annealing for Academic Timetabling (Go)

### Data Model

```go
type TimeSlot struct { Day, Period int }
type Assignment struct {
    ClassID, TeacherID, RoomID int
    Slot TimeSlot
}
type Schedule []Assignment
```

### Constraint Classification

**Hard constraints (must = 0 violations):**
- No teacher double-booked in same slot
- Room capacity >= class size
- Teacher unavailability (blocked slots)
- No class assigned twice in same slot

**Soft constraints (minimize penalty):**
- Teacher preference for morning slots
- Room type match (lab vs classroom)
- Minimize gaps in teacher schedule
- Balanced load per day

```go
func EvaluatePenalty(s Schedule, hard, soft []Constraint) (int, int) {
    h, sft := 0, 0
    for _, c := range hard { h += c.Violations(s) }
    for _, c := range soft { sft += c.Violations(s) }
    return h, sft
}
// Cost = hardViolations*10000 + softViolations
```

### Neighborhood Functions

```go
// Swap: exchange slots of two assignments
func SwapMove(s Schedule) Schedule { /* pick 2 random, swap slots */ }

// Move: reassign one class to different slot
func MoveSlot(s Schedule, slots []TimeSlot) Schedule { /* pick 1, new slot */ }

// Reassign room: keep slot, change room
func ReassignRoom(s Schedule, rooms []Room) Schedule { /* pick 1, new room */ }
```

Apply randomly weighted: Swap 50%, MoveSlot 35%, ReassignRoom 15%.

### SA Core Loop (Go)

```go
func Anneal(initial Schedule, cfg SAConfig) Schedule {
    current := initial
    best := initial
    T := cfg.TInitial // e.g. 1000.0

    for iter := 0; iter < cfg.MaxIter; iter++ {
        neighbor := randomNeighbor(current)
        delta := cost(neighbor) - cost(current)
        if delta < 0 || rand.Float64() < math.Exp(-float64(delta)/T) {
            current = neighbor
            if cost(current) < cost(best) { best = current }
        }
        T *= cfg.CoolingRate // e.g. 0.9995
        if T < cfg.TMin { break } // e.g. 0.01
    }
    return best
}
```

### Cooling Schedule Tuning

| Parameter | Typical Value | Notes |
|---|---|---|
| T_initial | 500–2000 | Accept ~80% bad moves at start |
| CoolingRate | 0.995–0.9999 | Slower = better quality, longer |
| T_min | 0.01–1.0 | Stop threshold |
| MaxIter | 100k–1M | Scale with problem size |

**Convergence detection:** Track best cost over last N=500 iters; stop if no improvement.

**Room as extra dimension:** Include RoomID in Assignment; add hard constraint checking room conflicts separately from teacher conflicts. Same SA loop handles both.

**Parallelism:** Run 4–8 independent SA chains (goroutines), take global best — cheap in Go.

```go
func ParallelAnneal(initial Schedule, cfg SAConfig, n int) Schedule {
    results := make(chan Schedule, n)
    for i := 0; i < n; i++ {
        go func() { results <- Anneal(shuffle(initial), cfg) }()
    }
    // collect best
}
```

---

## Topic 3: Multi-Provider LLM Tool-Calling in Go

### langchaingo vs Custom Abstraction

| | langchaingo | Custom thin wrapper |
|---|---|---|
| Pros | Provider adapters ready, tools interface defined | Full control, no dep bloat |
| Cons | Abstraction leaks, slow to adopt new APIs, heavy | Must implement each provider |
| Verdict | **Use langchaingo** for MVP, replace hot paths later |

langchaingo supports Claude (Anthropic), OpenAI, Ollama out of box as of 2024.

### Tool Definition Pattern (langchaingo)

```go
import "github.com/tmc/langchaingo/tools"

type ScheduleQueryTool struct{ db *DB }

func (t *ScheduleQueryTool) Name() string { return "query_schedule" }
func (t *ScheduleQueryTool) Description() string {
    return "Query current timetable. Input: {teacher_id, week}"
}
func (t *ScheduleQueryTool) Call(ctx context.Context, input string) (string, error) {
    // parse input JSON, query DB, return JSON string
}
```

**Module tool registration at startup:**
```go
// Each module exposes RegisterTools(registry *ToolRegistry)
type ToolRegistry struct { tools []tools.Tool }
func (r *ToolRegistry) Register(t tools.Tool) { r.tools = append(r.tools, t) }

// main.go
registry := &ToolRegistry{}
hrModule.RegisterTools(registry)
timetableModule.RegisterTools(registry)
agent := agents.NewOneShotAgent(llm, registry.tools, ...)
```

### SSE Streaming Pattern (Go net/http)

```go
func ChatHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    flusher := w.(http.Flusher)

    streamCh := make(chan string)
    go func() {
        // langchaingo streaming callback populates streamCh
        llm.GenerateContent(r.Context(), messages,
            llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
                streamCh <- string(chunk)
                return nil
            }),
        )
        close(streamCh)
    }()

    for token := range streamCh {
        fmt.Fprintf(w, "data: %s\n\n", token)
        flusher.Flush()
    }
    fmt.Fprintf(w, "data: [DONE]\n\n")
    flusher.Flush()
}
```

### Multi-Provider Config

```go
type LLMProvider string
const (
    ProviderClaude  LLMProvider = "claude"
    ProviderOpenAI  LLMProvider = "openai"
    ProviderOllama  LLMProvider = "ollama"
)

func NewLLM(cfg Config) llms.Model {
    switch cfg.Provider {
    case ProviderClaude:
        return anthropic.New(anthropic.WithModel(cfg.Model))
    case ProviderOpenAI:
        return openai.New(openai.WithModel(cfg.Model))
    case ProviderOllama:
        return ollama.New(ollama.WithModel(cfg.Model))
    }
}
```

Tool dispatch (function calling) is handled by langchaingo's agent executor — it parses LLM tool_call response, calls the matching tool, feeds result back automatically.

---

## Summary Recommendations

| Topic | Decision |
|---|---|
| Monorepo routing | Shell owns router; modules export `createXRoutes(parent)` factory; `lazyRouteComponent` for code-split |
| SA timetabling | Go goroutine-parallel SA, cost = hard×10000 + soft, 3 neighbor ops, cooling 0.9995 |
| AI Go stack | langchaingo for MVP; modules register tools at startup; SSE via `http.Flusher` |

---

## Unresolved Questions

1. TanStack Router v2 (if released post Aug 2025) may change file-based API — verify current version.
2. SA convergence time for large schedules (>500 classes) — may need reheating strategy.
3. langchaingo Claude tool-calling: verify Anthropic tool_use format is fully supported (some versions had gaps).
4. Ollama streaming + tool-calling support varies by model — test with specific models used.
