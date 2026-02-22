# Frontend Architecture

**Location:** `/web` (React 19 Turborepo monorepo)

The frontend is organized as a monorepo with a main shell application and feature-specific packages, enabling independent development and testing of each module's UI.

## Shell Application (`/web/apps/shell`)

The main React 19 SPA built with Vite, serving as the entry point for all users.

**Tech Stack:**
- **React 19** — Latest with concurrent rendering
- **TypeScript** — Type-safe frontend code
- **TanStack Router** — File-based routing with auth guards
- **TanStack Query** — Server state management + auto caching
- **TanStack Form** — Form state with validation
- **shadcn/ui + Tailwind v4** — Component library + styling
- **Vite** — Lightning-fast dev/build

**Key Features:**
- Auth guard middleware: JWT tokens auto-refreshed, protected routes redirect to login
- Collapsible sidebar navigation with module links
- Breadcrumb header for context
- Responsive design (mobile/tablet/desktop)
- Error boundaries + fallback UI

**Structure:**
```
apps/shell/src/
├── routes/                  # TanStack Router file-based routing
│   ├── __root.tsx          # Root layout (sidebar, header)
│   ├── login.tsx            # Login page
│   ├── dashboard/           # Dashboard routes
│   ├── hr/                  # HR module routes
│   ├── subject/             # Subject module routes
│   └── ...
├── providers/               # Context providers
│   ├── QueryProvider.tsx   # TanStack Query setup
│   ├── AuthProvider.tsx    # Auth context + token management
│   └── ThemeProvider.tsx   # Theme/dark mode
├── hooks/                   # Custom React hooks
│   ├── useAuth.ts          # Access auth context
│   └── useQuery.ts         # Custom query hooks
└── lib/                     # Utilities
    ├── api-client.ts       # Axios instance with auth
    └── constants.ts        # API endpoints, time slots
```

## Feature Packages

Each module has a dedicated package for encapsulation and independent deployment if needed.

### `packages/ui` (Shared Component Library)
Pre-built, reusable components using shadcn/ui patterns:
- Form components (Input, Select, Checkbox, etc.)
- Data table components (columns, sorting, pagination)
- Dialog/Modal components
- Layout components (Card, Tabs, etc.)
- All components accept TypeScript prop types for full IDE support

### `packages/api-client` (API Types + Hooks)
- **Generated types** from backend OpenAPI spec (if available)
- **TanStack Query hooks** for all endpoints (useListTeachers, useCreateTeacher, etc.)
- **Query key factories** for consistent cache invalidation
- Centralized API error handling

### `packages/module-hr` (Teachers & Departments)
**Routes:** `/hr/teachers`, `/hr/departments`

**Pages:**
- Teachers table with CRUD actions (create, edit, delete)
- Department list with modal-based CRUD
- Availability grid editor (7 days × 10 periods per teacher)

**Features:**
- Inline form validation
- Optimistic updates (UI updates before server confirm)
- Bulk availability edit
- Search/filter/sort on tables

### `packages/module-subject` (Subject Catalog)
**Routes:** `/subject/list`, `/subject/categories`, `/subject/{id}`

**Pages:**
- Subject list with CRUD
- Category manager
- Prerequisite DAG visualizer (directed graph)
- Prerequisite editor with cycle detection feedback

**Features:**
- Prerequisite dependency tree view
- Real-time cycle detection (disable "add" button if would create cycle)
- Drag-drop reordering of categories (optional)

### `packages/module-room` (Room Management)
**Routes:** `/room/list`, `/room/{id}`

**Pages:**
- Room list with capacity/equipment filters
- Room detail page with:
  - Metadata editor (name, capacity, equipment)
  - Availability grid editor (7 days × 10 periods)

### `packages/module-timetable` (Scheduling)
**Routes:** `/timetable/semesters`, `/timetable/schedule/{semesterId}`

