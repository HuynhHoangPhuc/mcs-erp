package domain

// TimeSlot represents a specific teaching period on a day of the week.
// Day: 0=Monday, 1=Tuesday, ..., 5=Saturday. Period: 1-10.
type TimeSlot struct {
	Day    int // 0-5 (Mon-Sat)
	Period int // 1-10
}

// AllSlots generates all valid TimeSlot combinations: days 0-5, periods 1-10 (60 total).
func AllSlots() []TimeSlot {
	slots := make([]TimeSlot, 0, 60)
	for day := 0; day <= 5; day++ {
		for period := 1; period <= 10; period++ {
			slots = append(slots, TimeSlot{Day: day, Period: period})
		}
	}
	return slots
}

// SameSlot reports whether two TimeSlots are equal.
func SameSlot(a, b TimeSlot) bool {
	return a.Day == b.Day && a.Period == b.Period
}
