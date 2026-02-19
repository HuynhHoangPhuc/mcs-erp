package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role defines a named set of permissions assignable to users.
type Role struct {
	ID          uuid.UUID
	Name        string
	Permissions []string
	Description string
	CreatedAt   time.Time
}
