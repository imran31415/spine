package mcp

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/imran31415/spine"
	"github.com/imran31415/spine/api"
)

var errEmptyName = errors.New("graph name must not be empty")

type nameArg struct {
	Name string `json:"name"`
}

func requireName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errEmptyName
	}
	return nil
}

func (s *Server) handleOpenGraph(args json.RawMessage) (any, error) {
	var a struct {
		Name     string `json:"name"`
		Directed *bool  `json:"directed,omitempty"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Name); err != nil {
		return nil, err
	}
	directed := true
	if a.Directed != nil {
		directed = *a.Directed
	}
	return s.mgr.OpenWithDirected(a.Name, directed)
}

func (s *Server) handleSaveGraph(args json.RawMessage) (any, error) {
	var a nameArg
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Name); err != nil {
		return nil, err
	}
	if err := s.mgr.Save(a.Name); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

func (s *Server) handleListGraphs(args json.RawMessage) (any, error) {
	return s.mgr.List()
}

func (s *Server) handleDeleteGraph(args json.RawMessage) (any, error) {
	var a nameArg
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Name); err != nil {
		return nil, err
	}
	if err := s.mgr.Delete(a.Name); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

func (s *Server) handleGraphSummary(args json.RawMessage) (any, error) {
	var a nameArg
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Name); err != nil {
		return nil, err
	}
	return s.mgr.Summary(a.Name)
}

func (s *Server) handleUpsert(args json.RawMessage) (any, error) {
	var req api.UpsertRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, err
	}
	if err := requireName(req.Graph); err != nil {
		return nil, err
	}
	return s.mgr.Upsert(req)
}

func (s *Server) handleReadNodes(args json.RawMessage) (any, error) {
	var req api.ReadNodesRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, err
	}
	if err := requireName(req.Graph); err != nil {
		return nil, err
	}
	return s.mgr.ReadNodes(req)
}

func (s *Server) handleTransition(args json.RawMessage) (any, error) {
	var req api.TransitionRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, err
	}
	if err := requireName(req.Graph); err != nil {
		return nil, err
	}
	return s.mgr.Transition(req)
}

func (s *Server) handleRemove(args json.RawMessage) (any, error) {
	var req api.RemoveRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, err
	}
	if err := requireName(req.Graph); err != nil {
		return nil, err
	}
	return s.mgr.Remove(req)
}

func (s *Server) handleSCC(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	comps := spine.StronglyConnectedComponents(g)
	return map[string]any{"components": comps}, nil
}

func (s *Server) handleMST(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	edges, totalWeight, err := spine.MinimumSpanningTree(g)
	if err != nil {
		return nil, err
	}
	type edgeResult struct {
		From   string  `json:"from"`
		To     string  `json:"to"`
		Weight float64 `json:"weight"`
	}
	result := make([]edgeResult, len(edges))
	for i, e := range edges {
		result[i] = edgeResult{From: e.From, To: e.To, Weight: e.Weight}
	}
	return map[string]any{"edges": result, "total_weight": totalWeight}, nil
}

func (s *Server) handleBFS(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
		Start string `json:"start"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	order := spine.BFS(g, a.Start, nil)
	return map[string]any{"order": order}, nil
}

func (s *Server) handleDFS(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
		Start string `json:"start"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	order := spine.DFS(g, a.Start, nil)
	return map[string]any{"order": order}, nil
}

func (s *Server) handleShortestPath(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
		Src   string `json:"src"`
		Dst   string `json:"dst"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	path, cost, err := spine.ShortestPath(g, a.Src, a.Dst)
	if err != nil {
		return nil, err
	}
	return map[string]any{"path": path, "cost": cost}, nil
}

func (s *Server) handleTopologicalSort(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	order, err := spine.TopologicalSort(g)
	if err != nil {
		return nil, err
	}
	return map[string]any{"order": order}, nil
}

func (s *Server) handleCycleDetect(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	hasCycle, cycle := spine.CycleDetect(g)
	return map[string]any{"has_cycle": hasCycle, "cycle": cycle}, nil
}

func (s *Server) handleConnectedComponents(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	comps := spine.ConnectedComponents(g)
	return map[string]any{"components": comps}, nil
}

func (s *Server) handleAncestors(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
		ID    string `json:"id"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	anc := spine.Ancestors(g, a.ID)
	return map[string]any{"ancestors": anc}, nil
}

func (s *Server) handleDescendants(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
		ID    string `json:"id"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	desc := spine.Descendants(g, a.ID)
	return map[string]any{"descendants": desc}, nil
}

func (s *Server) handleRoots(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	roots := spine.Roots(g)
	ids := make([]string, len(roots))
	for i, r := range roots {
		ids[i] = r.ID
	}
	return map[string]any{"roots": ids}, nil
}

func (s *Server) handleLeaves(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	leaves := spine.Leaves(g)
	ids := make([]string, len(leaves))
	for i, l := range leaves {
		ids[i] = l.ID
	}
	return map[string]any{"leaves": ids}, nil
}

func (s *Server) handleTransitiveClosure(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	tc, err := spine.TransitiveClosure(g)
	if err != nil {
		return nil, err
	}
	type edgeResult struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	edges := tc.Edges()
	edgeResults := make([]edgeResult, len(edges))
	for i, e := range edges {
		edgeResults[i] = edgeResult{From: e.From, To: e.To}
	}
	return map[string]any{
		"node_count": tc.Order(),
		"edge_count": tc.Size(),
		"edges":      edgeResults,
	}, nil
}

func (s *Server) handleValidateGraph(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	result := spine.Validate(g)
	return result, nil
}

func (s *Server) handleDiffGraphs(args json.RawMessage) (any, error) {
	var a struct {
		GraphA string `json:"graph_a"`
		GraphB string `json:"graph_b"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.GraphA); err != nil {
		return nil, err
	}
	if err := requireName(a.GraphB); err != nil {
		return nil, err
	}
	ga, err := s.mgr.OpenGraph(a.GraphA)
	if err != nil {
		return nil, err
	}
	gb, err := s.mgr.OpenGraph(a.GraphB)
	if err != nil {
		return nil, err
	}
	result, err := spine.Diff(ga, gb)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Server) handleDegreeCentrality(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	result := spine.DegreeCentrality(g)
	return result, nil
}

func (s *Server) handleBetweennessCentrality(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	result := spine.BetweennessCentrality(g)
	return result, nil
}

func (s *Server) handleClosenessCentrality(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	result := spine.ClosenessCentrality(g)
	return result, nil
}

func (s *Server) handlePageRank(args json.RawMessage) (any, error) {
	var a struct {
		Graph     string   `json:"graph"`
		Damping   *float64 `json:"damping,omitempty"`
		MaxIter   *int     `json:"max_iter,omitempty"`
		Tolerance *float64 `json:"tolerance,omitempty"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	damping := 0.85
	if a.Damping != nil {
		damping = *a.Damping
	}
	maxIter := 100
	if a.MaxIter != nil {
		maxIter = *a.MaxIter
	}
	tol := 1e-6
	if a.Tolerance != nil {
		tol = *a.Tolerance
	}
	result := spine.PageRank(g, damping, maxIter, tol)
	return result, nil
}

func (s *Server) handleAllPairsShortestPaths(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	result, err := spine.AllPairsShortestPaths(g)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Server) handleCriticalPath(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	result, err := spine.CriticalPath(g)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Server) handleMaxFlow(args json.RawMessage) (any, error) {
	var a struct {
		Graph  string `json:"graph"`
		Source string `json:"source"`
		Sink   string `json:"sink"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	result, err := spine.MaxFlow(g, a.Source, a.Sink)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Server) handleExplainPath(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
		Src   string `json:"src"`
		Dst   string `json:"dst"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	result, err := spine.ExplainPath(g, a.Src, a.Dst)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Server) handleExplainComponent(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
		ID    string `json:"id"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	result, err := spine.ExplainComponent(g, a.ID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Server) handleExplainCentrality(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
		ID    string `json:"id"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	result, err := spine.ExplainCentrality(g, a.ID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Server) handleExplainDependency(args json.RawMessage) (any, error) {
	var a struct {
		Graph string `json:"graph"`
		Src   string `json:"src"`
		Dst   string `json:"dst"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	if err := requireName(a.Graph); err != nil {
		return nil, err
	}
	g, err := s.mgr.OpenGraph(a.Graph)
	if err != nil {
		return nil, err
	}
	result, err := spine.ExplainDependency(g, a.Src, a.Dst)
	if err != nil {
		return nil, err
	}
	return result, nil
}
