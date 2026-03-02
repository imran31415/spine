# Spine

<img width="1512" height="913" alt="image" src="https://github.com/user-attachments/assets/abf30824-e661-4b4c-949c-0f314bd2d397" />
<img width="1511" height="904" alt="image" src="https://github.com/user-attachments/assets/c31abe0d-3ea2-4cd8-b6d6-58029bda491a" />


A generic graph library for Go with zero external dependencies. Supports directed and undirected graphs, traversal algorithms, attribute-based queries, task scheduling with dependency resolution, an MCP server for LLM integration, and an interactive web-based visualizer.

## Features

- **Generic graph** — typed node and edge data via Go generics
- **Directed & undirected** — toggle mode per graph instance
- **Traversal** — BFS, DFS, Dijkstra shortest path, topological sort
- **Cycle detection** — detect and return cycle paths
- **Connected components** — weakly connected component discovery
- **Strongly connected components** — Tarjan's SCC algorithm
- **Minimum spanning tree** — Kruskal's MST for undirected graphs
- **Graph analytics** — density, diameter, average degree, component count
- **Queries** — filter nodes/edges by predicate, find roots, leaves, ancestors, descendants
- **Task scheduler** — DAG-based task execution with state machine, dependency resolution, and concurrent runner
- **MCP server** — Model Context Protocol server exposing 21 tools for LLM-driven graph operations
- **API layer** — high-level `api.Manager` for named graph lifecycle, upsert, read, transitions
- **Web visualizer** — interactive canvas UI to build and explore graphs in the browser

## Install

```bash
go get github.com/imran31415/spine
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/imran31415/spine"
)

func main() {
    g := spine.NewGraph[string, float64](true) // directed

    g.AddNode("a", "Alpha")
    g.AddNode("b", "Beta")
    g.AddNode("c", "Gamma")

    _ = g.AddEdge("a", "b", 1.0, 1.0)
    _ = g.AddEdge("b", "c", 2.0, 3.0)
    _ = g.AddEdge("a", "c", 0.5, 10.0)

    // Shortest path (Dijkstra)
    path, cost, _ := spine.ShortestPath(g, "a", "c")
    fmt.Println(path, cost) // [a b c] 4

    // Topological sort
    order, _ := spine.TopologicalSort(g)
    fmt.Println(order) // [a b c]

    // Query
    roots := spine.Roots(g)
    fmt.Println(roots[0].ID) // a
}
```

## Task Scheduler

```go
tg := spine.NewTaskGraph[string]()
tg.AddTask("fetch", "Download data")
tg.AddTask("parse", "Parse response")
tg.AddTask("store", "Write to DB")

tg.AddDependency("parse", "fetch") // parse depends on fetch
tg.AddDependency("store", "parse") // store depends on parse

err := tg.Run(ctx, 4, func(task spine.Task[string]) error {
    fmt.Println("Running:", task.ID)
    return nil
})
```

## MCP Server

The MCP server exposes spine operations as tools over JSON-RPC 2.0 on stdio, enabling LLM agents to build and query graphs:

```bash
go run ./cmd/mcp/
```

### Available Tools (21)

| Category | Tools |
|----------|-------|
| **Lifecycle** | `open_graph`, `save_graph`, `list_graphs`, `delete_graph`, `graph_summary` |
| **CRUD** | `upsert`, `read_nodes`, `transition`, `remove` |
| **Traversal** | `bfs`, `dfs`, `shortest_path`, `topological_sort` |
| **Analysis** | `cycle_detect`, `connected_components`, `scc`, `mst` |
| **Queries** | `ancestors`, `descendants`, `roots`, `leaves` |

## API Layer

The `api.Manager` provides a high-level Go API for managing named graphs with persistence:

```go
import "github.com/imran31415/spine/api"

mgr, _ := api.NewManager("./graphs")
info, _ := mgr.Open("my-project")           // load or create
mgr.Upsert(api.UpsertRequest{...})          // batch create/update
mgr.Transition(api.TransitionRequest{...})   // status machine
mgr.Save("my-project")                      // persist to disk
```

## Visualizer

An interactive web UI with zero frontend dependencies:

```bash
go run ./cmd/visualizer/
# Open http://localhost:8090
```

- Double-click the canvas to add nodes
- Select a node, press `e`, then click another node to add an edge
- Drag nodes to reposition
- Run algorithms (BFS, DFS, shortest path, etc.) and see animated results
- Toggle directed/undirected mode

## Architecture

```
spine (core)          Generic graph, traversal, algorithms, serialization
  ├── api             High-level Manager (named graphs, upsert, transitions)
  ├── mcp             MCP server (JSON-RPC 2.0, 21 tools)
  └── cmd/
      ├── mcp         MCP server binary
      └── visualizer  Interactive web UI
```

## Testing

```bash
go test -v -race ./...
```

## License

MIT
