# Development Roadmap

## Current Status: MVP Complete ✅

All 8 implementation phases completed. System is feature-complete for course scheduling use case.

| Phase | Status | Completion Date | Notes |
|-------|--------|-----------------|-------|
| Phase 1: Core Infrastructure | ✅ Complete | Feb 2026 | Auth, RBAC, module registry, DDD structure |
| Phase 2: HR Module | ✅ Complete | Feb 2026 | Teachers, departments, availability |
| Phase 3: Subject Module | ✅ Complete | Feb 2026 | Subjects, categories, prerequisites (DAG) |
| Phase 4: Room Module | ✅ Complete | Feb 2026 | Rooms, capacity, availability |
| Phase 5: Timetable Engine | ✅ Complete | Feb 2026 | Scheduling algorithm, approval workflow |
| Phase 6: AI Agent | ✅ Complete | Feb 2026 | Multi-LLM, tool registry, conversations |
| Phase 7: Frontend Scaffold | ✅ Complete | Feb 2026 | React 19, TanStack, module placeholders |
| Phase 8: Testing & Docs | ✅ Complete | Feb 2026 | Unit tests, API docs, architecture docs |

---

## Immediate Next Steps (Week 1-2)

### 1. Backend Testing & Hardening
**Priority:** High | **Effort:** 2 days | **Owner:** Backend Engineer

**Tasks:**
- [ ] Add comprehensive unit tests for domain logic (target 80% coverage)
- [ ] Add integration tests for critical API flows (login → create schedule → approve)
- [ ] Add database migration tests (schema creation across tenants)
- [ ] Test error scenarios and edge cases
- [ ] Load test with 100+ concurrent users
- [ ] Security testing: JWT expiry, permission enforcement, tenant isolation

**Success Criteria:**
- All tests pass locally and in CI/CD
- Coverage report shows 80%+ coverage
- No SQL injection vulnerabilities
- No unauthorized cross-tenant access

**Files to Test:**
- core/application/services/auth_service.go
- timetable/infrastructure/scheduler.go (algorithm correctness)
- platform/module/registry.go (dependency resolution)
- All repository implementations

---

### 2. Frontend Module Implementation
**Priority:** High | **Effort:** 5 days | **Owner:** Frontend Engineer

**Tasks:**
- [ ] Implement Authentication UI (login page, token storage, logout)
- [ ] Implement HR Module UI (teacher list, departments, availability matrix)
- [ ] Implement Subject Module UI (subject list, prerequisites, category management)
- [ ] Implement Room Module UI (room list, availability, equipment management)
- [ ] Implement Timetable UI (semester CRUD, schedule generation, approval workflow)
- [ ] Implement Agent Module UI (chat interface, conversation history)

**Success Criteria:**
- All module UIs functional and responsive
- Forms validate input and display errors
- Tables support pagination and filtering
- Real-time updates work (SSE for agent chat)

**File Structure:**
```
web/apps/shell/src/routes/
├── auth/
│   ├── login.tsx
│   └── logout.tsx
├── hr/
│   ├── teachers/
│   ├── departments/
│   └── availability/
├── subjects/
├── rooms/
├── timetable/
│   ├── semesters/
│   ├── schedule-generation/
│   └── approval/
└── agent/
    ├── chat.tsx
    └── conversations/
```

---

### 3. API Documentation & OpenAPI
**Priority:** Medium | **Effort:** 1 day | **Owner:** Backend Engineer

**Tasks:**
- [ ] Generate OpenAPI/Swagger specs for all endpoints
- [ ] Add request/response examples to spec
- [ ] Deploy Swagger UI at /swagger
- [ ] Document authentication flow (JWT header)
- [ ] Document permission requirements per endpoint
- [ ] Add error response examples

**Success Criteria:**
- All 60+ endpoints documented
- Spec is valid OpenAPI 3.0
- Swagger UI accessible and interactive

---

### 4. Development Environment Setup
**Priority:** Medium | **Effort:** 1 day | **Owner:** DevOps/Backend

