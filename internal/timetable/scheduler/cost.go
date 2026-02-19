package scheduler

import "github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"

// hardWeight multiplies each hard-constraint violation to dominate soft penalties.
const hardWeight = 10000

// Cost computes the total scheduling cost for a set of assignments.
// Hard violations are weighted by hardWeight so any infeasibility dominates.
// Soft penalties are summed unweighted for relative quality comparison.
func Cost(assignments []domain.Assignment, hard, soft []domain.Constraint) int {
	total := 0
	for _, c := range hard {
		total += c.Evaluate(assignments) * hardWeight
	}
	for _, c := range soft {
		total += c.Evaluate(assignments)
	}
	return total
}
