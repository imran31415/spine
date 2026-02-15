package spine

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Snapshot is the top-level serialized form of a graph.
type Snapshot[N, E any] struct {
	Version  int              `json:"version"`
	Directed bool             `json:"directed"`
	Graph    *GraphData[N, E] `json:"graph,omitempty"`
	Meta     *MetaData        `json:"metadata,omitempty"`
}

// GraphData holds the graph topology (nodes + edges).
type GraphData[N, E any] struct {
	Nodes []NodeData[N] `json:"nodes"`
	Edges []EdgeData[E] `json:"edges"`
}

// NodeData is the serialized form of a node.
type NodeData[N any] struct {
	ID   string `json:"id"`
	Data N      `json:"data"`
}

// EdgeData is the serialized form of an edge.
type EdgeData[E any] struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Data   E       `json:"data"`
	Weight float64 `json:"weight"`
}

// MetaData holds all metadata for nodes and edges.
type MetaData struct {
	Nodes []NodeMetaData `json:"nodes"`
	Edges []EdgeMetaData `json:"edges"`
}

// NodeMetaData is the serialized metadata for a single node.
type NodeMetaData struct {
	ID      string         `json:"id"`
	Entries map[string]any `json:"entries"`
	Schema  Schema         `json:"schema,omitempty"`
}

// EdgeMetaData is the serialized metadata for a single edge.
type EdgeMetaData struct {
	From    string         `json:"from"`
	To      string         `json:"to"`
	Entries map[string]any `json:"entries"`
	Schema  Schema         `json:"schema,omitempty"`
}

// MarshalOptions controls what gets serialized.
type MarshalOptions struct {
	NodeIDs []string // if non-nil, only include these nodes + edges between them
	Graph   bool     // include graph topology section
	Meta    bool     // include metadata section
	Schemas bool     // include schema definitions in metadata
	Indent  bool     // pretty-print JSON
}

// Marshal serializes a graph to JSON. If opts is nil, everything is included with pretty-printing.
func Marshal[N, E any](g *Graph[N, E], opts *MarshalOptions) ([]byte, error) {
	if opts == nil {
		opts = &MarshalOptions{Graph: true, Meta: true, Schemas: true, Indent: true}
	}

	target := g
	if opts.NodeIDs != nil {
		target = Subgraph(g, opts.NodeIDs)
	}

	snap := Snapshot[N, E]{
		Version:  1,
		Directed: target.Directed,
	}

	if opts.Graph {
		gd := &GraphData[N, E]{
			Nodes: make([]NodeData[N], 0),
			Edges: make([]EdgeData[E], 0),
		}
		for _, n := range target.Nodes() {
			gd.Nodes = append(gd.Nodes, NodeData[N]{ID: n.ID, Data: n.Data})
		}
		edges := target.Edges()
		// Normalize undirected edges so From < To for deterministic output.
		if !target.Directed {
			for i := range edges {
				if edges[i].From > edges[i].To {
					edges[i].From, edges[i].To = edges[i].To, edges[i].From
				}
			}
		}
		sort.Slice(edges, func(i, j int) bool {
			if edges[i].From != edges[j].From {
				return edges[i].From < edges[j].From
			}
			return edges[i].To < edges[j].To
		})
		for _, e := range edges {
			gd.Edges = append(gd.Edges, EdgeData[E]{From: e.From, To: e.To, Data: e.Data, Weight: e.Weight})
		}
		snap.Graph = gd
	}

	if opts.Meta {
		md := &MetaData{
			Nodes: make([]NodeMetaData, 0),
			Edges: make([]EdgeMetaData, 0),
		}

		// Node metadata — iterate Nodes() which returns sorted by ID.
		for _, n := range target.Nodes() {
			store, ok := target.nodeMeta[n.ID]
			if !ok || store.Len() == 0 {
				continue
			}
			nm := NodeMetaData{
				ID:      n.ID,
				Entries: make(map[string]any, store.Len()),
			}
			for k, v := range store.entries {
				nm.Entries[k] = v
			}
			if opts.Schemas {
				if schema := store.GetSchema(); schema != nil {
					nm.Schema = schema
				}
			}
			md.Nodes = append(md.Nodes, nm)
		}

		// Edge metadata — collect and sort by (from, to).
		type edgeKey struct{ from, to string }
		var keys []edgeKey
		for from, m := range target.edgeMeta {
			for to, store := range m {
				if store.Len() > 0 {
					keys = append(keys, edgeKey{from, to})
				}
			}
		}
		sort.Slice(keys, func(i, j int) bool {
			if keys[i].from != keys[j].from {
				return keys[i].from < keys[j].from
			}
			return keys[i].to < keys[j].to
		})
		for _, k := range keys {
			store := target.edgeMeta[k.from][k.to]
			em := EdgeMetaData{
				From:    k.from,
				To:      k.to,
				Entries: make(map[string]any, store.Len()),
			}
			for key, val := range store.entries {
				em.Entries[key] = val
			}
			if opts.Schemas {
				if schema := store.GetSchema(); schema != nil {
					em.Schema = schema
				}
			}
			md.Edges = append(md.Edges, em)
		}

		snap.Meta = md
	}

	if opts.Indent {
		return json.MarshalIndent(snap, "", "  ")
	}
	return json.Marshal(snap)
}

