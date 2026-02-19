package subject

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/application/services"
	coredomain "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/delivery"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"
	subdelivery "github.com/HuynhHoangPhuc/mcs-erp/internal/subject/delivery"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/subject/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/subject/infrastructure"
)

// Module implements pkg/module.Module for the subject module.
type Module struct {
	pool         *pgxpool.Pool
	authSvc      *services.AuthService
	subjectRepo  domain.SubjectRepository
	categoryRepo domain.CategoryRepository
	prereqRepo   domain.PrerequisiteRepository
}

// NewModule creates the subject module wired with concrete dependencies.
func NewModule(pool *pgxpool.Pool, authSvc *services.AuthService) *Module {
	return &Module{
		pool:         pool,
		authSvc:      authSvc,
		subjectRepo:  infrastructure.NewPostgresSubjectRepo(pool),
		categoryRepo: infrastructure.NewPostgresCategoryRepo(pool),
		prereqRepo:   infrastructure.NewPostgresPrerequisiteRepo(pool),
	}
}

// SubjectRepo returns the subject repository for cross-module access.
func (m *Module) SubjectRepo() domain.SubjectRepository { return m.subjectRepo }

func (m *Module) Name() string          { return "subject" }
func (m *Module) Dependencies() []string { return []string{"core"} }
func (m *Module) Migrate(ctx context.Context) error { return nil }
func (m *Module) RegisterEvents(ctx context.Context) error { return nil }

// RegisterRoutes wires all subject, category, and prerequisite endpoints.
func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	subjectHandler := subdelivery.NewSubjectHandler(m.subjectRepo)
	categoryHandler := subdelivery.NewCategoryHandler(m.categoryRepo)
	prereqHandler := subdelivery.NewPrerequisiteHandler(m.prereqRepo, m.subjectRepo)

	authMw := delivery.AuthMiddleware(m.authSvc)
	readPerm := auth.RequirePermission(coredomain.PermSubjectRead)
	writePerm := auth.RequirePermission(coredomain.PermSubjectWrite)

	// Subject routes
	mux.Handle("POST /api/v1/subjects", authMw(writePerm(http.HandlerFunc(subjectHandler.CreateSubject))))
	mux.Handle("GET /api/v1/subjects", authMw(readPerm(http.HandlerFunc(subjectHandler.ListSubjects))))
	mux.Handle("GET /api/v1/subjects/{id}", authMw(readPerm(http.HandlerFunc(subjectHandler.GetSubject))))
	mux.Handle("PUT /api/v1/subjects/{id}", authMw(writePerm(http.HandlerFunc(subjectHandler.UpdateSubject))))

	// Category routes
	mux.Handle("POST /api/v1/categories", authMw(writePerm(http.HandlerFunc(categoryHandler.CreateCategory))))
	mux.Handle("GET /api/v1/categories", authMw(readPerm(http.HandlerFunc(categoryHandler.ListCategories))))
	mux.Handle("GET /api/v1/categories/{id}", authMw(readPerm(http.HandlerFunc(categoryHandler.GetCategory))))

	// Prerequisite routes
	mux.Handle("POST /api/v1/subjects/{id}/prerequisites",
		authMw(writePerm(http.HandlerFunc(prereqHandler.AddPrerequisite))))
	mux.Handle("DELETE /api/v1/subjects/{id}/prerequisites/{prereqId}",
		authMw(writePerm(http.HandlerFunc(prereqHandler.RemovePrerequisite))))
	mux.Handle("GET /api/v1/subjects/{id}/prerequisites",
		authMw(readPerm(http.HandlerFunc(prereqHandler.ListPrerequisites))))
	mux.Handle("GET /api/v1/subjects/{id}/prerequisite-chain",
		authMw(readPerm(http.HandlerFunc(prereqHandler.GetPrerequisiteChain))))
}
