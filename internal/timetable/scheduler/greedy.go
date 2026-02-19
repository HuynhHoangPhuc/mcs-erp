package scheduler

import (
	"sort"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
)

// GreedyAssign produces an initial assignment set using a greedy heuristic.
// Subjects are ordered by HoursPerWeek descending (most-constrained first).
// For each required hour, it picks the first valid (slot, room) pair that
// satisfies teacher and room availability with no conflicts.
// Slots that cannot be placed are skipped — SA will attempt to fix them.
func GreedyAssign(p Problem) []domain.Assignment {
	// Sort subjects: most hours first (most constrained).
	subjects := make([]SubjectInfo, len(p.Subjects))
	copy(subjects, p.Subjects)
	sort.Slice(subjects, func(i, j int) bool {
		return subjects[i].HoursPerWeek > subjects[j].HoursPerWeek
	})

	// Track usage: teacher slot occupancy and room slot occupancy.
	teacherBusy := make(map[uuid.UUID]map[domain.TimeSlot]bool)
	roomBusy := make(map[uuid.UUID]map[domain.TimeSlot]bool)

	for _, t := range p.Teachers {
		teacherBusy[t.ID] = make(map[domain.TimeSlot]bool)
	}
	for _, r := range p.Rooms {
		roomBusy[r.ID] = make(map[domain.TimeSlot]bool)
	}

	// Build fast lookup maps.
	teacherAvail := make(map[uuid.UUID]map[domain.TimeSlot]bool, len(p.Teachers))
	for _, t := range p.Teachers {
		teacherAvail[t.ID] = t.Available
	}
	roomAvail := make(map[uuid.UUID]map[domain.TimeSlot]bool, len(p.Rooms))
	for _, r := range p.Rooms {
		roomAvail[r.ID] = r.Available
	}

	var assignments []domain.Assignment

	for _, subj := range subjects {
		// Resolve pre-assigned teacher or skip if none available.
		teacherID, ok := resolveTeacher(subj.ID, p)
		if !ok {
			continue
		}

		for hour := 0; hour < subj.HoursPerWeek; hour++ {
			placed := false
			for _, slot := range p.Slots {
				// Teacher must be available and free.
				if avail := teacherAvail[teacherID]; len(avail) > 0 && !avail[slot] {
					continue
				}
				if teacherBusy[teacherID][slot] {
					continue
				}

				// Find a free, available room.
				roomID, found := pickRoom(slot, p.Rooms, roomAvail, roomBusy)
				if !found {
					continue
				}

				// Place assignment.
				a := domain.Assignment{
					ID:        uuid.New(),
					SubjectID: subj.ID,
					TeacherID: teacherID,
					RoomID:    roomID,
					Day:       slot.Day,
					Period:    slot.Period,
				}
				assignments = append(assignments, a)
				teacherBusy[teacherID][slot] = true
				roomBusy[roomID][slot] = true
				placed = true
				break
			}
			// If not placed, skip — SA will attempt corrections.
			_ = placed
		}
	}

	return assignments
}

// resolveTeacher returns the teacher for a subject, either from the pre-assignment
// map or by picking the first available teacher.
func resolveTeacher(subjectID uuid.UUID, p Problem) (uuid.UUID, bool) {
	if tid, ok := p.TeacherAssign[subjectID]; ok {
		return tid, true
	}
	if len(p.Teachers) == 0 {
		return uuid.UUID{}, false
	}
	return p.Teachers[0].ID, true
}

// pickRoom returns the first room that is available and not busy at the given slot.
func pickRoom(
	slot domain.TimeSlot,
	rooms []RoomInfo,
	roomAvail map[uuid.UUID]map[domain.TimeSlot]bool,
	roomBusy map[uuid.UUID]map[domain.TimeSlot]bool,
) (uuid.UUID, bool) {
	for _, r := range rooms {
		if avail := roomAvail[r.ID]; len(avail) > 0 && !avail[slot] {
			continue
		}
		if roomBusy[r.ID][slot] {
			continue
		}
		return r.ID, true
	}
	return uuid.UUID{}, false
}
