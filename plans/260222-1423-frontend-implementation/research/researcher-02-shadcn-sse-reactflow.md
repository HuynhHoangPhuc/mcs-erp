# Research Report: shadcn/ui + Tailwind v4 + SSE + react-flow Patterns

_Date: 2026-02-22 | Based on knowledge cutoff Aug 2025_

---

## 1. shadcn/ui in Turborepo pnpm Monorepo (Tailwind v4)

**Recommended structure:**
```
packages/
  ui/                  # shared component library
    src/components/ui/ # shadcn components live here
    src/lib/utils.ts
    package.json
    tsconfig.json
  tailwind-config/     # shared Tailwind config (optional with v4 CSS imports)
apps/
  web/                 # React 19 app consuming packages/ui
```

**Key steps:**

1. Run `pnpm dlx shadcn@latest init` inside `apps/web` first, then move components to `packages/ui`.
2. With **Tailwind v4**, config is CSS-first — no `tailwind.config.js`. Use `@import "tailwindcss"` in CSS.
3. `packages/ui/src/globals.css`:
```css
@import "tailwindcss";
@import "tw-animate-css";

@custom-variant dark (&:is(.dark *));

:root {
  --background: oklch(1 0 0);
  --foreground: oklch(0.145 0 0);
  /* ... shadcn CSS vars */
}
```
4. `packages/ui/package.json` — export CSS and components:
```json
{
  "name": "@repo/ui",
  "exports": {
    "./globals.css": "./src/globals.css",
    "./components/*": "./src/components/*.tsx",
    "./lib/utils": "./src/lib/utils.ts"
  },
  "peerDependencies": { "tailwindcss": "^4.0.0" }
}
```
5. In `apps/web`, import shared styles: `import "@repo/ui/globals.css"` in `main.tsx`.
6. `shadcn.json` in `packages/ui` — set `"rsc": false`, `"tsx": true`, point aliases to package paths.
7. Turborepo: add `packages/ui` as dependency in `apps/web/package.json` → `"@repo/ui": "workspace:*"`.

**Tailwind v4 content scanning**: v4 auto-detects files, but explicitly add `packages/ui/src/**` via `@source` if needed:
```css
@source "../../packages/ui/src";
```

---

## 2. DataTable (TanStack Table v8 + shadcn/ui)

**Server-side pattern with URL state:**
```tsx
// hooks/use-table-state.ts
export function useTableState() {
  const search = useSearch(); // TanStack Router
  return {
    pagination: { pageIndex: search.page ?? 0, pageSize: search.limit ?? 20 },
    sorting: search.sort ? [{ id: search.sort, desc: search.order === 'desc' }] : [],
    columnFilters: search.filter ? [{ id: search.filterCol, value: search.filter }] : [],
  };
}

// components/data-table.tsx
export function DataTable<TData>({ columns, data, pageCount, onStateChange }) {
  const table = useReactTable({
    data,
    columns,
    pageCount,
    manualPagination: true,
    manualSorting: true,
    manualFiltering: true,
    getCoreRowModel: getCoreRowModel(),
    state: { pagination, sorting, columnFilters, columnVisibility },
    onPaginationChange,  // → update URL params
    onSortingChange,
    onColumnFiltersChange,
    onColumnVisibilityChange: setColumnVisibility,
  });
  // render DataTableToolbar + DataTablePagination
}
```

**Column visibility toggle:**
```tsx
<DropdownMenu>
  <DropdownMenuTrigger asChild>
    <Button variant="outline" size="sm"><SlidersHorizontal /></Button>
  </DropdownMenuTrigger>
  <DropdownMenuContent>
    {table.getAllColumns().filter(c => c.getCanHide()).map(col => (
      <DropdownMenuCheckboxItem key={col.id}
        checked={col.getIsVisible()}
        onCheckedChange={v => col.toggleVisibility(v)}>
        {col.id}
      </DropdownMenuCheckboxItem>
    ))}
  </DropdownMenuContent>
</DropdownMenu>
```

**Row actions column:**
```tsx
{
  id: "actions",
  cell: ({ row }) => (
    <DropdownMenu>
      <DropdownMenuTrigger asChild><Button variant="ghost" size="icon"><MoreHorizontal /></Button></DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuItem onClick={() => onEdit(row.original)}>Edit</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem className="text-destructive" onClick={() => onDelete(row.original.id)}>Delete</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
```

---

## 3. Form Components (react-hook-form + zod + shadcn/ui)

