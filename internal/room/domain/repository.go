package domain

import (
	"context"

	"github.com/google/uuid"
)

// ListFilter holds optional filters for listing rooms.
type ListFilter struct {
	Building    string   // filter by building name (empty = no filter)
	MinCapacity int      // minimum seat capacity (0 = no filter)
	Equipment   []string // must have all listed equipment (nil = no filter)
}

// RoomRepository defines persistence operations for Room entities.
type RoomRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Room, error)
	FindByCode(ctx context.Context, code string) (*Room, error)
	Save(ctx context.Context, room *Room) error
	Update(ctx context.Context, room *Room) error
	List(ctx context.Context, filter ListFilter) ([]*Room, error)
}

// RoomAvailabilityRepository persists weekly slot availability for rooms.
type RoomAvailabilityRepository interface {
	// GetByRoomID returns the full availability map for a room.
	// Returns an empty map (not error) when no slots are configured.
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (RoomAvailability, error)

	// SetSlots replaces all availability records for the given room.
	SetSlots(ctx context.Context, roomID uuid.UUID, avail RoomAvailability) error
}
