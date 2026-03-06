package spine

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// PathExplanation provides a human-readable explanation of a shortest path.
type PathExplanation struct {
	Source      string   `json:"source"`
	Destination string   `json:"destination"`
	Path        []string `json:"path"`
	PathLength  int      `json:"path_length"`
	TotalWeight float64  `json:"total_weight"`
	Explanation string   `json:"explanation"`
	Steps       []string `json:"steps"`
}

// ComponentExplanation provides a human-readable explanation of a node's component membership.
type ComponentExplanation struct {
	NodeID        string   `json:"node_id"`
	ComponentID   int      `json:"component_id"`
	ComponentSize int      `json:"component_size"`
	Members       []string `json:"members"`
	Explanation   string   `json:"explanation"`
	Connections   []string `json:"connections"`
}

// CentralityExplanation provides a human-readable explanation of a node's centrality.
type CentralityExplanation struct {
	NodeID      string   `json:"node_id"`
	Rank        int      `json:"rank"`
	Score       float64  `json:"score"`
	TotalNodes  int      `json:"total_nodes"`
	Explanation string   `json:"explanation"`
	Factors     []string `json:"factors"`
}

// DependencyExplanation provides a human-readable explanation of a dependency between two nodes.
type DependencyExplanation struct {
	Source       string     `json:"source"`
	Target       string     `json:"target"`
	IsDirect     bool       `json:"is_direct"`
	IsTransitive bool       `json:"is_transitive"`
	Paths        [][]string `json:"paths"`
	Explanation  string     `json:"explanation"`
}

// ExplainPath computes the shortest path and returns a human-readable explanation.
func ExplainPath[N, E any](g *Graph[N, E], src, dst string) (*PathExplanation, error) {
	path, cost, err := ShortestPath(g, src, dst)
	if err != nil {
		return nil, err
	}

	steps := make([]string, 0, len(path)-1)
	for i := 0; i < len(path)-1; i++ {
		from := path[i]
		to := path[i+1]
		e, _ := g.GetEdge(from, to)
		step := fmt.Sprintf("Step %d: '%s' -> '%s' (weight: %.2f)", i+1, from, to, e.Weight)
		steps = append(steps, step)
	}

	hops := len(path) - 1
	explanation := fmt.Sprintf(
		"There is a path of %d hop(s) from '%s' to '%s' with a total weight of %.2f. "+
			"The path traverses through %d node(s): %s.",
		hops, src, dst, cost, len(path), strings.Join(path, " -> "),
	)

	return &PathExplanation{
		Source:      src,
		Destination: dst,
		Path:        path,
		PathLength:  hops,
		TotalWeight: cost,
		Explanation: explanation,
		Steps:       steps,
	}, nil
}

// ExplainComponent explains which component a node belongs to and its connections.
func ExplainComponent[N, E any](g *Graph[N, E], nodeID string) (*ComponentExplanation, error) {
	if !g.HasNode(nodeID) {
		return nil, errors.New("node not found")
	}

	var components [][]string
	if g.Directed {
		components = StronglyConnectedComponents(g)
	} else {
		components = ConnectedComponents(g)
	}

	compIdx := -1
	var members []string
	for i, comp := range components {
		for _, id := range comp {
			if id == nodeID {
				compIdx = i
				members = comp
				break
			}
		}
		if compIdx >= 0 {
			break
		}
	}

	if compIdx < 0 {
		return nil, errors.New("node not found in any component")
	}

	// Find direct connections within the component
	var connections []string
	for _, nb := range g.Neighbors(nodeID) {
		for _, m := range members {
			if nb == m {
				connections = append(connections, fmt.Sprintf("'%s' is directly connected to '%s'", nodeID, nb))
				break
			}
		}
	}
	// Also check incoming edges
	for _, e := range g.InEdges(nodeID) {
		for _, m := range members {
			if e.From == m {
				alreadyListed := false
				for _, c := range connections {
					if strings.Contains(c, fmt.Sprintf("'%s'", e.From)) {
						alreadyListed = true
						break
					}
				}
				if !alreadyListed {
					connections = append(connections, fmt.Sprintf("'%s' receives an edge from '%s'", nodeID, e.From))
				}
				break
			}
		}
	}

	compType := "connected component"
	if g.Directed {
		compType = "strongly connected component"
	}

	explanation := fmt.Sprintf(
		"Node '%s' belongs to %s #%d, which contains %d node(s): [%s]. "+
			"The graph has %d %s(s) in total.",
		nodeID, compType, compIdx+1, len(members),
		strings.Join(members, ", "),
		len(components), compType,
	)

	return &ComponentExplanation{
		NodeID:        nodeID,
		ComponentID:   compIdx + 1,
		ComponentSize: len(members),
		Members:       members,
		Explanation:   explanation,
		Connections:   connections,
	}, nil
}

