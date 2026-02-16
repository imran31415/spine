package api

import "spine"

// Upsert performs a batch of idempotent node and edge create/update operations.
func (m *Manager) Upsert(req UpsertRequest) (*UpsertResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	g, err := m.getGraph(req.Graph)
	if err != nil {
		return nil, err
	}

	res := &UpsertResult{}

	// Process nodes.
	for _, un := range req.Nodes {
		if un.ID == "" {
			continue
		}
		existing, exists := g.GetNode(un.ID)
		if exists {
			// Update: only overwrite non-empty fields.
			nd := existing.Data
			changed := false
			if un.Label != "" && un.Label != nd.Label {
				nd.Label = un.Label
				changed = true
			}
			if un.Status != "" && un.Status != nd.Status {
				nd.Status = un.Status
				changed = true
			}
			if changed {
				g.AddNode(un.ID, nd)
				res.NodesUpdated++
			}
		} else {
			g.AddNode(un.ID, NodeData{Label: un.Label, Status: un.Status})
			res.NodesCreated++
		}

		// Metadata operations.
		res.MetaKeysSet += setMeta(g.NodeMeta(un.ID), un.Meta)
		res.MetaKeysDeleted += deleteMeta(g.NodeMeta(un.ID), un.Delete)
	}

	// Process edges: auto-create endpoint nodes if missing.
	for _, ue := range req.Edges {
		if ue.From == "" || ue.To == "" {
			continue
		}
		if !g.HasNode(ue.From) {
			g.AddNode(ue.From, NodeData{})
			res.NodesCreated++
		}
		if !g.HasNode(ue.To) {
			g.AddNode(ue.To, NodeData{})
			res.NodesCreated++
		}

		if g.HasEdge(ue.From, ue.To) {
			// Update existing edge.
			e, _ := g.GetEdge(ue.From, ue.To)
			ed := e.Data
			w := e.Weight
			changed := false
			if ue.Label != "" && ue.Label != ed.Label {
				ed.Label = ue.Label
				changed = true
			}
			if ue.Weight != 0 && ue.Weight != w {
				w = ue.Weight
				changed = true
			}
			if changed {
				g.RemoveEdge(ue.From, ue.To)
				_ = g.AddEdge(ue.From, ue.To, ed, w)
				res.EdgesUpdated++
			}
		} else {
			_ = g.AddEdge(ue.From, ue.To, EdgeData{Label: ue.Label}, ue.Weight)
			res.EdgesCreated++
		}

		// Edge metadata.
		store := g.EdgeMeta(ue.From, ue.To)
		res.MetaKeysSet += setMeta(store, ue.Meta)
		res.MetaKeysDeleted += deleteMeta(store, ue.Delete)
	}

	return res, nil
}

func setMeta(store *spine.Store, meta map[string]any) int {
	if store == nil || len(meta) == 0 {
		return 0
	}
	count := 0
	for k, v := range meta {
		store.Set(k, v)
		count++
	}
	return count
}

func deleteMeta(store *spine.Store, keys []string) int {
	if store == nil || len(keys) == 0 {
		return 0
	}
	count := 0
	for _, k := range keys {
		if store.Delete(k) {
			count++
		}
	}
	return count
}
