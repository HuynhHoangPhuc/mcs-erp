package domain

import (
	"time"

	"github.com/google/uuid"
)

// SubjectCreated is published when a new subject is successfully persisted.
type SubjectCreated struct {
	SubjectID  uuid.UUID
	Name       string
	Code       string
	OccurredAt time.Time
}

// PrerequisiteAdded is published when a prerequisite edge is added to the graph.
type PrerequisiteAdded struct {
	SubjectID      uuid.UUID
	PrerequisiteID uuid.UUID
	OccurredAt     time.Time
}

// PrerequisiteRemoved is published when a prerequisite edge is removed from the graph.
type PrerequisiteRemoved struct {
	SubjectID      uuid.UUID
	PrerequisiteID uuid.UUID
	OccurredAt     time.Time
}
