package spine

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// TaskState represents the current state of a task.
type TaskState int

const (
	Pending TaskState = iota
	Ready
	Running
	Done
	Failed
	Skipped
)

func (s TaskState) String() string {
	switch s {
	case Pending:
		return "Pending"
	case Ready:
		return "Ready"
	case Running:
		return "Running"
	case Done:
		return "Done"
	case Failed:
		return "Failed"
	case Skipped:
		return "Skipped"
	default:
		return "Unknown"
	}
}

// validTransitions defines the allowed state transitions.
var validTransitions = map[TaskState][]TaskState{
	Pending: {Ready, Skipped},
	Ready:   {Running, Skipped},
	Running: {Done, Failed},
}

// Task represents a unit of work with typed data and a state.
type Task[T any] struct {
	ID    string
	Data  T
	State TaskState
}

// TaskGraph manages tasks with dependencies, state tracking, and execution.
type TaskGraph[T any] struct {
	mu    sync.Mutex
	graph *Graph[Task[T], struct{}]
}

// NewTaskGraph creates a new task graph.
func NewTaskGraph[T any]() *TaskGraph[T] {
	return &TaskGraph[T]{
		graph: NewGraph[Task[T], struct{}](true),
	}
}

// AddTask adds a task with the given ID and data. Initial state is Pending.
func (tg *TaskGraph[T]) AddTask(id string, data T) {
	tg.mu.Lock()
	defer tg.mu.Unlock()
	t := Task[T]{ID: id, Data: data, State: Pending}
	tg.graph.AddNode(id, t)
}

// AddDependency adds a dependency: task `from` depends on task `to`.
// This means `to` must complete before `from` can run.
func (tg *TaskGraph[T]) AddDependency(from, to string) error {
	tg.mu.Lock()
	defer tg.mu.Unlock()
	return tg.graph.AddEdge(to, from, struct{}{}, 0)
}

// Ready returns all tasks whose dependencies are all Done and whose state is Ready.
// It also transitions Pending tasks to Ready if all deps are met.
func (tg *TaskGraph[T]) Ready() []Task[T] {
	tg.mu.Lock()
	defer tg.mu.Unlock()
	return tg.readyLocked()
}

func (tg *TaskGraph[T]) readyLocked() []Task[T] {
	var ready []Task[T]
	for _, n := range tg.graph.Nodes() {
		task := n.Data
		if task.State == Pending && tg.allDepsDone(task.ID) {
			task.State = Ready
			tg.graph.AddNode(task.ID, task)
		}
		if task.State == Ready {
			ready = append(ready, task)
		}
	}
	return ready
}

func (tg *TaskGraph[T]) allDepsDone(id string) bool {
	// In our graph, edge to->from means from depends on to.
	// InEdges of id gives edges where id is the target, i.e., dep->id.
	// But we modeled it as: AddDependency(from, to) calls AddEdge(to, from).
	// So if from depends on to, edge is to->from.
	// InEdges(id) = edges pointing TO id, which means "id depends on source".
	// Wait: AddEdge(to, from) means edge.From=to, edge.To=from.
	// So InEdges(from) returns edges with To=from, i.e., the dependency edges.
	// The source of those edges (edge.From) are the dependencies.
	for _, e := range tg.graph.InEdges(id) {
		dep, ok := tg.graph.GetNode(e.From)
		if !ok || dep.Data.State != Done {
			return false
		}
	}
	return true
}

// Transition moves a task to a new state, validating the transition.
func (tg *TaskGraph[T]) Transition(id string, newState TaskState) error {
	tg.mu.Lock()
	defer tg.mu.Unlock()
	return tg.transitionLocked(id, newState)
}

func (tg *TaskGraph[T]) transitionLocked(id string, newState TaskState) error {
	n, ok := tg.graph.GetNode(id)
	if !ok {
		return fmt.Errorf("task %q not found", id)
	}
	task := n.Data
	allowed := validTransitions[task.State]
	for _, s := range allowed {
		if s == newState {
			task.State = newState
			tg.graph.AddNode(id, task)
			return nil
		}
	}
	return fmt.Errorf("invalid transition from %s to %s for task %q", task.State, newState, id)
}

// GetTask returns the current state of a task.
func (tg *TaskGraph[T]) GetTask(id string) (Task[T], bool) {
	tg.mu.Lock()
	defer tg.mu.Unlock()
	n, ok := tg.graph.GetNode(id)
	if !ok {
		var zero Task[T]
		return zero, false
	}
	return n.Data, true
}

// Graph returns the underlying graph for traversal/query operations.
func (tg *TaskGraph[T]) Graph() *Graph[Task[T], struct{}] {
	return tg.graph
}

// Run executes tasks in dependency order with the given concurrency limit.
// The fn function is called for each task. If fn returns an error, the task
// transitions to Failed; otherwise it transitions to Done.
// Returns an error if any task fails.
func (tg *TaskGraph[T]) Run(ctx context.Context, concurrency int, fn func(Task[T]) error) error {
	if concurrency < 1 {
		concurrency = 1
	}

	var mu sync.Mutex
	var taskErrors []error

	for {
		tg.mu.Lock()
		ready := tg.readyLocked()
		tg.mu.Unlock()

		if len(ready) == 0 {
			break
		}

		sem := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

		for _, task := range ready {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			tg.mu.Lock()
			err := tg.transitionLocked(task.ID, Running)
			tg.mu.Unlock()
			if err != nil {
				continue
			}

			wg.Add(1)
			sem <- struct{}{}
			go func(t Task[T]) {
				defer wg.Done()
				defer func() { <-sem }()

				// Re-read the task to get the Running state data.
				tg.mu.Lock()
				current, _ := tg.graph.GetNode(t.ID)
				tg.mu.Unlock()

				err := fn(current.Data)
				tg.mu.Lock()
				if err != nil {
					tg.transitionLocked(t.ID, Failed)
					mu.Lock()
					taskErrors = append(taskErrors, fmt.Errorf("task %q failed: %w", t.ID, err))
					mu.Unlock()
				} else {
					tg.transitionLocked(t.ID, Done)
				}
				tg.mu.Unlock()
			}(task)
		}
		wg.Wait()

		// If any tasks failed, stop scheduling new ones.
		mu.Lock()
		hasErrors := len(taskErrors) > 0
		mu.Unlock()
		if hasErrors {
			break
		}
	}

	if len(taskErrors) > 0 {
		return errors.Join(taskErrors...)
	}
	return nil
}

// Reset sets all tasks back to Pending.
func (tg *TaskGraph[T]) Reset() {
	tg.mu.Lock()
	defer tg.mu.Unlock()
	for _, n := range tg.graph.Nodes() {
		task := n.Data
		task.State = Pending
		tg.graph.AddNode(task.ID, task)
	}
}