**Base pattern:**
```tsx
const schema = z.object({
  name: z.string().min(1, "Required"),
  email: z.string().email(),
});

export function CreateTeacherDialog({ open, onOpenChange, onSuccess }) {
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", email: "" },
  });

  const mutation = useMutation({ mutationFn: createTeacher,
    onSuccess: () => { onSuccess(); onOpenChange(false); form.reset(); }
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader><DialogTitle>Add Teacher</DialogTitle></DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(d => mutation.mutate(d))}>
            <FormField control={form.control} name="name" render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormControl><Input {...field} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <DialogFooter>
              <Button type="submit" disabled={mutation.isPending}>
                {mutation.isPending ? <Loader2 className="animate-spin" /> : "Save"}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
```

**Sheet for edit** (slide-in panel preferred for complex forms):
- Replace `Dialog` → `Sheet`, `DialogContent` → `SheetContent side="right"`.
- Use same form structure; Sheet gives more vertical space.

**Edit pattern**: pass `defaultValues` from row data; `useEffect` to `form.reset(data)` when data changes.

---

## 4. Layout Components

**Collapsible sidebar (shadcn/ui Sidebar component — added Q4 2024):**
```tsx
// app-sidebar.tsx — uses shadcn's built-in <Sidebar> primitive
import { Sidebar, SidebarContent, SidebarHeader, SidebarMenu,
         SidebarMenuItem, SidebarMenuButton, useSidebar } from "@repo/ui/components/ui/sidebar";

export function AppSidebar() {
  const { state } = useSidebar(); // "expanded" | "collapsed"
  return (
    <Sidebar collapsible="icon">
      <SidebarHeader><Logo collapsed={state === "collapsed"} /></SidebarHeader>
      <SidebarContent>
        <SidebarMenu>
          {navItems.map(item => (
            <SidebarMenuItem key={item.href}>
              <SidebarMenuButton asChild tooltip={item.label}>
                <Link to={item.href}><item.icon /><span>{item.label}</span></Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarContent>
    </Sidebar>
  );
}
```

**Breadcrumbs** — derive from TanStack Router's `useMatches`:
```tsx
const matches = useMatches();
const crumbs = matches.filter(m => m.context?.crumb).map(m => m.context.crumb);
```

**Header user dropdown**: `<DropdownMenu>` with `<Avatar>`, profile link, logout action.

**Root layout**: wrap with `<SidebarProvider>` → `<AppSidebar />` + `<main>`.

---

## 5. SSE Streaming in React

### (a) AI Agent Chat

```tsx
// hooks/use-chat-sse.ts
export function useChatSSE(sessionId: string) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [streaming, setStreaming] = useState(false);

  const sendMessage = useCallback(async (content: string) => {
    setStreaming(true);
    // POST to create message, then open SSE stream
    const es = new EventSource(`/api/agent/sessions/${sessionId}/stream`);
    let buffer = "";

    es.addEventListener("delta", e => {
      buffer += e.data;
      setMessages(prev => {
        const last = prev[prev.length - 1];
        if (last?.role === "assistant" && last.streaming) {
          return [...prev.slice(0, -1), { ...last, content: buffer }];
        }
        return [...prev, { role: "assistant", content: buffer, streaming: true }];
      });
    });

    es.addEventListener("done", () => {
      setMessages(prev => prev.map((m, i) =>
        i === prev.length - 1 ? { ...m, streaming: false } : m
      ));
      setStreaming(false);
      es.close();
    });

    es.onerror = () => { setStreaming(false); es.close(); };
  }, [sessionId]);

  return { messages, streaming, sendMessage };
}
```

### (b) Schedule Generation Progress

```tsx
// hooks/use-schedule-progress.ts
export function useScheduleProgress(jobId: string | null) {
  const [progress, setProgress] = useState({ stage: "", percent: 0, done: false });

  useEffect(() => {
    if (!jobId) return;
    const es = new EventSource(`/api/timetable/jobs/${jobId}/progress`);

    es.addEventListener("progress", e => {
      const data = JSON.parse(e.data); // { stage, percent }
      setProgress(p => ({ ...p, ...data }));
    });

    es.addEventListener("complete", e => {
      setProgress(p => ({ ...p, done: true, percent: 100 }));
      es.close();
    });

    es.onerror = () => es.close();
    return () => es.close();
  }, [jobId]);

  return progress;
}
```

**SSE with auth (Bearer token)**: `EventSource` doesn't support custom headers. Options:
- Pass token as query param: `?token=...` (acceptable for short-lived tokens)
- Use `fetch` with `ReadableStream` + `TextDecoderStream` for header support
- Use `@microsoft/fetch-event-source` library for full header support

