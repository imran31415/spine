.PHONY: run test visualize

GRAPH_DIR ?= /home/dev/.spine-graphs
PORT ?= 8090

run: visualize

test:
	go test -v -race ./...

visualize:
	@-pkill -f 'spine-visualizer\|cmd/visualizer' 2>/dev/null; sleep 1
	go run ./cmd/visualizer/ -dir $(GRAPH_DIR)
