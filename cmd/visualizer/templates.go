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
	ID     string  `json:"id"`
	Label  string  `json:"label"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Status string  `json:"status,omitempty"`
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
			{ID: "push", Label: "Push", X: 450, Y: 50},
			{ID: "lint", Label: "Lint", X: 200, Y: 180},
			{ID: "test", Label: "Test", X: 450, Y: 180},
			{ID: "build", Label: "Build", X: 700, Y: 180},
			{ID: "scan", Label: "Security Scan", X: 200, Y: 330},
			{ID: "coverage", Label: "Coverage", X: 450, Y: 330},
			{ID: "docker", Label: "Docker", X: 700, Y: 330},
			{ID: "staging", Label: "Staging", X: 450, Y: 480},
			{ID: "deploy", Label: "Deploy", X: 450, Y: 620},
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
			{ID: "ml", Label: "Machine Learning", X: 450, Y: 60},
			{ID: "nn", Label: "Neural Nets", X: 250, Y: 180},
			{ID: "dl", Label: "Deep Learning", X: 100, Y: 330},
			{ID: "cnn", Label: "CNN", X: 250, Y: 450},
			{ID: "nlp", Label: "NLP", X: 450, Y: 200},
			{ID: "transformers", Label: "Transformers", X: 450, Y: 370},
			{ID: "stats", Label: "Statistics", X: 700, Y: 100},
			{ID: "regression", Label: "Regression", X: 800, Y: 250},
			{ID: "pca", Label: "PCA", X: 700, Y: 400},
			{ID: "rl", Label: "Reinforcement", X: 600, Y: 550},
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
			{ID: "gateway", Label: "Gateway", X: 450, Y: 50},
			{ID: "auth", Label: "Auth", X: 200, Y: 180},
			{ID: "users", Label: "Users", X: 450, Y: 180},
			{ID: "orders", Label: "Orders", X: 700, Y: 180},
			{ID: "products", Label: "Products", X: 200, Y: 350},
			{ID: "payments", Label: "Payments", X: 450, Y: 350},
			{ID: "inventory", Label: "Inventory", X: 700, Y: 350},
			{ID: "notify", Label: "Notify", X: 300, Y: 520},
			{ID: "cache", Label: "Cache", X: 550, Y: 520},
			{ID: "db", Label: "Database", X: 700, Y: 520},
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
		Description: "Task execution plan with statuses — watch auto-ready promotion",
		Directed:    true,
		Nodes: []templateNode{
			{ID: "analyze", Label: "Analyze", X: 450, Y: 50, Status: "done"},
			{ID: "plan", Label: "Plan", X: 450, Y: 180, Status: "done"},
			{ID: "research", Label: "Research", X: 250, Y: 320, Status: "running"},
			{ID: "setup", Label: "Setup", X: 650, Y: 320, Status: "done"},
			{ID: "implement", Label: "Implement", X: 450, Y: 460, Status: "ready"},
			{ID: "test", Label: "Test", X: 450, Y: 590, Status: "pending"},
			{ID: "review", Label: "Review", X: 450, Y: 720, Status: "pending"},
			{ID: "deploy", Label: "Deploy", X: 450, Y: 850, Status: "pending"},
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
}
