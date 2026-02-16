// Package api provides a high-level, LLM-optimized public API for spine graphs.
// It uses concrete types (not generics) for easy tool-schema integration.
package api

// NodeData is the concrete node payload used by the API layer.
// Rich data lives in metadata stores.
type NodeData struct {
	Label  string `json:"label"`
	Status string `json:"status"`
}

// EdgeData is the concrete edge payload used by the API layer.
type EdgeData struct {
	Label string `json:"label"`
}

// --- Upsert ---

// UpsertRequest describes a batch of node and edge create/update operations.
type UpsertRequest struct {
	Graph string       `json:"graph"`
	Nodes []UpsertNode `json:"nodes"`
	Edges []UpsertEdge `json:"edges"`
}

// UpsertNode describes a node to create or update.
type UpsertNode struct {
	ID     string         `json:"id"`
	Label  string         `json:"label,omitempty"`
	Status string         `json:"status,omitempty"`
	Meta   map[string]any `json:"meta,omitempty"`
	Delete []string       `json:"delete,omitempty"`
}

// UpsertEdge describes an edge to create or update.
type UpsertEdge struct {
	From   string         `json:"from"`
	To     string         `json:"to"`
	Label  string         `json:"label,omitempty"`
	Weight float64        `json:"weight,omitempty"`
	Meta   map[string]any `json:"meta,omitempty"`
	Delete []string       `json:"delete,omitempty"`
}

// UpsertResult summarises the side-effects of an upsert.
type UpsertResult struct {
	NodesCreated    int `json:"nodes_created"`
	NodesUpdated    int `json:"nodes_updated"`
	EdgesCreated    int `json:"edges_created"`
	EdgesUpdated    int `json:"edges_updated"`
	MetaKeysSet     int `json:"meta_keys_set"`
	MetaKeysDeleted int `json:"meta_keys_deleted"`
}

// --- Read ---

// ReadNodesRequest describes a selective read with optional filtering and projection.
type ReadNodesRequest struct {
	Graph        string       `json:"graph"`
	IDs          []string     `json:"ids,omitempty"`
	Keys         []string     `json:"keys,omitempty"`
	Filters      []MetaFilter `json:"filters,omitempty"`
	IncludeEdges bool         `json:"include_edges,omitempty"`
	Offset       int          `json:"offset,omitempty"`
	Limit        int          `json:"limit,omitempty"`
}

// MetaFilter is a single filter predicate applied to node metadata or structural fields.
type MetaFilter struct {
	Key   string `json:"key"`
	Op    string `json:"op"`
	Value any    `json:"value,omitempty"`
}

// NodeResult is a single node in a read response.
type NodeResult struct {
	ID        string         `json:"id"`
	Label     string         `json:"label"`
	Status    string         `json:"status"`
	Meta      map[string]any `json:"meta,omitempty"`
	InDegree  int            `json:"in_degree"`
	OutDegree int            `json:"out_degree"`
}

// EdgeResult is a single edge in a read response.
type EdgeResult struct {
	From   string         `json:"from"`
	To     string         `json:"to"`
	Label  string         `json:"label"`
	Weight float64        `json:"weight,omitempty"`
	Meta   map[string]any `json:"meta,omitempty"`
}

// ReadNodesResponse is the response to a ReadNodes request.
type ReadNodesResponse struct {
	Nodes   []NodeResult `json:"nodes"`
	Edges   []EdgeResult `json:"edges,omitempty"`
	Total   int          `json:"total"`
	HasMore bool         `json:"has_more"`
}

// --- Lifecycle ---

// GraphInfo describes a graph at a glance.
type GraphInfo struct {
	Name      string `json:"name"`
	NodeCount int    `json:"node_count"`
	EdgeCount int    `json:"edge_count"`
	Directed  bool   `json:"directed"`
}

// GraphSummary extends GraphInfo with structural statistics.
type GraphSummary struct {
	GraphInfo
	Roots        []string       `json:"roots"`
	Leaves       []string       `json:"leaves"`
	StatusCounts map[string]int `json:"status_counts"`
	Components   int            `json:"components"`
}

// --- Transition ---

// TransitionRequest asks to move a node to a new status.
type TransitionRequest struct {
	Graph  string `json:"graph"`
	ID     string `json:"id"`
	Status string `json:"status"`
}

// TransitionResult describes what happened after a status transition.
type TransitionResult struct {
	ID        string   `json:"id"`
	OldStatus string   `json:"old_status"`
	NewStatus string   `json:"new_status"`
	NewlyReady []string `json:"newly_ready,omitempty"`
}

// --- Remove ---

// RemoveRequest asks to delete nodes and/or edges.
type RemoveRequest struct {
	Graph string       `json:"graph"`
	Nodes []string     `json:"nodes,omitempty"`
	Edges []RemoveEdge `json:"edges,omitempty"`
}

// RemoveEdge identifies an edge to remove.
type RemoveEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// RemoveResult summarises what was removed.
type RemoveResult struct {
	NodesRemoved int `json:"nodes_removed"`
	EdgesRemoved int `json:"edges_removed"`
}
