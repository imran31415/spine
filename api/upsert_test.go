package api

import (
	"testing"
)

func TestUpsertCreateNodes(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("u")

	res, err := mgr.Upsert(UpsertRequest{
		Graph: "u",
		Nodes: []UpsertNode{
			{ID: "a", Label: "Alpha", Status: "pending"},
			{ID: "b", Label: "Beta"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.NodesCreated != 2 {
		t.Errorf("expected 2 created, got %d", res.NodesCreated)
	}
}

func TestUpsertUpdateNodes(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("u")

	mgr.Upsert(UpsertRequest{
		Graph: "u",
		Nodes: []UpsertNode{{ID: "a", Label: "Alpha", Status: "pending"}},
	})

	res, err := mgr.Upsert(UpsertRequest{
		Graph: "u",
		Nodes: []UpsertNode{{ID: "a", Status: "ready"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.NodesUpdated != 1 {
		t.Errorf("expected 1 updated, got %d", res.NodesUpdated)
	}
	if res.NodesCreated != 0 {
		t.Errorf("expected 0 created, got %d", res.NodesCreated)
	}
}

func TestUpsertIdempotent(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("u")

	req := UpsertRequest{
		Graph: "u",
		Nodes: []UpsertNode{{ID: "a", Label: "Alpha", Status: "pending"}},
	}
	mgr.Upsert(req)
	res, _ := mgr.Upsert(req)

	// Same label and status — no update counted.
	if res.NodesUpdated != 0 {
		t.Errorf("expected 0 updated on idempotent upsert, got %d", res.NodesUpdated)
	}
}

func TestUpsertEdgesAutoCreateNodes(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("u")

	res, err := mgr.Upsert(UpsertRequest{
		Graph: "u",
		Edges: []UpsertEdge{{From: "x", To: "y", Label: "dep", Weight: 1.5}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.NodesCreated != 2 {
		t.Errorf("expected 2 auto-created nodes, got %d", res.NodesCreated)
	}
	if res.EdgesCreated != 1 {
		t.Errorf("expected 1 edge created, got %d", res.EdgesCreated)
	}
}

func TestUpsertEdgeUpdate(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("u")

	mgr.Upsert(UpsertRequest{
		Graph: "u",
		Edges: []UpsertEdge{{From: "x", To: "y", Label: "old"}},
	})

	res, _ := mgr.Upsert(UpsertRequest{
		Graph: "u",
		Edges: []UpsertEdge{{From: "x", To: "y", Label: "new"}},
	})
	if res.EdgesUpdated != 1 {
		t.Errorf("expected 1 edge updated, got %d", res.EdgesUpdated)
	}
}

func TestUpsertMeta(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("u")

	res, _ := mgr.Upsert(UpsertRequest{
		Graph: "u",
		Nodes: []UpsertNode{
			{ID: "a", Meta: map[string]any{"x": 1, "y": 2}},
		},
	})
	if res.MetaKeysSet != 2 {
		t.Errorf("expected 2 meta keys set, got %d", res.MetaKeysSet)
	}

	// Delete one key.
	res2, _ := mgr.Upsert(UpsertRequest{
		Graph: "u",
		Nodes: []UpsertNode{
			{ID: "a", Delete: []string{"x"}},
		},
	})
	if res2.MetaKeysDeleted != 1 {
		t.Errorf("expected 1 meta key deleted, got %d", res2.MetaKeysDeleted)
	}
}

func TestUpsertNotOpen(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	_, err := mgr.Upsert(UpsertRequest{Graph: "nope"})
	if err == nil {
		t.Error("expected error for non-open graph")
	}
}
