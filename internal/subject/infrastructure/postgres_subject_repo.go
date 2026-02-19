package infrastructure

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/subject/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// PostgresSubjectRepo implements domain.SubjectRepository using pgx.
type PostgresSubjectRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresSubjectRepo creates a new subject repository.
func NewPostgresSubjectRepo(pool *pgxpool.Pool) *PostgresSubjectRepo {
	return &PostgresSubjectRepo{pool: pool}
}

func (r *PostgresSubjectRepo) schema(ctx context.Context) (string, error) {
	return tenant.FromContext(ctx)
}

// Save persists a new subject.
func (r *PostgresSubjectRepo) Save(ctx context.Context, s *domain.Subject) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`INSERT INTO subjects (id, name, code, description, category_id, credits, hours_per_week, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			s.ID, s.Name, s.Code, s.Description, s.CategoryID, s.Credits, s.HoursPerWeek, s.IsActive, s.CreatedAt, s.UpdatedAt,
		)
		return err
	})
}

// FindByID returns the subject with the given id or erptypes.ErrNotFound.
func (r *PostgresSubjectRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Subject, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var s domain.Subject
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT id, name, code, description, category_id, credits, hours_per_week, is_active, created_at, updated_at
			 FROM subjects WHERE id = $1`,
			id,
		).Scan(&s.ID, &s.Name, &s.Code, &s.Description, &s.CategoryID,
			&s.Credits, &s.HoursPerWeek, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, erptypes.ErrNotFound
		}
		return nil, fmt.Errorf("find subject by id: %w", err)
	}
	return &s, nil
}

// FindByCode returns the subject with the given code or erptypes.ErrNotFound.
func (r *PostgresSubjectRepo) FindByCode(ctx context.Context, code string) (*domain.Subject, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var s domain.Subject
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT id, name, code, description, category_id, credits, hours_per_week, is_active, created_at, updated_at
			 FROM subjects WHERE code = $1`,
			code,
		).Scan(&s.ID, &s.Name, &s.Code, &s.Description, &s.CategoryID,
			&s.Credits, &s.HoursPerWeek, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, erptypes.ErrNotFound
		}
		return nil, fmt.Errorf("find subject by code: %w", err)
	}
	return &s, nil
}

// Update persists changes to an existing subject.
func (r *PostgresSubjectRepo) Update(ctx context.Context, s *domain.Subject) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE subjects
			 SET name = $2, code = $3, description = $4, category_id = $5,
			     credits = $6, hours_per_week = $7, is_active = $8, updated_at = now()
			 WHERE id = $1`,
			s.ID, s.Name, s.Code, s.Description, s.CategoryID,
			s.Credits, s.HoursPerWeek, s.IsActive,
		)
		return err
	})
}

// List returns a paginated slice of subjects and the total count.
func (r *PostgresSubjectRepo) List(ctx context.Context, offset, limit int) ([]*domain.Subject, int, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, 0, err
	}

	var subjects []*domain.Subject
	var total int

	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM subjects").Scan(&total); err != nil {
			return err
		}

		rows, err := tx.Query(ctx,
			`SELECT id, name, code, description, category_id, credits, hours_per_week, is_active, created_at, updated_at
			 FROM subjects ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			limit, offset,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var s domain.Subject
			if err := rows.Scan(&s.ID, &s.Name, &s.Code, &s.Description, &s.CategoryID,
				&s.Credits, &s.HoursPerWeek, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
				return err
			}
			subjects = append(subjects, &s)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list subjects: %w", err)
	}
	return subjects, total, nil
}

// ListByCategory returns paginated subjects filtered by category.
func (r *PostgresSubjectRepo) ListByCategory(ctx context.Context, categoryID uuid.UUID, offset, limit int) ([]*domain.Subject, int, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, 0, err
	}

	var subjects []*domain.Subject
	var total int

	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx,
			"SELECT COUNT(*) FROM subjects WHERE category_id = $1", categoryID,
		).Scan(&total); err != nil {
			return err
		}

		rows, err := tx.Query(ctx,
			`SELECT id, name, code, description, category_id, credits, hours_per_week, is_active, created_at, updated_at
			 FROM subjects WHERE category_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			categoryID, limit, offset,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var s domain.Subject
			if err := rows.Scan(&s.ID, &s.Name, &s.Code, &s.Description, &s.CategoryID,
				&s.Credits, &s.HoursPerWeek, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
				return err
			}
			subjects = append(subjects, &s)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list subjects by category: %w", err)
	}
	return subjects, total, nil
}

// Ensure interface compliance.
var _ domain.SubjectRepository = (*PostgresSubjectRepo)(nil)
