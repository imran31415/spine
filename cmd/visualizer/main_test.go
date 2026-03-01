package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// --- Test helpers ---

func newTestServer(t *testing.T) *server {
	t.Helper()
	return newServer(true)
}

func doJSON(t *testing.T, handler http.HandlerFunc, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest("POST", "/", &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func decodeGraphResp(t *testing.T, w *httptest.ResponseRecorder) graphResp {
	t.Helper()
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp graphResp
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	return resp
}

// --- Tests ---

func TestAddNode(t *testing.T) {
	s := newTestServer(t)
	w := doJSON(t, s.handleAddNode, addNodeReq{ID: "a", Label: "Alpha", X: 10, Y: 20})
	resp := decodeGraphResp(t, w)

	if len(resp.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(resp.Nodes))
	}
	if resp.Nodes[0].ID != "a" || resp.Nodes[0].Label != "Alpha" {
		t.Fatalf("unexpected node: %+v", resp.Nodes[0])
	}
}

func TestRemoveNode(t *testing.T) {
	s := newTestServer(t)
	doJSON(t, s.handleAddNode, addNodeReq{ID: "a", Label: "A"})
	w := doJSON(t, s.handleRemoveNode, removeNodeReq{ID: "a"})
	resp := decodeGraphResp(t, w)

	if len(resp.Nodes) != 0 {
		t.Fatalf("expected 0 nodes, got %d", len(resp.Nodes))
	}
}

func TestAddEdge(t *testing.T) {
	s := newTestServer(t)
	doJSON(t, s.handleAddNode, addNodeReq{ID: "a"})
	doJSON(t, s.handleAddNode, addNodeReq{ID: "b"})
	w := doJSON(t, s.handleAddEdge, addEdgeReq{From: "a", To: "b", Weight: 5})
	resp := decodeGraphResp(t, w)

	if len(resp.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(resp.Edges))
	}
	if resp.Edges[0].From != "a" || resp.Edges[0].To != "b" || resp.Edges[0].Weight != 5 {
		t.Fatalf("unexpected edge: %+v", resp.Edges[0])
	}
}

func TestRemoveEdge(t *testing.T) {
	s := newTestServer(t)
	doJSON(t, s.handleAddNode, addNodeReq{ID: "a"})
	doJSON(t, s.handleAddNode, addNodeReq{ID: "b"})
	doJSON(t, s.handleAddEdge, addEdgeReq{From: "a", To: "b", Weight: 1})
	w := doJSON(t, s.handleRemoveEdge, removeEdgeReq{From: "a", To: "b"})
	resp := decodeGraphResp(t, w)

	if len(resp.Edges) != 0 {
		t.Fatalf("expected 0 edges, got %d", len(resp.Edges))
	}
}

func TestAddEdgeWithLabel(t *testing.T) {
	s := newTestServer(t)
	doJSON(t, s.handleAddNode, addNodeReq{ID: "x"})
	doJSON(t, s.handleAddNode, addNodeReq{ID: "y"})
	w := doJSON(t, s.handleAddEdge, addEdgeReq{From: "x", To: "y", Label: "depends-on", Weight: 1})
	resp := decodeGraphResp(t, w)

	if len(resp.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(resp.Edges))
	}
	if resp.Edges[0].Label != "depends-on" {
		t.Fatalf("expected label 'depends-on', got %q", resp.Edges[0].Label)
	}
}

func TestSetDirected(t *testing.T) {
	s := newTestServer(t)
	// Server starts as directed.
	w := doJSON(t, s.handleGetGraph, nil)
	resp := decodeGraphResp(t, w)
	if !resp.Directed {
		t.Fatal("expected directed=true initially")
	}

	// Toggle to undirected.
	w = doJSON(t, s.handleSetDirected, map[string]bool{"directed": false})
	resp = decodeGraphResp(t, w)
	if resp.Directed {
		t.Fatal("expected directed=false after toggle")
	}
}

func TestClearGraph(t *testing.T) {
	s := newTestServer(t)
	doJSON(t, s.handleAddNode, addNodeReq{ID: "a"})
	doJSON(t, s.handleAddNode, addNodeReq{ID: "b"})
	doJSON(t, s.handleAddEdge, addEdgeReq{From: "a", To: "b", Weight: 1})
	w := doJSON(t, s.handleClear, map[string]string{})
	resp := decodeGraphResp(t, w)

	if len(resp.Nodes) != 0 || len(resp.Edges) != 0 {
		t.Fatalf("expected empty graph, got %d nodes, %d edges", len(resp.Nodes), len(resp.Edges))
	}
}

