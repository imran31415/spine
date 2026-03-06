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
	if len(cycle) != 3 {
		t.Fatalf("expected 3-node cycle, got %v", cycle)
	}
	// Verify cycle is complete: each node in cycle should have an edge to the next,
	// and the last should have an edge back to the first.
	for i := 0; i < len(cycle); i++ {
		from := cycle[i]
		to := cycle[(i+1)%len(cycle)]
		if !g.HasEdge(from, to) {
			t.Fatalf("cycle broken: no edge %s -> %s in cycle %v", from, to, cycle)
		}
	}
}

func TestCycleDetectNestedCycle(t *testing.T) {
	// x -> a -> b -> c -> d -> b (cycle is b->c->d->b, not involving x or a)
	g := NewGraph[int, int](true)
	g.AddNode("x", 0)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddNode("c", 3)
	g.AddNode("d", 4)
	g.AddEdge("x", "a", 0, 0)
	g.AddEdge("a", "b", 0, 0)
	g.AddEdge("b", "c", 0, 0)
	g.AddEdge("c", "d", 0, 0)
	g.AddEdge("d", "b", 0, 0)

	hasCycle, cycle := CycleDetect(g)
	if !hasCycle {
		t.Fatal("expected cycle")
	}
	if len(cycle) < 2 {
		t.Fatalf("cycle too short: %v", cycle)
	}
	// Verify cycle is valid: each consecutive pair has an edge
	for i := 0; i < len(cycle); i++ {
		from := cycle[i]
		to := cycle[(i+1)%len(cycle)]
		if !g.HasEdge(from, to) {
			t.Fatalf("cycle broken: no edge %s -> %s in cycle %v", from, to, cycle)
		}
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

func TestSCC(t *testing.T) {
	// Graph with 2 SCCs: a->b->c->a (cycle) and c->d (bridge to singleton)
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c", "d"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 1)
	g.AddEdge("c", "a", 0, 1)
	g.AddEdge("c", "d", 0, 1)

	comps := StronglyConnectedComponents(g)
	if len(comps) != 2 {
		t.Fatalf("expected 2 SCCs, got %d: %v", len(comps), comps)
	}
	// One SCC should be {a,b,c}, the other {d}
	found3 := false
	found1 := false
	for _, c := range comps {
		if len(c) == 3 {
			found3 = true
			if c[0] != "a" || c[1] != "b" || c[2] != "c" {
				t.Fatalf("expected SCC [a b c], got %v", c)
			}
		}
		if len(c) == 1 && c[0] == "d" {
			found1 = true
		}
	}
	if !found3 || !found1 {
		t.Fatalf("unexpected SCCs: %v", comps)
	}
}

func TestSCCUndirected(t *testing.T) {
	// Undirected graph: should fall back to ConnectedComponents
	g := NewGraph[int, int](false)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddNode("c", 3)
	g.AddEdge("a", "b", 0, 1)
	// c is isolated

	comps := StronglyConnectedComponents(g)
	if len(comps) != 2 {
		t.Fatalf("expected 2 components for undirected fallback, got %d", len(comps))
	}
}

func TestSCCDAG(t *testing.T) {
	// DAG: each node is its own SCC
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 1)

	comps := StronglyConnectedComponents(g)
	if len(comps) != 3 {
		t.Fatalf("expected 3 SCCs in DAG, got %d: %v", len(comps), comps)
	}
	for _, c := range comps {
		if len(c) != 1 {
			t.Fatalf("expected singleton SCC, got %v", c)
		}
	}
}

func TestMST(t *testing.T) {
	// Triangle: a-b(1), b-c(2), a-c(3). MST should pick a-b and b-c (total=3).
	g := NewGraph[string, int](false)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddNode("c", "C")
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 2)
	g.AddEdge("a", "c", 0, 3)

	edges, total, err := MinimumSpanningTree(g)
	if err != nil {
		t.Fatal(err)
	}
	if len(edges) != 2 {
		t.Fatalf("expected 2 MST edges, got %d", len(edges))
	}
	if total != 3.0 {
		t.Fatalf("expected total weight 3, got %f", total)
	}
}

