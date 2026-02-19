package core

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/application/services"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/delivery"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/infrastructure"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"
)

// Module implements pkg/module.Module for the core (auth/user/role) module.
type Module struct {
	pool       *pgxpool.Pool
	jwtService *infrastructure.JWTService
	authSvc    *services.AuthService
	userRepo   domain.UserRepository
	roleRepo   domain.RoleRepository
	lookupRepo domain.UsersLookupRepository
}

// NewModuleWithDeps creates the core module wired with concrete dependencies.
func NewModuleWithDeps(pool *pgxpool.Pool, jwtSvc *infrastructure.JWTService) *Module {
	userRepo := infrastructure.NewPostgresUserRepo(pool)
	roleRepo := infrastructure.NewPostgresRoleRepo(pool)
	lookupRepo := infrastructure.NewPostgresUsersLookupRepo(pool)
	authSvc := services.NewAuthService(userRepo, roleRepo, lookupRepo, jwtSvc)

	return &Module{
		pool:       pool,
		jwtService: jwtSvc,
		authSvc:    authSvc,
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		lookupRepo: lookupRepo,
	}
}

func (m *Module) Name() string            { return "core" }
func (m *Module) Dependencies() []string   { return nil }
func (m *Module) Migrate(ctx context.Context) error { return nil }
func (m *Module) RegisterEvents(ctx context.Context) error { return nil }

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	authHandler := delivery.NewAuthHandler(m.authSvc)
	userHandler := delivery.NewUserHandler(m.userRepo, m.roleRepo, m.lookupRepo)
	roleHandler := delivery.NewRoleHandler(m.roleRepo)

	// Public auth routes (no JWT required)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)

	// Protected routes — wrapped with auth middleware + permission checks
	authMw := delivery.AuthMiddleware(m.authSvc)
	userPerm := auth.RequirePermission(domain.PermUserWrite)
	rolePerm := auth.RequirePermission(domain.PermRoleWrite)
	readPerm := auth.RequirePermission(domain.PermUserRead)

	// Users
	mux.Handle("POST /api/v1/users", authMw(userPerm(http.HandlerFunc(userHandler.CreateUser))))
	mux.Handle("GET /api/v1/users", authMw(readPerm(http.HandlerFunc(userHandler.ListUsers))))
	mux.Handle("GET /api/v1/users/{id}", authMw(readPerm(http.HandlerFunc(userHandler.GetUser))))
	mux.Handle("POST /api/v1/users/{id}/roles", authMw(userPerm(http.HandlerFunc(userHandler.AssignRole))))

	// Roles
	mux.Handle("POST /api/v1/roles", authMw(rolePerm(http.HandlerFunc(roleHandler.CreateRole))))
	mux.Handle("GET /api/v1/roles", authMw(auth.RequirePermission(domain.PermRoleRead)(http.HandlerFunc(roleHandler.ListRoles))))
	mux.Handle("GET /api/v1/roles/{id}", authMw(auth.RequirePermission(domain.PermRoleRead)(http.HandlerFunc(roleHandler.GetRole))))
	mux.Handle("DELETE /api/v1/roles/{id}", authMw(rolePerm(http.HandlerFunc(roleHandler.DeleteRole))))
}

// AuthService returns the auth service for use by other modules or main.
func (m *Module) AuthService() *services.AuthService { return m.authSvc }
