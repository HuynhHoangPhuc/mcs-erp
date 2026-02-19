package infrastructure

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/hr/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
)

// PostgresAvailabilityRepo implements domain.AvailabilityRepository using pgx.
type PostgresAvailabilityRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresAvailabilityRepo creates a new availability repository.
func NewPostgresAvailabilityRepo(pool *pgxpool.Pool) *PostgresAvailabilityRepo {
	return &PostgresAvailabilityRepo{pool: pool}
}

func (r *PostgresAvailabilityRepo) schema(ctx context.Context) (string, error) {
	return tenant.FromContext(ctx)
}

// GetByTeacherID returns all availability rows for a given teacher.
func (r *PostgresAvailabilityRepo) GetByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.Availability, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var slots []*domain.Availability
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT teacher_id, day, period, is_available
			 FROM teacher_availability WHERE teacher_id = $1 ORDER BY day, period`,
			teacherID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var a domain.Availability
			if err := rows.Scan(&a.TeacherID, &a.Day, &a.Period, &a.IsAvailable); err != nil {
				return err
			}
			slots = append(slots, &a)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("get availability by teacher: %w", err)
	}
	return slots, nil
}

// SetSlots replaces all availability rows for the given teacher within a single transaction.
// Deletes existing rows first, then inserts the new set.
func (r *PostgresAvailabilityRepo) SetSlots(ctx context.Context, teacherID uuid.UUID, slots []*domain.Availability) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		// Remove all existing slots for this teacher
		if _, err := tx.Exec(ctx,
			"DELETE FROM teacher_availability WHERE teacher_id = $1", teacherID,
		); err != nil {
			return fmt.Errorf("delete old slots: %w", err)
		}

		// Insert new slots
		for _, s := range slots {
			if _, err := tx.Exec(ctx,
				`INSERT INTO teacher_availability (teacher_id, day, period, is_available)
				 VALUES ($1, $2, $3, $4)`,
				s.TeacherID, s.Day, s.Period, s.IsAvailable,
			); err != nil {
				return fmt.Errorf("insert slot day=%d period=%d: %w", s.Day, s.Period, err)
			}
		}
		return nil
	})
}

// Ensure interface compliance.
var _ domain.AvailabilityRepository = (*PostgresAvailabilityRepo)(nil)
