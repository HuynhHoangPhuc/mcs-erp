package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// PostgresScheduleRepo implements domain.ScheduleRepository using pgx.
type PostgresScheduleRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresScheduleRepo creates a new schedule repository.
func NewPostgresScheduleRepo(pool *pgxpool.Pool) *PostgresScheduleRepo {
	return &PostgresScheduleRepo{pool: pool}
}

func (r *PostgresScheduleRepo) schema(ctx context.Context) (string, error) {
	return tenant.FromContext(ctx)
}

// Save persists a complete schedule and all its assignments within a single transaction.
func (r *PostgresScheduleRepo) Save(ctx context.Context, sched *domain.Schedule) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		// Upsert the schedule metadata row.
		_, err := tx.Exec(ctx,
			`INSERT INTO schedules (semester_id, version, hard_violations, soft_penalty, generated_at)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (semester_id, version) DO UPDATE
			   SET hard_violations = EXCLUDED.hard_violations,
			       soft_penalty    = EXCLUDED.soft_penalty,
			       generated_at    = EXCLUDED.generated_at`,
			sched.SemesterID, sched.Version,
			sched.HardViolations, sched.SoftPenalty, sched.GeneratedAt,
		)
		if err != nil {
			return fmt.Errorf("upsert schedule: %w", err)
		}

		// Delete existing assignments for this version before re-inserting.
		if _, err := tx.Exec(ctx,
			`DELETE FROM assignments WHERE semester_id = $1 AND version = $2`,
			sched.SemesterID, sched.Version,
		); err != nil {
			return fmt.Errorf("delete old assignments: %w", err)
		}

		// Bulk-insert assignments.
		for _, a := range sched.Assignments {
			if a.ID == uuid.Nil {
				a.ID = uuid.New()
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO assignments
				   (id, semester_id, subject_id, teacher_id, room_id, day, period, version)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				a.ID, sched.SemesterID, a.SubjectID, a.TeacherID, a.RoomID,
				a.Day, a.Period, sched.Version,
			); err != nil {
				return fmt.Errorf("insert assignment: %w", err)
			}
		}
		return nil
	})
}

// FindBySemester retrieves the schedule at a specific version.
func (r *PostgresScheduleRepo) FindBySemester(
	ctx context.Context, semesterID uuid.UUID, version int,
) (*domain.Schedule, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var sched domain.Schedule
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		var genAt time.Time
		err := tx.QueryRow(ctx,
			`SELECT semester_id, version, hard_violations, soft_penalty, generated_at
			 FROM schedules WHERE semester_id = $1 AND version = $2`,
			semesterID, version,
		).Scan(&sched.SemesterID, &sched.Version,
			&sched.HardViolations, &sched.SoftPenalty, &genAt)
		if err != nil {
			return err
		}
		sched.GeneratedAt = genAt

		return r.loadAssignments(ctx, tx, semesterID, version, &sched)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, erptypes.ErrNotFound
		}
		return nil, fmt.Errorf("find schedule: %w", err)
	}
	return &sched, nil
}

// FindLatestBySemester retrieves the highest-version schedule for a semester.
func (r *PostgresScheduleRepo) FindLatestBySemester(
	ctx context.Context, semesterID uuid.UUID,
) (*domain.Schedule, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var sched domain.Schedule
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx,
			`SELECT semester_id, version, hard_violations, soft_penalty, generated_at
			 FROM schedules WHERE semester_id = $1
			 ORDER BY version DESC LIMIT 1`,
			semesterID,
		).Scan(&sched.SemesterID, &sched.Version,
			&sched.HardViolations, &sched.SoftPenalty, &sched.GeneratedAt)
		if err != nil {
			return err
		}
		return r.loadAssignments(ctx, tx, semesterID, sched.Version, &sched)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, erptypes.ErrNotFound
		}
		return nil, fmt.Errorf("find latest schedule: %w", err)
	}
	return &sched, nil
}

// UpdateAssignment modifies a single assignment row in place.
func (r *PostgresScheduleRepo) UpdateAssignment(ctx context.Context, a *domain.Assignment) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE assignments
			 SET teacher_id = $2, room_id = $3, day = $4, period = $5
			 WHERE id = $1`,
			a.ID, a.TeacherID, a.RoomID, a.Day, a.Period,
		)
		return err
	})
}

// FindAssignmentByID retrieves a single assignment by its primary key.
func (r *PostgresScheduleRepo) FindAssignmentByID(
	ctx context.Context, id uuid.UUID,
) (*domain.Assignment, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var a domain.Assignment
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT id, semester_id, subject_id, teacher_id, room_id, day, period, version
			 FROM assignments WHERE id = $1`,
			id,
		).Scan(&a.ID, &a.SemesterID, &a.SubjectID, &a.TeacherID, &a.RoomID,
			&a.Day, &a.Period, &a.Version)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, erptypes.ErrNotFound
		}
		return nil, fmt.Errorf("find assignment by id: %w", err)
	}
	return &a, nil
}

// loadAssignments fetches all assignments for a given semester+version into sched.
func (r *PostgresScheduleRepo) loadAssignments(
	ctx context.Context, tx pgx.Tx,
	semesterID uuid.UUID, version int,
	sched *domain.Schedule,
) error {
	rows, err := tx.Query(ctx,
		`SELECT id, semester_id, subject_id, teacher_id, room_id, day, period, version
		 FROM assignments WHERE semester_id = $1 AND version = $2
		 ORDER BY day, period`,
		semesterID, version,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var a domain.Assignment
		if err := rows.Scan(&a.ID, &a.SemesterID, &a.SubjectID, &a.TeacherID,
			&a.RoomID, &a.Day, &a.Period, &a.Version); err != nil {
			return err
		}
		sched.Assignments = append(sched.Assignments, a)
	}
	return rows.Err()
}

// Ensure interface compliance.
var _ domain.ScheduleRepository = (*PostgresScheduleRepo)(nil)
