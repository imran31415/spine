package main

// Template defines a pre-built graph that users can load.
type Template struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Directed    bool           `json:"directed"`
	Nodes       []templateNode `json:"nodes"`
	Edges       []templateEdge `json:"edges"`
}

type templateNode struct {
	ID     string         `json:"id"`
	Label  string         `json:"label"`
	X      float64        `json:"x"`
	Y      float64        `json:"y"`
	Status string         `json:"status,omitempty"`
	Meta   map[string]any `json:"meta,omitempty"`
}

type templateEdge struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Label  string  `json:"label"`
	Weight float64 `json:"weight"`
}

// templateSummary is the compact form returned by GET /api/templates.
type templateSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Directed    bool   `json:"directed"`
	NodeCount   int    `json:"nodeCount"`
	EdgeCount   int    `json:"edgeCount"`
}

func (t Template) summary() templateSummary {
	return templateSummary{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Directed:    t.Directed,
		NodeCount:   len(t.Nodes),
		EdgeCount:   len(t.Edges),
	}
}

var templates = []Template{
	{
		ID:          "workflow",
		Name:        "CI/CD Workflow",
		Description: "Fan-out/fan-in DAG — try topo sort, roots, and leaves",
		Directed:    true,
		Nodes: []templateNode{
			{ID: "push", Label: "Push", X: 450, Y: 50, Meta: map[string]any{
				"trigger": "on_push", "branch": "main",
			}},
			{ID: "lint", Label: "Lint", X: 200, Y: 180, Meta: map[string]any{
				"tool": "golangci-lint", "config": ".golangci.yml",
			}},
			{ID: "test", Label: "Test", X: 450, Y: 180, Meta: map[string]any{
				"framework": "go test", "flags": "-race -v", "min_coverage": "80%",
			}},
			{ID: "build", Label: "Build", X: 700, Y: 180, Meta: map[string]any{
				"output": "./bin/app", "os": "linux", "arch": "amd64",
			}},
			{ID: "scan", Label: "Security Scan", X: 200, Y: 330, Meta: map[string]any{
				"tool": "gosec", "severity": "high",
			}},
			{ID: "coverage", Label: "Coverage", X: 450, Y: 330},
			{ID: "docker", Label: "Docker", X: 700, Y: 330, Meta: map[string]any{
				"base_image": "golang:1.22-alpine", "registry": "ghcr.io", "tag": "latest",
			}},
			{ID: "staging", Label: "Staging", X: 450, Y: 480},
			{ID: "deploy", Label: "Deploy", X: 450, Y: 620, Meta: map[string]any{
				"environment": "production", "region": "us-east-1", "replicas": 3,
			}},
		},
		Edges: []templateEdge{
			{From: "push", To: "lint", Weight: 1},
			{From: "push", To: "test", Weight: 1},
			{From: "push", To: "build", Weight: 1},
			{From: "lint", To: "scan", Weight: 1},
			{From: "test", To: "coverage", Weight: 1},
			{From: "build", To: "docker", Weight: 1},
			{From: "scan", To: "staging", Weight: 1},
			{From: "coverage", To: "staging", Weight: 1},
			{From: "docker", To: "staging", Weight: 1},
			{From: "staging", To: "deploy", Weight: 1},
		},
	},
	{
		ID:          "knowledge",
		Name:        "Knowledge Graph",
		Description: "Undirected concept map — try BFS, DFS, and components",
		Directed:    false,
		Nodes: []templateNode{
			{ID: "ml", Label: "Machine Learning", X: 450, Y: 60, Meta: map[string]any{
				"field": "Computer Science", "since": 1959,
			}},
			{ID: "nn", Label: "Neural Nets", X: 250, Y: 180, Meta: map[string]any{
				"inspired_by": "biological neurons", "key_paper": "Perceptrons (1969)",
			}},
			{ID: "dl", Label: "Deep Learning", X: 100, Y: 330},
			{ID: "cnn", Label: "CNN", X: 250, Y: 450, Meta: map[string]any{
				"use_case": "image recognition", "key_paper": "ImageNet (2012)",
			}},
			{ID: "nlp", Label: "NLP", X: 450, Y: 200, Meta: map[string]any{
				"applications": []any{"translation", "summarization", "chat"},
			}},
			{ID: "transformers", Label: "Transformers", X: 450, Y: 370, Meta: map[string]any{
				"key_paper": "Attention Is All You Need", "year": 2017, "authors": "Vaswani et al.",
			}},
			{ID: "stats", Label: "Statistics", X: 700, Y: 100},
			{ID: "regression", Label: "Regression", X: 800, Y: 250},
			{ID: "pca", Label: "PCA", X: 700, Y: 400},
			{ID: "rl", Label: "Reinforcement", X: 600, Y: 550, Meta: map[string]any{
				"applications": []any{"game AI", "robotics", "resource management"},
			}},
		},
		Edges: []templateEdge{
			{From: "ml", To: "nn", Weight: 1},
			{From: "ml", To: "nlp", Weight: 1},
			{From: "ml", To: "stats", Weight: 1},
			{From: "nn", To: "dl", Weight: 1},
			{From: "dl", To: "cnn", Weight: 1},
			{From: "nlp", To: "transformers", Weight: 1},
			{From: "nn", To: "transformers", Weight: 1},
			{From: "stats", To: "regression", Weight: 1},
			{From: "stats", To: "pca", Weight: 1},
			{From: "ml", To: "rl", Weight: 1},
			{From: "transformers", To: "rl", Weight: 1},
		},
	},
	{
		ID:          "dependencies",
		Name:        "Dependency Tree",
		Description: "Package dependency DAG — try ancestors, descendants, topo sort",
		Directed:    true,
		Nodes: []templateNode{
			{ID: "app", Label: "app", X: 450, Y: 50},
			{ID: "api", Label: "api", X: 250, Y: 170},
			{ID: "web", Label: "web", X: 650, Y: 170},
			{ID: "auth", Label: "auth", X: 100, Y: 310},
			{ID: "db", Label: "db", X: 350, Y: 310},
			{ID: "ui", Label: "ui", X: 550, Y: 310},
			{ID: "router", Label: "router", X: 750, Y: 310},
			{ID: "crypto", Label: "crypto", X: 100, Y: 460},
			{ID: "logger", Label: "logger", X: 300, Y: 460},
			{ID: "config", Label: "config", X: 500, Y: 460},
			{ID: "utils", Label: "utils", X: 700, Y: 460},
		},
		Edges: []templateEdge{
			{From: "app", To: "api", Weight: 1},
			{From: "app", To: "web", Weight: 1},
			{From: "api", To: "auth", Weight: 1},
			{From: "api", To: "db", Weight: 1},
			{From: "web", To: "ui", Weight: 1},
			{From: "web", To: "router", Weight: 1},
			{From: "auth", To: "crypto", Weight: 1},
			{From: "auth", To: "logger", Weight: 1},
			{From: "db", To: "logger", Weight: 1},
			{From: "db", To: "config", Weight: 1},
			{From: "ui", To: "config", Weight: 1},
			{From: "router", To: "utils", Weight: 1},
			{From: "router", To: "logger", Weight: 1},
		},
	},
	{
		ID:          "statemachine",
		Name:        "State Machine",
		Description: "Has cycles from retry/return flows — try cycle detection",
		Directed:    true,
		Nodes: []templateNode{
			{ID: "idle", Label: "Idle", X: 120, Y: 300},
			{ID: "validating", Label: "Validating", X: 300, Y: 150},
			{ID: "processing", Label: "Processing", X: 520, Y: 150},
			{ID: "waiting", Label: "Waiting", X: 700, Y: 300},
			{ID: "retrying", Label: "Retrying", X: 520, Y: 450},
			{ID: "completed", Label: "Completed", X: 700, Y: 530},
			{ID: "failed", Label: "Failed", X: 300, Y: 530},
			{ID: "cancelled", Label: "Cancelled", X: 120, Y: 500},
		},
		Edges: []templateEdge{
			{From: "idle", To: "validating", Weight: 1},
			{From: "validating", To: "processing", Weight: 1},
			{From: "validating", To: "failed", Weight: 1},
			{From: "processing", To: "waiting", Weight: 1},
			{From: "processing", To: "failed", Weight: 1},
			{From: "waiting", To: "completed", Weight: 1},
			{From: "waiting", To: "retrying", Weight: 1},
			{From: "retrying", To: "processing", Weight: 1},
			{From: "retrying", To: "failed", Weight: 1},
			{From: "idle", To: "cancelled", Weight: 1},
			{From: "failed", To: "idle", Weight: 1},
		},
	},
	{
		ID:          "orgchart",
		Name:        "Org Chart",
		Description: "Strict tree hierarchy — try roots, leaves, BFS",
		Directed:    true,
		Nodes: []templateNode{
			{ID: "ceo", Label: "CEO", X: 450, Y: 50},
			{ID: "cto", Label: "CTO", X: 220, Y: 180},
			{ID: "cfo", Label: "CFO", X: 450, Y: 180},
			{ID: "cmo", Label: "CMO", X: 680, Y: 180},
			{ID: "eng1", Label: "Eng Lead", X: 120, Y: 340},
			{ID: "eng2", Label: "Architect", X: 320, Y: 340},
			{ID: "fin1", Label: "Controller", X: 450, Y: 340},
			{ID: "mkt1", Label: "Brand Lead", X: 580, Y: 340},
			{ID: "mkt2", Label: "Growth Lead", X: 780, Y: 340},
			{ID: "dev1", Label: "Dev A", X: 70, Y: 500},
			{ID: "dev2", Label: "Dev B", X: 220, Y: 500},
		},
		Edges: []templateEdge{
			{From: "ceo", To: "cto", Weight: 1},
			{From: "ceo", To: "cfo", Weight: 1},
			{From: "ceo", To: "cmo", Weight: 1},
			{From: "cto", To: "eng1", Weight: 1},
			{From: "cto", To: "eng2", Weight: 1},
			{From: "cfo", To: "fin1", Weight: 1},
			{From: "cmo", To: "mkt1", Weight: 1},
			{From: "cmo", To: "mkt2", Weight: 1},
			{From: "eng1", To: "dev1", Weight: 1},
			{From: "eng1", To: "dev2", Weight: 1},
		},
	},
	{
		ID:          "microservices",
		Name:        "Microservices",
		Description: "Weighted edges (latency in ms) — try shortest path",
		Directed:    true,
		Nodes: []templateNode{
			{ID: "gateway", Label: "Gateway", X: 450, Y: 50, Meta: map[string]any{
				"port": 8080, "rate_limit": "1000/min", "timeout_ms": 30000,
			}},
			{ID: "auth", Label: "Auth", X: 200, Y: 180, Meta: map[string]any{
				"strategy": "JWT", "token_ttl": "1h", "issuer": "auth-service",
			}},
			{ID: "users", Label: "Users", X: 450, Y: 180},
			{ID: "orders", Label: "Orders", X: 700, Y: 180},
			{ID: "products", Label: "Products", X: 200, Y: 350},
			{ID: "payments", Label: "Payments", X: 450, Y: 350},
			{ID: "inventory", Label: "Inventory", X: 700, Y: 350},
			{ID: "notify", Label: "Notify", X: 300, Y: 520, Meta: map[string]any{
				"channels": []any{"email", "slack", "webhook"}, "retry_count": 3,
			}},
			{ID: "cache", Label: "Cache", X: 550, Y: 520, Meta: map[string]any{
				"engine": "Redis", "ttl_seconds": 300, "max_memory": "256mb",
			}},
			{ID: "db", Label: "Database", X: 700, Y: 520, Meta: map[string]any{
				"engine": "PostgreSQL", "version": "15", "pool_size": 20,
			}},
		},
		Edges: []templateEdge{
			{From: "gateway", To: "auth", Weight: 5},
			{From: "gateway", To: "users", Weight: 12},
			{From: "gateway", To: "orders", Weight: 15},
			{From: "auth", To: "users", Weight: 8},
			{From: "auth", To: "cache", Weight: 2},
			{From: "users", To: "notify", Weight: 20},
			{From: "orders", To: "payments", Weight: 25},
			{From: "orders", To: "inventory", Weight: 10},
			{From: "products", To: "inventory", Weight: 7},
			{From: "products", To: "cache", Weight: 3},
			{From: "payments", To: "notify", Weight: 15},
			{From: "inventory", To: "db", Weight: 5},
			{From: "cache", To: "db", Weight: 4},
		},
	},
	{
		ID:          "taskplan",
		Name:        "LLM Task Plan",
		Description: "Task execution plan with statuses and metadata — click nodes to inspect",
		Directed:    true,
		Nodes: []templateNode{
			{ID: "analyze", Label: "Analyze", X: 450, Y: 50, Status: "done", Meta: map[string]any{
				"prompt": "Analyze the codebase structure and identify key components",
				"model":  "claude-opus-4-6", "temperature": 0.3, "output_format": "structured_json",
			}},
			{ID: "plan", Label: "Plan", X: 450, Y: 180, Status: "done", Meta: map[string]any{
				"strategy": "top-down decomposition", "max_tasks": 8, "priority": "correctness",
			}},
			{ID: "research", Label: "Research", X: 250, Y: 320, Status: "running", Meta: map[string]any{
				"sources": []any{"documentation", "source code", "tests"},
				"depth":   "comprehensive", "query": "graph metadata systems",
			}},
			{ID: "setup", Label: "Setup", X: 650, Y: 320, Status: "done", Meta: map[string]any{
				"runtime": "go1.22", "environment": "development",
			}},
			{ID: "implement", Label: "Implement", X: 450, Y: 460, Status: "ready", Meta: map[string]any{
				"language": "Go", "pattern": "method receiver",
				"files": []any{"store.go", "graph.go"},
			}},
			{ID: "test", Label: "Test", X: 450, Y: 590, Status: "pending", Meta: map[string]any{
				"framework": "testing", "coverage_target": "90%", "race_detector": true,
			}},
			{ID: "review", Label: "Review", X: 450, Y: 720, Status: "pending", Meta: map[string]any{
				"checklist": []any{"correctness", "edge cases", "documentation"},
			}},
			{ID: "deploy", Label: "Deploy", X: 450, Y: 850, Status: "pending", Meta: map[string]any{
				"target": "production", "strategy": "blue-green", "rollback_plan": true,
			}},
		},
		Edges: []templateEdge{
			{From: "analyze", To: "plan", Weight: 1},
			{From: "plan", To: "research", Weight: 1},
			{From: "plan", To: "setup", Weight: 1},
			{From: "research", To: "implement", Weight: 1},
			{From: "setup", To: "implement", Weight: 1},
			{From: "implement", To: "test", Weight: 1},
			{From: "test", To: "review", Weight: 1},
			{From: "review", To: "deploy", Weight: 1},
		},
	},
	{
		ID:          "codebase",
		Name:        "Codebase File Tree",
		Description: "Source file tree with file metadata — try roots, leaves, descendants, ancestors",
		Directed:    true,
		Nodes: []templateNode{
			// Root
			{ID: "spine", Label: "spine/", X: 450, Y: 30, Meta: map[string]any{
				"type": "directory", "description": "Root of the spine graph library",
				"total_files": 20, "language": "Go",
			}},
			// Top-level source files
			{ID: "graph.go", Label: "graph.go", X: 80, Y: 170, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 313, "bytes": 7681,
				"description": "Core graph data structure with generic nodes and edges",
				"exports":     []any{"Graph[N,E]", "Node[T]", "Edge[T]", "NewGraph", "AddNode", "AddEdge"},
			}},
			{ID: "traverse.go", Label: "traverse.go", X: 230, Y: 170, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 328, "bytes": 7631,
				"description": "Graph traversal algorithms",
				"exports":     []any{"BFS", "DFS", "ShortestPath", "TopologicalSort", "CycleDetect", "ConnectedComponents"},
			}},
			{ID: "query.go", Label: "query.go", X: 380, Y: 170, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 88, "bytes": 2081,
				"description": "Query and filter operations on graphs",
				"exports":     []any{"FilterNodes", "FilterEdges", "Ancestors", "Descendants", "Roots", "Leaves"},
			}},
			{ID: "task.go", Label: "task.go", X: 530, Y: 170, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 254, "bytes": 5909,
				"description": "DAG-based task scheduler with concurrency control",
				"exports":     []any{"TaskGraph[T]", "TaskState", "Execute"},
			}},
			{ID: "store.go", Label: "store.go", X: 680, Y: 170, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 261, "bytes": 5291,
				"description": "Metadata key-value store with schema validation",
				"exports":     []any{"Store", "Schema", "FieldDef", "NewStore", "Set", "Get", "List"},
			}},
			{ID: "serialize.go", Label: "serialize.go", X: 830, Y: 170, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 275, "bytes": 6787,
				"description": "JSON serialization and deserialization of graphs",
				"exports":     []any{"Marshal", "Unmarshal", "Snapshot", "MarshalOptions"},
			}},
			// Test files
			{ID: "graph_test.go", Label: "graph_test.go", X: 80, Y: 320, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 206, "bytes": 4497,
				"description": "Tests for core graph operations",
				"test_file_for": "graph.go",
			}},
			{ID: "traverse_test.go", Label: "traverse_test.go", X: 230, Y: 320, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 277, "bytes": 6260,
				"description": "Tests for traversal algorithms",
				"test_file_for": "traverse.go",
			}},
			{ID: "query_test.go", Label: "query_test.go", X: 380, Y: 320, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 144, "bytes": 3078,
				"description": "Tests for query operations",
				"test_file_for": "query.go",
			}},
			{ID: "task_test.go", Label: "task_test.go", X: 530, Y: 320, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 277, "bytes": 6143,
				"description": "Tests for task scheduler",
				"test_file_for": "task.go",
			}},
			{ID: "store_test.go", Label: "store_test.go", X: 680, Y: 320, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 296, "bytes": 6166,
				"description": "Tests for metadata store",
				"test_file_for": "store.go",
			}},
			{ID: "serialize_test.go", Label: "serialize_test.go", X: 830, Y: 320, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 446, "bytes": 11093,
				"description": "Tests for serialization round-trips",
				"test_file_for": "serialize.go",
			}},
			{ID: "meta_test.go", Label: "meta_test.go", X: 80, Y: 440, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 203, "bytes": 4883,
				"description": "Tests for graph metadata integration",
			}},
			// Config/docs
			{ID: "go.mod", Label: "go.mod", X: 230, Y: 440, Meta: map[string]any{
				"type": "file", "language": "Go Module", "lines": 3, "bytes": 22,
				"description": "Go module definition — zero external dependencies",
				"module": "spine", "go_version": "1.22",
			}},
			{ID: "Makefile", Label: "Makefile", X: 380, Y: 440, Meta: map[string]any{
				"type": "file", "language": "Make", "lines": 9, "bytes": 112,
				"description": "Build automation: test, visualize, run targets",
				"targets": []any{"test", "visualize", "run"},
			}},
			{ID: "README.md", Label: "README.md", X: 530, Y: 440, Meta: map[string]any{
				"type": "file", "language": "Markdown", "lines": 99, "bytes": 2552,
				"description": "Project documentation with usage examples",
			}},
			{ID: "doc.go", Label: "doc.go", X: 680, Y: 440, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 7, "bytes": 387,
				"description": "Package-level Go doc comment",
			}},
			{ID: ".gitignore", Label: ".gitignore", X: 830, Y: 440, Meta: map[string]any{
				"type": "file", "lines": 21, "bytes": 156,
				"description": "Git ignore rules for binaries, coverage, IDE files",
			}},
			// cmd directory tree
			{ID: "cmd", Label: "cmd/", X: 450, Y: 560, Meta: map[string]any{
				"type": "directory", "description": "Application entry points",
			}},
			{ID: "visualizer", Label: "visualizer/", X: 450, Y: 660, Meta: map[string]any{
				"type": "directory", "description": "Web-based interactive graph visualizer",
			}},
			{ID: "main.go", Label: "main.go", X: 280, Y: 770, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 908, "bytes": 22060,
				"description": "HTTP server with REST API for graph manipulation and algorithm execution",
				"endpoints": 20, "port": 8090,
			}},
			{ID: "templates.go", Label: "templates.go", X: 450, Y: 770, Meta: map[string]any{
				"type": "file", "language": "Go", "lines": 319, "bytes": 12614,
				"description": "Pre-built graph templates for demo and exploration",
				"template_count": 8,
			}},
			{ID: "static", Label: "static/", X: 620, Y: 770, Meta: map[string]any{
				"type": "directory", "description": "Embedded static web assets",
			}},
			{ID: "index.html", Label: "index.html", X: 620, Y: 880, Meta: map[string]any{
				"type": "file", "language": "HTML/JS/CSS", "lines": 1982, "bytes": 54469,
				"description": "Self-contained interactive frontend with canvas rendering",
				"features": []any{"graph editing", "algorithm visualization", "metadata inspector", "import/export"},
			}},
		},
		Edges: []templateEdge{
			// Root -> top-level source files
			{From: "spine", To: "graph.go", Label: "contains", Weight: 1},
			{From: "spine", To: "traverse.go", Label: "contains", Weight: 1},
			{From: "spine", To: "query.go", Label: "contains", Weight: 1},
			{From: "spine", To: "task.go", Label: "contains", Weight: 1},
			{From: "spine", To: "store.go", Label: "contains", Weight: 1},
			{From: "spine", To: "serialize.go", Label: "contains", Weight: 1},
			// Root -> test files
			{From: "spine", To: "graph_test.go", Label: "contains", Weight: 1},
			{From: "spine", To: "traverse_test.go", Label: "contains", Weight: 1},
			{From: "spine", To: "query_test.go", Label: "contains", Weight: 1},
			{From: "spine", To: "task_test.go", Label: "contains", Weight: 1},
			{From: "spine", To: "store_test.go", Label: "contains", Weight: 1},
			{From: "spine", To: "serialize_test.go", Label: "contains", Weight: 1},
			{From: "spine", To: "meta_test.go", Label: "contains", Weight: 1},
			// Root -> config/docs
			{From: "spine", To: "go.mod", Label: "contains", Weight: 1},
			{From: "spine", To: "Makefile", Label: "contains", Weight: 1},
			{From: "spine", To: "README.md", Label: "contains", Weight: 1},
			{From: "spine", To: "doc.go", Label: "contains", Weight: 1},
			{From: "spine", To: ".gitignore", Label: "contains", Weight: 1},
			// Root -> cmd directory
			{From: "spine", To: "cmd", Label: "contains", Weight: 1},
			// cmd -> visualizer
			{From: "cmd", To: "visualizer", Label: "contains", Weight: 1},
			// visualizer -> its files
			{From: "visualizer", To: "main.go", Label: "contains", Weight: 1},
			{From: "visualizer", To: "templates.go", Label: "contains", Weight: 1},
			{From: "visualizer", To: "static", Label: "contains", Weight: 1},
			// static -> index.html
			{From: "static", To: "index.html", Label: "contains", Weight: 1},
			// Source-to-test relationships
			{From: "graph.go", To: "graph_test.go", Label: "tested by", Weight: 1},
			{From: "traverse.go", To: "traverse_test.go", Label: "tested by", Weight: 1},
			{From: "query.go", To: "query_test.go", Label: "tested by", Weight: 1},
			{From: "task.go", To: "task_test.go", Label: "tested by", Weight: 1},
			{From: "store.go", To: "store_test.go", Label: "tested by", Weight: 1},
			{From: "serialize.go", To: "serialize_test.go", Label: "tested by", Weight: 1},
		},
	},
}
