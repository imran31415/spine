.PHONY: run test visualize

run: visualize

test:
	go test -v -race ./...

visualize:
	go run ./cmd/visualizer/
