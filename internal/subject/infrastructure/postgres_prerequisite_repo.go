package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/subject/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// PostgresPrerequisiteRepo implements domain.PrerequisiteRepository using pgx.
// Optimistic locking is implemented via a version counter stored in a separate
// subject_prerequisite_versions table. Version starts at 0 and increments on
// every structural change to a subject's prerequisite set.
type PostgresPrerequisiteRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresPrerequisiteRepo creates a new prerequisite repository.
func NewPostgresPrerequisiteRepo(pool *pgxpool.Pool) *PostgresPrerequisiteRepo {
	return &PostgresPrerequisiteRepo{pool: pool}
}

func (r *PostgresPrerequisiteRepo) schema(ctx context.Context) (string, error) {
	return tenant.FromContext(ctx)
}

// GetEdges returns all prerequisite edges where the given subject is the dependent.
func (r *PostgresPrerequisiteRepo) GetEdges(ctx context.Context, subjectID uuid.UUID) ([]domain.PrerequisiteEdge, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var edges []domain.PrerequisiteEdge
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT sp.subject_id, sp.prerequisite_id, COALESCE(v.version, 1)
			 FROM subject_prerequisites sp
			 LEFT JOIN subject_prerequisite_versions v ON v.subject_id = sp.subject_id
			 WHERE sp.subject_id = $1`,
			subjectID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var e domain.PrerequisiteEdge
			if err := rows.Scan(&e.SubjectID, &e.PrerequisiteID, &e.Version); err != nil {
				return err
			}
			edges = append(edges, e)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("get prerequisite edges: %w", err)
	}
	return edges, nil
}

// GetAllEdges returns every prerequisite edge in the tenant (for full graph construction).
func (r *PostgresPrerequisiteRepo) GetAllEdges(ctx context.Context) ([]domain.PrerequisiteEdge, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var edges []domain.PrerequisiteEdge
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT sp.subject_id, sp.prerequisite_id, COALESCE(v.version, 1)
			 FROM subject_prerequisites sp
			 LEFT JOIN subject_prerequisite_versions v ON v.subject_id = sp.subject_id`,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var e domain.PrerequisiteEdge
			if err := rows.Scan(&e.SubjectID, &e.PrerequisiteID, &e.Version); err != nil {
				return err
			}
			edges = append(edges, e)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("get all prerequisite edges: %w", err)
	}
	return edges, nil
}

// GetVersion returns the current optimistic-locking version for a subject's prerequisite set.
// Returns 0 if no version row exists yet (first mutation will create it).
func (r *PostgresPrerequisiteRepo) GetVersion(ctx context.Context, subjectID uuid.UUID) (int, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return 0, err
	}

	var version int
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx,
			`SELECT version FROM subject_prerequisite_versions WHERE subject_id = $1`,
			subjectID,
		).Scan(&version)
		if errors.Is(err, pgx.ErrNoRows) {
			version = 0
			return nil
		}
		return err
	})
	if err != nil {
		return 0, fmt.Errorf("get prerequisite version: %w", err)
	}
	return version, nil
}

// AddEdge inserts a prerequisite edge and increments the version counter.
// Returns erptypes.ErrConflict if expectedVersion does not match the stored version.
func (r *PostgresPrerequisiteRepo) AddEdge(ctx context.Context, subjectID, prerequisiteID uuid.UUID, expectedVersion int) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		// Read current version with a row-level lock to prevent concurrent mutations.
		var currentVersion int
		err := tx.QueryRow(ctx,
			`SELECT version FROM subject_prerequisite_versions WHERE subject_id = $1 FOR UPDATE`,
			subjectID,
		).Scan(&currentVersion)

		if errors.Is(err, pgx.ErrNoRows) {
			// No version row yet — treat as version 0.
			currentVersion = 0
		} else if err != nil {
			return fmt.Errorf("lock prerequisite version: %w", err)
		}

		if currentVersion != expectedVersion {
			return fmt.Errorf("version mismatch (expected %d, got %d): %w",
				expectedVersion, currentVersion, erptypes.ErrConflict)
		}

		// Insert the edge.
		_, err = tx.Exec(ctx,
			`INSERT INTO subject_prerequisites (subject_id, prerequisite_id)
			 VALUES ($1, $2)
			 ON CONFLICT DO NOTHING`,
			subjectID, prerequisiteID,
		)
		if err != nil {
			return fmt.Errorf("insert prerequisite edge: %w", err)
		}

		// Upsert the version row, incrementing by 1.
		_, err = tx.Exec(ctx,
			`INSERT INTO subject_prerequisite_versions (subject_id, version)
			 VALUES ($1, 1)
			 ON CONFLICT (subject_id) DO UPDATE SET version = subject_prerequisite_versions.version + 1`,
			subjectID,
		)
		return err
	})
}

// RemoveEdge deletes a prerequisite edge and increments the version counter.
// Returns erptypes.ErrConflict if expectedVersion does not match the stored version.
func (r *PostgresPrerequisiteRepo) RemoveEdge(ctx context.Context, subjectID, prerequisiteID uuid.UUID, expectedVersion int) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		var currentVersion int
		err := tx.QueryRow(ctx,
			`SELECT version FROM subject_prerequisite_versions WHERE subject_id = $1 FOR UPDATE`,
			subjectID,
		).Scan(&currentVersion)

		if errors.Is(err, pgx.ErrNoRows) {
			currentVersion = 0
		} else if err != nil {
			return fmt.Errorf("lock prerequisite version: %w", err)
		}

		if currentVersion != expectedVersion {
			return fmt.Errorf("version mismatch (expected %d, got %d): %w",
				expectedVersion, currentVersion, erptypes.ErrConflict)
		}

		tag, err := tx.Exec(ctx,
			`DELETE FROM subject_prerequisites WHERE subject_id = $1 AND prerequisite_id = $2`,
			subjectID, prerequisiteID,
		)
		if err != nil {
			return fmt.Errorf("delete prerequisite edge: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return erptypes.ErrNotFound
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO subject_prerequisite_versions (subject_id, version)
			 VALUES ($1, 1)
			 ON CONFLICT (subject_id) DO UPDATE SET version = subject_prerequisite_versions.version + 1`,
			subjectID,
		)
		return err
	})
}

// Ensure interface compliance.
var _ domain.PrerequisiteRepository = (*PostgresPrerequisiteRepo)(nil)
