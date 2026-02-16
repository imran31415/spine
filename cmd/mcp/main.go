// Command spine-mcp runs the MCP server for spine graphs over stdio.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"spine/api"
	"spine/mcp"
)

func main() {
	log.SetOutput(os.Stderr)

	dir := flag.String("dir", "", "graph storage directory (default: SPINE_GRAPH_DIR or current dir)")
	flag.Parse()

	if *dir == "" {
		*dir = os.Getenv("SPINE_GRAPH_DIR")
	}
	if *dir == "" {
		var err error
		*dir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot determine working directory: %v\n", err)
			os.Exit(1)
		}
	}

	mgr, err := api.NewManager(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create manager: %v\n", err)
		os.Exit(1)
	}

	srv := mcp.NewServer(mgr)
	log.Printf("spine-mcp server started (dir=%s)", *dir)
	if err := srv.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
