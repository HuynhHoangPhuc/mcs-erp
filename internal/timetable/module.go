package timetable

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/application/services"
	coredelivery "github.com/HuynhHoangPhuc/mcs-erp/internal/core/delivery"
	coredomain "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	hrDomain      "github.com/HuynhHoangPhuc/mcs-erp/internal/hr/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"
	roomDomain    "github.com/HuynhHoangPhuc/mcs-erp/internal/room/domain"
	subjectDomain "github.com/HuynhHoangPhuc/mcs-erp/internal/subject/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/delivery"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/infrastructure"
)

// Module implements pkg/module.Module for the Timetable scheduling module.
type Module struct {
	pool           *pgxpool.Pool
	authSvc        *services.AuthService
	semesterRepo   domain.SemesterRepository
	scheduleRepo   domain.ScheduleRepository
	problemBuilder delivery.ProblemBuilder
}

// NewModule creates the Timetable module with a pre-built ProblemBuilder.
// Use this when you want full control over cross-module wiring in main.go.
func NewModule(
	pool *pgxpool.Pool,
	authSvc *services.AuthService,
	problemBuilder delivery.ProblemBuilder,
) *Module {
	return &Module{
		pool:           pool,
		authSvc:        authSvc,
		semesterRepo:   infrastructure.NewPostgresSemesterRepo(pool),
		scheduleRepo:   infrastructure.NewPostgresScheduleRepo(pool),
		problemBuilder: problemBuilder,
	}
}

// NewModuleWithRepos is a convenience constructor that wires cross-module
// adapters internally. Pass concrete repo instances from the hr/subject/room
// modules; no manual adapter creation needed in main.go.
func NewModuleWithRepos(
	pool         *pgxpool.Pool,
	authSvc      *services.AuthService,
	teacherRepo  hrDomain.TeacherRepository,
	availRepo    hrDomain.AvailabilityRepository,
	subjectRepo  subjectDomain.SubjectRepository,
	roomRepo     roomDomain.RoomRepository,
	roomAvail    roomDomain.RoomAvailabilityRepository,
) *Module {
	reader := infrastructure.NewCrossModuleReaderFromRepos(
		teacherRepo, availRepo, subjectRepo, roomRepo, roomAvail,
	)
	return NewModule(pool, authSvc, reader)
}

func (m *Module) Name() string           { return "timetable" }
func (m *Module) Dependencies() []string { return []string{"core", "hr", "subject", "room"} }
func (m *Module) Migrate(_ context.Context) error        { return nil }
func (m *Module) RegisterEvents(_ context.Context) error { return nil }

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	semHandler := delivery.NewSemesterHandler(m.semesterRepo)
	schedHandler := delivery.NewScheduleHandler(m.semesterRepo, m.scheduleRepo, m.problemBuilder)

	authMw := coredelivery.AuthMiddleware(m.authSvc)
	read  := auth.RequirePermission(coredomain.PermTimetableRead)
	write := auth.RequirePermission(coredomain.PermTimetableWrite)

	// Semester CRUD
	mux.Handle("POST /api/v1/timetable/semesters",
		authMw(write(http.HandlerFunc(semHandler.CreateSemester))))
	mux.Handle("GET /api/v1/timetable/semesters",
		authMw(read(http.HandlerFunc(semHandler.ListSemesters))))
	mux.Handle("GET /api/v1/timetable/semesters/{id}",
		authMw(read(http.HandlerFunc(semHandler.GetSemester))))

	// Semester subject management
	mux.Handle("POST /api/v1/timetable/semesters/{id}/subjects",
		authMw(write(http.HandlerFunc(semHandler.SetSubjects))))
	mux.Handle("POST /api/v1/timetable/semesters/{id}/subjects/{subjectId}/teacher",
		authMw(write(http.HandlerFunc(semHandler.AssignTeacher))))

	// Schedule generation, retrieval, approval
	mux.Handle("POST /api/v1/timetable/semesters/{id}/generate",
		authMw(write(http.HandlerFunc(schedHandler.GenerateSchedule))))
	mux.Handle("GET /api/v1/timetable/semesters/{id}/schedule",
		authMw(read(http.HandlerFunc(schedHandler.GetLatestSchedule))))
	mux.Handle("POST /api/v1/timetable/semesters/{id}/approve",
		authMw(write(http.HandlerFunc(schedHandler.ApproveSchedule))))

	// Manual assignment override
	mux.Handle("PUT /api/v1/timetable/assignments/{id}",
		authMw(write(http.HandlerFunc(schedHandler.UpdateAssignment))))
}
