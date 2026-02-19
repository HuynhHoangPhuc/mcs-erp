package services

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/infrastructure"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
)

// AuthService handles authentication (login, refresh, validate).
type AuthService struct {
	userRepo   domain.UserRepository
	roleRepo   domain.RoleRepository
	lookupRepo domain.UsersLookupRepository
	jwt        *infrastructure.JWTService
}

// NewAuthService creates a new auth service.
func NewAuthService(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	lookupRepo domain.UsersLookupRepository,
	jwt *infrastructure.JWTService,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		lookupRepo: lookupRepo,
		jwt:        jwt,
	}
}

// Login authenticates a user by email+password. Resolves tenant from users_lookup.
func (s *AuthService) Login(ctx context.Context, email, password string) (*infrastructure.TokenPair, error) {
	// Resolve tenant schema from public.users_lookup
	schema, err := s.lookupRepo.FindTenantByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Set tenant context for repo queries
	ctx = tenant.WithTenant(ctx, schema)

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Gather permissions from all roles
	perms, err := s.getUserPermissions(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("get permissions: %w", err)
	}

	return s.jwt.GenerateTokenPair(user.ID, schema, user.Email, perms)
}

// Refresh validates a refresh token and returns a new token pair.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*infrastructure.TokenPair, error) {
	claims, err := s.jwt.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Set tenant context
	ctx = tenant.WithTenant(ctx, claims.TenantID)

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	perms, err := s.getUserPermissions(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("get permissions: %w", err)
	}

	return s.jwt.GenerateTokenPair(user.ID, claims.TenantID, user.Email, perms)
}

// ValidateToken validates an access token and returns claims.
func (s *AuthService) ValidateToken(token string) (*auth.Claims, error) {
	return s.jwt.ValidateToken(token)
}

func (s *AuthService) getUserPermissions(ctx context.Context, user *domain.User) ([]string, error) {
	roles, err := s.roleRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	permSet := make(map[string]struct{})
	for _, role := range roles {
		for _, p := range role.Permissions {
			permSet[p] = struct{}{}
		}
	}

	perms := make([]string, 0, len(permSet))
	for p := range permSet {
		perms = append(perms, p)
	}
	return perms, nil
}