**Tasks:**
- [ ] Document .env setup (.env.example with all vars)
- [ ] Document Docker Compose setup (PostgreSQL, Redis)
- [ ] Create quick-start guide for new developers
- [ ] Add pre-commit hooks for linting/formatting
- [ ] Document database migration process
- [ ] Create Makefile targets for common tasks

**Success Criteria:**
- New developer can set up environment in < 30 minutes
- `make dev` starts backend with hot-reload
- `make test` runs all tests
- `make lint` checks code style

---

## Short-Term Roadmap (Week 3-6)

### Phase 9: Beta Deployment & User Testing

**Timeline:** 2 weeks

**Activities:**
1. Deploy MVP to staging environment (Kubernetes cluster)
2. Create beta program with 3-5 institutions
3. Collect feedback on UI/UX and scheduling accuracy
4. Fix critical bugs from beta testing
5. Performance tuning based on real usage

**Deliverables:**
- Staging deployment ready
- Beta user onboarding documentation
- Feature feedback survey
- Bug tracker populated with issues

---

### Phase 10: Performance Optimization

**Timeline:** 1 week

**Focus Areas:**
1. **Database Query Optimization**
   - Add missing indexes
   - Optimize N+1 queries
   - Profile slow endpoints

2. **Algorithm Tuning**
   - Optimize simulated annealing parameters
   - Add early termination conditions
   - Cache course conflict information

3. **Frontend Performance**
   - Code splitting for module bundles
   - Lazy loading of routes
   - Image optimization
   - Reduce bundle size

**Success Metrics:**
- API p95 response time < 200ms
- Schedule generation < 30s for 500 courses
- Frontend bundle size < 500KB

---

### Phase 11: Security Hardening

**Timeline:** 1 week

**Focus Areas:**
1. **Penetration Testing**
   - SQL injection attempts
   - JWT tampering
   - IDOR vulnerabilities
   - Cross-site scripting (XSS)

2. **Security Fixes**
   - Rate limiting on login endpoint
   - CSRF token on state-changing endpoints
   - Input sanitization
   - HTTPS enforcement

3. **Audit Logging**
   - Log all permission checks
   - Log data mutations with user info
   - Store audit logs in separate table

**Success Metrics:**
- Zero critical security vulnerabilities
- Audit log captures all sensitive operations

---

## Medium-Term Roadmap (Weeks 7-12)

### Phase 12: GA Release

**Timeline:** 2 weeks

**Preparation:**
- [ ] Production deployment (AWS, GCP, or self-hosted)
- [ ] Database backup & recovery procedures
- [ ] Monitoring setup (Prometheus, Grafana, or CloudWatch)
- [ ] Alerting for critical errors
- [ ] Runbook for common incidents
- [ ] SLA documentation (uptime, support response time)

**Launch Activities:**
- [ ] Public announcement
- [ ] Documentation website
- [ ] Support email/chat channel
- [ ] Getting started guide
- [ ] Migration guide from manual scheduling

---

### Phase 13: Advanced Scheduling Features

**Timeline:** 3 weeks

**Features:**
1. **Custom Time Slots**
   - Allow institutions to define their own schedule periods
   - UI to configure morning/afternoon/evening slots
   - Algorithm adapts to custom slots

2. **Advanced Constraints**
   - Minimize gaps in teacher schedules
   - Balance load across rooms
   - Prefer contiguous time slots (no spread-out assignments)

3. **Conflict Resolution**
   - Identify unavoidable conflicts
   - Suggest manual adjustments
   - Show impact of changes

**Deliverables:**
- Custom slot configuration API
- Conflict detection and suggestions
- Updated scheduling algorithm

---

### Phase 14: Analytics & Insights

**Timeline:** 2 weeks

**Features:**
1. **Dashboard Metrics**
   - Scheduling efficiency (% conflicts resolved)
   - Room utilization (% of time slots used)
   - Teacher load distribution
   - Most problematic constraints

2. **Reports**
   - Schedule summary (courses, teachers, rooms)
   - Conflict report
   - Room usage report
   - Export schedule as PDF/Excel

