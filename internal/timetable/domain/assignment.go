package domain

import "github.com/google/uuid"

// Assignment is a single scheduled class: subject+teacher+room at a day/period.
type Assignment struct {
	ID         uuid.UUID
	SemesterID uuid.UUID
	SubjectID  uuid.UUID
	TeacherID  uuid.UUID
	RoomID     uuid.UUID
	Day        int // 0-5 (Mon-Sat)
	Period     int // 1-10
	Version    int // schedule generation version
}

// Slot returns the TimeSlot for this assignment.
func (a Assignment) Slot() TimeSlot {
	return TimeSlot{Day: a.Day, Period: a.Period}
}
