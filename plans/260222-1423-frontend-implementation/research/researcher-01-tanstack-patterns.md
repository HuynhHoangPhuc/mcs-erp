# TanStack Ecosystem Best Practices — ERP Frontend

**Date:** 2026-02-22 | **Stack:** React 19 + TypeScript + TanStack Router v1 / Query v5 / Table v8 / Form

---

## 1. TanStack Router v1 — Route Tree + Auth Guards + Lazy Loading

### Route Tree Structure (file-based codegen preferred)

```
src/routes/
├── __root.tsx          # Root layout (providers, global shell)
├── _auth.tsx           # Auth layout route (sidebar + outlet)
├── _auth/
│   ├── dashboard.tsx
│   ├── hr/
│   │   ├── index.tsx
│   │   └── $id.tsx
│   └── subjects/index.tsx
├── login.tsx
└── -components/        # Colocated, not routed (prefix -)
```

### Root + Auth Layout

```tsx
// routes/__root.tsx
export const Route = createRootRoute({
  component: () => <Outlet />,
})

// routes/_auth.tsx — sidebar layout + auth guard
export const Route = createFileRoute('/_auth')({
  beforeLoad: async ({ context }) => {
    if (!context.auth.isAuthenticated) {
      throw redirect({ to: '/login' })
    }
  },
  component: () => (
    <div className="flex h-screen">
      <Sidebar />
      <main className="flex-1 overflow-auto"><Outlet /></main>
    </div>
  ),
})
```

### Auth Context via Router Context

```tsx
// main.tsx
const router = createRouter({
  routeTree,
  context: { auth: undefined! }, // typed, injected at render
})

function App() {
  const auth = useAuth()
  return <RouterProvider router={router} context={{ auth }} />
}
```

### Lazy Loading

```tsx
// routes/_auth/hr/index.tsx
export const Route = createFileRoute('/_auth/hr/')({
  component: lazyRouteComponent(() => import('./-components/HrPage')),
})
```

---

## 2. TanStack Query v5 — Hook Factory for CRUD + Pagination

### Query Key Factory (type-safe)

```ts
// lib/query-keys.ts
export const hrKeys = {
  all: ['hr'] as const,
  list: (p: PageParams) => [...hrKeys.all, 'list', p] as const,
  detail: (id: string) => [...hrKeys.all, 'detail', id] as const,
}
```

### Reusable Hook Factory

```ts
// lib/create-crud-hooks.ts
type PageParams = { page: number; per_page: number }

export function createCrudHooks<T, TCreate, TUpdate>(
  resource: string,
  api: {
    list: (p: PageParams) => Promise<{ data: T[]; total: number }>
    get: (id: string) => Promise<T>
    create: (body: TCreate) => Promise<T>
    update: (id: string, body: TUpdate) => Promise<T>
    remove: (id: string) => Promise<void>
  }
) {
  const keys = {
    all: [resource] as const,
    list: (p: PageParams) => [resource, 'list', p] as const,
    detail: (id: string) => [resource, 'detail', id] as const,
  }

  return {
    useList: (params: PageParams) =>
      useQuery({ queryKey: keys.list(params), queryFn: () => api.list(params) }),

    useDetail: (id: string) =>
      useQuery({ queryKey: keys.detail(id), queryFn: () => api.get(id) }),

    useCreate: () => {
      const qc = useQueryClient()
      return useMutation({
        mutationFn: api.create,
        onSuccess: () => qc.invalidateQueries({ queryKey: keys.all }),
      })
    },

    useUpdate: () => {
      const qc = useQueryClient()
      return useMutation({
        mutationFn: ({ id, body }: { id: string; body: TUpdate }) => api.update(id, body),
        // Optimistic update
        onMutate: async ({ id, body }) => {
          await qc.cancelQueries({ queryKey: keys.detail(id) })
          const prev = qc.getQueryData(keys.detail(id))
          qc.setQueryData(keys.detail(id), (old: T) => ({ ...old, ...body }))
          return { prev }
        },
        onError: (_e, { id }, ctx) => qc.setQueryData(keys.detail(id), ctx?.prev),
        onSettled: (_d, _e, { id }) => qc.invalidateQueries({ queryKey: keys.detail(id) }),
      })
    },

    useRemove: () => {
      const qc = useQueryClient()
      return useMutation({
        mutationFn: api.remove,
        onSuccess: () => qc.invalidateQueries({ queryKey: keys.all }),
      })
    },
  }
}

// Usage
export const useTeachers = createCrudHooks('teachers', teachersApi)
```

