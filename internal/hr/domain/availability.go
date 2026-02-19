package domain

import "github.com/google/uuid"

// WeeklySlot represents a specific teaching period on a day of the week.
// Day: 0=Monday ... 6=Sunday. Period: 1-10 (fixed school periods per day).
type WeeklySlot struct {
	Day    int // 0-6
	Period int // 1-10
}

// TeacherAvailability maps each weekly slot to whether the teacher is available.
type TeacherAvailability map[WeeklySlot]bool

// Availability is the flat row stored in the database for a single slot.
type Availability struct {
	TeacherID   uuid.UUID
	Day         int
	Period      int
	IsAvailable bool
}
