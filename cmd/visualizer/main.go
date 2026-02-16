package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"spine"
)

//go:embed static/index.html
var static embed.FS

// Position stores x,y for the visualizer layout.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NodeData is the data stored in each graph node.
type NodeData struct {
	Label  string `json:"label"`
	Status string `json:"status,omitempty"`
}

// EdgeData is the data stored in each graph edge.
type EdgeData struct {
	Label string `json:"label"`
}

type server struct {
	mu        sync.Mutex
	graph     *spine.Graph[NodeData, EdgeData]
	positions map[string]Position
}

// --- API request/response types ---

type addNodeReq struct {
	ID    string  `json:"id"`
	Label string  `json:"label"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}

type removeNodeReq struct {
	ID string `json:"id"`
}

type addEdgeReq struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Label  string  `json:"label"`
	Weight float64 `json:"weight"`
}

type removeEdgeReq struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type updatePosReq struct {
	ID string  `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
}

type algoReq struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type graphResp struct {
	Directed bool            `json:"directed"`
	Nodes    []nodeResp      `json:"nodes"`
	Edges    []edgeResp      `json:"edges"`
	Result   *algoResultResp `json:"result,omitempty"`
}

type nodeResp struct {
	ID        string  `json:"id"`
	Label     string  `json:"label"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Status    string  `json:"status,omitempty"`
	MetaCount int     `json:"metaCount,omitempty"`
}

type edgeResp struct {
	From      string  `json:"from"`
	To        string  `json:"to"`
	Label     string  `json:"label"`
	Weight    float64 `json:"weight"`
	MetaCount int     `json:"metaCount,omitempty"`
}

type algoResultResp struct {
	Algorithm      string      `json:"algorithm"`
	VisitedOrder   []string    `json:"visitedOrder,omitempty"`
	Path           []string    `json:"path,omitempty"`
	Cost           float64     `json:"cost,omitempty"`
	HasCycle       bool        `json:"hasCycle,omitempty"`
	Cycle          []string    `json:"cycle,omitempty"`
	Components     [][]string  `json:"components,omitempty"`
	Roots          []string    `json:"roots,omitempty"`
	Leaves         []string    `json:"leaves,omitempty"`
	Ancestors      []string    `json:"ancestors,omitempty"`
	Descendants    []string    `json:"descendants,omitempty"`
	HighlightNodes []string    `json:"highlightNodes,omitempty"`
	HighlightEdges [][2]string `json:"highlightEdges,omitempty"`
	Error          string      `json:"error,omitempty"`
}

// Metadata API types

type metaResp struct {
	Items   []metaEntry `json:"items"`
	Total   int         `json:"total"`
	Offset  int         `json:"offset"`
	HasMore bool        `json:"hasMore"`
}

type metaEntry struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

func newServer(directed bool) *server {
	return &server{
		graph:     spine.NewGraph[NodeData, EdgeData](directed),
		positions: make(map[string]Position),
	}
}

func (s *server) buildGraphResp(result *algoResultResp) graphResp {
	nodes := s.graph.Nodes()
	nr := make([]nodeResp, len(nodes))
	for i, n := range nodes {
		pos := s.positions[n.ID]
		nr[i] = nodeResp{
			ID:        n.ID,
			Label:     n.Data.Label,
			X:         pos.X,
			Y:         pos.Y,
			Status:    n.Data.Status,
			MetaCount: s.graph.NodeMetaCount(n.ID),
		}
	}
	edges := s.graph.Edges()
	er := make([]edgeResp, len(edges))
	for i, e := range edges {
		er[i] = edgeResp{
			From:      e.From,
			To:        e.To,
			Label:     e.Data.Label,
			Weight:    e.Weight,
			MetaCount: s.graph.EdgeMetaCount(e.From, e.To),
		}
	}
	return graphResp{
		Directed: s.graph.Directed,
		Nodes:    nr,
		Edges:    er,
		Result:   result,
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (s *server) handleGetGraph(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, s.buildGraphResp(nil))
}

func (s *server) handleAddNode(w http.ResponseWriter, r *http.Request) {
	var req addNodeReq
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if req.Label == "" {
		req.Label = req.ID
	}
	s.graph.AddNode(req.ID, NodeData{Label: req.Label})
	s.positions[req.ID] = Position{X: req.X, Y: req.Y}
	writeJSON(w, s.buildGraphResp(nil))
}

func (s *server) handleRemoveNode(w http.ResponseWriter, r *http.Request) {
	var req removeNodeReq
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.graph.RemoveNode(req.ID)
	delete(s.positions, req.ID)
	writeJSON(w, s.buildGraphResp(nil))
}

func (s *server) handleAddEdge(w http.ResponseWriter, r *http.Request) {
	var req addEdgeReq
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.graph.AddEdge(req.From, req.To, EdgeData{Label: req.Label}, req.Weight); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, s.buildGraphResp(nil))
}

func (s *server) handleRemoveEdge(w http.ResponseWriter, r *http.Request) {
	var req removeEdgeReq
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.graph.RemoveEdge(req.From, req.To)
	writeJSON(w, s.buildGraphResp(nil))
}

func (s *server) handleUpdatePos(w http.ResponseWriter, r *http.Request) {
	var req updatePosReq
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.positions[req.ID] = Position{X: req.X, Y: req.Y}
	w.WriteHeader(204)
}

func (s *server) handleSetDirected(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Directed bool `json:"directed"`
	}
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// Rebuild the graph with the new mode, keeping all nodes, edges, and metadata.
	old := s.graph
	s.graph = spine.NewGraph[NodeData, EdgeData](req.Directed)
	for _, n := range old.Nodes() {
		s.graph.AddNode(n.ID, n.Data)
	}
	for _, e := range old.Edges() {
		s.graph.AddEdge(e.From, e.To, e.Data, e.Weight)
	}
	// Preserve metadata
	for _, n := range old.Nodes() {
		if old.NodeMetaCount(n.ID) > 0 {
			src := old.NodeMeta(n.ID)
			dst := s.graph.NodeMeta(n.ID)
			src.Range(func(k string, v any) bool {
				dst.Set(k, v)
				return true
			})
		}
	}
	for _, e := range old.Edges() {
		if old.EdgeMetaCount(e.From, e.To) > 0 {
			src := old.EdgeMeta(e.From, e.To)
			dst := s.graph.EdgeMeta(e.From, e.To)
			src.Range(func(k string, v any) bool {
				dst.Set(k, v)
				return true
			})
		}
	}
	writeJSON(w, s.buildGraphResp(nil))
}

func (s *server) handleAlgo(w http.ResponseWriter, r *http.Request) {
	algo := r.URL.Query().Get("algo")
	var req algoReq
	readJSON(r, &req)

	s.mu.Lock()
	defer s.mu.Unlock()

	result := &algoResultResp{Algorithm: algo}

	switch algo {
	case "bfs":
		if req.Start == "" {
			result.Error = "start node required"
			break
		}
		order := spine.BFS(s.graph, req.Start, nil)
		result.VisitedOrder = order
		result.HighlightNodes = order
		result.HighlightEdges = pathToEdges(order)

	case "dfs":
		if req.Start == "" {
			result.Error = "start node required"
			break
		}
		order := spine.DFS(s.graph, req.Start, nil)
		result.VisitedOrder = order
		result.HighlightNodes = order
		result.HighlightEdges = pathToEdges(order)

	case "shortest-path":
		if req.Start == "" || req.End == "" {
			result.Error = "start and end nodes required"
			break
		}
		path, cost, err := spine.ShortestPath(s.graph, req.Start, req.End)
		if err != nil {
			result.Error = err.Error()
			break
		}
		result.Path = path
		result.Cost = cost
		result.HighlightNodes = path
		result.HighlightEdges = pathToEdges(path)

	case "topo-sort":
		order, err := spine.TopologicalSort(s.graph)
		if err != nil {
			result.Error = err.Error()
			break
		}
		result.VisitedOrder = order
		result.HighlightNodes = order

	case "cycle-detect":
		hasCycle, cycle := spine.CycleDetect(s.graph)
		result.HasCycle = hasCycle
		result.Cycle = cycle
		if hasCycle {
			result.HighlightNodes = cycle
			edges := pathToEdges(cycle)
			if len(cycle) >= 2 {
				edges = append(edges, [2]string{cycle[len(cycle)-1], cycle[0]})
			}
			result.HighlightEdges = edges
		}

	case "components":
		comps := spine.ConnectedComponents(s.graph)
		result.Components = comps

	case "roots":
		roots := spine.Roots(s.graph)
		ids := make([]string, len(roots))
		for i, r := range roots {
			ids[i] = r.ID
		}
		result.Roots = ids
		result.HighlightNodes = ids

	case "leaves":
		leaves := spine.Leaves(s.graph)
		ids := make([]string, len(leaves))
		for i, l := range leaves {
			ids[i] = l.ID
		}
		result.Leaves = ids
		result.HighlightNodes = ids

	case "ancestors":
		if req.Start == "" {
			result.Error = "start node required"
			break
		}
		anc := spine.Ancestors(s.graph, req.Start)
		result.Ancestors = anc
		result.HighlightNodes = append(anc, req.Start)

	case "descendants":
		if req.Start == "" {
			result.Error = "start node required"
			break
		}
		desc := spine.Descendants(s.graph, req.Start)
		result.Descendants = desc
		result.HighlightNodes = append(desc, req.Start)

	default:
		result.Error = fmt.Sprintf("unknown algorithm: %s", algo)
	}

	writeJSON(w, s.buildGraphResp(result))
}

func pathToEdges(path []string) [][2]string {
	if len(path) < 2 {
		return nil
	}
	edges := make([][2]string, 0, len(path)-1)
	for i := 0; i < len(path)-1; i++ {
		edges = append(edges, [2]string{path[i], path[i+1]})
	}
	return edges
}

func (s *server) handleGetTemplates(w http.ResponseWriter, r *http.Request) {
	summaries := make([]templateSummary, len(templates))
	for i, t := range templates {
		summaries[i] = t.summary()
	}
	writeJSON(w, summaries)
}

func (s *server) handleLoadTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	var tmpl *Template
	for i := range templates {
		if templates[i].ID == req.ID {
			tmpl = &templates[i]
			break
		}
	}
	if tmpl == nil {
		http.Error(w, "template not found", 404)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.graph = spine.NewGraph[NodeData, EdgeData](tmpl.Directed)
	s.positions = make(map[string]Position)
	for _, n := range tmpl.Nodes {
		s.graph.AddNode(n.ID, NodeData{Label: n.Label, Status: n.Status})
		s.positions[n.ID] = Position{X: n.X, Y: n.Y}
		if len(n.Meta) > 0 {
			store := s.graph.NodeMeta(n.ID)
			for k, v := range n.Meta {
				store.Set(k, v)
			}
		}
	}
	for _, e := range tmpl.Edges {
		s.graph.AddEdge(e.From, e.To, EdgeData{Label: e.Label}, e.Weight)
	}
	s.computeReady()
	writeJSON(w, s.buildGraphResp(nil))
}

var validStatuses = map[string]bool{
	"pending": true, "ready": true, "running": true,
	"done": true, "failed": true, "skipped": true,
}

// computeReady auto-promotes pending nodes to ready when all in-edge sources are done.
func (s *server) computeReady() {
	for _, n := range s.graph.Nodes() {
		if n.Data.Status != "pending" {
			continue
		}
		inEdges := s.graph.InEdges(n.ID)
		allDone := true
		for _, e := range inEdges {
			src, ok := s.graph.GetNode(e.From)
			if !ok || src.Data.Status != "done" {
				allDone = false
				break
			}
		}
		if allDone && len(inEdges) > 0 {
			s.graph.AddNode(n.ID, NodeData{Label: n.Data.Label, Status: "ready"})
		}
	}
}

func (s *server) handleUpdateNodeStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if !validStatuses[req.Status] {
		http.Error(w, fmt.Sprintf("invalid status: %s", req.Status), 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.graph.GetNode(req.ID)
	if !ok {
		http.Error(w, "node not found", 404)
		return
	}
	s.graph.AddNode(req.ID, NodeData{Label: n.Data.Label, Status: req.Status})
	s.computeReady()
	writeJSON(w, s.buildGraphResp(nil))
}

type planTask struct {
	ID           string   `json:"id"`
	Label        string   `json:"label"`
	Status       string   `json:"status"`
	Dependencies []string `json:"dependencies"`
	X            float64  `json:"x"`
	Y            float64  `json:"y"`
}

func (s *server) handleLoadPlan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Tasks []planTask `json:"tasks"`
	}
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.graph = spine.NewGraph[NodeData, EdgeData](true)
	s.positions = make(map[string]Position)

	// Build dependency map for auto-layout
	depMap := make(map[string][]string)
	taskMap := make(map[string]*planTask)
	for i := range req.Tasks {
		t := &req.Tasks[i]
		taskMap[t.ID] = t
		depMap[t.ID] = t.Dependencies
	}

	// Auto-layout: compute topological depth for each node
	needsLayout := false
	for _, t := range req.Tasks {
		if t.X == 0 && t.Y == 0 {
			needsLayout = true
			break
		}
	}

	if needsLayout {
		depth := make(map[string]int)
		var computeDepth func(id string) int
		computeDepth = func(id string) int {
			if d, ok := depth[id]; ok {
				return d
			}
			depth[id] = 0
			maxDep := 0
			for _, dep := range depMap[id] {
				d := computeDepth(dep) + 1
				if d > maxDep {
					maxDep = d
				}
			}
			depth[id] = maxDep
			return maxDep
		}
		for _, t := range req.Tasks {
			computeDepth(t.ID)
		}

		// Group by depth level
		levels := make(map[int][]string)
		maxLevel := 0
		for _, t := range req.Tasks {
			d := depth[t.ID]
			levels[d] = append(levels[d], t.ID)
			if d > maxLevel {
				maxLevel = d
			}
		}

		// Assign positions
		for level := 0; level <= maxLevel; level++ {
			ids := levels[level]
			for i, id := range ids {
				t := taskMap[id]
				t.X = 150 + float64(i)*200
				t.Y = 80 + float64(level)*150
			}
		}
	}

	// Create nodes
	for _, t := range req.Tasks {
		status := t.Status
		if status == "" {
			status = "pending"
		}
		label := t.Label
		if label == "" {
			label = t.ID
		}
		s.graph.AddNode(t.ID, NodeData{Label: label, Status: status})
		s.positions[t.ID] = Position{X: t.X, Y: t.Y}
	}

	// Create dependency edges (dep → task)
	for _, t := range req.Tasks {
		for _, dep := range t.Dependencies {
			s.graph.AddEdge(dep, t.ID, EdgeData{}, 1)
		}
	}

	s.computeReady()
	writeJSON(w, s.buildGraphResp(nil))
}

// ---- Directory Upload handler ----

type dirEntry struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	Content string `json:"content,omitempty"`
}

type loadDirReq struct {
	RootName string     `json:"rootName"`
	Entries  []dirEntry `json:"entries"`
}

func (s *server) handleLoadDirectory(w http.ResponseWriter, r *http.Request) {
	var req loadDirReq
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if req.RootName == "" {
		req.RootName = "root"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.graph = spine.NewGraph[NodeData, EdgeData](true)
	s.positions = make(map[string]Position)

	// Add root node.
	rootID := req.RootName
	s.graph.AddNode(rootID, NodeData{Label: rootID + "/"})
	s.graph.NodeMeta(rootID).Set("type", "directory")

	// Collect unique directories and files from entries.
	dirs := make(map[string]bool)
	dirs[rootID] = true

	for _, e := range req.Entries {
		nodeID := e.Path
		if e.IsDir {
			dirs[nodeID] = true
			s.graph.AddNode(nodeID, NodeData{Label: e.Name + "/"})
			meta := s.graph.NodeMeta(nodeID)
			meta.Set("type", "directory")
		} else {
			s.graph.AddNode(nodeID, NodeData{Label: e.Name})
			meta := s.graph.NodeMeta(nodeID)
			meta.Set("type", "file")
			meta.Set("size", e.Size)
			if e.Content != "" {
				meta.Set("content", e.Content)
			}
		}

		// Build parent edge. Parent is filepath.Dir(path) or rootName if top-level.
		parent := filepath.Dir(nodeID)
		if parent == "." || parent == "" {
			parent = rootID
		}
		// Ensure parent directory nodes exist in the graph.
		if !s.graph.HasNode(parent) {
			s.graph.AddNode(parent, NodeData{Label: filepath.Base(parent) + "/"})
			s.graph.NodeMeta(parent).Set("type", "directory")
			dirs[parent] = true
		}
		s.graph.AddEdge(parent, nodeID, EdgeData{}, 1)
	}

	// Compute tree layout: group nodes by depth (number of '/' separators).
	type nodeDepth struct {
		id    string
		depth int
	}
	var allNodes []nodeDepth
	for _, n := range s.graph.Nodes() {
		d := 0
		if n.ID != rootID {
			for _, c := range n.ID {
				if c == '/' {
					d++
				}
			}
			// Files/dirs at top level under root still have depth 1.
			d++ // +1 because root is depth 0
		}
		allNodes = append(allNodes, nodeDepth{id: n.ID, depth: d})
	}

	// Group by depth level.
	levels := make(map[int][]string)
	maxLevel := 0
	for _, nd := range allNodes {
		levels[nd.depth] = append(levels[nd.depth], nd.id)
		if nd.depth > maxLevel {
			maxLevel = nd.depth
		}
	}

	// Assign positions: center each level horizontally.
	for level := 0; level <= maxLevel; level++ {
		ids := levels[level]
		totalWidth := float64(len(ids)-1) * 200
		startX := 400 - totalWidth/2
		for i, id := range ids {
			s.positions[id] = Position{
				X: startX + float64(i)*200,
				Y: 80 + float64(level)*150,
			}
		}
	}

	writeJSON(w, s.buildGraphResp(nil))
}

func (s *server) handleClear(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	directed := s.graph.Directed
	s.graph = spine.NewGraph[NodeData, EdgeData](directed)
	s.positions = make(map[string]Position)
	writeJSON(w, s.buildGraphResp(nil))
}

// ---- Export / Import handlers ----

func (s *server) handleExport(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapBytes, err := spine.Marshal(s.graph, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var snapshot any
	if err := json.Unmarshal(snapBytes, &snapshot); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	wrapper := map[string]any{
		"positions": s.positions,
		"snapshot":  snapshot,
	}

	out, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (s *server) handleImport(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Positions map[string]Position `json:"positions"`
		Snapshot  json.RawMessage     `json:"snapshot"`
	}
	if err := readJSON(r, &payload); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	g, err := spine.Unmarshal[NodeData, EdgeData](payload.Snapshot)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.graph = g
	s.positions = make(map[string]Position)

	// Apply imported positions; grid-layout any nodes that are missing.
	col := 0
	for _, n := range s.graph.Nodes() {
		if pos, ok := payload.Positions[n.ID]; ok {
			s.positions[n.ID] = pos
		} else {
			s.positions[n.ID] = Position{
				X: 150 + float64(col%5)*200,
				Y: 80 + float64(col/5)*150,
			}
			col++
		}
	}

	writeJSON(w, s.buildGraphResp(nil))
}

// ---- Metadata API handlers ----

func (s *server) handleNodeMeta(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string `json:"id"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
	}
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.graph.HasNode(req.ID) {
		http.Error(w, "node not found", 404)
		return
	}
	store := s.graph.NodeMeta(req.ID)
	if req.Limit <= 0 {
		req.Limit = 50
	}
	page := store.List(req.Offset, req.Limit)
	items := make([]metaEntry, len(page.Items))
	for i, e := range page.Items {
		items[i] = metaEntry{Key: e.Key, Value: e.Value}
	}
	writeJSON(w, metaResp{Items: items, Total: page.Total, Offset: page.Offset, HasMore: page.HasMore})
}

