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
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// PostgresUserRepo implements domain.UserRepository using pgx.
type PostgresUserRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresUserRepo creates a new user repository.
func NewPostgresUserRepo(pool *pgxpool.Pool) *PostgresUserRepo {
	return &PostgresUserRepo{pool: pool}
}

func (r *PostgresUserRepo) schema(ctx context.Context) (string, error) {
	return tenant.FromContext(ctx)
}

func (r *PostgresUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var u domain.User
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			"SELECT id, email, password_hash, name, is_active, created_at, updated_at FROM users WHERE id = $1",
			id,
		).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	})
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}

func (r *PostgresUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var u domain.User
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			"SELECT id, email, password_hash, name, is_active, created_at, updated_at FROM users WHERE email = $1",
			email,
		).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	})
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &u, nil
}

func (r *PostgresUserRepo) Save(ctx context.Context, user *domain.User) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`INSERT INTO users (id, email, password_hash, name, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			user.ID, user.Email, user.PasswordHash, user.Name, user.IsActive, user.CreatedAt, user.UpdatedAt,
		)
		return err
	})
}

func (r *PostgresUserRepo) Update(ctx context.Context, user *domain.User) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE users SET email = $2, name = $3, is_active = $4, updated_at = now() WHERE id = $1`,
			user.ID, user.Email, user.Name, user.IsActive,
		)
		return err
	})
}

func (r *PostgresUserRepo) List(ctx context.Context, offset, limit int) ([]*domain.User, int, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, 0, err
	}

	var users []*domain.User
	var total int
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total); err != nil {
			return err
		}

		rows, err := tx.Query(ctx,
			"SELECT id, email, password_hash, name, is_active, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2",
			limit, offset,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var u domain.User
			if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
				return err
			}
			users = append(users, &u)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	return users, total, nil
}

// Ensure interface compliance.
var _ domain.UserRepository = (*PostgresUserRepo)(nil)

// PostgresUsersLookupRepo implements domain.UsersLookupRepository on public schema.
type PostgresUsersLookupRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresUsersLookupRepo(pool *pgxpool.Pool) *PostgresUsersLookupRepo {
	return &PostgresUsersLookupRepo{pool: pool}
}

func (r *PostgresUsersLookupRepo) FindTenantByEmail(ctx context.Context, email string) (string, error) {
	var schema string
	err := r.pool.QueryRow(ctx,
		"SELECT tenant_schema FROM public.users_lookup WHERE email = $1", email,
	).Scan(&schema)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", erptypes.ErrNotFound
		}
		return "", fmt.Errorf("lookup tenant by email: %w", err)
	}
	return schema, nil
}

func (r *PostgresUsersLookupRepo) Upsert(ctx context.Context, email, tenantSchema string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO public.users_lookup (email, tenant_schema) VALUES ($1, $2)
		 ON CONFLICT (email) DO UPDATE SET tenant_schema = $2`,
		email, tenantSchema,
	)
	return err
}

var _ domain.UsersLookupRepository = (*PostgresUsersLookupRepo)(nil)
