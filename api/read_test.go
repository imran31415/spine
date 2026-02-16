package api

import (
	"testing"
)

func setupReadGraph(t *testing.T) *Manager {
	t.Helper()
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("r")
	mgr.Upsert(UpsertRequest{
		Graph: "r",
		Nodes: []UpsertNode{
			{ID: "a", Label: "Alpha", Status: "done", Meta: map[string]any{"priority": float64(10), "tag": "core"}},
			{ID: "b", Label: "Beta", Status: "pending", Meta: map[string]any{"priority": float64(5)}},
			{ID: "c", Label: "Charlie", Status: "running", Meta: map[string]any{"priority": float64(8), "tag": "ui"}},
			{ID: "d", Label: "Delta", Status: "done", Meta: map[string]any{"priority": float64(3)}},
		},
		Edges: []UpsertEdge{
			{From: "a", To: "b", Label: "dep"},
			{From: "a", To: "c", Label: "dep"},
			{From: "c", To: "d", Label: "dep"},
		},
	})
	return mgr
}

func TestReadByID(t *testing.T) {
	mgr := setupReadGraph(t)
	resp, err := mgr.ReadNodes(ReadNodesRequest{
		Graph: "r",
		IDs:   []string{"a", "c"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(resp.Nodes))
	}
	if resp.Nodes[0].ID != "a" || resp.Nodes[1].ID != "c" {
		t.Errorf("unexpected node IDs: %s, %s", resp.Nodes[0].ID, resp.Nodes[1].ID)
	}
}

func TestReadByFilter(t *testing.T) {
	mgr := setupReadGraph(t)
	resp, err := mgr.ReadNodes(ReadNodesRequest{
		Graph:   "r",
		Filters: []MetaFilter{{Key: "status", Op: "eq", Value: "done"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Nodes) != 2 {
		t.Fatalf("expected 2 done nodes, got %d", len(resp.Nodes))
	}
}

func TestReadKeyProjection(t *testing.T) {
	mgr := setupReadGraph(t)
	resp, err := mgr.ReadNodes(ReadNodesRequest{
		Graph: "r",
		IDs:   []string{"a"},
		Keys:  []string{"priority"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Nodes) != 1 {
		t.Fatal("expected 1 node")
	}
	meta := resp.Nodes[0].Meta
	if meta == nil {
		t.Fatal("expected meta")
	}
	if _, ok := meta["priority"]; !ok {
		t.Error("expected priority key")
	}
	if _, ok := meta["tag"]; ok {
		t.Error("expected tag key to be excluded by projection")
	}
}

func TestReadPagination(t *testing.T) {
	mgr := setupReadGraph(t)
	resp, err := mgr.ReadNodes(ReadNodesRequest{
		Graph:  "r",
		Offset: 0,
		Limit:  2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(resp.Nodes))
	}
	if resp.Total != 4 {
		t.Errorf("expected total 4, got %d", resp.Total)
	}
	if !resp.HasMore {
		t.Error("expected HasMore=true")
	}

	// Second page.
	resp2, _ := mgr.ReadNodes(ReadNodesRequest{
		Graph:  "r",
		Offset: 2,
		Limit:  2,
	})
	if len(resp2.Nodes) != 2 {
		t.Fatalf("expected 2 nodes on page 2, got %d", len(resp2.Nodes))
	}
	if resp2.HasMore {
		t.Error("expected HasMore=false on last page")
	}
}

func TestReadIncludeEdges(t *testing.T) {
	mgr := setupReadGraph(t)
	resp, err := mgr.ReadNodes(ReadNodesRequest{
		Graph:        "r",
		IDs:          []string{"a", "b", "c"},
		IncludeEdges: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Edges) != 2 {
		t.Errorf("expected 2 edges among a,b,c; got %d", len(resp.Edges))
	}
}

func TestReadDegrees(t *testing.T) {
	mgr := setupReadGraph(t)
	resp, _ := mgr.ReadNodes(ReadNodesRequest{
		Graph: "r",
		IDs:   []string{"a"},
	})
	if len(resp.Nodes) != 1 {
		t.Fatal("expected 1 node")
	}
	n := resp.Nodes[0]
	if n.InDegree != 0 {
		t.Errorf("expected in_degree=0, got %d", n.InDegree)
	}
	if n.OutDegree != 2 {
		t.Errorf("expected out_degree=2, got %d", n.OutDegree)
	}
}

func TestReadEmptyResult(t *testing.T) {
	mgr := setupReadGraph(t)
	resp, err := mgr.ReadNodes(ReadNodesRequest{
		Graph:   "r",
		Filters: []MetaFilter{{Key: "status", Op: "eq", Value: "skipped"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(resp.Nodes))
	}
}

func TestReadNotOpen(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	_, err := mgr.ReadNodes(ReadNodesRequest{Graph: "nope"})
	if err == nil {
		t.Error("expected error for non-open graph")
	}
}
