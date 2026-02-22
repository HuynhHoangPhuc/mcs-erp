# Phase 04: Subject Module

## Context
- [Plan](./plan.md) | [Phase 03](./phase-03-hr-module.md)
- Depends on: Phase 01, Phase 02
- Backend routes: `GET/POST /api/v1/subjects`, `GET/PUT /api/v1/subjects/{id}`, `GET/POST /api/v1/categories`, `POST/DELETE /api/v1/subjects/{id}/prerequisites`, `GET /api/v1/subjects/{id}/prerequisites`, `GET /api/v1/subjects/{id}/prerequisite-chain`

## Overview
- **Priority:** P2
- **Status:** pending
- **Effort:** 5h

Build subject catalog UI with prerequisite DAG visualization using react-flow + dagre auto-layout.

## Key Insights
- Subject has: id, name, code, description, category_id, credits, hours_per_week, is_active
- Prerequisites form a DAG — backend validates no cycles on add
- Prerequisite chain endpoint returns transitive dependencies
- react-flow v12 with dagre for automatic top-down layout

## Requirements

### Functional
- Subject list with filtering (category, search) + pagination
- Subject create/edit form (name, code, description, category, credits, hours_per_week)
- Category management (simple CRUD list)
- Prerequisite DAG visualization page:
  - Nodes = subjects (show name + code)
  - Edges = prerequisite relationships (directed)
  - Auto-layout using dagre (top-to-bottom)
  - Add prerequisite: select subject dropdown → add edge
  - Remove prerequisite: click edge → confirm delete
  - Highlight prerequisite chain for selected subject

### Non-functional
- DAG re-layouts on add/remove without losing scroll position
- Backend cycle detection error displayed as toast

## Architecture

### File Structure
```
packages/module-subject/src/
├── index.ts
├── components/
│   ├── subject-list-page.tsx
│   ├── subject-form-dialog.tsx
│   ├── subject-columns.tsx
│   ├── subject-detail-page.tsx
│   ├── category-list-page.tsx
│   ├── category-form-dialog.tsx
│   ├── prerequisite-dag-page.tsx     # react-flow DAG visualization
│   ├── subject-node.tsx              # Custom react-flow node
│   └── add-prerequisite-dialog.tsx   # Dialog to add prerequisite edge
```

## Related Code Files

### Modify
- `web/packages/module-subject/package.json` — add deps (+ `@xyflow/react`, `@dagrejs/dagre`)
- `web/packages/module-subject/src/index.ts`
- Shell route stubs: `subjects.index.tsx`, `subjects.$subjectId.tsx`, `subjects.prerequisites.tsx`, `categories.index.tsx`

### Create
- All files under `packages/module-subject/src/components/`

## Implementation Steps

1. **Add deps**: `@mcs-erp/ui`, `@mcs-erp/api-client`, `@tanstack/react-table`, `@tanstack/react-query`, `react-hook-form`, `@hookform/resolvers`, `zod`, `@xyflow/react`, `@dagrejs/dagre`

2. **Create subject CRUD components** — Same pattern as HR: columns, form dialog, list page, detail page

3. **Create category CRUD** — Simple DataTable with create/edit/delete dialog

4. **Create `subject-node.tsx`** — Custom react-flow node showing subject name, code, credits as a card

5. **Create `prerequisite-dag-page.tsx`**:
   - Fetch all subjects + all prerequisites
   - Build nodes + edges arrays for react-flow
   - Apply dagre layout (TB direction)
   - `onConnect` handler → call `addPrerequisite` mutation → re-layout
   - Edge click → confirm → call `deletePrerequisite` mutation
   - On node click → highlight prerequisite chain (fetch from API)
   - Error toast if cycle detected (backend returns error)

6. **Create `add-prerequisite-dialog.tsx`** — Two select dropdowns: subject + prerequisite subject → submit

7. **Wire route pages** in shell

8. **Build check**

## Todo List
- [ ] Add dependencies (react-flow, dagre)
- [ ] Create subject columns + list page
- [ ] Create subject form dialog
- [ ] Create subject detail page
- [ ] Create category CRUD
- [ ] Create custom react-flow SubjectNode
- [ ] Create prerequisite DAG page with dagre layout
- [ ] Create add-prerequisite dialog
- [ ] Handle cycle detection errors
- [ ] Wire routes in shell
- [ ] Verify build passes

## Success Criteria
- Subject list with pagination and category filter
- Create/edit subject works
- DAG page renders all subjects as nodes with prerequisite edges
- Adding prerequisite re-layouts graph
- Cycle detection error displays as toast
- Removing prerequisite edge works

## Risk Assessment
- **react-flow bundle size** — ~150KB; acceptable for ERP
- **Large DAG rendering** — dagre handles 100+ nodes fine; beyond 500 may need virtualization

## Security Considerations
- Prerequisite mutations require `subject:subject:write` permission
- Backend handles optimistic locking; frontend shows retry toast on version conflict

## Next Steps
- Phase 05 (Room) — simpler CRUD, reuses availability grid from Phase 01
