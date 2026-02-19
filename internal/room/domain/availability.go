package domain

// WeeklySlot identifies a specific time slot in the weekly schedule.
// Day: 0=Monday ... 6=Sunday. Period: 1-10.
type WeeklySlot struct {
	Day    int // 0-6
	Period int // 1-10
}

// RoomAvailability maps each weekly slot to whether the room is available.
type RoomAvailability = map[WeeklySlot]bool