func TestExportImport(t *testing.T) {
	s := newTestServer(t)
	doJSON(t, s.handleAddNode, addNodeReq{ID: "a", Label: "Alpha", X: 100, Y: 200})
	doJSON(t, s.handleAddNode, addNodeReq{ID: "b", Label: "Beta", X: 300, Y: 400})
	doJSON(t, s.handleAddEdge, addEdgeReq{From: "a", To: "b", Weight: 3})

	// Export
	req := httptest.NewRequest("GET", "/api/graph/export", nil)
	w := httptest.NewRecorder()
	s.handleExport(w, req)
	if w.Code != 200 {
		t.Fatalf("export failed: %d %s", w.Code, w.Body.String())
	}
	exported := w.Body.Bytes()

	// Clear
	doJSON(t, s.handleClear, map[string]string{})

	// Import
	req = httptest.NewRequest("POST", "/api/graph/import", bytes.NewReader(exported))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.handleImport(w, req)
	resp := decodeGraphResp(t, w)

	if len(resp.Nodes) != 2 || len(resp.Edges) != 1 {
		t.Fatalf("expected 2 nodes, 1 edge after import, got %d nodes, %d edges", len(resp.Nodes), len(resp.Edges))
	}
}

func TestUpdateNodeStatus(t *testing.T) {
	s := newTestServer(t)
	doJSON(t, s.handleAddNode, addNodeReq{ID: "a"})
	doJSON(t, s.handleAddNode, addNodeReq{ID: "b"})
	doJSON(t, s.handleAddEdge, addEdgeReq{From: "a", To: "b", Weight: 1})

	// Set both to pending first.
	doJSON(t, s.handleUpdateNodeStatus, map[string]string{"id": "a", "status": "pending"})
	doJSON(t, s.handleUpdateNodeStatus, map[string]string{"id": "b", "status": "pending"})

	// Mark 'a' as done — 'b' should auto-promote to ready.
	w := doJSON(t, s.handleUpdateNodeStatus, map[string]string{"id": "a", "status": "done"})
	resp := decodeGraphResp(t, w)

	var aStatus, bStatus string
	for _, n := range resp.Nodes {
		if n.ID == "a" {
			aStatus = n.Status
		}
		if n.ID == "b" {
			bStatus = n.Status
		}
	}
	if aStatus != "done" {
		t.Fatalf("expected node a status=done, got %q", aStatus)
	}
	if bStatus != "ready" {
		t.Fatalf("expected node b auto-promoted to ready, got %q", bStatus)
	}
}

