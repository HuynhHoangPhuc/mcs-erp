# Phase 01: Shared Foundation

## Context
- [Plan](./plan.md)
- [TanStack Research](./research/researcher-01-tanstack-patterns.md)
- [shadcn Research](./research/researcher-02-shadcn-sse-reactflow.md)
- [Codebase Summary](../../docs/codebase-summary.md)

## Overview
- **Priority:** P1 (blocks all other phases)
- **Status:** pending
- **Effort:** 6h

Install dependencies, configure Tailwind v4 + shadcn/ui in `packages/ui`, build API client with TanStack Query hooks, and create shared data table / form / layout components.

## Key Insights
- Backend uses `{ items: T[], total: number }` list envelope with `?offset=&limit=` pagination
- All API paths prefixed `/api/v1/` — existing `apiFetch` wrapper handles auth + refresh
- **Move `apiFetch` from shell to `@mcs-erp/api-client`** — colocate with hooks (Validation Session 1)
- **Add `VITE_API_BASE_URL` env var** — default '' for same-origin production (Validation Session 1)
- shadcn/ui in monorepo: export CSS + components from `@mcs-erp/ui`, import in shell app
- Tailwind v4: CSS-first config (`@import "tailwindcss"`) — shadcn officially supports v4+React 19
- Follow shadcn v4 setup: https://ui.shadcn.com/docs/tailwind-v4
<!-- Updated: Validation Session 1 - Move apiFetch to api-client, add VITE_API_BASE_URL, confirm Tailwind v4 -->

## Requirements

### Functional
- TypeScript types matching all backend API response shapes
- TanStack Query hooks for every REST endpoint (CRUD pattern)
- Reusable `DataTable` component with server-side pagination
- Form components using react-hook-form + zod
- SSE utility hook for streaming endpoints

### Non-functional
- All packages must build with `tsc -b`
- Exports must be tree-shakable (named exports, no barrel re-exports of everything)

## Architecture

### API Client (`@mcs-erp/api-client`)
```
packages/api-client/src/
├── index.ts              # Re-export all
├── lib/
│   └── api-client.ts     # apiFetch wrapper (moved from shell) + VITE_API_BASE_URL
├── types/
│   ├── common.ts         # ListResponse<T>, PaginationParams, ApiError
│   ├── auth.ts           # LoginRequest, TokenResponse, User
│   ├── hr.ts             # Teacher, Department, Availability, filters
│   ├── subject.ts        # Subject, Category, Prerequisite
│   ├── room.ts           # Room, RoomAvailability
│   ├── timetable.ts      # Semester, Schedule, Assignment
│   └── agent.ts          # Conversation, Message, ChatRequest
├── hooks/
│   ├── use-auth.ts       # useLogin, useLogout, useRefresh
│   ├── use-teachers.ts   # useTeachers, useTeacher, useCreateTeacher, useUpdateTeacher
│   ├── use-departments.ts
│   ├── use-subjects.ts
│   ├── use-categories.ts
│   ├── use-rooms.ts
│   ├── use-semesters.ts
│   ├── use-schedule.ts
│   ├── use-conversations.ts
│   └── use-chat-sse.ts   # SSE streaming hook
└── query-keys.ts         # Centralized query key factory
```

### UI Package (`@mcs-erp/ui`)
```
packages/ui/src/
├── index.ts
├── globals.css           # Tailwind v4 + shadcn CSS vars
├── lib/
│   └── utils.ts          # cn() helper
├── components/
│   ├── data-table.tsx            # Generic server-side paginated table
│   ├── data-table-pagination.tsx # Pagination controls
│   ├── data-table-toolbar.tsx    # Search + filter bar
│   ├── form-dialog.tsx           # Dialog wrapper for create/edit forms
│   ├── confirm-dialog.tsx        # Delete confirmation
│   ├── availability-grid.tsx     # 7-day × 10-period checkbox matrix
│   ├── loading-spinner.tsx
│   └── empty-state.tsx
└── shadcn/               # shadcn/ui primitives (generated via CLI)
    ├── button.tsx
    ├── input.tsx
    ├── dialog.tsx
    ├── dropdown-menu.tsx
    ├── table.tsx
    ├── form.tsx
    ├── select.tsx
    ├── badge.tsx
    ├── card.tsx
    ├── sidebar.tsx
    ├── breadcrumb.tsx
    ├── separator.tsx
    ├── sheet.tsx
    ├── toast.tsx
    ├── toaster.tsx
    └── tooltip.tsx
```

## Related Code Files

