package domain

import (
	"time"

	"github.com/google/uuid"
)

// TeacherCreated is published when a new teacher is successfully persisted.
type TeacherCreated struct {
	TeacherID uuid.UUID
	Name      string
	Email     string
	OccurredAt time.Time
}

// TeacherUpdated is published when a teacher's core fields are changed.
type TeacherUpdated struct {
	TeacherID    uuid.UUID
	Name         string
	Email        string
	DepartmentID *uuid.UUID
	IsActive     bool
	OccurredAt   time.Time
}

// AvailabilityUpdated is published when a teacher's weekly availability is replaced.
type AvailabilityUpdated struct {
	TeacherID  uuid.UUID
	SlotCount  int // total slots set as available
	OccurredAt time.Time
}
