package domain

import (
	"time"

	"github.com/google/uuid"
)

// Teacher represents a teaching staff member within a tenant.
type Teacher struct {
	ID             uuid.UUID
	Name           string
	Email          string
	DepartmentID   *uuid.UUID
	Qualifications []string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
