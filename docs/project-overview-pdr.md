# MCS-ERP: Project Overview & Product Development Requirements

## Executive Summary

**MCS-ERP** is a multi-tenant, agentic-first Enterprise Resource Planning system designed for academic institutions. The MVP focuses on **course scheduling** (timetabling) with AI agent assistance, built as a modular monolith using Domain-Driven Design principles.

**Status:** MVP complete (6 modules, full backend + frontend scaffold)
**Target Users:** Academic administrators, teachers, scheduling coordinators
**Deployment Model:** Cloud-native (Kubernetes-ready)

---

## Product Vision

### Problem Statement
Academic institutions struggle to create conflict-free course schedules across:
- Multiple teachers with varying availability
- Limited classroom resources (rooms with capacity/equipment constraints)
- Complex prerequisite chains for subjects
- Manual processes prone to errors and inefficiency

### Solution
Provide an intuitive, AI-powered platform that:
1. Automates schedule generation using constraint optimization
2. Allows AI agent to interactively answer scheduling questions
3. Supports multi-tenant deployments for different institutions
4. Scales to handle 1000s of courses and 100s of rooms

### Vision Statement
"Be the go-to scheduling solution for academic institutions by combining powerful constraint solvers with intuitive AI assistance."

---

## Core Functionality (MVP)

### 1. Authentication & Authorization
- User login with email + password (JWT-based)
- Role-based access control (RBAC) with granular permissions
- Support for admin, coordinator, teacher roles
- Tenant isolation (no cross-tenant data access)

### 2. Human Resources Module
- Teacher management (name, email, department, active status)
- Department management (organization structure)
- Teacher availability tracking (7 days × 10 time periods per day)

### 3. Subject Catalog
- Subject CRUD (code, name, credits, category)
- Prerequisite management with cycle detection
- Category-based organization

### 4. Room Management
- Classroom/lab CRUD (name, code, building, capacity, equipment)
- Room availability tracking (7 days × 10 periods)
- Capacity and equipment constraints for scheduling

### 5. Timetable Scheduling
- Semester management (create, list, configure)
- Subject-teacher assignments
- Automated schedule generation (greedy + simulated annealing)
- Admin approval workflow (DRAFT → APPROVED)
- Manual schedule adjustments (override assignments)

### 6. AI Agent
- Multi-LLM support (Claude, OpenAI, Ollama)
- Tool registry for cross-module extensions
- Conversation management (create, list, delete)
- Real-time SSE streaming for agent responses
- Rule-based suggestions (no LLM call needed)

---

## Technical Architecture

### Backend Stack
- **Language:** Go 1.22+
- **Framework:** stdlib net/http (REST API)
- **Architecture:** Modular monolith with DDD layers
- **Database:** PostgreSQL 16 (schema-per-tenant)
- **Caching:** Redis 7 (optional, agent message cache)
- **Internal Communication:** gRPC with protobuf
- **Event Bus:** Watermill (in-process, extensible)

### Frontend Stack
- **Framework:** React 19 with TypeScript
- **Routing:** TanStack Router (file-based)
- **State Management:** TanStack Query + TanStack Form
- **UI Components:** shadcn/ui + Tailwind CSS
- **Monorepo:** Turborepo + pnpm
- **Modules:** Separate feature packages (hr, subject, room, timetable, agent)

### Deployment
- **Development:** Docker Compose (PostgreSQL + Redis)
- **Production:** Kubernetes with managed PostgreSQL/Redis
- **Frontend:** Static assets on CDN (Vercel/Netlify)
- **API Gateway:** Load balancer with TLS termination

---

## Functional Requirements

### FR1: User Authentication
- Users can log in with email + password
- JWT tokens issued with 24h expiry
- Refresh token endpoint for session extension
- Logout invalidates session

**Acceptance Criteria:**
- Valid credentials return 200 with JWT token
- Invalid credentials return 401
- Expired tokens return 401 on protected endpoints
- JWT payload includes user_id, roles, permissions

