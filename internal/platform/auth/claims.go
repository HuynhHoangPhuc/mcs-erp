package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents JWT token claims for authenticated users.
type Claims struct {
	jwt.RegisteredClaims
	UserID      uuid.UUID `json:"user_id"`
	TenantID    string    `json:"tenant_id"`
	Email       string    `json:"email"`
	Permissions []string  `json:"permissions,omitempty"`
}