```tsx
// fetch-based SSE with auth header
import { fetchEventSource } from "@microsoft/fetch-event-source";

await fetchEventSource(`/api/agent/stream`, {
  method: "POST",
  headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  body: JSON.stringify({ message }),
  onmessage(ev) { /* handle */ },
});
```

---

## 6. react-flow — DAG for Subject Prerequisites

```tsx
import ReactFlow, { Background, Controls, MiniMap,
  addEdge, useNodesState, useEdgesState } from "reactflow";
import dagre from "@dagrejs/dagre";

// Auto-layout with dagre
function getLayoutedElements(nodes, edges, direction = "TB") {
  const g = new dagre.graphlib.Graph();
  g.setDefaultEdgeLabel(() => ({}));
  g.setGraph({ rankdir: direction, ranksep: 80, nodesep: 60 });

  nodes.forEach(n => g.setNode(n.id, { width: 180, height: 60 }));
  edges.forEach(e => g.setEdge(e.source, e.target));
  dagre.layout(g);

  return {
    nodes: nodes.map(n => {
      const { x, y } = g.node(n.id);
      return { ...n, position: { x: x - 90, y: y - 30 } };
    }),
    edges,
  };
}

export function PrerequisiteGraph({ subjects, prerequisites, onAddEdge, onRemoveEdge }) {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);

  useEffect(() => {
    const rawNodes = subjects.map(s => ({
      id: s.id, data: { label: s.code }, type: "subjectNode",
    }));
    const rawEdges = prerequisites.map(p => ({
      id: `${p.subjectId}-${p.prerequisiteId}`,
      source: p.prerequisiteId, target: p.subjectId,
      markerEnd: { type: MarkerType.ArrowClosed },
    }));
    const { nodes: ln, edges: le } = getLayoutedElements(rawNodes, rawEdges);
    setNodes(ln); setEdges(le);
  }, [subjects, prerequisites]);

  const onConnect = useCallback(params => {
    onAddEdge(params.source, params.target); // call API
    setEdges(eds => addEdge(params, eds));
  }, [onAddEdge]);

  return (
    <div style={{ height: 600 }}>
      <ReactFlow nodes={nodes} edges={edges}
        onNodesChange={onNodesChange} onEdgesChange={onEdgesChange}
        onConnect={onConnect} onEdgeDoubleClick={(_, e) => onRemoveEdge(e)}
        fitView nodeTypes={nodeTypes}>
        <Background /> <Controls /> <MiniMap />
      </ReactFlow>
    </div>
  );
}
```

**Custom node type** for colored badges, status indicators:
```tsx
const SubjectNode = ({ data }) => (
  <div className="px-4 py-2 rounded-lg border bg-card shadow-sm min-w-[160px]">
    <Handle type="target" position={Position.Top} />
    <p className="font-mono text-sm font-semibold">{data.label}</p>
    <Handle type="source" position={Position.Bottom} />
  </div>
);
const nodeTypes = { subjectNode: SubjectNode };
```

---

## Key Dependencies Summary

| Package | Version | Purpose |
|---|---|---|
| `shadcn/ui` | latest (v4 compatible) | Component primitives |
| `tailwindcss` | ^4.0 | Utility CSS |
| `@tanstack/react-table` | ^8 | DataTable |
| `@tanstack/react-query` | ^5 | Server state |
| `@tanstack/react-router` | ^1 | Routing + URL state |
| `react-hook-form` | ^7 | Forms |
| `zod` | ^3 | Validation |
| `reactflow` | ^12 | DAG visualization |
| `@dagrejs/dagre` | ^1 | Auto-layout |
| `@microsoft/fetch-event-source` | ^2 | SSE with auth headers |

---

## Unresolved Questions

1. **SSE auth strategy** — Go backend needs to support token-in-query-param or implement SSE via POST body; confirm which approach with backend team.
2. **Tailwind v4 CSS vars** — shadcn's CSS variable naming changed between v3→v4 compatibility layer; verify `tw-animate-css` version matches shadcn's expectation.
3. **react-flow v12 license** — v12+ uses MIT; confirm no commercial restrictions for ERP use.
4. **Monorepo shadcn init path** — `shadcn@canary` may be needed for Tailwind v4 support; verify stable vs canary recommendation at time of implementation.
5. **Schedule SSE reconnect** — need retry/reconnect logic if job runs >30s and connection drops; consider polling fallback.
