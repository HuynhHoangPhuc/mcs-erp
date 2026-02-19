package domain

import (
	"time"

	"github.com/google/uuid"
)

// Room represents a physical room (classroom, lab, etc.) within a tenant.
type Room struct {
	ID        uuid.UUID
	Name      string
	Code      string // unique identifier e.g. "A101"
	Building  string
	Floor     int
	Capacity  int
	Equipment []string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
