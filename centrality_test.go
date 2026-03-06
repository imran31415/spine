package spine

import (
	"math"
	"testing"
)

func TestDegreeCentrality(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("a", "c", 0, 1)

	result := DegreeCentrality(g)
	// a has out-degree 2, n-1=2 => 1.0
	if result.Scores["a"] != 1.0 {
		t.Fatalf("expected a=1.0, got %f", result.Scores["a"])
	}
	if result.Scores["b"] != 0.0 {
		t.Fatalf("expected b=0.0, got %f", result.Scores["b"])
	}
}

func TestDegreeCentralitySingleNode(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")
	result := DegreeCentrality(g)
	if result.Scores["a"] != 0 {
		t.Fatalf("expected 0 for single node, got %f", result.Scores["a"])
	}
}

func TestDegreeCentralityUndirected(t *testing.T) {
	g := NewGraph[string, int](false)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddNode("c", "C")
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("a", "c", 0, 1)

	result := DegreeCentrality(g)
	// a has degree 2, n-1=2 => 1.0
	if result.Scores["a"] != 1.0 {
		t.Fatalf("expected a=1.0, got %f", result.Scores["a"])
	}
	// b has degree 1, n-1=2 => 0.5
	if result.Scores["b"] != 0.5 {
		t.Fatalf("expected b=0.5, got %f", result.Scores["b"])
	}
}

func TestBetweennessCentrality(t *testing.T) {
	// Star graph: a is the center connected to b, c, d
	g := NewGraph[string, int](false)
	for _, id := range []string{"a", "b", "c", "d"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("a", "c", 0, 1)
	g.AddEdge("a", "d", 0, 1)

	result := BetweennessCentrality(g)
	// a should be the most central (all shortest paths go through a)
	if result.Scores["a"] <= result.Scores["b"] {
		t.Fatalf("expected a to have highest betweenness: a=%f, b=%f",
			result.Scores["a"], result.Scores["b"])
	}
	// b, c, d should have 0 betweenness
	if result.Scores["b"] != 0 {
		t.Fatalf("expected b=0, got %f", result.Scores["b"])
	}
}

func TestBetweennessCentralityDirected(t *testing.T) {
	// Linear: a->b->c
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 1)

	result := BetweennessCentrality(g)
	// b is on the only path from a to c
	if result.Scores["b"] != 1.0 {
		t.Fatalf("expected b=1.0, got %f", result.Scores["b"])
	}
}

func TestClosenessCentrality(t *testing.T) {
	// Linear: a->b->c
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 1)

	result := ClosenessCentrality(g)
	// From a: dist(b)=1, dist(c)=2. closeness = 2/3
	expected := 2.0 / 3.0
	if math.Abs(result.Scores["a"]-expected) > 0.001 {
		t.Fatalf("expected a~%.4f, got %f", expected, result.Scores["a"])
	}
	// From c: no outgoing => reachable=0 => 0
	if result.Scores["c"] != 0 {
		t.Fatalf("expected c=0, got %f", result.Scores["c"])
	}
}

func TestPageRank(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 1)
	g.AddEdge("c", "a", 0, 1)

	result := PageRank(g, 0.85, 100, 1e-6)
	if !result.Converged {
		t.Fatal("expected convergence")
	}
	// In a cycle, all nodes should have equal PageRank
	diff := math.Abs(result.Scores["a"] - result.Scores["b"])
	if diff > 0.01 {
		t.Fatalf("expected similar scores in cycle: a=%f, b=%f", result.Scores["a"], result.Scores["b"])
	}
	// Sum should be ~1.0
	sum := result.Scores["a"] + result.Scores["b"] + result.Scores["c"]
	if math.Abs(sum-1.0) > 0.01 {
		t.Fatalf("expected sum~1.0, got %f", sum)
	}
}

func TestPageRankDanglingNodes(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddEdge("a", "b", 0, 1)
	// b is a dangling node (no outgoing)

	result := PageRank(g, 0.85, 100, 1e-6)
	if !result.Converged {
		t.Fatal("expected convergence")
	}
	// b should have higher rank than a since it receives from a + dangling redistribution
	if result.Scores["b"] <= result.Scores["a"] {
		t.Fatalf("expected b > a: b=%f, a=%f", result.Scores["b"], result.Scores["a"])
	}
}

func TestPageRankEmpty(t *testing.T) {
	g := NewGraph[string, int](true)
	result := PageRank(g, 0.85, 100, 1e-6)
	if !result.Converged {
		t.Fatal("expected convergence for empty graph")
	}
	if len(result.Scores) != 0 {
		t.Fatalf("expected empty scores, got %v", result.Scores)
	}
}