### FR2: Tenant Isolation
- Each institution is a separate tenant with isolated data
- Tenant determined by subdomain (tenant-abc.example.com) or X-Tenant-ID header
- Users can only access their own tenant's data

**Acceptance Criteria:**
- Queries from User A only return User A's tenant data
- User B cannot access User A's data even with valid JWT
- All SQL queries wrap `SET LOCAL search_path = tenant_schema`

### FR3: Role-Based Access Control
- System supports predefined roles (admin, coordinator, teacher)
- Roles have granular permissions (module:resource:action format)
- Permission checks enforced on all protected endpoints

**Acceptance Criteria:**
- Admin role has all permissions
- Coordinator role can read/write scheduling data
- Teacher role can view their schedule (read-only)
- Missing permission returns 403 Forbidden

### FR4: Teacher Management
- Create, read, update teacher records
- Assign teachers to departments
- Set teacher availability by day and period

**Acceptance Criteria:**
- All CRUD operations work via REST API
- Availability stored as 70 boolean slots (7 days × 10 periods)
- Teachers can only modify their own availability (with proper auth)

### FR5: Subject Management
- Create, read, update subject records with prerequisites
- Validate no circular prerequisites (DAG)
- Query prerequisite chains (transitive dependencies)

**Acceptance Criteria:**
- Adding cyclic prerequisite returns error
- GetPrerequisiteChain returns all transitive prerequisites
- Subjects can have multiple prerequisites

### FR6: Room Management
- Create, read, update room records with capacity and equipment
- Set room availability by day and period
- Query room constraints (capacity, equipment) during scheduling

**Acceptance Criteria:**
- Room capacity is required (>0)
- Equipment stored as array (e.g., ["projector", "whiteboard"])
- Availability same format as teacher (70 slots)

### FR7: Schedule Generation
- Generate conflict-free timetables for a semester
- Assign subjects to teachers, rooms, time slots
- Respect constraints: teacher availability, room capacity, prerequisites
- Return draft schedule for admin review

**Acceptance Criteria:**
- Generated schedule has no conflicts (one teacher/room per slot)
- Respects all teacher availability constraints
- Respects all room capacity constraints
- Respects prerequisite prerequisites (no same-slot prerequisites)
- Completes within 30s timeout or returns error

### FR8: Schedule Approval
- Admin reviews generated schedule (DRAFT status)
- Approve schedule (→ APPROVED status)
- Make manual adjustments to individual assignments

**Acceptance Criteria:**
- Only DRAFT schedules can be approved
- APPROVED schedule is immutable (read-only)
- Manual updates only allowed in DRAFT status
- Update assignment returns updated assignment object

### FR9: AI Agent Chat
- Users can start conversations with AI agent
- Agent can answer questions about schedules
- Real-time streaming of agent responses (SSE)
- Conversation history persisted

**Acceptance Criteria:**
- POST /api/v1/agent/chat accepts message and returns SSE stream
- Agent response chunks are delivered in real-time
- Conversation history retrievable via GET /api/v1/agent/conversations/{id}
- Tool calling works (agent can invoke scheduling tools)

### FR10: Multi-LLM Support
- System can switch between multiple LLM providers
- Providers: Claude, OpenAI, Ollama
- Configuration via environment variables

**Acceptance Criteria:**
- Setting AI_PROVIDER=claude uses Claude API
- Setting AI_PROVIDER=openai uses OpenAI API
- Agent response quality consistent across providers

---

## Non-Functional Requirements

### NFR1: Performance
- **Response Time:** API responses < 200ms for reads, < 500ms for writes (p95)
- **Throughput:** Support 100+ concurrent users
- **Scheduling:** Generate schedule for 500 courses in < 30s

**Metrics:**
- Database connection pool: 20 connections
- Query optimization: Indexes on foreign keys and search columns
- Caching: Redis optional for agent message cache

