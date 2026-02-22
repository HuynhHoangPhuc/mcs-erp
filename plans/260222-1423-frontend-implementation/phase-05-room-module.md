# Phase 05: Room Module

## Context
- [Plan](./plan.md) | [Phase 03](./phase-03-hr-module.md)
- Depends on: Phase 01, Phase 02
- Backend routes: `GET/POST /api/v1/rooms`, `GET/PUT /api/v1/rooms/{id}`, `GET/PUT /api/v1/rooms/{id}/availability`

## Overview
- **Priority:** P2
- **Status:** pending
- **Effort:** 3h

Build room management UI. Simplest module — CRUD + availability grid (reuses pattern from HR). Includes equipment tagging.

## Key Insights
- Room: id, name, code, building, floor, capacity, equipment[], is_active
- Availability: same 7-day × 10-period grid as teachers
- Equipment: free-form string array (projector, whiteboard, lab-equipment, etc.)
- Filtering: building, min capacity, equipment tag

## Requirements

### Functional
- Room list with filtering (building, capacity range, equipment tag) + pagination
- Room create/edit form (name, code, building, floor, capacity, equipment tags)
- Room detail page with availability grid
- Equipment shown as badges, editable as tag input

### Non-functional
- Reuse `<AvailabilityGrid>` from `@mcs-erp/ui`

## Architecture

### File Structure
```
packages/module-room/src/
├── index.ts
├── components/
│   ├── room-list-page.tsx
│   ├── room-form-dialog.tsx
│   ├── room-columns.tsx
│   ├── room-detail-page.tsx
│   └── room-filters.tsx        # Building select, capacity input, equipment select
```

## Related Code Files

### Modify
- `web/packages/module-room/package.json` — add deps
- `web/packages/module-room/src/index.ts`
- Shell route stubs: `rooms.index.tsx`, `rooms.$roomId.tsx`

### Create
- All files under `packages/module-room/src/components/`

## Implementation Steps

1. **Add deps**: same as HR module pattern
2. **Create room CRUD** — columns, form dialog, list page, detail page (mirrors HR teacher pattern)
3. **Room form** — zod schema: `{ name, code, building, floor: number, capacity: number, equipment: string[] }`
4. **Room filters** — Building text input, capacity min/max, equipment multi-select
5. **Room detail** — Info card + `<AvailabilityGrid>` (same component as teacher availability)
6. **Wire routes** in shell
7. **Build check**

## Todo List
- [ ] Add dependencies
- [ ] Create room columns + list page
- [ ] Create room form dialog with equipment tags
- [ ] Create room filter bar
- [ ] Create room detail page with availability grid
- [ ] Wire routes in shell
- [ ] Verify build passes

## Success Criteria
- Room list with filtering and pagination
- Create/edit room works with equipment tags
- Availability grid saves correctly

## Risk Assessment
- Minimal — follows established HR pattern

## Security Considerations
- `room:room:write` permission required for mutations

## Next Steps
- Phase 06 (Timetable) — depends on HR + Subject + Room data
