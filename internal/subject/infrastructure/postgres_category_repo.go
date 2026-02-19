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

// PostgresCategoryRepo implements domain.CategoryRepository using pgx.
type PostgresCategoryRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresCategoryRepo creates a new category repository.
func NewPostgresCategoryRepo(pool *pgxpool.Pool) *PostgresCategoryRepo {
	return &PostgresCategoryRepo{pool: pool}
}

func (r *PostgresCategoryRepo) schema(ctx context.Context) (string, error) {
	return tenant.FromContext(ctx)
}

// Save persists a new category.
func (r *PostgresCategoryRepo) Save(ctx context.Context, c *domain.Category) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`INSERT INTO subject_categories (id, name, description, created_at)
			 VALUES ($1, $2, $3, $4)`,
			c.ID, c.Name, c.Description, c.CreatedAt,
		)
		return err
	})
}

// FindByID returns the category with the given id or erptypes.ErrNotFound.
func (r *PostgresCategoryRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var c domain.Category
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT id, name, description, created_at FROM subject_categories WHERE id = $1`,
			id,
		).Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, erptypes.ErrNotFound
		}
		return nil, fmt.Errorf("find category by id: %w", err)
	}
	return &c, nil
}

// List returns all categories ordered by name.
func (r *PostgresCategoryRepo) List(ctx context.Context) ([]*domain.Category, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var categories []*domain.Category

	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT id, name, description, created_at FROM subject_categories ORDER BY name ASC`,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var c domain.Category
			if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt); err != nil {
				return err
			}
			categories = append(categories, &c)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return categories, nil
}

// Ensure interface compliance.
var _ domain.CategoryRepository = (*PostgresCategoryRepo)(nil)
