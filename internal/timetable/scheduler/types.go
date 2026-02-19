// Package scheduler provides a pure-Go timetable scheduling engine.
// It has no database dependencies; all inputs are passed as in-memory structs.
package scheduler

import (
	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
)

// SubjectInfo holds scheduling-relevant data for a subject.
type SubjectInfo struct {
	ID           uuid.UUID
	HoursPerWeek int
}

// TeacherInfo holds scheduling-relevant data for a teacher.
type TeacherInfo struct {
	ID             uuid.UUID
	Available      map[domain.TimeSlot]bool // weekly availability grid
	Qualifications []string
}

// RoomInfo holds scheduling-relevant data for a room.
type RoomInfo struct {
	ID        uuid.UUID
	Capacity  int
	Equipment []string
	Available map[domain.TimeSlot]bool // weekly availability grid
}

// Problem is the complete scheduling input passed to the engine.
type Problem struct {
	Subjects      []SubjectInfo
	Teachers      []TeacherInfo
	Rooms         []RoomInfo
	TeacherAssign map[uuid.UUID]uuid.UUID // subjectID -> pre-assigned teacherID
	Slots         []domain.TimeSlot
}

// SAConfig holds simulated annealing tuning parameters.
type SAConfig struct {
	TInitial    float64 // starting temperature, e.g. 1000.0
	CoolingRate float64 // multiplicative factor per iteration, e.g. 0.9995
	TMin        float64 // stop when temperature falls below this, e.g. 0.01
	MaxIter     int     // hard cap on iterations, e.g. 500000
}

// DefaultSAConfig returns sensible defaults for the SA algorithm.
func DefaultSAConfig() SAConfig {
	return SAConfig{
		TInitial:    1000.0,
		CoolingRate: 0.9995,
		TMin:        0.01,
		MaxIter:     500000,
	}
}