// Unmarshal deserializes JSON into a new graph. Both graph topology and metadata
// sections are applied when present.
func Unmarshal[N, E any](data []byte) (*Graph[N, E], error) {
	var snap Snapshot[N, E]
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	if snap.Version != 1 {
		return nil, fmt.Errorf("unsupported version: %d", snap.Version)
	}

	g := NewGraph[N, E](snap.Directed)

	if snap.Graph != nil {
		for _, n := range snap.Graph.Nodes {
			g.AddNode(n.ID, n.Data)
		}
		for _, e := range snap.Graph.Edges {
			if err := g.AddEdge(e.From, e.To, e.Data, e.Weight); err != nil {
				return nil, fmt.Errorf("unmarshal edge %s->%s: %w", e.From, e.To, err)
			}
		}
	}

	if snap.Meta != nil {
		for _, nm := range snap.Meta.Nodes {
			if !g.HasNode(nm.ID) {
				continue
			}
			store := g.NodeMeta(nm.ID)
			for k, v := range nm.Entries {
				store.Set(k, v)
			}
			if nm.Schema != nil {
				store.SetSchema(nm.Schema)
			}
		}
		for _, em := range snap.Meta.Edges {
			if !g.HasEdge(em.From, em.To) {
				continue
			}
			store := g.EdgeMeta(em.From, em.To)
			for k, v := range em.Entries {
				store.Set(k, v)
			}
			if em.Schema != nil {
				store.SetSchema(em.Schema)
			}
		}
	}

	return g, nil
}

// ApplyMeta reads the metadata section from JSON and applies it to an existing graph.
// Nodes and edges not present in the graph are silently skipped.
func ApplyMeta[N, E any](data []byte, g *Graph[N, E]) error {
	var raw struct {
		Meta *MetaData `json:"metadata"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("apply meta: %w", err)
	}
	if raw.Meta == nil {
		return nil
	}

	for _, nm := range raw.Meta.Nodes {
		if !g.HasNode(nm.ID) {
			continue
		}
		store := g.NodeMeta(nm.ID)
		for k, v := range nm.Entries {
			store.Set(k, v)
		}
		if nm.Schema != nil {
			store.SetSchema(nm.Schema)
		}
	}
	for _, em := range raw.Meta.Edges {
		if !g.HasEdge(em.From, em.To) {
			continue
		}
		store := g.EdgeMeta(em.From, em.To)
		for k, v := range em.Entries {
			store.Set(k, v)
		}
		if em.Schema != nil {
			store.SetSchema(em.Schema)
		}
	}

	return nil
}
