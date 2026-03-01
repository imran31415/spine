package spine

import "sort"

// FilterNodes returns all nodes matching the predicate.
func FilterNodes[N, E any](g *Graph[N, E], pred func(Node[N]) bool) []Node[N] {
	var result []Node[N]
	for _, n := range g.Nodes() {
		if pred(n) {
			result = append(result, n)
		}
	}
	return result
}

// FilterEdges returns all edges matching the predicate.
func FilterEdges[N, E any](g *Graph[N, E], pred func(Edge[E]) bool) []Edge[E] {
	var result []Edge[E]
	for _, e := range g.Edges() {
		if pred(e) {
			result = append(result, e)
		}
	}
	return result
}

// Ancestors returns all transitive predecessors of the given node in a directed graph.
// For undirected graphs, this returns all reachable nodes.
func Ancestors[N, E any](g *Graph[N, E], id string) []string {
	visited := make(map[string]bool)
	var walk func(string)
	walk = func(cur string) {
		for _, e := range g.InEdges(cur) {
			src := e.From
			if !visited[src] {
				visited[src] = true
				walk(src)
			}
		}
	}
	walk(id)
	result := make([]string, 0, len(visited))
	for v := range visited {
		result = append(result, v)
	}
	return result
}

// Descendants returns all transitive successors of the given node in a directed graph.
func Descendants[N, E any](g *Graph[N, E], id string) []string {
	visited := make(map[string]bool)
	var walk func(string)
	walk = func(cur string) {
		for _, e := range g.OutEdges(cur) {
			dst := e.To
			if !visited[dst] {
				visited[dst] = true
				walk(dst)
			}
		}
	}
	walk(id)
	result := make([]string, 0, len(visited))
	for v := range visited {
		result = append(result, v)
	}
	return result
}

// Roots returns nodes with in-degree 0 (no incoming edges).
func Roots[N, E any](g *Graph[N, E]) []Node[N] {
	var result []Node[N]
	for _, n := range g.Nodes() {
		if len(g.InEdges(n.ID)) == 0 {
			result = append(result, n)
		}
	}
	return result
}

// Leaves returns nodes with out-degree 0 (no outgoing edges).
func Leaves[N, E any](g *Graph[N, E]) []Node[N] {
	var result []Node[N]
	for _, n := range g.Nodes() {
		if len(g.OutEdges(n.ID)) == 0 {
			result = append(result, n)
		}
	}
	return result
}

// Analytics holds graph-level statistics.
type Analytics struct {
	NodeCount    int            `json:"node_count"`
	EdgeCount    int            `json:"edge_count"`
	Directed     bool           `json:"directed"`
	Density      float64        `json:"density"`
	AvgDegree    float64        `json:"avg_degree"`
	MaxInDegree  int            `json:"max_in_degree"`
	MaxOutDegree int            `json:"max_out_degree"`
	InDegrees    map[string]int `json:"in_degrees"`
	OutDegrees   map[string]int `json:"out_degrees"`
	Diameter     int            `json:"diameter"`
	Components   int            `json:"components"`
}

// GraphAnalytics computes structural statistics for a graph including
// degree distributions, density, diameter, and component count.
func GraphAnalytics[N, E any](g *Graph[N, E]) Analytics {
	nodes := g.Nodes()
	n := len(nodes)
	edges := g.Edges()
	e := len(edges)

	a := Analytics{
		NodeCount:  n,
		EdgeCount:  e,
		Directed:   g.Directed,
		InDegrees:  make(map[string]int, n),
		OutDegrees: make(map[string]int, n),
		Diameter:   -1,
	}

	if n == 0 {
		a.Components = 0
		return a
	}

	// Compute degrees.
	for _, nd := range nodes {
		a.InDegrees[nd.ID] = len(g.InEdges(nd.ID))
		a.OutDegrees[nd.ID] = len(g.OutEdges(nd.ID))
		if a.InDegrees[nd.ID] > a.MaxInDegree {
			a.MaxInDegree = a.InDegrees[nd.ID]
		}
		if a.OutDegrees[nd.ID] > a.MaxOutDegree {
			a.MaxOutDegree = a.OutDegrees[nd.ID]
		}
	}

	// Density.
	if n > 1 {
		if g.Directed {
			a.Density = float64(e) / float64(n*(n-1))
		} else {
			a.Density = float64(e) / (float64(n*(n-1)) / 2)
		}
	}

	// Average degree.
	if n > 0 {
		if g.Directed {
			a.AvgDegree = float64(e) / float64(n)
		} else {
			a.AvgDegree = 2 * float64(e) / float64(n)
		}
	}

	// Components.
	comps := ConnectedComponents(g)
	a.Components = len(comps)

	// Diameter via all-pairs BFS.
	// Build adjacency list (undirected view for diameter).
	adj := make(map[string][]string, n)
	for _, nd := range nodes {
		adj[nd.ID] = nil
	}
	for _, edge := range edges {
		adj[edge.From] = append(adj[edge.From], edge.To)
		if !g.Directed {
			adj[edge.To] = append(adj[edge.To], edge.From)
		}
	}
	// For directed graphs, use undirected view (both directions) for diameter.
	if g.Directed {
		for _, edge := range edges {
			adj[edge.To] = append(adj[edge.To], edge.From)
		}
	}

	// Sort adjacency lists for determinism.
	for id := range adj {
		sort.Strings(adj[id])
	}

	if len(comps) > 1 {
		// Disconnected graph: diameter is undefined (-1).
		a.Diameter = -1
	} else {
		// Connected: BFS from each node, track max distance.
		maxDist := 0
		for _, nd := range nodes {
			dist := bfsMaxDist(nd.ID, adj)
			if dist > maxDist {
				maxDist = dist
			}
		}
		a.Diameter = maxDist
	}

	return a
}

// bfsMaxDist returns the maximum distance from start to any reachable node.
func bfsMaxDist(start string, adj map[string][]string) int {
	dist := map[string]int{start: 0}
	queue := []string{start}
	maxD := 0
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, nb := range adj[cur] {
			if _, seen := dist[nb]; !seen {
				dist[nb] = dist[cur] + 1
				if dist[nb] > maxD {
					maxD = dist[nb]
				}
				queue = append(queue, nb)
			}
		}
	}
	return maxD
}
