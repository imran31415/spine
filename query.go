package spine

import (
	"errors"
	"fmt"
	"sort"
)

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

// Ancestors returns all transitive predecessors of the given node in a directed graph,
// sorted by ID. For undirected graphs, this returns all reachable nodes.
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
	sort.Strings(result)
	return result
}

// Descendants returns all transitive successors of the given node in a directed graph,
// sorted by ID.
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
	sort.Strings(result)
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

// TransitiveClosure computes the transitive closure of a directed graph.
// Returns a new graph where an edge u->v exists if v is reachable from u.
func TransitiveClosure[N, E any](g *Graph[N, E]) (*Graph[N, E], error) {
	if !g.Directed {
		return nil, errors.New("transitive closure requires a directed graph")
	}

	tc := NewGraph[N, E](true)
	for _, n := range g.Nodes() {
		tc.AddNode(n.ID, n.Data)
	}

	// BFS from each node
	for _, n := range g.Nodes() {
		visited := make(map[string]bool)
		queue := []string{n.ID}
		visited[n.ID] = true
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			for _, nb := range g.Neighbors(cur) {
				if !visited[nb] {
					visited[nb] = true
					queue = append(queue, nb)
				}
			}
		}
		for v := range visited {
			if v != n.ID {
				var zero E
				tc.AddEdge(n.ID, v, zero, 1)
			}
		}
	}

	return tc, nil
}

// ValidationError represents a single graph validation issue.
type ValidationError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	NodeID  string `json:"node_id,omitempty"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
}

// ValidationResult holds the result of graph validation.
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// Validate checks the internal consistency of a graph.
func Validate[N, E any](g *Graph[N, E]) ValidationResult {
	var errs []ValidationError

	// Check that every edge endpoint exists in nodes
	for from, m := range g.out {
		if !g.HasNode(from) {
			errs = append(errs, ValidationError{
				Type:    "dangling_edge",
				Message: fmt.Sprintf("edge source %q not in nodes map", from),
				From:    from,
			})
		}
		for to := range m {
			if !g.HasNode(to) {
				errs = append(errs, ValidationError{
					Type:    "dangling_edge",
					Message: fmt.Sprintf("edge target %q not in nodes map", to),
					To:      to,
					From:    from,
				})
			}
		}
	}

	// Check out/in map symmetry
	for from, m := range g.out {
		for to := range m {
			if _, ok := g.in[to][from]; !ok {
				errs = append(errs, ValidationError{
					Type:    "inconsistent_in_out",
					Message: fmt.Sprintf("edge %q->%q in out map but not in in map", from, to),
					From:    from,
					To:      to,
				})
			}
		}
	}
	for to, m := range g.in {
		for from := range m {
			if _, ok := g.out[from][to]; !ok {
				errs = append(errs, ValidationError{
					Type:    "inconsistent_in_out",
					Message: fmt.Sprintf("edge %q->%q in in map but not in out map", from, to),
					From:    from,
					To:      to,
				})
			}
		}
	}

	// Check rawEdgeCount matches actual count
	actualCount := 0
	for _, m := range g.out {
		actualCount += len(m)
	}
	if actualCount != g.rawEdgeCount {
		errs = append(errs, ValidationError{
			Type:    "count_mismatch",
			Message: fmt.Sprintf("rawEdgeCount=%d but actual out entries=%d", g.rawEdgeCount, actualCount),
		})
	}

	// Sort errors for deterministic output
	sort.Slice(errs, func(i, j int) bool {
		if errs[i].Type != errs[j].Type {
			return errs[i].Type < errs[j].Type
		}
		return errs[i].Message < errs[j].Message
	})

	return ValidationResult{
		Valid:  len(errs) == 0,
		Errors: errs,
	}
}

// DiffResult describes the differences between two graphs.
type DiffResult struct {
	NodesAdded    []string       `json:"nodes_added"`
	NodesRemoved  []string       `json:"nodes_removed"`
	EdgesAdded    [][2]string    `json:"edges_added"`
	EdgesRemoved  [][2]string    `json:"edges_removed"`
	WeightChanges []WeightChange `json:"weight_changes,omitempty"`
}

// WeightChange describes a weight difference for a shared edge.
type WeightChange struct {
	From      string  `json:"from"`
	To        string  `json:"to"`
	OldWeight float64 `json:"old_weight"`
	NewWeight float64 `json:"new_weight"`
}

// Diff computes the differences between two graphs.
func Diff[N, E any](a, b *Graph[N, E]) (*DiffResult, error) {
	if a.Directed != b.Directed {
		return nil, errors.New("cannot diff graphs with different directed modes")
	}

	result := &DiffResult{}

	// Node differences
	aNodes := make(map[string]bool)
	for _, n := range a.Nodes() {
		aNodes[n.ID] = true
	}
	bNodes := make(map[string]bool)
	for _, n := range b.Nodes() {
		bNodes[n.ID] = true
	}

	for id := range bNodes {
		if !aNodes[id] {
			result.NodesAdded = append(result.NodesAdded, id)
		}
	}
	sort.Strings(result.NodesAdded)

	for id := range aNodes {
		if !bNodes[id] {
			result.NodesRemoved = append(result.NodesRemoved, id)
		}
	}
	sort.Strings(result.NodesRemoved)

	// Edge differences
	aEdges := make(map[[2]string]float64)
	for _, e := range a.Edges() {
		aEdges[[2]string{e.From, e.To}] = e.Weight
	}
	bEdges := make(map[[2]string]float64)
	for _, e := range b.Edges() {
		bEdges[[2]string{e.From, e.To}] = e.Weight
	}

	for key, bw := range bEdges {
		if aw, ok := aEdges[key]; !ok {
			result.EdgesAdded = append(result.EdgesAdded, key)
		} else if aw != bw {
			result.WeightChanges = append(result.WeightChanges, WeightChange{
				From:      key[0],
				To:        key[1],
				OldWeight: aw,
				NewWeight: bw,
			})
		}
	}
	sort.Slice(result.EdgesAdded, func(i, j int) bool {
		if result.EdgesAdded[i][0] != result.EdgesAdded[j][0] {
			return result.EdgesAdded[i][0] < result.EdgesAdded[j][0]
		}
		return result.EdgesAdded[i][1] < result.EdgesAdded[j][1]
	})
	sort.Slice(result.WeightChanges, func(i, j int) bool {
		if result.WeightChanges[i].From != result.WeightChanges[j].From {
			return result.WeightChanges[i].From < result.WeightChanges[j].From
		}
		return result.WeightChanges[i].To < result.WeightChanges[j].To
	})

	for key := range aEdges {
		if _, ok := bEdges[key]; !ok {
			result.EdgesRemoved = append(result.EdgesRemoved, key)
		}
	}
	sort.Slice(result.EdgesRemoved, func(i, j int) bool {
		if result.EdgesRemoved[i][0] != result.EdgesRemoved[j][0] {
			return result.EdgesRemoved[i][0] < result.EdgesRemoved[j][0]
		}
		return result.EdgesRemoved[i][1] < result.EdgesRemoved[j][1]
	})

	return result, nil
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
