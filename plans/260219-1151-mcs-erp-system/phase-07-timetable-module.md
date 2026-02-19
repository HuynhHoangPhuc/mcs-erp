# Phase 07: Timetable Module

## Context Links
- [Parent Plan](./plan.md)
- Depends on: [Phase 04 - HR](./phase-04-hr-module.md), [Phase 05 - Subject](./phase-05-subject-module.md), [Phase 06 - Room](./phase-06-room-module.md)
- [SA Research](./research/researcher-02-frontend-scheduling-ai.md)
- [Go ERP Research](../reports/researcher-260219-1151-go-erp-architecture-research.md)

## Overview
- **Date:** 2026-02-19
- **Priority:** P1
- **Status:** Complete
- **Description:** Timetable bounded context: Semester management, auto-scheduling (greedy init + simulated annealing with parallel goroutines), teacher+room assignment, conflict detection, admin review workflow. Most complex module.

## Key Insights
- Scheduling is a constraint satisfaction problem; greedy+SA handles 80% of academic cases
- Hard constraints must be zero violations; soft constraints minimized via SA penalty function
- Parallel SA: run 4-8 independent chains as goroutines, take global best
- Semester workflow: create semester -> select subjects -> auto-schedule -> review -> approve/reject
- Schedule is immutable once approved; modifications create new version
- Cross-module reads: teacher availability (HR), subject hours/prerequisites (Subject), room capacity/availability (Room)

## Requirements

### Functional
- Semester CRUD: name, start_date, end_date, status (draft/scheduling/review/approved)
- Semester subject selection: pick which subjects to schedule this semester
- Teacher-subject assignment: hybrid manual + suggestions (admin assigns, system suggests based on qualifications + availability)
<!-- Updated: Validation Session 2 - Hybrid manual+suggestions for teacher assignment -->
- Auto-scheduling: generate timetable assignments (subject + teacher + room + timeslot)
- Hard constraints: no teacher double-book, no room double-book, teacher available, room available, room capacity >= 1
- Soft constraints: minimize teacher gaps, prefer teacher preferred slots, even day distribution
- Conflict detection: list all hard/soft violations in a generated schedule
- Admin review: view schedule, see conflicts, manually swap assignments, approve/reject
- Schedule versioning: each generation creates a new version; compare versions

### Non-Functional
- Scheduling for 200 subjects, 80 teachers, 50 rooms completes in < 30 seconds
- SA runs 500k iterations with cooling rate 0.9995
- Parallel: 4 goroutines per scheduling run

## Architecture

```
internal/timetable/
├── domain/
│   ├── semester.go            # Semester aggregate root
│   ├── assignment.go          # Assignment: subject+teacher+room+slot
│   ├── time_slot.go           # TimeSlot value object (day+period)
│   ├── schedule.go            # Schedule: collection of assignments + metadata
│   ├── constraint.go          # Constraint interface + hard/soft implementations
│   ├── events.go              # ScheduleGenerated, ScheduleApproved, AssignmentModified
│   └── repository.go          # SemesterRepository, ScheduleRepository
├── application/
│   ├── commands/
│   │   ├── create_semester.go
│   │   ├── select_subjects.go      # Add subjects to semester
│   │   ├── generate_schedule.go    # Trigger auto-scheduling
│   │   ├── modify_assignment.go    # Manual swap/move
│   │   └── approve_schedule.go     # Approve or reject
│   ├── queries/
│   │   ├── get_semester.go
│   │   ├── list_semesters.go
│   │   ├── get_schedule.go         # Current schedule + conflicts
│   │   └── get_conflicts.go        # List all violations
│   └── events/
│       └── handlers.go
├── scheduler/
│   ├── greedy.go              # Greedy initial assignment
│   ├── annealing.go           # Simulated annealing optimizer
│   ├── neighbor.go            # Neighbor functions (swap, move, reassign room)
│   ├── cost.go                # Cost function (hard*10000 + soft)
│   ├── constraints.go         # Constraint implementations
│   └── parallel.go            # Parallel SA runner
├── infrastructure/
│   ├── postgres_semester_repo.go
│   ├── postgres_schedule_repo.go
│   └── cross_module_reader.go # Read from HR, Subject, Room repos
├── delivery/
│   ├── semester_handler.go    # /api/timetable/semesters
│   ├── schedule_handler.go    # /api/timetable/semesters/:id/schedule
│   └── assignment_handler.go  # /api/timetable/assignments
├── tools/
│   ├── generate_schedule.go   # AI tool
│   ├── modify_assignment.go   # AI tool
│   └── explain_conflicts.go   # AI tool
└── module.go
```