// ExplainCentrality explains a node's degree centrality ranking.
func ExplainCentrality[N, E any](g *Graph[N, E], nodeID string) (*CentralityExplanation, error) {
	if !g.HasNode(nodeID) {
		return nil, errors.New("node not found")
	}

	result := DegreeCentrality(g)
	score := result.Scores[nodeID]
	totalNodes := len(result.Scores)

	// Rank nodes by score (descending)
	type nodeScore struct {
		ID    string
		Score float64
	}
	ranked := make([]nodeScore, 0, totalNodes)
	for id, s := range result.Scores {
		ranked = append(ranked, nodeScore{id, s})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Score != ranked[j].Score {
			return ranked[i].Score > ranked[j].Score
		}
		return ranked[i].ID < ranked[j].ID
	})

	rank := 1
	for i, ns := range ranked {
		if ns.ID == nodeID {
			rank = i + 1
			break
		}
	}

	var factors []string
	outDeg := len(g.OutEdges(nodeID))
	inDeg := len(g.InEdges(nodeID))
	neighbors := g.Neighbors(nodeID)

	if g.Directed {
		factors = append(factors, fmt.Sprintf("Out-degree: %d", outDeg))
		factors = append(factors, fmt.Sprintf("In-degree: %d", inDeg))
	} else {
		factors = append(factors, fmt.Sprintf("Degree: %d", len(neighbors)))
	}
	factors = append(factors, fmt.Sprintf("Neighbor count: %d", len(neighbors)))

	if len(neighbors) > 0 {
		nbs := make([]string, len(neighbors))
		copy(nbs, neighbors)
		if len(nbs) > 5 {
			nbs = nbs[:5]
			factors = append(factors, fmt.Sprintf("Connected to: %s, ... and %d more",
				strings.Join(nbs, ", "), len(neighbors)-5))
		} else {
			factors = append(factors, fmt.Sprintf("Connected to: %s", strings.Join(nbs, ", ")))
		}
	}

	explanation := fmt.Sprintf(
		"Node '%s' is ranked #%d out of %d nodes by degree centrality with a score of %.4f.",
		nodeID, rank, totalNodes, score,
	)

	return &CentralityExplanation{
		NodeID:      nodeID,
		Rank:        rank,
		Score:       score,
		TotalNodes:  totalNodes,
		Explanation: explanation,
		Factors:     factors,
	}, nil
}

// ExplainDependency explains the dependency relationship between two nodes.
func ExplainDependency[N, E any](g *Graph[N, E], src, dst string) (*DependencyExplanation, error) {
	if !g.HasNode(src) {
		return nil, errors.New("source node not found")
	}
	if !g.HasNode(dst) {
		return nil, errors.New("target node not found")
	}

	isDirect := g.HasEdge(src, dst)

	// BFS to find simple paths (limit to avoid explosion)
	const maxPaths = 5
	const maxDepth = 10
	var paths [][]string
	findPaths(g, src, dst, maxPaths, maxDepth, &paths)

	isTransitive := len(paths) > 0

	var explanation string
	if isDirect && isTransitive && len(paths) > 1 {
		explanation = fmt.Sprintf(
			"'%s' has a direct dependency on '%s' and there are %d path(s) between them. "+
				"This means '%s' depends on '%s' both directly and transitively.",
			src, dst, len(paths), src, dst,
		)
	} else if isDirect {
		explanation = fmt.Sprintf(
			"'%s' has a direct dependency on '%s' via a single edge.",
			src, dst,
		)
	} else if isTransitive {
		explanation = fmt.Sprintf(
			"'%s' has a transitive (indirect) dependency on '%s' through %d path(s). "+
				"There is no direct edge from '%s' to '%s'.",
			src, dst, len(paths), src, dst,
		)
	} else {
		explanation = fmt.Sprintf(
			"There is no dependency from '%s' to '%s'. No path exists between these nodes in the given direction.",
			src, dst,
		)
	}

	return &DependencyExplanation{
		Source:       src,
		Target:       dst,
		IsDirect:     isDirect,
		IsTransitive: isTransitive,
		Paths:        paths,
		Explanation:  explanation,
	}, nil
}

// findPaths uses DFS to find simple paths from src to dst, up to maxPaths paths and maxDepth hops.
func findPaths[N, E any](g *Graph[N, E], src, dst string, maxPaths, maxDepth int, result *[][]string) {
	visited := make(map[string]bool)
	path := []string{src}
	visited[src] = true
	var dfs func()
	dfs = func() {
		if len(*result) >= maxPaths {
			return
		}
		cur := path[len(path)-1]
		if cur == dst {
			p := make([]string, len(path))
			copy(p, path)
			*result = append(*result, p)
			return
		}
		if len(path) > maxDepth {
			return
		}
		for _, nb := range g.Neighbors(cur) {
			if !visited[nb] {
				visited[nb] = true
				path = append(path, nb)
				dfs()
				path = path[:len(path)-1]
				visited[nb] = false
			}
		}
	}
	dfs()
}
