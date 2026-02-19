# Phase 05: Subject Module

## Context Links
- [Parent Plan](./plan.md)
- Depends on: [Phase 03 - Auth & RBAC](./phase-03-auth-rbac.md)
- [Brainstorm Report](../reports/brainstorm-260219-1151-mcs-erp-architecture.md)

## Overview
- **Date:** 2026-02-19
- **Priority:** P1
- **Status:** Complete
- **Description:** Subject bounded context: Subject entity, Category grouping, prerequisite DAG with cycle detection (DFS), topological sort for semester ordering. Frontend: subject list, subject form, DAG visualization. AI tools: search_subjects, get_prerequisites.

## Key Insights
- Prerequisites form a DAG (Directed Acyclic Graph) — must reject cycles on edge creation
- Cycle detection via DFS with coloring (white/gray/black) on AddPrerequisite
- Topological sort (Kahn's) provides valid teaching order — used by timetable scheduling
- Credits/hours-per-week field drives timetable slot allocation
- Keep graph operations in domain layer (pure Go, no DB dependency) for testability

## Requirements

### Functional
- Subject CRUD: name, code, description, category, credits, hours_per_week, status
- Category CRUD: name, description (grouping subjects like "Science", "Languages")
- Prerequisite management: add/remove prerequisite edges between subjects
- Cycle detection: reject prerequisite if it would create a cycle
- Topological sort: return valid teaching order for a set of subjects
- Prerequisite chain query: "What must be completed before Subject X?"
- Frontend: subject list, form, DAG visualization with react-flow

### Non-Functional
- Cycle detection runs in O(V+E) where V=subjects, E=prerequisite edges
- DAG visualization handles up to 100 nodes smoothly
- Subject code must be unique per tenant

## Architecture

```
internal/subject/
├── domain/
│   ├── subject.go              # Subject aggregate root
│   ├── category.go             # Category entity
│   ├── prerequisite.go         # PrerequisiteEdge value object
│   ├── prerequisite_graph.go   # DAG: cycle detection, topo sort
│   ├── events.go               # SubjectCreated, PrerequisiteAdded
│   └── repository.go           # SubjectRepository, CategoryRepository
├── application/
│   ├── commands/
│   │   ├── create_subject.go
│   │   ├── update_subject.go
│   │   ├── add_prerequisite.go
│   │   ├── remove_prerequisite.go
│   │   └── create_category.go
│   ├── queries/
│   │   ├── list_subjects.go
│   │   ├── get_subject.go
│   │   ├── get_prerequisite_graph.go
│   │   └── get_prerequisite_chain.go
│   └── events/
│       └── handlers.go
├── infrastructure/
│   ├── postgres_subject_repo.go
│   └── postgres_category_repo.go
├── delivery/
│   ├── subject_handler.go      # /api/subjects
│   ├── category_handler.go     # /api/subjects/categories
│   └── prerequisite_handler.go # /api/subjects/:id/prerequisites
├── tools/
│   ├── search_subjects.go      # AI tool
│   └── get_prerequisites.go    # AI tool
└── module.go
```

### Frontend
```
web/packages/module-subject/src/
├── routes.ts
├── pages/
│   ├── subject-list-page.tsx
│   ├── subject-detail-page.tsx
│   ├── subject-form-page.tsx
│   ├── category-list-page.tsx
│   └── prerequisite-graph-page.tsx
├── components/
│   ├── subject-table.tsx
│   ├── subject-form.tsx
│   ├── category-select.tsx
│   ├── prerequisite-select.tsx  # Add prerequisite dropdown
│   └── dag-visualization.tsx    # react-flow DAG viewer
└── queries/
    ├── use-subjects.ts
    ├── use-subject.ts
    ├── use-categories.ts
    └── use-prerequisite-graph.ts
```

## Related Code Files

### Files to Create

**Backend Domain:**
- `internal/subject/domain/subject.go` — Subject: ID, Code, Name, Description, CategoryID, Credits int, HoursPerWeek int, IsActive, CreatedAt, UpdatedAt
- `internal/subject/domain/category.go` — Category: ID, Name, Description
- `internal/subject/domain/prerequisite.go` — PrerequisiteEdge: SubjectID, PrerequisiteID (the dep)
- `internal/subject/domain/prerequisite_graph.go` — `Graph` struct, `AddEdge(from, to) error` (cycle check), `TopologicalSort() ([]ID, error)`, `GetChain(subjectID) []ID`
- `internal/subject/domain/events.go` — SubjectCreated, SubjectUpdated, PrerequisiteAdded, PrerequisiteRemoved
- `internal/subject/domain/repository.go` — SubjectRepository, CategoryRepository, PrerequisiteRepository

**Backend Application:**
- `internal/subject/application/commands/create_subject.go`
- `internal/subject/application/commands/update_subject.go`
- `internal/subject/application/commands/add_prerequisite.go` — loads graph, validates no cycle, saves edge
- `internal/subject/application/commands/remove_prerequisite.go`
- `internal/subject/application/commands/create_category.go`
- `internal/subject/application/queries/list_subjects.go` — filter by category, status, search
- `internal/subject/application/queries/get_subject.go` — includes prerequisites list
- `internal/subject/application/queries/get_prerequisite_graph.go` — returns nodes + edges for visualization
- `internal/subject/application/queries/get_prerequisite_chain.go` — returns ordered chain for a subject

**Backend Infrastructure:**
- `internal/subject/infrastructure/postgres_subject_repo.go`
- `internal/subject/infrastructure/postgres_category_repo.go`

**Backend Delivery:**
- `internal/subject/delivery/subject_handler.go` — CRUD /api/subjects
- `internal/subject/delivery/category_handler.go` — CRUD /api/subjects/categories
- `internal/subject/delivery/prerequisite_handler.go` — POST/DELETE /api/subjects/:id/prerequisites, GET /api/subjects/graph

**AI Tools:**
- `internal/subject/tools/search_subjects.go`
- `internal/subject/tools/get_prerequisites.go`

**Module:**
- `internal/subject/module.go` — dependencies: `["core"]`

**SQL:**
- `sqlc/queries/subject/subjects.sql`
- `sqlc/queries/subject/categories.sql`
- `sqlc/queries/subject/prerequisites.sql`
- `migrations/subject/000001_create_categories_table.up.sql`
- `migrations/subject/000002_create_subjects_table.up.sql`
- `migrations/subject/000003_create_subject_prerequisites_table.up.sql`

**Frontend:**
- `web/packages/module-subject/src/routes.ts`
- `web/packages/module-subject/src/pages/subject-list-page.tsx`
- `web/packages/module-subject/src/pages/subject-detail-page.tsx`
- `web/packages/module-subject/src/pages/subject-form-page.tsx`
- `web/packages/module-subject/src/pages/category-list-page.tsx`
- `web/packages/module-subject/src/pages/prerequisite-graph-page.tsx`
- `web/packages/module-subject/src/components/subject-table.tsx`
- `web/packages/module-subject/src/components/subject-form.tsx`
- `web/packages/module-subject/src/components/category-select.tsx`
- `web/packages/module-subject/src/components/prerequisite-select.tsx`
- `web/packages/module-subject/src/components/dag-visualization.tsx`
- `web/packages/module-subject/src/queries/use-subjects.ts`
- `web/packages/module-subject/src/queries/use-subject.ts`
- `web/packages/module-subject/src/queries/use-categories.ts`
- `web/packages/module-subject/src/queries/use-prerequisite-graph.ts`

## Implementation Steps

1. **SQL migrations**
   ```sql
   -- categories
   CREATE TABLE categories (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       name VARCHAR(100) NOT NULL UNIQUE,
       description TEXT,
       created_at TIMESTAMPTZ NOT NULL DEFAULT now()
   );
   -- subjects
   CREATE TABLE subjects (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       code VARCHAR(20) NOT NULL UNIQUE,
       name VARCHAR(255) NOT NULL,
       description TEXT,
       category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
       credits INT NOT NULL DEFAULT 1,
       hours_per_week INT NOT NULL DEFAULT 1,
       is_active BOOLEAN NOT NULL DEFAULT true,
       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
       updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
   );
   -- subject_prerequisites (DAG edges, with version for optimistic locking)
   CREATE TABLE subject_prerequisites (
       subject_id UUID REFERENCES subjects(id) ON DELETE CASCADE,
       prerequisite_id UUID REFERENCES subjects(id) ON DELETE CASCADE,
       PRIMARY KEY (subject_id, prerequisite_id),
       CHECK (subject_id != prerequisite_id)
   );
   -- graph version for optimistic locking on concurrent edits
   CREATE TABLE subject_prerequisite_graph_version (
       id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
       version INT NOT NULL DEFAULT 0,
       updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
   );
   INSERT INTO subject_prerequisite_graph_version (id, version) VALUES (1, 0);
<!-- Updated: Validation Session 2 - Optimistic locking with version for DAG concurrency -->
   ```

2. **Domain: prerequisite_graph.go** — core algorithm, pure Go, no DB
   ```go
   type Graph struct {
       adjacency map[uuid.UUID][]uuid.UUID // subject -> prerequisites
   }
   func NewGraph(edges []PrerequisiteEdge) *Graph
   func (g *Graph) AddEdge(subjectID, prereqID uuid.UUID) error // returns ErrCyclicDependency
   func (g *Graph) HasCycle() bool // DFS with white/gray/black coloring
   func (g *Graph) TopologicalSort() ([]uuid.UUID, error) // Kahn's algorithm
   func (g *Graph) GetChain(subjectID uuid.UUID) []uuid.UUID // BFS/DFS all ancestors
   ```

3. **Cycle detection algorithm** — DFS with 3 colors:
   - White (unvisited), Gray (in current path), Black (fully processed)
   - If DFS hits a Gray node, cycle exists
   - Run after tentatively adding edge; rollback if cycle detected

4. **Domain entities** — Subject with `NewSubject(code, name, categoryID, credits, hoursPerWeek)`

5. **sqlc queries** — CRUD + prerequisite edge management + graph query (all edges for tenant)

6. **Command: add_prerequisite** — read current graph version, load all edges from DB, build Graph, add edge, check cycle, save edge + increment version with WHERE version = current_version (optimistic lock). Retry on version conflict (max 3 retries).
<!-- Updated: Validation Session 2 - Optimistic locking replaces advisory lock for DAG concurrency -->

7. **Query: get_prerequisite_graph** — return `{nodes: [{id, name, code}], edges: [{from, to}]}` for frontend

8. **HTTP handlers** — standard CRUD + `POST /api/subjects/:id/prerequisites` body `{prerequisite_id}`, `GET /api/subjects/graph`

9. **AI tools** — `search_subjects(query, category)` returns list, `get_prerequisites(subject_id)` returns chain

10. **Frontend: DAG visualization** — use `@xyflow/react` (react-flow):
    - Nodes = subjects, edges = prerequisites
    - Auto-layout with dagre (top-to-bottom)
    - Click node to navigate to subject detail
    - Highlight prerequisite chain on hover

11. **Frontend: subject form** — fields: code, name, description, category (select), credits, hours_per_week. Prerequisite management on detail page (separate UI).

12. **Frontend: prerequisite select** — dropdown of available subjects (exclude self + would-create-cycle) with add/remove buttons

## Todo List
- [x] SQL migrations (categories, subjects, prerequisites)
- [x] Domain entities (Subject, Category, PrerequisiteEdge)
- [x] Prerequisite graph with cycle detection (DFS)
- [x] Topological sort (Kahn's algorithm)
- [x] Domain events
- [x] Repository interfaces
- [x] sqlc queries
- [x] Infrastructure repos
- [x] Command handlers (CRUD + add/remove prerequisite)
- [x] Query handlers (list, get, graph, chain)
- [x] HTTP handlers
- [x] AI tools (search_subjects, get_prerequisites)
- [x] Subject module registration
- [x] Frontend routes + pages
- [x] Subject table + form components
- [x] DAG visualization with react-flow
- [x] Prerequisite management UI
- [x] TanStack Query hooks
- [x] Optimistic locking: graph version table + retry logic in add_prerequisite command
- [x] Unit tests: cycle detection, topo sort, optimistic lock retry (critical)
- [x] Integration test: prerequisite CRUD + concurrent edit conflict

## Success Criteria
- Subject CRUD works via API
- Adding a prerequisite that creates a cycle returns 400 error
- Topological sort returns valid order
- Prerequisite chain returns all transitive dependencies
- Graph endpoint returns correct nodes + edges
- Frontend DAG renders subjects with edges
- Subject list with filtering works

## Risk Assessment
- **Large graphs**: Cycle detection is O(V+E), acceptable for hundreds of subjects
- **Concurrent prerequisite edits**: Two users adding edges simultaneously could bypass cycle check. Mitigate with optimistic locking: version column on graph, check-and-increment atomically, retry on conflict (max 3).
<!-- Updated: Validation Session 2 - Optimistic locking chosen over advisory locks -->
- **react-flow bundle size**: ~150KB gzipped. Acceptable; load only on graph page via lazy import.

## Security Considerations
- All endpoints require `subject:read` / `subject:write` permissions
- Subject code uniqueness enforced at DB level (tenant-scoped via schema)
- Prerequisite modifications logged via domain events for audit

## Next Steps
- Phase 7 (Timetable) uses topological sort + hours_per_week for scheduling
- Phase 8 (AI Agent) uses subject search + prerequisite tools
