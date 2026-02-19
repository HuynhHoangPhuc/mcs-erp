package infrastructure

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// PostgresSemesterRepo implements domain.SemesterRepository using pgx.
type PostgresSemesterRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresSemesterRepo creates a new semester repository.
func NewPostgresSemesterRepo(pool *pgxpool.Pool) *PostgresSemesterRepo {
	return &PostgresSemesterRepo{pool: pool}
}

func (r *PostgresSemesterRepo) schema(ctx context.Context) (string, error) {
	return tenant.FromContext(ctx)
}

func (r *PostgresSemesterRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Semester, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var s domain.Semester
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT id, name, start_date, end_date, status, created_at, updated_at
			 FROM semesters WHERE id = $1`,
			id,
		).Scan(&s.ID, &s.Name, &s.StartDate, &s.EndDate, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, erptypes.ErrNotFound
		}
		return nil, fmt.Errorf("find semester by id: %w", err)
	}
	return &s, nil
}

func (r *PostgresSemesterRepo) Save(ctx context.Context, s *domain.Semester) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`INSERT INTO semesters (id, name, start_date, end_date, status, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			s.ID, s.Name, s.StartDate, s.EndDate, s.Status, s.CreatedAt, s.UpdatedAt,
		)
		return err
	})
}

func (r *PostgresSemesterRepo) Update(ctx context.Context, s *domain.Semester) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE semesters
			 SET name = $2, start_date = $3, end_date = $4, status = $5, updated_at = now()
			 WHERE id = $1`,
			s.ID, s.Name, s.StartDate, s.EndDate, s.Status,
		)
		return err
	})
}

func (r *PostgresSemesterRepo) List(ctx context.Context, offset, limit int) ([]*domain.Semester, int, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, 0, err
	}

	var semesters []*domain.Semester
	var total int

	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM semesters`).Scan(&total); err != nil {
			return err
		}

		rows, err := tx.Query(ctx,
			`SELECT id, name, start_date, end_date, status, created_at, updated_at
			 FROM semesters ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			limit, offset,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var s domain.Semester
			if err := rows.Scan(&s.ID, &s.Name, &s.StartDate, &s.EndDate, &s.Status,
				&s.CreatedAt, &s.UpdatedAt); err != nil {
				return err
			}
			semesters = append(semesters, &s)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list semesters: %w", err)
	}
	return semesters, total, nil
}

func (r *PostgresSemesterRepo) AddSubjects(ctx context.Context, semesterID uuid.UUID, subjectIDs []uuid.UUID) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		for _, sid := range subjectIDs {
			_, err := tx.Exec(ctx,
				`INSERT INTO semester_subjects (semester_id, subject_id)
				 VALUES ($1, $2)
				 ON CONFLICT (semester_id, subject_id) DO NOTHING`,
				semesterID, sid,
			)
			if err != nil {
				return fmt.Errorf("add subject %s: %w", sid, err)
			}
		}
		return nil
	})
}

func (r *PostgresSemesterRepo) GetSubjects(ctx context.Context, semesterID uuid.UUID) ([]*domain.SemesterSubject, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var subjects []*domain.SemesterSubject
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT semester_id, subject_id, teacher_id
			 FROM semester_subjects WHERE semester_id = $1`,
			semesterID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var ss domain.SemesterSubject
			if err := rows.Scan(&ss.SemesterID, &ss.SubjectID, &ss.TeacherID); err != nil {
				return err
			}
			subjects = append(subjects, &ss)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("get semester subjects: %w", err)
	}
	return subjects, nil
}

func (r *PostgresSemesterRepo) SetTeacherAssignment(
	ctx context.Context, semesterID, subjectID uuid.UUID, teacherID *uuid.UUID,
) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE semester_subjects SET teacher_id = $3
			 WHERE semester_id = $1 AND subject_id = $2`,
			semesterID, subjectID, teacherID,
		)
		return err
	})
}

// Ensure interface compliance.
var _ domain.SemesterRepository = (*PostgresSemesterRepo)(nil)
