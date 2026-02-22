# Phase 02: Shell & Layout

## Context
- [Plan](./plan.md) | [Phase 01](./phase-01-shared-foundation.md)
- Depends on: Phase 01

## Overview
- **Priority:** P1
- **Status:** pending
- **Effort:** 4h

Replace the bare-bones shell app with a proper layout: sidebar navigation, header with user info, breadcrumbs, TanStack Router route tree for all modules, TanStack Query provider.

## Key Insights
- Current shell: basic `<header>MCS-ERP</header>` + `<Outlet />`; login page with inline styles
- Need sidebar with nav links for: Dashboard, HR (Teachers, Departments), Subjects, Rooms, Timetable, Agent Chat
- Auth guard already works via `_authenticated` route with localStorage check
- TanStack Router `useMatches()` for breadcrumbs
- shadcn `<Sidebar>` primitive with `collapsible="icon"` mode

## Requirements

### Functional
- Collapsible sidebar with icon-only mode
- Navigation links grouped by module (HR, Academic, Scheduling, AI)
- Header showing user email + logout button + tenant name
- Breadcrumbs auto-generated from route matches
- QueryClientProvider wrapping the app
- Route tree with lazy-loaded module routes
- Dashboard page with summary cards (placeholder data for now)

### Non-functional
- Responsive: sidebar collapses to overlay on mobile
- Navigation state persists across page loads (collapsed/expanded)

## Architecture

### Route Tree Structure
```
__root.tsx (AuthProvider + QueryClientProvider + <Outlet />)
├── login.tsx
└── _authenticated.tsx (auth guard + AppLayout with sidebar)
    ├── index.tsx (Dashboard)
    ├── teachers/
    │   ├── index.tsx (list)
    │   └── $teacherId.tsx (detail/edit)
    ├── departments/
    │   └── index.tsx (list)
    ├── subjects/
    │   ├── index.tsx (list)
    │   ├── $subjectId.tsx (detail/edit)
    │   └── prerequisites.tsx (DAG view)
    ├── categories/
    │   └── index.tsx (list)
    ├── rooms/
    │   ├── index.tsx (list)
    │   └── $roomId.tsx (detail/edit)
    ├── timetable/
    │   ├── index.tsx (semester list)
    │   └── $semesterId.tsx (semester detail + schedule)
    └── chat/
        └── index.tsx (agent chat)
```

### Layout Components
```
apps/shell/src/
├── components/
│   ├── app-layout.tsx        # Sidebar + Header + main content area
│   ├── app-sidebar.tsx       # Navigation sidebar using shadcn Sidebar
│   ├── app-header.tsx        # Top bar with breadcrumbs + user menu
│   ├── nav-group.tsx         # Sidebar nav group (label + links)
│   └── user-menu.tsx         # User dropdown (email, logout)
├── routes/
│   ├── __root.tsx            # Updated: add QueryClientProvider
│   ├── login.tsx             # Updated: use shadcn/ui components
│   ├── _authenticated.tsx    # Updated: wrap with AppLayout
│   └── _authenticated/
│       ├── index.tsx         # Dashboard with cards
│       ├── teachers.index.tsx
│       ├── teachers.$teacherId.tsx
│       ├── departments.index.tsx
│       ├── subjects.index.tsx
│       ├── subjects.$subjectId.tsx
│       ├── subjects.prerequisites.tsx
│       ├── categories.index.tsx
│       ├── rooms.index.tsx
│       ├── rooms.$roomId.tsx
│       ├── timetable.index.tsx
│       ├── timetable.$semesterId.tsx
│       └── chat.index.tsx
└── router.ts                 # Updated route tree
```

## Related Code Files

### Modify
- `web/apps/shell/src/main.tsx` — add QueryClientProvider
- `web/apps/shell/src/router.ts` — new route tree with all module routes
- `web/apps/shell/src/routes/__root.tsx` — add QueryClientProvider, import globals.css
- `web/apps/shell/src/routes/_authenticated.tsx` — wrap with AppLayout
- `web/apps/shell/src/routes/_authenticated/index.tsx` — dashboard with cards
- `web/apps/shell/src/routes/login.tsx` — restyle with shadcn/ui

### Create
- `web/apps/shell/src/components/app-layout.tsx`
- `web/apps/shell/src/components/app-sidebar.tsx`
- `web/apps/shell/src/components/app-header.tsx`
- `web/apps/shell/src/components/nav-group.tsx`
- `web/apps/shell/src/components/user-menu.tsx`
- All route files under `_authenticated/` (stub pages initially, filled in Phase 3-7)

## Implementation Steps

1. **Update `__root.tsx`** — Import `globals.css` from `@mcs-erp/ui`, wrap children with `QueryClientProvider`

2. **Create `app-sidebar.tsx`** — shadcn `<Sidebar>` with navigation groups:
   - Dashboard (Home icon)
   - HR: Teachers, Departments (Users icon)
   - Academic: Subjects, Categories, Prerequisites (BookOpen icon)
   - Rooms (DoorOpen icon)
   - Scheduling: Timetable (Calendar icon)
   - AI: Chat (MessageSquare icon)

3. **Create `app-header.tsx`** — Breadcrumbs (from route matches) + `<UserMenu />` on right

4. **Create `app-layout.tsx`** — `<SidebarProvider>` wrapping `<AppSidebar />` + main area with `<AppHeader />` + `<Outlet />`

5. **Update `_authenticated.tsx`** — Replace bare `<Outlet />` with `<AppLayout />`

6. **Restyle `login.tsx`** — Use shadcn Card, Input, Button, Label components

7. **Create stub route files** — Each module gets an `index.tsx` with placeholder `<h2>Module Name</h2>`. Real content in Phase 3-7.

8. **Update `router.ts`** — Build complete route tree with all routes

9. **Update dashboard** — Summary cards: total teachers, subjects, rooms, semesters (use API hooks from Phase 01 or placeholder)

10. **Build check** — `pnpm --filter @mcs-erp/shell build`

## Todo List
- [ ] Add QueryClientProvider to root
- [ ] Import globals.css from @mcs-erp/ui
- [ ] Create AppSidebar with navigation groups
- [ ] Create AppHeader with breadcrumbs + user menu
- [ ] Create AppLayout combining sidebar + header + content
- [ ] Update _authenticated route to use AppLayout
- [ ] Restyle login page with shadcn/ui
- [ ] Create stub route files for all modules
- [ ] Update router.ts with complete route tree
- [ ] Create dashboard with summary cards
- [ ] Verify `pnpm build` passes

## Success Criteria
- Sidebar navigates between all module pages
- Sidebar collapses to icon-only mode
- Breadcrumbs show correct path
- User menu shows email + logout works
- Login page looks polished
- All routes load without errors

## Risk Assessment
- **Route tree size** — Many routes; ensure lazy loading to avoid bundle bloat
- **Sidebar state** — localStorage for collapse preference may conflict with SSR (not applicable — SPA only)

## Security Considerations
- Auth guard in `_authenticated` route prevents unauthorized access
- Logout clears tokens from memory + localStorage

## Next Steps
- Phase 3-7: Fill stub route pages with actual module UI
