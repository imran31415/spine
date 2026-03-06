package spine

import (
	"container/heap"
	"errors"
	"math"
	"sort"
)

// BFS performs a breadth-first search starting from the given node.
// The visitor function is called for each visited node. If visitor returns false,
// the traversal stops early. Returns the visited node IDs in BFS order.
func BFS[N, E any](g *Graph[N, E], start string, visitor func(Node[N]) bool) []string {
	if !g.HasNode(start) {
		return nil
	}
	visited := map[string]bool{start: true}
	queue := []string{start}
	var order []string
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		n, _ := g.GetNode(id)
		order = append(order, id)
		if visitor != nil && !visitor(n) {
			break
		}
		neighbors := g.Neighbors(id)
		for _, nb := range neighbors {
			if !visited[nb] {
				visited[nb] = true
				queue = append(queue, nb)
			}
		}
	}
	return order
}

// DFS performs a depth-first search starting from the given node.
// The visitor function is called for each visited node. If visitor returns false,
// the traversal stops early. Returns the visited node IDs in DFS order.
func DFS[N, E any](g *Graph[N, E], start string, visitor func(Node[N]) bool) []string {
	if !g.HasNode(start) {
		return nil
	}
	visited := make(map[string]bool)
	var order []string
	stopped := false
	var walk func(id string)
	walk = func(id string) {
		if stopped || visited[id] {
			return
		}
		visited[id] = true
		n, _ := g.GetNode(id)
		order = append(order, id)
		if visitor != nil && !visitor(n) {
			stopped = true
			return
		}
		neighbors := g.Neighbors(id)
		for _, nb := range neighbors {
			walk(nb)
		}
	}
	walk(start)
	return order
}

// ShortestPath computes the shortest weighted path from src to dst using Dijkstra's algorithm.
// Returns the path as a slice of node IDs and the total cost.
// Returns an error if src or dst don't exist, or no path exists.
func ShortestPath[N, E any](g *Graph[N, E], src, dst string) ([]string, float64, error) {
	if !g.HasNode(src) {
		return nil, 0, errors.New("source node not found")
	}
	if !g.HasNode(dst) {
		return nil, 0, errors.New("destination node not found")
	}

	dist := map[string]float64{src: 0}
	prev := map[string]string{}
	h := &dijkstraHeap{{id: src, dist: 0}}

	for h.Len() > 0 {
		cur := heap.Pop(h).(dijkstraItem)
		if cur.dist > dist[cur.id] {
			continue
		}
		if cur.id == dst {
			break
		}
		for _, e := range g.OutEdges(cur.id) {
			nd := cur.dist + e.Weight
			if d, ok := dist[e.To]; !ok || nd < d {
				dist[e.To] = nd
				prev[e.To] = cur.id
				heap.Push(h, dijkstraItem{id: e.To, dist: nd})
			}
		}
	}

	if _, ok := dist[dst]; !ok {
		return nil, 0, errors.New("no path found")
	}

	// Reconstruct path.
	var path []string
	for cur := dst; cur != ""; cur = prev[cur] {
		path = append(path, cur)
		if cur == src {
			break
		}
	}
	// Reverse.
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path, dist[dst], nil
}

type dijkstraItem struct {
	id   string
	dist float64
}

type dijkstraHeap []dijkstraItem