**Pages:**
- Semester wizard (create/edit semester metadata)
- Subject-teacher-room assignment form
- Schedule generation page with:
  - Progress stream (real-time status SSE)
  - Conflict visualization (red cells for overlaps)
  - Manual adjustment interface
- Schedule approval page (admin view)

**Features:**
- Form validation (subject → teacher already assigned?)
- SSE streaming for schedule generation progress
- Conflict highlighting in grid
- Manual period override + validation

### `packages/module-agent` (AI Chatbot)
**Routes:** `/agent/chat`, `/agent/conversations`

**Pages:**
- Chat interface with message list
- Conversation sidebar (list + new conversation)
- Message input with SSE streaming response

**Features:**
- Real-time message streaming (SSE)
- Conversation history navigation
- Tool call indicators (when agent invokes tools)
- Copy/regenerate message actions

## State Management Strategy

**Server State (TanStack Query):**
- All API data managed by Query (teachers, subjects, schedules, etc.)
- Automatic caching, invalidation, and deduplication
- Optimistic updates for better UX
- Stale-while-revalidate pattern for background sync

**Client State (React Context + Hooks):**
- Auth state (user, JWT token, refresh logic)
- Theme/UI preferences (light/dark mode)
- Pagination state per table
- Modal/dialog open/close state

**Form State (TanStack Form):**
- Form validation (client + server error display)
- Dirty tracking (warn before navigate if unsaved)
- Multi-step forms (semester wizard)

## Authentication Flow

1. **Login Page** — User enters email + password
2. **Auth Provider** — Calls POST /api/v1/auth/login, stores JWT in memory
3. **Token Refresh** — Before each request, check expiry; if <5min, call POST /api/v1/auth/refresh
4. **Protected Routes** — TanStack Router redirects to /login if token missing/invalid
5. **Auto-logout** — If refresh fails, clear token and redirect to login

**Token Storage:**
- Memory-only (no localStorage, safer for XSS)
- Refresh token stored in secure HttpOnly cookie (backend sends on login)

## API Integration Pattern

```typescript
// In packages/api-client/src/queries.ts
import { useQuery, useMutation } from '@tanstack/react-query'

export const useListTeachers = () => {
  return useQuery({
    queryKey: ['teachers'],
    queryFn: () => api.get('/api/v1/teachers'),
  })
}

export const useCreateTeacher = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data) => api.post('/api/v1/teachers', data),
    onSuccess: () => {
      // Invalidate and refetch
      queryClient.invalidateQueries({ queryKey: ['teachers'] })
    },
  })
}
```

**Usage in Component:**
```typescript
export const TeachersPage = () => {
  const { data: teachers, isLoading } = useListTeachers()
  const { mutate: createTeacher } = useCreateTeacher()

  return (
    <div>
      {isLoading && <Spinner />}
      {teachers?.map(t => <TeacherRow key={t.id} teacher={t} />)}
      <CreateTeacherForm onSubmit={createTeacher} />
    </div>
  )
}
```

## Responsive Design & Layout

**Breakpoints (Tailwind v4):**
- `sm`: 640px (tablets)
- `md`: 768px (larger tablets)
- `lg`: 1024px (desktops)
- `xl`: 1280px (large desktops)

**Main Layout:**
```
┌─────────────────────────────────────┐
│        Header (breadcrumb)          │
├──────────┬──────────────────────────┤
│ Sidebar  │                          │
│ (collapse│    Main Content          │
│ on mobile)│   (route pages)          │
│          │                          │
└──────────┴──────────────────────────┘
```

On mobile (`<md`), sidebar becomes a drawer/hamburger menu.

## Error Handling

**Network Errors:**
- 4xx errors: Display form-level error messages (e.g., validation)
- 5xx errors: Show toast notification + retry button
- Network timeout: Show "no connection" banner

**Validation Errors:**
- Field-level errors from TanStack Form validation
- Server-side errors mapped to form fields (if API returns field errors)

**Auth Errors:**
- 401 Unauthorized: Auto-refresh token; if fails, redirect to login
- 403 Forbidden: Show "permission denied" message
