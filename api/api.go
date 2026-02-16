package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"spine"
)

// Manager provides the high-level API for managing named spine graphs.
// All methods are safe for concurrent use.
type Manager struct {
	mu     sync.Mutex
	dir    string
	graphs map[string]*spine.Graph[NodeData, EdgeData]
}

// NewManager creates a Manager backed by the given directory.
// The directory is created if it does not exist.
func NewManager(dir string) (*Manager, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create graph dir: %w", err)
	}
	return &Manager{
		dir:    dir,
		graphs: make(map[string]*spine.Graph[NodeData, EdgeData]),
	}, nil
}

func (m *Manager) graphPath(name string) string {
	return filepath.Join(m.dir, name+".json")
}

func (m *Manager) getGraph(name string) (*spine.Graph[NodeData, EdgeData], error) {
	g, ok := m.graphs[name]
	if !ok {
		return nil, fmt.Errorf("graph %q not open", name)
	}
	return g, nil
}

// Open loads a graph from disk, or creates a new directed graph if the file
// does not exist. The graph is cached in memory for subsequent operations.
func (m *Manager) Open(name string) (*GraphInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if g, ok := m.graphs[name]; ok {
		return m.graphInfo(name, g), nil
	}

	path := m.graphPath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			g := spine.NewGraph[NodeData, EdgeData](true)
			m.graphs[name] = g
			return m.graphInfo(name, g), nil
		}
		return nil, fmt.Errorf("open %q: %w", name, err)
	}

	g, err := spine.Unmarshal[NodeData, EdgeData](data)
	if err != nil {
		return nil, fmt.Errorf("unmarshal %q: %w", name, err)
	}

	// After unmarshalling from JSON, NodeData fields (Label, Status) are
	// map-typed. We need to fix them up since json.Unmarshal fills structs
	// with matching keys but the generic N type in spine is decoded as a
	// map[string]any when the JSON has object values.
	m.fixupNodeData(g)

	m.graphs[name] = g
	return m.graphInfo(name, g), nil
}

// fixupNodeData re-parses node data that json.Unmarshal may have decoded as
// map[string]any instead of the concrete NodeData struct.
func (m *Manager) fixupNodeData(g *spine.Graph[NodeData, EdgeData]) {
	for _, n := range g.Nodes() {
		// Check if Data is actually a NodeData or was decoded as map
		switch d := any(n.Data).(type) {
		case map[string]any:
			nd := NodeData{}
			if l, ok := d["label"]; ok {
				nd.Label, _ = l.(string)
			}
			if s, ok := d["status"]; ok {
				nd.Status, _ = s.(string)
			}
			g.AddNode(n.ID, nd)
		}
	}
	// Same for edges
	for _, e := range g.Edges() {
		switch d := any(e.Data).(type) {
		case map[string]any:
			ed := EdgeData{}
			if l, ok := d["label"]; ok {
				ed.Label, _ = l.(string)
			}
			_ = g.AddEdge(e.From, e.To, ed, e.Weight)
		}
	}
}

// Save persists the named graph to disk as JSON.
func (m *Manager) Save(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	g, err := m.getGraph(name)
	if err != nil {
		return err
	}
	data, err := spine.Marshal(g, &spine.MarshalOptions{
		Graph:   true,
		Meta:    true,
		Schemas: true,
		Indent:  true,
	})
	if err != nil {
		return fmt.Errorf("marshal %q: %w", name, err)
	}
	return os.WriteFile(m.graphPath(name), data, 0o644)
}

// List returns info for every persisted graph (files on disk).
func (m *Manager) List() ([]GraphInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil, fmt.Errorf("list dir: %w", err)
	}

	var result []GraphInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".json")

		// If already loaded, use in-memory copy.
		if g, ok := m.graphs[name]; ok {
			result = append(result, *m.graphInfo(name, g))
			continue
		}

		// Peek at the file to get basic info.
		data, err := os.ReadFile(filepath.Join(m.dir, e.Name()))
		if err != nil {
			continue
		}
		var peek struct {
			Directed bool `json:"directed"`
			Graph    *struct {
				Nodes []json.RawMessage `json:"nodes"`
				Edges []json.RawMessage `json:"edges"`
			} `json:"graph"`
		}
		if json.Unmarshal(data, &peek) != nil {
			continue
		}
		info := GraphInfo{Name: name, Directed: peek.Directed}
		if peek.Graph != nil {
			info.NodeCount = len(peek.Graph.Nodes)
			info.EdgeCount = len(peek.Graph.Edges)
		}
		result = append(result, info)
	}
	return result, nil
}

// Delete removes a graph from disk and from the in-memory cache.
func (m *Manager) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.graphs, name)
	path := m.graphPath(name)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete %q: %w", name, err)
	}
	return nil
}

// Summary returns structural statistics for the named graph.
func (m *Manager) Summary(name string) (*GraphSummary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	g, err := m.getGraph(name)
	if err != nil {
		return nil, err
	}

	roots := spine.Roots(g)
	leaves := spine.Leaves(g)
	comps := spine.ConnectedComponents(g)

	rootIDs := make([]string, len(roots))
	for i, r := range roots {
		rootIDs[i] = r.ID
	}
	leafIDs := make([]string, len(leaves))
	for i, l := range leaves {
		leafIDs[i] = l.ID
	}

	statusCounts := make(map[string]int)
	for _, n := range g.Nodes() {
		s := n.Data.Status
		if s == "" {
			s = "(none)"
		}
		statusCounts[s]++
	}

	return &GraphSummary{
		GraphInfo:    *m.graphInfo(name, g),
		Roots:        rootIDs,
		Leaves:       leafIDs,
		StatusCounts: statusCounts,
		Components:   len(comps),
	}, nil
}

// Remove deletes nodes and/or edges from a graph.
func (m *Manager) Remove(req RemoveRequest) (*RemoveResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	g, err := m.getGraph(req.Graph)
	if err != nil {
		return nil, err
	}

	res := &RemoveResult{}
	for _, id := range req.Nodes {
		if g.HasNode(id) {
			g.RemoveNode(id)
			res.NodesRemoved++
		}
	}
	for _, e := range req.Edges {
		if g.HasEdge(e.From, e.To) {
			g.RemoveEdge(e.From, e.To)
			res.EdgesRemoved++
		}
	}
	return res, nil
}

func (m *Manager) graphInfo(name string, g *spine.Graph[NodeData, EdgeData]) *GraphInfo {
	return &GraphInfo{
		Name:      name,
		NodeCount: g.Order(),
		EdgeCount: g.Size(),
		Directed:  g.Directed,
	}
}

// sortedNodeIDs returns sorted node IDs from a graph.
func sortedNodeIDs(g *spine.Graph[NodeData, EdgeData]) []string {
	nodes := g.Nodes()
	ids := make([]string, len(nodes))
	for i, n := range nodes {
		ids[i] = n.ID
	}
	sort.Strings(ids)
	return ids
}
