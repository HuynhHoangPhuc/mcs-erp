package database

import (
	"context"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var schemaNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// NewPool creates a pgxpool connected to the given database URL.
func NewPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}
	cfg.MinConns = 5
	cfg.MaxConns = 25

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return pool, nil
}

// SetTenantSchema sets the search_path for the current transaction.
// MUST be called inside a transaction — uses SET LOCAL which is tx-scoped.
func SetTenantSchema(ctx context.Context, tx pgx.Tx, schema string) error {
	if !schemaNameRegex.MatchString(schema) {
		return fmt.Errorf("invalid schema name: %q", schema)
	}
	_, err := tx.Exec(ctx, fmt.Sprintf("SET LOCAL search_path = %s, public", schema))
	return err
}

// CreateSchema creates a tenant schema if it does not exist.
func CreateSchema(ctx context.Context, pool *pgxpool.Pool, schema string) error {
	if !schemaNameRegex.MatchString(schema) {
		return fmt.Errorf("invalid schema name: %q", schema)
	}
	_, err := pool.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema))
	return err
}

// WithTenantTx begins a transaction with the search_path set to the given tenant schema.
func WithTenantTx(ctx context.Context, pool *pgxpool.Pool, schema string, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := SetTenantSchema(ctx, tx, schema); err != nil {
		return fmt.Errorf("set tenant schema: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
