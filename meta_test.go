package spine

import (
	"testing"
)

func TestNodeMeta(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "A")

	store := g.NodeMeta("a")
	if store == nil {
		t.Fatal("expected non-nil store for existing node")
	}
	store.Set("color", "red")
	v, ok := store.Get("color")
	if !ok || v != "red" {
		t.Fatalf("expected red, got %v", v)
	}

	// Same store returned on second call
	store2 := g.NodeMeta("a")
	if store2 != store {
		t.Fatal("expected same store instance")
	}
}

func TestEdgeMeta(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddEdge("a", "b", "ab", 1.0)

	store := g.EdgeMeta("a", "b")
	if store == nil {
		t.Fatal("expected non-nil store for existing edge")
	}
	store.Set("label", "connects")
	v, ok := store.Get("label")
	if !ok || v != "connects" {
		t.Fatalf("expected connects, got %v", v)
	}

	// Same store returned on second call
	store2 := g.EdgeMeta("a", "b")
	if store2 != store {
		t.Fatal("expected same store instance")
	}
}

func TestNodeMetaNonexistent(t *testing.T) {
	g := NewGraph[int, int](true)
	if g.NodeMeta("missing") != nil {
		t.Fatal("expected nil for nonexistent node")
	}
}

func TestEdgeMetaNonexistent(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("a", 1)
	g.AddNode("b", 2)
	if g.EdgeMeta("a", "b") != nil {
		t.Fatal("expected nil for nonexistent edge")
	}
}

func TestNodeMetaLazy(t *testing.T) {
	g := NewGraph[int, int](true)
	g.AddNode("a", 1)

	// Before calling NodeMeta, internal map should have no entry
	if _, ok := g.nodeMeta["a"]; ok {
		t.Fatal("store should not be created until NodeMeta is called")
	}

	g.NodeMeta("a")
	if _, ok := g.nodeMeta["a"]; !ok {
		t.Fatal("store should be created after NodeMeta is called")
	}
}

func TestRemoveNodeCleansMetadata(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddEdge("a", "b", "ab", 1.0)

	g.NodeMeta("a").Set("key", "val")
	g.EdgeMeta("a", "b").Set("key", "val")

	g.RemoveNode("a")

	if g.NodeMeta("a") != nil {
		t.Fatal("node metadata should be gone after RemoveNode")
	}
	// Edge metadata should also be cleaned up
	if _, ok := g.edgeMeta["a"]; ok {
		t.Fatal("edge metadata for removed node should be cleaned up")
	}
}

func TestRemoveEdgeCleansMetadata(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddEdge("a", "b", "ab", 1.0)

	g.EdgeMeta("a", "b").Set("key", "val")
	g.RemoveEdge("a", "b")

	// Edge metadata should be gone
	if m, ok := g.edgeMeta["a"]; ok {
		if _, ok := m["b"]; ok {
			t.Fatal("edge metadata should be gone after RemoveEdge")
		}
	}
}

func TestCopyPreservesMetadata(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddEdge("a", "b", "ab", 1.0)

	g.NodeMeta("a").Set("key", "original")
	g.EdgeMeta("a", "b").Set("label", "edge-original")

	c := g.Copy()

	// Verify metadata exists in copy
	v, ok := c.NodeMeta("a").Get("key")
	if !ok || v != "original" {
		t.Fatalf("expected original, got %v", v)
	}
	v, ok = c.EdgeMeta("a", "b").Get("label")
	if !ok || v != "edge-original" {
		t.Fatalf("expected edge-original, got %v", v)
	}

	// Modify original, verify copy is independent
	g.NodeMeta("a").Set("key", "modified")
	v, _ = c.NodeMeta("a").Get("key")
	if v != "original" {
		t.Fatal("copy should be independent from original")
	}
}

func TestUndirectedEdgeMetaSymmetry(t *testing.T) {
	g := NewGraph[string, string](false)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddEdge("a", "b", "ab", 1.0)

	store1 := g.EdgeMeta("a", "b")
	store2 := g.EdgeMeta("b", "a")
	if store1 != store2 {
		t.Fatal("undirected edge metadata should be the same store in both directions")
	}

	store1.Set("weight", 42)
	v, ok := store2.Get("weight")
	if !ok || v != 42 {
		t.Fatal("changes in one direction should be visible in the other")
	}
}

func TestSubgraphPreservesMetadata(t *testing.T) {
	g := NewGraph[string, string](true)
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddNode("c", "C")
	g.AddEdge("a", "b", "ab", 1.0)
	g.AddEdge("b", "c", "bc", 1.0)

	g.NodeMeta("a").Set("role", "start")
	g.NodeMeta("b").Set("role", "middle")
	g.EdgeMeta("a", "b").Set("type", "link")

	sub := Subgraph(g, []string{"a", "b"})

	// Node metadata should be preserved
	v, ok := sub.NodeMeta("a").Get("role")
	if !ok || v != "start" {
		t.Fatalf("expected start, got %v", v)
	}
	v, ok = sub.NodeMeta("b").Get("role")
	if !ok || v != "middle" {
		t.Fatalf("expected middle, got %v", v)
	}

	// Edge metadata should be preserved
	v, ok = sub.EdgeMeta("a", "b").Get("type")
	if !ok || v != "link" {
		t.Fatalf("expected link, got %v", v)
	}

	// Subgraph should be independent
	g.NodeMeta("a").Set("role", "changed")
	v, _ = sub.NodeMeta("a").Get("role")
	if v != "start" {
		t.Fatal("subgraph metadata should be independent from parent")
	}
}
