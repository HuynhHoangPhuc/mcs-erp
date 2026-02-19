package domain

// Constraint evaluates a set of assignments and returns a violation count.
// Hard constraints must have zero violations for a feasible schedule.
// Soft constraints contribute to the penalty score for quality optimisation.
type Constraint interface {
	// Evaluate returns the number of violations found in the assignment list.
	Evaluate(assignments []Assignment) int
	// IsHard reports whether this is a hard (infeasibility) constraint.
	IsHard() bool
	// Name returns a human-readable identifier for logging/debugging.
	Name() string
}