func (h dijkstraHeap) Len() int            { return len(h) }
func (h dijkstraHeap) Less(i, j int) bool   { return h[i].dist < h[j].dist }
func (h dijkstraHeap) Swap(i, j int)        { h[i], h[j] = h[j], h[i] }
func (h *dijkstraHeap) Push(x interface{}) { *h = append(*h, x.(dijkstraItem)) }
func (h *dijkstraHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

// TopologicalSort returns a topological ordering of the nodes in a directed graph.
// Returns an error if the graph is not directed or contains a cycle.
func TopologicalSort[N, E any](g *Graph[N, E]) ([]string, error) {
	if !g.Directed {
		return nil, errors.New("topological sort requires a directed graph")
	}

	// Kahn's algorithm.
	inDeg := make(map[string]int)
	for _, n := range g.Nodes() {
		inDeg[n.ID] = len(g.InEdges(n.ID))
	}

	var queue []string
	for id, d := range inDeg {
		if d == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	var order []string
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		order = append(order, id)
		neighbors := g.Neighbors(id)
		for _, nb := range neighbors {
			inDeg[nb]--
			if inDeg[nb] == 0 {
				idx := sort.SearchStrings(queue, nb)
				queue = append(queue, "")
				copy(queue[idx+1:], queue[idx:])
				queue[idx] = nb
			}
		}
	}

	if len(order) != g.Order() {
		return nil, errors.New("graph contains a cycle")
	}
	return order, nil
}

// CycleDetect checks if a directed graph contains a cycle.
// Returns true and one cycle path if a cycle exists, false and nil otherwise.
// For undirected graphs it always returns false.
func CycleDetect[N, E any](g *Graph[N, E]) (bool, []string) {
	if !g.Directed {
		return false, nil
	}

	const (
		white = 0
		gray  = 1
		black = 2
	)

	color := make(map[string]int)
	parent := make(map[string]string)

	var cycle []string
	found := false

	var dfs func(id string)
	dfs = func(id string) {
		if found {
			return
		}
		color[id] = gray
		neighbors := g.Neighbors(id)
		for _, nb := range neighbors {
			if found {
				return
			}
			if color[nb] == gray {
				// Found cycle: reconstruct.
				cycle = []string{nb, id}
				for cur := id; cur != nb; {
					cur = parent[cur]
					if cur == nb {
						break
					}
					cycle = append(cycle, cur)
				}
				// Reverse to get forward order.
				for i, j := 0, len(cycle)-1; i < j; i, j = i+1, j-1 {
					cycle[i], cycle[j] = cycle[j], cycle[i]
				}
				found = true
				return
			}
			if color[nb] == white {
				parent[nb] = id
				dfs(nb)
			}
		}
		color[id] = black
	}

	nodes := g.Nodes()
	for _, n := range nodes {
		if color[n.ID] == white {
			dfs(n.ID)
			if found {
				return true, cycle
			}
		}
	}
	return false, nil
}

// Subgraph extracts a new graph containing only the specified node IDs
// and edges between them.
func Subgraph[N, E any](g *Graph[N, E], ids []string) *Graph[N, E] {
	sub := NewGraph[N, E](g.Directed)
	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}
	for _, id := range ids {
		if n, ok := g.GetNode(id); ok {
			sub.AddNode(n.ID, n.Data)
		}
	}
	for _, id := range ids {
		for _, e := range g.OutEdges(id) {
			if idSet[e.To] && !sub.HasEdge(e.From, e.To) {
				sub.AddEdge(e.From, e.To, e.Data, e.Weight)
			}
		}
	}
	// Copy metadata stores for included nodes and edges.
	for _, id := range ids {
		if store, ok := g.nodeMeta[id]; ok {
			sub.nodeMeta[id] = store.Copy()
		}
	}
	for from, m := range g.edgeMeta {
		for to, store := range m {
			if idSet[from] && idSet[to] && sub.HasEdge(from, to) {
				if sub.edgeMeta[from] == nil {
					sub.edgeMeta[from] = make(map[string]*Store)
				}
				sub.edgeMeta[from][to] = store.Copy()
			}
		}
	}
	return sub
}

// StronglyConnectedComponents returns the strongly connected components of a
// directed graph using Tarjan's algorithm. For undirected graphs it delegates
// to ConnectedComponents. Components and their contents are sorted
// deterministically.
func StronglyConnectedComponents[N, E any](g *Graph[N, E]) [][]string {
	if !g.Directed {
		return ConnectedComponents(g)
	}

	nodes := g.Nodes()
	index := make(map[string]int, len(nodes))
	lowlink := make(map[string]int, len(nodes))
	onStack := make(map[string]bool, len(nodes))
	var stack []string
	counter := 0
	var components [][]string

	var strongconnect func(id string)
	strongconnect = func(id string) {
		index[id] = counter
		lowlink[id] = counter
		counter++
		stack = append(stack, id)
		onStack[id] = true

		for _, e := range g.OutEdges(id) {
			w := e.To
			if _, visited := index[w]; !visited {
				strongconnect(w)
				if lowlink[w] < lowlink[id] {
					lowlink[id] = lowlink[w]
				}
			} else if onStack[w] {
				if index[w] < lowlink[id] {
					lowlink[id] = index[w]
				}
			}
		}

		if lowlink[id] == index[id] {
			var comp []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				comp = append(comp, w)
				if w == id {
					break
				}
			}
			sort.Strings(comp)
			components = append(components, comp)
		}
	}

	for _, n := range nodes {
		if _, visited := index[n.ID]; !visited {
			strongconnect(n.ID)
		}
	}

	// Sort components by first element for deterministic output.
	sort.Slice(components, func(i, j int) bool {
		return components[i][0] < components[j][0]
	})
	return components
}

