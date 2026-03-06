package spine

import (
	"testing"
)

func TestMaxFlowSimple(t *testing.T) {
	// s -> a (cap 10), s -> b (cap 5), a -> t (cap 5), b -> t (cap 10), a -> b (cap 3)
	g := NewGraph[string, int](true)
	for _, id := range []string{"s", "a", "b", "t"} {
		g.AddNode(id, id)
	}
	g.AddEdge("s", "a", 0, 10)
	g.AddEdge("s", "b", 0, 5)
	g.AddEdge("a", "t", 0, 5)
	g.AddEdge("b", "t", 0, 10)
	g.AddEdge("a", "b", 0, 3)

	result, err := MaxFlow(g, "s", "t")
	if err != nil {
		t.Fatal(err)
	}
	// Max flow should be 13: s->a->t(5), s->a->b->t(3), s->b->t(5)
	if result.MaxFlow != 13 {
		t.Fatalf("expected max flow 13, got %f", result.MaxFlow)
	}
	if len(result.MinCut) == 0 {
		t.Fatal("expected non-empty min-cut")
	}
}

func TestMaxFlowLinear(t *testing.T) {
	// s -> a (cap 3) -> t (cap 5). Bottleneck = 3
	g := NewGraph[string, int](true)
	for _, id := range []string{"s", "a", "t"} {
		g.AddNode(id, id)
	}
	g.AddEdge("s", "a", 0, 3)
	g.AddEdge("a", "t", 0, 5)

	result, err := MaxFlow(g, "s", "t")
	if err != nil {
		t.Fatal(err)
	}
	if result.MaxFlow != 3 {
		t.Fatalf("expected max flow 3, got %f", result.MaxFlow)
	}
}

func TestMaxFlowNoPath(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("s", "S")
	g.AddNode("t", "T")
	// No edge between them

	result, err := MaxFlow(g, "s", "t")
	if err != nil {
		t.Fatal(err)
	}
	if result.MaxFlow != 0 {
		t.Fatalf("expected max flow 0, got %f", result.MaxFlow)
	}
}

func TestMaxFlowUndirectedError(t *testing.T) {
	g := NewGraph[string, int](false)
	g.AddNode("s", "S")
	g.AddNode("t", "T")
	g.AddEdge("s", "t", 0, 5)

	_, err := MaxFlow(g, "s", "t")
	if err == nil {
		t.Fatal("expected error for undirected graph")
	}
}

func TestMaxFlowMissingNodes(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")

	_, err := MaxFlow(g, "x", "a")
	if err == nil {
		t.Fatal("expected error for missing source")
	}
	_, err = MaxFlow(g, "a", "x")
	if err == nil {
		t.Fatal("expected error for missing sink")
	}
}

func TestMaxFlowSameSourceSink(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")

	_, err := MaxFlow(g, "a", "a")
	if err == nil {
		t.Fatal("expected error for same source and sink")
	}
}
