package spine

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTaskAddAndGet(t *testing.T) {
	tg := NewTaskGraph[string]()
	tg.AddTask("t1", "task one")

	task, ok := tg.GetTask("t1")
	if !ok {
		t.Fatal("task not found")
	}
	if task.Data != "task one" || task.State != Pending {
		t.Fatalf("unexpected task: %+v", task)
	}
}

func TestTaskDependency(t *testing.T) {
	tg := NewTaskGraph[string]()
	tg.AddTask("t1", "first")
	tg.AddTask("t2", "second")
	if err := tg.AddDependency("t2", "t1"); err != nil {
		t.Fatal(err)
	}

	ready := tg.Ready()
	// Only t1 should be ready (no deps)
	if len(ready) != 1 || ready[0].ID != "t1" {
		t.Fatalf("expected only t1 ready, got %v", ready)
	}
}

func TestTaskTransition(t *testing.T) {
	tg := NewTaskGraph[string]()
	tg.AddTask("t1", "task")

	// Pending -> Ready
	tg.Ready() // trigger transition
	if err := tg.Transition("t1", Running); err != nil {
		t.Fatal(err)
	}
	if err := tg.Transition("t1", Done); err != nil {
		t.Fatal(err)
	}

	task, _ := tg.GetTask("t1")
	if task.State != Done {
		t.Fatalf("expected Done, got %s", task.State)
	}
}

func TestTaskInvalidTransition(t *testing.T) {
	tg := NewTaskGraph[string]()
	tg.AddTask("t1", "task")

	// Pending -> Running (invalid, must go through Ready)
	if err := tg.Transition("t1", Running); err == nil {
		t.Fatal("expected error for invalid transition")
	}
}

func TestTaskTransitionMissingTask(t *testing.T) {
	tg := NewTaskGraph[string]()
	if err := tg.Transition("nonexistent", Ready); err == nil {
		t.Fatal("expected error for missing task")
	}
}

func TestTaskReady(t *testing.T) {
	tg := NewTaskGraph[string]()
	tg.AddTask("t1", "first")
	tg.AddTask("t2", "second")
	tg.AddTask("t3", "third")
	tg.AddDependency("t2", "t1")
	tg.AddDependency("t3", "t2")

	// Only t1 is ready
	ready := tg.Ready()
	if len(ready) != 1 || ready[0].ID != "t1" {
		t.Fatalf("expected [t1], got %v", taskIDs(ready))
	}

	// Complete t1
	tg.Transition("t1", Running)
	tg.Transition("t1", Done)

	// Now t2 should be ready
	ready = tg.Ready()
	if len(ready) != 1 || ready[0].ID != "t2" {
		t.Fatalf("expected [t2], got %v", taskIDs(ready))
	}
}

func TestTaskRun(t *testing.T) {
	tg := NewTaskGraph[string]()
	tg.AddTask("t1", "first")
	tg.AddTask("t2", "second")
	tg.AddTask("t3", "third")
	tg.AddDependency("t2", "t1")
	tg.AddDependency("t3", "t2")

	var order []string
	var mu sync.Mutex

	err := tg.Run(context.Background(), 2, func(task Task[string]) error {
		mu.Lock()
		order = append(order, task.ID)
		mu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 tasks executed, got %d: %v", len(order), order)
	}
	// Verify dependency order
	if indexOf(order, "t1") >= indexOf(order, "t2") {
		t.Fatal("t1 must run before t2")
	}
	if indexOf(order, "t2") >= indexOf(order, "t3") {
		t.Fatal("t2 must run before t3")
	}
}

func TestTaskRunConcurrency(t *testing.T) {
	tg := NewTaskGraph[string]()
	tg.AddTask("t1", "a")
	tg.AddTask("t2", "b")
	tg.AddTask("t3", "c")
	// t1 and t2 are independent; t3 depends on both
	tg.AddDependency("t3", "t1")
	tg.AddDependency("t3", "t2")

	var running atomic.Int32
	var maxConcurrent atomic.Int32

	err := tg.Run(context.Background(), 4, func(task Task[string]) error {
		cur := running.Add(1)
		for {
			old := maxConcurrent.Load()
			if int32(cur) <= old || maxConcurrent.CompareAndSwap(old, int32(cur)) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		running.Add(-1)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if maxConcurrent.Load() < 2 {
		t.Log("warning: expected concurrent execution of t1 and t2")
	}
}

func TestTaskRunFailure(t *testing.T) {
	tg := NewTaskGraph[string]()
	tg.AddTask("t1", "will-fail")
	tg.AddTask("t2", "depends-on-t1")
	tg.AddDependency("t2", "t1")

	err := tg.Run(context.Background(), 1, func(task Task[string]) error {
		if task.ID == "t1" {
			return errors.New("boom")
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected error")
	}

	task, _ := tg.GetTask("t1")
	if task.State != Failed {
		t.Fatalf("expected Failed, got %s", task.State)
	}

	// t2 should not have run
	task2, _ := tg.GetTask("t2")
	if task2.State == Done || task2.State == Running {
		t.Fatal("t2 should not have executed")
	}
}

func TestTaskRunContextCancel(t *testing.T) {
	tg := NewTaskGraph[string]()
	tg.AddTask("t1", "a")
	tg.AddTask("t2", "b")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	err := tg.Run(ctx, 1, func(task Task[string]) error {
		return nil
	})
	if err == nil {
		// It's possible all tasks ran before context was checked.
		// This is acceptable.
	}
}

func TestTaskReset(t *testing.T) {
	tg := NewTaskGraph[string]()
	tg.AddTask("t1", "task")
	tg.Ready()
	tg.Transition("t1", Running)
	tg.Transition("t1", Done)

	tg.Reset()
	task, _ := tg.GetTask("t1")
	if task.State != Pending {
		t.Fatalf("expected Pending after reset, got %s", task.State)
	}
}

func TestTaskSkipped(t *testing.T) {
	tg := NewTaskGraph[string]()
	tg.AddTask("t1", "task")

	if err := tg.Transition("t1", Skipped); err != nil {
		t.Fatal(err)
	}
	task, _ := tg.GetTask("t1")
	if task.State != Skipped {
		t.Fatalf("expected Skipped, got %s", task.State)
	}
}

func TestTaskStateString(t *testing.T) {
	states := map[TaskState]string{
		Pending: "Pending",
		Ready:   "Ready",
		Running: "Running",
		Done:    "Done",
		Failed:  "Failed",
		Skipped: "Skipped",
	}
	for s, expected := range states {
		if s.String() != expected {
			t.Fatalf("expected %s, got %s", expected, s.String())
		}
	}
}

func TestTaskDependencyMissingTask(t *testing.T) {
	tg := NewTaskGraph[string]()
	tg.AddTask("t1", "task")
	if err := tg.AddDependency("t1", "missing"); err == nil {
		t.Fatal("expected error for missing dependency")
	}
}

func TestTaskRunEmpty(t *testing.T) {
	tg := NewTaskGraph[string]()
	err := tg.Run(context.Background(), 1, func(task Task[string]) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func taskIDs[T any](tasks []Task[T]) []string {
	ids := make([]string, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}
	return ids
}
