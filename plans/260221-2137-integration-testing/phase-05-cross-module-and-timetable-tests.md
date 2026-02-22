---
phase: 5
title: "Cross-Module & Timetable Tests"
status: completed
priority: P1
effort: 5h
depends_on: [1, 3, 4]
---

# Phase 5: Cross-Module & Timetable Tests

## Context Links
- [timetable/module.go](/internal/timetable/module.go) -- Timetable module with cross-module deps
- [timetable/infrastructure/cross_module_reader.go](/internal/timetable/infrastructure/cross_module_reader.go) -- CrossModuleReader assembles Problem
- [timetable/infrastructure/cross_module_adapters.go](/internal/timetable/infrastructure/cross_module_adapters.go) -- adapters from repo interfaces
- [timetable/scheduler/types.go](/internal/timetable/scheduler/types.go) -- Problem, SubjectInfo, TeacherInfo, RoomInfo
- [timetable/scheduler/greedy.go](/internal/timetable/scheduler/greedy.go) -- greedy scheduler
- [timetable/scheduler/annealing.go](/internal/timetable/scheduler/annealing.go) -- SA optimizer
- [timetable/delivery/schedule_handler.go](/internal/timetable/delivery/schedule_handler.go) -- generate/approve endpoints
- [timetable/delivery/semester_handler.go](/internal/timetable/delivery/semester_handler.go) -- semester CRUD

## Overview
Test the timetable module end-to-end: semester creation, subject selection, teacher assignment, schedule generation (greedy + SA), conflict detection, and admin approval. These tests validate cross-module data flow: HR teachers + Subject subjects + Room rooms -> Timetable scheduler.

## Key Insights
<!-- Updated: Validation Session 1 - Constraint-only validation for SA output -->
- `CrossModuleReader.BuildProblem()` reads from HR, Subject, Room repos via narrow adapter interfaces
- **Scheduler tests validate constraints only** (zero hard violations), not exact assignments — SA is non-deterministic
- `NewModuleWithRepos()` wires cross-module repos at construction time (see main.go)
- Scheduler input: `scheduler.Problem{Subjects, Teachers, Rooms, TeacherAssign, Slots}`
- TimeSlot: `{Day: 1-7, Period: 1-10}` = 70 slots per week
- Schedule generation: greedy initial solution -> SA refinement
- Conflict types: teacher double-booking, room double-booking
- Approval flow: generate -> review -> approve (status transitions)
- SSE progress streaming during generation

## Requirements

### Functional
- Semester CRUD (create, list, get)
- Add subjects to semester
- Assign teacher to semester subject
- Generate schedule with real data from HR/Subject/Room modules
- Generated schedule has no hard constraint violations (teacher/room conflicts)
- Retrieve latest schedule
- Approve schedule (status: pending -> approved)
- Update individual assignment (manual override)
- Conflict detection rejects invalid manual overrides

### Non-functional
- Schedule generation for 10 subjects, 5 teachers, 5 rooms < 5s
- All data reads cross-module boundaries correctly

## Architecture
```
internal/timetable/
  semester_integration_test.go  -- semester CRUD + subject management
  scheduling_integration_test.go -- end-to-end scheduling flow
  cross_module_reader_test.go   -- CrossModuleReader unit/integration test
```

## Related Code Files

### Files to Create
- `internal/timetable/semester_integration_test.go` (~120 lines)
- `internal/timetable/scheduling_integration_test.go` (~180 lines)
- `internal/timetable/cross_module_reader_test.go` (~100 lines)

### Files to Reference
- `internal/timetable/domain/semester.go` -- Semester, SemesterSubject structs
- `internal/timetable/domain/schedule.go` -- Schedule, status constants
- `internal/timetable/domain/assignment.go` -- Assignment struct
- `internal/timetable/domain/time_slot.go` -- TimeSlot, AllSlots()
- `internal/timetable/scheduler/constraints.go` -- hard/soft constraint checks

## Implementation Steps

### Step 1: Create `internal/timetable/cross_module_reader_test.go`

```go
package timetable_test

func TestCrossModuleReader_BuildProblem(t *testing.T) {
    // 1. Create test schema
    // 2. Seed: 3 teachers with availability, 5 subjects, 3 rooms with availability
    // 3. Create semester subjects linking to seeded subjects
    // 4. Call BuildProblem(ctx, semesterSubjects)
    // 5. Assert problem.Subjects has 5 entries
    // 6. Assert problem.Teachers has 3 entries with availability grids
    // 7. Assert problem.Rooms has 3 entries with availability grids
    // 8. Assert problem.Slots = AllSlots() (70 slots)
}

func TestCrossModuleReader_BuildProblem_WithTeacherAssignment(t *testing.T) {
    // Verify pre-assigned teacher IDs appear in problem.TeacherAssign map
}

func TestCrossModuleReader_BuildProblem_EmptyData(t *testing.T) {
    // No teachers/subjects/rooms -> empty problem, no error
}
```

