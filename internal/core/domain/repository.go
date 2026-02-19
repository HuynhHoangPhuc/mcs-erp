package domain

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository defines persistence operations for User entities.
type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Save(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	List(ctx context.Context, offset, limit int) ([]*User, int, error)
}

// RoleRepository defines persistence operations for Role entities.
type RoleRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Role, error)
	FindByName(ctx context.Context, name string) (*Role, error)
	Save(ctx context.Context, role *Role) error
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*Role, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
}

// UsersLookupRepository handles the public.users_lookup table for cross-tenant login.
type UsersLookupRepository interface {
	FindTenantByEmail(ctx context.Context, email string) (tenantSchema string, err error)
	Upsert(ctx context.Context, email, tenantSchema string) error
}
