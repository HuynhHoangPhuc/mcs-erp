package domain

import (
	"time"

	"github.com/google/uuid"
)

// Department represents an academic or administrative department within a tenant.
type Department struct {
	ID            uuid.UUID
	Name          string
	Description   string
	HeadTeacherID *uuid.UUID
	CreatedAt     time.Time
}
