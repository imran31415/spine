package spine

import (
	"math"
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

func TestGraphAnalytics(t *testing.T) {
	// Build a connected directed graph: a->b->c->d, a->c
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c", "d"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 1)
	g.AddEdge("c", "d", 0, 1)
	g.AddEdge("a", "c", 0, 1)

	a := GraphAnalytics(g)
	if a.NodeCount != 4 {
		t.Fatalf("expected 4 nodes, got %d", a.NodeCount)
	}
	if a.EdgeCount != 4 {
		t.Fatalf("expected 4 edges, got %d", a.EdgeCount)
	}
	if !a.Directed {
		t.Fatal("expected directed=true")
	}
	// Density for directed: 4/(4*3) = 0.333...
	expectedDensity := 4.0 / 12.0
	if a.Density < expectedDensity-0.001 || a.Density > expectedDensity+0.001 {
		t.Fatalf("expected density ~%.4f, got %.4f", expectedDensity, a.Density)
	}
	// AvgDegree for directed: 4/4 = 1.0
	if a.AvgDegree != 1.0 {
		t.Fatalf("expected avg degree 1.0, got %f", a.AvgDegree)
	}
	if a.MaxOutDegree != 2 { // node a has 2 out-edges
		t.Fatalf("expected max out-degree 2, got %d", a.MaxOutDegree)
	}
	if a.Components != 1 {
		t.Fatalf("expected 1 component, got %d", a.Components)
	}
	// Diameter should be 2 (a-c shortcut means a->c->d is longest shortest path)
	if a.Diameter != 2 {
		t.Fatalf("expected diameter 2, got %d", a.Diameter)
	}
}

func TestGraphAnalyticsDisconnected(t *testing.T) {
	g := NewGraph[int, int](false)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddNode("c", 3)
	g.AddNode("d", 4)
	g.AddEdge("a", "b", 0, 1)
	// c and d are isolated

	a := GraphAnalytics(g)
	if a.Diameter != -1 {
		t.Fatalf("expected diameter -1 for disconnected graph, got %d", a.Diameter)
	}
	if a.Components != 3 {
		t.Fatalf("expected 3 components, got %d", a.Components)
	}
}

func TestGraphAnalyticsEmpty(t *testing.T) {
	g := NewGraph[int, int](true)
	a := GraphAnalytics(g)
	if a.NodeCount != 0 || a.EdgeCount != 0 || a.Components != 0 {
		t.Fatalf("expected empty analytics, got %+v", a)
	}
}

func TestTransitiveClosure(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c", "d"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 1)
	g.AddEdge("c", "d", 0, 1)

	tc, err := TransitiveClosure(g)
	if err != nil {
		t.Fatal(err)
	}
	// a can reach b, c, d
	if !tc.HasEdge("a", "b") || !tc.HasEdge("a", "c") || !tc.HasEdge("a", "d") {
		t.Fatal("expected a to reach b, c, d")
	}
	// b can reach c, d
	if !tc.HasEdge("b", "c") || !tc.HasEdge("b", "d") {
		t.Fatal("expected b to reach c, d")
	}
	// c can reach d
	if !tc.HasEdge("c", "d") {
		t.Fatal("expected c to reach d")
	}
	// d can't reach anyone
	if tc.HasEdge("d", "a") || tc.HasEdge("d", "b") || tc.HasEdge("d", "c") {
		t.Fatal("d should not reach anyone")
	}
}

func TestTransitiveClosureUndirectedError(t *testing.T) {
	g := NewGraph[int, int](false)
	_, err := TransitiveClosure(g)
	if err == nil {
		t.Fatal("expected error for undirected graph")
	}
}

func TestValidate(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddEdge("a", "b", 0, 1)

	result := Validate(g)
	if !result.Valid {
		t.Fatalf("expected valid graph, got errors: %v", result.Errors)
	}
}

func TestValidateEmpty(t *testing.T) {
	g := NewGraph[int, int](true)
	result := Validate(g)
	if !result.Valid {
		t.Fatal("expected empty graph to be valid")
	}
}

func TestDiff(t *testing.T) {
	a := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c"} {
		a.AddNode(id, id)
	}
	a.AddEdge("a", "b", 0, 1)
	a.AddEdge("b", "c", 0, 2)

	b := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "d"} {
		b.AddNode(id, id)
	}
	b.AddEdge("a", "b", 0, 5) // weight changed
	b.AddEdge("a", "d", 0, 1) // new edge

	result, err := Diff(a, b)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.NodesAdded) != 1 || result.NodesAdded[0] != "d" {
		t.Fatalf("expected nodes added [d], got %v", result.NodesAdded)
	}
	if len(result.NodesRemoved) != 1 || result.NodesRemoved[0] != "c" {
		t.Fatalf("expected nodes removed [c], got %v", result.NodesRemoved)
	}
	if len(result.EdgesAdded) != 1 {
		t.Fatalf("expected 1 edge added, got %v", result.EdgesAdded)
	}
	if len(result.EdgesRemoved) != 1 {
		t.Fatalf("expected 1 edge removed, got %v", result.EdgesRemoved)
	}
	if len(result.WeightChanges) != 1 {
		t.Fatalf("expected 1 weight change, got %v", result.WeightChanges)
	}
	if result.WeightChanges[0].OldWeight != 1 || result.WeightChanges[0].NewWeight != 5 {
		t.Fatalf("expected weight change 1->5, got %v", result.WeightChanges[0])
	}
}

func TestDiffMixedDirectedError(t *testing.T) {
	a := NewGraph[int, int](true)
	b := NewGraph[int, int](false)
	_, err := Diff(a, b)
	if err == nil {
		t.Fatal("expected error for mixed directed modes")
	}
}

// Suppress unused import warning
var _ = math.Abs