---

## 3. TanStack Table v8 — Server-Side Pagination + Sort + Filter

```tsx
// hooks/use-server-table.ts
export function useServerTable<T>(
  data: T[] | undefined,
  total: number,
  columns: ColumnDef<T>[],
  pagination: PaginationState,
  onPaginationChange: OnChangeFn<PaginationState>,
  sorting: SortingState,
  onSortingChange: OnChangeFn<SortingState>,
) {
  return useReactTable({
    data: data ?? [],
    columns,
    pageCount: Math.ceil(total / pagination.pageSize),
    state: { pagination, sorting },
    onPaginationChange,
    onSortingChange,
    manualPagination: true,
    manualSorting: true,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
  })
}

// Column definition example
const columns: ColumnDef<Teacher>[] = [
  { id: 'select', header: ({ table }) => <Checkbox checked={table.getIsAllRowsSelected()} onChange={table.toggleAllRowsSelected} />, cell: ({ row }) => <Checkbox checked={row.getIsSelected()} onChange={row.toggleSelected} /> },
  { accessorKey: 'name', header: 'Name', enableSorting: true },
  { accessorKey: 'department', header: 'Department' },
  { id: 'actions', cell: ({ row }) => <RowActions row={row} /> },
]

// Component wiring
function TeachersTable() {
  const [pagination, setPagination] = useState<PaginationState>({ pageIndex: 0, pageSize: 20 })
  const [sorting, setSorting] = useState<SortingState>([])
  const { data } = useTeachers.useList({ page: pagination.pageIndex + 1, per_page: pagination.pageSize })
  const table = useServerTable(data?.data, data?.total ?? 0, columns, pagination, setPagination, sorting, setSorting)
  return <DataTable table={table} />
}
```

---

## 4. TanStack Form — Validation + Query Mutation Integration

```tsx
// forms/teacher-form.tsx
import { useForm } from '@tanstack/react-form'
import { z } from 'zod'

const schema = z.object({
  name: z.string().min(2),
  email: z.string().email(),
  department_id: z.string().uuid(),
})

export function TeacherForm({ onSuccess }: { onSuccess: () => void }) {
  const createTeacher = useTeachers.useCreate()

  const form = useForm({
    defaultValues: { name: '', email: '', department_id: '' },
    validators: {
      onChange: schema,          // field-level on change
      onSubmit: schema,          // full validation on submit
    },
    onSubmit: async ({ value }) => {
      await createTeacher.mutateAsync(value)
      onSuccess()
    },
  })

  return (
    <form onSubmit={(e) => { e.preventDefault(); form.handleSubmit() }}>
      <form.Field name="name" children={(field) => (
        <>
          <Input value={field.state.value} onChange={e => field.handleChange(e.target.value)} />
          {field.state.meta.errors.map(e => <p key={e} className="text-red-500 text-sm">{e}</p>)}
        </>
      )} />
      <Button type="submit" disabled={form.state.isSubmitting || createTeacher.isPending}>
        {createTeacher.isPending ? 'Saving...' : 'Save'}
      </Button>
    </form>
  )
}
```

---

## Key Decisions for MCS-ERP

| Concern | Pattern |
|---|---|
| Auth guard | `beforeLoad` on `_auth` layout route |
| Query key namespacing | resource-keyed factory, invalidate `keys.all` on mutations |
| Pagination params | `page` (1-based) + `per_page=20`, map to `pageIndex = page-1` |
| Optimistic updates | only for `useUpdate`, not create/delete (list invalidation simpler) |
| Form validation | zod schema passed to TanStack Form validators |
| Table row selection | `id: 'select'` column with `enableRowSelection: true` on table |

---

## Unresolved Questions

1. Does the backend return `{ data: T[], total: number }` envelope or something else? (affects query hook shape)
2. Sorting: does API accept `sort_by=name&sort_dir=asc` or a different convention?
3. Are there role-based route guards needed (RBAC per module), or just auth?