### Modify
- `web/packages/api-client/package.json` — add dependencies (@tanstack/react-query, zod)
- `web/packages/api-client/src/index.ts` — replace stub with real exports
- `web/packages/ui/package.json` — add shadcn/ui deps (tailwindcss, @radix-ui/*, class-variance-authority, clsx, tailwind-merge, lucide-react, react-hook-form, @hookform/resolvers, zod)
- `web/packages/ui/src/index.ts` — replace stub with component exports
- `web/apps/shell/package.json` — add @tanstack/react-query, @microsoft/fetch-event-source

### Create
- All files under `packages/api-client/src/types/` and `packages/api-client/src/hooks/`
- All files under `packages/ui/src/components/` and `packages/ui/src/shadcn/`
- `packages/ui/src/globals.css`
- `packages/ui/src/lib/utils.ts`
- `packages/api-client/src/query-keys.ts`

## Implementation Steps

1. **Install deps** in `packages/ui`:
   ```bash
   cd web && pnpm --filter @mcs-erp/ui add tailwindcss@^4 @radix-ui/react-dialog @radix-ui/react-dropdown-menu @radix-ui/react-select @radix-ui/react-separator @radix-ui/react-slot @radix-ui/react-tooltip class-variance-authority clsx tailwind-merge lucide-react react-hook-form @hookform/resolvers zod
   ```

2. **Install deps** in `packages/api-client`:
   ```bash
   pnpm --filter @mcs-erp/api-client add @tanstack/react-query react
   ```

3. **Install deps** in `apps/shell`:
   ```bash
   pnpm --filter @mcs-erp/shell add @tanstack/react-query @microsoft/fetch-event-source
   ```

4. **Create `packages/ui/src/globals.css`** — Tailwind v4 imports + shadcn CSS variables (dark/light theme)

5. **Create `packages/ui/src/lib/utils.ts`** — `cn()` function using `clsx` + `tailwind-merge`

6. **Generate shadcn primitives** — Create all files in `packages/ui/src/shadcn/` (button, input, dialog, table, form, select, badge, card, sidebar, breadcrumb, separator, sheet, toast, dropdown-menu, tooltip)

7. **Create API types** — One file per module in `packages/api-client/src/types/`; types derived from backend handler request/response shapes

8. **Create query key factory** — `packages/api-client/src/query-keys.ts` — centralized keys for cache invalidation

9. **Create CRUD hooks per module** — Each hook file exports useList, useDetail, useCreate, useUpdate (+ useDelete where applicable). All use `apiFetch` from shell's `api-client.ts` (passed via QueryClient default config or imported directly).

10. **Create SSE hook** — `use-chat-sse.ts` using `@microsoft/fetch-event-source` for POST-based SSE with Bearer auth

11. **Create shared components** — `data-table.tsx` (TanStack Table v8, server-side), `availability-grid.tsx` (7×10 checkbox matrix), `form-dialog.tsx`, `confirm-dialog.tsx`

12. **Update exports** — `packages/ui/src/index.ts` and `packages/api-client/src/index.ts`

13. **Build check** — Run `pnpm --filter @mcs-erp/ui build && pnpm --filter @mcs-erp/api-client build`

## Todo List
- [ ] Install UI package dependencies (Tailwind v4, Radix, shadcn deps)
- [ ] Install API client dependencies (TanStack Query)
- [ ] Install shell dependencies (TanStack Query, fetch-event-source)
- [ ] Create globals.css with Tailwind v4 + shadcn CSS vars
- [ ] Create cn() utility
- [ ] Create shadcn primitive components (~15 files)
- [ ] Create API types (7 type files)
- [ ] Create query key factory
- [ ] Create CRUD hooks (10 hook files)
- [ ] Create SSE hook
- [ ] Create DataTable component
- [ ] Create AvailabilityGrid component
- [ ] Create FormDialog + ConfirmDialog
- [ ] Update package exports
- [ ] Verify `pnpm build` passes

## Success Criteria
- `pnpm build` passes for `@mcs-erp/ui` and `@mcs-erp/api-client`
- All types match backend handler shapes exactly
- DataTable renders with server-side pagination
- AvailabilityGrid renders 7×10 checkbox matrix

## Risk Assessment
- **Tailwind v4 + shadcn/ui compatibility** — shadcn CLI may not fully support Tailwind v4 yet; manually create components if needed
- **Monorepo CSS resolution** — Tailwind may not scan cross-package; use `@source` directive

## Security Considerations
- API client must send Bearer token via Authorization header
- No credentials stored in code; all via httpOnly cookies + in-memory token
- SSE hook must include auth token in request headers

## Next Steps
- Phase 02: Shell layout uses these shared components
- All module phases import from `@mcs-erp/ui` and `@mcs-erp/api-client`
