package scheduler

import (
	"math"
	"math/rand"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
)

// Anneal runs simulated annealing starting from the given assignments.
// It uses the standard SA acceptance criterion:
//   - always accept improvements (delta < 0)
//   - accept degradations with probability e^(-delta/T)
//
// Returns the best assignment set found over the entire run.
func Anneal(assignments []domain.Assignment, p Problem, cfg SAConfig, rng *rand.Rand) []domain.Assignment {
	hard := BuildHardConstraints(p)
	soft := BuildSoftConstraints()

	current := make([]domain.Assignment, len(assignments))
	copy(current, assignments)

	currentCost := Cost(current, hard, soft)
	best := make([]domain.Assignment, len(current))
	copy(best, current)
	bestCost := currentCost

	T := cfg.TInitial

	for iter := 0; iter < cfg.MaxIter && T > cfg.TMin; iter++ {
		candidate := Neighbor(current, p, rng)
		candidateCost := Cost(candidate, hard, soft)

		delta := candidateCost - currentCost

		// Accept candidate if better, or probabilistically if worse.
		if delta < 0 || rng.Float64() < math.Exp(-float64(delta)/T) {
			current = candidate
			currentCost = candidateCost

			if currentCost < bestCost {
				best = make([]domain.Assignment, len(current))
				copy(best, current)
				bestCost = currentCost
			}
		}

		T *= cfg.CoolingRate
	}

	return best
}