// unionFind implements a disjoint-set data structure for Kruskal's algorithm.
type unionFind struct {
	parent map[string]string
	rank   map[string]int
}

func newUnionFind(ids []string) *unionFind {
	uf := &unionFind{
		parent: make(map[string]string, len(ids)),
		rank:   make(map[string]int, len(ids)),
	}
	for _, id := range ids {
		uf.parent[id] = id
	}
	return uf
}

func (uf *unionFind) find(x string) string {
	for uf.parent[x] != x {
		uf.parent[x] = uf.parent[uf.parent[x]] // path compression
		x = uf.parent[x]
	}
	return x
}

func (uf *unionFind) union(x, y string) bool {
	rx, ry := uf.find(x), uf.find(y)
	if rx == ry {
		return false
	}
	if uf.rank[rx] < uf.rank[ry] {
		rx, ry = ry, rx
	}
	uf.parent[ry] = rx
	if uf.rank[rx] == uf.rank[ry] {
		uf.rank[rx]++
	}
	return true
}

// MinimumSpanningTree computes a minimum spanning tree (or forest) of an
// undirected graph using Kruskal's algorithm. Returns the selected edges,
// the total weight, and an error if the graph is directed.
func MinimumSpanningTree[N, E any](g *Graph[N, E]) ([]Edge[E], float64, error) {
	if g.Directed {
		return nil, 0, errors.New("minimum spanning tree requires an undirected graph")
	}

	edges := g.Edges()
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Weight != edges[j].Weight {
			return edges[i].Weight < edges[j].Weight
		}
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})

	nodes := g.Nodes()
	ids := make([]string, len(nodes))
	for i, n := range nodes {
		ids[i] = n.ID
	}
	uf := newUnionFind(ids)

	var mst []Edge[E]
	totalWeight := 0.0
	for _, e := range edges {
		if uf.union(e.From, e.To) {
			mst = append(mst, e)
			totalWeight += e.Weight
		}
	}
	return mst, totalWeight, nil
}

// AllPairsResult holds the result of all-pairs shortest paths (Floyd-Warshall).
type AllPairsResult struct {
	Dist map[string]map[string]float64 `json:"dist"`
	Next map[string]map[string]string  `json:"next"`
}

// AllPairsShortestPaths computes shortest paths between all pairs using Floyd-Warshall.
// Returns error if a negative cycle is detected.
func AllPairsShortestPaths[N, E any](g *Graph[N, E]) (*AllPairsResult, error) {
	nodes := g.Nodes()
	n := len(nodes)
	ids := make([]string, n)
	for i, nd := range nodes {
		ids[i] = nd.ID
	}

	dist := make(map[string]map[string]float64, n)
	next := make(map[string]map[string]string, n)
	for _, u := range ids {
		dist[u] = make(map[string]float64, n)
		next[u] = make(map[string]string, n)
		for _, v := range ids {
			if u == v {
				dist[u][v] = 0
			} else {
				dist[u][v] = math.Inf(1)
			}
		}
	}

	for _, e := range g.Edges() {
		if e.Weight < dist[e.From][e.To] {
			dist[e.From][e.To] = e.Weight
			next[e.From][e.To] = e.To
		}
		if !g.Directed {
			if e.Weight < dist[e.To][e.From] {
				dist[e.To][e.From] = e.Weight
				next[e.To][e.From] = e.From
			}
		}
	}

	// Floyd-Warshall
	for _, k := range ids {
		for _, i := range ids {
			for _, j := range ids {
				if dist[i][k]+dist[k][j] < dist[i][j] {
					dist[i][j] = dist[i][k] + dist[k][j]
					next[i][j] = next[i][k]
				}
			}
		}
	}

	// Check for negative cycles
	for _, u := range ids {
		if dist[u][u] < 0 {
			return nil, errors.New("graph contains a negative cycle")
		}
	}

	// Remove infinite distances (unreachable pairs) for JSON compatibility
	for u := range dist {
		for v, d := range dist[u] {
			if math.IsInf(d, 1) {
				delete(dist[u], v)
				delete(next[u], v)
			}
		}
	}

	return &AllPairsResult{Dist: dist, Next: next}, nil
}

