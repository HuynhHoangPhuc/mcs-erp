package scheduler

import (
	"math/rand"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
)

// moveType enumerates the three neighborhood move types.
type moveType int

const (
	moveSwap         moveType = iota // swap two assignments' (day, period, room)
	moveMoveSlot                     // relocate one assignment to a random slot
	moveReassignRoom                 // assign a different room to one assignment
)

// Neighbor generates a copy of assignments with a single random move applied.
// Move selection is weighted: Swap 50%, MoveSlot 35%, ReassignRoom 15%.
// Returns the mutated copy; the original slice is not modified.
func Neighbor(assignments []domain.Assignment, p Problem, rng *rand.Rand) []domain.Assignment {
	if len(assignments) == 0 {
		return assignments
	}

	// Copy the slice so the caller retains the original.
	next := make([]domain.Assignment, len(assignments))
	copy(next, assignments)

	move := pickMoveType(rng)
	switch move {
	case moveSwap:
		applySwap(next, rng)
	case moveMoveSlot:
		applyMoveSlot(next, p, rng)
	case moveReassignRoom:
		applyReassignRoom(next, p, rng)
	}
	return next
}

// pickMoveType selects a move type using weighted probabilities.
func pickMoveType(rng *rand.Rand) moveType {
	r := rng.Intn(100)
	switch {
	case r < 50:
		return moveSwap
	case r < 85:
		return moveMoveSlot
	default:
		return moveReassignRoom
	}
}

// applySwap swaps the (Day, Period, RoomID) of two randomly chosen assignments.
func applySwap(a []domain.Assignment, rng *rand.Rand) {
	if len(a) < 2 {
		return
	}
	i := rng.Intn(len(a))
	j := rng.Intn(len(a))
	for j == i {
		j = rng.Intn(len(a))
	}
	a[i].Day, a[j].Day = a[j].Day, a[i].Day
	a[i].Period, a[j].Period = a[j].Period, a[i].Period
	a[i].RoomID, a[j].RoomID = a[j].RoomID, a[i].RoomID
}

// applyMoveSlot moves one randomly chosen assignment to a random slot.
func applyMoveSlot(a []domain.Assignment, p Problem, rng *rand.Rand) {
	if len(a) == 0 || len(p.Slots) == 0 {
		return
	}
	i := rng.Intn(len(a))
	slot := p.Slots[rng.Intn(len(p.Slots))]
	a[i].Day = slot.Day
	a[i].Period = slot.Period
}

// applyReassignRoom assigns a different random room to one assignment.
func applyReassignRoom(a []domain.Assignment, p Problem, rng *rand.Rand) {
	if len(a) == 0 || len(p.Rooms) == 0 {
		return
	}
	i := rng.Intn(len(a))
	room := p.Rooms[rng.Intn(len(p.Rooms))]
	a[i].RoomID = room.ID
}