### Frontend
```
web/packages/module-timetable/src/
├── routes.ts
├── pages/
│   ├── semester-list-page.tsx
│   ├── semester-detail-page.tsx
│   ├── semester-setup-page.tsx     # Select subjects + teachers
│   ├── schedule-view-page.tsx      # Timetable grid
│   └── schedule-review-page.tsx    # Review + approve
├── components/
│   ├── semester-table.tsx
│   ├── semester-form.tsx
│   ├── subject-selection.tsx       # Multi-select subjects for semester
│   ├── teacher-assignment.tsx      # Assign teachers to subjects
│   ├── timetable-grid.tsx          # Weekly grid view (day x period)
│   ├── conflict-panel.tsx          # List conflicts with severity
│   ├── assignment-card.tsx         # Draggable assignment in grid
│   └── schedule-progress.tsx       # Progress indicator during generation
└── queries/
    ├── use-semesters.ts
    ├── use-schedule.ts
    ├── use-conflicts.ts
    └── use-generate-schedule.ts    # Mutation + polling for result
```

## Related Code Files

### Files to Create

**Backend Domain:**
- `internal/timetable/domain/semester.go` — Semester: ID, Name, StartDate, EndDate, Status (draft/scheduling/review/approved), SubjectIDs []uuid, CreatedAt
- `internal/timetable/domain/assignment.go` — Assignment: ID, SemesterID, SubjectID, TeacherID, RoomID, Day int, Period int, Version int
- `internal/timetable/domain/time_slot.go` — TimeSlot{Day, Period}, slot constants, `Conflicts(a, b)` helper
- `internal/timetable/domain/schedule.go` — Schedule: SemesterID, Version, Assignments []Assignment, HardViolations int, SoftPenalty float64, GeneratedAt
- `internal/timetable/domain/constraint.go` — `Constraint` interface: `Violations(schedule) int`, `IsHard() bool`
- `internal/timetable/domain/events.go` — ScheduleGenerated, ScheduleApproved, AssignmentModified
- `internal/timetable/domain/repository.go` — SemesterRepository, ScheduleRepository

**Scheduler (pure Go, no DB):**
- `internal/timetable/scheduler/greedy.go` — `GreedyAssign(problem) Schedule`: sort subjects by constraint count desc, assign to first valid slot
- `internal/timetable/scheduler/annealing.go` — `Anneal(initial Schedule, cfg SAConfig) Schedule`: SA main loop
- `internal/timetable/scheduler/neighbor.go` — `SwapMove`, `MoveSlot`, `ReassignRoom` — weighted random selection (50/35/15)
- `internal/timetable/scheduler/cost.go` — `Cost(schedule) int`: hard*10000 + soft. `EvaluateHard(schedule)`, `EvaluateSoft(schedule)`
- `internal/timetable/scheduler/constraints.go` — `TeacherConflict`, `RoomConflict`, `TeacherUnavailable`, `RoomUnavailable` (hard). `TeacherGap`, `PreferredSlot`, `EvenDistribution` (soft).
- `internal/timetable/scheduler/parallel.go` — `ParallelAnneal(problem, cfg, numWorkers) Schedule`: spawn goroutines, collect best

**Backend Application:**
- `internal/timetable/application/commands/create_semester.go`
- `internal/timetable/application/commands/select_subjects.go` — set subject list for semester
- `internal/timetable/application/commands/generate_schedule.go` — load data from HR/Subject/Room, build problem, run parallel SA, save result
- `internal/timetable/application/commands/modify_assignment.go` — swap/move single assignment, re-evaluate conflicts
- `internal/timetable/application/commands/approve_schedule.go` — set status approved/rejected
- `internal/timetable/application/queries/get_semester.go`
- `internal/timetable/application/queries/list_semesters.go`
- `internal/timetable/application/queries/get_schedule.go` — returns schedule + conflict summary
- `internal/timetable/application/queries/get_conflicts.go` — detailed conflict list
- `internal/timetable/application/queries/suggest_teachers.go` — for a subject, query HR teachers with matching qualifications + available slots, rank by fit
<!-- Updated: Validation Session 2 - Teacher suggestion query for hybrid assignment -->

**Backend Infrastructure:**
- `internal/timetable/infrastructure/postgres_semester_repo.go`
- `internal/timetable/infrastructure/postgres_schedule_repo.go`
- `internal/timetable/infrastructure/cross_module_reader.go` — reads teacher availability, subject data, room data via gRPC reader service clients (HR, Subject, Room)
<!-- Updated: Validation Session 1 - Cross-module reads via gRPC services -->

