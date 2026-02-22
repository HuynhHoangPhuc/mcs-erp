---
title: "Frontend Implementation"
description: "Build complete React frontend for MCS-ERP: API client, shared UI, shell layout, and 5 module UIs"
status: complete
priority: P1
effort: 32h
branch: main
tags: [frontend, react, typescript, tanstack, shadcn-ui, tailwind]
created: 2026-02-22
---

# Frontend Implementation Plan

## Overview

Build the full React 19 frontend for MCS-ERP. Backend Go API is complete with 6 modules. Frontend currently has: basic shell with login page, auth provider, api-client fetch wrapper. All module packages are empty stubs.

## Architecture

- **Monorepo:** Turborepo + pnpm (existing)
- **Routing:** TanStack Router v1 (layout routes, auth guards)
- **State:** TanStack Query v5 (server state), React context (auth)
- **Tables:** TanStack Table v8 (server-side pagination/sorting)
- **Forms:** react-hook-form + zod + shadcn/ui Form
- **UI:** shadcn/ui + Tailwind CSS v4
- **SSE:** `@microsoft/fetch-event-source` for auth-header support
- **DAG viz:** react-flow + dagre auto-layout

## API Patterns (from backend handlers)

- **List response:** `{ items: T[], total: number }`
- **Pagination:** `?offset=0&limit=20`
- **Single resource:** flat object
- **Errors:** `{ error: string }`
- **SSE chat:** POST body → `text/event-stream` with `data: "token"\n\n` + `data: [DONE]\n\n`

## Phases

| # | Phase | Effort | Status | Link |
|---|-------|--------|--------|------|
| 1 | [Shared Foundation](./phase-01-shared-foundation.md) | 6h | complete | API client hooks, UI package, Tailwind/shadcn setup |
| 2 | [Shell & Layout](./phase-02-shell-and-layout.md) | 4h | complete | Sidebar nav, header, breadcrumbs, route tree |
| 3 | [HR Module](./phase-03-hr-module.md) | 5h | complete | Teachers, departments, availability grid |
| 4 | [Subject Module](./phase-04-subject-module.md) | 5h | complete | Subjects, categories, prerequisite DAG |
| 5 | [Room Module](./phase-05-room-module.md) | 3h | complete | Rooms, availability, equipment tags |
| 6 | [Timetable Module](./phase-06-timetable-module.md) | 5h | complete | Semesters, schedule grid, generation, approval |
| 7 | [Agent Module](./phase-07-agent-module.md) | 4h | complete | Chat SSE, conversations, suggestions |

## Dependencies

```
Phase 1 (Foundation) — API client + UI components
  → Phase 2 (Shell) — layout, routing
    → Phase 3 (HR)     \
    → Phase 4 (Subject)  > independent, can parallelize
    → Phase 5 (Room)   /
      → Phase 6 (Timetable) — needs HR/Subject/Room data
      → Phase 7 (Agent) — independent but last priority
```

## Key Research

- [TanStack Patterns](./research/researcher-01-tanstack-patterns.md)
- [shadcn/SSE/react-flow](./research/researcher-02-shadcn-sse-reactflow.md)

## Key Decisions

1. **Pagination:** offset/limit (matches backend), not page-based
2. **Forms:** react-hook-form + zod (not TanStack Form — more ecosystem support)
3. **SSE:** `@microsoft/fetch-event-source` for POST+Bearer header support
4. **Module exports:** Each module package exports page components only (not route objects)
5. **Tailwind v4:** CSS-first config with `@import "tailwindcss"` — shadcn supports v4+React 19
6. **apiFetch location:** Move from shell to `@mcs-erp/api-client` package (colocated with hooks)
7. **Timetable name resolution:** Parallel queries for teachers/subjects/rooms on page load, build ID→name maps
8. **API base URL:** `VITE_API_BASE_URL` env var (default '' for same-origin production)
9. **Markdown rendering:** `react-markdown` for AI chat responses (handles code blocks, tables)
10. **Scope:** All 7 phases — full frontend implementation

## Validation Log

### Session 1 — 2026-02-22
**Trigger:** Initial plan creation validation
**Questions asked:** 7

#### Questions & Answers

