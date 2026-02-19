package domain

import (
	"time"

	"github.com/google/uuid"
)

// ScheduleGenerated is published when the scheduler produces a new schedule version.
type ScheduleGenerated struct {
	SemesterID     uuid.UUID
	Version        int
	HardViolations int
	SoftPenalty    float64
	GeneratedAt    time.Time
}

// ScheduleApproved is published when an admin approves a generated schedule.
type ScheduleApproved struct {
	SemesterID  uuid.UUID
	Version     int
	ApprovedAt  time.Time
}

// AssignmentModified is published when a single assignment is manually adjusted.
type AssignmentModified struct {
	AssignmentID uuid.UUID
	SemesterID   uuid.UUID
	ModifiedAt   time.Time
}
