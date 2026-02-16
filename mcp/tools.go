package mcp

func (s *Server) registerTools() {
	s.addTool("open_graph", "Open or create a named graph",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string", "description": "Graph name"},
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
}
