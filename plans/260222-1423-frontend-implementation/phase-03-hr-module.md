# Phase 03: HR Module

## Context
- [Plan](./plan.md) | [Phase 01](./phase-01-shared-foundation.md) | [Phase 02](./phase-02-shell-and-layout.md)
- Depends on: Phase 01, Phase 02
- Backend routes: `GET/POST /api/v1/teachers`, `GET/PUT /api/v1/teachers/{id}`, `GET/PUT /api/v1/teachers/{id}/availability`, `GET/POST/PUT/DELETE /api/v1/departments`

## Overview
- **Priority:** P2
- **Status:** pending
- **Effort:** 5h

Build teacher management UI (list, create, edit, availability grid) and department management. First real module — establishes patterns for Subject and Room modules.

## Key Insights
- Teacher list: `{ items: Teacher[], total: number }` with `?offset=&limit=&department_id=&status=&qualification=`
- Availability: 7-day × 10-period boolean grid per teacher
- Departments: simple CRUD (name, description)
- Teacher has optional `department_id` FK

## Requirements

### Functional
- Teacher list page with filtering (department dropdown, status toggle, search by name/email)
- Teacher create/edit form in dialog (name, email, department, qualifications as tags)
- Teacher detail page with availability grid
- Availability grid: checkboxes for day 0-6 × period 1-10, save on submit
- Department list with inline create/edit/delete
- Pagination on teacher list

### Non-functional
- Optimistic updates on availability grid save
- Loading skeletons while data loads

## Architecture

### File Structure
```
packages/module-hr/src/
├── index.ts                        # Export route components
├── components/
│   ├── teacher-list-page.tsx        # Main teacher list with filters + table
│   ├── teacher-form-dialog.tsx      # Create/edit teacher dialog
│   ├── teacher-detail-page.tsx      # Teacher detail + availability
│   ├── teacher-columns.tsx          # TanStack Table column definitions
│   ├── teacher-filters.tsx          # Filter bar (department, status, search)
│   ├── department-list-page.tsx     # Department management
│   └── department-form-dialog.tsx   # Create/edit department
```

### Route Pages (in shell)
```
apps/shell/src/routes/_authenticated/
├── teachers.index.tsx      → imports TeacherListPage from @mcs-erp/module-hr
├── teachers.$teacherId.tsx → imports TeacherDetailPage from @mcs-erp/module-hr
└── departments.index.tsx   → imports DepartmentListPage from @mcs-erp/module-hr
```

## Related Code Files

### Modify
- `web/packages/module-hr/package.json` — add deps (@mcs-erp/ui, @mcs-erp/api-client, @tanstack/react-table, @tanstack/react-query, react-hook-form, @hookform/resolvers, zod)
- `web/packages/module-hr/src/index.ts` — export page components
- `web/apps/shell/src/routes/_authenticated/teachers.index.tsx` — import TeacherListPage
- `web/apps/shell/src/routes/_authenticated/teachers.$teacherId.tsx` — import TeacherDetailPage
- `web/apps/shell/src/routes/_authenticated/departments.index.tsx` — import DepartmentListPage

### Create
- All files under `packages/module-hr/src/components/`

## Implementation Steps

1. **Add deps** to `module-hr/package.json`: `@mcs-erp/ui`, `@mcs-erp/api-client`, `@tanstack/react-table`, `@tanstack/react-query`, `react-hook-form`, `@hookform/resolvers`, `zod`

2. **Create `teacher-columns.tsx`** — Column defs: name, email, department, qualifications (badges), status (active/inactive badge), actions (edit, view)

3. **Create `teacher-filters.tsx`** — Department select dropdown, status toggle (all/active/inactive), search input with debounce

4. **Create `teacher-form-dialog.tsx`** — react-hook-form + zod schema: `{ name: string, email: email, department_id?: string, qualifications: string[] }`. Tag input for qualifications.

5. **Create `teacher-list-page.tsx`** — DataTable with filters, create button, pagination. Uses `useTeachers` hook with filter/pagination state.

6. **Create `teacher-detail-page.tsx`** — Teacher info card + `<AvailabilityGrid>` component. Uses `useTeacher(id)` + `useTeacherAvailability(id)`. Save button calls `useUpdateTeacherAvailability`.

7. **Create `department-form-dialog.tsx`** — Simple form: name, description

8. **Create `department-list-page.tsx`** — DataTable with create/edit/delete. Uses `useDepartments` hook.

9. **Wire route pages** — Update stub routes in shell to import from `@mcs-erp/module-hr`

10. **Build check** — `pnpm --filter @mcs-erp/module-hr build`

## Todo List
- [ ] Add dependencies to module-hr package
- [ ] Create teacher column definitions
- [ ] Create teacher filter bar
- [ ] Create teacher form dialog (create/edit)
- [ ] Create teacher list page with DataTable
- [ ] Create teacher detail page with availability grid
- [ ] Create department form dialog
- [ ] Create department list page
- [ ] Wire route pages in shell
- [ ] Verify build passes

## Success Criteria
- Teacher list shows paginated data with working filters
- Create/edit teacher dialog submits and refreshes list
- Availability grid displays 7×10 matrix, saves changes
- Department CRUD works end-to-end

## Risk Assessment
- **Qualifications tag input** — No built-in shadcn tag input; may need simple comma-separated input or custom component

## Security Considerations
- All API calls go through authenticated `apiFetch` wrapper
- Department delete should show confirmation dialog

## Next Steps
- Phase 04 (Subject) follows same patterns established here