func (s *server) handleNodeMetaSet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID    string `json:"id"`
		Key   string `json:"key"`
		Value any    `json:"value"`
	}
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if req.Key == "" {
		http.Error(w, "key is required", 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.graph.HasNode(req.ID) {
		http.Error(w, "node not found", 404)
		return
	}
	s.graph.NodeMeta(req.ID).Set(req.Key, req.Value)
	writeJSON(w, s.buildGraphResp(nil))
}

func (s *server) handleNodeMetaDelete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID  string `json:"id"`
		Key string `json:"key"`
	}
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.graph.HasNode(req.ID) {
		http.Error(w, "node not found", 404)
		return
	}
	s.graph.NodeMeta(req.ID).Delete(req.Key)
	writeJSON(w, s.buildGraphResp(nil))
}

func (s *server) handleEdgeMeta(w http.ResponseWriter, r *http.Request) {
	var req struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
	}
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.graph.HasEdge(req.From, req.To) {
		http.Error(w, "edge not found", 404)
		return
	}
	store := s.graph.EdgeMeta(req.From, req.To)
	if req.Limit <= 0 {
		req.Limit = 50
	}
	page := store.List(req.Offset, req.Limit)
	items := make([]metaEntry, len(page.Items))
	for i, e := range page.Items {
		items[i] = metaEntry{Key: e.Key, Value: e.Value}
	}
	writeJSON(w, metaResp{Items: items, Total: page.Total, Offset: page.Offset, HasMore: page.HasMore})
}

