package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/hr/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// PostgresTeacherRepo implements domain.TeacherRepository using pgx.
type PostgresTeacherRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresTeacherRepo creates a new teacher repository.
func NewPostgresTeacherRepo(pool *pgxpool.Pool) *PostgresTeacherRepo {
	return &PostgresTeacherRepo{pool: pool}
}

func (r *PostgresTeacherRepo) schema(ctx context.Context) (string, error) {
	return tenant.FromContext(ctx)
}

func (r *PostgresTeacherRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Teacher, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var t domain.Teacher
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT id, name, email, department_id, qualifications, is_active, created_at, updated_at
			 FROM teachers WHERE id = $1`,
			id,
		).Scan(&t.ID, &t.Name, &t.Email, &t.DepartmentID, &t.Qualifications, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, erptypes.ErrNotFound
		}
		return nil, fmt.Errorf("find teacher by id: %w", err)
	}
	return &t, nil
}

func (r *PostgresTeacherRepo) FindByEmail(ctx context.Context, email string) (*domain.Teacher, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var t domain.Teacher
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT id, name, email, department_id, qualifications, is_active, created_at, updated_at
			 FROM teachers WHERE email = $1`,
			email,
		).Scan(&t.ID, &t.Name, &t.Email, &t.DepartmentID, &t.Qualifications, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, erptypes.ErrNotFound
		}
		return nil, fmt.Errorf("find teacher by email: %w", err)
	}
	return &t, nil
}

func (r *PostgresTeacherRepo) Save(ctx context.Context, t *domain.Teacher) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`INSERT INTO teachers (id, name, email, department_id, qualifications, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			t.ID, t.Name, t.Email, t.DepartmentID, t.Qualifications, t.IsActive, t.CreatedAt, t.UpdatedAt,
		)
		return err
	})
}

func (r *PostgresTeacherRepo) Update(ctx context.Context, t *domain.Teacher) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE teachers
			 SET name = $2, email = $3, department_id = $4, qualifications = $5, is_active = $6, updated_at = now()
			 WHERE id = $1`,
			t.ID, t.Name, t.Email, t.DepartmentID, t.Qualifications, t.IsActive,
		)
		return err
	})
}

func (r *PostgresTeacherRepo) List(ctx context.Context, filter domain.TeacherFilter, offset, limit int) ([]*domain.Teacher, int, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Build dynamic WHERE clause
	conds := []string{}
	args := []any{}
	argIdx := 1

	if filter.DepartmentID != nil {
		conds = append(conds, fmt.Sprintf("department_id = $%d", argIdx))
		args = append(args, *filter.DepartmentID)
		argIdx++
	}
	if filter.IsActive != nil {
		conds = append(conds, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *filter.IsActive)
		argIdx++
	}
	if filter.Qualification != "" {
		conds = append(conds, fmt.Sprintf("$%d = ANY(qualifications)", argIdx))
		args = append(args, filter.Qualification)
		argIdx++
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	var teachers []*domain.Teacher
	var total int

	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM teachers %s", where)
		if err := tx.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
			return err
		}

		listArgs := append(args, limit, offset)
		listQuery := fmt.Sprintf(
			`SELECT id, name, email, department_id, qualifications, is_active, created_at, updated_at
			 FROM teachers %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
			where, argIdx, argIdx+1,
		)
		rows, err := tx.Query(ctx, listQuery, listArgs...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var t domain.Teacher
			if err := rows.Scan(&t.ID, &t.Name, &t.Email, &t.DepartmentID, &t.Qualifications,
				&t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
				return err
			}
			teachers = append(teachers, &t)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list teachers: %w", err)
	}
	return teachers, total, nil
}

// Ensure interface compliance.
var _ domain.TeacherRepository = (*PostgresTeacherRepo)(nil)
