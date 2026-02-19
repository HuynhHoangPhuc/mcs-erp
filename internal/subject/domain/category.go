package domain

import (
	"time"

	"github.com/google/uuid"
)

// Category groups subjects into logical curriculum categories.
type Category struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
}
