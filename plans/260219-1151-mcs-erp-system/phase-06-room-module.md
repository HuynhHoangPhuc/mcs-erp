# Phase 06: Room Module

## Context Links
- [Parent Plan](./plan.md)
- Depends on: [Phase 03 - Auth & RBAC](./phase-03-auth-rbac.md)
- [Brainstorm Report](../reports/brainstorm-260219-1151-mcs-erp-architecture.md)

## Overview
- **Date:** 2026-02-19
- **Priority:** P2
- **Status:** Complete
- **Description:** Room bounded context: Room entity with capacity, equipment tags, and availability. Simplest domain module — straightforward CRUD. Provides room data to timetable scheduling.

## Key Insights
- Simplest module; follows exact same pattern as HR but with fewer entities
- Equipment tags (e.g., "projector", "lab_equipment", "computers") used by scheduler to match room to subject requirements
- Room availability reuses same weekly slot grid pattern as teacher availability
- Room capacity used as hard constraint in scheduling (class size <= room capacity)

## Requirements

### Functional
- Room CRUD: name, code, building, floor, capacity, equipment tags, status
- Room availability: weekly slot grid (same pattern as teacher availability)
- List rooms with filtering (building, min capacity, equipment tag) + pagination
- Domain events: RoomCreated, RoomUpdated, RoomAvailabilityUpdated

### Non-Functional
- Room list handles 200+ rooms with server-side pagination
- Equipment tags are free-form strings (no separate entity)

## Architecture

```
internal/room/
├── domain/
│   ├── room.go              # Room aggregate root
│   ├── availability.go      # RoomAvailability (weekly slot grid)
│   ├── events.go            # RoomCreated, RoomUpdated, RoomAvailabilityUpdated
│   └── repository.go        # RoomRepository interface
├── application/
│   ├── commands/
│   │   ├── create_room.go
│   │   ├── update_room.go
│   │   └── set_room_availability.go
│   └── queries/
│       ├── list_rooms.go
│       ├── get_room.go
│       └── get_room_availability.go
├── infrastructure/
│   └── postgres_room_repo.go
├── delivery/
│   ├── room_handler.go          # /api/rooms
│   └── room_availability_handler.go # /api/rooms/:id/availability
├── tools/
│   ├── search_rooms.go          # AI tool
│   └── check_room_availability.go # AI tool
└── module.go
```

### Frontend
```
web/packages/module-room/src/
├── routes.ts
├── pages/
│   ├── room-list-page.tsx
│   ├── room-detail-page.tsx
│   └── room-form-page.tsx
├── components/
│   ├── room-table.tsx
│   ├── room-form.tsx
│   ├── equipment-tags.tsx
│   └── room-availability-grid.tsx
└── queries/
    ├── use-rooms.ts
    ├── use-room.ts
    └── use-room-availability.ts
```

## Related Code Files

### Files to Create

**Backend Domain:**
- `internal/room/domain/room.go` — Room: ID, Code, Name, Building, Floor int, Capacity int, Equipment []string, IsActive, CreatedAt, UpdatedAt
- `internal/room/domain/availability.go` — RoomAvailability: RoomID, Slots []TimeSlot
- `internal/room/domain/events.go` — RoomCreated, RoomUpdated, RoomAvailabilityUpdated
- `internal/room/domain/repository.go` — RoomRepository interface

**Backend Application:**
- `internal/room/application/commands/create_room.go`
- `internal/room/application/commands/update_room.go`
- `internal/room/application/commands/set_room_availability.go`
- `internal/room/application/queries/list_rooms.go` — filter: building, min_capacity, equipment, status
- `internal/room/application/queries/get_room.go`
- `internal/room/application/queries/get_room_availability.go`

**Backend Infrastructure:**
- `internal/room/infrastructure/postgres_room_repo.go`

**Backend Delivery:**
- `internal/room/delivery/room_handler.go` — CRUD /api/rooms
- `internal/room/delivery/room_availability_handler.go` — GET/PUT /api/rooms/:id/availability

**AI Tools:**
- `internal/room/tools/search_rooms.go` — search_rooms(query, building, min_capacity, equipment)
- `internal/room/tools/check_room_availability.go` — check_room_availability(room_id, day, period)

**Module:**
- `internal/room/module.go` — dependencies: `["core"]`

**SQL:**
- `sqlc/queries/room/rooms.sql`
- `sqlc/queries/room/room_availability.sql`
- `migrations/room/000001_create_rooms_table.up.sql`
- `migrations/room/000002_create_room_availability_table.up.sql`