3. **Visualizations**
   - Calendar view of generated schedule
   - Heatmap of room utilization
   - Teacher load chart

**Deliverables:**
- Analytics API endpoints
- React dashboard components
- PDF export functionality

---

## Long-Term Roadmap (3-6 months)

### Phase 15: AI Enhancements

**Timeline:** 3 weeks

**Features:**
1. **Advanced LLM Capabilities**
   - Agent can suggest schedule improvements
   - Natural language schedule queries ("Which rooms are free on Monday?")
   - Conflict prediction and resolution suggestions

2. **ML-Based Optimization**
   - Learn from historical schedules
   - Predict teacher preferences
   - Optimize for institution-specific goals

3. **Tool Ecosystem**
   - Easy API for modules to register tools
   - Tool versioning and governance
   - Tool marketplace concept

**Deliverables:**
- Enhanced agent service with new capabilities
- Tool registration documentation
- Sample tools for each module

---

### Phase 16: Multi-Region Deployment

**Timeline:** 2 weeks

**Features:**
1. **Geo-Replication**
   - PostgreSQL replication to multiple regions
   - Redis cluster for distributed caching
   - CDN for frontend assets

2. **Disaster Recovery**
   - Automated failover on region outage
   - Point-in-time recovery capability
   - RTO/RPO targets (< 1 hour)

3. **Data Residency**
   - EU data centers for GDPR compliance
   - Data localization options

**Deliverables:**
- Multi-region infrastructure
- DR playbook
- Compliance documentation

---

### Phase 17: Mobile App

**Timeline:** 4 weeks

**Features:**
1. **Native Apps (iOS/Android)**
   - View schedule on the go
   - Receive notifications for conflicts
   - Quick actions (approve schedule, view teacher availability)

2. **Features**
   - Offline mode (cached schedule)
   - Push notifications
   - QR code scanning for course details

**Deliverables:**
- React Native mobile app (iOS + Android)
- Mobile-specific APIs
- App store listings

---

### Phase 18: Enterprise Features

**Timeline:** 2 weeks

**Features:**
1. **Advanced RBAC**
   - Custom role creation
   - Delegated administration
   - Audit trail of permission changes

2. **Integration APIs**
   - Webhook support for external systems
   - Zapier integration
   - SAML/OIDC SSO

3. **Multi-Tenant Management**
   - Tenant provisioning API
   - Usage billing/metering
   - Per-tenant customization

**Deliverables:**
- Enterprise APIs
- Integration documentation
- SSO configuration guide

---

## Known Issues & Technical Debt

### Current Issues
1. **Frontend placeholder components** — Need full implementation
2. **Missing error handling edge cases** — Some error scenarios untested
3. **No real-time updates** — WebSocket not implemented (using polling)
4. **Limited scheduling algorithm** — Only greedy + annealing (no OR-Tools)

### Technical Debt
1. **Database migrations** — Manual migration process (need golang-migrate)
2. **gRPC unused** — Protobuf definitions created but not used yet
3. **Event bus unused** — Watermill set up but no event handlers
4. **Frontend state management** — Could use more sophisticated patterns

### Priority Fixes
1. Complete frontend module UIs
2. Add comprehensive test coverage
3. Implement proper database migration tooling
4. Add WebSocket support for real-time updates

---

## Metrics & Success Criteria

### Development Metrics
| Metric | Target | Current |
|--------|--------|---------|
| Code Coverage | 80%+ | TBD (after testing phase) |
| API Response Time (p95) | < 200ms | TBD (after optimization) |
| Uptime | 99.5% | N/A (not deployed yet) |
| Security Scan Pass Rate | 100% | TBD (after hardening) |

### Business Metrics
| Metric | Target by GA | Notes |
|--------|--------------|-------|
| Beta Users | 5+ institutions | Collecting feedback |
| User Adoption | 100+ admins | By end of Q2 2026 |
| Courses Scheduled | 1000+/month | By end of Q3 2026 |
| Customer Satisfaction | NPS > 50 | Survey at GA |
| Uptime SLA | 99.5% | After production deployment |

