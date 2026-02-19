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
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// PostgresDepartmentRepo implements domain.DepartmentRepository using pgx.
type PostgresDepartmentRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresDepartmentRepo creates a new department repository.
func NewPostgresDepartmentRepo(pool *pgxpool.Pool) *PostgresDepartmentRepo {
	return &PostgresDepartmentRepo{pool: pool}
}

func (r *PostgresDepartmentRepo) schema(ctx context.Context) (string, error) {
	return tenant.FromContext(ctx)
}

func (r *PostgresDepartmentRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Department, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var d domain.Department
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT id, name, description, head_teacher_id, created_at
			 FROM departments WHERE id = $1`,
			id,
		).Scan(&d.ID, &d.Name, &d.Description, &d.HeadTeacherID, &d.CreatedAt)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, erptypes.ErrNotFound
		}
		return nil, fmt.Errorf("find department by id: %w", err)
	}
	return &d, nil
}

func (r *PostgresDepartmentRepo) Save(ctx context.Context, d *domain.Department) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`INSERT INTO departments (id, name, description, head_teacher_id, created_at)
			 VALUES ($1, $2, $3, $4, $5)`,
			d.ID, d.Name, d.Description, d.HeadTeacherID, d.CreatedAt,
		)
		return err
	})
}

func (r *PostgresDepartmentRepo) Update(ctx context.Context, d *domain.Department) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE departments SET name = $2, description = $3, head_teacher_id = $4 WHERE id = $1`,
			d.ID, d.Name, d.Description, d.HeadTeacherID,
		)
		return err
	})
}

func (r *PostgresDepartmentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "DELETE FROM departments WHERE id = $1", id)
		return err
	})
}

func (r *PostgresDepartmentRepo) List(ctx context.Context) ([]*domain.Department, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var depts []*domain.Department
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT id, name, description, head_teacher_id, created_at FROM departments ORDER BY name`,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var d domain.Department
			if err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.HeadTeacherID, &d.CreatedAt); err != nil {
				return err
			}
			depts = append(depts, &d)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}
	return depts, nil
}

// Ensure interface compliance.
var _ domain.DepartmentRepository = (*PostgresDepartmentRepo)(nil)
