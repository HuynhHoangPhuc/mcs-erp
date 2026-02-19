package domain

import (
	"time"

	"github.com/google/uuid"
)

// SemesterStatus defines valid lifecycle states for a semester scheduling run.
type SemesterStatus string

const (
	SemesterStatusDraft      SemesterStatus = "draft"
	SemesterStatusScheduling SemesterStatus = "scheduling"
	SemesterStatusReview     SemesterStatus = "review"
	SemesterStatusApproved   SemesterStatus = "approved"
	SemesterStatusRejected   SemesterStatus = "rejected"
)

// Semester represents an academic semester that holds a timetable.
type Semester struct {
	ID        uuid.UUID
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Status    SemesterStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SemesterSubject links a subject (and optional teacher) to a semester.
type SemesterSubject struct {
	SemesterID uuid.UUID
	SubjectID  uuid.UUID
	TeacherID  *uuid.UUID // nil until a teacher is assigned
}
