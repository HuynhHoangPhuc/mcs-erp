package scheduler

import (
	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
)

// --- Hard constraints ---

// TeacherConflictConstraint detects teacher double-booking: same teacher at the same day+period.
type TeacherConflictConstraint struct{}

func (c TeacherConflictConstraint) Name() string { return "teacher_conflict" }
func (c TeacherConflictConstraint) IsHard() bool { return true }

func (c TeacherConflictConstraint) Evaluate(assignments []domain.Assignment) int {
	type key struct {
		teacherID uuid.UUID
		day, period int
	}
	seen := make(map[key]int, len(assignments))
	for _, a := range assignments {
		k := key{a.TeacherID, a.Day, a.Period}
		seen[k]++
	}
	violations := 0
	for _, count := range seen {
		if count > 1 {
			violations += count - 1
		}
	}
	return violations
}

// RoomConflictConstraint detects room double-booking: same room at the same day+period.
type RoomConflictConstraint struct{}

func (c RoomConflictConstraint) Name() string { return "room_conflict" }
func (c RoomConflictConstraint) IsHard() bool { return true }

func (c RoomConflictConstraint) Evaluate(assignments []domain.Assignment) int {
	type key struct {
		roomID      uuid.UUID
		day, period int
	}
	seen := make(map[key]int, len(assignments))
	for _, a := range assignments {
		k := key{a.RoomID, a.Day, a.Period}
		seen[k]++
	}
	violations := 0
	for _, count := range seen {
		if count > 1 {
			violations += count - 1
		}
	}
	return violations
}

// TeacherUnavailableConstraint penalises assignments where the teacher is not available.
type TeacherUnavailableConstraint struct {
	// TeacherAvail maps teacherID -> set of available TimeSlots.
	TeacherAvail map[uuid.UUID]map[domain.TimeSlot]bool
}

func (c TeacherUnavailableConstraint) Name() string { return "teacher_unavailable" }
func (c TeacherUnavailableConstraint) IsHard() bool { return true }

func (c TeacherUnavailableConstraint) Evaluate(assignments []domain.Assignment) int {
	violations := 0
	for _, a := range assignments {
		avail, ok := c.TeacherAvail[a.TeacherID]
		if !ok {
			// No availability data → treat all slots as available (permissive default).
			continue
		}
		slot := domain.TimeSlot{Day: a.Day, Period: a.Period}
		if !avail[slot] {
			violations++
		}
	}
	return violations
}

// RoomUnavailableConstraint penalises assignments where the room is not available.
type RoomUnavailableConstraint struct {
	// RoomAvail maps roomID -> set of available TimeSlots.
	RoomAvail map[uuid.UUID]map[domain.TimeSlot]bool
}

func (c RoomUnavailableConstraint) Name() string { return "room_unavailable" }
func (c RoomUnavailableConstraint) IsHard() bool { return true }

func (c RoomUnavailableConstraint) Evaluate(assignments []domain.Assignment) int {
	violations := 0
	for _, a := range assignments {
		avail, ok := c.RoomAvail[a.RoomID]
		if !ok {
			continue
		}
		slot := domain.TimeSlot{Day: a.Day, Period: a.Period}
		if !avail[slot] {
			violations++
		}
	}
	return violations
}

// --- Soft constraints ---

// TeacherGapConstraint counts scheduling gaps (idle periods between classes) per teacher per day.
// Fewer gaps → more compact teaching days → lower penalty.
type TeacherGapConstraint struct{}

func (c TeacherGapConstraint) Name() string { return "teacher_gap" }
func (c TeacherGapConstraint) IsHard() bool { return false }

func (c TeacherGapConstraint) Evaluate(assignments []domain.Assignment) int {
	// Group periods by (teacherID, day).
	type dayKey struct {
		teacherID uuid.UUID
		day       int
	}
	periodsMap := make(map[dayKey][]int)
	for _, a := range assignments {
		k := dayKey{a.TeacherID, a.Day}
		periodsMap[k] = append(periodsMap[k], a.Period)
	}

	gaps := 0
	for _, periods := range periodsMap {
		if len(periods) < 2 {
			continue
		}
		// Sort periods (simple insertion sort for small slices).
		for i := 1; i < len(periods); i++ {
			for j := i; j > 0 && periods[j] < periods[j-1]; j-- {
				periods[j], periods[j-1] = periods[j-1], periods[j]
			}
		}
		// Count gaps between consecutive periods.
		for i := 1; i < len(periods); i++ {
			gap := periods[i] - periods[i-1] - 1
			if gap > 0 {
				gaps += gap
			}
		}
	}
	return gaps
}

// EvenDistributionConstraint penalises uneven spread of assignments across weekdays.
// It computes the variance of the per-day assignment counts (scaled to int).
type EvenDistributionConstraint struct{}

func (c EvenDistributionConstraint) Name() string { return "even_distribution" }
func (c EvenDistributionConstraint) IsHard() bool { return false }

func (c EvenDistributionConstraint) Evaluate(assignments []domain.Assignment) int {
	const numDays = 6 // Mon-Sat
	counts := make([]int, numDays)
	for _, a := range assignments {
		if a.Day >= 0 && a.Day < numDays {
			counts[a.Day]++
		}
	}

	// Compute mean.
	total := 0
	for _, c := range counts {
		total += c
	}
	mean := float64(total) / float64(numDays)

	// Sum squared deviations (scaled by 10 to keep int resolution).
	variance := 0.0
	for _, c := range counts {
		diff := float64(c) - mean
		variance += diff * diff
	}
	return int(variance * 10)
}

// BuildHardConstraints assembles the standard hard-constraint set for a problem.
func BuildHardConstraints(p Problem) []domain.Constraint {
	teacherAvail := make(map[uuid.UUID]map[domain.TimeSlot]bool, len(p.Teachers))
	for _, t := range p.Teachers {
		teacherAvail[t.ID] = t.Available
	}
	roomAvail := make(map[uuid.UUID]map[domain.TimeSlot]bool, len(p.Rooms))
	for _, r := range p.Rooms {
		roomAvail[r.ID] = r.Available
	}
	return []domain.Constraint{
		TeacherConflictConstraint{},
		RoomConflictConstraint{},
		TeacherUnavailableConstraint{TeacherAvail: teacherAvail},
		RoomUnavailableConstraint{RoomAvail: roomAvail},
	}
}

// BuildSoftConstraints assembles the standard soft-constraint set.
func BuildSoftConstraints() []domain.Constraint {
	return []domain.Constraint{
		TeacherGapConstraint{},
		EvenDistributionConstraint{},
	}
}
