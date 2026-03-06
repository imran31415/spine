package spine

import (
	"strings"
	"testing"
)

func TestExplainPath(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 2)

	result, err := ExplainPath(g, "a", "c")
	if err != nil {
		t.Fatal(err)
	}
	if result.PathLength != 2 {
		t.Fatalf("expected path length 2, got %d", result.PathLength)
	}
	if result.TotalWeight != 3.0 {
		t.Fatalf("expected total weight 3.0, got %f", result.TotalWeight)
	}
	if len(result.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(result.Steps))
	}
	if !strings.Contains(result.Explanation, "2 hop(s)") {
		t.Fatalf("explanation should mention hops: %s", result.Explanation)
	}
}

func TestExplainPathNoPath(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")

	_, err := ExplainPath(g, "a", "b")
	if err == nil {
		t.Fatal("expected error for no path")
	}
}

func TestExplainComponent(t *testing.T) {
	g := NewGraph[string, int](false)
	for _, id := range []string{"a", "b", "c", "d"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 1)
	// d is isolated

	result, err := ExplainComponent(g, "a")
	if err != nil {
		t.Fatal(err)
	}
	if result.ComponentSize != 3 {
		t.Fatalf("expected component size 3, got %d", result.ComponentSize)
	}
	if !strings.Contains(result.Explanation, "connected component") {
		t.Fatalf("should mention component type: %s", result.Explanation)
	}
}

func TestExplainComponentDirected(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 1)
	g.AddEdge("c", "a", 0, 1)

	result, err := ExplainComponent(g, "a")
	if err != nil {
		t.Fatal(err)
	}
	if result.ComponentSize != 3 {
		t.Fatalf("expected component size 3, got %d", result.ComponentSize)
	}
	if !strings.Contains(result.Explanation, "strongly connected component") {
		t.Fatalf("should mention SCC: %s", result.Explanation)
	}
}

func TestExplainComponentMissing(t *testing.T) {
	g := NewGraph[string, int](true)
	_, err := ExplainComponent(g, "x")
	if err == nil {
		t.Fatal("expected error for missing node")
	}
}

func TestExplainCentrality(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("a", "c", 0, 1)

	result, err := ExplainCentrality(g, "a")
	if err != nil {
		t.Fatal(err)
	}
	if result.Rank != 1 {
		t.Fatalf("expected rank 1, got %d", result.Rank)
	}
	if result.TotalNodes != 3 {
		t.Fatalf("expected 3 total nodes, got %d", result.TotalNodes)
	}
	if !strings.Contains(result.Explanation, "#1") {
		t.Fatalf("should mention rank: %s", result.Explanation)
	}
}

func TestExplainCentralityMissing(t *testing.T) {
	g := NewGraph[string, int](true)
	_, err := ExplainCentrality(g, "x")
	if err == nil {
		t.Fatal("expected error for missing node")
	}
}

func TestExplainDependencyDirect(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddEdge("a", "b", 0, 1)

	result, err := ExplainDependency(g, "a", "b")
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsDirect {
		t.Fatal("expected direct dependency")
	}
	if !result.IsTransitive {
		t.Fatal("expected transitive (path exists)")
	}
}

func TestExplainDependencyTransitive(t *testing.T) {
	g := NewGraph[string, int](true)
	for _, id := range []string{"a", "b", "c"} {
		g.AddNode(id, id)
	}
	g.AddEdge("a", "b", 0, 1)
	g.AddEdge("b", "c", 0, 1)

	result, err := ExplainDependency(g, "a", "c")
	if err != nil {
		t.Fatal(err)
	}
	if result.IsDirect {
		t.Fatal("expected no direct dependency")
	}
	if !result.IsTransitive {
		t.Fatal("expected transitive dependency")
	}
	if len(result.Paths) == 0 {
		t.Fatal("expected at least one path")
	}
}

func TestExplainDependencyNone(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")

	result, err := ExplainDependency(g, "a", "b")
	if err != nil {
		t.Fatal(err)
	}
	if result.IsDirect {
		t.Fatal("expected no direct dependency")
	}
	if result.IsTransitive {
		t.Fatal("expected no transitive dependency")
	}
	if !strings.Contains(result.Explanation, "no dependency") {
		t.Fatalf("should mention no dependency: %s", result.Explanation)
	}
}

func TestExplainDependencyMissing(t *testing.T) {
	g := NewGraph[string, int](true)
	g.AddNode("a", "A")

	_, err := ExplainDependency(g, "a", "x")
	if err == nil {
		t.Fatal("expected error for missing target")
	}
	_, err = ExplainDependency(g, "x", "a")
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}
