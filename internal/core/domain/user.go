package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a system user within a tenant.
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Name         string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
