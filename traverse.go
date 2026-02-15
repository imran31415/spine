package spine

import (
	"container/heap"
	"errors"
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
				queue = append(queue, nb)
				sort.Strings(queue)
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