### NFR2: Scalability
- **Horizontal:** Stateless HTTP handlers (scale with load balancer)
- **Vertical:** Increase CPU/memory and database resources
- **Multi-tenancy:** Support 1000+ tenants on single deployment

**Metrics:**
- Schema-per-tenant isolation enables per-tenant scaling
- No hard limits on number of tenants (limited by DB resources)

### NFR3: Security
- **Authentication:** JWT with HMAC-SHA256 signing
- **Authorization:** RBAC with granular permissions
- **Data Protection:** Tenant isolation at database level (SET LOCAL search_path)
- **Input Validation:** All inputs validated before use
- **SQL Safety:** Parameterized queries (no string concatenation)

**Metrics:**
- No SQL injection vulnerabilities
- No unauthorized cross-tenant data access
- Passwords hashed with bcrypt (never stored plaintext)

### NFR4: Reliability
- **Uptime:** 99.5% availability
- **Data Durability:** PostgreSQL ACID transactions
- **Backup:** Daily backups with point-in-time recovery
- **Graceful Degradation:** Agent chat optional (rest of app works if Redis down)

**Metrics:**
- No data loss in case of pod restart
- Automatic connection pool recovery on DB reconnect

### NFR5: Maintainability
- **Code Structure:** DDD layers with clear separation of concerns
- **Testing:** Unit tests for domain logic, integration tests for API
- **Documentation:** API docs (OpenAPI/Swagger), architecture docs, code standards
- **Monitoring:** Structured logging with slog

**Metrics:**
- Code coverage target: 80%+
- All public APIs documented with examples

### NFR6: Usability
- **UI:** Intuitive React frontend with TanStack components
- **Accessibility:** WCAG 2.1 AA compliance
- **Mobile:** Responsive design (works on tablet/mobile)
- **Error Messages:** Clear, actionable error messages

**Metrics:**
- All forms have validation feedback
- All errors return descriptive JSON messages

---

## Data Model (Core Entities)

### Users & Authentication
```
users
├── id (UUID)
├── email (string, unique)
├── password_hash (string)
├── roles[] (foreign key → roles)
└── created_at, updated_at

roles
├── id (UUID)
├── name (string, unique)
├── permissions[] (foreign key → permissions)
└── created_at

permissions
├── id (UUID)
├── name (string, format: module:resource:action)
└── description (string)
```

### HR Module
```
teachers
├── id (UUID)
├── name (string)
├── email (string)
├── department_id (UUID → departments)
├── is_active (boolean)
└── created_at, updated_at

departments
├── id (UUID)
├── name (string)
├── description (string)
└── created_at, updated_at

teacher_availability
├── teacher_id (UUID → teachers)
├── day (0-6, Monday-Sunday)
├── period (1-10, time slots)
├── is_available (boolean)
└── PRIMARY KEY (teacher_id, day, period)
```

### Subject Module
```
subjects
├── id (UUID)
├── code (string, unique)
├── name (string)
├── credits (integer)
├── category_id (UUID → categories)
└── created_at, updated_at

categories
├── id (UUID)
├── name (string)
└── created_at, updated_at

subject_prerequisites
├── subject_id (UUID → subjects)
├── prerequisite_subject_id (UUID → subjects)
└── PRIMARY KEY (subject_id, prerequisite_subject_id)
```

### Room Module
```
rooms
├── id (UUID)
├── name (string)
├── code (string, unique)
├── building (string)
├── floor (integer)
├── capacity (integer)
├── equipment[] (string array)
├── is_active (boolean)
└── created_at, updated_at

room_availability
├── room_id (UUID → rooms)
├── day (0-6)
├── period (1-10)
├── is_available (boolean)
└── PRIMARY KEY (room_id, day, period)
```

