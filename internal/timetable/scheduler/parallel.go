package scheduler

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
)

// workerResult carries a single worker's best schedule and its cost.
type workerResult struct {
	assignments []domain.Assignment
	cost        int
}

// ParallelAnneal spawns numWorkers goroutines, each seeding its own RNG,
// running GreedyAssign + Anneal independently, and returns the best result.
// If numWorkers <= 0, it defaults to runtime.NumCPU().
func ParallelAnneal(p Problem, cfg SAConfig, numWorkers int) []domain.Assignment {
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	hard := BuildHardConstraints(p)
	soft := BuildSoftConstraints()

	results := make(chan workerResult, numWorkers)
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		seed := int64(w*1_000_003 + 42) // deterministic but distinct per worker
		go func(seed int64) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(seed))
			initial := GreedyAssign(p)
			best := Anneal(initial, p, cfg, rng)
			results <- workerResult{
				assignments: best,
				cost:        Cost(best, hard, soft),
			}
		}(seed)
	}

	// Close channel once all workers finish.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and pick the lowest-cost result.
	var best []domain.Assignment
	bestCost := int(^uint(0) >> 1) // MaxInt

	for r := range results {
		if r.cost < bestCost {
			bestCost = r.cost
			best = r.assignments
		}
	}

	return best
}
