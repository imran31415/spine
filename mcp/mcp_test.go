package mcp

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/imran31415/spine/api"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(os.TempDir(), "spine-mcp-test-"+t.Name())
	os.RemoveAll(dir)
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	mgr, err := api.NewManager(tempDir(t))
	if err != nil {
		t.Fatal(err)
	}
	return NewServer(mgr)
}

// call sends a JSON-RPC request and returns the parsed response.
func call(t *testing.T, srv *Server, method string, params any) *Response {
	t.Helper()

	var paramsRaw json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			t.Fatal(err)
		}
		paramsRaw = b
	}

	req := Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  method,
		Params:  paramsRaw,
	}
	line, _ := json.Marshal(req)
	line = append(line, '\n')

	var in bytes.Buffer
	in.Write(line)

	var out bytes.Buffer
	if err := srv.Run(&in, &out); err != nil {
		t.Fatal(err)
	}

	var resp Response
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v\nbody: %s", err, out.String())
	}
	return &resp
}

// callTool sends a tools/call request and returns the parsed ToolCallResult.
func callTool(t *testing.T, srv *Server, name string, args any) *ToolCallResult {
	t.Helper()

	var argsRaw json.RawMessage
	if args != nil {
		b, _ := json.Marshal(args)
		argsRaw = b
	}

	resp := call(t, srv, "tools/call", ToolCallParams{Name: name, Arguments: argsRaw})
	if resp.Error != nil {
		t.Fatalf("RPC error: %s", resp.Error.Message)
	}

	// resp.Result is the ToolCallResult, but it's unmarshalled as map — re-marshal.
	b, _ := json.Marshal(resp.Result)
	var tcr ToolCallResult
	if err := json.Unmarshal(b, &tcr); err != nil {
		t.Fatal(err)
	}
	return &tcr
}

func TestInitialize(t *testing.T) {
	srv := newTestServer(t)
	resp := call(t, srv, "initialize", nil)
	if resp.Error != nil {
		t.Fatal(resp.Error.Message)
	}

	b, _ := json.Marshal(resp.Result)
	var init InitializeResult
	json.Unmarshal(b, &init)

	if init.ProtocolVersion != "2024-11-05" {
		t.Errorf("unexpected protocol version: %s", init.ProtocolVersion)
	}
	if init.ServerInfo.Name != "spine-mcp" {
		t.Errorf("unexpected server name: %s", init.ServerInfo.Name)
	}
}

func TestToolsList(t *testing.T) {
	srv := newTestServer(t)
	resp := call(t, srv, "tools/list", nil)
	if resp.Error != nil {
		t.Fatal(resp.Error.Message)
	}

	b, _ := json.Marshal(resp.Result)
	var result struct {
		Tools []ToolDefinition `json:"tools"`
	}
	json.Unmarshal(b, &result)

	if len(result.Tools) != 35 {
		t.Errorf("expected 35 tools, got %d", len(result.Tools))
	}

	names := make(map[string]bool)
	for _, td := range result.Tools {
		names[td.Name] = true
	}
	for _, expected := range []string{
		"open_graph", "save_graph", "list_graphs", "delete_graph",
		"graph_summary", "upsert", "read_nodes", "transition", "remove",
		"scc", "mst",
		"bfs", "dfs", "shortest_path", "topological_sort", "cycle_detect",
		"connected_components", "ancestors", "descendants", "roots", "leaves",
		"transitive_closure", "validate_graph", "diff_graphs",
		"degree_centrality", "betweenness_centrality", "closeness_centrality", "pagerank",
		"all_pairs_shortest_paths", "critical_path", "max_flow",
		"explain_path", "explain_component", "explain_centrality", "explain_dependency",
	} {
		if !names[expected] {
			t.Errorf("missing tool: %s", expected)
		}
	}
}

func TestNotification(t *testing.T) {
	// Notifications have no ID and should produce no response.
	srv := newTestServer(t)

	req := Request{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	line, _ := json.Marshal(req)
	line = append(line, '\n')

	var in, out bytes.Buffer
	in.Write(line)
	srv.Run(&in, &out)

	if out.Len() != 0 {
		t.Errorf("expected no response for notification, got: %s", out.String())
	}
}

func TestUnknownMethod(t *testing.T) {
	srv := newTestServer(t)
	resp := call(t, srv, "bogus/method", nil)
	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected code -32601, got %d", resp.Error.Code)
	}
}

