package infrastructure

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
)

// PostgresRoleRepo implements domain.RoleRepository using pgx.
type PostgresRoleRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRoleRepo(pool *pgxpool.Pool) *PostgresRoleRepo {
	return &PostgresRoleRepo{pool: pool}
}

func (r *PostgresRoleRepo) schema(ctx context.Context) (string, error) {
	return tenant.FromContext(ctx)
}

func (r *PostgresRoleRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var role domain.Role
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			"SELECT id, name, permissions, description, created_at FROM roles WHERE id = $1", id,
		).Scan(&role.ID, &role.Name, &role.Permissions, &role.Description, &role.CreatedAt)
	})
	if err != nil {
		return nil, fmt.Errorf("find role by id: %w", err)
	}
	return &role, nil
}

func (r *PostgresRoleRepo) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var role domain.Role
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			"SELECT id, name, permissions, description, created_at FROM roles WHERE name = $1", name,
		).Scan(&role.ID, &role.Name, &role.Permissions, &role.Description, &role.CreatedAt)
	})
	if err != nil {
		return nil, fmt.Errorf("find role by name: %w", err)
	}
	return &role, nil
}

func (r *PostgresRoleRepo) Save(ctx context.Context, role *domain.Role) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			"INSERT INTO roles (id, name, permissions, description, created_at) VALUES ($1, $2, $3, $4, $5)",
			role.ID, role.Name, role.Permissions, role.Description, role.CreatedAt,
		)
		return err
	})
}

func (r *PostgresRoleRepo) Update(ctx context.Context, role *domain.Role) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			"UPDATE roles SET name = $2, permissions = $3, description = $4 WHERE id = $1",
			role.ID, role.Name, role.Permissions, role.Description,
		)
		return err
	})
}

func (r *PostgresRoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "DELETE FROM roles WHERE id = $1", id)
		return err
	})
}

func (r *PostgresRoleRepo) List(ctx context.Context) ([]*domain.Role, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var roles []*domain.Role
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, "SELECT id, name, permissions, description, created_at FROM roles ORDER BY name")
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var role domain.Role
			if err := rows.Scan(&role.ID, &role.Name, &role.Permissions, &role.Description, &role.CreatedAt); err != nil {
				return err
			}
			roles = append(roles, &role)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	return roles, nil
}

func (r *PostgresRoleRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var roles []*domain.Role
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT r.id, r.name, r.permissions, r.description, r.created_at
			 FROM roles r JOIN user_roles ur ON r.id = ur.role_id
			 WHERE ur.user_id = $1`, userID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var role domain.Role
			if err := rows.Scan(&role.ID, &role.Name, &role.Permissions, &role.Description, &role.CreatedAt); err != nil {
				return err
			}
			roles = append(roles, &role)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("find roles by user: %w", err)
	}
	return roles, nil
}

func (r *PostgresRoleRepo) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, roleID)
		return err
	})
}

func (r *PostgresRoleRepo) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2", userID, roleID)
		return err
	})
}

var _ domain.RoleRepository = (*PostgresRoleRepo)(nil)
