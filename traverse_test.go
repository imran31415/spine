package spine

import (
	"testing"
)

func TestBFS(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c", "d"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("a", "c", 0, 1)
	g.AddEdge("b", "d", 0, 1)

	order := BFS(g, "a", nil)
	if len(order) != 4 || order[0] != "a" {
		t.Fatalf("unexpected BFS order: %v", order)
	}
	// d should come after b and c
	dIdx := indexOf(order, "d")
	bIdx := indexOf(order, "b")
	if dIdx <= bIdx {
		t.Fatalf("d should come after b in BFS: %v", order)
	}
}

func TestBFSEarlyStop(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "a")
	g.AddNode("b", "b")
	g.AddNode("c", "c")
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 1)

	order := BFS(g, "a", func(n Node[string]) bool {
		return n.ID != "b" // stop at b
	})
	if indexOf(order, "c") != -1 {
		t.Fatalf("should not visit c: %v", order)
	}
}

func TestBFSMissingStart(t *testing.T) {
	g := NewGraph[int, int](true)
	order := BFS(g, "x", nil)
	if order != nil {
		t.Fatal("expected nil for missing start")
	}
}

func TestDFS(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c", "d"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 1)
	g.AddEdge("a", "d", 0, 1)

	order := DFS(g, "a", nil)
	if len(order) != 4 || order[0] != "a" {
		t.Fatalf("unexpected DFS order: %v", order)
	}
}

func TestDFSEarlyStop(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "a")
	g.AddNode("b", "b")
	g.AddNode("c", "c")
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 1)

	order := DFS(g, "a", func(n Node[string]) bool {
		return n.ID != "b"
	})
	if indexOf(order, "c") != -1 {
		t.Fatalf("should not visit c after stopping at b: %v", order)
	}
}

func TestShortestPath(t *testing.T) {
	g := NewGraph[string, string](true)
	for _, id := range []string{"a", "b", "c", "d"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", "", 1)
	g.AddEdge("b", "d", "", 2)
	g.AddEdge("a", "c", "", 1)
	g.AddEdge("c", "d", "", 1)

	path, cost, err := ShortestPath(g, "a", "d")
	if err != nil {
		t.Fatal(err)
	}
	if cost != 2 {
		t.Fatalf("expected cost 2, got %f", cost)
	}
	// Should go a -> c -> d (cost 2) not a -> b -> d (cost 3)
	if len(path) != 3 || path[0] != "a" || path[1] != "c" || path[2] != "d" {
		t.Fatalf("expected [a c d], got %v", path)
	}
}

func TestShortestPathNoPath(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	// No edge between them.
	_, _, err := ShortestPath(g, "a", "b")
	if err == nil {
		t.Fatal("expected error for no path")
	}
}

func TestShortestPathMissingNode(t *testing.T) {
	g := NewGraph[int, int](true)
	_, _, err := ShortestPath(g, "a", "b")
	if err == nil {
		t.Fatal("expected error for missing node")
	}
}

func TestTopologicalSort(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c", "d"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 0)
	g.AddEdge("a", "c", 0, 0)
	g.AddEdge("b", "d", 0, 0)
	g.AddEdge("c", "d", 0, 0)

	order, err := TopologicalSort(g)
	if err != nil {
		t.Fatal(err)
	}
	if len(order) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(order))
	}
	// a must come before b and c; b and c before d
	if indexOf(order, "a") >= indexOf(order, "b") || indexOf(order, "a") >= indexOf(order, "c") {
		t.Fatalf("a should come first: %v", order)
	}
	if indexOf(order, "d") <= indexOf(order, "b") || indexOf(order, "d") <= indexOf(order, "c") {
		t.Fatalf("d should come last: %v", order)
	}
}

func TestTopologicalSortCycle(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddEdge("a", "b", 0, 0)
	g.AddEdge("b", "a", 0, 0)

	_, err := TopologicalSort(g)
	if err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestTopologicalSortUndirected(t *testing.T) {
	g := NewGraph[int, int](false)
	_, err := TopologicalSort(g)
	if err == nil {
		t.Fatal("expected error for undirected graph")
	}
}

func TestCycleDetect(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddNode("c", 3)
	g.AddEdge("a", "b", 0, 0)
	g.AddEdge("b", "c", 0, 0)
	g.AddEdge("c", "a", 0, 0)

	hasCycle, cycle := CycleDetect(g)
	if !hasCycle {
		t.Fatal("expected cycle")
	}
	if len(cycle) < 2 {
		t.Fatalf("cycle too short: %v", cycle)
	}
}

func TestCycleDetectNoCycle(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddEdge("a", "b", 0, 0)

	hasCycle, _ := CycleDetect(g)
	if hasCycle {
		t.Fatal("expected no cycle")
	}
}

func TestCycleDetectUndirected(t *testing.T) {
	g := NewGraph[int, int](false)
	hasCycle, _ := CycleDetect(g)
	if hasCycle {
		t.Fatal("undirected should return false")
	}
}

func TestSubgraph(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c", "d"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 1, 1)
	g.AddEdge("b", "c", 2, 1)
	g.AddEdge("c", "d", 3, 1)

	sub := Subgraph(g, []string{"a", "b", "c"})
	if sub.Order() != 3 {
		t.Fatalf("expected 3 nodes, got %d", sub.Order())
	}
	if !sub.HasEdge("a", "b") || !sub.HasEdge("b", "c") {
		t.Fatal("expected edges a->b and b->c")
	}
	if sub.HasEdge("c", "d") {
		t.Fatal("d should not be in subgraph")
	}
}

func TestConnectedComponents(t *testing.T) {
	g := NewGraph[int, int](false)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddNode("c", 3)
	g.AddNode("d", 4)
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("c", "d", 0, 1)

	comps := ConnectedComponents(g)
	if len(comps) != 2 {
		t.Fatalf("expected 2 components, got %d", len(comps))
	}
}

func TestConnectedComponentsDirected(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddNode("c", 3)
	g.AddEdge("a", "b", 0, 1)
	// c is isolated

	comps := ConnectedComponents(g)
	if len(comps) != 2 {
		t.Fatalf("expected 2 components, got %d: %v", len(comps), comps)
	}
}

func TestConnectedComponentsSingleNode(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("a", 1)

	comps := ConnectedComponents(g)
	if len(comps) != 1 || len(comps[0]) != 1 {
		t.Fatalf("expected 1 component with 1 node, got %v", comps)
	}
}

func indexOf(s []string, v string) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return -1
}
