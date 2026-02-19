package erptypes

import "github.com/google/uuid"

// ID is the standard identifier type used across all modules.
type ID = uuid.UUID

// NewID generates a new random UUID.
func NewID() ID {
	return uuid.New()
}

// ParseID parses a string into an ID. Returns error if invalid.
func ParseID(s string) (ID, error) {
	return uuid.Parse(s)
}

// NilID returns the zero-value UUID.
func NilID() ID {
	return uuid.Nil
}
