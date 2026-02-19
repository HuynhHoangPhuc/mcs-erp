package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/room/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// PostgresRoomRepo implements domain.RoomRepository using pgx.
type PostgresRoomRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresRoomRepo creates a new room repository.
func NewPostgresRoomRepo(pool *pgxpool.Pool) *PostgresRoomRepo {
	return &PostgresRoomRepo{pool: pool}
}

func (r *PostgresRoomRepo) schema(ctx context.Context) (string, error) {
	return tenant.FromContext(ctx)
}

func (r *PostgresRoomRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var room domain.Room
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT id, name, code, building, floor, capacity, equipment, is_active, created_at, updated_at
			 FROM rooms WHERE id = $1`,
			id,
		).Scan(
			&room.ID, &room.Name, &room.Code, &room.Building,
			&room.Floor, &room.Capacity, &room.Equipment,
			&room.IsActive, &room.CreatedAt, &room.UpdatedAt,
		)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, erptypes.ErrNotFound
		}
		return nil, fmt.Errorf("find room by id: %w", err)
	}
	return &room, nil
}

func (r *PostgresRoomRepo) FindByCode(ctx context.Context, code string) (*domain.Room, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var room domain.Room
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT id, name, code, building, floor, capacity, equipment, is_active, created_at, updated_at
			 FROM rooms WHERE code = $1`,
			code,
		).Scan(
			&room.ID, &room.Name, &room.Code, &room.Building,
			&room.Floor, &room.Capacity, &room.Equipment,
			&room.IsActive, &room.CreatedAt, &room.UpdatedAt,
		)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, erptypes.ErrNotFound
		}
		return nil, fmt.Errorf("find room by code: %w", err)
	}
	return &room, nil
}

func (r *PostgresRoomRepo) Save(ctx context.Context, room *domain.Room) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`INSERT INTO rooms (id, name, code, building, floor, capacity, equipment, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			room.ID, room.Name, room.Code, room.Building, room.Floor,
			room.Capacity, room.Equipment, room.IsActive, room.CreatedAt, room.UpdatedAt,
		)
		return err
	})
}

func (r *PostgresRoomRepo) Update(ctx context.Context, room *domain.Room) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE rooms
			 SET name = $2, code = $3, building = $4, floor = $5,
			     capacity = $6, equipment = $7, is_active = $8, updated_at = now()
			 WHERE id = $1`,
			room.ID, room.Name, room.Code, room.Building, room.Floor,
			room.Capacity, room.Equipment, room.IsActive,
		)
		return err
	})
}

func (r *PostgresRoomRepo) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Room, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var rooms []*domain.Room
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		query, args := buildListQuery(filter)
		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var room domain.Room
			if err := rows.Scan(
				&room.ID, &room.Name, &room.Code, &room.Building,
				&room.Floor, &room.Capacity, &room.Equipment,
				&room.IsActive, &room.CreatedAt, &room.UpdatedAt,
			); err != nil {
				return err
			}
			rooms = append(rooms, &room)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	return rooms, nil
}

// buildListQuery constructs a parameterized SELECT with optional WHERE clauses.
func buildListQuery(filter domain.ListFilter) (string, []any) {
	base := `SELECT id, name, code, building, floor, capacity, equipment, is_active, created_at, updated_at
	         FROM rooms`

	var clauses []string
	var args []any
	idx := 1

	if filter.Building != "" {
		clauses = append(clauses, fmt.Sprintf("building = $%d", idx))
		args = append(args, filter.Building)
		idx++
	}

	if filter.MinCapacity > 0 {
		clauses = append(clauses, fmt.Sprintf("capacity >= $%d", idx))
		args = append(args, filter.MinCapacity)
		idx++
	}

	// equipment filter: room must contain ALL requested items
	for _, eq := range filter.Equipment {
		clauses = append(clauses, fmt.Sprintf("$%d = ANY(equipment)", idx))
		args = append(args, eq)
		idx++
	}

	query := base
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY code ASC"

	return query, args
}

// Ensure interface compliance.
var _ domain.RoomRepository = (*PostgresRoomRepo)(nil)
