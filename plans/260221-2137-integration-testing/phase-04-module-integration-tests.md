---
phase: 4
title: "Module Integration Tests"
status: pending
priority: P1
effort: 6h
depends_on: [1]
---

# Phase 4: Module Integration Tests

## Context Links
- [hr/module.go](/internal/hr/module.go) -- HR module (teachers, departments, availability)
- [subject/module.go](/internal/subject/module.go) -- Subject module (subjects, categories, prerequisites)
- [room/module.go](/internal/room/module.go) -- Room module (rooms, availability)
- [agent/module.go](/internal/agent/module.go) -- Agent module (AI chat, conversations)
- [subject/domain/prerequisite_graph.go](/internal/subject/domain/prerequisite_graph.go) -- DAG cycle detection

## Overview
Integration tests for each individual module: HR, Subject, Room, Agent. Each test validates CRUD operations, domain logic, event publishing, and error handling through the HTTP API against real Postgres.

## Key Insights
- All modules share the same pattern: `module.go` -> `delivery/` handlers -> `infrastructure/` repos -> `domain/` entities
- HR availability: 7 days x 10 periods grid
- Subject prerequisites: directed acyclic graph with cycle detection via DFS
- Room availability: same 7x10 grid pattern as HR
<!-- Updated: Validation Session 1 - Use mock LLM provider -->
- Agent: requires Docker Compose Redis for cache; uses MockLLMProvider from testutil (no real LLM calls)
- Permission format: `module:resource:action` (e.g., `hr:teacher:write`)

## Requirements

### Functional per Module

**HR Module:**
- Teacher CRUD (create, list, get, update)
- Department CRUD (create, list, get, update, delete)
- Availability grid set/get (7 days x 10 periods)
- Teacher with non-existent department returns error

**Subject Module:**
- Subject CRUD (create, list, get)
- Category CRUD (create, list, get)
- Prerequisite add/remove
- Cycle detection: adding A->B->C->A rejected
- Topological sort returns valid ordering
- Prerequisite chain retrieval (transitive deps)
- Optimistic locking conflict on concurrent prerequisite update

**Room Module:**
- Room CRUD (create, list, get)
- Room availability set/get (7 days x 10 periods)
- Equipment tagging

**Agent Module:**
- Conversation CRUD (create, list, get, delete)
- Message persistence within conversation
- Permission-filtered tool listing
- SSE chat endpoint (validate event stream format)
- Redis cache hit/miss behavior

### Non-functional
- Each module test suite < 15s
- Table-driven tests for CRUD validation

## Architecture
Test files alongside source, split by concern:

```
internal/hr/
  teacher_integration_test.go      -- teacher CRUD + availability
  department_integration_test.go   -- department CRUD

internal/subject/
  subject_integration_test.go      -- subject + category CRUD
  prerequisite_integration_test.go -- DAG operations + cycle detection

internal/room/
  room_integration_test.go         -- room CRUD + availability

internal/agent/
  conversation_integration_test.go -- conversation CRUD + messages
  chat_sse_integration_test.go     -- SSE streaming test
```

## Related Code Files

### Files to Create
- `internal/hr/teacher_integration_test.go` (~160 lines)
- `internal/hr/department_integration_test.go` (~100 lines)
- `internal/subject/subject_integration_test.go` (~120 lines)
- `internal/subject/prerequisite_integration_test.go` (~180 lines)
- `internal/room/room_integration_test.go` (~140 lines)
- `internal/agent/conversation_integration_test.go` (~140 lines)
- `internal/agent/chat_sse_integration_test.go` (~80 lines)

### Files to Reference
- All `delivery/*_handler.go` files for API endpoints
- All `domain/repository.go` files for expected operations
- `internal/hr/domain/availability.go` -- availability grid structure
- `internal/subject/domain/prerequisite.go` -- PrerequisiteEdge type

## Implementation Steps

### Step 1: HR Module Tests

#### `internal/hr/teacher_integration_test.go`
```go
package hr_test

func TestMain(m *testing.M) { /* testcontainer setup */ }

func TestCreateTeacher_Success(t *testing.T) {
    // POST /api/v1/teachers with valid payload
    // Assert 201, response has ID, name, email, qualifications
}

func TestCreateTeacher_MissingFields_Returns400(t *testing.T)

func TestListTeachers_Paginated(t *testing.T) {
    // Seed 5 teachers, GET /api/v1/teachers?offset=0&limit=3
    // Assert 3 returned, total=5
}

func TestGetTeacher_ByID(t *testing.T)
func TestGetTeacher_NotFound_Returns404(t *testing.T)
func TestUpdateTeacher_Success(t *testing.T)

func TestSetAvailability_FullGrid(t *testing.T) {
    // PUT /api/v1/teachers/{id}/availability
    // Body: array of {day, period, is_available} for 7x10 grid
    // Assert 200
    // GET /api/v1/teachers/{id}/availability -> verify grid
}

func TestSetAvailability_InvalidTeacher_Returns404(t *testing.T)
```

