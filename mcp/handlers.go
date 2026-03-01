package mcp

import (
	"encoding/json"

	"github.com/imran31415/spine"
	"github.com/imran31415/spine/api"
)

type nameArg struct {
	Name string `json:"name"`
}

func (s *Server) handleOpenGraph(args json.RawMessage) (any, error) {
	var a nameArg
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, err
	}
	return s.mgr.Open(a.Name)
}

func (s *Server) handleSaveGraph(args json.RawMessage) (any, error) {
	var a nameArg
	if err := json.Unmarshal(args, &a); err != nil {
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
	return s.mgr.Summary(a.Name)
}

func (s *Server) handleUpsert(args json.RawMessage) (any, error) {
	var req api.UpsertRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, err
	}
	return s.mgr.Upsert(req)
}

func (s *Server) handleReadNodes(args json.RawMessage) (any, error) {
	var req api.ReadNodesRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, err
	}
	return s.mgr.ReadNodes(req)
}

func (s *Server) handleTransition(args json.RawMessage) (any, error) {
	var req api.TransitionRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, err
	}
	return s.mgr.Transition(req)
}

func (s *Server) handleRemove(args json.RawMessage) (any, error) {
	var req api.RemoveRequest
	if err := json.Unmarshal(args, &req); err != nil {
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
