package spine

import (
	"errors"
	"math"
)

// MaxFlowResult holds the result of a max flow computation.
type MaxFlowResult struct {
	MaxFlow float64                       `json:"max_flow"`
	Flow    map[string]map[string]float64 `json:"flow"`
	MinCut  [][2]string                   `json:"min_cut"`
}

// MaxFlow computes maximum flow from source to sink using Edmonds-Karp (BFS-based Ford-Fulkerson).
// Edge weights are used as capacities. Returns error if source/sink missing or graph is undirected.
func MaxFlow[N, E any](g *Graph[N, E], source, sink string) (*MaxFlowResult, error) {
	if !g.Directed {
		return nil, errors.New("max flow requires a directed graph")
	}
	if !g.HasNode(source) {
		return nil, errors.New("source node not found")
	}
	if !g.HasNode(sink) {
		return nil, errors.New("sink node not found")
	}
	if source == sink {
		return nil, errors.New("source and sink must be different")
	}

	nodes := g.Nodes()
	nodeIDs := make([]string, len(nodes))
	for i, n := range nodes {
		nodeIDs[i] = n.ID
	}

	// Build capacity and flow maps
	capacity := make(map[string]map[string]float64)
	flow := make(map[string]map[string]float64)
	for _, id := range nodeIDs {
		capacity[id] = make(map[string]float64)
		flow[id] = make(map[string]float64)
	}
	for _, e := range g.Edges() {
		capacity[e.From][e.To] = e.Weight
	}

	// Build adjacency for residual graph (includes reverse edges)
	adj := make(map[string]map[string]bool)
	for _, id := range nodeIDs {
		adj[id] = make(map[string]bool)
	}
	for _, e := range g.Edges() {
		adj[e.From][e.To] = true
		adj[e.To][e.From] = true
	}

	totalFlow := 0.0

	for {
		// BFS to find augmenting path
		parent := make(map[string]string)
		parent[source] = source
		queue := []string{source}
		found := false

		for len(queue) > 0 && !found {
			u := queue[0]
			queue = queue[1:]
			for v := range adj[u] {
				if _, visited := parent[v]; !visited {
					residual := capacity[u][v] - flow[u][v]
					if residual > 0 {
						parent[v] = u
						if v == sink {
							found = true
							break
						}
						queue = append(queue, v)
					}
				}
			}
		}

		if !found {
			break
		}

		// Find bottleneck
		pathFlow := math.Inf(1)
		v := sink
		for v != source {
			u := parent[v]
			residual := capacity[u][v] - flow[u][v]
			if residual < pathFlow {
				pathFlow = residual
			}
			v = u
		}

		// Update flow along path
		v = sink
		for v != source {
			u := parent[v]
			flow[u][v] += pathFlow
			flow[v][u] -= pathFlow
			v = u
		}

		totalFlow += pathFlow
	}

	// Find min-cut: BFS from source in residual graph
	reachable := make(map[string]bool)
	queue := []string{source}
	reachable[source] = true
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for v := range adj[u] {
			if !reachable[v] && (capacity[u][v]-flow[u][v]) > 0 {
				reachable[v] = true
				queue = append(queue, v)
			}
		}
	}

	var minCut [][2]string
	for _, e := range g.Edges() {
		if reachable[e.From] && !reachable[e.To] {
			minCut = append(minCut, [2]string{e.From, e.To})
		}
	}

	// Build positive-only flow map for output
	posFlow := make(map[string]map[string]float64)
	for u, m := range flow {
		for v, f := range m {
			if f > 0 {
				if posFlow[u] == nil {
					posFlow[u] = make(map[string]float64)
				}
				posFlow[u][v] = f
			}
		}
	}

	return &MaxFlowResult{
		MaxFlow: totalFlow,
		Flow:    posFlow,
		MinCut:  minCut,
	}, nil
}