func (s *server) handleEdgeMetaSet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		From  string `json:"from"`
		To    string `json:"to"`
		Key   string `json:"key"`
		Value any    `json:"value"`
	}
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if req.Key == "" {
		http.Error(w, "key is required", 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.graph.HasEdge(req.From, req.To) {
		http.Error(w, "edge not found", 404)
		return
	}
	s.graph.EdgeMeta(req.From, req.To).Set(req.Key, req.Value)
	writeJSON(w, s.buildGraphResp(nil))
}

func (s *server) handleEdgeMetaDelete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		From string `json:"from"`
		To   string `json:"to"`
		Key  string `json:"key"`
	}
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.graph.HasEdge(req.From, req.To) {
		http.Error(w, "edge not found", 404)
		return
	}
	s.graph.EdgeMeta(req.From, req.To).Delete(req.Key)
	writeJSON(w, s.buildGraphResp(nil))
}

// enrichCodebaseTemplate reads actual source files from disk and injects their
// content as metadata into the "codebase" template nodes.
func enrichCodebaseTemplate() {
	// Find repo root relative to this source file's location.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return
	}
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))

	// Map node IDs to file paths relative to repo root.
	fileMap := map[string]string{
		"graph.go":          "graph.go",
		"traverse.go":       "traverse.go",
		"query.go":          "query.go",
		"task.go":           "task.go",
		"store.go":          "store.go",
		"serialize.go":      "serialize.go",
		"doc.go":            "doc.go",
		"go.mod":            "go.mod",
		"Makefile":          "Makefile",
		".gitignore":        ".gitignore",
		"README.md":         "README.md",
		"main.go":           "cmd/visualizer/main.go",
		"templates.go":      "cmd/visualizer/templates.go",
		"graph_test.go":     "graph_test.go",
		"traverse_test.go":  "traverse_test.go",
		"query_test.go":     "query_test.go",
		"task_test.go":      "task_test.go",
		"store_test.go":     "store_test.go",
		"serialize_test.go": "serialize_test.go",
		"meta_test.go":      "meta_test.go",
	}

	// Find the codebase template.
	var tmpl *Template
	for i := range templates {
		if templates[i].ID == "codebase" {
			tmpl = &templates[i]
			break
		}
	}
	if tmpl == nil {
		return
	}

	for i := range tmpl.Nodes {
		relPath, ok := fileMap[tmpl.Nodes[i].ID]
		if !ok {
			continue
		}
		fullPath := filepath.Join(repoRoot, relPath)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}
		if tmpl.Nodes[i].Meta == nil {
			tmpl.Nodes[i].Meta = make(map[string]any)
		}
		tmpl.Nodes[i].Meta["content"] = string(data)
	}
}