1. **[Architecture]** The plan uses Tailwind CSS v4 (CSS-first config, no tailwind.config.js). shadcn/ui CLI may not fully support v4 yet, requiring manual component creation. Should we use Tailwind v4 or stick with the battle-tested v3?
   - Options: Tailwind v3 (Recommended) | Tailwind v4
   - **Answer:** Tailwind v4
   - **Custom input:** use v4, I see shadcn newest version support tailwind v4 and react 19 https://ui.shadcn.com/docs/tailwind-v4
   - **Rationale:** shadcn/ui now officially supports Tailwind v4 + React 19. No compatibility risk.

2. **[Architecture]** The API client hooks are placed in `@mcs-erp/api-client` package, but `apiFetch` (with auth token management) lives in `apps/shell/src/lib/api-client.ts`. How should hooks access the fetch wrapper?
   - Options: Move apiFetch to api-client package (Recommended) | Keep in shell via QueryClient | Keep in shell via import
   - **Answer:** Move apiFetch to api-client package
   - **Rationale:** Colocates fetch wrapper with hooks. Clean dependency graph. Shell only provides QueryClientProvider.

3. **[Scope]** The plan creates ~50 component files across 7 packages. For an MVP, should we implement all modules or focus on a critical subset first?
   - Options: All 7 phases (full plan) | Phases 1-3+6 | Phases 1-2+3-5
   - **Answer:** All 7 phases (full plan)
   - **Rationale:** Complete frontend wanted. Every module gets UI.

4. **[Architecture]** Phase 07 (Agent) plans a custom lightweight markdown renderer (regex-based). For rendering AI responses with code blocks, should we use react-markdown instead?
   - Options: react-markdown (Recommended) | Custom regex renderer | Defer Agent module
   - **Answer:** react-markdown
   - **Rationale:** Battle-tested, handles edge cases (nested lists, code blocks, tables). ~30KB acceptable.

5. **[Architecture]** The timetable grid shows assignments as day × period cells. Each cell needs teacher name, subject code, and room code — but the schedule API only returns IDs. How should we resolve display names?
   - Options: Parallel queries on page load (Recommended) | Backend enriched response | Fetch on-demand per cell
   - **Answer:** Parallel queries on page load
   - **Rationale:** Simple client-side ID→name maps. No backend changes. Works with existing API.

6. **[Architecture]** Phase 02 creates stub route files in the shell that import from module packages. How should module packages expose their page components?
   - Options: Export page components only (Recommended) | Export route definitions | Export both
   - **Answer:** Export page components only
   - **Rationale:** Clean separation. Shell owns routing. Modules own UI. No router coupling in packages.

7. **[Architecture]** The plan assumes the existing `apiFetch` base path `/api/v1` is correct and Vite proxy handles dev routing. Should we also add CORS headers handling or an env-based API URL for production?
   - Options: Env-based API URL (Recommended) | Vite proxy only | Both
   - **Answer:** Env-based API URL
   - **Rationale:** `VITE_API_BASE_URL` env var for flexibility. Default '' (same-origin) in production.

#### Confirmed Decisions
- **Tailwind v4:** Confirmed — shadcn officially supports it now
- **apiFetch location:** Move to `@mcs-erp/api-client` package
- **Scope:** Full 7-phase implementation
- **Markdown:** react-markdown for AI chat
- **Name resolution:** Parallel queries, client-side ID→name maps
- **Module exports:** Page components only, no route objects
- **API URL:** VITE_API_BASE_URL env var

#### Action Items
- [ ] Phase 01: Move `apiFetch` from `apps/shell/src/lib/api-client.ts` to `packages/api-client/src/lib/api-client.ts`
- [ ] Phase 01: Add `VITE_API_BASE_URL` support in apiFetch base URL
- [ ] Phase 01: Use shadcn v4 setup guide (https://ui.shadcn.com/docs/tailwind-v4)
- [ ] Phase 04: Update module export pattern — page components only
- [ ] Phase 06: Add parallel queries for teacher/subject/room name resolution
- [ ] Phase 07: Add `react-markdown` dependency instead of custom renderer

#### Impact on Phases
- Phase 01: Move apiFetch to api-client package. Add VITE_API_BASE_URL env var. Follow shadcn v4 setup guide.
- Phase 02: Shell route files import page components (not route objects) from module packages.
- Phase 06: Timetable detail page fetches teachers/subjects/rooms lists in parallel, builds lookup maps.
- Phase 07: Replace custom markdown renderer with react-markdown. Add react-markdown as dependency.