### Timetable Module
```
semesters
├── id (UUID)
├── name (string)
├── start_date (date)
├── end_date (date)
├── is_active (boolean)
└── created_at, updated_at

semester_subjects
├── id (UUID)
├── semester_id (UUID → semesters)
├── subject_id (UUID → subjects)
├── teacher_id (UUID → teachers)
├── requested_hours (integer)
└── created_at, updated_at

schedules
├── id (UUID)
├── semester_id (UUID → semesters)
├── status (DRAFT|APPROVED)
└── created_at, updated_at

assignments
├── id (UUID)
├── schedule_id (UUID → schedules)
├── subject_id (UUID → subjects)
├── teacher_id (UUID → teachers)
├── room_id (UUID → rooms)
├── day (0-6)
├── period (1-10)
└── created_at, updated_at
```

### Agent Module
```
conversations
├── id (UUID)
├── user_id (UUID → users)
├── title (string)
└── created_at, updated_at

messages
├── id (UUID)
├── conversation_id (UUID → conversations)
├── role (USER|ASSISTANT)
├── content (text)
└── created_at

tools
├── id (UUID)
├── name (string)
├── description (text)
├── schema (JSON, function signature)
└── created_at
```

---

## API Endpoints

### Authentication
```
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
```

### Users & Roles
```
POST   /api/v1/users
GET    /api/v1/users
GET    /api/v1/users/{id}
POST   /api/v1/users/{id}/roles
POST   /api/v1/roles
GET    /api/v1/roles
GET    /api/v1/roles/{id}
DELETE /api/v1/roles/{id}
```

### HR Module
```
POST   /api/v1/teachers
GET    /api/v1/teachers
GET    /api/v1/teachers/{id}
PUT    /api/v1/teachers/{id}
GET    /api/v1/teachers/{id}/availability
PUT    /api/v1/teachers/{id}/availability
POST   /api/v1/departments
GET    /api/v1/departments
GET    /api/v1/departments/{id}
PUT    /api/v1/departments/{id}
DELETE /api/v1/departments/{id}
```

### Subject Module
```
POST   /api/v1/subjects
GET    /api/v1/subjects
GET    /api/v1/subjects/{id}
PUT    /api/v1/subjects/{id}
POST   /api/v1/categories
GET    /api/v1/categories
GET    /api/v1/categories/{id}
POST   /api/v1/subjects/{id}/prerequisites
DELETE /api/v1/subjects/{id}/prerequisites/{prereqId}
GET    /api/v1/subjects/{id}/prerequisites
GET    /api/v1/subjects/{id}/prerequisite-chain
```

### Room Module
```
POST   /api/v1/rooms
GET    /api/v1/rooms
GET    /api/v1/rooms/{id}
PUT    /api/v1/rooms/{id}
GET    /api/v1/rooms/{id}/availability
PUT    /api/v1/rooms/{id}/availability
```

### Timetable Module
```
POST   /api/v1/timetable/semesters
GET    /api/v1/timetable/semesters
GET    /api/v1/timetable/semesters/{id}
POST   /api/v1/timetable/semesters/{id}/subjects
POST   /api/v1/timetable/semesters/{id}/subjects/{subjectId}/teacher
POST   /api/v1/timetable/semesters/{id}/generate
GET    /api/v1/timetable/semesters/{id}/schedule
POST   /api/v1/timetable/semesters/{id}/approve
PUT    /api/v1/timetable/assignments/{id}
```

### Agent Module
```
POST   /api/v1/agent/chat (SSE)
GET    /api/v1/agent/conversations
POST   /api/v1/agent/conversations
GET    /api/v1/agent/conversations/{id}
PATCH  /api/v1/agent/conversations/{id}
DELETE /api/v1/agent/conversations/{id}
GET    /api/v1/agent/suggestions
```

---

## Success Metrics

### Adoption
- [ ] 10+ institutions deployed (beta users)
- [ ] 100+ administrators using system
- [ ] 1000+ courses scheduled per month

### Reliability
- [ ] 99.5% uptime
- [ ] Zero data loss incidents
- [ ] < 1% error rate on schedule generation

### Performance
- [ ] API p95 response time < 200ms
- [ ] Schedule generation < 30s for 500 courses
- [ ] Support 100+ concurrent users per tenant