func TestMSTDirectedError(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddEdge("a", "b", 0, 1)

	_, _, err := MinimumSpanningTree(g)
	if err == nil {
		t.Fatal("expected error for directed graph")
	}
}

func TestMSTDisconnected(t *testing.T) {
	// Two components: a-b(1) and c-d(2). MST forest should have 2 edges.
	g := NewGraph[string, int](false)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddNode("c", "C")
	g.AddNode("d", "D")
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("c", "d", 0, 2)

	edges, total, err := MinimumSpanningTree(g)
	if err != nil {
		t.Fatal(err)
	}
	if len(edges) != 2 {
		t.Fatalf("expected 2 MST forest edges, got %d", len(edges))
	}
	if total != 3.0 {
		t.Fatalf("expected total weight 3, got %f", total)
	}
}

func TestAllPairsShortestPaths(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 2)
	g.AddEdge("a", "c", 0, 10) // longer direct path

	result, err := AllPairsShortestPaths(g)
	if err != nil {
		t.Fatal(err)
	}
	// a->c via b should be 3, not 10
	if result.Dist["a"]["c"] != 3 {
		t.Fatalf("expected dist a->c = 3, got %f", result.Dist["a"]["c"])
	}
	if result.Dist["a"]["b"] != 1 {
		t.Fatalf("expected dist a->b = 1, got %f", result.Dist["a"]["b"])
	}
}

func TestAllPairsReconstructPath(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 2)
	g.AddEdge("a", "c", 0, 10)

	result, err := AllPairsShortestPaths(g)
	if err != nil {
		t.Fatal(err)
	}
	path, err := ReconstructPath(result, "a", "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(path) != 3 || path[0] != "a" || path[1] != "b" || path[2] != "c" {
		t.Fatalf("expected [a b c], got %v", path)
	}
}

func TestAllPairsNoPath(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")

	result, err := AllPairsShortestPaths(g)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ReconstructPath(result, "a", "b")
	if err == nil {
		t.Fatal("expected error for no path")
	}
}

func TestCriticalPath(t *testing.T) {
	// DAG: a->b(3), a->c(2), b->d(1), c->d(4)
	// Critical path: a->c->d (length 6)
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c", "d"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 3)
	g.AddEdge("a", "c", 0, 2)
	g.AddEdge("b", "d", 0, 1)
	g.AddEdge("c", "d", 0, 4)

	result, err := CriticalPath(g)
	if err != nil {
		t.Fatal(err)
	}
	if result.Length != 6 {
		t.Fatalf("expected critical path length 6, got %f", result.Length)
	}
	// a and d must be on the critical path
	if indexOf(result.Path, "a") < 0 || indexOf(result.Path, "d") < 0 {
		t.Fatalf("expected a and d on critical path, got %v", result.Path)
	}
	// c should be on critical path (a->c->d = 6 vs a->b->d = 4)
	if indexOf(result.Path, "c") < 0 {
		t.Fatalf("expected c on critical path, got %v", result.Path)
	}
}

func TestCriticalPathLinear(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 5)
	g.AddEdge("b", "c", 0, 3)

	result, err := CriticalPath(g)
	if err != nil {
		t.Fatal(err)
	}
	if result.Length != 8 {
		t.Fatalf("expected length 8, got %f", result.Length)
	}
	// All nodes on critical path
	if len(result.Path) != 3 {
		t.Fatalf("expected 3 nodes on critical path, got %v", result.Path)
	}
}

func TestCriticalPathUndirectedError(t *testing.T) {
	g := NewGraph[int, int](false)
	_, err := CriticalPath(g)
	if err == nil {
		t.Fatal("expected error for undirected graph")
	}
}

func TestCriticalPathCycleError(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "a", 0, 1)

	_, err := CriticalPath(g)
	if err == nil {
		t.Fatal("expected error for cyclic graph")
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
