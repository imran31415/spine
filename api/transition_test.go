package api

import (
	"testing"
)

func TestTransitionBasic(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("t")
	mgr.Upsert(UpsertRequest{
		Graph: "t",
		Nodes: []UpsertNode{{ID: "a", Status: "pending"}},
	})

	res, err := mgr.Transition(TransitionRequest{Graph: "t", ID: "a", Status: "ready"})
	if err != nil {
		t.Fatal(err)
	}
	if res.OldStatus != "pending" || res.NewStatus != "ready" {
		t.Errorf("unexpected: %+v", res)
	}
}

func TestTransitionInvalid(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("t")
	mgr.Upsert(UpsertRequest{
		Graph: "t",
		Nodes: []UpsertNode{{ID: "a", Status: "pending"}},
	})

	_, err := mgr.Transition(TransitionRequest{Graph: "t", ID: "a", Status: "done"})
	if err == nil {
		t.Error("expected error for invalid transition pending->done")
	}
}

func TestTransitionAutoReady(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("t")
	mgr.Upsert(UpsertRequest{
		Graph: "t",
		Nodes: []UpsertNode{
			{ID: "a", Status: "running"},
			{ID: "b", Status: "running"},
			{ID: "c", Status: "pending"}, // depends on both a and b
		},
		Edges: []UpsertEdge{
			{From: "a", To: "c"},
			{From: "b", To: "c"},
		},
	})

	// Complete a — c should NOT be ready yet (b still running).
	res1, err := mgr.Transition(TransitionRequest{Graph: "t", ID: "a", Status: "done"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res1.NewlyReady) != 0 {
		t.Errorf("expected no newly ready, got %v", res1.NewlyReady)
	}

	// Complete b — now c should auto-promote.
	res2, err := mgr.Transition(TransitionRequest{Graph: "t", ID: "b", Status: "done"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res2.NewlyReady) != 1 || res2.NewlyReady[0] != "c" {
		t.Errorf("expected c to be newly ready, got %v", res2.NewlyReady)
	}
}

func TestTransitionAutoReadySkipsNonPending(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("t")
	mgr.Upsert(UpsertRequest{
		Graph: "t",
		Nodes: []UpsertNode{
			{ID: "a", Status: "running"},
			{ID: "b", Status: "ready"}, // already ready, should not appear in NewlyReady
		},
		Edges: []UpsertEdge{{From: "a", To: "b"}},
	})

	res, _ := mgr.Transition(TransitionRequest{Graph: "t", ID: "a", Status: "done"})
	if len(res.NewlyReady) != 0 {
		t.Errorf("expected no newly ready (b already ready), got %v", res.NewlyReady)
	}
}

func TestTransitionFullLifecycle(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("t")
	mgr.Upsert(UpsertRequest{
		Graph: "t",
		Nodes: []UpsertNode{{ID: "a"}}, // empty status
	})

	transitions := []string{"pending", "ready", "running", "done"}
	for _, s := range transitions {
		_, err := mgr.Transition(TransitionRequest{Graph: "t", ID: "a", Status: s})
		if err != nil {
			t.Fatalf("transition to %q failed: %v", s, err)
		}
	}
}

func TestTransitionRetry(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("t")
	mgr.Upsert(UpsertRequest{
		Graph: "t",
		Nodes: []UpsertNode{{ID: "a", Status: "running"}},
	})

	// running -> failed
	mgr.Transition(TransitionRequest{Graph: "t", ID: "a", Status: "failed"})
	// failed -> pending (retry)
	res, err := mgr.Transition(TransitionRequest{Graph: "t", ID: "a", Status: "pending"})
	if err != nil {
		t.Fatal(err)
	}
	if res.NewStatus != "pending" {
		t.Errorf("expected pending after retry, got %s", res.NewStatus)
	}
}

func TestTransitionMissingNode(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	mgr.Open("t")
	_, err := mgr.Transition(TransitionRequest{Graph: "t", ID: "nope", Status: "ready"})
	if err == nil {
		t.Error("expected error for missing node")
	}
}

func TestTransitionNotOpen(t *testing.T) {
	dir := tempDir(t)
	mgr, _ := NewManager(dir)
	_, err := mgr.Transition(TransitionRequest{Graph: "nope", ID: "a", Status: "ready"})
	if err == nil {
		t.Error("expected error for non-open graph")
	}
}