func TestAlgoBFS(t *testing.T) {
	s := newTestServer(t)
	doJSON(t, s.handleAddNode, addNodeReq{ID: "1"})
	doJSON(t, s.handleAddNode, addNodeReq{ID: "2"})
	doJSON(t, s.handleAddNode, addNodeReq{ID: "3"})
	doJSON(t, s.handleAddEdge, addEdgeReq{From: "1", To: "2", Weight: 1})
	doJSON(t, s.handleAddEdge, addEdgeReq{From: "2", To: "3", Weight: 1})

	req := httptest.NewRequest("POST", "/api/algo?algo=bfs", bytes.NewBufferString(`{"start":"1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleAlgo(w, req)
	resp := decodeGraphResp(t, w)

	if resp.Result == nil {
		t.Fatal("expected result, got nil")
	}
	if len(resp.Result.VisitedOrder) != 3 {
		t.Fatalf("expected 3 visited nodes, got %d", len(resp.Result.VisitedOrder))
	}
	if resp.Result.VisitedOrder[0] != "1" {
		t.Fatalf("expected BFS to start at '1', got %q", resp.Result.VisitedOrder[0])
	}
}

func TestLoadTemplate(t *testing.T) {
	s := newTestServer(t)
	if len(templates) == 0 {
		t.Skip("no templates available")
	}
	w := doJSON(t, s.handleLoadTemplate, map[string]string{"id": templates[0].ID})
	resp := decodeGraphResp(t, w)

	if len(resp.Nodes) == 0 {
		t.Fatal("expected template to produce nodes")
	}
	if len(resp.Edges) == 0 {
		t.Fatal("expected template to produce edges")
	}
}

// File-related tests (cannot be parallel because they share graphDir).

func TestListFiles(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir := graphDir
	graphDir = tmpDir
	t.Cleanup(func() { graphDir = oldDir })

	// Create some test files.
	os.WriteFile(filepath.Join(tmpDir, "graph1.json"), []byte(`{}`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "graph2.json"), []byte(`{}`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "notjson.txt"), []byte(`hello`), 0644)

	s := newTestServer(t)
	req := httptest.NewRequest("GET", "/api/files/list", nil)
	w := httptest.NewRecorder()
	s.handleListFiles(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var names []string
	json.NewDecoder(w.Body).Decode(&names)
	if len(names) != 2 {
		t.Fatalf("expected 2 json files, got %d: %v", len(names), names)
	}
}

func TestLoadFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir := graphDir
	graphDir = tmpDir
	t.Cleanup(func() { graphDir = oldDir })

	// Create a valid spine graph file using the wrapper format.
	content := `{"positions":{"x":{"x":100,"y":200}},"snapshot":{"version":1,"directed":true,"graph":{"nodes":[{"id":"x","data":{"label":"X"}}],"edges":[]}}}`
	os.WriteFile(filepath.Join(tmpDir, "test.json"), []byte(content), 0644)

	s := newTestServer(t)
	w := doJSON(t, s.handleLoadFile, map[string]string{"name": "test"})
	resp := decodeGraphResp(t, w)

	if len(resp.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(resp.Nodes))
	}
	if resp.Nodes[0].ID != "x" {
		t.Fatalf("expected node id 'x', got %q", resp.Nodes[0].ID)
	}
	// Verify positions were loaded from wrapper format.
	if resp.Nodes[0].X != 100 || resp.Nodes[0].Y != 200 {
		t.Fatalf("expected positions (100,200), got (%v,%v)", resp.Nodes[0].X, resp.Nodes[0].Y)
	}
}

func TestSaveFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir := graphDir
	graphDir = tmpDir
	t.Cleanup(func() { graphDir = oldDir })

	s := newTestServer(t)
	doJSON(t, s.handleAddNode, addNodeReq{ID: "s1", Label: "Saved", X: 50, Y: 60})

	w := doJSON(t, s.handleSaveFile, map[string]string{"name": "mysave"})
	if w.Code != 200 {
		t.Fatalf("save failed: %d %s", w.Code, w.Body.String())
	}

	// Verify file exists on disk.
	filePath := filepath.Join(tmpDir, "mysave.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("saved file not found: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("saved file is empty")
	}

	// Verify the saved file can be loaded back.
	w2 := doJSON(t, s.handleLoadFile, map[string]string{"name": "mysave"})
	resp := decodeGraphResp(t, w2)
	if len(resp.Nodes) != 1 || resp.Nodes[0].ID != "s1" {
		t.Fatalf("roundtrip failed: expected node s1, got %+v", resp.Nodes)
	}
	// Verify positions survived the roundtrip.
	if resp.Nodes[0].X != 50 || resp.Nodes[0].Y != 60 {
		t.Fatalf("expected positions (50,60) after roundtrip, got (%v,%v)", resp.Nodes[0].X, resp.Nodes[0].Y)
	}
}

func TestAlgoSCC(t *testing.T) {
	s := newTestServer(t)
	// Create a cycle: 1->2->3->1
	doJSON(t, s.handleAddNode, addNodeReq{ID: "1"})
	doJSON(t, s.handleAddNode, addNodeReq{ID: "2"})
	doJSON(t, s.handleAddNode, addNodeReq{ID: "3"})
	doJSON(t, s.handleAddEdge, addEdgeReq{From: "1", To: "2", Weight: 1})
	doJSON(t, s.handleAddEdge, addEdgeReq{From: "2", To: "3", Weight: 1})
	doJSON(t, s.handleAddEdge, addEdgeReq{From: "3", To: "1", Weight: 1})

	req := httptest.NewRequest("POST", "/api/algo?algo=scc", nil)
	w := httptest.NewRecorder()
	s.handleAlgo(w, req)
	resp := decodeGraphResp(t, w)

	if resp.Result == nil {
		t.Fatal("expected result")
	}
	if len(resp.Result.Components) != 1 {
		t.Fatalf("expected 1 SCC, got %d", len(resp.Result.Components))
	}
	if len(resp.Result.Components[0]) != 3 {
		t.Fatalf("expected SCC with 3 nodes, got %d", len(resp.Result.Components[0]))
	}
}

func TestAlgoMST(t *testing.T) {
	s := newServer(false) // undirected graph for MST
	doJSON(t, s.handleAddNode, addNodeReq{ID: "a"})
	doJSON(t, s.handleAddNode, addNodeReq{ID: "b"})
	doJSON(t, s.handleAddNode, addNodeReq{ID: "c"})
	doJSON(t, s.handleAddEdge, addEdgeReq{From: "a", To: "b", Weight: 1})
	doJSON(t, s.handleAddEdge, addEdgeReq{From: "b", To: "c", Weight: 2})
	doJSON(t, s.handleAddEdge, addEdgeReq{From: "a", To: "c", Weight: 3})

	req := httptest.NewRequest("POST", "/api/algo?algo=mst", nil)
	w := httptest.NewRecorder()
	s.handleAlgo(w, req)
	resp := decodeGraphResp(t, w)

	if resp.Result == nil {
		t.Fatal("expected result")
	}
	if resp.Result.Error != "" {
		t.Fatalf("unexpected error: %s", resp.Result.Error)
	}
	if len(resp.Result.MSTEdges) != 2 {
		t.Fatalf("expected 2 MST edges, got %d", len(resp.Result.MSTEdges))
	}
	if resp.Result.MSTWeight != 3 {
		t.Fatalf("expected MST weight 3, got %f", resp.Result.MSTWeight)
	}
}

func TestAlgoMSTDirectedError(t *testing.T) {
	s := newTestServer(t) // directed by default
	doJSON(t, s.handleAddNode, addNodeReq{ID: "a"})
	doJSON(t, s.handleAddNode, addNodeReq{ID: "b"})
	doJSON(t, s.handleAddEdge, addEdgeReq{From: "a", To: "b", Weight: 1})

	req := httptest.NewRequest("POST", "/api/algo?algo=mst", nil)
	w := httptest.NewRecorder()
	s.handleAlgo(w, req)
	resp := decodeGraphResp(t, w)

	if resp.Result == nil || resp.Result.Error == "" {
		t.Fatal("expected error for MST on directed graph")
	}
}

func TestAlgoAnalytics(t *testing.T) {
	s := newTestServer(t)
	doJSON(t, s.handleAddNode, addNodeReq{ID: "1"})
	doJSON(t, s.handleAddNode, addNodeReq{ID: "2"})
	doJSON(t, s.handleAddNode, addNodeReq{ID: "3"})
	doJSON(t, s.handleAddEdge, addEdgeReq{From: "1", To: "2", Weight: 1})
	doJSON(t, s.handleAddEdge, addEdgeReq{From: "2", To: "3", Weight: 1})

	req := httptest.NewRequest("POST", "/api/algo?algo=analytics", nil)
	w := httptest.NewRecorder()
	s.handleAlgo(w, req)
	resp := decodeGraphResp(t, w)

	if resp.Result == nil {
		t.Fatal("expected result")
	}
	if resp.Result.Analytics == nil {
		t.Fatal("expected analytics data")
	}
	// Verify analytics has the expected fields by marshalling and unmarshalling.
	b, _ := json.Marshal(resp.Result.Analytics)
	var analytics map[string]any
	json.Unmarshal(b, &analytics)
	if analytics["node_count"].(float64) != 3 {
		t.Fatalf("expected 3 nodes in analytics, got %v", analytics["node_count"])
	}
}

func TestSaveFilePathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir := graphDir
	graphDir = tmpDir
	t.Cleanup(func() { graphDir = oldDir })

	s := newTestServer(t)

	cases := []string{"../evil", "foo/bar", "..\\evil", "a..b/c", "..", "foo\\bar"}
	for _, name := range cases {
		w := doJSON(t, s.handleSaveFile, map[string]string{"name": name})
		if w.Code != 400 {
			t.Errorf("expected 400 for name %q, got %d", name, w.Code)
		}
	}

	// Also test load with traversal.
	for _, name := range cases {
		w := doJSON(t, s.handleLoadFile, map[string]string{"name": name})
		if w.Code != 400 {
			t.Errorf("load: expected 400 for name %q, got %d", name, w.Code)
		}
	}
}
