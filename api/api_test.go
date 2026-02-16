package api

import (
	"os"
	"path/filepath"
	"testing"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(os.TempDir(), "spine-api-test-"+t.Name())
	os.RemoveAll(dir)
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestNewManager(t *testing.T) {
	dir := tempDir(t)
	mgr, err := NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatal("expected directory to be created")
	}
}

func TestOpenCreatesSave(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)

	info, err := mgr.Open("test")
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "test" || info.NodeCount != 0 || !info.Directed {
		t.Errorf("unexpected info: %+v", info)
	}

	// Opening again returns same graph.
	info2, err := mgr.Open("test")
	if err != nil {
		t.Fatal(err)
	}
	if info2.Name != "test" {
		t.Error("expected same graph on re-open")
	}
}

func TestSaveAndReload(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("proj")

	// Add some data.
	mgr.Upsert(UpsertRequest{
		Graph: "proj",
		Nodes: []UpsertNode{
			{ID: "a", Label: "Alpha", Status: "pending"},
			{ID: "b", Label: "Beta", Status: "ready"},
		},
		Edges: []UpsertEdge{{From: "a", To: "b", Label: "dep"}},
	})

	if err := mgr.Save("proj"); err != nil {
		t.Fatal(err)
	}

	// Check file exists.
	if _, err := os.Stat(filepath.Join(dir, "proj.json")); err != nil {
		t.Fatal("expected file on disk")
	}

	// Load in a fresh manager.
	mgr2, _ := NewManager(dir)
	info, err := mgr2.Open("proj")
	if err != nil {
		t.Fatal(err)
	}
	if info.NodeCount != 2 || info.EdgeCount != 1 {
		t.Errorf("unexpected reloaded info: %+v", info)
	}
}

func TestList(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("aaa")
	mgr.Open("bbb")
	mgr.Save("aaa")
	mgr.Save("bbb")

	list, err := mgr.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 graphs, got %d", len(list))
	}
}

func TestDelete(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("del")
	mgr.Save("del")

	if err := mgr.Delete("del"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "del.json")); !os.IsNotExist(err) {
		t.Error("expected file to be removed")
	}

	list, _ := mgr.List()
	if len(list) != 0 {
		t.Error("expected empty list after delete")
	}
}

func TestSummary(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("sum")
	mgr.Upsert(UpsertRequest{
		Graph: "sum",
		Nodes: []UpsertNode{
			{ID: "root", Label: "Root", Status: "done"},
			{ID: "mid", Label: "Mid", Status: "running"},
			{ID: "leaf", Label: "Leaf", Status: "pending"},
		},
		Edges: []UpsertEdge{
			{From: "root", To: "mid"},
			{From: "mid", To: "leaf"},
		},
	})

	sum, err := mgr.Summary("sum")
	if err != nil {
		t.Fatal(err)
	}
	if sum.NodeCount != 3 || sum.EdgeCount != 2 {
		t.Errorf("unexpected counts: nodes=%d edges=%d", sum.NodeCount, sum.EdgeCount)
	}
	if len(sum.Roots) != 1 || sum.Roots[0] != "root" {
		t.Errorf("unexpected roots: %v", sum.Roots)
	}
	if len(sum.Leaves) != 1 || sum.Leaves[0] != "leaf" {
		t.Errorf("unexpected leaves: %v", sum.Leaves)
	}
	if sum.StatusCounts["done"] != 1 || sum.StatusCounts["running"] != 1 {
		t.Errorf("unexpected status counts: %v", sum.StatusCounts)
	}
	if sum.Components != 1 {
		t.Errorf("expected 1 component, got %d", sum.Components)
	}
}

func TestRemove(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("rem")
	mgr.Upsert(UpsertRequest{
		Graph: "rem",
		Nodes: []UpsertNode{{ID: "x"}, {ID: "y"}, {ID: "z"}},
		Edges: []UpsertEdge{{From: "x", To: "y"}, {From: "y", To: "z"}},
	})

	res, err := mgr.Remove(RemoveRequest{
		Graph: "rem",
		Nodes: []string{"z"},
		Edges: []RemoveEdge{{From: "x", To: "y"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.NodesRemoved != 1 || res.EdgesRemoved != 1 {
		t.Errorf("unexpected remove result: %+v", res)
	}
}

func TestSaveNotOpen(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	if err := mgr.Save("nope"); err == nil {
		t.Error("expected error saving non-open graph")
	}
}
