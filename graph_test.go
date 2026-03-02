package spine

import (
	"testing"
)

func TestAddNodeAndGetNode(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "alpha")
	g.AddNode("b", "beta")

	n, ok := g.GetNode("a")
	if !ok || n.Data != "alpha" {
		t.Fatalf("expected alpha, got %v", n)
	}

	_, ok = g.GetNode("z")
	if ok {
		t.Fatal("expected missing node")
	}
}

func TestAddEdge(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")

	if err := g.AddEdge("a", "b", "edge-ab", 1.5); err != nil {
		t.Fatal(err)
	}
	e, ok := g.GetEdge("a", "b")
	if !ok {
		t.Fatal("edge not found")
	}
	if e.Weight != 1.5 || e.Data != "edge-ab" {
		t.Fatalf("unexpected edge: %+v", e)
	}

	// Missing node
	if err := g.AddEdge("a", "z", "x", 0); err == nil {
		t.Fatal("expected error for missing node")
	}
}

func TestRemoveNode(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddNode("c", 3)
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 1)

	g.RemoveNode("b")
	if g.HasNode("b") {
		t.Fatal("b should be removed")
	}
	if g.HasEdge("a", "b") || g.HasEdge("b", "c") {
		t.Fatal("edges involving b should be removed")
	}
	if g.Order() != 2 {
		t.Fatalf("expected 2 nodes, got %d", g.Order())
	}
}

func TestRemoveEdge(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddEdge("a", "b", 0, 1)

	g.RemoveEdge("a", "b")
	if g.HasEdge("a", "b") {
		t.Fatal("edge should be removed")
	}
}

func TestUndirectedGraph(t *testing.T) {
	g := NewGraph[string, int](false)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddEdge("a", "b", 1, 1.0)

	if !g.HasEdge("a", "b") || !g.HasEdge("b", "a") {
		t.Fatal("undirected edge should exist in both directions")
	}

	neighbors := g.Neighbors("b")
	if len(neighbors) != 1 || neighbors[0] != "a" {
		t.Fatalf("expected [a], got %v", neighbors)
	}

	// Size should count undirected edge once
	if g.Size() != 1 {
		t.Fatalf("expected size 1, got %d", g.Size())
	}

	// Remove edge from one direction
	g.RemoveEdge("b", "a")
	if g.HasEdge("a", "b") || g.HasEdge("b", "a") {
		t.Fatal("both directions should be removed")
	}
}

func TestNeighborsAndEdges(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddNode("c", 3)
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("a", "c", 0, 2)

	neighbors := g.Neighbors("a")
	if len(neighbors) != 2 {
		t.Fatalf("expected 2 neighbors, got %d", len(neighbors))
	}

	out := g.OutEdges("a")
	if len(out) != 2 {
		t.Fatalf("expected 2 out edges, got %d", len(out))
	}

	in := g.InEdges("b")
	if len(in) != 1 || in[0].From != "a" {
		t.Fatalf("expected 1 in edge from a, got %v", in)
	}
}

func TestNodesSorted(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("c", 3)
	g.AddNode("a", 1)
	g.AddNode("b", 2)

	nodes := g.Nodes()
	if len(nodes) != 3 || nodes[0].ID != "a" || nodes[1].ID != "b" || nodes[2].ID != "c" {
		t.Fatalf("nodes not sorted: %v", nodes)
	}
}

func TestOrderAndSize(t *testing.T) {
	g := NewGraph[int, int](true)
	if g.Order() != 0 || g.Size() != 0 {
		t.Fatal("empty graph should have order 0 size 0")
	}
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddEdge("a", "b", 0, 1)
	if g.Order() != 2 || g.Size() != 1 {
		t.Fatalf("expected order=2 size=1, got order=%d size=%d", g.Order(), g.Size())
	}
}

func TestCopy(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddEdge("a", "b", 1, 2.0)

	c := g.Copy()
	// Modify original
	g.RemoveNode("a")

	if !c.HasNode("a") {
		t.Fatal("copy should be independent")
	}
	if !c.HasEdge("a", "b") {
		t.Fatal("copy should retain edges")
	}
}

