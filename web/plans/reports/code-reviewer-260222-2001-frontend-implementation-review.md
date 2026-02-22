# Code Review: MCS-ERP Frontend Implementation

## Code Review Summary

### Scope
- Files: ~85 source files (TS/TSX) across 8 packages
- Packages: `api-client`, `ui`, `module-hr`, `module-subject`, `module-room`, `module-timetable`, `module-agent`, `apps/shell`
- Focus: Full initial implementation review
- Scout findings: Swapped function arguments, inconsistent API base URL, missing React keys, state sync anti-pattern, auth guard weakness

### Overall Assessment

Solid MVP implementation with clean architecture: well-structured monorepo, consistent patterns across modules, proper TanStack Query usage with centralized query keys, and good component decomposition. However, there are 2 critical bugs, several high-priority security/correctness issues, and a few medium-priority improvements needed.

---

### Critical Issues

#### 1. CRITICAL BUG: Swapped arguments in `useUpdateAssignment` call

**File:** `/Users/phuc/Developer/mcs-erp/web/packages/module-timetable/src/components/assignment-edit-dialog.tsx:28`

The hook `useUpdateAssignment` is defined as `(assignmentId, semesterId)` but called with arguments in reverse order:

```typescript
// Definition in use-schedule.ts:38
export function useUpdateAssignment(assignmentId: string, semesterId: string) {

// Call in assignment-edit-dialog.tsx:28
const mutation = useUpdateAssignment(assignment.semester_id, assignment.id);
//                                    ^^^^^^^^^^^^^^^^^^^^^^^^  ^^^^^^^^^^^^^^
//                                    This is semesterId!       This is assignmentId!
```

**Impact:** Every assignment edit will PUT to the wrong URL (`/timetable/assignments/{semesterId}` instead of `/timetable/assignments/{assignmentId}`) and invalidate the wrong cache key. This will cause 404s or corrupt data.

**Fix:** Swap the arguments:
```typescript
const mutation = useUpdateAssignment(assignment.id, assignment.semester_id);
```

#### 2. CRITICAL: Auth refresh URL inconsistency between `api-client.ts` and `auth-provider.tsx`

**Files:**
- `/Users/phuc/Developer/mcs-erp/web/packages/api-client/src/lib/api-client.ts:64` uses `${API_BASE_URL}/api/v1/auth/refresh`
- `/Users/phuc/Developer/mcs-erp/web/apps/shell/src/providers/auth-provider.tsx:38` uses `"/api/v1/auth/refresh"` (relative, no `API_BASE_URL`)

**Impact:** If the API is deployed on a different origin (e.g., `api.mcs-erp.com` while the frontend is on `app.mcs-erp.com`), the auth-provider's session restore will fail silently. Users will be forced to re-login on every page load even with a valid refresh token.

**Fix:** Reuse `apiFetch` from the api-client package:
```typescript
// auth-provider.tsx - replace the raw fetch with:
import { apiFetch } from "../lib/api-client";

const data = await apiFetch<TokenResponse>("/auth/refresh", {
  method: "POST",
  body: JSON.stringify({ refresh_token: refreshToken }),
});
```

Note: This introduces a chicken-and-egg issue since `apiFetch` adds a Bearer token. The refresh endpoint should accept unauthenticated requests, so either strip the auth header for refresh calls or create a separate `publicFetch` helper.

---

### High Priority

#### 3. Auth guard checks only `localStorage` presence, not token validity

**File:** `/Users/phuc/Developer/mcs-erp/web/apps/shell/src/routes/_authenticated.tsx:9`

```typescript
beforeLoad: () => {
  const hasToken = !!localStorage.getItem("refresh_token");
  if (!hasToken) {
    throw redirect({ to: "/login" });
  }
},
```

**Issue:** This only checks if a string exists in localStorage. An expired, revoked, or malformed token will pass this check. The user will see the authenticated shell momentarily before the first API call fails with 401 and triggers a redirect.

