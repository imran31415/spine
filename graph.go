package spine

import (
	"fmt"
	"sort"
)

// Node represents a vertex in the graph with typed data.
type Node[T any] struct {
	ID   string
	Data T
}

// Edge represents a connection between two nodes with typed data and a weight.
type Edge[T any] struct {
	From   string
	To     string
	Data   T
	Weight float64
}

// Graph is a generic graph supporting both directed and undirected modes.
// N is the node data type and E is the edge data type.
type Graph[N, E any] struct {
	Directed bool
	nodes    map[string]Node[N]
	out      map[string]map[string]Edge[E] // from -> to -> edge
	in       map[string]map[string]Edge[E] // to -> from -> edge
}

// NewGraph creates a new graph. If directed is true, edges are one-way.
func NewGraph[N, E any](directed bool) *Graph[N, E] {
	return &Graph[N, E]{
		Directed: directed,
		nodes:    make(map[string]Node[N]),
		out:      make(map[string]map[string]Edge[E]),
		in:       make(map[string]map[string]Edge[E]),
	}
}

// AddNode adds a node to the graph. If a node with the same ID exists, it is overwritten.
func (g *Graph[N, E]) AddNode(id string, data N) {
	g.nodes[id] = Node[N]{ID: id, Data: data}
	if g.out[id] == nil {
		g.out[id] = make(map[string]Edge[E])
	}
	if g.in[id] == nil {
		g.in[id] = make(map[string]Edge[E])
	}
}

// AddEdge adds an edge between two nodes. Both nodes must already exist.
// Returns an error if either node is missing.
func (g *Graph[N, E]) AddEdge(from, to string, data E, weight float64) error {
	if !g.HasNode(from) {
		return fmt.Errorf("node %q not found", from)
	}
	if !g.HasNode(to) {
		return fmt.Errorf("node %q not found", to)
	}
	e := Edge[E]{From: from, To: to, Data: data, Weight: weight}
	g.out[from][to] = e
	g.in[to][from] = e
	if !g.Directed {
		rev := Edge[E]{From: to, To: from, Data: data, Weight: weight}
		g.out[to][from] = rev
		g.in[from][to] = rev
	}
	return nil
}

// RemoveNode removes a node and all its incident edges.
func (g *Graph[N, E]) RemoveNode(id string) {
	if !g.HasNode(id) {
		return
	}
	// Remove outgoing edges
	for to := range g.out[id] {
		delete(g.in[to], id)
	}
	// Remove incoming edges
	for from := range g.in[id] {
		delete(g.out[from], id)
	}
	delete(g.out, id)
	delete(g.in, id)
	delete(g.nodes, id)
}

// RemoveEdge removes the edge from -> to.
func (g *Graph[N, E]) RemoveEdge(from, to string) {
	delete(g.out[from], to)
	delete(g.in[to], from)
	if !g.Directed {
		delete(g.out[to], from)
		delete(g.in[from], to)
	}
}

// GetNode returns the node with the given ID and true, or the zero value and false.
func (g *Graph[N, E]) GetNode(id string) (Node[N], bool) {
	n, ok := g.nodes[id]
	return n, ok
}

// GetEdge returns the edge from -> to and true, or the zero value and false.
func (g *Graph[N, E]) GetEdge(from, to string) (Edge[E], bool) {
	if m, ok := g.out[from]; ok {
		e, ok := m[to]
		return e, ok
	}
	var zero Edge[E]
	return zero, false
}

// HasNode returns true if the node exists.
func (g *Graph[N, E]) HasNode(id string) bool {
	_, ok := g.nodes[id]
	return ok
}

// HasEdge returns true if an edge from -> to exists.
func (g *Graph[N, E]) HasEdge(from, to string) bool {
	if m, ok := g.out[from]; ok {
		_, ok := m[to]
		return ok
	}
	return false
}

// Neighbors returns the IDs of nodes adjacent to the given node (outgoing direction).
func (g *Graph[N, E]) Neighbors(id string) []string {
	m := g.out[id]
	result := make([]string, 0, len(m))
	for to := range m {
		result = append(result, to)
	}
	sort.Strings(result)
	return result
}

// OutEdges returns all edges originating from the given node.
func (g *Graph[N, E]) OutEdges(id string) []Edge[E] {
	m := g.out[id]
	result := make([]Edge[E], 0, len(m))
	for _, e := range m {
		result = append(result, e)
	}
	return result
}

// InEdges returns all edges pointing to the given node.
func (g *Graph[N, E]) InEdges(id string) []Edge[E] {
	m := g.in[id]
	result := make([]Edge[E], 0, len(m))
	for _, e := range m {
		result = append(result, e)
	}
	return result
}

// Nodes returns all nodes in the graph in sorted order by ID.
func (g *Graph[N, E]) Nodes() []Node[N] {
	result := make([]Node[N], 0, len(g.nodes))
	for _, n := range g.nodes {
		result = append(result, n)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

// Edges returns all edges in the graph.
func (g *Graph[N, E]) Edges() []Edge[E] {
	seen := make(map[[2]string]bool)
	var result []Edge[E]
	for from, m := range g.out {
		for to, e := range m {
			key := [2]string{from, to}
			if !g.Directed {
				// For undirected graphs, only include each edge once.
				revKey := [2]string{to, from}
				if seen[revKey] {
					continue
				}
			}
			if !seen[key] {
				seen[key] = true
				result = append(result, e)
			}
		}
	}
	return result
}

// Order returns the number of nodes.
func (g *Graph[N, E]) Order() int {
	return len(g.nodes)
}

// Size returns the number of edges.
func (g *Graph[N, E]) Size() int {
	return len(g.Edges())
}

// Copy returns a deep copy of the graph.
func (g *Graph[N, E]) Copy() *Graph[N, E] {
	c := NewGraph[N, E](g.Directed)
	for id, n := range g.nodes {
		c.nodes[id] = n
		c.out[id] = make(map[string]Edge[E])
		c.in[id] = make(map[string]Edge[E])
	}
	for from, m := range g.out {
		for to, e := range m {
			c.out[from][to] = e
		}
	}
	for to, m := range g.in {
		for from, e := range m {
			c.in[to][from] = e
		}
	}
	return c
}