**Frontend:**
- `web/packages/module-room/src/routes.ts`
- `web/packages/module-room/src/pages/room-list-page.tsx`
- `web/packages/module-room/src/pages/room-detail-page.tsx`
- `web/packages/module-room/src/pages/room-form-page.tsx`
- `web/packages/module-room/src/components/room-table.tsx`
- `web/packages/module-room/src/components/room-form.tsx`
- `web/packages/module-room/src/components/equipment-tags.tsx`
- `web/packages/module-room/src/components/room-availability-grid.tsx`
- `web/packages/module-room/src/queries/use-rooms.ts`
- `web/packages/module-room/src/queries/use-room.ts`
- `web/packages/module-room/src/queries/use-room-availability.ts`

## Implementation Steps

1. **SQL migrations**
   ```sql
   -- rooms
   CREATE TABLE rooms (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       code VARCHAR(20) NOT NULL UNIQUE,
       name VARCHAR(255) NOT NULL,
       building VARCHAR(100),
       floor INT DEFAULT 0,
       capacity INT NOT NULL DEFAULT 30,
       equipment TEXT[] NOT NULL DEFAULT '{}',
       is_active BOOLEAN NOT NULL DEFAULT true,
       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
       updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
   );
   -- room_availability (same pattern as teacher_availability)
   CREATE TABLE room_availability (
       room_id UUID REFERENCES rooms(id) ON DELETE CASCADE,
       day INT NOT NULL CHECK (day BETWEEN 0 AND 6),
       period INT NOT NULL CHECK (period >= 0),
       available BOOLEAN NOT NULL DEFAULT true,
       PRIMARY KEY (room_id, day, period)
   );
   ```

2. **Domain entity** — Room with `NewRoom(code, name, building, floor, capacity, equipment)`, validation: capacity > 0, code non-empty

3. **Availability** — reuse same `TimeSlot{Day, Period, Available}` pattern from HR; consider extracting to `pkg/erptypes/` if identical

4. **Repository interface** — `RoomRepository`: Save, FindByID, List(filters, page), Update, Delete, SaveAvailability, GetAvailability

5. **sqlc queries** — CRUD + filtered list (building, capacity >= min, equipment array overlap `&&`)
   ```sql
   -- name: ListRooms :many
   SELECT * FROM rooms
   WHERE (sqlc.narg('building')::VARCHAR IS NULL OR building = sqlc.narg('building'))
     AND (sqlc.narg('min_capacity')::INT IS NULL OR capacity >= sqlc.narg('min_capacity'))
     AND (sqlc.narg('equipment')::TEXT[] IS NULL OR equipment && sqlc.narg('equipment'))
     AND is_active = true
   ORDER BY name
   LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
   ```

6. **Command + query handlers** — standard pattern, same as HR module

7. **HTTP handlers** — CRUD at `/api/rooms`, availability at `/api/rooms/:id/availability`

8. **AI tools**
   - `search_rooms`: filter by name/building/capacity/equipment, return list
   - `check_room_availability`: given room_id + day + period, return available/unavailable

9. **Module registration** — `room.Module`, dependencies: `["core"]`

10. **Frontend: room table** — TanStack Table: code, name, building, floor, capacity, equipment (badge list)

11. **Frontend: room form** — code, name, building, floor, capacity (number input), equipment (tag input)

12. **Frontend: availability grid** — reuse same grid component pattern from HR (extract shared component to `packages/ui` if identical)

## Todo List
- [x] SQL migrations (rooms, room_availability)
- [x] Domain entities (Room, RoomAvailability)
- [x] Domain events
- [x] Repository interface
- [x] sqlc queries (with equipment array filter)
- [x] Infrastructure repo
- [x] Command handlers (create, update, set availability)
- [x] Query handlers (list, get, availability)
- [x] HTTP handlers
- [x] AI tools (search_rooms, check_availability)
- [x] Room module registration
- [x] Frontend routes + pages
- [x] Room table + form components
- [x] Equipment tags component
- [x] Room availability grid (reuse from HR if possible)
- [x] TanStack Query hooks
- [x] Extract shared availability grid to packages/ui (if HR pattern identical)

## Success Criteria
- Room CRUD works via API
- Room list filters by building, capacity, and equipment
- Availability grid saves/loads per room
- AI tools return correct results
- Frontend room list and form work end-to-end

## Risk Assessment
- **Equipment tag consistency**: Free-form tags may have typos. Add autocomplete from existing tags in frontend.
- **Availability grid reuse**: If HR and Room grids diverge, maintain separately rather than premature abstraction.

## Security Considerations
- All endpoints require `room:read` / `room:write` permissions
- Room data is not PII; lower sensitivity than HR data

## Next Steps
- Phase 7 (Timetable) uses room capacity + availability as scheduling constraints
- Phase 8 (AI Agent) uses room search + availability tools
