package domain

import (
	"context"

	"github.com/google/uuid"
)

// SubjectRepository defines persistence operations for Subject entities.
type SubjectRepository interface {
	Save(ctx context.Context, subject *Subject) error
	FindByID(ctx context.Context, id uuid.UUID) (*Subject, error)
	FindByCode(ctx context.Context, code string) (*Subject, error)
	Update(ctx context.Context, subject *Subject) error
	List(ctx context.Context, offset, limit int) ([]*Subject, int, error)
	ListByCategory(ctx context.Context, categoryID uuid.UUID, offset, limit int) ([]*Subject, int, error)
}

// CategoryRepository defines persistence operations for Category entities.
type CategoryRepository interface {
	Save(ctx context.Context, category *Category) error
	FindByID(ctx context.Context, id uuid.UUID) (*Category, error)
	List(ctx context.Context) ([]*Category, error)
}

// PrerequisiteRepository manages the prerequisite edge set with optimistic locking.
// The version is stored per-subject: each subject has a single version counter
// that is incremented on every AddEdge or RemoveEdge for that subject.
type PrerequisiteRepository interface {
	// GetEdges returns all prerequisite edges for the given subject.
	GetEdges(ctx context.Context, subjectID uuid.UUID) ([]PrerequisiteEdge, error)
	// GetAllEdges returns every edge in the tenant (for full graph construction).
	GetAllEdges(ctx context.Context) ([]PrerequisiteEdge, error)
	// GetVersion returns the current optimistic-locking version for a subject's prerequisite set.
	GetVersion(ctx context.Context, subjectID uuid.UUID) (int, error)
	// AddEdge inserts a new prerequisite edge.
	// Returns erptypes.ErrConflict if expectedVersion does not match the stored version.
	AddEdge(ctx context.Context, subjectID, prerequisiteID uuid.UUID, expectedVersion int) error
	// RemoveEdge deletes a prerequisite edge.
	// Returns erptypes.ErrConflict if expectedVersion does not match the stored version.
	RemoveEdge(ctx context.Context, subjectID, prerequisiteID uuid.UUID, expectedVersion int) error
}
