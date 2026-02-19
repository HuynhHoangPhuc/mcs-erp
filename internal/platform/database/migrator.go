package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrator handles per-tenant schema migrations.
type Migrator struct {
	pool *pgxpool.Pool
}

// NewMigrator creates a new migration runner.
func NewMigrator(pool *pgxpool.Pool) *Migrator {
	return &Migrator{pool: pool}
}

// EnsureTemplateSchema creates the _template schema used for sqlc codegen.
func (m *Migrator) EnsureTemplateSchema(ctx context.Context) error {
	return CreateSchema(ctx, m.pool, "_template")
}

// ActiveTenantSchemas queries the public.tenants table for all active tenant schema names.
func (m *Migrator) ActiveTenantSchemas(ctx context.Context) ([]string, error) {
	rows, err := m.pool.Query(ctx, "SELECT schema_name FROM public.tenants WHERE is_active = true")
	if err != nil {
		return nil, fmt.Errorf("query tenant schemas: %w", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, fmt.Errorf("scan schema: %w", err)
		}
		schemas = append(schemas, s)
	}
	return schemas, rows.Err()
}

// MigrateAll applies the given SQL to _template and all active tenant schemas.
func (m *Migrator) MigrateAll(ctx context.Context, sql string) error {
	slog.Info("migrating _template schema")
	if err := m.runSQLInSchema(ctx, "_template", sql); err != nil {
		return fmt.Errorf("migrate _template: %w", err)
	}

	schemas, err := m.ActiveTenantSchemas(ctx)
	if err != nil {
		return err
	}

	for _, schema := range schemas {
		slog.Info("migrating tenant schema", "schema", schema)
		if err := m.runSQLInSchema(ctx, schema, sql); err != nil {
			return fmt.Errorf("migrate %s: %w", schema, err)
		}
	}
	return nil
}

func (m *Migrator) runSQLInSchema(ctx context.Context, schema string, sqlStr string) error {
	return WithTenantTx(ctx, m.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sqlStr)
		return err
	})
}