---

## Dependencies & Blockers

### External Dependencies
- **PostgreSQL:** v16+ (database)
- **Redis:** v7+ (optional, for caching)
- **Claude/OpenAI API:** For LLM features
- **Kubernetes:** For production deployment

### Internal Dependencies
1. **Frontend completion** blocks user acceptance testing
2. **Performance optimization** blocks GA release
3. **Security hardening** blocks customer deployment

### Risks
1. **AI API costs** — May escalate with user base (mitigate: usage quotas, batching)
2. **Database scalability** — Schema-per-tenant may hit limits at 10K+ tenants (mitigate: shard by tenant)
3. **Scheduling algorithm** — May not handle complex constraints (mitigate: add OR-Tools support)

---

## Resource Planning

### Recommended Team
- **1 Backend Engineer** (full-time)
  - Testing, performance optimization, security hardening
  - Advanced scheduling features, analytics API

- **1 Frontend Engineer** (full-time)
  - Complete module UIs
  - Dashboard and analytics UI
  - Mobile app (later phases)

- **1 DevOps/Infra Engineer** (part-time, 50%)
  - Deployment automation
  - Monitoring setup
  - Performance profiling

### Training Needs
- Code review process (new engineers)
- Architecture overview (new team members)
- Deployment procedures (all team)
- Customer support (if applicable)

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 0.1.0 (MVP) | Feb 2026 | Initial release: 6 modules, auth, scheduling, AI agent |
| 0.2.0 (Beta) | Q2 2026 | Frontend completion, performance optimization, security hardening |
| 1.0.0 (GA) | Q3 2026 | Production-ready, SLA support, advanced features |
| 1.1.0 | Q4 2026 | Analytics, advanced scheduling, mobile app |
| 2.0.0 | 2027 | Enterprise features, multi-region, integrations |

---

## Communication & Handoff

### Weekly Standup
- [ ] Deploy progress against roadmap
- [ ] Blockers and risks
- [ ] Next week's priorities

### Stakeholder Updates
- [ ] Monthly progress reports (high-level status)
- [ ] Quarterly feature reviews (demo new features)
- [ ] Roadmap updates (scope changes)

### Documentation
- [ ] API changelog (breaking changes, new endpoints)
- [ ] Architecture decision records (ADRs)
- [ ] Release notes (features, fixes, known issues)

---

## Decision Points

### Scheduling Algorithm
**Decision:** Start with greedy + simulated annealing
**Alternative:** Integrate Google OR-Tools (more powerful, slower)
**Trigger for Upgrade:** If beta users report unsolvable conflicts

### Frontend Framework
**Decision:** React 19 + TanStack
**Alternative:** Vue.js or Svelte
**Rationale:** React ecosystem largest, TanStack mature

### Database Strategy
**Decision:** Schema-per-tenant (current)
**Alternative:** Row-level security with shared schema
**Rationale:** Stronger isolation, easier multi-tenancy

### Deployment Platform
**Decision:** Kubernetes (AWS EKS recommended)
**Alternative:** Serverless (AWS Lambda), Platform-as-a-Service (Heroku)
**Rationale:** Better control, cost-effective at scale

---

## Appendix: Detailed Phase Breakdown

### Current Phase: Testing & Documentation (Week 1-2)
**Goal:** Ensure system is robust and well-documented

**Acceptance Criteria:**
- 80%+ test coverage
- All APIs documented with examples
- Security vulnerabilities identified and documented
- Performance benchmarks established

### Next Phase: Beta Deployment (Week 3-4)
**Goal:** Collect real-world feedback from 3-5 institutions

**Acceptance Criteria:**
- Production-like staging environment
- Beta users onboarded and trained
- Feedback survey deployed
- Critical bugs fixed within 24 hours

### Future Phase: GA Release (Week 5-6)
**Goal:** Production-ready system with SLA support

**Acceptance Criteria:**
- Zero critical security vulnerabilities
- 99.5% uptime in staging
- Customer support process defined
- Documentation complete
