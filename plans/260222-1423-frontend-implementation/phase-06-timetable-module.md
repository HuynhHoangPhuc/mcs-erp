# Phase 06: Timetable Module

## Context
- [Plan](./plan.md) | Depends on: Phase 01, 02, 03, 04, 05
- Backend routes: `GET/POST /api/v1/timetable/semesters`, `GET /api/v1/timetable/semesters/{id}`, `POST /api/v1/timetable/semesters/{id}/subjects`, `POST /api/v1/timetable/semesters/{id}/subjects/{subjectId}/teacher`, `POST /api/v1/timetable/semesters/{id}/generate`, `GET /api/v1/timetable/semesters/{id}/schedule`, `POST /api/v1/timetable/semesters/{id}/approve`, `PUT /api/v1/timetable/assignments/{id}`

## Overview
- **Priority:** P1
- **Status:** pending
- **Effort:** 5h

Most complex module. Semester wizard flow, schedule grid visualization, conflict panel, admin review/approval workflow.

## Key Insights
- Semester statuses: DRAFT → SCHEDULING → REVIEW → APPROVED
- Schedule generation is synchronous (returns JSON directly, not SSE — per actual handler code)
- Schedule response: `{ semester_id, version, hard_violations, soft_penalty, generated_at, assignments[] }`
- Assignment: `{ id, semester_id, subject_id, teacher_id, room_id, day, period, version }`
- Manual assignment swap via PUT /assignments/{id}
- Need cross-module data: fetch teachers/subjects/rooms lists in parallel on page load, build ID→name lookup maps (Validation Session 1)
<!-- Updated: Validation Session 1 - Parallel queries for name resolution -->

## Requirements

### Functional
- Semester list page (name, dates, status badge) + create dialog
- Semester detail page — wizard-like flow:
  1. **Setup:** Select subjects for semester (multi-select from subject list)
  2. **Assign:** Assign teachers to subjects (dropdown per subject, teacher suggestion)
  3. **Generate:** Trigger schedule generation, show loading state
  4. **Review:** View generated timetable grid + conflict panel
  5. **Approve:** Approve or regenerate
- Timetable grid view: 6 columns (Mon-Sat) × 10 rows (periods 1-10)
  - Each cell shows: subject code, teacher name, room code
  - Color-coded by subject
  - Click cell to edit assignment (change teacher/room/slot)
- Conflict panel: list hard violations (red) and soft penalties (yellow)
- Teacher suggestion: button next to teacher dropdown → fetches available teachers

### Non-functional
- Grid must handle 200+ assignments visually
- Loading indicator during schedule generation (~25s)

## Architecture

### File Structure
```
packages/module-timetable/src/
├── index.ts
├── components/
│   ├── semester-list-page.tsx
│   ├── semester-form-dialog.tsx
│   ├── semester-detail-page.tsx      # Main wizard container
│   ├── semester-setup-step.tsx       # Step 1: select subjects
│   ├── semester-assign-step.tsx      # Step 2: assign teachers
│   ├── semester-review-step.tsx      # Step 3+4: view grid + approve
│   ├── timetable-grid.tsx            # Day × Period grid
│   ├── timetable-cell.tsx            # Single cell (subject/teacher/room)
│   ├── assignment-edit-dialog.tsx    # Edit assignment slot
│   ├── conflict-panel.tsx            # Hard/soft violations list
│   └── teacher-suggest-button.tsx    # Suggest teachers for subject
```

## Related Code Files

### Modify
- `web/packages/module-timetable/package.json` — add deps
- `web/packages/module-timetable/src/index.ts`
- Shell routes: `timetable.index.tsx`, `timetable.$semesterId.tsx`

### Create
- All files under `packages/module-timetable/src/components/`

## Implementation Steps

1. **Add deps**: standard module deps

2. **Create semester list** — DataTable: name, start/end date, status (badge colored by status), create button

3. **Create semester form dialog** — name, start_date (date picker), end_date (date picker). Zod: end > start.

4. **Create semester detail page** — Tab-based or step-based layout based on semester status:
   - DRAFT: show Setup + Assign steps
   - SCHEDULING: show loading spinner
   - REVIEW: show Review step (grid + conflicts + approve button)
   - APPROVED: show read-only grid

5. **Create `semester-setup-step.tsx`** — Fetch all subjects. Checkbox list or transfer list. Submit calls `POST /semesters/{id}/subjects` with selected IDs.

6. **Create `semester-assign-step.tsx`** — For each semester subject, show subject name + teacher dropdown. Dropdown populated from teachers list. "Suggest" button next to each. Submit calls `POST /semesters/{id}/subjects/{subjectId}/teacher` per subject.

7. **Create `timetable-grid.tsx`** — CSS Grid: 7 columns (header + Mon-Sat) × 11 rows (header + periods 1-10). Map assignments array to grid cells by `day` + `period`. Color each subject with deterministic hue from subject ID hash.

8. **Create `timetable-cell.tsx`** — Shows subject code, teacher name (truncated), room code. Click opens assignment edit dialog.

9. **Create `assignment-edit-dialog.tsx`** — Dropdowns for teacher, room, day, period. Submit calls `PUT /assignments/{id}`.

10. **Create `conflict-panel.tsx`** — List of hard violations (red icon + description) and soft penalties (yellow icon + description). Derived from `hard_violations` and `soft_penalty` counts in schedule response.

11. **Create `semester-review-step.tsx`** — Combines grid + conflict panel + approve/regenerate buttons

12. **Wire routes, build check**

## Todo List
- [ ] Create semester list page
- [ ] Create semester form dialog
- [ ] Create semester detail page with status-based steps
- [ ] Create subject selection step
- [ ] Create teacher assignment step with suggestion
- [ ] Create timetable grid component
- [ ] Create timetable cell component
- [ ] Create assignment edit dialog
- [ ] Create conflict panel
- [ ] Create review step with approve/regenerate
- [ ] Wire routes
- [ ] Verify build passes

## Success Criteria
- Full semester wizard flow: create → select subjects → assign teachers → generate → review → approve
- Timetable grid renders assignments correctly
- Assignment editing works via dialog
- Conflict panel shows violations
- Approve transitions semester to APPROVED status

## Risk Assessment
- **Grid complexity** — Many assignments in small cells; may need tooltip on hover for full info
- **Generation timeout** — 25s generation; must show loading state, not timeout the fetch
- **Cross-module data** — Need teacher/subject/room names to display in grid; requires parallel queries

## Security Considerations
- `timetable:timetable:write` for mutations
- `timetable:timetable:read` for viewing schedules

## Next Steps
- Phase 07 (Agent) — independent module
