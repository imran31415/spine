package spine

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
)

// benchGraph creates a directed graph with n nodes and n*edgeFactor random edges.
func benchGraph(n, edgeFactor int) *Graph[string, string] {
	g := NewGraph[string, string](true)
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("n%d", i)
		g.AddNode(id, id)
	}
	for i := 0; i < n*edgeFactor; i++ {
		from := fmt.Sprintf("n%d", rng.Intn(n))
		to := fmt.Sprintf("n%d", rng.Intn(n))
		if from != to {
			g.AddEdge(from, to, "", float64(rng.Intn(100)+1))
		}
	}
	return g
}

func BenchmarkAddNode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		g := NewGraph[string, string](true)
		for j := 0; j < 1000; j++ {
			g.AddNode(fmt.Sprintf("n%d", j), "data")
		}
	}
}

func BenchmarkAddEdge(b *testing.B) {
	g := NewGraph[string, string](true)
	for j := 0; j < 1000; j++ {
		g.AddNode(fmt.Sprintf("n%d", j), "data")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g2 := g.Copy()
		rng := rand.New(rand.NewSource(int64(i)))
		for j := 0; j < 1000; j++ {
			from := fmt.Sprintf("n%d", rng.Intn(1000))
			to := fmt.Sprintf("n%d", rng.Intn(1000))
			if from != to {
				g2.AddEdge(from, to, "", 1.0)
			}
		}
	}
}

func BenchmarkBFS(b *testing.B) {
	g := benchGraph(1000, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BFS(g, "n0", nil)
	}
}

func BenchmarkDFS(b *testing.B) {
	g := benchGraph(1000, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DFS(g, "n0", nil)
	}
}

func BenchmarkShortestPath(b *testing.B) {
	g := benchGraph(1000, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ShortestPath(g, "n0", "n999")
	}
}

func BenchmarkTopologicalSort(b *testing.B) {
	// Build a DAG: edges only go from lower to higher indices.
	g := NewGraph[string, string](true)
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 1000; i++ {
		g.AddNode(fmt.Sprintf("n%d", i), "")
	}
	for i := 0; i < 3000; i++ {
		a := rng.Intn(999)
		b2 := a + 1 + rng.Intn(1000-a-1)
		g.AddEdge(fmt.Sprintf("n%d", a), fmt.Sprintf("n%d", b2), "", 1.0)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TopologicalSort(g)
	}
}

func BenchmarkConnectedComponents(b *testing.B) {
	g := benchGraph(1000, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConnectedComponents(g)
	}
}

func BenchmarkSCC(b *testing.B) {
	g := benchGraph(1000, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StronglyConnectedComponents(g)
	}
}

func BenchmarkGraphAnalytics(b *testing.B) {
	g := benchGraph(1000, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GraphAnalytics(g)
	}
}

func BenchmarkNeighbors(b *testing.B) {
	g := benchGraph(1000, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000; j++ {
			g.Neighbors(fmt.Sprintf("n%d", j))
		}
	}
}

func BenchmarkMarshal(b *testing.B) {
	g := benchGraph(1000, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(struct {
			Nodes []Node[string]
			Edges []Edge[string]
		}{g.Nodes(), g.Edges()})
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	g := benchGraph(1000, 3)
	data, _ := json.Marshal(struct {
		Nodes []Node[string] `json:"nodes"`
		Edges []Edge[string] `json:"edges"`
	}{g.Nodes(), g.Edges()})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var v struct {
			Nodes []Node[string] `json:"nodes"`
			Edges []Edge[string] `json:"edges"`
		}
		json.Unmarshal(data, &v)
	}
}