func main() {
	enrichCodebaseTemplate()
	s := newServer(true)

	mux := http.NewServeMux()

	// Serve the embedded static file.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := static.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	// API routes.
	mux.HandleFunc("/api/graph", s.handleGetGraph)
	mux.HandleFunc("/api/node/add", s.handleAddNode)
	mux.HandleFunc("/api/node/remove", s.handleRemoveNode)
	mux.HandleFunc("/api/edge/add", s.handleAddEdge)
	mux.HandleFunc("/api/edge/remove", s.handleRemoveEdge)
	mux.HandleFunc("/api/node/position", s.handleUpdatePos)
	mux.HandleFunc("/api/graph/directed", s.handleSetDirected)
	mux.HandleFunc("/api/graph/clear", s.handleClear)
	mux.HandleFunc("/api/templates", s.handleGetTemplates)
	mux.HandleFunc("/api/template/load", s.handleLoadTemplate)
	mux.HandleFunc("/api/algo", s.handleAlgo)
	mux.HandleFunc("/api/node/status", s.handleUpdateNodeStatus)
	mux.HandleFunc("/api/plan/load", s.handleLoadPlan)
	mux.HandleFunc("/api/directory/load", s.handleLoadDirectory)

	// Export/Import API routes.
	mux.HandleFunc("/api/graph/export", s.handleExport)
	mux.HandleFunc("/api/graph/import", s.handleImport)

	// Metadata API routes.
	mux.HandleFunc("/api/node/meta", s.handleNodeMeta)
	mux.HandleFunc("/api/node/meta/set", s.handleNodeMetaSet)
	mux.HandleFunc("/api/node/meta/delete", s.handleNodeMetaDelete)
	mux.HandleFunc("/api/edge/meta", s.handleEdgeMeta)
	mux.HandleFunc("/api/edge/meta/set", s.handleEdgeMetaSet)
	mux.HandleFunc("/api/edge/meta/delete", s.handleEdgeMetaDelete)

	addr := ":8090"
	fmt.Printf("Spine visualizer running at http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
