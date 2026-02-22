# Project Roadmap

## Phase 7: Integration Testing · Feb 2026 — Completed ✅
- All six phases of the 2026 integration-testing plan (infrastructure, smoke validation, auth/tenant, module suites, cross-module timetable flows, and security/performance regression) are now completed as of 2026-02-22.
- Coverage includes timetable cross-module reader validation, semester CRUD/permissions, scheduling generation/approval/update flow with double-booking guards, IDOR scans, tenant claim/header mismatch prevention, SQL/schema injection handling, JWT `alg=none` rejection, and RBAC bypass checks.
- Validation runs: `make test-integration` (full suite, `-tags integration -race`), `go test ./internal/security_test/...`, `go test ./internal/timetable/... -race`, and `go test ./internal/platform/... -run Concurrent -race` all passed under race safety flags.

## Phase 8: Frontend Implementation · Feb 2026 — Completed ✅
- Complete React 19 frontend implementation with all 6 backend modules mapped to UI pages.
- **Tech Stack:** React 19, TypeScript, TanStack (Router/Query/Table/Form), shadcn/ui, Tailwind v4, Turborepo+pnpm
- **Structure:** `web/` monorepo with `apps/shell` (Vite SPA) and `packages/` (ui, api-client, module-{hr,subject,room,timetable,agent})
- **Features Implemented:**
  - **HR Module:** Teachers table, departments CRUD, availability grid editor (7×10 time slots)
  - **Subject Module:** Subject CRUD, category manager, prerequisite DAG visualizer with cycle detection
  - **Room Module:** Room CRUD, capacity/equipment metadata, availability grid editor
  - **Timetable Module:** Semester wizard, subject-teacher-room assignment form, schedule generation with SSE progress streaming, conflict detection grid, admin approval view
  - **Agent Module:** AI chat interface with real-time SSE message streaming, conversation history, tool-call indicators
- **Auth:** JWT-based login with auto-refresh, protected routes via TanStack Router, token refresh middleware
- **Layout:** Collapsible sidebar navigation, breadcrumb context header, responsive design (mobile/tablet/desktop)
- **State Management:** TanStack Query for server state (caching, invalidation), React Context for auth/theme, TanStack Form for validation
- **Build & Dev:** Turborepo for parallelized dev/build/lint, pnpm workspaces, Vite for fast HMR

## Next Milestones
1. **Phase 9: Beta Deployment & User Testing** – Prepare staging rollout, onboard pilot institutions, and collect feedback on scheduling accuracy and UI/UX.
2. **Phase 10: Performance Optimization** – Profile expensive queries, optimize scheduling algorithm, monitor API latency p95/p99 targets, implement pagination/lazy-loading.
3. **Phase 11: Security Hardening** – Penetration testing, rate limits, audit logging, CSRF protection, HTTPS enforcement, helmet headers.

_Source: Integration-testing plan (Feb 2026) and frontend implementation (Feb 2026) artifacts._
