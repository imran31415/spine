package api

import (
	"sort"

	"spine"
)

const defaultLimit = 100

// ReadNodes performs a selective read with optional ID lookup, filtering,
// key projection, and pagination.
func (m *Manager) ReadNodes(req ReadNodesRequest) (*ReadNodesResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	g, err := m.getGraph(req.Graph)
	if err != nil {
		return nil, err
	}

	// Collect candidate node IDs.
	var ids []string
	if len(req.IDs) > 0 {
		// Explicit ID lookup — only return nodes that exist.
		for _, id := range req.IDs {
			if g.HasNode(id) {
				ids = append(ids, id)
			}
		}
	} else {
		ids = sortedNodeIDs(g)
	}

	// Apply filters.
	var matched []string
	for _, id := range ids {
		if matchesFilters(g, id, req.Filters) {
			matched = append(matched, id)
		}
	}
	sort.Strings(matched)

	total := len(matched)

	// Pagination.
	limit := req.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	offset := req.Offset
	if offset > len(matched) {
		offset = len(matched)
	}
	end := offset + limit
	if end > len(matched) {
		end = len(matched)
	}
	page := matched[offset:end]

	// Build node results.
	keySet := makeKeySet(req.Keys)
	nodes := make([]NodeResult, 0, len(page))
	for _, id := range page {
		n, _ := g.GetNode(id)
		nr := NodeResult{
			ID:        id,
			Label:     n.Data.Label,
			Status:    n.Data.Status,
			InDegree:  len(g.InEdges(id)),
			OutDegree: len(g.OutEdges(id)),
		}
		nr.Meta = projectMeta(g.NodeMeta(id), keySet)
		nodes = append(nodes, nr)
	}

	resp := &ReadNodesResponse{
		Nodes:   nodes,
		Total:   total,
		HasMore: end < total,
	}

	// Optionally include edges between matched nodes.
	if req.IncludeEdges && len(page) > 0 {
		sub := spine.Subgraph(g, page)
		for _, e := range sub.Edges() {
			er := EdgeResult{
				From:   e.From,
				To:     e.To,
				Label:  e.Data.Label,
				Weight: e.Weight,
			}
			edgeMeta := g.EdgeMeta(e.From, e.To)
			er.Meta = projectMeta(edgeMeta, nil)
			resp.Edges = append(resp.Edges, er)
		}
	}

	return resp, nil
}

// makeKeySet builds a set from a slice. nil means "all keys".
func makeKeySet(keys []string) map[string]bool {
	if len(keys) == 0 {
		return nil
	}
	s := make(map[string]bool, len(keys))
	for _, k := range keys {
		s[k] = true
	}
	return s
}

// projectMeta returns metadata entries filtered by keySet.
// If keySet is nil, all entries are returned. Returns nil if store is nil or empty.
func projectMeta(store *spine.Store, keySet map[string]bool) map[string]any {
	if store == nil || store.Len() == 0 {
		return nil
	}
	result := make(map[string]any)
	store.Range(func(key string, value any) bool {
		if keySet == nil || keySet[key] {
			result[key] = value
		}
		return true
	})
	if len(result) == 0 {
		return nil
	}
	return result
}
