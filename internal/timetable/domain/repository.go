package domain

import (
	"context"

	"github.com/google/uuid"
)

// SemesterRepository defines persistence operations for Semester aggregates.
type SemesterRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Semester, error)
	Save(ctx context.Context, s *Semester) error
	Update(ctx context.Context, s *Semester) error
	List(ctx context.Context, offset, limit int) ([]*Semester, int, error)

	// AddSubjects links a set of subjects to the semester (idempotent).
	AddSubjects(ctx context.Context, semesterID uuid.UUID, subjectIDs []uuid.UUID) error
	// GetSubjects returns all SemesterSubject rows for the given semester.
	GetSubjects(ctx context.Context, semesterID uuid.UUID) ([]*SemesterSubject, error)
	// SetTeacherAssignment assigns (or clears) a teacher for a semester subject.
	SetTeacherAssignment(ctx context.Context, semesterID, subjectID uuid.UUID, teacherID *uuid.UUID) error
}

// ScheduleRepository persists generated schedule versions and their assignments.
type ScheduleRepository interface {
	// Save stores a complete schedule (assignments + metadata) under a new version.
	Save(ctx context.Context, schedule *Schedule) error
	// FindBySemester retrieves the schedule at a specific version.
	FindBySemester(ctx context.Context, semesterID uuid.UUID, version int) (*Schedule, error)
	// FindLatestBySemester retrieves the highest-version schedule for a semester.
	FindLatestBySemester(ctx context.Context, semesterID uuid.UUID) (*Schedule, error)
	// UpdateAssignment modifies a single assignment (manual override).
	UpdateAssignment(ctx context.Context, a *Assignment) error
	// FindAssignmentByID retrieves a single assignment.
	FindAssignmentByID(ctx context.Context, id uuid.UUID) (*Assignment, error)
}