### User Satisfaction
- [ ] > 80% of users rate system "very easy to use"
- [ ] < 5% of generated schedules require manual adjustment
- [ ] AI agent resolves 70%+ of user questions without escalation

---

## Implementation Phases (Completed)

### Phase 1: Core Infrastructure & Authentication ✅
- Database schema, tenant setup, JWT authentication, RBAC
- Module registry, DDD layer structure
- Core API (auth, users, roles)

### Phase 2: HR Module ✅
- Teachers, departments, availability tracking
- REST API for HR operations

### Phase 3: Subject Module ✅
- Subject catalog, prerequisites with cycle detection
- REST API for subject management

### Phase 4: Room Module ✅
- Room management, capacity tracking, availability
- REST API for room operations

### Phase 5: Timetable Scheduling Engine ✅
- Semester management, schedule generation algorithm
- Greedy + simulated annealing optimization
- Admin approval workflow

### Phase 6: AI Agent ✅
- Multi-LLM provider integration
- Tool registry for cross-module capabilities
- Conversation management, SSE streaming

### Phase 7: Frontend Scaffold ✅
- React 19 + TanStack setup
- Module placeholder components
- API client generation

### Phase 8: Testing & Documentation ✅
- Unit tests for core logic
- API documentation
- System architecture docs
- Code standards documentation

---

## Known Limitations & Future Work

### MVP Limitations
1. **Single scheduling algorithm:** Only greedy + simulated annealing (no advanced constraint solvers like OR-Tools)
2. **Fixed time slots:** 10 periods per day (no custom slot definitions)
3. **No real-time updates:** WebSocket not implemented
4. **Basic AI tools:** Limited tool registry for agent
5. **Frontend incomplete:** Module UIs are scaffolds only

### Future Enhancements
1. **Advanced Scheduling:** Integrate Google OR-Tools or similar
2. **Custom Time Slots:** Allow institutions to define their own schedule periods
3. **Real-time Collaboration:** WebSocket for live schedule updates
4. **Event-Driven Architecture:** Full event sourcing with Watermill
5. **ML-Powered Recommendations:** Predict optimal schedule assignments
6. **Analytics Dashboard:** Visualize scheduling metrics and conflicts
7. **Mobile App:** Native iOS/Android apps
8. **Audit Logging:** Track all data changes with who/when/why
9. **Multi-Region Deployment:** Geo-replication for disaster recovery
10. **GraphQL API:** Alternative to REST

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Database scalability (many tenants) | Medium | High | Schema-per-tenant allows sharding per tenant DB |
| Scheduling algorithm inefficiency | Medium | Medium | Optimize annealing params, add timeout |
| LLM API downtime | Medium | Low | Fall back to rule-based suggestions |
| Security breach (data leak) | Low | Critical | Encryption at rest, RBAC, tenant isolation testing |
| Team attrition | Low | Medium | Good documentation, modular code structure |

---

## Budget & Timeline

### Development Resources
- **1 Backend Engineer:** Full-time (Done)
- **1 Frontend Engineer:** Full-time (Done)
- **1 DevOps/Infrastructure:** Part-time (Done)

### Timeline to Production
- **MVP:** Q1 2026 (Complete)
- **Beta Deployment:** Q2 2026
- **GA Release:** Q3 2026
- **Advanced Features:** Q4 2026+

### Cost Estimate
- **Development:** ~3-4 engineer-months completed
- **Infrastructure (AWS):** ~$2K/month (PostgreSQL + Redis + Kubernetes)
- **LLM API costs:** ~$500-1K/month (Claude API for beta users)

---

## Stakeholders & Communication

### Primary Stakeholders
- **Administrators:** Schedule coordinators, IT staff
- **Teachers:** End users viewing/approving schedules
- **Institutions:** C-level decision makers, IT directors

### Communication Plan
- Monthly progress reports to stakeholders
- Quarterly feature reviews with beta users
- Public documentation and API changelog