func TestUnknownTool(t *testing.T) {
	srv := newTestServer(t)
	resp := call(t, srv, "tools/call", ToolCallParams{Name: "no_such_tool"})
	if resp.Error == nil {
		t.Fatal("expected error for unknown tool")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("expected code -32602, got %d", resp.Error.Code)
	}
}

func TestFullRoundtrip(t *testing.T) {
	srv := newTestServer(t)

	// 1. Open a graph.
	tcr := callTool(t, srv, "open_graph", map[string]any{"name": "proj"})
	if tcr.IsError {
		t.Fatalf("open_graph failed: %s", tcr.Content[0].Text)
	}

	// 2. Upsert nodes and edges.
	tcr = callTool(t, srv, "upsert", map[string]any{
		"graph": "proj",
		"nodes": []map[string]any{
			{"id": "a", "label": "Alpha", "status": "pending"},
			{"id": "b", "label": "Beta", "status": "pending"},
		},
		"edges": []map[string]any{
			{"from": "a", "to": "b", "label": "blocks"},
		},
	})
	if tcr.IsError {
		t.Fatalf("upsert failed: %s", tcr.Content[0].Text)
	}
	var upsertRes api.UpsertResult
	json.Unmarshal([]byte(tcr.Content[0].Text), &upsertRes)
	if upsertRes.NodesCreated != 2 || upsertRes.EdgesCreated != 1 {
		t.Errorf("unexpected upsert result: %+v", upsertRes)
	}

	// 3. Read nodes.
	tcr = callTool(t, srv, "read_nodes", map[string]any{
		"graph":         "proj",
		"include_edges": true,
	})
	if tcr.IsError {
		t.Fatalf("read_nodes failed: %s", tcr.Content[0].Text)
	}
	var readRes api.ReadNodesResponse
	json.Unmarshal([]byte(tcr.Content[0].Text), &readRes)
	if readRes.Total != 2 {
		t.Errorf("expected 2 nodes, got %d", readRes.Total)
	}
	if len(readRes.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(readRes.Edges))
	}

	// 4. Transition a->ready->running->done, check b becomes ready.
	callTool(t, srv, "transition", map[string]any{"graph": "proj", "id": "a", "status": "ready"})
	callTool(t, srv, "transition", map[string]any{"graph": "proj", "id": "a", "status": "running"})
	tcr = callTool(t, srv, "transition", map[string]any{"graph": "proj", "id": "a", "status": "done"})
	if tcr.IsError {
		t.Fatalf("transition failed: %s", tcr.Content[0].Text)
	}
	var transRes api.TransitionResult
	json.Unmarshal([]byte(tcr.Content[0].Text), &transRes)
	if len(transRes.NewlyReady) != 1 || transRes.NewlyReady[0] != "b" {
		t.Errorf("expected b to become ready, got: %v", transRes.NewlyReady)
	}

	// 5. Save the graph.
	tcr = callTool(t, srv, "save_graph", map[string]any{"name": "proj"})
	if tcr.IsError {
		t.Fatalf("save_graph failed: %s", tcr.Content[0].Text)
	}

	// 6. Summary.
	tcr = callTool(t, srv, "graph_summary", map[string]any{"name": "proj"})
	if tcr.IsError {
		t.Fatalf("graph_summary failed: %s", tcr.Content[0].Text)
	}
	var summary api.GraphSummary
	json.Unmarshal([]byte(tcr.Content[0].Text), &summary)
	if summary.NodeCount != 2 || summary.EdgeCount != 1 {
		t.Errorf("unexpected summary: %+v", summary)
	}

	// 7. List should show the graph.
	tcr = callTool(t, srv, "list_graphs", nil)
	if tcr.IsError {
		t.Fatalf("list_graphs failed: %s", tcr.Content[0].Text)
	}

	// 8. Remove an edge.
	tcr = callTool(t, srv, "remove", map[string]any{
		"graph": "proj",
		"edges": []map[string]any{{"from": "a", "to": "b"}},
	})
	if tcr.IsError {
		t.Fatalf("remove failed: %s", tcr.Content[0].Text)
	}
	var removeRes api.RemoveResult
	json.Unmarshal([]byte(tcr.Content[0].Text), &removeRes)
	if removeRes.EdgesRemoved != 1 {
		t.Errorf("expected 1 edge removed, got %d", removeRes.EdgesRemoved)
	}

	// 9. Delete the graph.
	tcr = callTool(t, srv, "delete_graph", map[string]any{"name": "proj"})
	if tcr.IsError {
		t.Fatalf("delete_graph failed: %s", tcr.Content[0].Text)
	}
}

