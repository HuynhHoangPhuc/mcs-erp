package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const (
	defaultTestDatabaseURL = "postgres://mcs:mcs_dev_pass@localhost:5432/mcs_erp_test?sslmode=disable"
	operationTimeout       = 20 * time.Second
)

var migrationDirs = []string{
	"migrations",
	"migrations/core",
	"migrations/hr",
	"migrations/subject",
	"migrations/room",
	"migrations/timetable",
	"migrations/agent",
}

var (
	gooseInitOnce sync.Once
	gooseInitErr  error
	gooseMu       sync.Mutex
)

// TestDB wraps a pgxpool.Pool connected to the integration test database.
type TestDB struct {
	Pool        *pgxpool.Pool
	databaseURL string
}

// NewTestDB connects to the integration database and ensures it exists.
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if databaseURL == "" {
		databaseURL = defaultTestDatabaseURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	if err := ensureDatabase(ctx, databaseURL); err != nil {
		t.Fatalf("ensure test database: %v", err)
	}

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse test database config: %v", err)
	}
	cfg.MinConns = 0
	cfg.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping test database: %v", err)
	}
	if _, err := pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto`); err != nil {
		pool.Close()
		t.Fatalf("ensure pgcrypto extension: %v", err)
	}

	db := &TestDB{Pool: pool, databaseURL: databaseURL}
	t.Cleanup(func() { db.Close(t) })
	return db
}

// CreateTenantSchema creates a unique tenant schema and applies all migrations.
func (db *TestDB) CreateTenantSchema(t *testing.T) string {
	t.Helper()

	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "_")
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	if _, err := db.Pool.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+pgx.Identifier{schema}.Sanitize()); err != nil {
		t.Fatalf("create tenant schema %q: %v", schema, err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), operationTimeout)
		defer cleanupCancel()
		_, _ = db.Pool.Exec(cleanupCtx, "DELETE FROM public.users_lookup WHERE tenant_schema = $1", schema)
		_, _ = db.Pool.Exec(cleanupCtx, "DELETE FROM public.tenants WHERE schema_name = $1", schema)
		_, _ = db.Pool.Exec(cleanupCtx, "DROP SCHEMA IF EXISTS "+pgx.Identifier{schema}.Sanitize()+" CASCADE")
	})

	if err := db.applyMigrations(ctx, schema); err != nil {
		t.Fatalf("apply migrations for schema %q: %v", schema, err)
	}

	return schema
}

// Close closes the test pool.
func (db *TestDB) Close(t *testing.T) {
	t.Helper()
	if db != nil && db.Pool != nil {
		db.Pool.Close()
	}
}

func (db *TestDB) applyMigrations(ctx context.Context, schema string) error {
	if err := initGoose(); err != nil {
		return err
	}

	sqlDB, err := db.openSchemaSQLDB(schema)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	root, err := projectRoot()
	if err != nil {
		return err
	}

	gooseMu.Lock()
	defer gooseMu.Unlock()

	for _, rel := range migrationDirs {
		sourceDir := filepath.Join(root, rel)
		upOnlyDir, cleanup, err := prepareUpOnlyMigrationDir(sourceDir)
		if err != nil {
			return fmt.Errorf("prepare migrations %s: %w", rel, err)
		}

		goose.SetTableName(gooseTableName(schema, rel))
		if err := goose.UpContext(ctx, sqlDB, upOnlyDir, goose.WithAllowMissing()); err != nil {
			cleanup()
			return fmt.Errorf("goose up %s: %w", rel, err)
		}
		cleanup()
	}
	return nil
}

func (db *TestDB) openSchemaSQLDB(schema string) (*sql.DB, error) {
	cfg, err := pgx.ParseConfig(db.databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}
	if cfg.RuntimeParams == nil {
		cfg.RuntimeParams = map[string]string{}
	}
	cfg.RuntimeParams["search_path"] = schema + ",public"

	sqlDB := stdlib.OpenDB(*cfg)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	return sqlDB, nil
}

func initGoose() error {
	gooseInitOnce.Do(func() {
		goose.SetLogger(log.New(io.Discard, "", 0))
		gooseInitErr = goose.SetDialect("postgres")
	})
	return gooseInitErr
}

func ensureDatabase(ctx context.Context, databaseURL string) error {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("parse database url: %w", err)
	}
	dbName := strings.TrimPrefix(parsed.Path, "/")
	if dbName == "" {
		return fmt.Errorf("database name missing in URL")
	}

	adminURL := *parsed
	adminURL.Path = "/postgres"

	cfg, err := pgxpool.ParseConfig(adminURL.String())
	if err != nil {
		return fmt.Errorf("parse admin database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("connect admin database: %w", err)
	}
	defer pool.Close()

	var exists bool
	if err := pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists); err != nil {
		return fmt.Errorf("check database existence: %w", err)
	}
	if exists {
		return nil
	}

	_, err = pool.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{dbName}.Sanitize())
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "already exists") {
		return fmt.Errorf("create database %q: %w", dbName, err)
	}
	return nil
}

func projectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found from %s", wd)
}

func gooseTableName(schema, rel string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(schema + "/" + rel))
	return fmt.Sprintf("goose_%x", h.Sum64())
}

func prepareUpOnlyMigrationDir(sourceDir string) (string, func(), error) {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return "", nil, err
	}

	tmpDir, err := os.MkdirTemp("", "mcs-erp-up-migrations-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		srcPath := filepath.Join(sourceDir, name)
		dstPath := filepath.Join(tmpDir, name)

		content, err := os.ReadFile(srcPath)
		if err != nil {
			cleanup()
			return "", nil, err
		}
		content = wrapGooseSQLMigration(content)
		if err := os.WriteFile(dstPath, content, 0o644); err != nil {
			cleanup()
			return "", nil, err
		}
	}

	return tmpDir, cleanup, nil
}

func wrapGooseSQLMigration(content []byte) []byte {
	s := strings.TrimSpace(string(content))
	if strings.Contains(s, "-- +goose Up") {
		return []byte(s + "\n")
	}
	return []byte("-- +goose Up\n-- +goose StatementBegin\n" + s + "\n-- +goose StatementEnd\n")
}
