package room

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/application/services"
	coredel "github.com/HuynhHoangPhuc/mcs-erp/internal/core/delivery"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/room/delivery"
	roomdomain "github.com/HuynhHoangPhuc/mcs-erp/internal/room/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/room/infrastructure"
)

// Module implements pkg/module.Module for the room module.
type Module struct {
	pool      *pgxpool.Pool
	authSvc   *services.AuthService
	roomRepo  *infrastructure.PostgresRoomRepo
	availRepo *infrastructure.PostgresAvailabilityRepo
}

// NewModule creates the room module wired with concrete dependencies.
func NewModule(pool *pgxpool.Pool, authSvc *services.AuthService) *Module {
	return &Module{
		pool:      pool,
		authSvc:   authSvc,
		roomRepo:  infrastructure.NewPostgresRoomRepo(pool),
		availRepo: infrastructure.NewPostgresAvailabilityRepo(pool),
	}
}

// RoomRepo returns the room repository for cross-module access.
func (m *Module) RoomRepo() roomdomain.RoomRepository { return m.roomRepo }

// RoomAvailabilityRepo returns the room availability repository for cross-module access.
func (m *Module) RoomAvailabilityRepo() roomdomain.RoomAvailabilityRepository { return m.availRepo }

func (m *Module) Name() string          { return "room" }
func (m *Module) Dependencies() []string { return []string{"core"} }

// Migrate runs room table migrations across all active tenant schemas.
func (m *Module) Migrate(ctx context.Context) error {
	migrator := database.NewMigrator(m.pool)

	if err := migrator.MigrateAll(ctx, sqlCreateRoomsTable); err != nil {
		return err
	}
	return migrator.MigrateAll(ctx, sqlCreateRoomAvailabilityTable)
}

func (m *Module) RegisterEvents(_ context.Context) error { return nil }

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	roomHandler := delivery.NewRoomHandler(m.roomRepo)
	availHandler := delivery.NewAvailabilityHandler(m.roomRepo, m.availRepo)

	authMw := coredel.AuthMiddleware(m.authSvc)
	readPerm := auth.RequirePermission(domain.PermRoomRead)
	writePerm := auth.RequirePermission(domain.PermRoomWrite)

	mux.Handle("POST /api/v1/rooms", authMw(writePerm(http.HandlerFunc(roomHandler.CreateRoom))))
	mux.Handle("GET /api/v1/rooms", authMw(readPerm(http.HandlerFunc(roomHandler.ListRooms))))
	mux.Handle("GET /api/v1/rooms/{id}", authMw(readPerm(http.HandlerFunc(roomHandler.GetRoom))))
	mux.Handle("PUT /api/v1/rooms/{id}", authMw(writePerm(http.HandlerFunc(roomHandler.UpdateRoom))))

	mux.Handle("GET /api/v1/rooms/{id}/availability", authMw(readPerm(http.HandlerFunc(availHandler.GetAvailability))))
	mux.Handle("PUT /api/v1/rooms/{id}/availability", authMw(writePerm(http.HandlerFunc(availHandler.SetAvailability))))
}

// sqlCreateRoomsTable is the DDL for the rooms table.
const sqlCreateRoomsTable = `
CREATE TABLE IF NOT EXISTS rooms (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name      VARCHAR(255) NOT NULL,
    code      VARCHAR(50)  NOT NULL,
    building  VARCHAR(255) NOT NULL DEFAULT '',
    floor     INTEGER      NOT NULL DEFAULT 0,
    capacity  INTEGER      NOT NULL CHECK (capacity > 0),
    equipment TEXT[]       NOT NULL DEFAULT '{}',
    is_active BOOLEAN      NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT rooms_code_unique UNIQUE (code)
);
CREATE INDEX IF NOT EXISTS idx_rooms_building ON rooms(building);
CREATE INDEX IF NOT EXISTS idx_rooms_capacity ON rooms(capacity);
`

// sqlCreateRoomAvailabilityTable is the DDL for the room_availability table.
const sqlCreateRoomAvailabilityTable = `
CREATE TABLE IF NOT EXISTS room_availability (
    room_id      UUID    NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    day          SMALLINT NOT NULL CHECK (day >= 0 AND day <= 6),
    period       SMALLINT NOT NULL CHECK (period >= 1 AND period <= 10),
    is_available BOOLEAN  NOT NULL DEFAULT true,
    CONSTRAINT room_availability_unique UNIQUE (room_id, day, period)
);
CREATE INDEX IF NOT EXISTS idx_room_availability_room_id ON room_availability(room_id);
`
