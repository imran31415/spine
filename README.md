# Spine

A generic graph library for Go with zero external dependencies. Supports directed and undirected graphs, traversal algorithms, attribute-based queries, task scheduling with dependency resolution, and an interactive web-based visualizer.

## Features

- **Generic graph** — typed node and edge data via Go generics
- **Directed & undirected** — toggle mode per graph instance
- **Traversal** — BFS, DFS, Dijkstra shortest path, topological sort
- **Cycle detection** — detect and return cycle paths
- **Connected components** — weakly connected component discovery
- **Queries** — filter nodes/edges by predicate, find roots, leaves, ancestors, descendants
- **Task scheduler** — DAG-based task execution with state machine, dependency resolution, and concurrent runner
- **Web visualizer** — interactive canvas UI to build and explore graphs in the browser

## Install

```bash
go get spine
```

## Quick Start

```go
package main

import (
    "fmt"
    "spine"
)

func main() {
    g := spine.NewGraph[string, float64](true) // directed

    g.AddNode("a", "Alpha")
    g.AddNode("b", "Beta")
    g.AddNode("c", "Gamma")

    g.AddEdge("a", "b", 1.0, 1.0)
    g.AddEdge("b", "c", 2.0, 3.0)
    g.AddEdge("a", "c", 0.5, 10.0)

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

## Testing

```bash
go test -v -race ./...
```

## License

MIT
