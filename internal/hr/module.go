package hr

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/application/services"
	coredelivery "github.com/HuynhHoangPhuc/mcs-erp/internal/core/delivery"
	coredomain "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/hr/delivery"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/hr/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/hr/infrastructure"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"
)

// Module implements pkg/module.Module for the HR (teachers/departments/availability) module.
type Module struct {
	pool          *pgxpool.Pool
	authSvc       *services.AuthService
	teacherRepo   domain.TeacherRepository
	deptRepo      domain.DepartmentRepository
	availRepo     domain.AvailabilityRepository
}

// NewModule creates the HR module wired with concrete dependencies.
func NewModule(pool *pgxpool.Pool, authSvc *services.AuthService) *Module {
	return &Module{
		pool:        pool,
		authSvc:     authSvc,
		teacherRepo: infrastructure.NewPostgresTeacherRepo(pool),
		deptRepo:    infrastructure.NewPostgresDepartmentRepo(pool),
		availRepo:   infrastructure.NewPostgresAvailabilityRepo(pool),
	}
}

// TeacherRepo returns the teacher repository for cross-module access.
func (m *Module) TeacherRepo() domain.TeacherRepository { return m.teacherRepo }

// AvailabilityRepo returns the availability repository for cross-module access.
func (m *Module) AvailabilityRepo() domain.AvailabilityRepository { return m.availRepo }

func (m *Module) Name() string           { return "hr" }
func (m *Module) Dependencies() []string { return []string{"core"} }
func (m *Module) Migrate(ctx context.Context) error        { return nil }
func (m *Module) RegisterEvents(ctx context.Context) error { return nil }

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	teacherHandler := delivery.NewTeacherHandler(m.teacherRepo)
	deptHandler := delivery.NewDepartmentHandler(m.deptRepo)
	availHandler := delivery.NewAvailabilityHandler(m.availRepo, m.teacherRepo)

	authMw := coredelivery.AuthMiddleware(m.authSvc)

	teacherRead  := auth.RequirePermission(coredomain.PermTeacherRead)
	teacherWrite := auth.RequirePermission(coredomain.PermTeacherWrite)
	deptRead     := auth.RequirePermission(coredomain.PermDeptRead)
	deptWrite    := auth.RequirePermission(coredomain.PermDeptWrite)

	// Teacher routes
	mux.Handle("POST /api/v1/teachers",
		authMw(teacherWrite(http.HandlerFunc(teacherHandler.CreateTeacher))))
	mux.Handle("GET /api/v1/teachers",
		authMw(teacherRead(http.HandlerFunc(teacherHandler.ListTeachers))))
	mux.Handle("GET /api/v1/teachers/{id}",
		authMw(teacherRead(http.HandlerFunc(teacherHandler.GetTeacher))))
	mux.Handle("PUT /api/v1/teachers/{id}",
		authMw(teacherWrite(http.HandlerFunc(teacherHandler.UpdateTeacher))))

	// Availability routes (nested under teacher)
	mux.Handle("GET /api/v1/teachers/{id}/availability",
		authMw(teacherRead(http.HandlerFunc(availHandler.GetAvailability))))
	mux.Handle("PUT /api/v1/teachers/{id}/availability",
		authMw(teacherWrite(http.HandlerFunc(availHandler.SetAvailability))))

	// Department routes
	mux.Handle("POST /api/v1/departments",
		authMw(deptWrite(http.HandlerFunc(deptHandler.CreateDepartment))))
	mux.Handle("GET /api/v1/departments",
		authMw(deptRead(http.HandlerFunc(deptHandler.ListDepartments))))
	mux.Handle("GET /api/v1/departments/{id}",
		authMw(deptRead(http.HandlerFunc(deptHandler.GetDepartment))))
	mux.Handle("PUT /api/v1/departments/{id}",
		authMw(deptWrite(http.HandlerFunc(deptHandler.UpdateDepartment))))
	mux.Handle("DELETE /api/v1/departments/{id}",
		authMw(deptWrite(http.HandlerFunc(deptHandler.DeleteDepartment))))
}
