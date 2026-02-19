package domain

import (
	"time"

	"github.com/google/uuid"
)

// Schedule is the complete timetable result for a semester at a given version.
type Schedule struct {
	SemesterID     uuid.UUID
	Version        int
	Assignments    []Assignment
	HardViolations int
	SoftPenalty    float64
	GeneratedAt    time.Time
}
