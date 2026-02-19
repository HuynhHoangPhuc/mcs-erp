package infrastructure

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/room/domain"
)

// PostgresAvailabilityRepo implements domain.RoomAvailabilityRepository using pgx.
type PostgresAvailabilityRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresAvailabilityRepo creates a new availability repository.
func NewPostgresAvailabilityRepo(pool *pgxpool.Pool) *PostgresAvailabilityRepo {
	return &PostgresAvailabilityRepo{pool: pool}
}

func (r *PostgresAvailabilityRepo) schema(ctx context.Context) (string, error) {
	return tenant.FromContext(ctx)
}

func (r *PostgresAvailabilityRepo) GetByRoomID(ctx context.Context, roomID uuid.UUID) (domain.RoomAvailability, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	avail := make(domain.RoomAvailability)
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT day, period, is_available FROM room_availability WHERE room_id = $1`,
			roomID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var slot domain.WeeklySlot
			var isAvailable bool
			if err := rows.Scan(&slot.Day, &slot.Period, &isAvailable); err != nil {
				return err
			}
			avail[slot] = isAvailable
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("get room availability: %w", err)
	}
	return avail, nil
}

func (r *PostgresAvailabilityRepo) SetSlots(ctx context.Context, roomID uuid.UUID, avail domain.RoomAvailability) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		// Replace all slots for this room in a single transaction
		if _, err := tx.Exec(ctx,
			`DELETE FROM room_availability WHERE room_id = $1`, roomID,
		); err != nil {
			return err
		}

		for slot, isAvailable := range avail {
			if _, err := tx.Exec(ctx,
				`INSERT INTO room_availability (room_id, day, period, is_available)
				 VALUES ($1, $2, $3, $4)`,
				roomID, slot.Day, slot.Period, isAvailable,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

// Ensure interface compliance.
var _ domain.RoomAvailabilityRepository = (*PostgresAvailabilityRepo)(nil)
