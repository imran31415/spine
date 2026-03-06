package spine

import "math"

// CentralityResult holds centrality scores for each node.
type CentralityResult struct {
	Scores map[string]float64 `json:"scores"`
}

// PageRankResult holds PageRank scores with convergence info.
type PageRankResult struct {
	Scores     map[string]float64 `json:"scores"`
	Iterations int                `json:"iterations"`
	Converged  bool               `json:"converged"`
}

// DegreeCentrality computes degree centrality for each node.
// For directed graphs: out-degree / (n-1). For undirected: degree / (n-1).
func DegreeCentrality[N, E any](g *Graph[N, E]) CentralityResult {
	nodes := g.Nodes()
	n := len(nodes)
	scores := make(map[string]float64, n)
	if n <= 1 {
		for _, nd := range nodes {
			scores[nd.ID] = 0
		}
		return CentralityResult{Scores: scores}
	}
	denom := float64(n - 1)
	for _, nd := range nodes {
		if g.Directed {
			scores[nd.ID] = float64(len(g.OutEdges(nd.ID))) / denom
		} else {
			scores[nd.ID] = float64(len(g.Neighbors(nd.ID))) / denom
		}
	}
	return CentralityResult{Scores: scores}
}

// BetweennessCentrality computes betweenness centrality using Brandes' algorithm.
// O(V*E) time complexity.
func BetweennessCentrality[N, E any](g *Graph[N, E]) CentralityResult {
	nodes := g.Nodes()
	cb := make(map[string]float64, len(nodes))
	for _, nd := range nodes {
		cb[nd.ID] = 0
	}

	for _, s := range nodes {
		// BFS from s
		stack := make([]string, 0)
		pred := make(map[string][]string)
		sigma := make(map[string]float64)
		dist := make(map[string]int)
		for _, nd := range nodes {
			sigma[nd.ID] = 0
			dist[nd.ID] = -1
		}
		sigma[s.ID] = 1
		dist[s.ID] = 0
		queue := []string{s.ID}

		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			stack = append(stack, v)
			for _, nb := range g.Neighbors(v) {
				if dist[nb] < 0 {
					queue = append(queue, nb)
					dist[nb] = dist[v] + 1
				}
				if dist[nb] == dist[v]+1 {
					sigma[nb] += sigma[v]
					pred[nb] = append(pred[nb], v)
				}
			}
		}

		delta := make(map[string]float64)
		for i := len(stack) - 1; i >= 0; i-- {
			w := stack[i]
			for _, v := range pred[w] {
				delta[v] += (sigma[v] / sigma[w]) * (1 + delta[w])
			}
			if w != s.ID {
				cb[w] += delta[w]
			}
		}
	}

	// For undirected graphs, each pair is counted twice
	if !g.Directed {
		for id := range cb {
			cb[id] /= 2
		}
	}

	return CentralityResult{Scores: cb}
}

// ClosenessCentrality computes closeness centrality for each node.
// closeness(v) = (reachable-1) / sum_of_distances for reachable nodes.
func ClosenessCentrality[N, E any](g *Graph[N, E]) CentralityResult {
	nodes := g.Nodes()
	scores := make(map[string]float64, len(nodes))

	for _, s := range nodes {
		// BFS to compute distances
		dist := map[string]int{s.ID: 0}
		queue := []string{s.ID}
		totalDist := 0
		reachable := 0

		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			for _, nb := range g.Neighbors(v) {
				if _, seen := dist[nb]; !seen {
					dist[nb] = dist[v] + 1
					totalDist += dist[nb]
					reachable++
					queue = append(queue, nb)
				}
			}
		}

		if reachable > 0 && totalDist > 0 {
			scores[s.ID] = float64(reachable) / float64(totalDist)
		} else {
			scores[s.ID] = 0
		}
	}

	return CentralityResult{Scores: scores}
}

// PageRank computes PageRank scores using power iteration.
// damping is typically 0.85. Converges when max score change < tol.
func PageRank[N, E any](g *Graph[N, E], damping float64, maxIter int, tol float64) PageRankResult {
	nodes := g.Nodes()
	n := len(nodes)
	if n == 0 {
		return PageRankResult{Scores: map[string]float64{}, Iterations: 0, Converged: true}
	}

	scores := make(map[string]float64, n)
	initial := 1.0 / float64(n)
	for _, nd := range nodes {
		scores[nd.ID] = initial
	}

	// Precompute out-degree
	outDeg := make(map[string]int, n)
	for _, nd := range nodes {
		outDeg[nd.ID] = len(g.OutEdges(nd.ID))
	}

	converged := false
	iter := 0
	for iter = 0; iter < maxIter; iter++ {
		newScores := make(map[string]float64, n)

		// Collect dangling node rank
		danglingSum := 0.0
		for _, nd := range nodes {
			if outDeg[nd.ID] == 0 {
				danglingSum += scores[nd.ID]
			}
		}

		for _, nd := range nodes {
			sum := 0.0
			for _, e := range g.InEdges(nd.ID) {
				sum += scores[e.From] / float64(outDeg[e.From])
			}
			newScores[nd.ID] = (1-damping)/float64(n) + damping*(sum+danglingSum/float64(n))
		}

		// Check convergence
		maxDiff := 0.0
		for _, nd := range nodes {
			diff := math.Abs(newScores[nd.ID] - scores[nd.ID])
			if diff > maxDiff {
				maxDiff = diff
			}
		}
		scores = newScores
		if maxDiff < tol {
			converged = true
			iter++
			break
		}
	}

	return PageRankResult{Scores: scores, Iterations: iter, Converged: converged}
}
