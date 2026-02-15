package spine

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
