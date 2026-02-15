package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	Label string `json:"label"`
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
	ID    string  `json:"id"`
	Label string  `json:"label"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}

type edgeResp struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Label  string  `json:"label"`
	Weight float64 `json:"weight"`
}

type algoResultResp struct {
	Algorithm      string     `json:"algorithm"`
	VisitedOrder   []string   `json:"visitedOrder,omitempty"`
	Path           []string   `json:"path,omitempty"`
	Cost           float64    `json:"cost,omitempty"`
	HasCycle       bool       `json:"hasCycle,omitempty"`
	Cycle          []string   `json:"cycle,omitempty"`
	Components     [][]string `json:"components,omitempty"`
	Roots          []string   `json:"roots,omitempty"`
	Leaves         []string   `json:"leaves,omitempty"`
	Ancestors      []string   `json:"ancestors,omitempty"`
	Descendants    []string   `json:"descendants,omitempty"`
	HighlightNodes []string   `json:"highlightNodes,omitempty"`
	HighlightEdges [][2]string `json:"highlightEdges,omitempty"`
	Error          string     `json:"error,omitempty"`
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
		nr[i] = nodeResp{ID: n.ID, Label: n.Data.Label, X: pos.X, Y: pos.Y}
	}
	edges := s.graph.Edges()
	er := make([]edgeResp, len(edges))
	for i, e := range edges {
		er[i] = edgeResp{From: e.From, To: e.To, Label: e.Data.Label, Weight: e.Weight}
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
	// Rebuild the graph with the new mode, keeping all nodes and edges.
	old := s.graph
	s.graph = spine.NewGraph[NodeData, EdgeData](req.Directed)
	for _, n := range old.Nodes() {
		s.graph.AddNode(n.ID, n.Data)
	}
	for _, e := range old.Edges() {
		s.graph.AddEdge(e.From, e.To, e.Data, e.Weight)
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

func (s *server) handleClear(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	directed := s.graph.Directed
	s.graph = spine.NewGraph[NodeData, EdgeData](directed)
	s.positions = make(map[string]Position)
	writeJSON(w, s.buildGraphResp(nil))
}

func main() {
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
	mux.HandleFunc("/api/algo", s.handleAlgo)

	addr := ":8090"
	fmt.Printf("Spine visualizer running at http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