func TestSelfLoop(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("a", 1)
	if err := g.AddEdge("a", "a", 0, 1); err != nil {
		t.Fatal(err)
	}
	if !g.HasEdge("a", "a") {
		t.Fatal("self-loop should exist")
	}
	if g.Size() != 1 {
		t.Fatalf("expected size 1, got %d", g.Size())
	}
}

func TestDuplicateAddNode(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "first")
	g.AddNode("a", "second")
	n, _ := g.GetNode("a")
	if n.Data != "second" {
		t.Fatal("duplicate add should overwrite")
	}
}

func TestSizeCounter(t *testing.T) {
	// Directed: add, overwrite, remove
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddNode("c", "C")

	g.AddEdge("a", "b", 1, 1)
	if g.Size() != 1 {
		t.Fatalf("expected 1, got %d", g.Size())
	}
	// Overwrite same edge — size should stay 1
	g.AddEdge("a", "b", 2, 2)
	if g.Size() != 1 {
		t.Fatalf("expected 1 after overwrite, got %d", g.Size())
	}
	g.AddEdge("b", "c", 1, 1)
	if g.Size() != 2 {
		t.Fatalf("expected 2, got %d", g.Size())
	}
	g.RemoveEdge("a", "b")
	if g.Size() != 1 {
		t.Fatalf("expected 1 after remove, got %d", g.Size())
	}
	// Remove non-existent edge — no change
	g.RemoveEdge("a", "b")
	if g.Size() != 1 {
		t.Fatalf("expected 1 after duplicate remove, got %d", g.Size())
	}

	// Undirected: add, remove
	u := NewGraph[string, int](false)
	u.AddNode("x", "X")
	u.AddNode("y", "Y")
	u.AddEdge("x", "y", 1, 1)
	if u.Size() != 1 {
		t.Fatalf("undirected: expected 1, got %d", u.Size())
	}
	u.RemoveEdge("y", "x")
	if u.Size() != 0 {
		t.Fatalf("undirected: expected 0 after remove, got %d", u.Size())
	}

	// Self-loop
	s := NewGraph[string, int](true)
	s.AddNode("a", "A")
	s.AddEdge("a", "a", 1, 1)
	if s.Size() != 1 {
		t.Fatalf("self-loop: expected 1, got %d", s.Size())
	}
	s.RemoveNode("a")
	if s.Size() != 0 {
		t.Fatalf("self-loop after RemoveNode: expected 0, got %d", s.Size())
	}

	// RemoveNode cascading
	g2 := NewGraph[string, int](true)
	g2.AddNode("a", "A")
	g2.AddNode("b", "B")
	g2.AddNode("c", "C")
	g2.AddEdge("a", "b", 1, 1)
	g2.AddEdge("b", "c", 1, 1)
	g2.AddEdge("c", "b", 1, 1)
	if g2.Size() != 3 {
		t.Fatalf("expected 3, got %d", g2.Size())
	}
	g2.RemoveNode("b")
	if g2.Size() != 0 {
		t.Fatalf("expected 0 after removing b, got %d", g2.Size())
	}

	// Copy preserves size
	g3 := NewGraph[string, int](true)
	g3.AddNode("a", "A")
	g3.AddNode("b", "B")
	g3.AddEdge("a", "b", 1, 1)
	c := g3.Copy()
	if c.Size() != 1 {
		t.Fatalf("copy: expected 1, got %d", c.Size())
	}
}

func TestEmptyGraphOperations(t *testing.T) {
	g := NewGraph[int, int](true)
	g.RemoveNode("nonexistent")
	g.RemoveEdge("a", "b")
	if g.HasNode("a") || g.HasEdge("a", "b") {
		t.Fatal("operations on empty graph should be safe")
	}
	neighbors := g.Neighbors("a")
	if len(neighbors) != 0 {
		t.Fatal("neighbors of nonexistent node should be empty")
	}
}
