package mcp

func (s *Server) registerTools() {
	s.addTool("open_graph", "Open or create a named graph",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name":     map[string]any{"type": "string", "description": "Graph name"},
				"directed": map[string]any{"type": "boolean", "description": "Whether the graph is directed (default true)"},
			},
			"required": []string{"name"},
		}, s.handleOpenGraph)

	s.addTool("save_graph", "Persist a graph to disk",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"name"},
		}, s.handleSaveGraph)

	s.addTool("list_graphs", "List all persisted graphs",
		map[string]any{
			"type": "object",
			"properties": map[string]any{},
		}, s.handleListGraphs)

	s.addTool("delete_graph", "Delete a graph from disk and memory",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"name"},
		}, s.handleDeleteGraph)

	s.addTool("graph_summary", "Get structural statistics for a graph (roots, leaves, status counts, components)",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"name"},
		}, s.handleGraphSummary)

	s.addTool("upsert", "Batch create/update nodes, edges, and metadata",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
				"nodes": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id":     map[string]any{"type": "string"},
							"label":  map[string]any{"type": "string"},
							"status": map[string]any{"type": "string"},
							"meta":   map[string]any{"type": "object"},
							"delete": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						},
						"required": []string{"id"},
					},
				},
				"edges": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"from":   map[string]any{"type": "string"},
							"to":     map[string]any{"type": "string"},
							"label":  map[string]any{"type": "string"},
							"weight": map[string]any{"type": "number"},
							"meta":   map[string]any{"type": "object"},
							"delete": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						},
						"required": []string{"from", "to"},
					},
				},
			},
			"required": []string{"graph"},
		}, s.handleUpsert)

	s.addTool("read_nodes", "Selective read with filters, key projection, and pagination",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
				"ids":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"keys":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"filters": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"key":   map[string]any{"type": "string"},
							"op":    map[string]any{"type": "string"},
							"value": map[string]any{},
						},
						"required": []string{"key", "op"},
					},
				},
				"include_edges": map[string]any{"type": "boolean"},
				"offset":        map[string]any{"type": "integer"},
				"limit":         map[string]any{"type": "integer"},
			},
			"required": []string{"graph"},
		}, s.handleReadNodes)

	s.addTool("transition", "Change node status with auto-ready propagation",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph":  map[string]any{"type": "string", "description": "Graph name"},
				"id":     map[string]any{"type": "string", "description": "Node ID"},
				"status": map[string]any{"type": "string", "description": "Target status"},
			},
			"required": []string{"graph", "id", "status"},
		}, s.handleTransition)

	s.addTool("remove", "Delete nodes and/or edges from a graph",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
				"nodes": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"edges": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"from": map[string]any{"type": "string"},
							"to":   map[string]any{"type": "string"},
						},
						"required": []string{"from", "to"},
					},
				},
			},
			"required": []string{"graph"},
		}, s.handleRemove)

	s.addTool("scc", "Compute strongly connected components of a graph",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"graph"},
		}, s.handleSCC)

	s.addTool("mst", "Compute minimum spanning tree of an undirected graph",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"graph"},
		}, s.handleMST)

	s.addTool("bfs", "Breadth-first search traversal from a start node",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
				"start": map[string]any{"type": "string", "description": "Start node ID"},
			},
			"required": []string{"graph", "start"},
		}, s.handleBFS)

	s.addTool("dfs", "Depth-first search traversal from a start node",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
				"start": map[string]any{"type": "string", "description": "Start node ID"},
			},
			"required": []string{"graph", "start"},
		}, s.handleDFS)

	s.addTool("shortest_path", "Find shortest path between two nodes (Dijkstra)",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
				"src":   map[string]any{"type": "string", "description": "Source node ID"},
				"dst":   map[string]any{"type": "string", "description": "Destination node ID"},
			},
			"required": []string{"graph", "src", "dst"},
		}, s.handleShortestPath)

	s.addTool("topological_sort", "Compute topological ordering of a directed acyclic graph",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"graph"},
		}, s.handleTopologicalSort)

	s.addTool("cycle_detect", "Detect cycles in the graph",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"graph"},
		}, s.handleCycleDetect)

	s.addTool("connected_components", "Find connected components of the graph",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"graph"},
		}, s.handleConnectedComponents)

	s.addTool("ancestors", "Find all ancestor nodes of a given node",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
				"id":    map[string]any{"type": "string", "description": "Node ID"},
			},
			"required": []string{"graph", "id"},
		}, s.handleAncestors)

	s.addTool("descendants", "Find all descendant nodes of a given node",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
				"id":    map[string]any{"type": "string", "description": "Node ID"},
			},
			"required": []string{"graph", "id"},
		}, s.handleDescendants)

	s.addTool("roots", "Find all root nodes (no incoming edges)",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"graph"},
		}, s.handleRoots)

	s.addTool("leaves", "Find all leaf nodes (no outgoing edges)",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"graph"},
		}, s.handleLeaves)

	s.addTool("transitive_closure", "Compute the transitive closure of a directed graph",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"graph"},
		}, s.handleTransitiveClosure)

	s.addTool("validate_graph", "Validate internal consistency of a graph",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"graph"},
		}, s.handleValidateGraph)

	s.addTool("diff_graphs", "Compute differences between two graphs",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph_a": map[string]any{"type": "string", "description": "First graph name"},
				"graph_b": map[string]any{"type": "string", "description": "Second graph name"},
			},
			"required": []string{"graph_a", "graph_b"},
		}, s.handleDiffGraphs)

	s.addTool("degree_centrality", "Compute degree centrality for all nodes",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"graph"},
		}, s.handleDegreeCentrality)

	s.addTool("betweenness_centrality", "Compute betweenness centrality for all nodes",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"graph"},
		}, s.handleBetweennessCentrality)

	s.addTool("closeness_centrality", "Compute closeness centrality for all nodes",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"graph"},
		}, s.handleClosenessCentrality)

	s.addTool("pagerank", "Compute PageRank scores for all nodes",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph":     map[string]any{"type": "string", "description": "Graph name"},
				"damping":   map[string]any{"type": "number", "description": "Damping factor (default 0.85)"},
				"max_iter":  map[string]any{"type": "integer", "description": "Maximum iterations (default 100)"},
				"tolerance": map[string]any{"type": "number", "description": "Convergence tolerance (default 1e-6)"},
			},
			"required": []string{"graph"},
		}, s.handlePageRank)

	s.addTool("all_pairs_shortest_paths", "Compute shortest paths between all pairs of nodes (Floyd-Warshall)",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"graph"},
		}, s.handleAllPairsShortestPaths)

	s.addTool("critical_path", "Compute the critical path in a DAG",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
			},
			"required": []string{"graph"},
		}, s.handleCriticalPath)

	s.addTool("max_flow", "Compute maximum flow between source and sink (Edmonds-Karp)",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph":  map[string]any{"type": "string", "description": "Graph name"},
				"source": map[string]any{"type": "string", "description": "Source node ID"},
				"sink":   map[string]any{"type": "string", "description": "Sink node ID"},
			},
			"required": []string{"graph", "source", "sink"},
		}, s.handleMaxFlow)

	s.addTool("explain_path", "Get a human-readable explanation of the shortest path between two nodes",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
				"src":   map[string]any{"type": "string", "description": "Source node ID"},
				"dst":   map[string]any{"type": "string", "description": "Destination node ID"},
			},
			"required": []string{"graph", "src", "dst"},
		}, s.handleExplainPath)

	s.addTool("explain_component", "Get a human-readable explanation of a node's component membership",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
				"id":    map[string]any{"type": "string", "description": "Node ID"},
			},
			"required": []string{"graph", "id"},
		}, s.handleExplainComponent)

	s.addTool("explain_centrality", "Get a human-readable explanation of a node's centrality ranking",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
				"id":    map[string]any{"type": "string", "description": "Node ID"},
			},
			"required": []string{"graph", "id"},
		}, s.handleExplainCentrality)

	s.addTool("explain_dependency", "Get a human-readable explanation of the dependency between two nodes",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"graph": map[string]any{"type": "string", "description": "Graph name"},
				"src":   map[string]any{"type": "string", "description": "Source node ID"},
				"dst":   map[string]any{"type": "string", "description": "Target node ID"},
			},
			"required": []string{"graph", "src", "dst"},
		}, s.handleExplainDependency)
}