func TestToolError(t *testing.T) {
	srv := newTestServer(t)

	// Calling save on a non-open graph should return isError: true.
	tcr := callTool(t, srv, "save_graph", map[string]any{"name": "nope"})
	if !tcr.IsError {
		t.Fatal("expected isError for saving non-open graph")
	}
	if len(tcr.Content) == 0 || tcr.Content[0].Text == "" {
		t.Fatal("expected error message in content")
	}
}

func TestSCC(t *testing.T) {
	srv := newTestServer(t)

	// Create a graph with a cycle: a->b->c->a and a bridge c->d.
	callTool(t, srv, "open_graph", map[string]any{"name": "scc-test"})
	callTool(t, srv, "upsert", map[string]any{
		"graph": "scc-test",
		"nodes": []map[string]any{
			{"id": "a"}, {"id": "b"}, {"id": "c"}, {"id": "d"},
		},
		"edges": []map[string]any{
			{"from": "a", "to": "b"},
			{"from": "b", "to": "c"},
			{"from": "c", "to": "a"},
			{"from": "c", "to": "d"},
		},
	})

	tcr := callTool(t, srv, "scc", map[string]any{"graph": "scc-test"})
	if tcr.IsError {
		t.Fatalf("scc failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Components [][]string `json:"components"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if len(result.Components) != 2 {
		t.Fatalf("expected 2 SCCs, got %d", len(result.Components))
	}
}

func TestMST(t *testing.T) {
	srv := newTestServer(t)

	// Create an undirected graph. The default is directed, so we need
	// to create an undirected one. Since the API always creates directed
	// graphs, MST should return an error.
	callTool(t, srv, "open_graph", map[string]any{"name": "mst-test"})
	callTool(t, srv, "upsert", map[string]any{
		"graph": "mst-test",
		"nodes": []map[string]any{
			{"id": "a"}, {"id": "b"},
		},
		"edges": []map[string]any{
			{"from": "a", "to": "b", "weight": 1.0},
		},
	})

	// Should fail because the graph is directed.
	tcr := callTool(t, srv, "mst", map[string]any{"graph": "mst-test"})
	if !tcr.IsError {
		t.Fatal("expected error for MST on directed graph")
	}
}

// setupDAG creates a graph "dag" with a->b->c for algorithm tests.
func setupDAG(t *testing.T, srv *Server) {
	t.Helper()
	callTool(t, srv, "open_graph", map[string]any{"name": "dag"})
	callTool(t, srv, "upsert", map[string]any{
		"graph": "dag",
		"nodes": []map[string]any{{"id": "a"}, {"id": "b"}, {"id": "c"}},
		"edges": []map[string]any{
			{"from": "a", "to": "b", "weight": 1.0},
			{"from": "b", "to": "c", "weight": 2.0},
		},
	})
}

func TestBFS(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "bfs", map[string]any{"graph": "dag", "start": "a"})
	if tcr.IsError {
		t.Fatalf("bfs failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Order []string `json:"order"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if len(result.Order) != 3 || result.Order[0] != "a" {
		t.Fatalf("unexpected bfs order: %v", result.Order)
	}
}

func TestDFS(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "dfs", map[string]any{"graph": "dag", "start": "a"})
	if tcr.IsError {
		t.Fatalf("dfs failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Order []string `json:"order"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if len(result.Order) != 3 || result.Order[0] != "a" {
		t.Fatalf("unexpected dfs order: %v", result.Order)
	}
}

func TestShortestPath(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "shortest_path", map[string]any{"graph": "dag", "src": "a", "dst": "c"})
	if tcr.IsError {
		t.Fatalf("shortest_path failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Path []string `json:"path"`
		Cost float64  `json:"cost"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if len(result.Path) != 3 || result.Cost != 3.0 {
		t.Fatalf("unexpected shortest_path: path=%v cost=%v", result.Path, result.Cost)
	}
}

func TestTopologicalSort(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "topological_sort", map[string]any{"graph": "dag"})
	if tcr.IsError {
		t.Fatalf("topological_sort failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Order []string `json:"order"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if len(result.Order) != 3 {
		t.Fatalf("expected 3 nodes in topo sort, got %d", len(result.Order))
	}
}

func TestCycleDetect(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	// DAG has no cycle
	tcr := callTool(t, srv, "cycle_detect", map[string]any{"graph": "dag"})
	if tcr.IsError {
		t.Fatalf("cycle_detect failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		HasCycle bool     `json:"has_cycle"`
		Cycle    []string `json:"cycle"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if result.HasCycle {
		t.Fatalf("expected no cycle, got cycle: %v", result.Cycle)
	}
}

func TestConnectedComponents(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "connected_components", map[string]any{"graph": "dag"})
	if tcr.IsError {
		t.Fatalf("connected_components failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Components [][]string `json:"components"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if len(result.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(result.Components))
	}
}

func TestAncestors(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "ancestors", map[string]any{"graph": "dag", "id": "c"})
	if tcr.IsError {
		t.Fatalf("ancestors failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Ancestors []string `json:"ancestors"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if len(result.Ancestors) != 2 {
		t.Fatalf("expected 2 ancestors, got %d: %v", len(result.Ancestors), result.Ancestors)
	}
}

func TestDescendants(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "descendants", map[string]any{"graph": "dag", "id": "a"})
	if tcr.IsError {
		t.Fatalf("descendants failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Descendants []string `json:"descendants"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if len(result.Descendants) != 2 {
		t.Fatalf("expected 2 descendants, got %d: %v", len(result.Descendants), result.Descendants)
	}
}

func TestRoots(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "roots", map[string]any{"graph": "dag"})
	if tcr.IsError {
		t.Fatalf("roots failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Roots []string `json:"roots"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if len(result.Roots) != 1 || result.Roots[0] != "a" {
		t.Fatalf("expected root [a], got %v", result.Roots)
	}
}

func TestLeaves(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "leaves", map[string]any{"graph": "dag"})
	if tcr.IsError {
		t.Fatalf("leaves failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Leaves []string `json:"leaves"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if len(result.Leaves) != 1 || result.Leaves[0] != "c" {
		t.Fatalf("expected leaf [c], got %v", result.Leaves)
	}
}

func TestEmptyGraphName(t *testing.T) {
	srv := newTestServer(t)

	// Tools that accept "name" param.
	for _, tool := range []string{"open_graph", "save_graph", "delete_graph", "graph_summary"} {
		tcr := callTool(t, srv, tool, map[string]any{"name": ""})
		if !tcr.IsError {
			t.Errorf("%s: expected error for empty name", tool)
		}
	}

	// Tools that accept "graph" param.
	for _, tool := range []string{
		"upsert", "read_nodes", "transition", "remove",
		"scc", "mst", "bfs", "dfs", "shortest_path", "topological_sort",
		"cycle_detect", "connected_components", "ancestors", "descendants",
		"roots", "leaves",
		"transitive_closure", "validate_graph",
		"degree_centrality", "betweenness_centrality", "closeness_centrality", "pagerank",
		"all_pairs_shortest_paths", "critical_path", "max_flow",
		"explain_path", "explain_component", "explain_centrality", "explain_dependency",
	} {
		tcr := callTool(t, srv, tool, map[string]any{"graph": ""})
		if !tcr.IsError {
			t.Errorf("%s: expected error for empty graph name", tool)
		}
	}

	// Whitespace-only should also be rejected.
	tcr := callTool(t, srv, "open_graph", map[string]any{"name": "   "})
	if !tcr.IsError {
		t.Error("open_graph: expected error for whitespace-only name")
	}
}

func TestDiffGraphsEmptyName(t *testing.T) {
	srv := newTestServer(t)
	tcr := callTool(t, srv, "diff_graphs", map[string]any{"graph_a": "", "graph_b": "x"})
	if !tcr.IsError {
		t.Error("diff_graphs: expected error for empty graph_a name")
	}
	tcr = callTool(t, srv, "diff_graphs", map[string]any{"graph_a": "x", "graph_b": ""})
	if !tcr.IsError {
		t.Error("diff_graphs: expected error for empty graph_b name")
	}
}

func TestTransitiveClosure(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "transitive_closure", map[string]any{"graph": "dag"})
	if tcr.IsError {
		t.Fatalf("transitive_closure failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		NodeCount int `json:"node_count"`
		EdgeCount int `json:"edge_count"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if result.NodeCount != 3 {
		t.Fatalf("expected 3 nodes, got %d", result.NodeCount)
	}
	// a->b, a->c, b->c = 3 edges
	if result.EdgeCount != 3 {
		t.Fatalf("expected 3 edges in closure, got %d", result.EdgeCount)
	}
}

func TestValidateGraph(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "validate_graph", map[string]any{"graph": "dag"})
	if tcr.IsError {
		t.Fatalf("validate_graph failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Valid bool `json:"valid"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if !result.Valid {
		t.Fatal("expected valid graph")
	}
}

func TestDiffGraphs(t *testing.T) {
	srv := newTestServer(t)

	// Create two graphs
	callTool(t, srv, "open_graph", map[string]any{"name": "g1"})
	callTool(t, srv, "upsert", map[string]any{
		"graph": "g1",
		"nodes": []map[string]any{{"id": "a"}, {"id": "b"}},
		"edges": []map[string]any{{"from": "a", "to": "b", "weight": 1.0}},
	})

	callTool(t, srv, "open_graph", map[string]any{"name": "g2"})
	callTool(t, srv, "upsert", map[string]any{
		"graph": "g2",
		"nodes": []map[string]any{{"id": "a"}, {"id": "c"}},
		"edges": []map[string]any{{"from": "a", "to": "c", "weight": 2.0}},
	})

	tcr := callTool(t, srv, "diff_graphs", map[string]any{"graph_a": "g1", "graph_b": "g2"})
	if tcr.IsError {
		t.Fatalf("diff_graphs failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		NodesAdded   []string `json:"nodes_added"`
		NodesRemoved []string `json:"nodes_removed"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if len(result.NodesAdded) != 1 || result.NodesAdded[0] != "c" {
		t.Fatalf("expected node c added, got %v", result.NodesAdded)
	}
}

func TestDegreeCentrality(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "degree_centrality", map[string]any{"graph": "dag"})
	if tcr.IsError {
		t.Fatalf("degree_centrality failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Scores map[string]float64 `json:"scores"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if len(result.Scores) != 3 {
		t.Fatalf("expected 3 scores, got %d", len(result.Scores))
	}
}

func TestBetweennessCentrality(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "betweenness_centrality", map[string]any{"graph": "dag"})
	if tcr.IsError {
		t.Fatalf("betweenness_centrality failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Scores map[string]float64 `json:"scores"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if len(result.Scores) != 3 {
		t.Fatalf("expected 3 scores, got %d", len(result.Scores))
	}
}

func TestClosenessCentrality(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "closeness_centrality", map[string]any{"graph": "dag"})
	if tcr.IsError {
		t.Fatalf("closeness_centrality failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Scores map[string]float64 `json:"scores"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if len(result.Scores) != 3 {
		t.Fatalf("expected 3 scores, got %d", len(result.Scores))
	}
}

func TestPageRank(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "pagerank", map[string]any{"graph": "dag"})
	if tcr.IsError {
		t.Fatalf("pagerank failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Scores    map[string]float64 `json:"scores"`
		Converged bool               `json:"converged"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if len(result.Scores) != 3 {
		t.Fatalf("expected 3 scores, got %d", len(result.Scores))
	}
}

func TestPageRankCustomParams(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "pagerank", map[string]any{
		"graph":     "dag",
		"damping":   0.9,
		"max_iter":  50,
		"tolerance": 0.001,
	})
	if tcr.IsError {
		t.Fatalf("pagerank with params failed: %s", tcr.Content[0].Text)
	}
}

func TestAllPairsShortestPaths(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "all_pairs_shortest_paths", map[string]any{"graph": "dag"})
	if tcr.IsError {
		t.Fatalf("all_pairs_shortest_paths failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Dist map[string]map[string]float64 `json:"dist"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	// a->c should be 3 (a->b:1 + b->c:2)
	if result.Dist["a"]["c"] != 3.0 {
		t.Fatalf("expected dist a->c = 3, got %f", result.Dist["a"]["c"])
	}
}

func TestCriticalPath(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "critical_path", map[string]any{"graph": "dag"})
	if tcr.IsError {
		t.Fatalf("critical_path failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Path   []string `json:"path"`
		Length float64  `json:"length"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if result.Length != 3.0 {
		t.Fatalf("expected length 3, got %f", result.Length)
	}
}

func TestMaxFlow(t *testing.T) {
	srv := newTestServer(t)

	callTool(t, srv, "open_graph", map[string]any{"name": "flow"})
	callTool(t, srv, "upsert", map[string]any{
		"graph": "flow",
		"nodes": []map[string]any{{"id": "s"}, {"id": "a"}, {"id": "t"}},
		"edges": []map[string]any{
			{"from": "s", "to": "a", "weight": 10.0},
			{"from": "a", "to": "t", "weight": 5.0},
		},
	})

	tcr := callTool(t, srv, "max_flow", map[string]any{
		"graph": "flow", "source": "s", "sink": "t",
	})
	if tcr.IsError {
		t.Fatalf("max_flow failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		MaxFlow float64 `json:"max_flow"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if result.MaxFlow != 5.0 {
		t.Fatalf("expected max flow 5, got %f", result.MaxFlow)
	}
}

func TestExplainPath(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "explain_path", map[string]any{
		"graph": "dag", "src": "a", "dst": "c",
	})
	if tcr.IsError {
		t.Fatalf("explain_path failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		PathLength  int    `json:"path_length"`
		Explanation string `json:"explanation"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if result.PathLength != 2 {
		t.Fatalf("expected path length 2, got %d", result.PathLength)
	}
}

func TestExplainComponent(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "explain_component", map[string]any{
		"graph": "dag", "id": "a",
	})
	if tcr.IsError {
		t.Fatalf("explain_component failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		ComponentSize int    `json:"component_size"`
		Explanation   string `json:"explanation"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if result.ComponentSize == 0 {
		t.Fatal("expected non-zero component size")
	}
}

func TestExplainCentrality(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "explain_centrality", map[string]any{
		"graph": "dag", "id": "a",
	})
	if tcr.IsError {
		t.Fatalf("explain_centrality failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		Rank       int    `json:"rank"`
		TotalNodes int    `json:"total_nodes"`
		Explanation string `json:"explanation"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if result.TotalNodes != 3 {
		t.Fatalf("expected 3 total nodes, got %d", result.TotalNodes)
	}
}

func TestExplainDependency(t *testing.T) {
	srv := newTestServer(t)
	setupDAG(t, srv)

	tcr := callTool(t, srv, "explain_dependency", map[string]any{
		"graph": "dag", "src": "a", "dst": "c",
	})
	if tcr.IsError {
		t.Fatalf("explain_dependency failed: %s", tcr.Content[0].Text)
	}
	var result struct {
		IsDirect     bool   `json:"is_direct"`
		IsTransitive bool   `json:"is_transitive"`
		Explanation  string `json:"explanation"`
	}
	json.Unmarshal([]byte(tcr.Content[0].Text), &result)
	if result.IsDirect {
		t.Fatal("expected no direct dependency from a to c")
	}
	if !result.IsTransitive {
		t.Fatal("expected transitive dependency from a to c")
	}
}

func TestParseError(t *testing.T) {
	srv := newTestServer(t)

	var in bytes.Buffer
	in.WriteString("this is not json\n")

	var out bytes.Buffer
	srv.Run(&in, &out)

	var resp Response
	json.Unmarshal(out.Bytes(), &resp)
	if resp.Error == nil || resp.Error.Code != -32700 {
		t.Errorf("expected parse error, got: %+v", resp.Error)
	}
}
