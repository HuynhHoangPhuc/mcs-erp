# Phase 04: HR Module

## Context Links
- [Parent Plan](./plan.md)
- Depends on: [Phase 03 - Auth & RBAC](./phase-03-auth-rbac.md)
- [Brainstorm Report](../reports/brainstorm-260219-1151-mcs-erp-architecture.md)

## Overview
- **Date:** 2026-02-19
- **Priority:** P1
- **Status:** Complete
- **Description:** HR bounded context: Teacher, Department, Availability, Qualification entities with full CQRS pattern. Frontend: teacher list, teacher form, availability calendar. AI tools: search_teachers, get_teacher_availability.

## Key Insights
- First domain module — establishes the pattern all other modules follow
- Teacher availability is critical input for timetable scheduling (Phase 7)
- Availability modeled as weekly recurring slots (day + period), not calendar dates
- Department is lightweight: just grouping + head teacher reference
- Events published for cross-module use: `TeacherCreated`, `AvailabilityUpdated`

## Requirements

### Functional
- Teacher CRUD: name, email, department, qualifications, status (active/inactive)
- Department CRUD: name, description, head teacher
- Availability management: per-teacher weekly slot grid (available/unavailable per day+period)
- Qualification tags: free-form list on teacher (e.g., "Math", "Physics", "Lab Certified")
- List teachers with filtering (department, status, qualification) + pagination
- Domain events: TeacherCreated, TeacherUpdated, AvailabilityUpdated

### Non-Functional
- Teacher list supports 500+ records with server-side pagination
- Availability grid loads in < 200ms

## Architecture

```
internal/hr/
├── domain/
│   ├── teacher.go           # Teacher aggregate root
│   ├── department.go        # Department entity
│   ├── availability.go      # Availability value object (day+period grid)
│   ├── qualification.go     # Qualification value object
│   ├── events.go            # TeacherCreated, TeacherUpdated, AvailabilityUpdated
│   └── repository.go        # TeacherRepository, DepartmentRepository interfaces
├── application/
│   ├── commands/
│   │   ├── create_teacher.go
│   │   ├── update_teacher.go
│   │   ├── set_availability.go
│   │   ├── create_department.go
│   │   └── update_department.go
│   ├── queries/
│   │   ├── list_teachers.go
│   │   ├── get_teacher.go
│   │   ├── get_availability.go
│   │   └── list_departments.go
│   └── events/
│       └── handlers.go      # React to own events if needed
├── infrastructure/
│   ├── postgres_teacher_repo.go
│   └── postgres_department_repo.go
├── delivery/
│   ├── teacher_handler.go   # /api/hr/teachers
│   ├── department_handler.go # /api/hr/departments
│   └── availability_handler.go # /api/hr/teachers/:id/availability
├── tools/
│   ├── search_teachers.go   # AI tool: search teachers by criteria
│   └── get_availability.go  # AI tool: get teacher availability
└── module.go                # HR Module implementation
```

### Frontend
```
web/packages/module-hr/src/
├── routes.ts                # createHrRoutes(parentRoute)
├── pages/
│   ├── teacher-list-page.tsx
│   ├── teacher-detail-page.tsx
│   ├── teacher-form-page.tsx
│   ├── department-list-page.tsx
│   └── availability-page.tsx
├── components/
│   ├── teacher-table.tsx     # TanStack Table
│   ├── teacher-form.tsx      # TanStack Form + Zod
│   ├── department-select.tsx
│   ├── qualification-tags.tsx
│   └── availability-grid.tsx # Weekly slot grid (checkboxes)
└── queries/
    ├── use-teachers.ts
    ├── use-teacher.ts
    ├── use-departments.ts
    └── use-availability.ts
```

## Related Code Files

### Files to Create

**Backend Domain:**
- `internal/hr/domain/teacher.go` — Teacher: ID, Name, Email, DepartmentID, Qualifications []string, IsActive, CreatedAt, UpdatedAt
- `internal/hr/domain/department.go` — Department: ID, Name, Description, HeadTeacherID
- `internal/hr/domain/availability.go` — WeeklyAvailability: TeacherID, Slots []TimeSlot where TimeSlot = {Day int, Period int, Available bool}
- `internal/hr/domain/qualification.go` — Qualification type alias (string)
- `internal/hr/domain/events.go` — TeacherCreated{ID, Name, TenantID}, TeacherUpdated{ID, Fields}, AvailabilityUpdated{TeacherID}
- `internal/hr/domain/repository.go` — TeacherRepository, DepartmentRepository interfaces

**Backend Application:**
- `internal/hr/application/commands/create_teacher.go` — validates, saves, publishes TeacherCreated
- `internal/hr/application/commands/update_teacher.go` — validates, updates, publishes TeacherUpdated
- `internal/hr/application/commands/set_availability.go` — replace availability grid, publish AvailabilityUpdated
- `internal/hr/application/commands/create_department.go`
- `internal/hr/application/commands/update_department.go`
- `internal/hr/application/queries/list_teachers.go` — filter by dept, status, qualification; paginated
- `internal/hr/application/queries/get_teacher.go` — by ID, includes department + availability
- `internal/hr/application/queries/get_availability.go` — by teacher ID
- `internal/hr/application/queries/list_departments.go`

**Backend Infrastructure:**
- `internal/hr/infrastructure/postgres_teacher_repo.go` — sqlc-backed
- `internal/hr/infrastructure/postgres_department_repo.go` — sqlc-backed

**Backend Delivery:**
- `internal/hr/delivery/teacher_handler.go` — GET /api/hr/teachers, GET /:id, POST, PUT /:id, DELETE /:id
- `internal/hr/delivery/department_handler.go` — CRUD /api/hr/departments
- `internal/hr/delivery/availability_handler.go` — GET/PUT /api/hr/teachers/:id/availability