### Step 2: Create `internal/timetable/semester_integration_test.go`

```go
func TestMain(m *testing.M) { /* testcontainer setup with ALL module repos */ }

func TestCreateSemester_Success(t *testing.T) {
    // POST /api/v1/timetable/semesters {name, year, term, start_date, end_date}
    // Assert 201
}

func TestListSemesters(t *testing.T)
func TestGetSemester_ByID(t *testing.T)

func TestSetSemesterSubjects(t *testing.T) {
    // 1. Seed subjects A, B, C
    // 2. Create semester
    // 3. POST /api/v1/timetable/semesters/{id}/subjects [A, B, C]
    // 4. GET semester -> verify subjects attached
}

func TestAssignTeacherToSemesterSubject(t *testing.T) {
    // 1. Seed teacher + subject, create semester with subject
    // 2. POST /api/v1/timetable/semesters/{id}/subjects/{subjectId}/teacher {teacher_id}
    // 3. Verify assignment persisted
}
```

### Step 3: Create `internal/timetable/scheduling_integration_test.go`

```go
func TestScheduleGeneration_EndToEnd(t *testing.T) {
    // FULL E2E FLOW:
    // 1. Seed 5 teachers with varied availability
    // 2. Seed 10 subjects (2-4 hours/week each)
    // 3. Seed 5 rooms (different capacities)
    // 4. Create semester, attach subjects, assign some teachers
    // 5. POST /api/v1/timetable/semesters/{id}/generate
    // 6. Assert 200/202 (generation started)
    // 7. GET /api/v1/timetable/semesters/{id}/schedule
    // 8. Verify: all subjects scheduled, no teacher conflicts, no room conflicts
}

func TestScheduleGeneration_NoConflicts(t *testing.T) {
    // Verify generated schedule has zero hard constraint violations:
    // - No teacher teaching two subjects in same slot
    // - No room used by two subjects in same slot
    // - All slots within teacher/room availability
}

func TestScheduleGeneration_RespectsTeacherAvailability(t *testing.T) {
    // Teacher available only Mon-Wed periods 1-5
    // All assignments for that teacher must be within availability
}

func TestApproveSchedule_Success(t *testing.T) {
    // 1. Generate schedule (status: pending)
    // 2. POST /api/v1/timetable/semesters/{id}/approve
    // 3. Assert status changed to approved
}

func TestApproveSchedule_AlreadyApproved_Returns409(t *testing.T)

func TestUpdateAssignment_ManualOverride(t *testing.T) {
    // 1. Generate schedule
    // 2. PUT /api/v1/timetable/assignments/{id} with new room/slot
    // 3. Verify assignment updated
}

func TestUpdateAssignment_ConflictDetection(t *testing.T) {
    // 1. Generate schedule
    // 2. Try moving assignment to slot already occupied by same teacher
    // 3. Assert conflict error
}

func TestScheduleGeneration_SmallScale_Under5Seconds(t *testing.T) {
    // 10 subjects, 5 teachers, 5 rooms
    // time.Now() before, assert elapsed < 5s
}
```

## Todo List
- [ ] Create cross_module_reader_test.go
- [ ] Create semester_integration_test.go
- [ ] Create scheduling_integration_test.go
- [ ] Implement BuildProblem tests (data assembly from 3 modules)
- [ ] Implement semester CRUD + subject management tests
- [ ] Implement full E2E schedule generation test
- [ ] Implement conflict detection tests
- [ ] Implement approval workflow tests
- [ ] Implement manual override + conflict test
- [ ] Verify all pass: `go test ./internal/timetable/... -v -race -timeout 60s`

## Success Criteria
- CrossModuleReader correctly assembles Problem from HR+Subject+Room data
- Semester CRUD + subject attachment works
- Schedule generation produces conflict-free timetable
- Teacher availability constraints are respected
- Approval state machine works (pending -> approved, no double-approve)
- Manual override detects and rejects conflicts
- Small-scale generation completes under 5 seconds

## Risk Assessment
| Risk | Impact | Mitigation |
|------|--------|------------|
| Scheduler non-deterministic (SA randomness) | Medium | Verify constraints, not exact assignments |
| Large fixture setup | Medium | Use helper functions from Phase 1 |
| Schedule generation timeout in CI | Low | Use small scale (10 subjects) and generous timeout |
| Cross-module import cycles | High | Timetable uses narrow interfaces, not direct imports |

## Security Considerations
- Verify only users with `timetable:schedule:write` can generate/approve
- Verify schedule data is tenant-scoped

## Next Steps
- Phase 6 uses scheduling flow for performance benchmarks (200 subjects)
