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

	if len(result.Tools) != 21 {
		t.Errorf("expected 21 tools, got %d", len(result.Tools))
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