**Recommendation:** Check the AuthProvider's state instead. Pass the auth context via router context so `beforeLoad` can check `isAuthenticated` from the actual auth state. At minimum, add a TODO comment documenting this known limitation.

#### 4. JWT payload decoding without validation

**File:** `/Users/phuc/Developer/mcs-erp/web/apps/shell/src/providers/auth-provider.tsx:51,70`

```typescript
const payload = JSON.parse(atob(data.access_token.split(".")[1]));
setUser({ email: payload.email, permissions: payload.permissions || [] });
```

**Issues:**
- No try/catch around `atob`/`JSON.parse` -- a malformed JWT will crash the app
- No validation that `payload.email` exists or is a string
- `atob` does not handle URL-safe base64 (`-` and `_` characters) which JWTs use per RFC 7515

**Fix:**
```typescript
function decodeJwtPayload(token: string): AuthUser {
  try {
    const base64 = token.split(".")[1]
      .replace(/-/g, "+")
      .replace(/_/g, "/");
    const payload = JSON.parse(atob(base64));
    return {
      email: typeof payload.email === "string" ? payload.email : "",
      permissions: Array.isArray(payload.permissions) ? payload.permissions : [],
    };
  } catch {
    return { email: "", permissions: [] };
  }
}
```

#### 5. Race condition: `useChatSSE` token fetched twice per request

**File:** `/Users/phuc/Developer/mcs-erp/web/packages/api-client/src/hooks/use-chat-sse.ts:45`

```typescript
headers: {
  ...(getAccessToken() ? { Authorization: `Bearer ${getAccessToken()}` } : {}),
},
```

`getAccessToken()` is called twice. If the token is refreshed between the two calls (e.g., by a concurrent 401 retry in `apiFetch`), the check and the value could be inconsistent. Minor but could lead to intermittent auth failures.

**Fix:** Capture once:
```typescript
const token = getAccessToken();
headers: {
  ...(token ? { Authorization: `Bearer ${token}` } : {}),
},
```

#### 6. `SemesterSetupStep` state sync during render (anti-pattern)

**File:** `/Users/phuc/Developer/mcs-erp/web/packages/module-timetable/src/components/semester-setup-step.tsx:21-23`

```typescript
if (semesterSubjects && selectedIds.size === 0 && existingIds.size > 0) {
  setSelectedIds(new Set(existingIds));
}
```

**Issue:** Calling `setState` during render triggers an extra render cycle and can cause infinite loops in edge cases. Also, the condition `selectedIds.size === 0` means if a user intentionally deselects all subjects, the state will be reset back on the next render.

**Fix:** Use `useEffect` for data sync:
```typescript
useEffect(() => {
  if (semesterSubjects) {
    setSelectedIds(new Set(semesterSubjects.map(s => s.subject_id)));
  }
}, [semesterSubjects]);
```

#### 7. Missing React keys on fragment in `TimetableGrid`

**File:** `/Users/phuc/Developer/mcs-erp/web/packages/module-timetable/src/components/timetable-grid.tsx:59`

```tsx
{PERIODS.map((p) => (
  <>   {/* <-- no key! */}
    <div key={`label-${p}`} ... />
    {DAYS.map((_, dayIdx) => ...)}
  </>
))}
```

**Impact:** React will log a warning and may have trouble efficiently reconciling the DOM. Use `<Fragment key={...}>` instead.

**Fix:**
```tsx
import { Fragment } from "react";
// ...
{PERIODS.map((p) => (
  <Fragment key={`period-${p}`}>
    <div className="...">P{p}</div>
    {DAYS.map(...)}
  </Fragment>
))}
```

---

### Medium Priority

#### 8. `apiFetch` sets Content-Type for all requests including file uploads

**File:** `/Users/phuc/Developer/mcs-erp/web/packages/api-client/src/lib/api-client.ts:23`

```typescript
if (!headers.has("Content-Type")) {
  headers.set("Content-Type", "application/json");
}
```

This forces `application/json` on every request. If multipart/form-data uploads are added later, this will break them because the browser needs to set the Content-Type with the boundary. Not a problem today but a landmine.

