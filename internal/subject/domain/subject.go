package domain

import (
	"time"

	"github.com/google/uuid"
)

// Subject represents a course/subject within a tenant's curriculum.
type Subject struct {
	ID           uuid.UUID
	Name         string
	Code         string // unique per tenant
	Description  string
	CategoryID   *uuid.UUID
	Credits      int
	HoursPerWeek int
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
