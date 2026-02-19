package domain

import (
	"time"

	"github.com/google/uuid"
)

// RoomCreated is published when a new room is created.
type RoomCreated struct {
	RoomID    uuid.UUID `json:"room_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Capacity  int       `json:"capacity"`
	OccurredAt time.Time `json:"occurred_at"`
}

// RoomUpdated is published when room details change.
type RoomUpdated struct {
	RoomID    uuid.UUID `json:"room_id"`
	Code      string    `json:"code"`
	OccurredAt time.Time `json:"occurred_at"`
}

// RoomAvailabilityUpdated is published when a room's availability slots change.
type RoomAvailabilityUpdated struct {
	RoomID    uuid.UUID `json:"room_id"`
	OccurredAt time.Time `json:"occurred_at"`
}