#### 9. Token refresh has no concurrency protection

**File:** `/Users/phuc/Developer/mcs-erp/web/packages/api-client/src/lib/api-client.ts:32-44`

If multiple API calls receive 401 simultaneously, each will independently call `tryRefresh()`, resulting in multiple refresh requests. Only the first one should succeed (depending on server's refresh token rotation policy), and subsequent ones may fail, causing unnecessary logouts.

**Recommendation:** Add a mutex/promise cache:
```typescript
let refreshPromise: Promise<boolean> | null = null;

async function tryRefreshOnce(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = tryRefresh().finally(() => { refreshPromise = null; });
  }
  return refreshPromise;
}
```

#### 10. Pagination offset=0 is falsy, skipped in query params

**File:** `/Users/phuc/Developer/mcs-erp/web/packages/api-client/src/hooks/use-teachers.ts:9`

```typescript
if (filters?.offset) params.set("offset", String(filters.offset));
```

When `offset` is `0`, this condition is falsy and the param is omitted. This is correct behavior if the backend defaults to 0, but it means explicit `offset=0` is never sent. Same pattern in `use-subjects.ts`, `use-rooms.ts`, `use-semesters.ts`.

**Minor concern:** If the backend requires an explicit offset param, page 1 would break. Likely not an issue but worth noting.

#### 11. `QueryClient` created outside component but inside module scope

**File:** `/Users/phuc/Developer/mcs-erp/web/apps/shell/src/routes/__root.tsx:6`

```typescript
const queryClient = new QueryClient({ ... });
```

This is fine for SPA but would cause shared state between requests in SSR. Not a concern for MVP but worth noting if SSR is planned.

#### 12. Duplicate type definitions for `AuthUser` and `TokenResponse`

- `AuthUser` defined in both `auth-provider.tsx` and `types/auth.ts`
- `TokenResponse` defined in both `auth-provider.tsx` and `types/auth.ts`

**Recommendation:** Import from `@mcs-erp/api-client` instead of re-declaring.

#### 13. `ConflictPanel` `detectConflicts` runs on every render

**File:** `/Users/phuc/Developer/mcs-erp/web/packages/module-timetable/src/components/conflict-panel.tsx:68`

```typescript
const conflicts = detectConflicts(assignments, teacherMap, roomMap);
```

Called directly in render without `useMemo`. For large schedules this could be expensive (O(n) with map operations).

**Fix:** Wrap in `useMemo` keyed on assignments length or use the caller's already-computed conflicts.

#### 14. Hard-coded `limit: 200` / `limit: 1000` in multiple components

**Files:**
- `semester-review-step.tsx:21-23` uses `limit: 200`
- `semester-setup-step.tsx:13` uses `limit: 200`
- `semester-assign-step.tsx:14-15` uses `limit: 200`
- `prerequisite-dag-page.tsx:32` uses `limit: 1000`

**Issue:** These assume all data fits in a single page. Will silently drop items if the dataset exceeds the limit.

**Recommendation:** Either implement proper server-side iteration or document this as a known MVP limitation.

---

### Low Priority

#### 15. `AssignmentEditDialog` state not reset when `assignment` prop changes

**File:** `/Users/phuc/Developer/mcs-erp/web/packages/module-timetable/src/components/assignment-edit-dialog.tsx:23-26`

```typescript
const [teacherId, setTeacherId] = useState(assignment.teacher_id);
```

`useState` only uses the initial value. If the parent changes the `assignment` prop, the dialog will show stale data. Should add an `useEffect` to sync, or use `key={assignment.id}` on the dialog to force remount.

#### 16. `useParams({ strict: false })` with type assertion across 4 files

All detail pages use `as { paramName: string }` type assertion after `strict: false`. This bypasses TanStack Router's type safety. Consider using the typed `Route.useParams()` pattern from TanStack Router's file-based routing.

#### 17. `onNodeClick` in prerequisite DAG has stale closure over `highlightedChain`

**File:** `/Users/phuc/Developer/mcs-erp/web/packages/module-subject/src/components/prerequisite-dag-page.tsx:100`

```typescript
const onNodeClick = useCallback(
  (_, node) => {
    if (highlightedChain.has(node.id) && highlightedChain.size === 1) { ... }
  },
  [highlightedChain, rawEdges]
);
```

This recreates the callback every time `highlightedChain` changes. Since `highlightedChain` is a `Set` that's replaced on each click, this is correct but causes unnecessary re-renders in ReactFlow. Consider using a ref for the highlight state.

---

### Edge Cases Found by Scout

1. **Swapped arguments** in `useUpdateAssignment` -- caught by tracing call/definition signatures
2. **Refresh URL mismatch** -- auth-provider uses relative URL while api-client uses `API_BASE_URL` prefix
3. **State-during-render** in `SemesterSetupStep` -- found by analyzing render flow
4. **Missing Fragment key** in `TimetableGrid` -- found by reviewing `.map()` patterns
5. **JWT decode crash** on malformed tokens -- found by tracing error paths
6. **Concurrent refresh race** -- found by analyzing multi-request 401 scenarios

---

### Positive Observations

1. **Clean monorepo structure** -- clear separation between `api-client` (data layer), `ui` (presentation primitives), and module packages
2. **Centralized query keys** -- `query-keys.ts` factory pattern prevents key collisions and makes cache invalidation predictable
3. **Consistent hook patterns** -- all TanStack Query hooks follow the same structure (query key, queryFn, mutation with invalidation)
4. **Proper form validation** -- zod schemas with `react-hook-form` for teacher forms
5. **Good component decomposition** -- `FormDialog`, `ConfirmDialog`, `DataTable`, `AvailabilityGrid` are well-abstracted shared components
6. **No `dangerouslySetInnerHTML`** -- all user-facing content is rendered safely
7. **Markdown rendering via ReactMarkdown** -- the AI chat uses `react-markdown` which sanitizes by default (no XSS)
8. **SSE streaming implementation** -- clean `useChatSSE` hook with proper abort handling
9. **File sizes under 200 lines** -- every source file respects the modularization guideline
10. **Conflict detection** -- timetable module has client-side conflict detection for both teacher and room double-booking

---

### Recommended Actions (Priority Order)

1. **Fix swapped `useUpdateAssignment` arguments** -- critical bug, 1-line fix
2. **Fix auth-provider refresh URL** to use `API_BASE_URL` or reuse `apiFetch`
3. **Add try/catch and URL-safe base64 handling** to JWT decode in auth-provider
4. **Add Fragment key** to `TimetableGrid` period rows
5. **Move `SemesterSetupStep` state sync** from render to `useEffect`
6. **Capture `getAccessToken()` once** in `useChatSSE`
7. **Add refresh token mutex** to prevent concurrent refresh requests
8. **Import shared types** (`AuthUser`, `TokenResponse`) from `@mcs-erp/api-client` instead of re-declaring
9. **Memoize `detectConflicts`** call in `ConflictPanel`
10. **Add `key={assignment.id}`** to `AssignmentEditDialog` to reset state on prop change

---

### Metrics

- Type Coverage: High (all API types defined, zod validation on forms, proper generics on hooks)
- Test Coverage: 0% (no frontend tests found -- **major gap** for a production path)
- Linting Issues: 1 eslint-disable comment (justified, in prerequisite-dag-page.tsx)
- Security: No XSS vectors, no `dangerouslySetInnerHTML`, JWT handled client-side (acceptable for MVP)
- Build: Not verified (no `tsc --noEmit` run)

---

### Unresolved Questions

1. Is the backend `refresh` endpoint idempotent with refresh token rotation? If not, the concurrent refresh race (item #9) will cause logouts under load.
2. Should pagination hard limits (200/1000) be documented as MVP constraints or addressed now?
3. Is SSR planned? If so, the module-level `QueryClient` instantiation and `localStorage` usage in route guards need refactoring.
4. Are frontend tests in scope for this phase?
