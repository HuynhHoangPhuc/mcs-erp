package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// nodeColor represents DFS traversal state for cycle detection.
type nodeColor int

const (
	white nodeColor = iota // not yet visited
	gray                   // currently in the recursion stack
	black                  // fully processed
)

// Graph is an in-memory directed acyclic graph of subject prerequisites.
// Each edge A → B means "A requires B as a prerequisite".
// This struct has NO database dependency — all operations are pure Go.
type Graph struct {
	adjacency map[uuid.UUID][]uuid.UUID
}

// NewGraph builds a Graph from the given prerequisite edges.
func NewGraph(edges []PrerequisiteEdge) *Graph {
	adj := make(map[uuid.UUID][]uuid.UUID, len(edges))
	for _, e := range edges {
		adj[e.SubjectID] = append(adj[e.SubjectID], e.PrerequisiteID)
		// Ensure the prerequisite node exists in the map even if it has no outgoing edges.
		if _, ok := adj[e.PrerequisiteID]; !ok {
			adj[e.PrerequisiteID] = nil
		}
	}
	return &Graph{adjacency: adj}
}

// HasCycle reports whether adding the directed edge (from → to) to the graph
// would introduce a cycle. Uses DFS with white/gray/black coloring.
func HasCycle(g *Graph, newFrom, newTo uuid.UUID) bool {
	// Build a temporary adjacency map that includes the candidate edge.
	adj := make(map[uuid.UUID][]uuid.UUID, len(g.adjacency)+2)
	for k, v := range g.adjacency {
		adj[k] = v
	}
	adj[newFrom] = append(append([]uuid.UUID{}, adj[newFrom]...), newTo)
	if _, ok := adj[newTo]; !ok {
		adj[newTo] = nil
	}

	colors := make(map[uuid.UUID]nodeColor, len(adj))

	var dfs func(uuid.UUID) bool
	dfs = func(n uuid.UUID) bool {
		colors[n] = gray
		for _, neighbor := range adj[n] {
			switch colors[neighbor] {
			case gray:
				return true // back-edge → cycle
			case white:
				if dfs(neighbor) {
					return true
				}
			}
		}
		colors[n] = black
		return false
	}

	for node := range adj {
		if colors[node] == white {
			if dfs(node) {
				return true
			}
		}
	}
	return false
}

// TopologicalSort returns a valid topological ordering of the graph nodes using
// Kahn's algorithm. Returns an error if the graph contains a cycle.
func TopologicalSort(g *Graph) ([]uuid.UUID, error) {
	inDegree := make(map[uuid.UUID]int, len(g.adjacency))
	for node := range g.adjacency {
		if _, ok := inDegree[node]; !ok {
			inDegree[node] = 0
		}
		for _, neighbor := range g.adjacency[node] {
			inDegree[neighbor]++
		}
	}

	queue := make([]uuid.UUID, 0)
	for node, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, node)
		}
	}

	sorted := make([]uuid.UUID, 0, len(inDegree))
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		sorted = append(sorted, cur)
		for _, neighbor := range g.adjacency[cur] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(sorted) != len(inDegree) {
		return nil, fmt.Errorf("cycle detected: topological sort failed")
	}
	return sorted, nil
}

// GetPrerequisiteChain returns all transitive prerequisites of the given subjectID
// (i.e., every node reachable from subjectID in the prerequisite graph).
func GetPrerequisiteChain(g *Graph, subjectID uuid.UUID) []uuid.UUID {
	visited := make(map[uuid.UUID]bool)
	result := make([]uuid.UUID, 0)

	var dfs func(uuid.UUID)
	dfs = func(n uuid.UUID) {
		for _, dep := range g.adjacency[n] {
			if !visited[dep] {
				visited[dep] = true
				result = append(result, dep)
				dfs(dep)
			}
		}
	}

	dfs(subjectID)
	return result
}