#### `internal/hr/department_integration_test.go`
```go
func TestDepartmentCRUD(t *testing.T) {
    // Table-driven: create, list, get, update, delete
    // Verify delete returns 204
    // Verify deleted department returns 404
}
```

### Step 2: Subject Module Tests

#### `internal/subject/subject_integration_test.go`
```go
func TestCreateSubject_Success(t *testing.T)
func TestCreateSubject_DuplicateCode_Returns409(t *testing.T)
func TestListSubjects_WithCategoryFilter(t *testing.T)
func TestCreateCategory_Success(t *testing.T)
func TestListCategories(t *testing.T)
```

#### `internal/subject/prerequisite_integration_test.go`
```go
func TestAddPrerequisite_Success(t *testing.T) {
    // Create subjects A, B
    // POST add prerequisite A -> B
    // GET prerequisites of A -> contains B
}

func TestAddPrerequisite_CycleDetection(t *testing.T) {
    // Create A -> B -> C
    // Try adding C -> A
    // Assert 409/422 with cycle error
}

func TestTopologicalSort_ValidOrdering(t *testing.T) {
    // Create DAG: A->B, A->C, B->D, C->D
    // GET topological order
    // Assert D before B,C; B,C before A
}

func TestPrerequisiteChain_TransitiveDeps(t *testing.T) {
    // A->B->C->D
    // Get chain for A -> [B, C, D]
}

func TestRemovePrerequisite_Success(t *testing.T)

func TestOptimisticLocking_ConcurrentUpdate(t *testing.T) {
    // 1. Read prerequisite with version=1
    // 2. Update with version=1 (succeeds)
    // 3. Update again with version=1 (fails with conflict)
}
```

### Step 3: Room Module Tests

#### `internal/room/room_integration_test.go`
```go
func TestCreateRoom_Success(t *testing.T) {
    // POST /api/v1/rooms {name, capacity, equipment: ["projector","whiteboard"]}
    // Assert 201
}

func TestListRooms_FilterByCapacity(t *testing.T)
func TestGetRoom_ByID(t *testing.T)

func TestSetRoomAvailability_FullGrid(t *testing.T) {
    // Same pattern as teacher availability
}

func TestRoomEquipment_TaggingAndFilter(t *testing.T) {
    // Create rooms with different equipment
    // Filter by equipment type
}
```

### Step 4: Agent Module Tests

#### `internal/agent/conversation_integration_test.go`
```go
func TestMain(m *testing.M) {
    // Start Postgres + Redis testcontainers
}

func TestCreateConversation_Success(t *testing.T) {
    // POST /api/v1/agent/conversations {title: "Test Chat"}
    // Assert 201
}

func TestListConversations_UserScoped(t *testing.T) {
    // User A creates conversation
    // User B lists conversations -> does not see A's
}

func TestGetConversation_ByID(t *testing.T)
func TestDeleteConversation_Success(t *testing.T)
func TestDeleteConversation_NotOwner_Returns403(t *testing.T)
```

#### `internal/agent/chat_sse_integration_test.go`
```go
func TestChatSSE_StreamFormat(t *testing.T) {
    // 1. Wire agent module with MockLLMProvider from testutil
    // 2. POST /api/v1/agent/chat with Accept: text/event-stream
    // 3. Read SSE events, verify format: "data: {...}\n\n"
    // 4. Assert mock response chunks arrive correctly
    // 5. Verify conversation + messages persisted in DB
}
```

## Todo List
- [ ] Create HR teacher integration tests
- [ ] Create HR department integration tests
- [ ] Create Subject CRUD integration tests
- [ ] Create Prerequisite DAG integration tests (cycle detection, topo sort)
- [ ] Create Room integration tests
- [ ] Create Agent conversation integration tests
- [ ] Create Agent SSE chat test
- [ ] Verify all pass: `go test ./internal/hr/... ./internal/subject/... ./internal/room/... ./internal/agent/... -v -race`

## Success Criteria
- All CRUD operations return correct status codes and response bodies
- Prerequisite cycle detection correctly rejects circular dependencies
- Topological sort produces valid ordering
- Optimistic locking catches concurrent conflicts
- Availability grids persist and retrieve correctly (7x10)
- Agent conversations are user-scoped (IDOR prevention)
- All tests pass with `-race`

## Risk Assessment
| Risk | Impact | Mitigation |
|------|--------|------------|
| Agent LLM calls in tests | Medium | Skip LLM tests or use mock provider |
| Redis availability | Low | testcontainers-go redis module |
| Large test files | Medium | Split by concern (teacher vs dept, subject vs prereq) |

## Security Considerations
- Verify IDOR: user A cannot access user B's conversations
- Verify permission checks on all CRUD endpoints
- Equipment/tag fields should not accept arbitrary HTML/scripts

## Next Steps
- Phase 5 uses fixtures from all modules for cross-module timetable tests