**Backend Delivery:**
- `internal/timetable/delivery/semester_handler.go` — CRUD /api/timetable/semesters, POST /:id/subjects
- `internal/timetable/delivery/schedule_handler.go` — POST /api/timetable/semesters/:id/generate, GET /:id/schedule, POST /:id/approve
- `internal/timetable/delivery/assignment_handler.go` — PUT /api/timetable/assignments/:id (modify)

**AI Tools:**
- `internal/timetable/tools/generate_schedule.go` — trigger schedule generation
- `internal/timetable/tools/modify_assignment.go` — modify a specific assignment
- `internal/timetable/tools/explain_conflicts.go` — explain why conflicts exist in natural language

**Module:**
- `internal/timetable/module.go` — dependencies: `["core", "hr", "subject", "room"]`

**SQL:**
- `sqlc/queries/timetable/semesters.sql`
- `sqlc/queries/timetable/assignments.sql`
- `migrations/timetable/000001_create_semesters_table.up.sql`
- `migrations/timetable/000002_create_semester_subjects_table.up.sql`
- `migrations/timetable/000003_create_assignments_table.up.sql`

**Frontend:**
- All files listed in Frontend section above

## Implementation Steps

1. **SQL migrations**
   ```sql
   -- semesters
   CREATE TABLE semesters (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       name VARCHAR(255) NOT NULL,
       start_date DATE NOT NULL,
       end_date DATE NOT NULL,
       status VARCHAR(20) NOT NULL DEFAULT 'draft'
           CHECK (status IN ('draft','scheduling','review','approved','rejected')),
       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
       updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
   );
   -- semester_subjects (which subjects are in this semester)
   CREATE TABLE semester_subjects (
       semester_id UUID REFERENCES semesters(id) ON DELETE CASCADE,
       subject_id UUID NOT NULL,
       teacher_id UUID,  -- assigned teacher (nullable until scheduled)
       PRIMARY KEY (semester_id, subject_id)
   );
   -- assignments (the actual timetable)
   CREATE TABLE assignments (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       semester_id UUID REFERENCES semesters(id) ON DELETE CASCADE,
       subject_id UUID NOT NULL,
       teacher_id UUID NOT NULL,
       room_id UUID NOT NULL,
       day INT NOT NULL CHECK (day BETWEEN 0 AND 6),
       period INT NOT NULL CHECK (period >= 0),
       version INT NOT NULL DEFAULT 1,
       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
       UNIQUE (semester_id, room_id, day, period, version),
       UNIQUE (semester_id, teacher_id, day, period, version)
   );
   ```

2. **Domain entities** — Semester with state machine (draft->scheduling->review->approved), Assignment as value object within Schedule aggregate

3. **Scheduler: constraints.go** — implement each constraint:
   - `TeacherConflict`: count assignments where same teacher+day+period appears twice
   - `RoomConflict`: count assignments where same room+day+period appears twice
   - `TeacherUnavailable`: check teacher availability from HR data
   - `RoomUnavailable`: check room availability from Room data
   - `TeacherGap`: count gaps between teacher's classes in a day
   - `EvenDistribution`: variance of assignments per day

4. **Scheduler: greedy.go** — sort subjects by constraint difficulty (most constrained first = fewest valid slots). For each, try all slots, pick first valid one. Return partial schedule if some unplaceable.

5. **Scheduler: annealing.go** — SA with config:
   ```go
   type SAConfig struct {
       TInitial    float64 // 1000.0
       CoolingRate float64 // 0.9995
       TMin        float64 // 0.01
       MaxIter     int     // 500000
   }
   ```

6. **Scheduler: neighbor.go** — three moves, weighted random:
   - `SwapMove` (50%): pick 2 random assignments, swap their slots
   - `MoveSlot` (35%): pick 1 assignment, move to random empty slot
   - `ReassignRoom` (15%): pick 1 assignment, assign different room

7. **Scheduler: parallel.go** — spawn N goroutines each running Anneal with shuffled initial schedule. Collect results via channel, return best (lowest cost).

8. **Cross-module reader** — `cross_module_reader.go` accepts HR, Subject, Room repository interfaces. Builds `ScheduleProblem` struct:
   ```go
   type ScheduleProblem struct {
       Subjects      []SubjectInfo     // id, hours_per_week
       Teachers      []TeacherInfo     // id, availability slots
       Rooms         []RoomInfo        // id, capacity, equipment, availability
       TeacherAssign map[SubjectID]TeacherID // pre-assigned teachers
       Slots         []TimeSlot        // all available day+period combos
   }
   ```

