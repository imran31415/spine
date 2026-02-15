package spine

import (
	"sort"
	"testing"
)

func TestFilterNodes(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddNode("c", 3)

	result := FilterNodes(g, func(n Node[int]) bool {
		return n.Data > 1
	})
	if len(result) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(result))
	}
}

func TestFilterEdges(t *testing.T) {
	g := NewGraph[int, float64](true)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddNode("c", 3)
	g.AddEdge("a", "b", 1.0, 1)
	g.AddEdge("b", "c", 2.0, 1)

	result := FilterEdges(g, func(e Edge[float64]) bool {
		return e.Data > 1.5
	})
	if len(result) != 1 || result[0].From != "b" {
		t.Fatalf("expected edge b->c, got %v", result)
	}
}

func TestAncestors(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c", "d"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 0)
	g.AddEdge("b", "c", 0, 0)
	g.AddEdge("a", "d", 0, 0)
	g.AddEdge("d", "c", 0, 0)

	anc := Ancestors(g, "c")
	sort.Strings(anc)
	if len(anc) != 3 {
		t.Fatalf("expected 3 ancestors, got %v", anc)
	}
	// a, b, d are all ancestors of c
	expected := []string{"a", "b", "d"}
	for i, e := range expected {
		if anc[i] != e {
			t.Fatalf("expected %v, got %v", expected, anc)
		}
	}
}

func TestDescendants(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c", "d"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 0)
	g.AddEdge("b", "c", 0, 0)
	g.AddEdge("b", "d", 0, 0)

	desc := Descendants(g, "a")
	sort.Strings(desc)
	expected := []string{"b", "c", "d"}
	if len(desc) != 3 {
		t.Fatalf("expected 3 descendants, got %v", desc)
	}
	for i, e := range expected {
		if desc[i] != e {
			t.Fatalf("expected %v, got %v", expected, desc)
		}
	}
}

func TestRoots(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddNode("c", "C")
	g.AddEdge("a", "b", 0, 0)
	g.AddEdge("a", "c", 0, 0)

	roots := Roots(g)
	if len(roots) != 1 || roots[0].ID != "a" {
		t.Fatalf("expected root [a], got %v", roots)
	}
}

func TestLeaves(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddNode("c", "C")
	g.AddEdge("a", "b", 0, 0)
	g.AddEdge("a", "c", 0, 0)

	leaves := Leaves(g)
	if len(leaves) != 2 {
		t.Fatalf("expected 2 leaves, got %v", leaves)
	}
}

func TestRootsAndLeavesEmpty(t *testing.T) {
	g := NewGraph[int, int](true)
	if roots := Roots(g); len(roots) != 0 {
		t.Fatal("expected no roots")
	}
	if leaves := Leaves(g); len(leaves) != 0 {
		t.Fatal("expected no leaves")
	}
}

func TestAncestorsOfRoot(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddEdge("a", "b", 0, 0)

	anc := Ancestors(g, "a")
	if len(anc) != 0 {
		t.Fatalf("root should have no ancestors, got %v", anc)
	}
}

func TestDescendantsOfLeaf(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddEdge("a", "b", 0, 0)

	desc := Descendants(g, "b")
	if len(desc) != 0 {
		t.Fatalf("leaf should have no descendants, got %v", desc)
	}
}
