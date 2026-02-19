package domain

import (
	"context"

	"github.com/google/uuid"
)

// TeacherFilter holds optional filters for listing teachers.
type TeacherFilter struct {
	DepartmentID  *uuid.UUID
	IsActive      *bool
	Qualification string // filter by any matching qualification
}

// TeacherRepository defines persistence operations for Teacher entities.
type TeacherRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Teacher, error)
	FindByEmail(ctx context.Context, email string) (*Teacher, error)
	Save(ctx context.Context, teacher *Teacher) error
	Update(ctx context.Context, teacher *Teacher) error
	List(ctx context.Context, filter TeacherFilter, offset, limit int) ([]*Teacher, int, error)
}

// DepartmentRepository defines persistence operations for Department entities.
type DepartmentRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Department, error)
	Save(ctx context.Context, dept *Department) error
	Update(ctx context.Context, dept *Department) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*Department, error)
}

// AvailabilityRepository manages teacher weekly slot availability.
type AvailabilityRepository interface {
	// GetByTeacherID returns all stored slots for the given teacher.
	GetByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*Availability, error)
	// SetSlots replaces all availability rows for the teacher (upsert + delete).
	SetSlots(ctx context.Context, teacherID uuid.UUID, slots []*Availability) error
}