9. **Command: generate_schedule** — load problem data, call `ParallelAnneal`, save result with version, set semester status to `review`, publish `ScheduleGenerated`

10. **Command: modify_assignment** — load current schedule, apply modification, re-evaluate conflicts, save as same version (or new), publish `AssignmentModified`

11. **Command: approve_schedule** — validate zero hard violations, set semester status to `approved`, publish `ScheduleApproved`

12. **Query: get_schedule** — return assignments grouped by day+period + conflict summary

13. **HTTP handlers** — semester CRUD, generate (POST, returns SSE stream with progress events: `{type: "progress", percent: 45, message: "SA iteration 225000/500000"}` and final `{type: "complete", schedule_id: "..."}` or `{type: "error", message: "..."}`), schedule view, assignment modify, approve/reject
<!-- Updated: Validation Session 1 - SSE progress stream for schedule generation instead of HTTP 202+polling -->

14. **Frontend: semester setup wizard** — Step 1: create semester (name, dates). Step 2: select subjects (multi-select from active subjects). Step 3: assign teachers (manual + "Suggest Teachers" button that queries HR for teachers with matching qualifications + available slots; admin confirms/overrides). Step 4: click "Generate Schedule".
<!-- Updated: Validation Session 2 - Hybrid teacher assignment with suggestion button -->

15. **Frontend: timetable grid** — CSS Grid or table: columns = days (Mon-Sat), rows = periods (1-10). Each cell shows assignment cards (subject name, teacher, room). Color-code by conflict status (green=ok, red=hard violation, yellow=soft violation).

16. **Frontend: conflict panel** — sidebar listing conflicts grouped by type. Click conflict to highlight relevant cells in grid.

17. **Frontend: schedule progress** — poll `/api/timetable/semesters/:id/schedule` while status is `scheduling`. Show progress bar or spinner.

18. **AI tools** — `generate_schedule(semester_id)`, `modify_assignment(assignment_id, new_day, new_period, new_room)`, `explain_conflicts(semester_id)` returns human-readable conflict summary

## Todo List
- [x] SQL migrations (semesters, semester_subjects, assignments)
- [x] Domain entities (Semester, Assignment, Schedule, TimeSlot)
- [x] Constraint interface + hard/soft implementations
- [x] Greedy initial assignment
- [x] Simulated annealing optimizer
- [x] Neighbor functions (swap, move, reassign room)
- [x] Cost function (hard*10000 + soft)
- [x] Parallel SA runner
- [x] Cross-module reader (HR, Subject, Room data)
- [x] Domain events
- [x] Repository interfaces + infrastructure
- [x] Command handlers (create semester, select subjects, generate, modify, approve)
- [x] Query handlers (semesters, schedule, conflicts, suggest_teachers)
- [x] HTTP handlers
- [x] AI tools (generate, modify, explain)
- [x] Timetable module registration
- [x] Frontend semester setup wizard
- [x] Frontend timetable grid
- [x] Frontend conflict panel
- [x] Frontend schedule progress indicator
- [x] TanStack Query hooks
- [x] Unit tests: constraints, greedy, SA (critical)
- [x] Integration test: full generate + review flow

## Success Criteria
- Auto-scheduling generates a valid timetable (zero hard violations) for test dataset
- SA completes in < 30s for 200 subjects / 80 teachers / 50 rooms
- Conflict detection correctly identifies all violations
- Admin can manually modify assignments and see updated conflicts
- Semester workflow: draft -> scheduling -> review -> approved works end-to-end
- Frontend grid renders schedule with color-coded conflict indicators

## Risk Assessment
- **SA convergence**: May not reach zero hard violations for over-constrained problems. Mitigation: return best found + list remaining hard violations for manual resolution.
- **Cross-module coupling**: Timetable reads from 3 other modules. Use read-only interfaces, not direct repo access. Event-driven cache refresh if perf needed.
- **Long-running generation**: 30s blocking HTTP is bad. Use SSE stream for real-time progress updates during SA iterations. Frontend shows progress bar with iteration count and current best cost.
<!-- Updated: Validation Session 1 - SSE confirmed as async pattern -->

## Security Considerations
- Schedule generation requires `timetable:schedule:write` permission
- Approval requires `timetable:schedule:approve` (separate from write)
- Cross-module reads respect tenant isolation (same schema context)
- Schedule data includes teacher assignments — treat as internal/confidential

## Next Steps
- Phase 8 (AI Agent) uses timetable tools for conversational schedule management