// ReconstructPath reconstructs the shortest path from src to dst using the Next matrix.
func ReconstructPath(result *AllPairsResult, src, dst string) ([]string, error) {
	if _, ok := result.Dist[src]; !ok {
		return nil, errors.New("source node not found in result")
	}
	if _, ok := result.Dist[dst]; !ok {
		return nil, errors.New("destination node not found in result")
	}
	d, ok := result.Dist[src][dst]
	if !ok || math.IsInf(d, 1) {
		return nil, errors.New("no path found")
	}
	if src == dst {
		return []string{src}, nil
	}

	path := []string{src}
	cur := src
	for cur != dst {
		nxt, ok := result.Next[cur][dst]
		if !ok || nxt == "" {
			return nil, errors.New("no path found")
		}
		path = append(path, nxt)
		cur = nxt
		if len(path) > len(result.Dist)+1 {
			return nil, errors.New("path reconstruction loop detected")
		}
	}
	return path, nil
}

// CriticalPathResult holds the critical path analysis result.
type CriticalPathResult struct {
	Path      []string           `json:"path"`
	Length    float64            `json:"length"`
	NodeSlack map[string]float64 `json:"node_slack"`
}

// CriticalPath computes the critical path in a DAG.
// Edge weights represent task durations. Returns error if graph has cycles or is undirected.
func CriticalPath[N, E any](g *Graph[N, E]) (*CriticalPathResult, error) {
	if !g.Directed {
		return nil, errors.New("critical path requires a directed graph")
	}

	order, err := TopologicalSort(g)
	if err != nil {
		return nil, err
	}

	n := len(order)
	if n == 0 {
		return &CriticalPathResult{
			Path:      nil,
			Length:    0,
			NodeSlack: map[string]float64{},
		}, nil
	}

	// Forward pass: compute earliest start time
	earliest := make(map[string]float64, n)
	for _, id := range order {
		earliest[id] = 0
	}
	for _, id := range order {
		for _, e := range g.OutEdges(id) {
			if earliest[id]+e.Weight > earliest[e.To] {
				earliest[e.To] = earliest[id] + e.Weight
			}
		}
	}

	// Find the maximum earliest time (project duration)
	maxTime := 0.0
	for _, t := range earliest {
		if t > maxTime {
			maxTime = t
		}
	}

	// Backward pass: compute latest start time
	latest := make(map[string]float64, n)
	for _, id := range order {
		latest[id] = maxTime
	}
	for i := n - 1; i >= 0; i-- {
		id := order[i]
		for _, e := range g.OutEdges(id) {
			if latest[e.To]-e.Weight < latest[id] {
				latest[id] = latest[e.To] - e.Weight
			}
		}
	}

	// Compute slack and find critical path
	slack := make(map[string]float64, n)
	for _, id := range order {
		slack[id] = latest[id] - earliest[id]
	}

	// Critical path: nodes with zero slack, in topological order
	var critPath []string
	for _, id := range order {
		if slack[id] == 0 {
			critPath = append(critPath, id)
		}
	}

	return &CriticalPathResult{
		Path:      critPath,
		Length:    maxTime,
		NodeSlack: slack,
	}, nil
}

// ConnectedComponents returns the connected components of the graph
// as a list of node-ID sets. For directed graphs, this finds weakly connected components.
func ConnectedComponents[N, E any](g *Graph[N, E]) [][]string {
	visited := make(map[string]bool)
	var components [][]string

	// For directed graphs, build an undirected view.
	adj := make(map[string]map[string]bool)
	for _, n := range g.Nodes() {
		adj[n.ID] = make(map[string]bool)
	}
	for _, e := range g.Edges() {
		adj[e.From][e.To] = true
		adj[e.To][e.From] = true
	}

	nodes := g.Nodes()
	for _, n := range nodes {
		if visited[n.ID] {
			continue
		}
		var comp []string
		queue := []string{n.ID}
		visited[n.ID] = true
		for len(queue) > 0 {
			id := queue[0]
			queue = queue[1:]
			comp = append(comp, id)
			for nb := range adj[id] {
				if !visited[nb] {
					visited[nb] = true
					queue = append(queue, nb)
				}
			}
		}
		sort.Strings(comp)
		components = append(components, comp)
	}
	return components
}
