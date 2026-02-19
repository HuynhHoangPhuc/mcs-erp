package domain

import "github.com/google/uuid"

// PrerequisiteEdge represents a directed edge in the prerequisite graph:
// SubjectID requires PrerequisiteID to be completed first.
type PrerequisiteEdge struct {
	SubjectID      uuid.UUID
	PrerequisiteID uuid.UUID
	// Version is used for optimistic locking on the prerequisite set for SubjectID.
	Version int
}