**AI Tools:**
- `internal/hr/tools/search_teachers.go` — tool: search_teachers(query, department, qualification)
- `internal/hr/tools/get_availability.go` — tool: get_teacher_availability(teacher_id)

**Module:**
- `internal/hr/module.go` — implements Module interface, registers routes + events + tools

**SQL:**
- `sqlc/queries/hr/teachers.sql`
- `sqlc/queries/hr/departments.sql`
- `sqlc/queries/hr/availability.sql`
- `migrations/hr/000001_create_departments_table.up.sql`
- `migrations/hr/000002_create_teachers_table.up.sql`
- `migrations/hr/000003_create_teacher_availability_table.up.sql`

**Frontend:**
- `web/packages/module-hr/src/routes.ts`
- `web/packages/module-hr/src/pages/teacher-list-page.tsx`
- `web/packages/module-hr/src/pages/teacher-detail-page.tsx`
- `web/packages/module-hr/src/pages/teacher-form-page.tsx`
- `web/packages/module-hr/src/pages/department-list-page.tsx`
- `web/packages/module-hr/src/pages/availability-page.tsx`
- `web/packages/module-hr/src/components/teacher-table.tsx`
- `web/packages/module-hr/src/components/teacher-form.tsx`
- `web/packages/module-hr/src/components/availability-grid.tsx`
- `web/packages/module-hr/src/components/qualification-tags.tsx`
- `web/packages/module-hr/src/queries/use-teachers.ts`
- `web/packages/module-hr/src/queries/use-teacher.ts`
- `web/packages/module-hr/src/queries/use-departments.ts`
- `web/packages/module-hr/src/queries/use-availability.ts`

## Implementation Steps

1. **SQL migrations** — departments table, teachers table (FK to departments), teacher_availability table (teacher_id + day + period, unique constraint)

2. **Domain entities** — Teacher aggregate with `NewTeacher(name, email, deptID)`, `Update(...)`, `SetAvailability(slots)`. Department entity.

3. **Domain events** — define event structs, embed base event with ID + timestamp + tenant

4. **Repository interfaces** — TeacherRepository: `Save`, `FindByID`, `List(filters, page)`, `Update`, `Delete`. DepartmentRepository similar.

5. **sqlc queries** — write raw SQL for all CRUD + filtered list with dynamic WHERE
   - Teacher list query: use `sqlc.arg` for optional filters
   - Availability: batch upsert (DELETE + INSERT for teacher's weekly grid)

6. **Infrastructure repos** — implement interfaces using sqlc generated code, handle tenant schema via tx

7. **Command handlers** — each command: validate input -> call repo -> publish event via event bus

8. **Query handlers** — each query: call repo with filters -> return DTO (not domain entity)

9. **HTTP handlers** — parse request, call command/query handler, return JSON response
   - Teacher list: `GET /api/hr/teachers?department_id=&status=&q=&page=&per_page=`
   - Availability: `PUT /api/hr/teachers/:id/availability` with body `{slots: [{day, period, available}]}`

10. **AI tools** — implement `tools.Tool` interface from langchaingo
    - `search_teachers`: accepts JSON `{query, department, qualification}`, returns teacher list
    - `get_teacher_availability`: accepts `{teacher_id}`, returns availability grid

11. **Module registration** — `hr.Module` implements `pkg/module.Module`, dependencies: `["core"]`

12. **Frontend routes** — `createHrRoutes(parent)` with lazy-loaded pages

13. **Teacher table** — TanStack Table with columns: name, email, department, qualifications, status. Server-side pagination.

14. **Teacher form** — TanStack Form + Zod: name (required), email (required, email format), department (select), qualifications (tag input)

15. **Availability grid** — 7-column (days) x N-row (periods) checkbox grid. PUT on save.

16. **Query hooks** — `useTeachers(filters)`, `useTeacher(id)`, `useDepartments()`, `useAvailability(teacherId)`

## Todo List
- [x] SQL migrations (departments, teachers, availability)
- [x] Domain entities (Teacher, Department, Availability)
- [x] Domain events (TeacherCreated, AvailabilityUpdated)
- [x] Repository interfaces
- [x] sqlc queries
- [x] Infrastructure repos
- [x] Command handlers (create/update teacher, set availability, department CRUD)
- [x] Query handlers (list/get teachers, availability, departments)
- [x] HTTP handlers (teacher, department, availability)
- [x] AI tools (search_teachers, get_availability)
- [x] HR module registration
- [x] Frontend routes + pages
- [x] Teacher table component
- [x] Teacher form component
- [x] Availability grid component
- [x] TanStack Query hooks
- [x] Integration test: teacher CRUD end-to-end

## Success Criteria
- Teacher CRUD works via API (create, read, update, deactivate)
- Department CRUD works
- Availability grid saves and loads correctly per teacher
- Teacher list supports filtering by department, status, and search query
- Events published on teacher create/update and availability change
- AI tools return correct data when called
- Frontend teacher list renders with pagination
- Frontend teacher form creates/updates teacher

## Risk Assessment
- **Availability schema**: weekly recurring slots (not dates) may not cover exceptions (holidays). Acceptable for MVP.
- **Qualification free-text**: No standardization. May cause AI search mismatches. Add autocomplete from existing values.

## Security Considerations
- All endpoints require auth + `hr:teacher:read` / `hr:teacher:write` permissions
- Teacher email is PII — do not log or expose in error messages
- Department head assignment restricted to users with `hr:department:write`

## Next Steps
- Phase 7 (Timetable) consumes teacher availability data
- Phase 8 (AI Agent) uses registered HR tools
